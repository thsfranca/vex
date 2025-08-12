package analysis

import (
	"fmt"
	"strings"

	"github.com/antlr4-go/antlr/v4"
	"github.com/thsfranca/vex/internal/transpiler/diagnostics"
	"github.com/thsfranca/vex/internal/transpiler/parser"
)

// AnalyzerImpl implements the Analyzer interface
// AnalyzerImpl performs semantic analysis and explicit type checking over Vex AST.
type AnalyzerImpl struct {
	symbolTable   *SymbolTableImpl
	errorReporter *ErrorReporterImpl
	typeEnv       TypeEnv
	freshId       int
	subst         Subst
	// Package environment (optional): provided by resolver/build when available
	ignoreImports map[string]bool                   // local Vex packages (not real Go imports)
	pkgExports    map[string]map[string]bool        // package -> exported symbols
	pkgSchemes    map[string]map[string]*TypeScheme // package -> symbol -> scheme
}

// installBuiltinSchemes seeds the type environment with polymorphic types for core functions
func (a *AnalyzerImpl) installBuiltinSchemes() {
	fresh := func() int { a.freshId++; return a.freshId }
	// helper to build a scheme from a type with its current free vars
	makeScheme := func(t Type) *TypeScheme { return generalize(a.typeEnv, t) }

	// Arithmetic & comparison & logic schemes
	num := &TypeConstant{Name: "number"}
	boolean := &TypeConstant{Name: "bool"}
	a.typeEnv["+"] = makeScheme(&TypeFunction{Params: []Type{num, num}, Result: num})
	a.typeEnv["-"] = makeScheme(&TypeFunction{Params: []Type{num, num}, Result: num})
	a.typeEnv["*"] = makeScheme(&TypeFunction{Params: []Type{num, num}, Result: num})
	a.typeEnv["/"] = makeScheme(&TypeFunction{Params: []Type{num, num}, Result: num})
	a.typeEnv[">"] = makeScheme(&TypeFunction{Params: []Type{num, num}, Result: boolean})
	a.typeEnv["<"] = makeScheme(&TypeFunction{Params: []Type{num, num}, Result: boolean})
	aVarEq := &TypeVariable{ID: fresh()}
	a.typeEnv["="] = makeScheme(&TypeFunction{Params: []Type{aVarEq, aVarEq}, Result: boolean})
	a.typeEnv["not"] = makeScheme(&TypeFunction{Params: []Type{boolean}, Result: boolean})
	a.typeEnv["and"] = makeScheme(&TypeFunction{Params: []Type{boolean, boolean}, Result: boolean})
	a.typeEnv["or"] = makeScheme(&TypeFunction{Params: []Type{boolean, boolean}, Result: boolean})

	// (first [a]) -> a
	aVar := &TypeVariable{ID: fresh()}
	firstT := &TypeFunction{Params: []Type{&TypeArray{Elem: aVar}}, Result: aVar}
	a.typeEnv["first"] = makeScheme(firstT)

	// (rest [a]) -> [a]
	aVar2 := &TypeVariable{ID: fresh()}
	restT := &TypeFunction{Params: []Type{&TypeArray{Elem: aVar2}}, Result: &TypeArray{Elem: aVar2}}
	a.typeEnv["rest"] = makeScheme(restT)

	// (cons a [a]) -> [a]
	aVar3 := &TypeVariable{ID: fresh()}
	consT := &TypeFunction{Params: []Type{aVar3, &TypeArray{Elem: aVar3}}, Result: &TypeArray{Elem: aVar3}}
	a.typeEnv["cons"] = makeScheme(consT)

	// (count [a]) -> number (int internally)
	a.typeEnv["count"] = makeScheme(&TypeFunction{Params: []Type{&TypeArray{Elem: &TypeVariable{ID: fresh()}}}, Result: &TypeConstant{Name: "int"}})

	// (empty? [a]) -> bool
	a.typeEnv["empty?"] = makeScheme(&TypeFunction{Params: []Type{&TypeArray{Elem: &TypeVariable{ID: fresh()}}}, Result: &TypeConstant{Name: "bool"}})

	// (len [a]) -> number (int internally)
	a.typeEnv["len"] = makeScheme(&TypeFunction{Params: []Type{&TypeArray{Elem: &TypeVariable{ID: fresh()}}}, Result: &TypeConstant{Name: "int"}})
}

// NewAnalyzer creates a new analyzer
// NewAnalyzer constructs an AnalyzerImpl with built-in type schemes installed.
func NewAnalyzer() *AnalyzerImpl {
	a := &AnalyzerImpl{
		symbolTable:   NewSymbolTable(),
		errorReporter: NewErrorReporter(),
		typeEnv:       make(TypeEnv),
		freshId:       0,
		subst:         Subst{},
		ignoreImports: make(map[string]bool),
		pkgExports:    make(map[string]map[string]bool),
		pkgSchemes:    make(map[string]map[string]*TypeScheme),
	}
	a.installBuiltinSchemes()
	return a
}

// Analyze performs semantic analysis on the AST
func (a *AnalyzerImpl) Analyze(ast AST) (SymbolTable, error) {
	// Reset for new analysis
	a.symbolTable = NewSymbolTable()
	a.errorReporter.Clear()
	a.typeEnv = make(TypeEnv)
	a.installBuiltinSchemes()
	a.freshId = 0
	a.subst = Subst{}

	// Visit the AST
	if err := ast.Accept(a); err != nil {
		return nil, err
	}

	// Check for errors
	if a.errorReporter.HasErrors() {
		return nil, fmt.Errorf("analysis failed with errors:\n%s", a.errorReporter.FormatErrors())
	}

	return a.symbolTable, nil
}

// SetPackageEnv injects package boundary information for local Vex packages.
// ignore marks import paths that are local Vex packages; exports lists visible symbols
// and schemes provides typing for exported symbols.
func (a *AnalyzerImpl) SetPackageEnv(ignore map[string]bool, exports map[string]map[string]bool, schemes map[string]map[string]*TypeScheme) {
	if ignore != nil {
		a.ignoreImports = ignore
	} else {
		a.ignoreImports = make(map[string]bool)
	}
	if exports != nil {
		a.pkgExports = exports
	} else {
		a.pkgExports = make(map[string]map[string]bool)
	}
	if schemes != nil {
		a.pkgSchemes = schemes
	} else {
		a.pkgSchemes = make(map[string]map[string]*TypeScheme)
	}
}

// SetErrorReporter sets the error reporter
func (a *AnalyzerImpl) SetErrorReporter(reporter ErrorReporter) {
	if impl, ok := reporter.(*ErrorReporterImpl); ok {
		a.errorReporter = impl
	}
}

// GetErrorReporter returns the current error reporter
func (a *AnalyzerImpl) GetErrorReporter() *ErrorReporterImpl {
	return a.errorReporter
}

// GetTypeScheme returns the generalized type scheme for a symbol if present in the environment.
func (a *AnalyzerImpl) GetTypeScheme(name string) (*TypeScheme, bool) {
	sch, ok := a.typeEnv[name]
	if !ok {
		return nil, false
	}
	return sch, true
}

// VisitProgram analyzes a program node
func (a *AnalyzerImpl) VisitProgram(ctx *parser.ProgramContext) error {
	for _, child := range ctx.GetChildren() {
		if listCtx, ok := child.(*parser.ListContext); ok {
			_, err := a.VisitList(listCtx)
			if err != nil {
				// Error already reported, continue analysis
				continue
			}
		}
	}
	return nil
}

