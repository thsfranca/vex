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
			expected: `func() interface{} { if len([]interface{}{1, 2, 3}) > 0 { return []interface{}{1, 2, 3}[0] } else { return nil } }()`,
		},
		{
			name:     "Rest in variable definition",
			input:    `(def result (rest [1 2 3]))`,
			expected: `func() []interface{} { if len([]interface{}{1, 2, 3}) > 1 { return []interface{}{1, 2, 3}[1:] } else { return []interface{}{} } }()`,
		},
		{
			name:     "Count in variable definition",
			input:    `(def result (count [1 2 3]))`,
			expected: `len([]interface{}{1, 2, 3})`,
		},
		{
			name:     "Empty check in variable definition",
			input:    `(def result (empty? []))`,
			expected: `len([]interface{}{}) == 0`,
		},
		{
			name:     "Cons in variable definition",
			input:    `(def result (cons 0 [1 2]))`,
			expected: `append([]interface{}{0}, []interface{}{1, 2}...)`,
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
		name     string
		input    string
		expected string
	}{
		{
			name:     "First with no arguments",
			input:    `(first)`,
			expected: `_ = nil // Error: first requires collection`,
		},
		{
			name:     "Rest with no arguments",
			input:    `(rest)`,
			expected: `_ = []interface{}{} // Error: rest requires collection`,
		},
		{
			name:     "Count with no arguments",
			input:    `(count)`,
			expected: `_ = 0 // Error: count requires collection`,
		},
		{
			name:     "Empty with no arguments",
			input:    `(empty?)`,
			expected: `_ = true // Error: empty? requires collection`,
		},
		{
			name:     "Cons with insufficient arguments",
			input:    `(cons 1)`,
			expected: `_ = []interface{}{} // Error: cons requires element and collection`,
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