// Package transpiler provides a minimal Vex to Go transpiler
package transpiler

import (
	"fmt"
	"strings"

	"github.com/antlr4-go/antlr/v4"
	"github.com/thsfranca/vex/internal/transpiler/parser"
)

// Transpiler converts Vex code to Go code
type Transpiler struct {
	output  strings.Builder
	imports map[string]bool
}

// New creates a new transpiler instance
func New() *Transpiler {
	return &Transpiler{
		imports: make(map[string]bool),
	}
}

// TranspileFromInput transpiles Vex source code to Go
func (t *Transpiler) TranspileFromInput(input string) (string, error) {
	// Parse the input
	inputStream := antlr.NewInputStream(input)
	lexer := parser.NewVexLexer(inputStream)
	tokenStream := antlr.NewCommonTokenStream(lexer, 0)
	vexParser := parser.NewVexParser(tokenStream)

	// Parse the program
	tree := vexParser.Program()
	
	// Reset output and imports
	t.output.Reset()
	t.imports = make(map[string]bool)
	
	// Generate Go code
	if programCtx, ok := tree.(*parser.ProgramContext); ok {
		t.visitProgram(programCtx)
	}
	
	return t.generateGoCode(), nil
}

// TranspileFromFile transpiles a Vex file to Go
func (t *Transpiler) TranspileFromFile(filename string) (string, error) {
	// Simple file reading would go here
	// For now, just return an error
	return "", fmt.Errorf("file transpilation not implemented yet")
}

// visitProgram processes the top-level program
func (t *Transpiler) visitProgram(ctx *parser.ProgramContext) {
	// Visit all lists in the program
	for _, listCtx := range ctx.AllList() {
		if concreteList, ok := listCtx.(*parser.ListContext); ok {
			t.visitList(concreteList)
		}
	}
}

// visitList handles list expressions (function calls, special forms)
func (t *Transpiler) visitList(ctx *parser.ListContext) {
	childCount := ctx.GetChildCount()
	if childCount <= 2 { // Just parentheses
		return
	}

	// Get the first element (function name)
	firstChild := ctx.GetChild(1)
	if firstChild == nil {
		return
	}

	var funcName string
	if parseTree, ok := firstChild.(antlr.ParseTree); ok {
		funcName = parseTree.GetText()
	} else {
		return
	}

	// Extract arguments
	args := make([]string, 0)
	for i := 2; i < childCount-1; i++ { // Skip '(' and ')'
		child := ctx.GetChild(i)
		if child != nil {
			args = append(args, t.visitNode(child))
		}
	}

	// Handle special forms
	switch funcName {
	case "def":
		t.handleDefinition(args)
	case "import":
		t.handleImport(args)
	case "+", "-", "*", "/":
		t.handleArithmetic(funcName, args)
	default:
		t.handleFunctionCall(funcName, args)
	}
}

// visitNode handles any AST node and returns its value
func (t *Transpiler) visitNode(node antlr.Tree) string {
	switch n := node.(type) {
	case *parser.ListContext:
		// For nested lists, evaluate them and return the expression
		return t.evaluateExpression(n)
	case *parser.ArrayContext:
		return t.visitArray(n)
	case antlr.TerminalNode:
		return n.GetText()
	default:
		if parseTree, ok := node.(antlr.ParseTree); ok {
			return parseTree.GetText()
		}
		return "/* unknown */"
	}
}

// evaluateExpression evaluates a list as an expression and returns the result
func (t *Transpiler) evaluateExpression(ctx *parser.ListContext) string {
	childCount := ctx.GetChildCount()
	if childCount <= 2 { // Just parentheses
		return "nil"
	}

	// Get the first element (function name)
	firstChild := ctx.GetChild(1)
	if firstChild == nil {
		return "nil"
	}

	var funcName string
	if parseTree, ok := firstChild.(antlr.ParseTree); ok {
		funcName = parseTree.GetText()
	} else {
		return "nil"
	}

	// Extract arguments
	args := make([]string, 0)
	for i := 2; i < childCount-1; i++ { // Skip '(' and ')'
		child := ctx.GetChild(i)
		if child != nil {
			args = append(args, t.visitNode(child))
		}
	}

	// Handle special expressions
	switch funcName {
	case "+", "-", "*", "/":
		if len(args) < 2 {
			return "0"
		}
		// Left-associative chaining for multiple operands
		result := "(" + args[0] + " " + funcName + " " + args[1] + ")"
		for i := 2; i < len(args); i++ {
			result = "(" + result + " " + funcName + " " + args[i] + ")"
		}
		return result
	default:
		// Handle function calls
		if strings.Contains(funcName, "/") {
			funcName = strings.ReplaceAll(funcName, "/", ".")
		}
		argStr := strings.Join(args, ", ")
		return fmt.Sprintf("%s(%s)", funcName, argStr)
	}
}