// VisitList analyzes a list expression
func (a *AnalyzerImpl) VisitList(ctx *parser.ListContext) (Value, error) {
	childCount := ctx.GetChildCount()
	if childCount < 3 { // Need at least: '(', function, ')'
		line := ctx.GetStart().GetLine()
		column := ctx.GetStart().GetColumn()
		diag := diagnostics.New(diagnostics.CodeSyntaxEmpty, diagnostics.SeverityError, "", line, column, nil)
		a.errorReporter.ReportDiagnosticBody(line, column, diag.RenderBody(), SemanticError)
		return nil, fmt.Errorf("empty expression")
	}

	// Get function name (first child after '(')
	funcNameNode := ctx.GetChild(1)
	funcName := a.nodeToString(funcNameNode)

	// Special handling: if funcName is a declared record, support constructor and field calls
	if sym, ok := a.symbolTable.GetSymbol(funcName); ok {
		if _, isRec := sym.Value.(*RecordValue); isRec {
			// Constructor: (RecordName [fields...])
			if childCount >= 4 {
				if _, okArr := ctx.GetChild(2).(*parser.ArrayContext); okArr {
					return a.analyzeRecordConstructor(ctx, sym.Value.(*RecordValue))
				}
			}
			// Field call: (RecordName [:field])
			return a.analyzeRecordFieldCall(ctx, sym)
		}
	}

	// Extract arguments
	var args []Value
	for i := 2; i < childCount-1; i++ { // Skip '(' and ')'
		child := ctx.GetChild(i)
		if child == nil {
			continue
		}
		switch funcName {
		case "macro":
			argText := a.nodeToFlatText(child)
			args = append(args, NewBasicValue(argText, "raw").MarkRaw())
		case "def":
			if len(args) == 0 {
				argText := a.nodeToFlatText(child)
				args = append(args, NewBasicValue(argText, "raw").MarkRaw())
				continue
			}
			fallthrough
		case "fn":
			// For 'fn', capture arguments as raw to preserve type annotations
			// Handle both (fn [params] body) and (fn [params] -> type body)
			if funcName == "fn" && len(args) < 4 {
				argText := a.nodeToFlatText(child)
				args = append(args, NewBasicValue(argText, "raw").MarkRaw())
				continue
			}
			fallthrough
		case "record":
			// Treat record children as raw to avoid premature symbol lookups
			if funcName == "record" {
				argText := a.nodeToFlatText(child)
				args = append(args, NewBasicValue(argText, "raw").MarkRaw())
				continue
			}
			fallthrough
		case "export":
			if funcName == "export" {
				argText := a.nodeToFlatText(child)
				args = append(args, NewBasicValue(argText, "raw"))
				continue
			}
			fallthrough
		default:
			arg, err := a.visitNode(child)
			if err != nil {
				continue
			}
			args = append(args, arg)
		}
	}

	// Analyze special forms
	switch funcName {
	case "def":
		return a.analyzeDef(ctx, args)
	case "defn":
		return a.analyzeDefn(ctx, args)
	case "if":
		return a.analyzeIf(ctx, args)
	case "do":
		return a.analyzeDo(ctx, args)
	case "fn":
		return a.analyzeFn(ctx, args)
	case "macro":
		return a.analyzeMacro(ctx, args)
	case "import":
		// Handle stdlib imports - they are already loaded during macro expansion phase
		if len(args) > 0 {
			importName := args[0].String()
			// Remove quotes if present
			importName = strings.Trim(importName, "\"")
			// Stdlib imports (vex.*) are handled by macro system, just accept them
			if strings.HasPrefix(importName, "vex.") {
				return NewBasicValue("import", "void"), nil
			}
		}
		// Other imports are handled in codegen/pipeline
		return NewBasicValue("import", "void"), nil
	case "export":
		return a.analyzeExport(ctx, args)
	case "record":
		return a.analyzeRecord(ctx, args)
	case "map":
		return a.analyzeMap(ctx, args)
	default:
		return a.analyzeFunctionCall(ctx, funcName, args)
	}
}

// analyzeExport records exported symbols for current package scope
func (a *AnalyzerImpl) analyzeExport(ctx *parser.ListContext, args []Value) (Value, error) {
	if len(args) < 1 {
		line := ctx.GetStart().GetLine()
		column := ctx.GetStart().GetColumn()
		diag := diagnostics.New(diagnostics.CodeExportArguments, diagnostics.SeverityError, "", line, column, nil)
		a.errorReporter.ReportDiagnosticBody(line, column, diag.RenderBody(), SemanticError)
		return nil, fmt.Errorf("invalid export")
	}
	// Minimal placeholder: accept and no-op. Enforcement will be performed cross-package in future step.
	return NewBasicValue("export", "void"), nil
}

// VisitArray analyzes an array literal
func (a *AnalyzerImpl) VisitArray(ctx *parser.ArrayContext) (Value, error) {
	elements := make([]Value, 0)

	// Analyze all elements
	for i := 1; i < ctx.GetChildCount()-1; i++ { // Skip '[' and ']'
		child := ctx.GetChild(i)
		if child != nil {
			element, err := a.visitNode(child)
			if err != nil {
				continue // Error already reported
			}
			elements = append(elements, element)
		}
	}
	// HM: unify element types using global substitution; outward type remains []interface{}
	var elemT Type
	if len(elements) > 0 {
		elemT = a.typeFromValue(elements[0]).apply(a.subst)
		for i := 1; i < len(elements); i++ {
			nextT := a.typeFromValue(elements[i]).apply(a.subst)
			if s, err := unify(elemT, nextT); err == nil {
				a.subst = a.subst.compose(s)
				elemT = elemT.apply(a.subst)
			} else {
				// Strict HM: incompatible element types should be an error
				line := ctx.GetStart().GetLine()
				column := ctx.GetStart().GetColumn()
				diag := diagnostics.New(diagnostics.CodeTypArrayElem, diagnostics.SeverityError, "", line, column, map[string]any{"Got": fmt.Sprintf("%s vs %s", a.publicTypeString(elemT), a.publicTypeString(nextT))})
				a.errorReporter.ReportDiagnosticBody(line, column, diag.RenderBody(), TypeError)
			}
		}
	} else {
		elemT = a.freshTypeVar()
	}
	return NewBasicValue("array", "[]interface{}").WithType(&TypeArray{Elem: elemT}), nil
}

// VisitTerminal analyzes a terminal node
func (a *AnalyzerImpl) VisitTerminal(node antlr.TerminalNode) (Value, error) {
	text := node.GetText()

	// Determine type based on content
	if strings.HasPrefix(text, "\"") && strings.HasSuffix(text, "\"") {
		return NewBasicValue(text, "string"), nil
	}

	// Numbers: float or int; keep outward type as "number" and store concrete internally
	if isFloat(text) {
		return NewBasicValue(text, "number").WithType(&TypeConstant{Name: "float"}), nil
	}
	if isInt(text) {
		return NewBasicValue(text, "number").WithType(&TypeConstant{Name: "int"}), nil
	}

	// Check if it's a boolean
	if text == "true" || text == "false" {
		return NewBasicValue(text, "bool"), nil
	}

	// Symbol: if defined, instantiate scheme and return a fresh value copy with attached type
	if value, exists := a.symbolTable.Lookup(text); exists {
		if sch, ok := a.typeEnv[text]; ok {
			inst := instantiate(sch, func() int { a.freshId++; return a.freshId })
			// return a fresh value rather than mutating the stored symbol value
			return NewBasicValue(value.String(), value.Type()).WithType(inst), nil
		}
		// If we have a concrete internal type on the stored value, also return a fresh copy
		if bv, okb := value.(*BasicValue); okb && bv.getType() != nil {
			return NewBasicValue(bv.String(), bv.Type()).WithType(bv.getType()), nil
		}
		return value, nil
	}
	// Unknown symbol: report error for unresolved identifier
	// Do not treat as a fresh type var; enforce strict HM unknown-identifier policy
	line, col := 0, 0
	if tok := node.GetSymbol(); tok != nil {
		line = tok.GetLine()
		col = tok.GetColumn()
	}
	diag := diagnostics.New(diagnostics.CodeTypeUndefined, diagnostics.SeverityError, "", line, col, map[string]any{"Name": text})
	a.errorReporter.ReportDiagnosticBody(line, col, diag.RenderBody(), TypeError)
	return NewBasicValue(text, "undefined"), fmt.Errorf("undefined identifier: %s", text)
}

// analyzeDef analyzes a definition
func (a *AnalyzerImpl) analyzeDef(ctx *parser.ListContext, args []Value) (Value, error) {
	if len(args) < 2 {
		line := ctx.GetStart().GetLine()
		column := ctx.GetStart().GetColumn()
		diag := diagnostics.New(diagnostics.CodeDefArguments, diagnostics.SeverityError, "", line, column, nil).WithSuggestion("use (def name value)")
		a.errorReporter.ReportDiagnosticBody(line, column, diag.RenderBody(), SemanticError)
		return nil, fmt.Errorf("invalid def")
	}

	name := args[0].String()
	value := args[1]

	// Validate symbol naming convention for all symbols
	if !isValidSymbolName(name) {
		line := ctx.GetStart().GetLine()
		column := ctx.GetStart().GetColumn()
		diag := diagnostics.New(diagnostics.CodeSymNaming, diagnostics.SeverityError, "", line, column, map[string]any{"Name": name}).WithSuggestion("use kebab-case with dashes (e.g., 'my-symbol')")
		a.errorReporter.ReportDiagnosticBody(line, column, diag.RenderBody(), SemanticError)
		return nil, fmt.Errorf("invalid symbol name: %s", name)
	}

	// Define the symbol
	if err := a.symbolTable.Define(name, value); err != nil {
		line := ctx.GetStart().GetLine()
		column := ctx.GetStart().GetColumn()
		a.errorReporter.ReportError(line, column, err.Error())
		return nil, err
	}

	// Generalize and store scheme in type environment with value restriction
	if bv, ok := value.(*BasicValue); ok {
		var t Type
		if bv.getType() != nil {
			t = bv.getType()
		} else {
			switch bv.Type() {
			case "string", "bool", "number":
				t = &TypeConstant{Name: bv.Type()}
			case "[]interface{}":
				t = &TypeArray{Elem: a.freshTypeVar()}
			case "map[interface{}]interface{}":
				t = &TypeMap{Key: a.freshTypeVar(), Val: a.freshTypeVar()}
			case "record":
				t = &TypeConstant{Name: "record"}
			case "func":
				t = &TypeFunction{Params: []Type{}, Result: a.freshTypeVar()}
			default:
				t = &TypeConstant{Name: bv.Type()}
			}
		}
		if a.shouldGeneralizeValue(bv, t) {
			a.typeEnv[name] = generalize(a.typeEnv, t)
		}
	}

	return value, nil
}

