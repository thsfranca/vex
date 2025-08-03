package tests

import (
	"strings"
	"testing"

	"github.com/antlr4-go/antlr/v4"
	"github.com/thsfranca/vex/internal/transpiler"
	"github.com/thsfranca/vex/internal/transpiler/parser"
)

func TestNewMacroRegistry(t *testing.T) {
	registry := transpiler.NewMacroRegistry()
	if registry == nil {
		t.Error("Expected NewMacroRegistry() to return a non-nil registry")
	}
}

func TestMacroRegistry_RegisterMacro(t *testing.T) {
	registry := transpiler.NewMacroRegistry()
	
	// Create a simple template (mock AST node)
	template := createMockTemplateNode("(~param1 + ~param2)")
	
	registry.RegisterMacro("add", []string{"param1", "param2"}, template)
	
	// Check that macro is registered
	if !registry.IsMacro("add") {
		t.Error("Expected 'add' to be registered as a macro")
	}
	
	// Check that non-existent macro is not registered
	if registry.IsMacro("nonexistent") {
		t.Error("Expected 'nonexistent' to not be registered as a macro")
	}
}

func TestMacroRegistry_GetMacro(t *testing.T) {
	registry := transpiler.NewMacroRegistry()
	
	template := createMockTemplateNode("(~x + ~y)")
	registry.RegisterMacro("test-macro", []string{"x", "y"}, template)
	
	// Get existing macro
	macro, exists := registry.GetMacro("test-macro")
	if !exists {
		t.Error("Expected to find registered macro")
		return
	}
	if macro == nil {
		t.Error("Expected non-nil macro definition")
		return
	}
	if macro.Name != "test-macro" {
		t.Errorf("Expected macro name 'test-macro', got '%s'", macro.Name)
	}
	if len(macro.Parameters) != 2 {
		t.Errorf("Expected 2 parameters, got %d", len(macro.Parameters))
	}
	if macro.Parameters[0] != "x" || macro.Parameters[1] != "y" {
		t.Errorf("Expected parameters [x, y], got %v", macro.Parameters)
	}
	
	// Get non-existent macro
	_, exists = registry.GetMacro("nonexistent")
	if exists {
		t.Error("Expected non-existent macro to not be found")
	}
}

func TestMacroRegistry_IsMacro(t *testing.T) {
	registry := transpiler.NewMacroRegistry()
	
	// Initially no macros
	if registry.IsMacro("anything") {
		t.Error("Expected no macros initially")
	}
	
	// Register a macro
	template := createMockTemplateNode("test")
	registry.RegisterMacro("my-macro", []string{}, template)
	
	// Now it should be found
	if !registry.IsMacro("my-macro") {
		t.Error("Expected registered macro to be found")
	}
	
	// Case sensitive
	if registry.IsMacro("My-Macro") {
		t.Error("Expected macro lookup to be case sensitive")
	}
}

func TestMacroRegistry_ExpandMacro_SimpleSubstitution(t *testing.T) {
	registry := transpiler.NewMacroRegistry()
	
	// Register a simple macro: (add ~x ~y) -> (~x + ~y)
	template := createMockTemplateNode("(~x + ~y)")
	registry.RegisterMacro("add", []string{"x", "y"}, template)
	
	// Create mock arguments
	args := []antlr.Tree{
		createMockTerminalNode("10"),
		createMockTerminalNode("20"),
	}
	
	result, err := registry.ExpandMacro("add", args)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	
	// Check substitution occurred
	if !strings.Contains(result, "10") || !strings.Contains(result, "20") {
		t.Errorf("Expected result to contain arguments, got: %s", result)
	}
}

func TestMacroRegistry_ExpandMacro_NotFound(t *testing.T) {
	registry := transpiler.NewMacroRegistry()
	
	args := []antlr.Tree{
		createMockTerminalNode("arg1"),
	}
	
	_, err := registry.ExpandMacro("nonexistent", args)
	if err == nil {
		t.Error("Expected error for non-existent macro")
	}
	
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("Expected 'not found' error, got: %v", err)
	}
}

