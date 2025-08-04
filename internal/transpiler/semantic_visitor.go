// Package transpiler provides semantic analysis with type checking
package transpiler

import (
	"fmt"
	"strings"

	"github.com/antlr4-go/antlr/v4"
	"github.com/thsfranca/vex/internal/transpiler/parser"
)

// SemanticVisitor implements semantic analysis with type checking
type SemanticVisitor struct {
	parser.BaseVexVisitor
	codeGen          *CodeGenerator
	macroRegistry    *MacroRegistry
	namespaceManager *NamespaceManager
	typeChecker      *TypeChecker
	typeInference    *TypeInference
	typeUtils        *TypeUtils
	errors           []string
	currentNode      antlr.Tree // Track current node for error reporting
}

// NewSemanticVisitor creates a new semantic visitor with type system integration
func NewSemanticVisitor() *SemanticVisitor {
	namespaceManager := NewNamespaceManager()
	typeChecker := NewTypeChecker(namespaceManager)
	
	return &SemanticVisitor{
		codeGen:          NewCodeGenerator(),
		macroRegistry:    NewMacroRegistry(),
		namespaceManager: namespaceManager,
		typeChecker:      typeChecker,
		typeInference:    NewTypeInference(namespaceManager),
		typeUtils:        &TypeUtils{},
		errors:           make([]string, 0, 16),
	}
}

// AnalyzeProgram performs complete semantic analysis on a program
func (sv *SemanticVisitor) AnalyzeProgram(programCtx *parser.ProgramContext) error {
	sv.errors = sv.errors[:0] // Reset errors
	
	// Step 1: Perform type checking
	typeErrors, err := sv.typeChecker.CheckProgram(programCtx)
	if err != nil {
		return fmt.Errorf("type checking failed: %w", err)
	}
	
	// Convert type errors to string errors
	for _, typeErr := range typeErrors {
		sv.errors = append(sv.errors, typeErr.String())
	}
	
	// Step 2: If type checking passed, generate code with type information
	if len(typeErrors) == 0 {
		programCtx.Accept(sv)
	}
	
	return nil
}

// GetGeneratedCode returns the generated Go code
func (sv *SemanticVisitor) GetGeneratedCode() string {
	return sv.codeGen.GetCode()
}

// GetCodeGenerator returns the code generator
func (sv *SemanticVisitor) GetCodeGenerator() *CodeGenerator {
	return sv.codeGen
}

// GetErrors returns all semantic analysis errors
func (sv *SemanticVisitor) GetErrors() []string {
	return sv.errors
}

// HasErrors returns true if there are semantic errors
func (sv *SemanticVisitor) HasErrors() bool {
	return len(sv.errors) > 0
}

// GetNamespaceManager returns the namespace manager
func (sv *SemanticVisitor) GetNamespaceManager() *NamespaceManager {
	return sv.namespaceManager
}

// VisitProgram visits the root node with semantic analysis
func (sv *SemanticVisitor) VisitProgram(ctx *parser.ProgramContext) interface{} {
	// Visit all child lists with type information
	for _, listCtx := range ctx.AllList() {
		sv.visitWithTypeInfo(listCtx)
	}
	return nil
}

// visitWithTypeInfo visits a node and infers its type
func (sv *SemanticVisitor) visitWithTypeInfo(node antlr.Tree) VexType {
	// Infer the type of the node
	inferredType, err := sv.typeInference.InferTypes(node)
	if err != nil {
		sv.addError(fmt.Sprintf("Type inference failed: %s", err.Error()))
		inferredType = NewUnknownType(0)
	}
	
	// Visit the node for code generation
	switch ctx := node.(type) {
	case parser.IListContext:
		sv.VisitList(ctx.(*parser.ListContext))
	case parser.IArrayContext:
		sv.VisitArray(ctx.(*parser.ArrayContext))
	case *antlr.TerminalNodeImpl:
		// Terminal nodes don't need visiting
	}
	
	return inferredType
}

