package transpiler

import (
	"fmt"
	"strings"

	"github.com/antlr4-go/antlr/v4"
	"github.com/thsfranca/vex/internal/transpiler/parser"
)

// ASTVisitor implements the visitor pattern for Vex AST nodes
type ASTVisitor struct {
	parser.BaseVexVisitor
	codeGen       *CodeGenerator
	macroRegistry *MacroRegistry
}

// NewASTVisitor creates a new AST visitor with a code generator
func NewASTVisitor() *ASTVisitor {
	return &ASTVisitor{
		codeGen:       NewCodeGenerator(),
		macroRegistry: NewMacroRegistry(),
	}
}

// GetGeneratedCode returns the generated Go code
func (v *ASTVisitor) GetGeneratedCode() string {
	return v.codeGen.GetCode()
}

// GetCodeGenerator returns the code generator for accessing imports
func (v *ASTVisitor) GetCodeGenerator() *CodeGenerator {
	return v.codeGen
}

// VisitProgram visits the root node of the parse tree
func (v *ASTVisitor) VisitProgram(ctx *parser.ProgramContext) interface{} {
	// Visit all child lists
	for _, listCtx := range ctx.AllList() {
		listCtx.Accept(v)
	}
	return nil
}

// VisitList visits a list expression (s-expression)
func (v *ASTVisitor) VisitList(ctx *parser.ListContext) interface{} {
	children := ctx.GetChildren()
	if len(children) < 3 { // '(' ... ')'
		return nil
	}

	// Skip opening parenthesis, process content, skip closing parenthesis
	content := children[1 : len(children)-1]

	if len(content) == 0 {
		return nil
	}

	// Check if this is a special form (function call, def, etc.)
	firstChild := content[0]

	// Get the text of the first element
	var firstElement string
	if symbolCtx, ok := firstChild.(*antlr.TerminalNodeImpl); ok {
		firstElement = symbolCtx.GetText()
	}

	switch firstElement {
	case "def":
		v.handleDefinition(content)
	case "import":
		v.handleImport(content[1:]) // Skip the "import" symbol
	case "macro":
		v.handleMacroDefinition(content[1:]) // Skip the "macro" symbol
	case "fn":
		v.handleFunctionLiteral(content[1:]) // Skip the "fn" symbol
	case "+", "-", "*", "/":
		v.handleArithmetic(firstElement, content[1:])
	default:
		// Check if it's a registered macro
		if v.macroRegistry.IsMacro(firstElement) {
			v.handleMacroCall(firstElement, content[1:])
		} else if strings.HasPrefix(firstElement, ".") {
			// Check if it's a method call (starts with .)
			v.handleMethodCall(firstElement, content[1:])
		} else {
			// Handle other expressions or function calls
			v.handleExpression(content)
		}
	}

	return nil
}

// VisitArray visits an array literal
func (v *ASTVisitor) VisitArray(ctx *parser.ArrayContext) interface{} {
	// For now, just visit children
	// Arrays will be implemented in a later phase
	return v.VisitChildren(ctx)
}

// handleDefinition handles variable definitions like (def x 10)
func (v *ASTVisitor) handleDefinition(content []antlr.Tree) {
	if len(content) < 3 {
		v.codeGen.writeIndented("// Invalid definition\n")
		return
	}

	// Get variable name
	var varName string
	if symbolCtx, ok := content[1].(*antlr.TerminalNodeImpl); ok {
		varName = symbolCtx.GetText()
	}

	// Get value
	value := v.evaluateExpression(content[2])
	v.codeGen.EmitVariableDefinition(varName, value)
}

// handleArithmetic handles arithmetic expressions like (+ 1 2)
func (v *ASTVisitor) handleArithmetic(operator string, operands []antlr.Tree) {
	var operandValues []string
	for _, operand := range operands {
		operandValues = append(operandValues, v.evaluateExpression(operand))
	}
	v.codeGen.EmitArithmeticExpression(operator, operandValues)
}

// handleImport handles import statements like (import "net/http")
func (v *ASTVisitor) handleImport(content []antlr.Tree) {
	if len(content) < 1 {
		v.codeGen.writeIndented("// Invalid import\n")
		return
	}

	// Get the import path
	importPath := v.evaluateExpression(content[0])
	v.codeGen.EmitImport(importPath)
}

// handleMethodCall handles method calls like (.HandleFunc router "/path" handler)
func (v *ASTVisitor) handleMethodCall(methodName string, content []antlr.Tree) {
	if len(content) < 1 {
		v.codeGen.writeIndented("// Invalid method call\n")
		return
	}

	// Get the receiver (first argument)
	receiver := v.evaluateExpression(content[0])
	
	// Get the method arguments
	var args []string
	for _, arg := range content[1:] {
		args = append(args, v.evaluateExpression(arg))
	}
	
	v.codeGen.EmitMethodCall(receiver, methodName[1:], args) // Remove the dot from method name
}

