package transpiler

import (
	"strings"
	"testing"
)

// Test the enhanced defn macro functionality
func TestTranspiler_DefnMacro(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expected    []string
		notExpected []string
	}{
		{
			name: "Simple function definition",
			input: `(defn add [a: int b: int] -> int (+ a b))
(add 3 4)`,
			expected: []string{
				"add := func(a interface{}, b interface{}) interface{} { return (a + b) }",
				"_ = add(3, 4)",
			},
		},
		{
			name: "Function with single parameter",
			input: `(defn square [x: int] -> int (* x x))
(square 5)`,
			expected: []string{
				"square := func(x interface{}) interface{} { return (x * x) }",
				"_ = square(5)",
			},
		},
		{
			name: "Function with no parameters",
			input: `(defn get-answer [] -> int 42)
(get-answer)`,
			expected: []string{
				"get_answer := func() interface{} { return 42 }",
				"_ = get-answer()",
			},
		},
		{
			name: "Function with string body",
			input: `(defn greet [name: string] -> string (fmt/Sprintf "Hello %s" name))
(greet "World")`,
			expected: []string{
				"greet := func(name interface{}) interface{} { return fmt.Sprintf(\"Hello %s\", name) }",
				`_ = greet("World")`,
			},
		},
		{
			name: "Multiple function definitions",
			input: `(defn add [a: int b: int] -> int (+ a b))
(defn multiply [x: int y: int] -> int (* x y))
(add 1 2)
(multiply 3 4)`,
			expected: []string{
				"add := func(a interface{}, b interface{}) interface{} { return (a + b) }",
				"multiply := func(x interface{}, y interface{}) interface{} { return (x * y) }",
				"_ = add(1, 2)",
				"_ = multiply(3, 4)",
			},
		},
		{
			name: "Function calling another function",
			input: `(defn double [x: int] -> int (* x 2))
(defn quadruple [x: int] -> int (double (double x)))
(quadruple 5)`,
			expected: []string{
				"double := func(x interface{}) interface{} { return (x * 2) }",
				"quadruple := func(x interface{}) interface{} { return double(double(x)) }",
				"_ = quadruple(5)",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tr := New()
			result, err := tr.TranspileFromInput(tt.input)

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			// Check expected content
			for _, expected := range tt.expected {
				if !strings.Contains(result, expected) {
					t.Errorf("Expected output to contain:\n%s\n\nActual output:\n%s", expected, result)
				}
			}

			// Check that unwanted content is not present
			for _, notExpected := range tt.notExpected {
				if strings.Contains(result, notExpected) {
					t.Errorf("Output should NOT contain:\n%s\n\nActual output:\n%s", notExpected, result)
				}
			}
		})
	}
}

func TestTranspiler_DefnMacroTypeInference(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:  "Integer arithmetic function",
			input: `(defn calculate [a: int b: int] -> int (+ (* a 2) b))`,
			expected: []string{
				"calculate := func(a interface{}, b interface{}) interface{} { return ((a * 2) + b) }",
			},
		},
		{
			name:  "String manipulation function",
			input: `(defn concat [s1: string s2: string] -> string (+ s1 s2))`,
			expected: []string{
				"concat := func(s1 interface{}, s2 interface{}) interface{} { return (s1 + s2) }",
			},
		},
		{
			name:  "Boolean function",
			input: `(defn is-positive [x: int] -> bool (> x 0))`,
			expected: []string{
				"is_positive := func(x interface{}) interface{} { return (x > 0) }",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tr := New()
			result, err := tr.TranspileFromInput(tt.input)

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			for _, expected := range tt.expected {
				if !strings.Contains(result, expected) {
					t.Errorf("Expected output to contain:\n%s\n\nActual output:\n%s", expected, result)
				}
			}
		})
	}
}

func TestTranspiler_DefnMacroErrorHandling(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		expectedError bool
		errorMessage  string
	}{
		{
			name:          "defn with missing function name",
			input:         `(defn [x] (* x x))`,
			expectedError: true,
			errorMessage:  "defn requires function name",
		},
		{
			name:          "defn with missing parameter list",
			input:         `(defn square (* x x))`,
			expectedError: true,
			errorMessage:  "defn requires parameter list",
		},
		{
			name:          "defn with missing body",
			input:         `(defn square [x])`,
			expectedError: true,
			errorMessage:  "defn requires function body",
		},
		{
			name:          "defn with invalid parameter list",
			input:         `(defn square "not-a-list" (* x x))`,
			expectedError: true,
			errorMessage:  "defn requires parameter list",
		},
		{
			name:          "defn with snake_case function name",
			input:         `(defn my_function [x: int] -> int (* x 2))`,
			expectedError: true,
			errorMessage:  "function names must use kebab-case",
		},
		{
			name:          "defn with valid kebab-case function name",
			input:         `(defn my-function [x: int] -> int (* x 2))`,
			expectedError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tr := New()
			result, err := tr.TranspileFromInput(tt.input)

			if tt.expectedError {
				// Check if error occurred or error message appears in output
				if err == nil && !strings.Contains(result, tt.errorMessage) {
					t.Errorf("Expected error or error message '%s' but got result: %s", tt.errorMessage, result)
				}
			} else {
				if err != nil {
					t.Fatalf("Unexpected error: %v", err)
				}
			}
		})
	}
}

func TestTranspiler_DefnMacroComplexBodies(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:  "Function with conditional body",
			input: `(defn abs [x: int] -> int (if (< x 0) (- x) x))`,
			expected: []string{
				"abs := func(x interface{}) interface{} { return func() interface{} { if (x < 0) { return 0 } else { return x } }() }",
			},
		},
		{
			name:  "Function with nested operations",
			input: `(defn complex [a: int b: int c: int] -> int (+ (* a b) (/ c 2)))`,
			expected: []string{
				"complex := func(a interface{}, b interface{}, c interface{}) interface{} { return ((a * b) + (c / 2)) }",
			},
		},
		{
			name:  "Function with array operations",
			input: `(defn get-first [arr: [int]] -> int (get arr 0))`,
			expected: []string{
				"get_first := func(arr interface{}) interface{} { return func() interface{} { if len(arr) > 0 { return arr[0] } else",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tr := New()
			result, err := tr.TranspileFromInput(tt.input)

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			for _, expected := range tt.expected {
				if !strings.Contains(result, expected) {
					t.Errorf("Expected output to contain:\n%s\n\nActual output:\n%s", expected, result)
				}
			}
		})
	}
}