// VisitList visits a list expression with type-aware code generation
func (sv *SemanticVisitor) VisitList(ctx *parser.ListContext) interface{} {
	// Set current node for error reporting
	oldNode := sv.currentNode
	sv.currentNode = ctx
	defer func() { sv.currentNode = oldNode }()
	
	children := ctx.GetChildren()
	if len(children) < 3 { // '(' ... ')'
		return nil
	}

	content := children[1 : len(children)-1]
	if len(content) == 0 {
		return nil
	}

	// Get the first element to determine the operation
	firstChild := content[0]
	var firstElement string
	if symbolCtx, ok := firstChild.(*antlr.TerminalNodeImpl); ok {
		firstElement = symbolCtx.GetText()
	}

	switch firstElement {
	case "def":
		sv.handleTypedDefinition(content)
	case "import":
		sv.handleImport(content[1:])
	case "macro":
		sv.handleMacroDefinition(content[1:])
	case "fn":
		sv.handleTypedFunctionLiteral(content[1:])
	case "+", "-", "*", "/":
		sv.handleTypedArithmetic(firstElement, content[1:])
	case "if":
		sv.handleTypedConditional(content[1:])
	default:
		// Check if it's a registered macro
		if sv.macroRegistry.IsMacro(firstElement) {
			sv.handleMacroCall(firstElement, content[1:])
		} else if strings.HasPrefix(firstElement, ".") {
			sv.handleTypedMethodCall(firstElement, content[1:])
		} else if strings.Contains(firstElement, "/") {
			sv.handleTypedSlashNotationCall(firstElement, content[1:])
		} else {
			sv.handleTypedFunctionCall(firstElement, content[1:])
		}
	}

	return nil
}

// VisitArray visits an array literal with type checking
func (sv *SemanticVisitor) VisitArray(ctx *parser.ArrayContext) interface{} {
	// Infer array type and validate homogeneity
	arrayType, err := sv.typeInference.InferTypes(ctx)
	if err != nil {
		sv.addError(fmt.Sprintf("Array type inference failed: %s", err.Error()))
	}
	
	// Generate Go slice code with proper type
	if listType, ok := arrayType.(*ListType); ok {
		sv.codeGen.EmitTypedArray(listType)
	} else {
		sv.codeGen.EmitArray() // Fallback to generic array
	}
	
	return sv.VisitChildren(ctx)
}

// handleTypedDefinition handles variable definitions with type information
func (sv *SemanticVisitor) handleTypedDefinition(content []antlr.Tree) {
	if len(content) < 3 {
		sv.addError("Invalid definition: expected (def name value)")
		return
	}

	// Get variable name
	var varName string
	if symbolCtx, ok := content[1].(*antlr.TerminalNodeImpl); ok {
		varName = symbolCtx.GetText()
	} else {
		sv.addError("Variable name must be a symbol")
		return
	}

	// Infer the type of the value
	valueType := sv.visitWithTypeInfo(content[2])
	
	// Generate typed Go code
	value := sv.evaluateExpressionWithType(content[2], valueType)
	sv.codeGen.EmitTypedVariableDefinition(varName, value, valueType)
	
	// Update namespace binding
	sv.namespaceManager.GetCurrentNamespace().Bind(varName, valueType, false)
}



// handleTypedFunctionLiteral handles function literals with type inference
func (sv *SemanticVisitor) handleTypedFunctionLiteral(content []antlr.Tree) {
	if len(content) < 2 {
		sv.addError("Invalid function literal: expected (fn [params] body)")
		return
	}

	// Parse parameters with type inference
	paramTypes, paramNames, err := sv.parseTypedParameters(content[0])
	if err != nil {
		sv.addError(fmt.Sprintf("Parameter parsing failed: %s", err.Error()))
		return
	}

	// Infer return type from body
	returnType := sv.visitWithTypeInfo(content[1])

	// Generate Go function literal
	sv.codeGen.EmitTypedFunctionLiteral(paramNames, paramTypes, returnType)
}

