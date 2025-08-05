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
	macros     map[string]*Macro // Registered macros
}

// New creates a new transpiler instance
func New() *Transpiler {
	t := &Transpiler{
		imports:   make(map[string]bool),
		goModules: make(map[string]string),
		macros:    make(map[string]*Macro),
	}
	
	// Register built-in macros
	t.registerBuiltinMacros()
	
	return t
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

	// Visit the parse tree
	t.visitProgram(tree.(*parser.ProgramContext))

	return t.generateGoCode(), nil
}

// TranspileFromFile transpiles a Vex file to Go
func (t *Transpiler) TranspileFromFile(filename string) (string, error) {
	content, err := os.ReadFile(filename)
	if err != nil {
		return "", fmt.Errorf("failed to read file %s: %v", filename, err)
	}

	return t.TranspileFromInput(string(content))
}

// visitProgram handles the top-level program node
func (t *Transpiler) visitProgram(ctx *parser.ProgramContext) {
	children := ctx.GetChildren()
	
	// First, collect all the list expressions (skip terminal nodes like EOF)
	var listExpressions []*parser.ListContext
	for _, child := range children {
		if listCtx, ok := child.(*parser.ListContext); ok {
			listExpressions = append(listExpressions, listCtx)
		}
	}
	
	// Process each list expression
	for i, listCtx := range listExpressions {
		// Check if this is the last expression in the program
		isLast := i == len(listExpressions)-1
		if isLast {
			t.handleLastExpression(listCtx)
		} else {
			t.visitList(listCtx)
		}
	}
}

// visitList handles list expressions: (function arg1 arg2 ...)
func (t *Transpiler) visitList(ctx *parser.ListContext) {
	childCount := ctx.GetChildCount()
	if childCount < 3 { // Need at least: '(', function, ')'
		return
	}

	// Get function name (first child after '(')
	funcNameNode := ctx.GetChild(1)
	funcName := t.visitNode(funcNameNode)

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
	case "fn":
		t.handleLambda(args)
	case "macro":
		t.handleMacroWithContext(ctx, childCount)
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
		return t.evaluateExpression(n)
	case *parser.ArrayContext:
		return t.visitArray(n)
	case antlr.TerminalNode:
		text := n.GetText()
		// Remove quotes from strings but keep the quotes in output
		if strings.HasPrefix(text, "\"") && strings.HasSuffix(text, "\"") {
			return text // Keep quotes for Go string literals
		}
		return text
	default:
		return ""
	}
}

// visitArray handles array literals: [element1 element2 ...]
func (t *Transpiler) visitArray(ctx *parser.ArrayContext) string {
	var elements []string
	
	// Process all children except '[' and ']'
	for i := 1; i < ctx.GetChildCount()-1; i++ {
		child := ctx.GetChild(i)
		if child != nil {
			elements = append(elements, t.visitNode(child))
		}
	}
	
	return fmt.Sprintf("[]interface{}{%s}", strings.Join(elements, ", "))
}

// generateGoCode generates the final Go code with package and imports
func (t *Transpiler) generateGoCode() string {
	var result strings.Builder
	
	result.WriteString("package main\n\n")
	
	// Add imports
	if len(t.imports) > 0 {
		for importPath := range t.imports {
			result.WriteString(fmt.Sprintf("import \"%s\"\n", importPath))
		}
		result.WriteString("\n")
	}
	
	result.WriteString("func main() {\n")
	
	// Add the transpiled code with proper indentation
	lines := strings.Split(t.output.String(), "\n")
	for _, line := range lines {
		if strings.TrimSpace(line) != "" {
			result.WriteString("\t" + line + "\n")
		}
	}
	
	result.WriteString("}\n")
	
	return result.String()
}

// handleLastExpression ensures the last expression in a program is properly handled
func (t *Transpiler) handleLastExpression(ctx *parser.ListContext) {
	childCount := ctx.GetChildCount()
	if childCount < 3 {
		return
	}

	// Get function name (first child after '(')
	funcNameNode := ctx.GetChild(1)
	funcName := t.visitNode(funcNameNode)

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
	case "fn":
		t.handleLambda(args)
	case "macro":
		t.handleMacroWithContext(ctx, childCount)
	case "first", "rest", "cons", "count", "empty?":
		t.handleCollectionOp(funcName, args)
	case "+", "-", "*", "/":
		t.handleArithmetic(funcName, args)
	default:
		// For function calls, capture the result
		if strings.Contains(funcName, "Println") || strings.Contains(funcName, "Printf") {
			// These functions are called for their side effects
			t.handleFunctionCall(funcName, args)
		} else {
			// Regular function - capture result
			t.handleFunctionCall(funcName, args)
		}
	}
}

// GetDetectedModules returns the detected third-party modules for go.mod generation
func (t *Transpiler) GetDetectedModules() map[string]string {
	return t.goModules
}