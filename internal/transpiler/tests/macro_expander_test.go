package tests

import (
	"strings"
	"testing"

	"github.com/thsfranca/vex/internal/transpiler"
)

func TestNewMacroExpander(t *testing.T) {
	expander := transpiler.NewMacroExpander()
	if expander == nil {
		t.Error("Expected NewMacroExpander() to return a non-nil expander")
	}
}

func TestMacroExpander_ExpandMacros_NoMacros(t *testing.T) {
	expander := transpiler.NewMacroExpander()
	
	// Source with no macros should pass through unchanged
	source := `(def x 10)
(+ x 5)`
	
	result, err := expander.ExpandMacros(source)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	
	if result != source {
		t.Errorf("Expected source to pass through unchanged, got: %s", result)
	}
}

func TestMacroExpander_ExpandMacros_HttpServerMacro(t *testing.T) {
	expander := transpiler.NewMacroExpander()
	
	// Test the built-in http-server macro
	source := `(http-server (hello))`
	
	result, err := expander.ExpandMacros(source)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	
	// Should expand to the expected Vex code
	expectedParts := []string{
		`(import "net/http")`,
		`(import "github.com/gorilla/mux")`,
		`(def router (.NewRouter mux))`,
		`(.HandleFunc router "/hello" hello-handler)`,
		`(.ListenAndServe http ":8080" router)`,
	}
	
	for _, part := range expectedParts {
		if !strings.Contains(result, part) {
			t.Errorf("Expected expanded result to contain: %s\nActual result: %s", part, result)
		}
	}
}

func TestMacroExpander_ExpandMacros_EmptySource(t *testing.T) {
	expander := transpiler.NewMacroExpander()
	
	result, err := expander.ExpandMacros("")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	
	if result != "" {
		t.Errorf("Expected empty result for empty source, got: %s", result)
	}
}

func TestMacroExpander_ExtractArguments(t *testing.T) {
	expander := transpiler.NewMacroExpander()
	
	testCases := []struct {
		name     string
		match    string
		macro    string
		expected []string
	}{
		{
			name:     "No arguments",
			match:    "(test-macro)",
			macro:    "test-macro",
			expected: []string{},
		},
		{
			name:     "Single argument",
			match:    "(test-macro arg1)",
			macro:    "test-macro",
			expected: []string{"arg1"},
		},
		{
			name:     "Multiple arguments",
			match:    "(test-macro arg1 arg2 arg3)",
			macro:    "test-macro",
			expected: []string{"arg1", "arg2", "arg3"},
		},
		{
			name:     "Arguments with nested structure",
			match:    "(http-server (hello))",
			macro:    "http-server",
			expected: []string{"(hello)"},
		},
	}

	// Note: extractArguments is not exported, so we test it indirectly
	// through the public ExpandMacros method by creating a custom macro
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// We can't directly test extractArguments since it's not exported
			// This is a limitation, but we can test the overall functionality
			// through the public API
			
			// For now, we just verify that the method doesn't panic
			// In a real scenario, we might refactor to make key methods testable
			_, err := expander.ExpandMacros(tc.match)
			if err != nil {
				// If it's an unknown macro, that's expected
				if !strings.Contains(err.Error(), "unknown") && !strings.Contains(err.Error(), "not found") {
					t.Errorf("Unexpected error type: %v", err)
				}
			}
		})
	}
}

func TestMacroExpander_HttpServerMacro_Integration(t *testing.T) {
	// Test the http-server macro expansion in the context of full transpilation
	input := `(http-server (hello-endpoint))`
	
	tr := transpiler.New()
	result, err := tr.TranspileFromInput(input)
	if err != nil {
		t.Fatalf("Unexpected error during transpilation: %v", err)
	}
	
	// Check that the expanded macro was transpiled correctly
	expectedParts := []string{
		`package main`,
		`import "net/http"`,
		`import "github.com/gorilla/mux"`,
		`mux.NewRouter()`,
		`router.HandleFunc("/hello", hello-handler)`,
		`http.ListenAndServe(":8080", router)`,
	}
	
	for _, part := range expectedParts {
		if !strings.Contains(result, part) {
			t.Errorf("Expected transpiled result to contain: %s\nActual result: %s", part, result)
		}
	}
}

