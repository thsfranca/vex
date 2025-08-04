// Package transpiler provides a minimal Vex to Go transpiler
package transpiler

import (
	"fmt"
	"os"
	"strings"

	"github.com/antlr4-go/antlr/v4"
	"github.com/thsfranca/vex/internal/transpiler/parser"
)

// Transpiler converts Vex code to Go code
type Transpiler struct {
	output     strings.Builder
	imports    map[string]bool
	goModules  map[string]string // Track third-party modules for go.mod
}

// New creates a new transpiler instance
func New() *Transpiler {
	return &Transpiler{
		imports:   make(map[string]bool),
		goModules: make(map[string]string),
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

	// Reset output, imports, and modules
	t.output.Reset()
	t.imports = make(map[string]bool)
	t.goModules = make(map[string]string)
	
	// Generate Go code
	if programCtx, ok := tree.(*parser.ProgramContext); ok {
		t.visitProgram(programCtx)
	}
	
	return t.generateGoCode(), nil
}

// TranspileFromFile transpiles a Vex file to Go
func (t *Transpiler) TranspileFromFile(filename string) (string, error) {
	// Read the file content
	content, err := os.ReadFile(filename)
	if err != nil {
		return "", fmt.Errorf("error reading file %s: %v", filename, err)
	}

	// Use the existing TranspileFromInput method
	return t.TranspileFromInput(string(content))
}

// visitProgram processes the top-level program
func (t *Transpiler) visitProgram(ctx *parser.ProgramContext) {
	lists := ctx.AllList()
	
	// Visit all lists in the program
	for i, listCtx := range lists {
		if concreteList, ok := listCtx.(*parser.ListContext); ok {
			// For the last expression, make it the return value
			if i == len(lists)-1 {
				t.handleLastExpression(concreteList)
			} else {
				t.visitList(concreteList)
			}
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
	case "if":
		t.handleIf(args)
	case "do":
		t.handleDo(args)
	case "first", "rest", "cons", "count", "empty?":
		t.handleCollectionOp(funcName, args)
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
	case "if":
		return t.evaluateIf(args)
	case "do":
		return t.evaluateDo(args)
	case "first", "rest", "cons", "count", "empty?":
		return t.evaluateCollectionOp(funcName, args)
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

	// Generate variable definition and use it immediately (functional style)
	t.output.WriteString(fmt.Sprintf("var %s = %s\n", varName, varValue))
	t.output.WriteString(fmt.Sprintf("_ = %s // Use variable to satisfy Go compiler\n", varName))
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
	
	// Track third-party modules for go.mod generation
	if t.isThirdPartyModule(importPath) {
		t.goModules[importPath] = "v1.0.0" // Use a reasonable default version
	}
}

// isThirdPartyModule checks if an import is a third-party module
func (t *Transpiler) isThirdPartyModule(importPath string) bool {
	// Standard library packages don't need go.mod entries
	standardPkgs := []string{
		"fmt", "strings", "strconv", "time", "os", "io", "net", "http",
		"database", "encoding", "crypto", "math", "sort", "regexp",
		"context", "sync", "testing", "flag", "log", "path", "reflect",
	}
	
	for _, stdPkg := range standardPkgs {
		if strings.HasPrefix(importPath, stdPkg) {
			return false
		}
	}
	
	// If it contains a dot, it's likely a third-party module
	return strings.Contains(importPath, ".")
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

// handleIf processes if conditionals: (if condition then-expr else-expr)
func (t *Transpiler) handleIf(args []string) {
	if len(args) < 2 {
		t.output.WriteString("// Error: if requires at least condition and then-expr\n")
		return
	}
	
	condition := args[0]
	thenExpr := args[1]
	var elseExpr string
	if len(args) > 2 {
		elseExpr = args[2]
	} else {
		elseExpr = "nil"
	}
	
	t.output.WriteString(fmt.Sprintf("if %s {\n\t%s\n} else {\n\t%s\n}\n", condition, thenExpr, elseExpr))
}

// evaluateIf evaluates if conditionals as expressions
func (t *Transpiler) evaluateIf(args []string) string {
	if len(args) < 2 {
		return "nil"
	}
	
	condition := args[0]
	thenExpr := args[1]
	var elseExpr string
	if len(args) > 2 {
		elseExpr = args[2]
	} else {
		elseExpr = "nil"
	}
	
	return fmt.Sprintf("func() interface{} { if %s { return %s } else { return %s } }()", condition, thenExpr, elseExpr)
}

// handleDo processes do blocks: (do expr1 expr2 expr3)
func (t *Transpiler) handleDo(args []string) {
	for _, expr := range args {
		t.output.WriteString(expr + "\n")
	}
}

// evaluateDo evaluates do blocks as expressions (returns last expression)
func (t *Transpiler) evaluateDo(args []string) string {
	if len(args) == 0 {
		return "nil"
	}
	
	// Generate function that executes all expressions and returns the last one
	var statements []string
	for i, expr := range args {
		if i == len(args)-1 {
			statements = append(statements, "return "+expr)
		} else {
			statements = append(statements, expr)
		}
	}
	
	return fmt.Sprintf("func() interface{} { %s }()", strings.Join(statements, "; "))
}

// handleCollectionOp processes collection operations
func (t *Transpiler) handleCollectionOp(op string, args []string) {
	switch op {
	case "first":
		if len(args) < 1 {
			t.output.WriteString("_ = nil // Error: first requires collection\n")
			return
		}
		t.output.WriteString(fmt.Sprintf("_ = func() interface{} { if len(%s) > 0 { return %s[0] } else { return nil } }()\n", args[0], args[0]))
	
	case "rest":
		if len(args) < 1 {
			t.output.WriteString("_ = []interface{}{} // Error: rest requires collection\n")
			return
		}
		t.output.WriteString(fmt.Sprintf("_ = func() []interface{} { if len(%s) > 1 { return %s[1:] } else { return []interface{}{} } }()\n", args[0], args[0]))
	
	case "cons":
		if len(args) < 2 {
			t.output.WriteString("_ = []interface{}{} // Error: cons requires element and collection\n")
			return
		}
		t.output.WriteString(fmt.Sprintf("_ = append([]interface{}{%s}, %s...)\n", args[0], args[1]))
	
	case "count":
		if len(args) < 1 {
			t.output.WriteString("_ = 0 // Error: count requires collection\n")
			return
		}
		t.output.WriteString(fmt.Sprintf("_ = len(%s)\n", args[0]))
	
	case "empty?":
		if len(args) < 1 {
			t.output.WriteString("_ = true // Error: empty? requires collection\n")
			return
		}
		t.output.WriteString(fmt.Sprintf("_ = len(%s) == 0\n", args[0]))
	}
}

// evaluateCollectionOp evaluates collection operations as expressions
func (t *Transpiler) evaluateCollectionOp(op string, args []string) string {
	switch op {
	case "first":
		if len(args) < 1 {
			return "nil"
		}
		return fmt.Sprintf("func() interface{} { if len(%s) > 0 { return %s[0] } else { return nil } }()", args[0], args[0])
	
	case "rest":
		if len(args) < 1 {
			return "[]interface{}{}"
		}
		return fmt.Sprintf("func() []interface{} { if len(%s) > 1 { return %s[1:] } else { return []interface{}{} } }()", args[0], args[0])
	
	case "cons":
		if len(args) < 2 {
			return "[]interface{}{}"
		}
		return fmt.Sprintf("append([]interface{}{%s}, %s...)", args[0], args[1])
	
	case "count":
		if len(args) < 1 {
			return "0"
		}
		return fmt.Sprintf("len(%s)", args[0])
	
	case "empty?":
		if len(args) < 1 {
			return "true"
		}
		return fmt.Sprintf("len(%s) == 0", args[0])
	
	default:
		return "nil"
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

// handleLastExpression processes the last expression in a program (implicit return)
func (t *Transpiler) handleLastExpression(ctx *parser.ListContext) {
	childCount := ctx.GetChildCount()
	if childCount <= 2 { // Just parentheses
		t.output.WriteString("_ = nil // Empty expression\n")
		return
	}

	// Get the first element (function name)
	firstChild := ctx.GetChild(1)
	if firstChild == nil {
		t.output.WriteString("_ = nil // Invalid expression\n")
		return
	}

	var funcName string
	if parseTree, ok := firstChild.(antlr.ParseTree); ok {
		funcName = parseTree.GetText()
	} else {
		t.output.WriteString("_ = nil // Invalid expression\n")
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

	// Handle the last expression - ensure it's used
	switch funcName {
	case "def":
		t.handleDefinition(args)
		// Return the defined variable
		if len(args) >= 1 {
			t.output.WriteString(fmt.Sprintf("_ = %s // Return last defined value\n", args[0]))
		}
	case "import":
		t.handleImport(args)
		t.output.WriteString("_ = \"import completed\" // Import statement result\n")
	case "if":
		t.handleIf(args)
	case "do":
		t.handleDo(args)
	case "first", "rest", "cons", "count", "empty?":
		t.handleCollectionOp(funcName, args)
	case "+", "-", "*", "/":
		t.handleArithmetic(funcName, args)
	default:
		// For function calls, capture the result
		if strings.Contains(funcName, "Println") || strings.Contains(funcName, "Printf") {
			t.output.WriteString(fmt.Sprintf("%s(%s) // Last expression\n", 
				strings.ReplaceAll(funcName, "/", "."), strings.Join(args, ", ")))
		} else {
			t.handleFunctionCall(funcName, args)
		}
	}
}

// GetDetectedModules returns the Go modules detected during transpilation
func (t *Transpiler) GetDetectedModules() map[string]string {
	return t.goModules
}