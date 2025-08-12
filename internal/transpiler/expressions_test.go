package transpiler

import (
	"strings"
	"testing"
)

func TestTranspiler_ExpressionEvaluation(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Arithmetic in assignment",
			input:    `(def result (+ 1 2))`,
			expected: `result := (1 + 2)`,
		},
		{
			name:     "Nested arithmetic",
			input:    `(def result (+ (* 2 3) (/ 8 4)))`,
			expected: `result := ((2 * 3) + (8 / 4))`,
		},
		{
			name:     "Function call in assignment",
			input:    `(def result (calculate 5 10))`,
			expected: `result := calculate(5, 10)`,
		},
		{
			name:     "If expression in assignment",
			input:    `(def result (if true "yes" "no"))`,
			expected: `func() interface{} { if true { return "yes" } else { return "no" } }()`,
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

func TestTranspiler_NestedExpressions(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Nested function calls",
			input:    `(def result (print (add 1 (multiply 2 3))))`,
			expected: `result := print(add(1, multiply(2, 3)))`,
		},
		{
			name:     "Nested arithmetic",
			input:    `(def result (+ (- (* 5 6) (/ 8 2)) (+ 1 2)))`,
			expected: `result := (((5 * 6) - (8 / 2)) + (1 + 2))`,
		},
		{
			name:     "Mixed expression types",
			input:    `(def result (if (> (+ 1 1) 1) "greater" "not greater"))`,
			expected: `func() interface{} { if ((1 + 1) > 1) { return "greater" } else { return "not greater" } }()`,
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

func TestTranspiler_ArithmeticExpressions(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Addition expression",
			input:    `(def x (+ 1 2))`,
			expected: `x := (1 + 2)`,
		},
		{
			name:     "Subtraction expression",
			input:    `(def x (- 10 5))`,
			expected: `x := (10 - 5)`,
		},
		{
			name:     "Multiplication expression",
			input:    `(def x (* 3 4))`,
			expected: `x := (3 * 4)`,
		},
		{
			name:     "Division expression",
			input:    `(def x (/ 8 2))`,
			expected: `x := (8 / 2)`,
		},
		{
			name:     "Chained arithmetic",
			input:    `(def x (+ 1 2 3 4))`,
			expected: `x := (((1 + 2) + 3) + 4)`,
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

func TestTranspiler_FunctionCallExpressions(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Simple function call in assignment",
			input:    `(def result (getValue))`,
			expected: `result := getValue()`,
		},
		{
			name:     "Function call with arguments",
			input:    `(def result (add 1 2))`,
			expected: `result := add(1, 2)`,
		},
		{
			name:     "Namespaced function call",
			input:    `(def result (math/sqrt 16))`,
			expected: `result := math.sqrt(16)`,
		},
		{
			name:     "Function call with array argument",
			input:    `(def result (process [1 2 3]))`,
			expected: `result := process([]interface{}{1, 2, 3})`,
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

func TestTranspiler_ConditionalExpressions(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Simple if expression",
			input:    `(def result (if true "yes" "no"))`,
			expected: `func() interface{} { if true { return "yes" } else { return "no" } }()`,
		},
		{
			name:     "If with function condition",
			input:    `(def result (if (empty? []) "empty" "not empty"))`,
			expected: `func() interface{} { if true { return "empty" } else { return "not empty" } }()`,
		},
		{
			name:     "If with arithmetic condition",
			input:    `(def result (if (> 5 3) "greater" "lesser"))`,
			expected: `func() interface{} { if (5 > 3) { return "greater" } else { return "lesser" } }()`,
		},
		{
			name:     "Nested if expressions",
			input:    `(def result (if true (if false "inner" "outer") "else"))`,
			expected: `func() interface{} { if true { return func() interface{} { if false { return "inner" } else { return "outer" } }() } else { return "else" } }()`,
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

func TestTranspiler_DoBlockExpressions(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Simple do expression",
			input:    `(def result (do "hello"))`,
			expected: `result := func() interface{} { return "hello" }()`,
		},
		{
			name:     "Do with multiple expressions",
			input:    `(def result (do (print "computing") 42))`,
			expected: ``,
		},
		{
			name:     "Do with arithmetic",
			input:    `(def result (do (+ 1 2) (* 3 4)))`,
			expected: `result := func() interface{} { (1 + 2); return (3 * 4) }()`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tr := New()
			result, err := tr.TranspileFromInput(tt.input)
			if strings.Contains(tt.name, "Do with multiple expressions") {
				if err == nil {
					t.Fatalf("Expected type error for do-block mismatch, got: %s", result)
				}
				return
			}
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if !strings.Contains(result, tt.expected) {
				t.Errorf("Expected output to contain:\n%s\n\nActual output:\n%s", tt.expected, result)
			}
		})
	}
}

func TestTranspiler_LambdaExpressions(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Simple lambda expression",
			input:    `(def f (fn [x: int] -> int x))`,
			expected: `func(x interface{}) interface{} { return x }`,
		},
		{
			name:     "Lambda with multiple parameters",
			input:    `(def add (fn [x: int y: int] -> int (+ x y)))`,
			expected: `func(x interface{}, y interface{}) interface{} { return (x + y) }`,
		},
		{
			name:     "Lambda with no parameters",
			input:    `(def greet (fn [] -> string "hello"))`,
			expected: `func() interface{} { return "hello" }`,
		},
		{
			name:     "Lambda with complex body",
			input:    `(def calc (fn [x: int] -> int (+ (* x 2) 1)))`,
			expected: `func(x interface{}) interface{} { return ((x * 2) + 1) }`,
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

func TestTranspiler_ExpressionEdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Empty do block",
			input:    `(def result (do))`,
			expected: `result := func() interface{} { return nil }()`,
		},
		{
			name:     "If with only condition and then",
			input:    `(def result (if true "yes"))`,
			expected: `func() interface{} { if true { return "yes" } else { return nil } }()`,
		},
		{
			name:     "Arithmetic with insufficient args",
			input:    `(def result (+))`,
			expected: `result := 0`,
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

func TestTranspiler_ComplexNestedExpressions(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Complex nested expression",
			input:    `(def result (+ (if (> 5 3) 10 0) (* (count [1 2 3]) 2)))`,
			expected: `result := (func() interface{} { if (5 > 3) { return 10 } else { return 0 } }() + (len([]interface{}{1, 2, 3}) * 2))`,
		},
		{
			name:     "Deep nesting with do blocks",
			input:    `(def result (do (def x (+ 1 2)) (if (> x 2) (* x 3) (/ x 2))))`,
			expected: `result := func() interface{} { def(x, (1 + 2)); return func() interface{} { if (x > 2) { return (x * 3) } else { return (x / 2) } }() }()`,
		},
		{
			name:     "Lambda in expression context",
			input:    `(def operation (fn [f x y] (f x y)))`,
			expected: `func(f interface{}, x interface{}, y interface{}) interface{} { return f(x, y) }`,
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