// analyzeDefn analyzes a function definition (defn name [params] body)
// Converts to (def name (fn [params] body)) and delegates to analyzeDef
func (a *AnalyzerImpl) analyzeDefn(ctx *parser.ListContext, args []Value) (Value, error) {
	if len(args) < 3 {
		line := ctx.GetStart().GetLine()
		column := ctx.GetStart().GetColumn()
		diag := diagnostics.New(diagnostics.CodeDefArguments, diagnostics.SeverityError, "", line, column, nil).WithSuggestion("use (defn name [params] body)")
		a.errorReporter.ReportDiagnosticBody(line, column, diag.RenderBody(), SemanticError)
		return nil, fmt.Errorf("invalid defn")
	}

	name := args[0]
	params := args[1]
	body := args[2]

	// Create function value from params and body: (fn [params] body)
	fnValue, err := a.analyzeFn(ctx, []Value{params, body})
	if err != nil {
		return nil, err
	}

	// Now define the function: (def name fnValue)
	return a.analyzeDef(ctx, []Value{name, fnValue})
}

// shouldGeneralizeValue applies a lightweight value restriction:
// generalize only syntactic values such as literals, functions, and record constructors.
func (a *AnalyzerImpl) shouldGeneralizeValue(v *BasicValue, t Type) bool {
	// Disallow generalization for non-values like application/do results
	if v.value == "call-result" || v.value == "do-result" {
		return false
	}
	switch v.Type() {
	case "string", "bool", "number":
		return true
	case "func":
		return true
	case "record":
		return true
	case "[]interface{}", "map[interface{}]interface{}":
		// Conservatively allow generalization of empty/constructed collections
		return true
	default:
		// Do not generalize results of applications or unknown shapes
		return false
	}
}

// analyzeIf analyzes an if expression
func (a *AnalyzerImpl) analyzeIf(ctx *parser.ListContext, args []Value) (Value, error) {
	if len(args) < 2 {
		line := ctx.GetStart().GetLine()
		column := ctx.GetStart().GetColumn()
		diag := diagnostics.New(diagnostics.CodeIfArgs, diagnostics.SeverityError, "", line, column, nil).WithSuggestion("use (if condition then-expr [else-expr])")
		a.errorReporter.ReportDiagnosticBody(line, column, diag.RenderBody(), SemanticError)
		return nil, fmt.Errorf("invalid if")
	}

	// Analyze condition
	condition := args[0]
	thenBranch := args[1]

	// Optional else branch
	var elseBranch Value
	if len(args) > 2 {
		elseBranch = args[2]
	}

	// Type checking: condition should be boolean-compatible
	if condition.Type() != "bool" {
		line := ctx.GetStart().GetLine()
		column := ctx.GetStart().GetColumn()
		diag := diagnostics.New(diagnostics.CodeTypCond, diagnostics.SeverityError, "", line, column, map[string]any{"Got": condition.Type()})
		a.errorReporter.ReportDiagnosticBody(line, column, diag.RenderBody(), TypeError)
	}

	// Result type: unify then/else via global substitution
	thenT := a.typeFromValue(thenBranch).apply(a.subst)
	var resultT Type = thenT
	if elseBranch != nil {
		elseT := a.typeFromValue(elseBranch).apply(a.subst)
		if s, err := unify(thenT, elseT); err == nil {
			a.subst = a.subst.compose(s)
			resultT = resultT.apply(a.subst)
		} else {
			line := ctx.GetStart().GetLine()
			column := ctx.GetStart().GetColumn()
			// If both are nominal records but different names, emit specific nominal mismatch
			if ttc, ok1 := thenT.(*TypeConstant); ok1 && ttc.Name != "record" {
				if etc, ok2 := elseT.(*TypeConstant); ok2 && etc.Name != "record" && etc.Name != ttc.Name {
					diag := diagnostics.New(diagnostics.CodeRecNominal, diagnostics.SeverityError, "", line, column, map[string]any{"Got": fmt.Sprintf("then=%s, else=%s", ttc.Name, etc.Name)})
					a.errorReporter.ReportDiagnosticBody(line, column, diag.RenderBody(), TypeError)
				}
			}
			diag := diagnostics.New(diagnostics.CodeTypIfMismatch, diagnostics.SeverityError, "", line, column, map[string]any{"Expected": "type(then) == type(else)", "Got": fmt.Sprintf("then=%s, else=%s", a.publicTypeString(thenT), a.publicTypeString(elseT))})
			a.errorReporter.ReportDiagnosticBody(line, column, diag.RenderBody(), TypeError)
		}
	}
	return NewBasicValue("if-result", a.publicTypeString(resultT)).WithType(resultT), nil
}

// analyzeDo analyzes a do/block by evaluating all forms and returning the last form's type
func (a *AnalyzerImpl) analyzeDo(ctx *parser.ListContext, args []Value) (Value, error) {
	if len(args) == 0 {
		return NewBasicValue("do", "interface{}"), nil
	}

	// Process all expressions in the do block
	// We don't need to unify all types - just evaluate each expression for side effects
	// and return the type of the last expression
	var lastValue Value = args[0]
	for i := 1; i < len(args); i++ {
		lastValue = args[i]
	}

	// Return the type of the last expression
	lastT := a.typeFromValue(lastValue).apply(a.subst)
	return NewBasicValue("do-result", a.publicTypeString(lastT)).WithType(lastT), nil
}

// analyzeFn analyzes a function definition
func (a *AnalyzerImpl) analyzeFn(ctx *parser.ListContext, args []Value) (Value, error) {
	if len(args) < 2 {
		line := ctx.GetStart().GetLine()
		column := ctx.GetStart().GetColumn()
		diag := diagnostics.New(diagnostics.CodeFnArgs, diagnostics.SeverityError, "", line, column, nil).WithSuggestion("use (fn [params] body)")
		a.errorReporter.ReportDiagnosticBody(line, column, diag.RenderBody(), SemanticError)
		return nil, fmt.Errorf("invalid fn")
	}

	// Enter new scope for function parameters
	a.symbolTable.EnterScope()
	defer a.symbolTable.ExitScope()

	// Parse parameter list with required type annotations
	paramTypes := make([]Type, 0)
	paramList := args[0].String()
	if strings.HasPrefix(paramList, "[") && strings.HasSuffix(paramList, "]") {
		inner := strings.TrimSpace(paramList[1 : len(paramList)-1])
		if inner != "" {
			paramTypes = a.parseParameterListWithTypes(inner)
			if paramTypes == nil {
				line := ctx.GetStart().GetLine()
				column := ctx.GetStart().GetColumn()
				diag := diagnostics.New(diagnostics.CodeFnParams, diagnostics.SeverityError, "", line, column, nil)
				a.errorReporter.ReportDiagnosticBody(line, column, diag.RenderBody(), TypeError)
				return nil, fmt.Errorf("missing parameter type annotations")
			}
		}
	}

	// Parse required return type annotation
	var resultType Type
	if len(args) >= 3 && args[1].String() == "->" {
		// (fn [params] -> ReturnType body)
		returnTypeStr := args[2].String()
		resultType = a.typeNameToType(returnTypeStr)
		// Body is now args[3]
		bodyText := ""
		if len(args) > 3 {
			bodyText = args[3].String()
		}
		args = append(args[:1], &BasicValue{value: bodyText, typ: "expression"}) // Replace with body
	} else {
		// ERROR: Return type annotation is required
		line := ctx.GetStart().GetLine()
		column := ctx.GetStart().GetColumn()
		diag := diagnostics.New(diagnostics.CodeFnRetType, diagnostics.SeverityError, "", line, column, nil)
		a.errorReporter.ReportDiagnosticBody(line, column, diag.RenderBody(), TypeError)
		return nil, fmt.Errorf("missing return type annotation")
	}

	// Create function type with explicit result type
	fnType := &TypeFunction{Params: paramTypes, Result: resultType}

	// Check body type by parsing the body and visiting the AST
	bodyText := args[1].String()
	if len(bodyText) > 0 && bodyText[0] != '(' && bodyText[0] != '[' {
		// Primitive or symbol literal in body; determine type directly without parsing
		primitiveType := inferPrimitiveTypeString(bodyText)
		bodyType := a.typeNameToType(primitiveType)
		if s, errU := unify(fnType.Result, bodyType); errU == nil {
			appliedParams := make([]Type, len(fnType.Params))
			for i, p := range fnType.Params {
				appliedParams[i] = p.apply(s)
			}
			fnType = &TypeFunction{Params: appliedParams, Result: fnType.Result.apply(s)}
		}
	} else {
		node := a.parseSingleExpression(bodyText)
		if node != nil {
			// Visit with current scope (params already bound)
			if bodyVal, err := a.visitNode(node); err == nil && bodyVal != nil {
				bodyType := a.typeFromValue(bodyVal)
				if s, errU := unify(fnType.Result, bodyType); errU == nil {
					appliedParams := make([]Type, len(fnType.Params))
					for i, p := range fnType.Params {
						appliedParams[i] = p.apply(s)
					}
					fnType = &TypeFunction{Params: appliedParams, Result: fnType.Result.apply(s)}
				}
			}
		}
	}

	return NewBasicValue("function", "func").WithType(fnType), nil
}