// handleTypedArithmetic handles arithmetic with type checking
func (sv *SemanticVisitor) handleTypedArithmetic(operator string, operands []antlr.Tree) {
	if len(operands) < 2 {
		sv.addError(fmt.Sprintf("Arithmetic operation %s requires at least 2 operands", operator))
		return
	}

	// Infer types of all operands
	var operandTypes []VexType
	var operandValues []string
	
	for i, operand := range operands {
		operandType := sv.visitWithTypeInfo(operand)
		operandTypes = append(operandTypes, operandType)
		
		// Validate numeric type
		if !sv.isNumericType(operandType) {
			sv.addError(fmt.Sprintf("Arithmetic operand %d must be numeric, got %s", i, operandType.String()))
		}
		
		operandValues = append(operandValues, sv.evaluateExpressionWithType(operand, operandType))
	}

	// Determine result type (promote int to float if needed)
	resultType := sv.determineArithmeticResultType(operandTypes)
	
	// Generate typed arithmetic expression
	sv.codeGen.EmitTypedArithmeticExpression(operator, operandValues, resultType)
}

// handleTypedConditional handles if expressions with type checking
func (sv *SemanticVisitor) handleTypedConditional(content []antlr.Tree) {
	if len(content) < 3 {
		sv.addError("Conditional requires condition, then-branch, and else-branch")
		return
	}

	// Check condition type
	conditionType := sv.visitWithTypeInfo(content[0])
	if !conditionType.Equals(BoolType) {
		sv.addError(fmt.Sprintf("Condition must be boolean, got %s", conditionType.String()))
	}

	// Check branch types
	thenType := sv.visitWithTypeInfo(content[1])
	elseType := sv.visitWithTypeInfo(content[2])

	if !thenType.IsAssignableFrom(elseType) {
		sv.addError(fmt.Sprintf("Conditional branch type mismatch: then=%s, else=%s", 
			thenType.String(), elseType.String()))
	}

	// Generate typed conditional
	condition := sv.evaluateExpressionWithType(content[0], conditionType)
	thenBranch := sv.evaluateExpressionWithType(content[1], thenType)
	elseBranch := sv.evaluateExpressionWithType(content[2], elseType)
	
	sv.codeGen.EmitTypedConditional(condition, thenBranch, elseBranch, thenType)
}

// handleTypedMethodCall handles method calls with type checking
func (sv *SemanticVisitor) handleTypedMethodCall(methodName string, content []antlr.Tree) {
	if len(content) < 1 {
		sv.addError("Invalid method call: expected receiver")
		return
	}

	// Infer receiver type
	receiverType := sv.visitWithTypeInfo(content[0])
	receiver := sv.evaluateExpressionWithType(content[0], receiverType)

	// Infer argument types
	var args []string
	var argTypes []VexType
	for _, arg := range content[1:] {
		argType := sv.visitWithTypeInfo(arg)
		argTypes = append(argTypes, argType)
		args = append(args, sv.evaluateExpressionWithType(arg, argType))
	}

	// Generate typed method call
	sv.codeGen.EmitTypedMethodCall(receiver, methodName[1:], args, receiverType, argTypes)
}

// handleTypedSlashNotationCall handles package function calls with type checking
func (sv *SemanticVisitor) handleTypedSlashNotationCall(functionName string, content []antlr.Tree) {
	// Parse package/function
	parts := strings.SplitN(functionName, "/", 2)
	if len(parts) != 2 {
		sv.addError(fmt.Sprintf("Invalid slash notation: %s", functionName))
		return
	}

	packageName, funcName := parts[0], parts[1]

	// Infer argument types
	var args []string
	var argTypes []VexType
	for _, arg := range content {
		argType := sv.visitWithTypeInfo(arg)
		argTypes = append(argTypes, argType)
		args = append(args, sv.evaluateExpressionWithType(arg, argType))
	}

	// Generate typed package function call
	sv.codeGen.EmitTypedSlashNotationCall(packageName, funcName, args, argTypes)
}