func TestMacroRegistry_ExpandMacro_NoParameters(t *testing.T) {
	registry := transpiler.NewMacroRegistry()
	
	// Macro with no parameters
	template := createMockTemplateNode("(println \"Hello World\")")
	registry.RegisterMacro("hello", []string{}, template)
	
	result, err := registry.ExpandMacro("hello", []antlr.Tree{})
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	
	if !strings.Contains(result, "Hello World") {
		t.Errorf("Expected result to contain template content, got: %s", result)
	}
}

func TestMacroRegistry_ExpandMacro_ExtraArguments(t *testing.T) {
	registry := transpiler.NewMacroRegistry()
	
	// Macro with one parameter
	template := createMockTemplateNode("(print ~msg)")
	registry.RegisterMacro("print-msg", []string{"msg"}, template)
	
	// Provide more arguments than parameters
	args := []antlr.Tree{
		createMockTerminalNode("\"hello\""),
		createMockTerminalNode("\"extra\""),
	}
	
	result, err := registry.ExpandMacro("print-msg", args)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	
	// Should use the first argument and ignore extras
	if !strings.Contains(result, "hello") {
		t.Errorf("Expected result to contain first argument, got: %s", result)
	}
}

func TestMacroRegistry_ExpandMacro_FewerArguments(t *testing.T) {
	registry := transpiler.NewMacroRegistry()
	
	// Macro with two parameters
	template := createMockTemplateNode("(~x + ~y)")
	registry.RegisterMacro("add", []string{"x", "y"}, template)
	
	// Provide fewer arguments than parameters
	args := []antlr.Tree{
		createMockTerminalNode("10"),
	}
	
	result, err := registry.ExpandMacro("add", args)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	
	// Should substitute available arguments, leave others as template placeholders
	if !strings.Contains(result, "10") {
		t.Errorf("Expected result to contain provided argument, got: %s", result)
	}
}

func TestMacroRegistry_Integration_WithTranspiler(t *testing.T) {
	// Test macro registry integration with the full transpiler
	input := `
(macro simple-def [name value] (def ~name ~value))
(simple-def x 42)
(simple-def y "hello")
`
	
	tr := transpiler.New()
	result, err := tr.TranspileFromInput(input)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	
	// Check that macros were expanded and transpiled
	if !strings.Contains(result, "x := 42") {
		t.Error("Expected macro expansion to create variable definition for x")
	}
	if !strings.Contains(result, `y := "hello"`) {
		t.Error("Expected macro expansion to create variable definition for y")
	}
}

func TestMacroRegistry_ComplexMacro(t *testing.T) {
	// Test a more complex macro with multiple substitutions
	input := `
(macro simple-macro [name value] (def ~name ~value))
(simple-macro test-var 42)
`
	
	tr := transpiler.New()
	result, err := tr.TranspileFromInput(input)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	
	// Check that the macro was expanded properly
	// Note: This is a simpler test since macro expansion is complex
	expectedParts := []string{
		"// Registered macro: simple-macro",
		"test-var := 42",
	}
	
	for _, part := range expectedParts {
		if !strings.Contains(result, part) {
			t.Errorf("Expected result to contain: %s\nActual result: %s", part, result)
		}
	}
}

// Helper functions for creating mock AST nodes

func createMockTerminalNode(text string) *antlr.TerminalNodeImpl {
	// Create a mock terminal node
	// In a real scenario, this would be created by the ANTLR parser
	source := antlr.NewInputStream(text)
	lexer := parser.NewVexLexer(source)
	token := lexer.NextToken()
	token.SetText(text)
	return antlr.NewTerminalNodeImpl(token)
}

func createMockTemplateNode(content string) antlr.Tree {
	// For testing purposes, we can use a terminal node as a simple template
	// In reality, this would be a more complex AST structure
	return createMockTerminalNode(content)
}