// analyzeMacro analyzes a macro definition
func (a *AnalyzerImpl) analyzeMacro(ctx *parser.ListContext, args []Value) (Value, error) {
	if len(args) < 3 {
		line := ctx.GetStart().GetLine()
		column := ctx.GetStart().GetColumn()
		diag := diagnostics.New(diagnostics.CodeMacArgs, diagnostics.SeverityError, "", line, column, nil)
		a.errorReporter.ReportDiagnosticBody(line, column, diag.RenderBody(), MacroError)
		return nil, fmt.Errorf("invalid macro")
	}

	name := args[0].String()

	// Check for valid macro name
	if isReservedWord(name) {
		line := ctx.GetStart().GetLine()
		column := ctx.GetStart().GetColumn()
		diag := diagnostics.New(diagnostics.CodeMacReserved, diagnostics.SeverityError, "", line, column, map[string]any{"Name": name}).WithSuggestion("choose a different macro name")
		a.errorReporter.ReportDiagnosticBody(line, column, diag.RenderBody(), MacroError)
		return nil, fmt.Errorf("invalid macro name")
	}

	// Validate macro naming convention
	if !isValidSymbolName(name) {
		line := ctx.GetStart().GetLine()
		column := ctx.GetStart().GetColumn()
		diag := diagnostics.New(diagnostics.CodeSymNaming, diagnostics.SeverityError, "", line, column, map[string]any{"Name": name}).WithSuggestion("use kebab-case with dashes (e.g., 'my-macro')")
		a.errorReporter.ReportDiagnosticBody(line, column, diag.RenderBody(), MacroError)
		return nil, fmt.Errorf("invalid macro name: %s", name)
	}

	// Define the macro as a symbol
	macroValue := NewBasicValue(name, "macro")
	a.symbolTable.Define(name, macroValue)

	return macroValue, nil
}

