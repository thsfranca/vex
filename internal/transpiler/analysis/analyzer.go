package analysis

import (
	"fmt"
	"strings"

	"github.com/antlr4-go/antlr/v4"
	"github.com/thsfranca/vex/internal/transpiler/parser"
)

// AnalyzerImpl implements the Analyzer interface
type AnalyzerImpl struct {
	symbolTable   *SymbolTableImpl
	errorReporter *ErrorReporterImpl
}

// NewAnalyzer creates a new analyzer
func NewAnalyzer() *AnalyzerImpl {
	return &AnalyzerImpl{
		symbolTable:   NewSymbolTable(),
		errorReporter: NewErrorReporter(),
	}
}

// Analyze performs semantic analysis on the AST
func (a *AnalyzerImpl) Analyze(ast AST) (SymbolTable, error) {
	// Reset for new analysis
	a.symbolTable = NewSymbolTable()
	a.errorReporter.Clear()
	
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
		a.errorReporter.ReportError(line, column, "empty expression")
		return nil, fmt.Errorf("empty expression")
	}

	// Get function name (first child after '(')
	funcNameNode := ctx.GetChild(1)
	funcName := a.nodeToString(funcNameNode)

	// Extract arguments
	var args []Value
	for i := 2; i < childCount-1; i++ { // Skip '(' and ')'
		child := ctx.GetChild(i)
		if child != nil {
			// For special forms that need raw arguments, handle specially
			if funcName == "macro" || funcName == "def" {
				// For these special forms, just get the text representation
				argText := a.nodeToString(child)
				args = append(args, NewBasicValue(argText, "raw"))
			} else {
				arg, err := a.visitNode(child)
				if err != nil {
					continue // Error already reported
				}
				args = append(args, arg)
			}
		}
	}

	// Analyze special forms
	switch funcName {
	case "def":
		return a.analyzeDef(ctx, args)
	case "if":
		return a.analyzeIf(ctx, args)
	case "fn":
		return a.analyzeFn(ctx, args)
	case "macro":
		return a.analyzeMacro(ctx, args)
	default:
		return a.analyzeFunctionCall(ctx, funcName, args)
	}
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
	
	// Create array value
	return NewBasicValue("array", "[]interface{}"), nil
}

// VisitTerminal analyzes a terminal node
func (a *AnalyzerImpl) VisitTerminal(node antlr.TerminalNode) (Value, error) {
	text := node.GetText()
	
	// Determine type based on content
	if strings.HasPrefix(text, "\"") && strings.HasSuffix(text, "\"") {
		return NewBasicValue(text, "string"), nil
	}
	
	// Check if it's a number
	if isNumber(text) {
		return NewBasicValue(text, "number"), nil
	}
	
	// Check if it's a boolean
	if text == "true" || text == "false" {
		return NewBasicValue(text, "bool"), nil
	}
	
	// Must be a symbol - check if it's defined
	if value, exists := a.symbolTable.Lookup(text); exists {
		return value, nil
	}
	
	// Undefined symbol
	line := 1
	column := 0
	if symbol := node.GetSymbol(); symbol != nil {
		line = symbol.GetLine()
		column = symbol.GetColumn()
	}
	a.errorReporter.ReportError(line, column, fmt.Sprintf("undefined symbol '%s'", text))
	
	return NewBasicValue(text, "undefined"), fmt.Errorf("undefined symbol")
}

// analyzeDef analyzes a definition
func (a *AnalyzerImpl) analyzeDef(ctx *parser.ListContext, args []Value) (Value, error) {
	if len(args) < 2 {
		line := ctx.GetStart().GetLine()
		column := ctx.GetStart().GetColumn()
		a.errorReporter.ReportError(line, column, "def requires name and value")
		return nil, fmt.Errorf("invalid def")
	}
	
	name := args[0].String()
	value := args[1]
	
	// Define the symbol
	if err := a.symbolTable.Define(name, value); err != nil {
		line := ctx.GetStart().GetLine()
		column := ctx.GetStart().GetColumn()
		a.errorReporter.ReportError(line, column, err.Error())
		return nil, err
	}
	
	return value, nil
}

