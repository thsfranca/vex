package transpiler

import (
	"strings"
	"testing"
)

func TestTranspiler_CollectionOperations(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "First operation",
			input:    `(first [1 2 3])`,
			expected: `[]interface{}{1, 2, 3}[0]`,
		},
		{
			name:     "Count operation",
			input:    `(count [1 2 3])`,
			expected: `len([]interface{}{1, 2, 3})`,
		},
		{
			name:     "Empty check",
			input:    `(empty? [1 2])`,
			expected: `len([]interface{}{1, 2}) == 0`,
		},
		{
			name:     "Cons operation",
			input:    `(cons 0 [1 2 3])`,
			expected: `append([]interface{}{0}, []interface{}{1, 2, 3}...)`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tr := New()
			result, err := tr.TranspileFromInput(tt.input)
			
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
			
			if !strings.Contains(result, tt.expected) {
				t.Errorf("Expected output to contain:\n%s\n\nActual output:\n%s", tt.expected, result)
			}
		})
	}
}

func TestTranspiler_CollectionOperationsAsExpressions(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "First in variable definition",
			input:    `(def result (first [1 2 3]))`,
			expected: `first([]interface{}{1, 2, 3})`,
		},
		{
			name:     "Rest in variable definition",
			input:    `(def result (rest [1 2 3]))`,
			expected: `rest([]interface{}{1, 2, 3})`,
		},
		{
			name:     "Count in variable definition",
			input:    `(def result (count [1 2 3]))`,
			expected: `count([]interface{}{1, 2, 3})`,
		},
		{
			name:     "Empty check in variable definition",
			input:    `(def result (empty? []))`,
			expected: `empty?([]interface{}{})`,
		},
		{
			name:     "Cons in variable definition",
			input:    `(def result (cons 0 [1 2]))`,
			expected: `cons(0, []interface{}{1, 2})`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tr := New()
			result, err := tr.TranspileFromInput(tt.input)
			
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
			
			if !strings.Contains(result, tt.expected) {
				t.Errorf("Expected output to contain:\n%s\n\nActual output:\n%s", tt.expected, result)
			}
		})
	}
}

func TestTranspiler_HandleCollectionOpComprehensive(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "First with empty array",
			input:    `(first [])`,
			expected: `[]interface{}{}[0]`,
		},
		{
			name:     "First with non-empty array",
			input:    `(first [1 2 3])`,
			expected: `[]interface{}{1, 2, 3}[0]`,
		},
		{
			name:     "Rest with empty array",
			input:    `(rest [])`,
			expected: `[]interface{}{}[1:]`,
		},
		{
			name:     "Rest with single element",
			input:    `(rest [1])`,
			expected: `[]interface{}{1}[1:]`,
		},
		{
			name:     "Rest with multiple elements",
			input:    `(rest [1 2 3 4])`,
			expected: `[]interface{}{1, 2, 3, 4}[1:]`,
		},
		{
			name:     "Cons with single element",
			input:    `(cons 1 [])`,
			expected: `append([]interface{}{1}, []interface{}{}...)`,
		},
		{
			name:     "Cons with multiple elements",
			input:    `(cons 0 [1 2 3])`,
			expected: `append([]interface{}{0}, []interface{}{1, 2, 3}...)`,
		},
		{
			name:     "Count with empty array",
			input:    `(count [])`,
			expected: `len([]interface{}{})`,
		},
		{
			name:     "Count with multiple elements",
			input:    `(count [1 2 3 4 5])`,
			expected: `len([]interface{}{1, 2, 3, 4, 5})`,
		},
		{
			name:     "Empty check with empty array",
			input:    `(empty? [])`,
			expected: `len([]interface{}{}) == 0`,
		},
		{
			name:     "Empty check with non-empty array",
			input:    `(empty? [1 2])`,
			expected: `len([]interface{}{1, 2}) == 0`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tr := New()
			result, err := tr.TranspileFromInput(tt.input)
			
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
			
			if !strings.Contains(result, tt.expected) {
				t.Errorf("Expected output to contain:\n%s\n\nActual output:\n%s", tt.expected, result)
			}
		})
	}
}