// analyzeFunctionCall analyzes a function call
func (a *AnalyzerImpl) analyzeFunctionCall(ctx *parser.ListContext, funcName string, args []Value) (Value, error) {
	// Equality specialized typing: ∀a. a -> a -> bool with explicit mismatch diagnostic
	if funcName == "=" {
		if len(args) != 2 {
			line := ctx.GetStart().GetLine()
			column := ctx.GetStart().GetColumn()
			a.errorReporter.ReportTypedError(line, column, "[VEX-ARI-ARGS]: '=' expects 2 arguments", TypeError)
			return NewBasicValue("call-result", "bool"), nil
		}
		t1 := a.typeFromValue(args[0]).apply(a.subst)
		t2 := a.typeFromValue(args[1]).apply(a.subst)
		if s, err := unify(t1, t2); err == nil {
			a.subst = a.subst.compose(s)
			return NewBasicValue("call-result", "bool").WithType(&TypeConstant{Name: "bool"}), nil
		}
		line := ctx.GetStart().GetLine()
		column := ctx.GetStart().GetColumn()
		diag := diagnostics.New(diagnostics.CodeTypEq, diagnostics.SeverityError, "", line, column, map[string]any{"Got": a.publicTypeString(t1) + " vs " + a.publicTypeString(t2)}).WithMessage("occur-check or mismatch: cannot unify")
		a.errorReporter.ReportDiagnosticBody(line, column, diag.RenderBody(), TypeError)
		return NewBasicValue("call-result", "bool").WithType(&TypeConstant{Name: "bool"}), nil
	}
	// Namespaced call handling: pkg/path/Func
	if strings.Contains(funcName, "/") {
		parts := strings.Split(funcName, "/")
		if len(parts) >= 2 {
			importPath := strings.Join(parts[:len(parts)-1], "/")
			function := parts[len(parts)-1]
			// Local Vex package: enforce exports and attempt scheme-based typing
			if a.ignoreImports[importPath] {
				if ex, ok := a.pkgExports[importPath]; ok {
					if !ex[function] {
						line := ctx.GetStart().GetLine()
						column := ctx.GetStart().GetColumn()
						diag := diagnostics.New(diagnostics.CodePkgNotExported, diagnostics.SeverityError, "", line, column, map[string]any{"Name": function, "Pkg": importPath})
						a.errorReporter.ReportDiagnosticBody(line, column, diag.RenderBody(), TypeError)
						return NewBasicValue("call-result", "interface{}"), fmt.Errorf("symbol '%s' is not exported from package '%s'", function, importPath)
					}
				}
				// Type via provided package scheme, if available
				if schs, ok := a.pkgSchemes[importPath]; ok {
					if sch, ok2 := schs[function]; ok2 {
						ft, okf := instantiate(sch, func() int { a.freshId++; return a.freshId }).(*TypeFunction)
						if okf {
							if len(ft.Params) != len(args) {
								line := ctx.GetStart().GetLine()
								column := ctx.GetStart().GetColumn()
								a.errorReporter.ReportTypedError(line, column, fmt.Sprintf("[VEX-ARI-ARGS]: %s expects %d arguments; got %d", funcName, len(ft.Params), len(args)), TypeError)
								return NewBasicValue("call-result", "interface{}"), nil
							}
							subst := a.subst
							for i := range ft.Params {
								pt := ft.Params[i].apply(subst)
								at := a.typeFromValue(args[i]).apply(subst)
								s, err := unify(pt, at)
								if err != nil {
									line := ctx.GetStart().GetLine()
									column := ctx.GetStart().GetColumn()
									a.errorReporter.ReportTypedError(line, column, fmt.Sprintf("[VEX-TYP-ARG]: argument %d type mismatch", i), TypeError)
									return NewBasicValue("call-result", "interface{}"), nil
								}
								subst = subst.compose(s)
							}
							a.subst = subst
							rt := ft.Result.apply(a.subst)
							return NewBasicValue("call-result", a.publicTypeString(rt)).WithType(rt), nil
						}
					}
				}
				// No scheme: accept call without typing info
				return NewBasicValue("call-result", "interface{}"), nil
			}
		}
	}
	// Variadic arithmetic folding for '+' to match language/tests semantics
	if funcName == "+" {
		if len(args) == 0 {
			return NewBasicValue("0", "number").WithType(&TypeConstant{Name: "int"}), nil
		}
		if len(args) == 1 {
			// (+ x) => x
			return args[0], nil
		}
		// Fold left: (((a0 + a1) + a2) + ...)
		sawFloat := false
		accType := a.typeFromValue(args[0]).apply(a.subst)
		if tc, ok := accType.(*TypeConstant); ok && tc.Name == "float" {
			sawFloat = true
		}
		for i := 1; i < len(args); i++ {
			t := a.typeFromValue(args[i]).apply(a.subst)
			if tc, ok := t.(*TypeConstant); ok && tc.Name == "float" {
				sawFloat = true
			}
			if s, err := unify(accType, t); err == nil {
				a.subst = a.subst.compose(s)
				accType = accType.apply(a.subst)
			} else {
				line := ctx.GetStart().GetLine()
				column := ctx.GetStart().GetColumn()
				a.errorReporter.ReportTypedError(line, column, fmt.Sprintf("[VEX-TYP-ARG]: argument %d type mismatch", i), TypeError)
				return NewBasicValue("call-result", "interface{}"), nil
			}
		}
		if sawFloat {
			return NewBasicValue("call-result", "number").WithType(&TypeConstant{Name: "float"}), nil
		}
		return NewBasicValue("call-result", "number").WithType(&TypeConstant{Name: "int"}), nil
	}
	// Unary minus support: (- x) => negate x
	if funcName == "-" && len(args) == 1 {
		t := a.typeFromValue(args[0]).apply(a.subst)
		// Enforce numeric
		if tc, ok := t.(*TypeConstant); ok {
			if tc.Name == "float" {
				return NewBasicValue("call-result", "number").WithType(&TypeConstant{Name: "float"}), nil
			}
			if tc.Name == "int" || tc.Name == "number" {
				return NewBasicValue("call-result", "number").WithType(&TypeConstant{Name: "int"}), nil
			}
		}
		// Fallback to number
		return NewBasicValue("call-result", "number").WithType(&TypeConstant{Name: "number"}), nil
	}

	// Precise HM for collection helpers using local unification, then fall back to schemes
	switch funcName {
	case "cons":
		if len(args) != 2 {
			goto SCHEME
		}
		elemT := a.typeFromValue(args[0]).apply(a.subst)
		arrT := a.typeFromValue(args[1]).apply(a.subst)
		// unify arrT with [a]
		aVar := a.freshTypeVar()
		s1, err1 := unify(arrT, &TypeArray{Elem: aVar})
		if err1 == nil {
			a.subst = a.subst.compose(s1)
		}
		// unify elemT with aVar (or elem of arr)
		s2, err2 := unify(elemT.apply(a.subst), aVar)
		if err2 == nil {
			a.subst = a.subst.compose(s2)
			rt := (&TypeArray{Elem: aVar}).apply(a.subst)
			return NewBasicValue("call-result", a.publicTypeString(rt)).WithType(rt), nil
		}
		goto SCHEME
	case "rest":
		if len(args) != 1 {
			goto SCHEME
		}
		arrT := a.typeFromValue(args[0]).apply(a.subst)
		aVar := a.freshTypeVar()
		if s, err := unify(arrT, &TypeArray{Elem: aVar}); err == nil {
			a.subst = a.subst.compose(s)
			rt := (&TypeArray{Elem: aVar}).apply(a.subst)
			return NewBasicValue("call-result", a.publicTypeString(rt)).WithType(rt), nil
		}
		goto SCHEME
	case "first":
		if len(args) != 1 {
			goto SCHEME
		}
		arrT := a.typeFromValue(args[0]).apply(a.subst)
		aVar := a.freshTypeVar()
		if s, err := unify(arrT, &TypeArray{Elem: aVar}); err == nil {
			a.subst = a.subst.compose(s)
			rt := aVar.apply(a.subst)
			return NewBasicValue("call-result", a.publicTypeString(rt)).WithType(rt), nil
		}
		goto SCHEME
	case "count":
		if len(args) != 1 {
			goto SCHEME
		}
		arrT := a.typeFromValue(args[0]).apply(a.subst)
		aVar := a.freshTypeVar()
		if s, err := unify(arrT, &TypeArray{Elem: aVar}); err == nil {
			a.subst = a.subst.compose(s)
			rt := (&TypeConstant{Name: "int"}).apply(a.subst)
			return NewBasicValue("call-result", a.publicTypeString(rt)).WithType(rt), nil
		}
		goto SCHEME
	case "empty?":
		if len(args) != 1 {
			goto SCHEME
		}
		arrT := a.typeFromValue(args[0]).apply(a.subst)
		aVar := a.freshTypeVar()
		if s, err := unify(arrT, &TypeArray{Elem: aVar}); err == nil {
			a.subst = a.subst.compose(s)
			rt := (&TypeConstant{Name: "bool"}).apply(a.subst)
			return NewBasicValue("call-result", a.publicTypeString(rt)).WithType(rt), nil
		}
		goto SCHEME
	case "get":
		if len(args) != 2 {
			goto SCHEME
		}
		arrT := a.typeFromValue(args[0]).apply(a.subst)
		idxT := a.typeFromValue(args[1]).apply(a.subst)
		// index must be int
		if sidx, err := unify(idxT, &TypeConstant{Name: "int"}); err == nil {
			a.subst = a.subst.compose(sidx)
		} else {
			line := ctx.GetStart().GetLine()
			column := ctx.GetStart().GetColumn()
			diag := diagnostics.New(diagnostics.CodeTypeIndex, diagnostics.SeverityError, "", line, column, nil)
			a.errorReporter.ReportDiagnosticBody(line, column, diag.RenderBody(), TypeError)
		}
		elemVar := a.freshTypeVar()
		if s, err := unify(arrT, &TypeArray{Elem: elemVar}); err == nil {
			a.subst = a.subst.compose(s)
			rt := elemVar.apply(a.subst)
			return NewBasicValue("call-result", a.publicTypeString(rt)).WithType(rt), nil
		}
		goto SCHEME
	case "slice":
		if len(args) != 2 {
			goto SCHEME
		}
		arrT := a.typeFromValue(args[0]).apply(a.subst)
		idxT := a.typeFromValue(args[1]).apply(a.subst)
		// index must be int
		if sidx, err := unify(idxT, &TypeConstant{Name: "int"}); err == nil {
			a.subst = a.subst.compose(sidx)
		} else {
			line := ctx.GetStart().GetLine()
			column := ctx.GetStart().GetColumn()
			diag := diagnostics.New(diagnostics.CodeTypeIndex, diagnostics.SeverityError, "", line, column, nil)
			a.errorReporter.ReportDiagnosticBody(line, column, diag.RenderBody(), TypeError)
		}
		elemVar := a.freshTypeVar()
		if s, err := unify(arrT, &TypeArray{Elem: elemVar}); err == nil {
			a.subst = a.subst.compose(s)
			rt := (&TypeArray{Elem: elemVar}).apply(a.subst)
			return NewBasicValue("call-result", a.publicTypeString(rt)).WithType(rt), nil
		}
		goto SCHEME
	case "append":
		if len(args) != 2 {
			goto SCHEME
		}
		arrT := a.typeFromValue(args[0]).apply(a.subst)
		elsT := a.typeFromValue(args[1]).apply(a.subst)
		elemVar := a.freshTypeVar()
		// unify first arg as [a]
		s1, err1 := unify(arrT, &TypeArray{Elem: elemVar})
		if err1 == nil {
			a.subst = a.subst.compose(s1)
		}
		// unify second arg as [a]
		if s2, err2 := unify(elsT.apply(a.subst), &TypeArray{Elem: elemVar}); err2 == nil {
			a.subst = a.subst.compose(s2)
			rt := (&TypeArray{Elem: elemVar}).apply(a.subst)
			return NewBasicValue("call-result", a.publicTypeString(rt)).WithType(rt), nil
		}
		goto SCHEME
	}

SCHEME:
	// Check if function is defined
	if _, exists := a.symbolTable.Lookup(funcName); !exists {
		// Allow builtins or namespaced calls (fmt/Println) via isBuiltinFunction
		if !isBuiltinFunction(funcName) {
			line := ctx.GetStart().GetLine()
			column := ctx.GetStart().GetColumn()
			diag := diagnostics.New(diagnostics.CodeTypeUndefined, diagnostics.SeverityError, "", line, column, map[string]any{"Name": funcName})
			a.errorReporter.ReportDiagnosticBody(line, column, diag.RenderBody(), TypeError)
			return NewBasicValue("call-result", "interface{}"), fmt.Errorf("undefined function: %s", funcName)
		}
	}

	// Attempt function typing via environment/symbols
	var ft Type
	if sch, ok := a.typeEnv[funcName]; ok {
		ft = instantiate(sch, func() int { a.freshId++; return a.freshId })
	} else if sym, ok := a.symbolTable.GetSymbol(funcName); ok {
		if bv, okb := sym.Value.(*BasicValue); okb && bv.getType() != nil {
			ft = bv.getType()
		}
	}
	if fn, ok := ft.(*TypeFunction); ok {
		if len(fn.Params) != len(args) {
			line := ctx.GetStart().GetLine()
			column := ctx.GetStart().GetColumn()
			a.errorReporter.ReportTypedError(line, column, fmt.Sprintf("[VEX-ARI-ARGS]: %s expects %d arguments; got %d", funcName, len(fn.Params), len(args)), TypeError)
			return NewBasicValue("call-result", "interface{}"), nil
		}
		subst := a.subst
		for i := range fn.Params {
			pt := fn.Params[i].apply(subst)
			at := a.typeFromValue(args[i]).apply(subst)
			s, err := unify(pt, at)
			if err != nil {
				// Allow collection helpers to pass to codegen-level routing
				if isCollectionHelper(funcName) {
					continue
				}
				line := ctx.GetStart().GetLine()
				column := ctx.GetStart().GetColumn()
				a.errorReporter.ReportTypedError(line, column, fmt.Sprintf("[VEX-TYP-ARG]: argument %d type mismatch", i), TypeError)
				return NewBasicValue("call-result", "interface{}"), nil
			}
			subst = subst.compose(s)
		}
		a.subst = subst
		rt := fn.Result.apply(a.subst)
		return NewBasicValue("call-result", a.publicTypeString(rt)).WithType(rt), nil
	}
	// If we have a non-function type (likely a type variable, e.g., parameter 'f'),
	// synthesize a function type and unify with it
	if ft != nil {
		paramTVs := make([]Type, 0, len(args))
		for range args {
			paramTVs = append(paramTVs, a.freshTypeVar())
		}
		resultTV := a.freshTypeVar()
		fnSynth := &TypeFunction{Params: paramTVs, Result: resultTV}
		subst := a.subst
		if s, err := unify(ft.apply(subst), fnSynth); err == nil {
			subst = subst.compose(s)
			// unify params with args
			for i := range args {
				pt := fnSynth.Params[i].apply(subst)
				at := a.typeFromValue(args[i]).apply(subst)
				si, err := unify(pt, at)
				if err != nil {
					line := ctx.GetStart().GetLine()
					column := ctx.GetStart().GetColumn()
					a.errorReporter.ReportTypedError(line, column, fmt.Sprintf("[VEX-TYP-ARG]: argument %d type mismatch", i), TypeError)
					return NewBasicValue("call-result", "interface{}"), nil
				}
				subst = subst.compose(si)
			}
			a.subst = subst
			rt := fnSynth.Result.apply(a.subst)
			return NewBasicValue("call-result", a.publicTypeString(rt)).WithType(rt), nil
		}
	}

	// If it's a recognized builtin/external op, accept with unknown result type
	if isBuiltinFunction(funcName) {
		// For operators backed by schemes, attempt to get scheme and unify arity accordingly
		if sch, ok := a.typeEnv[funcName]; ok {
			ft = instantiate(sch, func() int { a.freshId++; return a.freshId })
			if fn, ok := ft.(*TypeFunction); ok {
				if len(fn.Params) != len(args) {
					// For '+' handled earlier; others must match arity
					line := ctx.GetStart().GetLine()
					column := ctx.GetStart().GetColumn()
					diag := diagnostics.New(diagnostics.CodeAriArgs, diagnostics.SeverityError, "", line, column, map[string]any{"Op": funcName, "Expected": len(fn.Params), "Got": len(args)})
					a.errorReporter.ReportDiagnosticBody(line, column, diag.RenderBody(), TypeError)
					return NewBasicValue("call-result", "interface{}"), nil
				}
				subst := a.subst
				for i := range fn.Params {
					pt := fn.Params[i].apply(subst)
					at := a.typeFromValue(args[i]).apply(subst)
					s, err := unify(pt, at)
					if err != nil {
						// Be tolerant for collection helpers to allow nested chaining
						if isCollectionHelper(funcName) {
							continue
						}
						line := ctx.GetStart().GetLine()
						column := ctx.GetStart().GetColumn()
						a.errorReporter.ReportTypedError(line, column, fmt.Sprintf("[VEX-TYP-ARG]: argument %d type mismatch", i), TypeError)
						return NewBasicValue("call-result", "interface{}"), nil
					}
					subst = subst.compose(s)
				}
				a.subst = subst
				rt := fn.Result.apply(a.subst)
				return NewBasicValue("call-result", a.publicTypeString(rt)).WithType(rt), nil
			}
		}
		return NewBasicValue("call-result", "interface{}"), nil
	}
	// Unknown function: strict HM requires a definition or scheme
	line := ctx.GetStart().GetLine()
	column := ctx.GetStart().GetColumn()
	diag := diagnostics.New(diagnostics.CodeTypeUndefined, diagnostics.SeverityError, "", line, column, map[string]any{"Name": funcName})
	a.errorReporter.ReportDiagnosticBody(line, column, diag.RenderBody(), TypeError)
	return NewBasicValue("call-result", "interface{}"), fmt.Errorf("undefined function: %s", funcName)

}