// handleTypedFunctionCall handles regular function calls with type checking
func (sv *SemanticVisitor) handleTypedFunctionCall(functionName string, content []antlr.Tree) {
	// Resolve function binding
	binding, exists := sv.namespaceManager.GetCurrentNamespace().Resolve(functionName)
	if !exists {
		sv.addError(fmt.Sprintf("Undefined function: %s", functionName))
		return
	}

	if !binding.IsFunction {
		sv.addError(fmt.Sprintf("Symbol %s is not a function", functionName))
		return
	}

	funcType, ok := binding.Type.(*FunctionType)
	if !ok {
		sv.addError(fmt.Sprintf("Invalid function type for %s", functionName))
		return
	}

	// Check argument count
	if len(content) != len(funcType.Parameters) {
		sv.addError(fmt.Sprintf("Function %s expects %d arguments, got %d", 
			functionName, len(funcType.Parameters), len(content)))
		return
	}

	// Infer and validate argument types
	var args []string
	for i, arg := range content {
		argType := sv.visitWithTypeInfo(arg)
		expectedType := funcType.Parameters[i]
		
		if !expectedType.IsAssignableFrom(argType) {
			sv.addError(fmt.Sprintf("Argument %d to %s: expected %s, got %s", 
				i, functionName, expectedType.String(), argType.String()))
		}
		
		args = append(args, sv.evaluateExpressionWithType(arg, argType))
	}

	// Generate typed function call
	sv.codeGen.EmitTypedFunctionCall(functionName, args, funcType.ReturnType)
}

// Helper methods (reusing some from original AST visitor)

// handleImport handles import statements (unchanged)
func (sv *SemanticVisitor) handleImport(content []antlr.Tree) {
	if len(content) < 1 {
		sv.addError("Invalid import: expected package path")
		return
	}

	importPath := sv.evaluateExpression(content[0])
	sv.codeGen.EmitImport(importPath)
}

// handleMacroDefinition handles macro definitions like (macro name [params] body)
func (sv *SemanticVisitor) handleMacroDefinition(content []antlr.Tree) {
	if len(content) < 3 {
		sv.addError("Invalid macro definition: expected (macro name [params] body)")
		return
	}

	// Get macro name
	var macroName string
	if symbolCtx, ok := content[0].(*antlr.TerminalNodeImpl); ok {
		macroName = symbolCtx.GetText()
	} else {
		sv.addError("Macro name must be a symbol")
		return
	}

	// Get parameters from array
	paramList := content[1]
	var params []string
	if arrayCtx, ok := paramList.(*parser.ArrayContext); ok {
		children := arrayCtx.GetChildren()
		// Skip [ and ] brackets
		for i := 1; i < len(children)-1; i++ {
			if terminalNode, ok := children[i].(*antlr.TerminalNodeImpl); ok {
				params = append(params, terminalNode.GetText())
			}
		}
	} else {
		sv.addError("Macro parameters must be in an array")
		return
	}

	// Macro body template (remaining content)
	if len(content) > 2 {
		template := content[2] // For now, just take the first body element
		sv.macroRegistry.RegisterMacro(macroName, params, template)
		
		// Don't generate any Go code for macro definitions
		sv.codeGen.writeIndented("// Registered macro: " + macroName + "\n")
	}
}

// handleMacroCall handles macro calls (unchanged)
func (sv *SemanticVisitor) handleMacroCall(macroName string, content []antlr.Tree) {
	expanded, err := sv.macroRegistry.ExpandMacro(macroName, content)
	if err != nil {
		sv.addError(fmt.Sprintf("Macro expansion failed: %s", err.Error()))
		return
	}
	sv.codeGen.writeIndented("// Macro expansion: " + macroName + "\n")
	sv.codeGen.writeIndented(expanded + "\n")
}