// analyzeIf analyzes an if expression
func (a *AnalyzerImpl) analyzeIf(ctx *parser.ListContext, args []Value) (Value, error) {
	if len(args) < 2 {
		line := ctx.GetStart().GetLine()
		column := ctx.GetStart().GetColumn()
		a.errorReporter.ReportError(line, column, "if requires condition and then-branch")
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
	if condition.Type() != "bool" && condition.Type() != "number" {
		line := ctx.GetStart().GetLine()
		column := ctx.GetStart().GetColumn()
		a.errorReporter.ReportWarning(line, column, 
			fmt.Sprintf("condition has type '%s', expected bool", condition.Type()))
	}
	
	// Result type is the type of the then branch
	// (In a more sophisticated system, we'd unify then and else types)
	resultType := thenBranch.Type()
	if elseBranch != nil && elseBranch.Type() != resultType {
		resultType = "interface{}" // Mixed types
	}
	
	return NewBasicValue("if-result", resultType), nil
}

// analyzeFn analyzes a function definition
func (a *AnalyzerImpl) analyzeFn(ctx *parser.ListContext, args []Value) (Value, error) {
	if len(args) < 2 {
		line := ctx.GetStart().GetLine()
		column := ctx.GetStart().GetColumn()
		a.errorReporter.ReportError(line, column, "fn requires parameter list and body")
		return nil, fmt.Errorf("invalid fn")
	}
	
	// Enter new scope for function parameters
	a.symbolTable.EnterScope()
	defer a.symbolTable.ExitScope()
	
	// Analyze parameter list (args[0])
	// For now, we'll treat parameters as symbols with interface{} type
	paramList := args[0].String()
	if strings.HasPrefix(paramList, "[") && strings.HasSuffix(paramList, "]") {
		// Extract parameter names
		inner := strings.TrimSpace(paramList[1 : len(paramList)-1])
		if inner != "" {
			params := strings.Fields(inner)
			for _, param := range params {
				paramValue := NewBasicValue(param, "interface{}")
				a.symbolTable.Define(param, paramValue)
			}
		}
	}
	
	// Analyze function body (args[1])
	_ = args[1] // bodyValue not used in this simplified implementation
	
	return NewBasicValue("function", "func"), nil
}

// analyzeMacro analyzes a macro definition
func (a *AnalyzerImpl) analyzeMacro(ctx *parser.ListContext, args []Value) (Value, error) {
	if len(args) < 3 {
		line := ctx.GetStart().GetLine()
		column := ctx.GetStart().GetColumn()
		a.errorReporter.ReportError(line, column, "macro requires name, parameters, and body")
		return nil, fmt.Errorf("invalid macro")
	}
	
	name := args[0].String()
	
	// Check for valid macro name
	if isReservedWord(name) {
		line := ctx.GetStart().GetLine()
		column := ctx.GetStart().GetColumn()
		a.errorReporter.ReportError(line, column, 
			fmt.Sprintf("'%s' is a reserved word and cannot be used as macro name", name))
		return nil, fmt.Errorf("invalid macro name")
	}
	
	// Define the macro as a symbol
	macroValue := NewBasicValue(name, "macro")
	a.symbolTable.Define(name, macroValue)
	
	return macroValue, nil
}

// analyzeFunctionCall analyzes a function call
func (a *AnalyzerImpl) analyzeFunctionCall(ctx *parser.ListContext, funcName string, args []Value) (Value, error) {
	// Check if function is defined
	if _, exists := a.symbolTable.Lookup(funcName); !exists {
		// For built-in functions, we'll allow them
		if !isBuiltinFunction(funcName) {
			line := ctx.GetStart().GetLine()
			column := ctx.GetStart().GetColumn()
			a.errorReporter.ReportWarning(line, column, 
				fmt.Sprintf("function '%s' may not be defined", funcName))
		}
	}
	
	// For now, assume function calls return interface{}
	return NewBasicValue("call-result", "interface{}"), nil
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
		return NewBasicValue("unknown", "interface{}"), nil
	}
}

func (a *AnalyzerImpl) nodeToString(node antlr.Tree) string {
	if terminal, ok := node.(antlr.TerminalNode); ok {
		return terminal.GetText()
	}
	return "unknown"
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
	builtins := []string{"+", "-", "*", "/", ">", "<", "=", "not", "and", "or", 
		"first", "rest", "cons", "count", "empty?", "println", "printf"}
	for _, b := range builtins {
		if name == b {
			return true
		}
	}
	return strings.Contains(name, "/") // Package functions like fmt/Println
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

