package transpiler

import (
	"strings"
	"testing"
)

func TestTranspiler_New(t *testing.T) {
	tr := New()
	if tr == nil {
		t.Error("Expected New() to return a non-nil transpiler")
	}
}

func TestTranspiler_SimpleExpressions(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
		hasError bool
	}{
		{
			name:     "Simple arithmetic",
			input:    "(+ 1 2)",
			expected: "_ = (1 + 2)",
			hasError: false,
		},
		{
			name:     "Variable definition",
			input:    "(def x 10)",
			expected: "var x = 10",
			hasError: false,
		},
		{
			name:     "Multiple arithmetic operands",
			input:    "(+ 1 2 3)",
			expected: "_ = ((1 + 2) + 3)",
			hasError: false,
		},
		{
			name:     "Import statement",
			input:    `(import "fmt")`,
			expected: `import "fmt"`,
			hasError: false,
		},
		{
			name:     "Function call",
			input:    "(fmt/Println \"Hello\")",
			expected: "fmt.Println(\"Hello\")",
			hasError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tr := New()
			result, err := tr.TranspileFromInput(tt.input)
			
			if tt.hasError && err == nil {
				t.Error("Expected error, but got success")
				return
			}
			
			if !tt.hasError && err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
			
			if !strings.Contains(result, tt.expected) {
				t.Errorf("Expected output to contain:\n%s\n\nActual output:\n%s", tt.expected, result)
			}
		})
	}
}

func TestTranspiler_FullProgram(t *testing.T) {
	input := `(import "fmt")
(def x (+ 5 3))
(fmt/Println "Hello World")`

	tr := New()
	result, err := tr.TranspileFromInput(input)
	
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	expectedParts := []string{
		"package main",
		"import \"fmt\"",
		"func main() {",
		"var x = (5 + 3)",
		"fmt.Println(\"Hello World\")",
		"}",
	}

	for _, expected := range expectedParts {
		if !strings.Contains(result, expected) {
			t.Errorf("Expected output to contain: %s\nActual output:\n%s", expected, result)
		}
	}
}