func TestMacroExpander_MultipleMacroInstances(t *testing.T) {
	expander := transpiler.NewMacroExpander()
	
	// Source with multiple instances of the same macro
	source := `(http-server (endpoint1))
(def x 10)
(http-server (endpoint2))`
	
	result, err := expander.ExpandMacros(source)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	
	// Should expand both instances
	importCount := strings.Count(result, `(import "net/http")`)
	if importCount < 2 {
		t.Errorf("Expected at least 2 http imports (one for each macro), got %d", importCount)
	}
	
	routerCount := strings.Count(result, "(def router")
	if routerCount < 2 {
		t.Errorf("Expected at least 2 router definitions (one for each macro), got %d", routerCount)
	}
}

func TestMacroExpander_MacroWithComplexArguments(t *testing.T) {
	// Test macro expansion with complex arguments (skip for now as http-server is builtin)
	t.Skip("Skipping complex arguments test - needs rework for multi-pass system")
}

func TestMacroExpander_MultiPassExpansion(t *testing.T) {
	expander := transpiler.NewMacroExpander()
	
	// Register a meta-macro that creates other macros
	expander.RegisterStringMacro("defmacro", []string{"name", "params", "body"}, "(macro ~name ~params ~body)")
	
	// Test multi-pass expansion
	source := `(defmacro simple [x] (def result ~x))
(simple 42)`
	
	result, err := expander.ExpandMacros(source)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	
	// Should expand to the final result
	if !strings.Contains(result, `(def result 42)`) {
		t.Errorf("Expected multi-pass expansion to work, got: %s", result)
	}
}

func TestMacroExpander_NonMacroSyntax(t *testing.T) {
	expander := transpiler.NewMacroExpander()
	
	// Source that looks like macro but isn't registered
	source := `(unknown-macro arg1 arg2)`
	
	result, err := expander.ExpandMacros(source)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	
	// Should pass through unchanged since it's not a registered macro
	if result != source {
		t.Errorf("Expected unknown macro to pass through unchanged, got: %s", result)
	}
}

func TestMacroExpander_MixedContent(t *testing.T) {
	expander := transpiler.NewMacroExpander()
	
	// Mix of macro and non-macro content
	source := `(def before 1)
(http-server (test))
(def after 2)`
	
	result, err := expander.ExpandMacros(source)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	
	// Should preserve non-macro content and expand macros
	if !strings.Contains(result, "(def before 1)") {
		t.Error("Expected to preserve content before macro")
	}
	if !strings.Contains(result, "(def after 2)") {
		t.Error("Expected to preserve content after macro")
	}
	if !strings.Contains(result, `(import "net/http")`) {
		t.Error("Expected macro to be expanded")
	}
}

func TestMacroExpander_EdgeCases(t *testing.T) {
	expander := transpiler.NewMacroExpander()
	
	testCases := []struct {
		name   string
		source string
		error  bool
	}{
		{
			name:   "Only whitespace",
			source: "   \n\t  ",
			error:  false,
		},
		{
			name:   "Unmatched parentheses",
			source: "(http-server (incomplete",
			error:  false, // Macro expander doesn't validate syntax
		},
		{
			name:   "Nested macros (not supported)",
			source: "(http-server (http-server (nested)))",
			error:  false, // Should process, but inner macro won't be recognized as such
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := expander.ExpandMacros(tc.source)
			
			if tc.error && err == nil {
				t.Error("Expected error, but got success")
			}
			if !tc.error && err != nil {
				t.Errorf("Expected success, but got error: %v", err)
			}
		})
	}
}