// Helper methods
func (a *AnalyzerImpl) visitNode(node antlr.Tree) (Value, error) {
	switch n := node.(type) {
	case *parser.ListContext:
		return a.VisitList(n)
	case *parser.ArrayContext:
		return a.VisitArray(n)
	case antlr.TerminalNode:
		return a.VisitTerminal(n)
	default:
		// Unknown node kind in analysis is an internal condition; treat as undefined
		return NewBasicValue("unknown", "undefined"), fmt.Errorf("unknown node in analysis")
	}
}

func (a *AnalyzerImpl) nodeToString(node antlr.Tree) string {
	if terminal, ok := node.(antlr.TerminalNode); ok {
		return terminal.GetText()
	}
	return "unknown"
}

// HM helpers
func (a *AnalyzerImpl) freshTypeVar() *TypeVariable {
	a.freshId++
	return &TypeVariable{ID: a.freshId}
}

func (a *AnalyzerImpl) typeFromValue(v Value) Type {
	if bv, ok := v.(*BasicValue); ok && bv.getType() != nil {
		return bv.getType()
	}
	switch v.Type() {
	case "string", "bool", "number", "record", "func":
		return &TypeConstant{Name: v.Type()}
	case "[]interface{}":
		return &TypeArray{Elem: a.freshTypeVar()}
	case "map[interface{}]interface{}":
		return &TypeConstant{Name: v.Type()}
	case "symbol":
		// Undefined/bare symbols are treated as fresh type variables for unification
		return a.freshTypeVar()
	default:
		return &TypeConstant{Name: v.Type()}
	}
}

func (a *AnalyzerImpl) publicTypeString(t Type) string {
	switch tt := t.(type) {
	case *TypeConstant:
		switch tt.Name {
		case "string", "bool", "number":
			return tt.Name
		case "int", "float":
			return "number"
		default:
			// Treat unknown constants as nominal records for outward typing
			return "record"
		}
	case *TypeArray:
		return "[]interface{}"
	case *TypeMap:
		return "map[interface{}]interface{}"
	case *TypeFunction:
		return "func"
	default:
		return "interface{}"
	}
}

// typeNameToType maps outward type string tokens to internal types
func (a *AnalyzerImpl) typeNameToType(name string) Type {
	switch name {
	case "int", "float":
		return &TypeConstant{Name: name}
	case "number":
		return &TypeConstant{Name: "number"}
	case "string", "bool", "record", "func":
		return &TypeConstant{Name: name}
	default:
		// Record names map to record for now; future: distinct nominal types
		return &TypeConstant{Name: name}
	}
}

// parseParameterListWithTypes parses parameter list with required type annotations
// Only supports explicit types: "x: int y: string z: bool" syntax
func (a *AnalyzerImpl) parseParameterListWithTypes(paramListInner string) []Type {
	paramTypes := make([]Type, 0)

	// Split by whitespace but handle type annotations
	tokens := strings.Fields(paramListInner)

	for i := 0; i < len(tokens); {
		paramToken := tokens[i]

		// Check if this token ends with ":" indicating type annotation follows
		if strings.HasSuffix(paramToken, ":") && i+1 < len(tokens) {
			// param: type format - "x:" followed by "int"
			paramName := strings.TrimSuffix(paramToken, ":")
			typeStr := tokens[i+1]
			paramType := a.typeNameToType(typeStr)
			_ = a.symbolTable.Define(paramName, NewBasicValue(paramName, "interface{}").WithType(paramType))
			paramTypes = append(paramTypes, paramType)
			i += 2 // Skip param: and type
		} else if strings.Contains(paramToken, ":") {
			// param:type format (no space) - "x:int"
			parts := strings.Split(paramToken, ":")
			if len(parts) == 2 && parts[1] != "" {
				paramName := parts[0]
				typeStr := parts[1]
				paramType := a.typeNameToType(typeStr)
				_ = a.symbolTable.Define(paramName, NewBasicValue(paramName, "interface{}").WithType(paramType))
				paramTypes = append(paramTypes, paramType)
				i += 1 // Skip this combined token
			} else {
				return nil // Malformed type annotation
			}
		} else {
			// param without type annotation - ERROR in explicit-only mode
			return nil
		}
	}

	return paramTypes
}

// parseSingleExpression parses a Vex expression from a raw string into a parse tree node
func (a *AnalyzerImpl) parseSingleExpression(expr string) antlr.Tree {
	input := antlr.NewInputStream(expr)
	lexer := parser.NewVexLexer(input)
	tokenStream := antlr.NewCommonTokenStream(lexer, 0)
	p := parser.NewVexParser(tokenStream)
	// Try list first, then array, then program
	if l := p.List(); l != nil && l.GetChildCount() > 0 {
		return l
	}
	if arr := p.Array(); arr != nil && arr.GetChildCount() > 0 {
		return arr
	}
	if prog := p.Program(); prog != nil && prog.GetChildCount() > 0 {
		return prog
	}
	return nil
}

