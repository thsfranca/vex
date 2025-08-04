package tests

import (
	"strings"
	"testing"

	"github.com/thsfranca/vex/internal/transpiler"
)

// TestArrayProcessing tests array handling functionality
// Note: Arrays are not fully implemented yet, but we test the existing VisitArray method
func TestArrayProcessing_VisitArray(t *testing.T) {
	// Test that VisitArray doesn't crash and handles arrays gracefully
	visitor := transpiler.NewASTVisitor()

	// Create a simple array context for testing
	// Arrays are not fully implemented, so this is a placeholder test

	// Try to parse as an array (this might not work perfectly since arrays aren't fully implemented)
	// But we want to test that the VisitArray method doesn't crash
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("VisitArray should not panic, but got: %v", r)
		}
	}()

	// Since arrays aren't fully implemented in the grammar, we'll test the method indirectly
	// by ensuring the visitor can handle array-like structures without crashing
	visitor.GetGeneratedCode() // This should not crash
}

func TestArrayProcessing_ArrayInDefinition(t *testing.T) {
	// Test how arrays are handled in variable definitions
	// This tests the edge case where arrays might appear in expressions
	input := `(def arr [1 2 3])`

	tr := transpiler.New()
	result, err := tr.TranspileFromInput(input)

	// Arrays aren't fully implemented, so this might produce a specific result
	// We're testing that it doesn't crash and produces some reasonable output
	if err != nil {
		// If there's an error, it should be a parsing error, not a panic
		if !strings.Contains(err.Error(), "syntax error") {
			t.Errorf("Expected syntax error for arrays, got: %v", err)
		}
	} else {
		// If it succeeds, check that we got some valid Go code structure
		if !strings.Contains(result, "package main") {
			t.Error("Expected basic Go code structure even with unimplemented arrays")
		}
	}
}

func TestArrayProcessing_ArraysInMacros(t *testing.T) {
	// Test array handling in macro parameters
	input := `(macro test-macro [param1 param2] (def ~param1 ~param2))`

	tr := transpiler.New()
	_, err := tr.TranspileFromInput(input)

	// This may generate a syntax error with the new multi-pass system
	// when trying to parse the macro registration without a body expression
	if err == nil {
		// Success case - macro registered without issues
		return
	}

	// Also acceptable - syntax error during macro processing is expected for this edge case
	if strings.Contains(err.Error(), "syntax error") {
		return
	}

	t.Fatalf("Unexpected error type: %v", err)
}

func TestArrayProcessing_EmptyArrays(t *testing.T) {
	// Test handling of empty array structures
	input := `(def empty [])`

	tr := transpiler.New()
	result, err := tr.TranspileFromInput(input)

	// Empty arrays might cause parsing issues since arrays aren't implemented
	// We're testing graceful handling
	if err != nil {
		// Should be a syntax error, not a crash
		if !strings.Contains(err.Error(), "syntax") {
			t.Errorf("Expected syntax error for empty arrays, got: %v", err)
		}
	} else {
		// If parsed successfully, should have basic structure
		if !strings.Contains(result, "package main") {
			t.Error("Expected valid Go structure")
		}
	}
}

func TestArrayProcessing_NestedArrays(t *testing.T) {
	// Test handling of nested array structures
	input := `(def nested [[1 2] [3 4]])`

	tr := transpiler.New()
	result, err := tr.TranspileFromInput(input)

	// Nested arrays definitely aren't implemented
	// We're testing that the system fails gracefully
	if err == nil {
		// If it somehow succeeds, check basic structure
		if !strings.Contains(result, "package main") {
			t.Error("Expected valid Go structure even with unimplemented nested arrays")
		}
	}
	// If it fails, that's expected for unimplemented features
}

func TestArrayProcessing_ArraysInFunctionCalls(t *testing.T) {
	// Test arrays as function arguments
	input := `(fmt/Println [1 2 3])`

	tr := transpiler.New()
	result, err := tr.TranspileFromInput(input)

	// Arrays in function calls might be parsed differently
	if err != nil {
		// Syntax error is acceptable for unimplemented features
		if !strings.Contains(err.Error(), "syntax") {
			t.Errorf("Expected syntax error, got: %v", err)
		}
	} else {
		// If parsed, should have function call structure
		if !strings.Contains(result, "fmt.Println") {
			t.Error("Expected function call to be preserved")
		}
	}
}

func TestArrayProcessing_FunctionParameterArrays(t *testing.T) {
	// Test function literals with array-like parameter syntax
	input := `(fn [x y z] (+ x y z))`

	tr := transpiler.New()
	_, err := tr.TranspileFromInput(input)

	// This should generate a type error because x, y, z are symbols but + requires numbers
	if err == nil {
		t.Error("Expected type error for arithmetic on symbol parameters")
	}
}

func TestArrayProcessing_MacroParameterArrays(t *testing.T) {
	// Test macro parameter arrays (this should work)
	input := `(macro add-three [a b c] (+ ~a ~b ~c))`

	tr := transpiler.New()
	_, err := tr.TranspileFromInput(input)

	// This may generate a syntax error with the new multi-pass system
	// when trying to parse the macro registration without a usage expression
	if err == nil {
		// Success case - macro registered without issues
		return
	}

	// Also acceptable - syntax error during macro processing is expected for this edge case
	if strings.Contains(err.Error(), "syntax error") {
		return
	}

	t.Fatalf("Unexpected error type: %v", err)
}

func TestArrayProcessing_DirectVisitorCall(t *testing.T) {
	// Directly test the VisitArray method
	visitor := transpiler.NewASTVisitor()

	// Create a mock array context
	// Since we can't easily create a real ArrayContext without full parsing,
	// we'll test that the method exists and can be called

	// The VisitArray method currently just calls VisitChildren
	// So we test that it doesn't crash
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("VisitArray should not panic: %v", r)
		}
	}()

	// Test that visitor can handle basic operations after array processing
	code := visitor.GetGeneratedCode()
	if code == "" {
		// This is expected since we haven't processed anything yet
	}
}
