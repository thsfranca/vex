// Package transpiler provides the core Vex to Go transpiler functionality
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
}

// New creates a new transpiler instance
func New() *Transpiler {
	return &Transpiler{}
}

// TranspileFromInput transpiles Vex source code to Go code
func (t *Transpiler) TranspileFromInput(input string) (string, error) {
	// Phase 1: Pre-process to find and register user-defined macros
	expander := NewMacroExpander()
	err := t.registerUserDefinedMacros(input, expander)
	if err != nil {
		return "", fmt.Errorf("macro registration failed: %w", err)
	}

	// Phase 2: Macro expansion (now includes user-defined macros)
	expandedInput, err := expander.ExpandMacros(input)
	if err != nil {
		return "", fmt.Errorf("macro expansion failed: %w", err)
	}

	// Debug output removed - macro system is working

	// Create input stream from the expanded source code
	inputStream := antlr.NewInputStream(expandedInput)

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

	// Create semantic visitor with type system integration
	semanticVisitor := NewSemanticVisitor()

	// Perform semantic analysis and type checking
	programCtx, ok := tree.(*parser.ProgramContext)
	if !ok {
		return "", fmt.Errorf("expected ProgramContext, got %T", tree)
	}

	err = semanticVisitor.AnalyzeProgram(programCtx)
	if err != nil {
		return "", fmt.Errorf("semantic analysis failed: %w", err)
	}

	// Check for semantic errors
	if semanticVisitor.HasErrors() {
		errorMsg := strings.Join(semanticVisitor.GetErrors(), "; ")
		return "", fmt.Errorf("semantic errors: %s", errorMsg)
	}

	// Generate final Go code with type information
	return t.generateGoCodeWithSemanticVisitor(semanticVisitor), nil
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

// generateGoCodeWithVisitor generates Go code using visitor data (imports + content)
func (t *Transpiler) generateGoCodeWithVisitor(visitor *ASTVisitor) string {
	var result strings.Builder

	// Package declaration
	result.WriteString("package main\n\n")

	// Add imports at the top
	codeGen := visitor.GetCodeGenerator()
	imports := codeGen.GetImports()
	if len(imports) > 0 {
		for _, imp := range imports {
			result.WriteString("import " + imp + "\n")
		}
		result.WriteString("\n")
	}

	// Main function with the generated code
	result.WriteString("func main() {\n")
	content := visitor.GetGeneratedCode()
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
	// No state to reset in current implementation
}

// registerUserDefinedMacros pre-processes source to find and register user-defined macros
func (t *Transpiler) registerUserDefinedMacros(input string, expander *MacroExpander) error {
	fmt.Printf("DEBUG: Starting macro registration for input: %s\n", input)

	// Create a temporary parse to find macro definitions
	inputStream := antlr.NewInputStream(input)
	lexer := parser.NewVexLexer(inputStream)
	tokenStream := antlr.NewCommonTokenStream(lexer, 0)
	vexParser := parser.NewVexParser(tokenStream)

	// Add error listener but allow parsing to continue for macro collection
	errorListener := &ErrorListener{}
	vexParser.RemoveErrorListeners()
	vexParser.AddErrorListener(errorListener)

	// Parse the program
	tree := vexParser.Program()

	// If there are syntax errors, we can't register macros safely
	if errorListener.hasError {
		return fmt.Errorf("syntax error prevents macro registration: %s", errorListener.errorMsg)
	}

	fmt.Printf("DEBUG: Successfully parsed for macro collection\n")

	// Create a macro registry collector
	macroCollector := NewMacroCollector(expander)

	// Walk the tree to find macro definitions
	programCtx, ok := tree.(*parser.ProgramContext)
	if !ok {
		return fmt.Errorf("expected ProgramContext for macro collection, got %T", tree)
	}

	err := macroCollector.CollectMacros(programCtx)
	if err != nil {
		return err
	}

	fmt.Printf("DEBUG: Macro collection completed\n")
	return nil
}

// generateGoCodeWithSemanticVisitor generates Go code using semantic visitor with type information
func (t *Transpiler) generateGoCodeWithSemanticVisitor(semanticVisitor *SemanticVisitor) string {
	var result strings.Builder

	// Package declaration
	result.WriteString("package main\n\n")

	// Add imports at the top
	codeGen := semanticVisitor.GetCodeGenerator()
	imports := codeGen.GetImports()
	if len(imports) > 0 {
		for _, imp := range imports {
			result.WriteString("import " + imp + "\n")
		}
		result.WriteString("\n")
	}

	// Main function with the generated code
	result.WriteString("func main() {\n")
	content := semanticVisitor.GetGeneratedCode()
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