// nodeToFlatText reconstructs a flat text for arrays, lists, and terminals
func (a *AnalyzerImpl) nodeToFlatText(node antlr.Tree) string {
	switch n := node.(type) {
	case antlr.TerminalNode:
		return n.GetText()
	case *parser.ArrayContext:
		var b strings.Builder
		b.WriteString("[")
		first := true
		for i := 1; i < n.GetChildCount()-1; i++ {
			child := n.GetChild(i)
			if child == nil {
				continue
			}
			if !first {
				b.WriteString(" ")
			} else {
				first = false
			}
			b.WriteString(a.nodeToFlatText(child))
		}
		b.WriteString("]")
		return b.String()
	case *parser.ListContext:
		var b strings.Builder
		b.WriteString("(")
		first := true
		for i := 1; i < n.GetChildCount()-1; i++ {
			child := n.GetChild(i)
			if child == nil {
				continue
			}
			if !first {
				b.WriteString(" ")
			} else {
				first = false
			}
			b.WriteString(a.nodeToFlatText(child))
		}
		b.WriteString(")")
		return b.String()
	default:
		var b strings.Builder
		for i := 0; i < node.GetChildCount(); i++ {
			if i > 0 {
				b.WriteString(" ")
			}
			b.WriteString(a.nodeToFlatText(node.GetChild(i)))
		}
		return b.String()
	}
}

// Helper functions
func isNumber(s string) bool {
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return len(s) > 0
}