// visitArray handles array literals
func (t *Transpiler) visitArray(ctx *parser.ArrayContext) string {
	childCount := ctx.GetChildCount()
	if childCount <= 2 { // Just brackets
		return "[]interface{}{}"
	}

	elements := make([]string, 0)
	for i := 1; i < childCount-1; i++ { // Skip '[' and ']'
		child := ctx.GetChild(i)
		if child != nil {
			elements = append(elements, t.visitNode(child))
		}
	}

	return "[]interface{}{" + strings.Join(elements, ", ") + "}"
}

// handleDefinition processes variable definitions
func (t *Transpiler) handleDefinition(args []string) {
	if len(args) < 2 {
		t.output.WriteString("// Error: def requires name and value\n")
		return
	}

	varName := args[0]
	varValue := args[1]

	t.output.WriteString(fmt.Sprintf("var %s = %s\n", varName, varValue))
}

// handleImport processes import statements
func (t *Transpiler) handleImport(args []string) {
	if len(args) < 1 {
		t.output.WriteString("// Error: import requires package path\n")
		return
	}

	importPath := args[0]
	// Remove quotes if present
	importPath = strings.Trim(importPath, "\"")
	
	// Store import for later use at package level
	t.imports[importPath] = true
}

// handleArithmetic processes arithmetic operations
func (t *Transpiler) handleArithmetic(op string, args []string) {
	if len(args) < 2 {
		t.output.WriteString("// Error: arithmetic requires at least 2 operands\n")
		return
	}

	// Left-associative chaining for multiple operands
	result := "(" + args[0] + " " + op + " " + args[1] + ")"
	for i := 2; i < len(args); i++ {
		result = "(" + result + " " + op + " " + args[i] + ")"
	}

	t.output.WriteString("_ = " + result + "\n")
}

// handleFunctionCall processes regular function calls
func (t *Transpiler) handleFunctionCall(funcName string, args []string) {
	// Handle package/function notation (fmt/Println -> fmt.Println)
	if strings.Contains(funcName, "/") {
		funcName = strings.ReplaceAll(funcName, "/", ".")
	}

	argStr := strings.Join(args, ", ")
	
	// For functions that return multiple values or don't need assignment, just call them
	if strings.Contains(funcName, "Println") || strings.Contains(funcName, "Printf") {
		t.output.WriteString(fmt.Sprintf("%s(%s)\n", funcName, argStr))
	} else {
		t.output.WriteString(fmt.Sprintf("_ = %s(%s)\n", funcName, argStr))
	}
}

// generateGoCode wraps the generated code in a proper Go program
func (t *Transpiler) generateGoCode() string {
	var result strings.Builder
	
	// Package declaration
	result.WriteString("package main\n\n")
	
	// Add collected imports
	content := t.output.String()
	
	// Auto-detect imports from function calls if not explicitly imported
	if strings.Contains(content, "fmt.") && !t.imports["fmt"] {
		t.imports["fmt"] = true
	}
	
	// Write all imports
	for importPath := range t.imports {
		result.WriteString(fmt.Sprintf("import \"%s\"\n", importPath))
	}
	
	if len(t.imports) > 0 {
		result.WriteString("\n")
	}
	
	// Main function
	result.WriteString("func main() {\n")
	
	// Add the generated code with indentation, but skip any import statements
	lines := strings.Split(strings.TrimSpace(content), "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" && !strings.HasPrefix(trimmed, "import ") {
			result.WriteString("\t" + line + "\n")
		}
	}
	
	result.WriteString("}\n")
	
	return result.String()
}