// handleMacroDefinition handles macro definitions like (macro name [params] body)
func (v *ASTVisitor) handleMacroDefinition(content []antlr.Tree) {
	if len(content) < 3 {
		v.codeGen.writeIndented("// Invalid macro definition\n")
		return
	}

	// Get macro name
	var macroName string
	if symbolCtx, ok := content[0].(*antlr.TerminalNodeImpl); ok {
		macroName = symbolCtx.GetText()
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
	}

	// Macro body template (remaining content)
	if len(content) > 2 {
		template := content[2] // For now, just take the first body element
		v.macroRegistry.RegisterMacro(macroName, params, template)
		
		// Don't generate any Go code for macro definitions
		v.codeGen.writeIndented(fmt.Sprintf("// Registered macro: %s\n", macroName))
	}
}

// handleMacroCall handles calls to user-defined macros
func (v *ASTVisitor) handleMacroCall(macroName string, args []antlr.Tree) {
	expanded, err := v.macroRegistry.ExpandMacro(macroName, args)
	if err != nil {
		v.codeGen.writeIndented(fmt.Sprintf("// Error expanding macro %s: %s\n", macroName, err.Error()))
		return
	}

	// Re-parse and process the expanded code
	err = v.processExpandedCode(expanded)
	if err != nil {
		v.codeGen.writeIndented(fmt.Sprintf("// Error processing expanded macro %s: %s\n", macroName, err.Error()))
		v.codeGen.writeIndented(fmt.Sprintf("// Expanded code was: %s\n", expanded))
	}
}

// handleFunctionLiteral handles function literals like (fn [w r] body)
func (v *ASTVisitor) handleFunctionLiteral(content []antlr.Tree) {
	if len(content) < 2 {
		v.codeGen.writeIndented("// Invalid function literal\n")
		return
	}

	// First element should be parameter list (array)
	paramList := content[0]
	
	// Extract parameter names
	var params []string
	if arrayCtx, ok := paramList.(*parser.ArrayContext); ok {
		children := arrayCtx.GetChildren()
		// Skip [ and ] brackets
		for i := 1; i < len(children)-1; i++ {
			if terminalNode, ok := children[i].(*antlr.TerminalNodeImpl); ok {
				params = append(params, terminalNode.GetText())
			}
		}
	}

	// Function body (rest of the arguments)
	bodyElements := content[1:]
	
	functionLiteral := v.codeGen.EmitFunctionLiteral(params, bodyElements, v)
	// This returns a function literal expression, but we need to handle it as a value
	v.codeGen.EmitExpressionStatement(functionLiteral)
}

// handleExpression handles general expressions
func (v *ASTVisitor) handleExpression(content []antlr.Tree) {
	// For now, just evaluate each element as a standalone expression
	for _, element := range content {
		value := v.evaluateExpression(element)
		v.codeGen.EmitExpressionStatement(value)
	}
}

// evaluateExpression evaluates a single expression and returns its Go representation
func (v *ASTVisitor) evaluateExpression(node antlr.Tree) string {
	switch ctx := node.(type) {
	case *antlr.TerminalNodeImpl:
		text := ctx.GetText()

		// Check if it's a string literal
		if strings.HasPrefix(text, "\"") && strings.HasSuffix(text, "\"") {
			return text // Already properly quoted
		}

		// Check if it's a number
		if isNumber(text) {
			return text
		}

		// Otherwise, it's a symbol/identifier
		return text

	case *parser.ListContext:
		// Handle nested lists by recursively visiting them
		// Create a temporary visitor to handle the nested list
		nestedVisitor := NewASTVisitor()
		ctx.Accept(nestedVisitor)
		generatedCode := nestedVisitor.GetGeneratedCode()
		
		// If it generated code, extract the expression part
		// For method calls like (.NewRouter mux), we want "mux.NewRouter()"
		if strings.Contains(generatedCode, "=") {
			// Extract the right side of the assignment
			parts := strings.Split(strings.TrimSpace(generatedCode), "=")
			if len(parts) >= 2 {
				return strings.TrimSpace(parts[len(parts)-1])
			}
		}
		
		return "/* nested list */"

	default:
		return "/* unknown */"
	}
}

// processExpandedCode re-parses and processes expanded macro code
func (v *ASTVisitor) processExpandedCode(expandedCode string) error {
	// Create a new lexer and parser for the expanded code
	inputStream := antlr.NewInputStream(expandedCode)
	lexer := parser.NewVexLexer(inputStream)
	tokenStream := antlr.NewCommonTokenStream(lexer, 0)
	vexParser := parser.NewVexParser(tokenStream)
	
	// Parse the expanded code
	tree := vexParser.List() // Parse as a single list expression
	
	// Visit the parsed tree with the current visitor
	tree.Accept(v)
	
	return nil
}

// isNumber checks if a string represents a numeric value
func isNumber(s string) bool {
	for _, char := range s {
		if char < '0' || char > '9' {
			if char != '.' && char != '-' {
				return false
			}
		}
	}
	return len(s) > 0
}