// Enhanced numeric checks for internal typing
func isInt(s string) bool {
	if s == "" {
		return false
	}
	if s[0] == '-' || s[0] == '+' {
		s = s[1:]
	}
	if s == "" {
		return false
	}
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

func isFloat(s string) bool {
	if s == "" {
		return false
	}
	if s[0] == '-' || s[0] == '+' {
		s = s[1:]
	}
	if s == "" {
		return false
	}
	dot := false
	digitsBefore := 0
	digitsAfter := 0
	for _, r := range s {
		if r == '.' {
			if dot {
				return false
			}
			dot = true
			continue
		}
		if r < '0' || r > '9' {
			return false
		}
		if !dot {
			digitsBefore++
		} else {
			digitsAfter++
		}
	}
	return dot && digitsBefore > 0 && digitsAfter > 0
}

// (numeric/equality helpers removed; rely on schemes and unification)

// isUnknownType allows interface{} and undefined to pass early checks in analyzer
func isUnknownType(t string) bool {
	return t == "interface{}" || t == "undefined"
}

// --- Record support ---

// analyzeRecord registers a record type declaration: (record Name [field: Type ...])
func (a *AnalyzerImpl) analyzeRecord(ctx *parser.ListContext, args []Value) (Value, error) {
	// Expect: ( record Name [fields...] )
	if ctx.GetChildCount() < 4 {
		line := ctx.GetStart().GetLine()
		column := ctx.GetStart().GetColumn()
		diag := diagnostics.New(diagnostics.CodeRecArgs, diagnostics.SeverityError, "", line, column, nil).WithSuggestion("use (record Name [field: Type ...])")
		a.errorReporter.ReportDiagnosticBody(line, column, diag.RenderBody(), SemanticError)
		return nil, fmt.Errorf("invalid record")
	}
	nameNode := ctx.GetChild(2)
	name := a.nodeToString(nameNode)
	if name == "" || isReservedWord(name) {
		line := ctx.GetStart().GetLine()
		column := ctx.GetStart().GetColumn()
		diag := diagnostics.New(diagnostics.CodeRecordName, diagnostics.SeverityError, "", line, column, nil)
		a.errorReporter.ReportDiagnosticBody(line, column, diag.RenderBody(), SemanticError)
		return nil, fmt.Errorf("invalid record name")
	}

	// Validate record naming convention
	if !isValidSymbolName(name) {
		line := ctx.GetStart().GetLine()
		column := ctx.GetStart().GetColumn()
		diag := diagnostics.New(diagnostics.CodeSymNaming, diagnostics.SeverityError, "", line, column, map[string]any{"Name": name}).WithSuggestion("use kebab-case with dashes (e.g., 'my-record')")
		a.errorReporter.ReportDiagnosticBody(line, column, diag.RenderBody(), SemanticError)
		return nil, fmt.Errorf("invalid record name: %s", name)
	}
	var fields map[string]string
	var order []string
	if arr, ok := ctx.GetChild(3).(*parser.ArrayContext); ok {
		fm, ord, ferr := a.parseRecordFields(arr)
		if ferr != nil {
			line := ctx.GetStart().GetLine()
			column := ctx.GetStart().GetColumn()
			a.errorReporter.ReportError(line, column, ferr.Error())
			return nil, ferr
		}
		fields = fm
		order = ord
	} else {
		line := ctx.GetStart().GetLine()
		column := ctx.GetStart().GetColumn()
		diag := diagnostics.New(diagnostics.CodeRecordFields, diagnostics.SeverityError, "", line, column, nil)
		a.errorReporter.ReportDiagnosticBody(line, column, diag.RenderBody(), SemanticError)
		return nil, fmt.Errorf("invalid record fields")
	}
	rec := NewRecordValue(name, fields, order)
	if err := a.symbolTable.Define(name, rec); err != nil {
		line := ctx.GetStart().GetLine()
		column := ctx.GetStart().GetColumn()
		a.errorReporter.ReportError(line, column, err.Error())
		return nil, err
	}
	return rec, nil
}

// (duplicate placeholder removed; concrete implementations below)

// parseRecordFields parses [name: Type ...] into a map
func (a *AnalyzerImpl) parseRecordFields(arr *parser.ArrayContext) (map[string]string, []string, error) {
	fields := make(map[string]string)
	order := make([]string, 0)
	tokens := make([]string, 0, arr.GetChildCount()-2)
	for i := 1; i < arr.GetChildCount()-1; i++ { // skip '[' and ']'
		child := arr.GetChild(i)
		if child == nil {
			continue
		}
		tokens = append(tokens, a.nodeToString(child))
	}
	for i := 0; i < len(tokens); {
		if i >= len(tokens) {
			break
		}
		nameTok := tokens[i]
		var typeTok string
		if strings.HasSuffix(nameTok, ":") {
			nameTok = strings.TrimSuffix(nameTok, ":")
			if i+1 >= len(tokens) {
				return nil, nil, fmt.Errorf("missing type for field '%s'", nameTok)
			}
			typeTok = tokens[i+1]
			i += 2
		} else {
			if i+2 >= len(tokens) {
				return nil, nil, fmt.Errorf("invalid field format; expected name : Type")
			}
			colonTok := tokens[i+1]
			if colonTok != ":" {
				return nil, nil, fmt.Errorf("invalid field format; missing ':' after %s", nameTok)
			}
			typeTok = tokens[i+2]
			i += 3
		}
		if nameTok == "" || isReservedWord(nameTok) {
			return nil, nil, fmt.Errorf("invalid field name '%s'", nameTok)
		}
		if !isValidSymbolName(nameTok) {
			return nil, nil, fmt.Errorf("field name '%s' must use kebab-case (e.g., 'my-field')", nameTok)
		}
		if typeTok == "" {
			return nil, nil, fmt.Errorf("missing type for field '%s'", nameTok)
		}
		if _, exists := fields[nameTok]; exists {
			return nil, nil, fmt.Errorf("duplicate field '%s'", nameTok)
		}
		fields[nameTok] = typeTok
		order = append(order, nameTok)
	}
	return fields, order, nil
}

// analyzeRecordConstructor validates a record construction list against the record's fields
// Syntax: (RecordName [key: value key: value ...])
func (a *AnalyzerImpl) analyzeRecordConstructor(ctx *parser.ListContext, rec *RecordValue) (Value, error) {
	if ctx.GetChildCount() < 4 {
		line := ctx.GetStart().GetLine()
		column := ctx.GetStart().GetColumn()
		diag := diagnostics.New(diagnostics.CodeRecordConstruct, diagnostics.SeverityError, "", line, column, nil)
		a.errorReporter.ReportDiagnosticBody(line, column, diag.RenderBody(), TypeError)
		return nil, fmt.Errorf("invalid record construction")
	}
	arr, ok := ctx.GetChild(2).(*parser.ArrayContext)
	if !ok {
		line := ctx.GetStart().GetLine()
		column := ctx.GetStart().GetColumn()
		diag := diagnostics.New(diagnostics.CodeRecordConstruct, diagnostics.SeverityError, "", line, column, nil).WithMessage("record construction expects [fields]")
		a.errorReporter.ReportDiagnosticBody(line, column, diag.RenderBody(), TypeError)
		return nil, fmt.Errorf("invalid record construction")
	}

	fieldTypes := rec.GetFields()
	// Walk children: name (with or without ':') ':' valueNode
	for i := 1; i < arr.GetChildCount()-1; {
		if i >= arr.GetChildCount()-1 {
			break
		}
		nameNode := arr.GetChild(i)
		if nameNode == nil {
			i++
			continue
		}
		nameTok := a.nodeToString(nameNode)
		var valNode antlr.Tree
		// Handle name: or name : value
		if strings.HasSuffix(nameTok, ":") {
			nameTok = strings.TrimSuffix(nameTok, ":")
			if i+1 >= arr.GetChildCount()-1 {
				break
			}
			valNode = arr.GetChild(i + 1)
			i += 2
		} else {
			if i+2 >= arr.GetChildCount()-1 {
				break
			}
			colon := a.nodeToString(arr.GetChild(i + 1))
			if colon != ":" {
				i++
				continue
			}
			valNode = arr.GetChild(i + 2)
			i += 3
		}
		expected, ok := fieldTypes[nameTok]
		if !ok {
			line := ctx.GetStart().GetLine()
			column := ctx.GetStart().GetColumn()
			a.errorReporter.ReportTypedError(line, column, fmt.Sprintf("unknown field '%s' for %s", nameTok, rec.String()), TypeError)
			continue
		}
		// Visit value node to obtain HM type
		val, err := a.visitNode(valNode)
		if err != nil {
			continue
		}
		valT := a.typeFromValue(val).apply(a.subst)
		expT := a.typeNameToType(expected)
		if s, err := unify(expT.apply(a.subst), valT); err == nil {
			a.subst = a.subst.compose(s)
		} else {
			line := ctx.GetStart().GetLine()
			column := ctx.GetStart().GetColumn()
			a.errorReporter.ReportTypedError(line, column, fmt.Sprintf("field '%s' expects %s, got %s", nameTok, expected, a.publicTypeString(valT)), TypeError)
		}
	}
	return NewBasicValue(rec.String(), rec.Type()).WithType(&TypeConstant{Name: rec.String()}), nil
}

// analyzeRecordFieldCall supports (User :field) calls equivalent to (get-field User :field)
func (a *AnalyzerImpl) analyzeRecordFieldCall(ctx *parser.ListContext, sym *Symbol) (Value, error) {
	// Expect: (User :field)
	if ctx.GetChildCount() < 4 { // '(', User, arg, ')'
		line := ctx.GetStart().GetLine()
		column := ctx.GetStart().GetColumn()
		diag := diagnostics.New(diagnostics.CodeRecordConstruct, diagnostics.SeverityError, "", line, column, nil).WithMessage("record field call requires one field argument")
		a.errorReporter.ReportDiagnosticBody(line, column, diag.RenderBody(), TypeError)
		return nil, fmt.Errorf("invalid record field call")
	}
	argNode := ctx.GetChild(2)
	fieldTok := a.nodeToString(argNode)
	fieldName := strings.TrimPrefix(fieldTok, ":")
	// Resolve record value
	rv, ok := sym.Value.(*RecordValue)
	if !ok {
		return NewBasicValue("field", "interface{}"), nil
	}
	if t, ok := rv.GetFields()[fieldName]; ok {
		// Attach internal type when possible
		return NewBasicValue("field", t).WithType(a.typeNameToType(t)), nil
	}
	line := ctx.GetStart().GetLine()
	column := ctx.GetStart().GetColumn()
	a.errorReporter.ReportTypedError(line, column, fmt.Sprintf("unknown field '%s' for %s", fieldName, rv.String()), TypeError)
	return NewBasicValue("field", "undefined"), nil
}

// analyzeMap processes (map [k v k v ...]) constructing a map type
func (a *AnalyzerImpl) analyzeMap(ctx *parser.ListContext, args []Value) (Value, error) {
	if ctx.GetChildCount() < 3 {
		line := ctx.GetStart().GetLine()
		column := ctx.GetStart().GetColumn()
		a.errorReporter.ReportDiagnosticBody(line, column, diagnostics.New(diagnostics.CodeMapArguments, diagnostics.SeverityError, "", line, column, nil).RenderBody(), SemanticError)
		return nil, fmt.Errorf("invalid map")
	}
	arr, ok := ctx.GetChild(2).(*parser.ArrayContext)
	if !ok {
		line := ctx.GetStart().GetLine()
		column := ctx.GetStart().GetColumn()
		a.errorReporter.ReportDiagnosticBody(line, column, diagnostics.New(diagnostics.CodeMapArguments, diagnostics.SeverityError, "", line, column, nil).WithMessage("map expects [key value ...]").RenderBody(), SemanticError)
		return nil, fmt.Errorf("invalid map")
	}
	// HM: visit child nodes and unify key/value types across pairs
	var keyT Type
	var valT Type
	pairIndex := 0
	for i := 1; i < arr.GetChildCount()-1; {
		if i >= arr.GetChildCount()-1 {
			break
		}
		kNode := arr.GetChild(i)
		if kNode == nil {
			break
		}
		vIdx := i + 1
		if vIdx >= arr.GetChildCount()-1 {
			break
		}
		vNode := arr.GetChild(vIdx)
		kVal, _ := a.visitNode(kNode)
		vVal, _ := a.visitNode(vNode)
		kt := a.typeFromValue(kVal).apply(a.subst)
		vt := a.typeFromValue(vVal).apply(a.subst)
		if keyT == nil {
			keyT = kt
		} else {
			if s, err := unify(keyT, kt); err == nil {
				a.subst = a.subst.compose(s)
				keyT = keyT.apply(a.subst)
			} else {
				line := ctx.GetStart().GetLine()
				column := ctx.GetStart().GetColumn()
				diag := diagnostics.New(diagnostics.CodeTypeMapKey, diagnostics.SeverityError, "", line, column, map[string]any{"Offender": fmt.Sprintf("pair %d", pairIndex)})
				a.errorReporter.ReportDiagnosticBody(line, column, diag.RenderBody(), TypeError)
			}
		}
		if valT == nil {
			valT = vt
		} else {
			if s, err := unify(valT, vt); err == nil {
				a.subst = a.subst.compose(s)
				valT = valT.apply(a.subst)
			} else {
				line := ctx.GetStart().GetLine()
				column := ctx.GetStart().GetColumn()
				diag := diagnostics.New(diagnostics.CodeTypeMapValue, diagnostics.SeverityError, "", line, column, map[string]any{"Offender": fmt.Sprintf("pair %d", pairIndex)})
				a.errorReporter.ReportDiagnosticBody(line, column, diag.RenderBody(), TypeError)
			}
		}
		pairIndex++
		i += 2
	}
	if keyT == nil {
		keyT = a.freshTypeVar()
	}
	if valT == nil {
		valT = a.freshTypeVar()
	}
	tm := &TypeMap{Key: keyT, Val: valT}
	return NewBasicValue("map", "map[interface{}]interface{}").WithType(tm), nil
}

func inferPrimitiveTypeString(tok string) string {
	if strings.HasPrefix(tok, "\"") && strings.HasSuffix(tok, "\"") {
		return "string"
	}
	if isFloat(tok) || isInt(tok) {
		return "number"
	}
	if tok == "true" || tok == "false" {
		return "bool"
	}
	return "interface{}"
}

func isReservedWord(word string) bool {
	reserved := []string{"if", "def", "fn", "let", "do", "when", "unless", "import"}
	for _, r := range reserved {
		if word == r {
			return true
		}
	}
	return false
}

func isBuiltinFunction(name string) bool {
	builtins := []string{
		"+", "-", "*", "/", ">", "<", "=", "not", "and", "or", "len",
		// Collections kept for codegen routing; typing comes from schemes above
		"first", "rest", "cons", "count", "empty?", "len",
	}
	for _, b := range builtins {
		if name == b {
			return true
		}
	}
	return strings.Contains(name, "/") // Package functions like fmt/Println
}

// isCollectionHelper retained for future use; strict HM no longer tolerates mismatches
func isCollectionHelper(name string) bool {
	switch name {
	case "first", "rest", "cons", "count", "empty?", "len", "get", "slice", "append":
		return true
	default:
		return false
	}
}

// isValidSymbolName validates that symbol names use kebab-case (dashes, not underscores)
func isValidSymbolName(name string) bool {
	// Skip validation for reserved words, builtin functions, and package functions
	if isReservedWord(name) || isBuiltinFunction(name) || strings.Contains(name, "/") {
		return true
	}

	// Symbol names should not contain underscores (use dashes instead)
	return !strings.Contains(name, "_")
}

// Interface implementations to satisfy the imports
type AST interface {
	Accept(visitor ASTVisitor) error
}

type ASTVisitor interface {
	VisitProgram(ctx *parser.ProgramContext) error
	VisitList(ctx *parser.ListContext) (Value, error)
	VisitArray(ctx *parser.ArrayContext) (Value, error)
	VisitTerminal(node antlr.TerminalNode) (Value, error)
}

type SymbolTable interface {
	Define(name string, value Value) error
	Lookup(name string) (Value, bool)
	EnterScope()
	ExitScope()
}

type ErrorReporter interface {
	ReportError(line, column int, message string)
	ReportWarning(line, column int, message string)
	HasErrors() bool
	GetErrors() []CompilerError
}
