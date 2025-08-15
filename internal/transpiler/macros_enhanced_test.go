package transpiler

import (
	"strings"
	"testing"
)

// Test the enhanced macro parameter substitution functionality
func TestTranspiler_SafeParameterSubstitution(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expected    []string
		notExpected []string // Things that should NOT appear (unsafe replacements)
	}{
		{
			name: "Safe parameter substitution - no partial matches",
			input: `(macro test [x] (+ x x))
(test 5)`,
			expected: []string{
				"(5 + 5)", // Should replace 'x' but not affect '+'
			},
			notExpected: []string{
				"(5 5 5)", // Should NOT happen with string replacement
			},
		},
		{
			name: "Parameter inside strings should not be replaced",
			input: `(macro greet [name] (fmt/Println "Hello name, welcome!"))
(greet "Alice")`,
			expected: []string{
				`"Hello name, welcome!"`, // 'name' in string should stay
			},
			notExpected: []string{
				`"Hello Alice, welcome!"`, // Should NOT replace inside strings
			},
		},
		{
			name: "Multiple parameters with overlapping names",
			input: `(macro test [x xx] (+ x xx))
(test 1 2)`,
			expected: []string{
				"(1 + 2)",
			},
			notExpected: []string{
				"(1 + 12)", // Should not replace 'x' inside 'xx'
				"(11 + 2)", // Should not replace 'x' inside 'xx'
			},
		},
		{
			name: "Parameter as part of larger symbol",
			input: `(def x-value 10)
(macro test [x] (+ x-value x))
(test 42)`,
			expected: []string{
				"(x_value + 42)", // 'x-value' should not become '42-value'
			},
			notExpected: []string{
				"(42-value + 42)",
			},
		},
		{
			name: "Nested expressions with parameters",
			input: `(macro calc [a b] (+ (* a a) (* b b)))
(calc 3 4)`,
			expected: []string{
				"((3 * 3) + (4 * 4))",
			},
		},
		{
			name: "Parameter used in function position",
			input: `(macro call-fn [fn arg] (fn arg))
(call-fn fmt/Println "test")`,
			expected: []string{
				`fmt.Println("test")`,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tr, _ := NewBuilder().Build()
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

			// Check that unsafe replacements didn't happen
			for _, notExpected := range tt.notExpected {
				if strings.Contains(result, notExpected) {
					t.Errorf("Output should NOT contain (unsafe replacement):\n%s\n\nActual output:\n%s", notExpected, result)
				}
			}
		})
	}
}

func TestTranspiler_MacroSymbolBoundaries(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name: "Symbol boundaries respected",
			input: `(def max-value 100)
(def a-min 1)
(macro test [a] (+ max-value a a-min))
(test 10)`,
			expected: "((max_value + 10) + a_min)", // Only standalone 'a' should be replaced
		},
		{
			name: "Hyphenated parameters",
			input: `(macro test [param-name] (len param-name))
(test "value")`,
			expected: `len("value")`,
		},
		{
			name: "Parameters with special characters",
			input: `(macro test [is-valid] (if is-valid "yes" "no"))
(test true)`,
			expected: `func() interface{} { if true { return "yes" } else { return "no" } }()`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tr, _ := NewBuilder().Build()
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

func TestTranspiler_MacroNestedStructures(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name: "Parameters in nested lists",
			input: `(macro test [x y] (+ (+ x 1) (- y 1)))
(test 5 10)`,
			expected: []string{
				"((5 + 1) + (10 - 1))",
			},
		},
		{
			name: "Parameters in arrays",
			input: `(macro make-array [a b c] [a b c])
(make-array 1 2 3)`,
			expected: []string{
				"[]interface{}{1, 2, 3}",
			},
		},
		{
			name: "Mixed nested structures",
			input: `(macro complex [x] (do (def arr [x x]) (len arr)))
(complex "value")`,
			expected: []string{
				`arr := []interface{}{"value", "value"}`,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tr, _ := NewBuilder().Build()
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

func TestTranspiler_MacroEdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{

		{
			name: "Parameter same as built-in function",
			input: `(macro test [def] (+ def 1))
(test 5)`,
			expected: []string{
				"(5 + 1)", // Updated to match actual Go infix output
			},
		},
		{
			name: "Parameter that matches macro name",
			input: `(macro test [test] (len test))
(test "value")`,
			expected: []string{
				`len("value")`, // Updated to match actual function call output
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tr, _ := NewBuilder().Build()
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