func TestTranspiler_EvaluateCollectionOpComprehensive(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "First as expression",
			input:    `(def x (first [1 2 3]))`,
			expected: `func() interface{} { if len([]interface{}{1, 2, 3}) > 0 { return []interface{}{1, 2, 3}[0] } else { return nil } }()`,
		},
		{
			name:     "Rest as expression",
			input:    `(def x (rest [1 2 3]))`,
			expected: `func() []interface{} { if len([]interface{}{1, 2, 3}) > 1 { return []interface{}{1, 2, 3}[1:] } else { return []interface{}{} } }()`,
		},
		{
			name:     "Cons as expression",
			input:    `(def x (cons 1 [2 3]))`,
			expected: `append([]interface{}{1}, []interface{}{2, 3}...)`,
		},
		{
			name:     "Count as expression",
			input:    `(def x (count [1 2 3]))`,
			expected: `len([]interface{}{1, 2, 3})`,
		},
		{
			name:     "Empty as expression",
			input:    `(def x (empty? [1 2]))`,
			expected: `len([]interface{}{1, 2}) == 0`,
		},
		{
			name:     "Nested collection operations",
			input:    `(def x (first (rest [1 2 3])))`,
			expected: `func() interface{} { if len(func() []interface{} { if len([]interface{}{1, 2, 3}) > 1 { return []interface{}{1, 2, 3}[1:] } else { return []interface{}{} } }()) > 0 { return func() []interface{} { if len([]interface{}{1, 2, 3}) > 1 { return []interface{}{1, 2, 3}[1:] } else { return []interface{}{} } }()[0] } else { return nil } }()`,
		},
		{
			name:     "Collection operation in function call",
			input:    `(def x (print (count [1 2 3])))`,
			expected: `print(len([]interface{}{1, 2, 3}))`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tr := New()
			result, err := tr.TranspileFromInput(tt.input)
			
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
			
			if !strings.Contains(result, tt.expected) {
				t.Errorf("Expected output to contain:\n%s\n\nActual output:\n%s", tt.expected, result)
			}
		})
	}
}

func TestTranspiler_EvaluateCollectionOpErrorPaths(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		expectError   bool
		errorContains string
	}{
		{
			name:          "First with no arguments",
			input:         `(first)`,
			expectError:   true,
			errorContains: "macro 'first' expects 1 arguments, got 0",
		},
		{
			name:          "Rest with no arguments",
			input:         `(rest)`,
			expectError:   true,
			errorContains: "macro 'rest' expects 1 arguments, got 0",
		},
		{
			name:          "Count with no arguments",
			input:         `(count)`,
			expectError:   true,
			errorContains: "macro 'count' expects 1 arguments, got 0",
		},
		{
			name:          "Empty with no arguments",
			input:         `(empty?)`,
			expectError:   true,
			errorContains: "macro 'empty?' expects 1 arguments, got 0",
		},
		{
			name:          "Cons with insufficient arguments",
			input:         `(cons 1)`,
			expectError:   true,
			errorContains: "macro 'cons' expects 2 arguments, got 1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tr := New()
			result, err := tr.TranspileFromInput(tt.input)
			
			if tt.expectError {
				if err == nil {
					t.Fatalf("Expected error but got none. Result: %s", result)
				}
				if !strings.Contains(err.Error(), tt.errorContains) {
					t.Errorf("Expected error to contain '%s', but got: %v", tt.errorContains, err)
				}
			} else {
				if err != nil {
					t.Fatalf("Unexpected error: %v", err)
				}
			}
		})
	}
}

func TestTranspiler_CollectionOperationsWithComplexExpressions(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "First with arithmetic result",
			input:    `(first [(+ 1 2) (* 3 4)])`,
			expected: `[]interface{}{(1 + 2), (3 * 4)}[0]`,
		},
		{
			name:     "Count with nested arrays",
			input:    `(count [[1 2] [3 4]])`,
			expected: `len([]interface{}{[]interface{}{1, 2}, []interface{}{3, 4}})`,
		},
		{
			name:     "Cons with function call",
			input:    `(cons (+ 1 1) [2 3])`,
			expected: `append([]interface{}{(1 + 1)}, []interface{}{2, 3}...)`,
		},
		{
			name:     "Collection operations chained",
			input:    `(count (rest (cons 0 [1 2 3])))`,
			expected: `len(func() []interface{} { if len(append([]interface{}{0}, []interface{}{1, 2, 3}...)) > 1 { return append([]interface{}{0}, []interface{}{1, 2, 3}...)[1:] } else { return []interface{}{} } }())`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tr := New()
			result, err := tr.TranspileFromInput(tt.input)
			
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
			
			if !strings.Contains(result, tt.expected) {
				t.Errorf("Expected output to contain:\n%s\n\nActual output:\n%s", tt.expected, result)
			}
		})
	}
}