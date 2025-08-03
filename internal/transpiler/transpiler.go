package transpiler

import (
	"fmt"
	"os"
	"strings"

	"github.com/antlr4-go/antlr/v4"
	"github.com/thsfranca/vex/internal/transpiler/parser"
)

// Transpiler handles the conversion from Vex AST to Go code
type Transpiler struct {
	output strings.Builder
}

// New creates a new transpiler instance
func New() *Transpiler {
	return &Transpiler{
		output: strings.Builder{},
	}
}

// TranspileFromInput transpiles Vex source code to Go code
func (t *Transpiler) TranspileFromInput(input string) (string, error) {
	// Create input stream from the source code
	inputStream := antlr.NewInputStream(input)
	
	// Create lexer
	lexer := parser.NewVexLexer(inputStream)
	
	// Create token stream
	tokenStream := antlr.NewCommonTokenStream(lexer, 0)
	
	// Create parser
	vexParser := parser.NewVexParser(tokenStream)
	
	// Add error listener to catch syntax errors
	errorListener := &ErrorListener{}
	vexParser.RemoveErrorListeners()
	vexParser.AddErrorListener(errorListener)
	
	// Parse starting from the 'program' rule (root rule)
	tree := vexParser.Program()
	
	// Check for parse errors
	if errorListener.hasError {
		return "", fmt.Errorf("syntax error: %s", errorListener.errorMsg)
	}
	
	// Create AST visitor and traverse the tree
	visitor := NewASTVisitor()
	tree.Accept(visitor)
	
	// Generate final Go code
	return t.generateGoCodeWithContent(visitor.GetGeneratedCode()), nil
}

// TranspileFromFile transpiles a .vex file to Go code
func (t *Transpiler) TranspileFromFile(filename string) (string, error) {
	// Read the file content
	content, err := os.ReadFile(filename)
	if err != nil {
		return "", fmt.Errorf("failed to read file %s: %w", filename, err)
	}
	
	// Transpile the content
	return t.TranspileFromInput(string(content))
}

// generateGoCode generates the final Go code
func (t *Transpiler) generateGoCode() string {
	// Add basic Go package structure
	var result strings.Builder
	result.WriteString("package main\n\n")
	result.WriteString("func main() {\n")
	result.WriteString(t.output.String())
	result.WriteString("}\n")
	
	return result.String()
}

// generateGoCodeWithContent generates Go code with the provided content
func (t *Transpiler) generateGoCodeWithContent(content string) string {
	var result strings.Builder
	result.WriteString("package main\n\n")
	result.WriteString("func main() {\n")
	if content != "" {
		// Add indentation to the content
		lines := strings.Split(content, "\n")
		for _, line := range lines {
			if strings.TrimSpace(line) != "" {
				result.WriteString("\t" + line + "\n")
			}
		}
	}
	result.WriteString("}\n")
	
	return result.String()
}

// Reset clears the transpiler state for reuse
func (t *Transpiler) Reset() {
	t.output.Reset()
}

// ErrorListener captures syntax errors during parsing
type ErrorListener struct {
	*antlr.DefaultErrorListener
	hasError bool
	errorMsg string
}

func (el *ErrorListener) SyntaxError(recognizer antlr.Recognizer, offendingSymbol interface{}, line, column int, msg string, e antlr.RecognitionException) {
	el.hasError = true
	el.errorMsg = fmt.Sprintf("line %d:%d - %s", line, column, msg)
}