// parseTypedParameters parses function parameters with type inference
func (sv *SemanticVisitor) parseTypedParameters(paramNode antlr.Tree) ([]VexType, []string, error) {
	arrayCtx, ok := paramNode.(*parser.ArrayContext)
	if !ok {
		return nil, nil, fmt.Errorf("parameters must be an array")
	}

	children := arrayCtx.GetChildren()
	if len(children) < 3 { // '[' ... ']'
		return []VexType{}, []string{}, nil
	}

	content := children[1 : len(children)-1]
	var paramTypes []VexType
	var paramNames []string

	for _, paramNode := range content {
		if symbolCtx, ok := paramNode.(*antlr.TerminalNodeImpl); ok {
			paramName := symbolCtx.GetText()
			paramNames = append(paramNames, paramName)
			// For now, use unknown type (will be inferred from usage)
			paramTypes = append(paramTypes, NewUnknownType(sv.typeInference.getNextUnknownID()))
		}
	}

	return paramTypes, paramNames, nil
}

// evaluateExpressionWithType evaluates an expression with type information
func (sv *SemanticVisitor) evaluateExpressionWithType(node antlr.Tree, expectedType VexType) string {
	// Use type information to generate more specific Go code
	switch ctx := node.(type) {
	case *antlr.TerminalNodeImpl:
		text := ctx.GetText()
		return sv.formatLiteralWithType(text, expectedType)
	default:
		// Fallback to basic evaluation
		return sv.evaluateExpression(node)
	}
}

// formatLiteralWithType formats literals with type information
func (sv *SemanticVisitor) formatLiteralWithType(literal string, vexType VexType) string {
	// Add type-specific formatting
	if vexType.Equals(IntType) && !strings.Contains(literal, ".") {
		return literal // Ensure integer literals stay as integers
	}
	if vexType.Equals(FloatType) && !strings.Contains(literal, ".") {
		return literal + ".0" // Convert integers to floats when needed
	}
	return literal
}

// evaluateExpression basic expression evaluation (fallback)
func (sv *SemanticVisitor) evaluateExpression(node antlr.Tree) string {
	switch ctx := node.(type) {
	case *antlr.TerminalNodeImpl:
		return ctx.GetText()
	default:
		return "/* unknown */"
	}
}

// isNumericType checks if a type is numeric
func (sv *SemanticVisitor) isNumericType(vexType VexType) bool {
	return vexType.Equals(IntType) || vexType.Equals(FloatType)
}

// determineArithmeticResultType determines the result type of arithmetic operations
func (sv *SemanticVisitor) determineArithmeticResultType(operandTypes []VexType) VexType {
	hasFloat := false
	for _, t := range operandTypes {
		if t.Equals(FloatType) {
			hasFloat = true
			break
		}
	}
	if hasFloat {
		return FloatType
	}
	return IntType
}

// addError adds an error to the error list
func (sv *SemanticVisitor) addError(message string) {
	sv.addErrorWithNode(message, sv.currentNode)
}

// addErrorWithNode adds an error with specific node position information
func (sv *SemanticVisitor) addErrorWithNode(message string, node antlr.Tree) {
	line, column := sv.getNodePosition(node)
	if line > 0 {
		formattedError := fmt.Sprintf("line %d:%d: %s", line, column, message)
		sv.errors = append(sv.errors, formattedError)
	} else {
		sv.errors = append(sv.errors, message)
	}
}

// getNodePosition extracts line and column information from an AST node
func (sv *SemanticVisitor) getNodePosition(node antlr.Tree) (int, int) {
	if node == nil {
		return 0, 0
	}
	
	switch ctx := node.(type) {
	case *antlr.TerminalNodeImpl:
		token := ctx.GetSymbol()
		return token.GetLine(), token.GetColumn()
	case *parser.ListContext:
		if ctx.GetStart() != nil {
			return ctx.GetStart().GetLine(), ctx.GetStart().GetColumn()
		}
	case *parser.ArrayContext:
		if ctx.GetStart() != nil {
			return ctx.GetStart().GetLine(), ctx.GetStart().GetColumn()
		}
	case *parser.ProgramContext:
		if ctx.GetStart() != nil {
			return ctx.GetStart().GetLine(), ctx.GetStart().GetColumn()
		}
	}
	return 0, 0
}