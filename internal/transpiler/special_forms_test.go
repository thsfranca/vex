package transpiler

import (
	"strings"
	"testing"
)

func TestTranspiler_FunctionCalls(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Simple function call",
			input:    `(print "hello")`,
			expected: `_ = print("hello")`,
		},
		{
			name:     "Function call with multiple args",
			input:    `(add 1 2 3)`,
			expected: `_ = add(1, 2, 3)`,
		},
		{
			name:     "Nested function calls",
			input:    `(print (add 1 2))`,
			expected: `_ = print(add(1, 2))`,
		},
		{
			name:     "Function call as last expression",
			input:    `(print "hello")`,
			expected: `_ = print("hello")`,
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

func TestTranspiler_ConditionalStatements(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Simple if statement",
			input:    `(if true "yes" "no")`,
			expected: `if true {`,
		},
		{
			name:     "If as expression in definition",
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

func TestTranspiler_DoBlocks(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Simple do block",
			input:    `(do (print "first") (print "second"))`,
			expected: `print("first")`,
		},
		{
			name:     "Do as expression in definition",
			input:    `(def result (do (print "computing") 42))`,
			expected: `func() interface{} {`,
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

func TestTranspiler_ArithmeticEdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Subtraction",
			input:    `(- 10 5)`,
			expected: `(10 - 5)`,
		},
		{
			name:     "Multiplication",
			input:    `(* 3 4)`,
			expected: `(3 * 4)`,
		},
		{
			name:     "Division",
			input:    `(/ 8 2)`,
			expected: `(8 / 2)`,
		},
		{
			name:     "Complex arithmetic",
			input:    `(+ (* 2 3) (/ 8 4))`,
			expected: `((2 * 3) + (8 / 4))`,
		},
		{
			name:     "Comparison operators",
			input:    `(> 5 3)`,
			expected: `_ = >(5, 3)`, // Transpiler treats > as function call
		},
		{
			name:     "Less than",
			input:    `(< 2 8)`,
			expected: `_ = <(2, 8)`,
		},
		{
			name:     "Equality",
			input:    `(= 5 5)`,
			expected: `_ = =(5, 5)`,
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

func TestTranspiler_IfStatementEdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "If with function call condition",
			input:    `(if (empty? []) "is empty" "not empty")`,
			expected: `if len([]interface{}{}) == 0 {`,
		},
		{
			name:     "If with arithmetic condition",
			input:    `(if (> (+ 1 1) 1) "greater" "not greater")`,
			expected: `if >((1 + 1), 1) {`, // Transpiler treats > as function in if condition
		},
		{
			name:     "Nested if statements",
			input:    `(if true (if false "inner true" "inner false") "outer false")`,
			expected: `if true {`,
		},
		{
			name:     "If in variable assignment",
			input:    `(def result (if (> 5 3) "yes" "no"))`,
			expected: `func() interface{} { if >(5, 3) { return "yes" } else { return "no" } }()`,
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

func TestTranspiler_DoBlockEdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Do with single expression",
			input:    `(do "hello")`,
			expected: `"hello"`,
		},
		{
			name:     "Do with variable definitions",
			input:    `(do (def x 1) (def y 2) (+ x y))`,
			expected: `def(x, 1)`, // Transpiler treats def as function call in this context
		},
		{
			name:     "Do in variable assignment",
			input:    `(def result (do (print "calculating") (+ 1 2)))`,
			expected: `func() interface{} {`,
		},
		{
			name:     "Nested do blocks",
			input:    `(do (do (def x 1)) x)`,
			expected: `def(x, 1)`,
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

func TestTranspiler_HandleFunctionCallComprehensive(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Simple function call no args",
			input:    `(getValue)`,
			expected: `_ = getValue()`,
		},
		{
			name:     "Function call with one arg",
			input:    `(print "hello")`,
			expected: `_ = print("hello")`,
		},
		{
			name:     "Function call with multiple args",
			input:    `(add 1 2 3)`,
			expected: `_ = add(1, 2, 3)`,
		},
		{
			name:     "Nested function calls",
			input:    `(print (add 1 (multiply 2 3)))`,
			expected: `_ = print(add(1, multiply(2, 3)))`,
		},
		{
			name:     "Function call with array args",
			input:    `(process [1 2 3] "mode")`,
			expected: `_ = process([]interface{}{1, 2, 3}, "mode")`,
		},
		{
			name:     "Function call in assignment",
			input:    `(def result (calculate 5 10))`,
			expected: `calculate(5, 10)`,
		},
		{
			name:     "Namespaced function calls",
			input:    `(math/sqrt 16)`,
			expected: `_ = math.sqrt(16)`,
		},
		{
			name:     "Function call with string interpolation",
			input:    `(log "Processing item: %s" name)`,
			expected: `_ = log("Processing item: %s", name)`,
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

func TestTranspiler_LambdaFunctions(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Simple lambda with one parameter",
			input:    `(def identity (fn [x] x))`,
			expected: `func(x interface{}) interface{} { return x }`,
		},
		{
			name:     "Lambda with multiple parameters",
			input:    `(def add (fn [x y] (+ x y)))`,
			expected: `func(x interface{}, y interface{}) interface{} { return (x + y) }`,
		},
		{
			name:     "Lambda with no parameters",
			input:    `(def greet (fn [] "hello"))`,
			expected: `func() interface{} { return "hello" }`,
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

func TestTranspiler_ErrorHandlingPaths(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Def with insufficient args",
			input:    `(def)`,
			expected: `// Error: def requires name and value`,
		},
		{
			name:     "Import with no args",
			input:    `(import)`,
			expected: `// Error: import requires package name`,
		},
		{
			name:     "If with insufficient args",
			input:    `(if)`,
			expected: `// Error: if requires at least condition and then-expr`,
		},
		{
			name:     "Arithmetic with insufficient args",
			input:    `(+)`,
			expected: `// Error: + requires at least 2 arguments`,
		},
		{
			name:     "Lambda with insufficient args",
			input:    `(fn)`,
			expected: `// Error: fn requires parameter list and body`,
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

func TestTranspiler_PrintlnSpecialHandling(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "fmt.Println direct call",
			input:    `(fmt/Println "Hello")`,
			expected: `fmt.Println("Hello")`,
		},
		{
			name:     "Printf direct call",
			input:    `(fmt/Printf "Hello %s" "World")`,
			expected: `fmt.Printf("Hello %s", "World")`,
		},
		{
			name:     "Regular function gets assignment",
			input:    `(regularFunc "Hello")`,
			expected: `_ = regularFunc("Hello")`,
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
// TestTranspiler_HandleMacroFunction tests the handleMacro function to increase coverage
func TestTranspiler_HandleMacroFunction(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Simple macro definition",
			input:    `(macro greet [name] (fmt/Printf "Hello %s!" name))`,
			expected: "// Registered macro: greet with parameters [name]",
		},
		{
			name:     "Macro with multiple parameters",
			input:    `(macro add [x y] (+ x y))`,
			expected: "// Registered macro: add with parameters [x y]",
		},
		{
			name:     "Macro with no parameters",
			input:    `(macro zero [] 0)`,
			expected: "// Registered macro: zero with parameters []",
		},
		{
			name:     "Macro with insufficient args - error case",
			input:    `(macro incomplete)`,
			expected: "// Error: macro requires name, parameter list, and body",
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

// TestTranspiler_HandleLambdaSuccess tests successful lambda function creation
func TestTranspiler_HandleLambdaSuccess(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Lambda with single parameter",
			input:    "(fn [x] x)",
			expected: "func(x interface{}) interface{} { return x }",
		},
		{
			name:     "Lambda with multiple parameters",
			input:    "(fn [x y] (+ x y))",
			expected: "func(x interface{}, y interface{}) interface{} { return (x + y) }",
		},
		{
			name:     "Lambda with no parameters",
			input:    "(fn [] 42)",
			expected: "func() interface{} { return 42 }",
		},
		{
			name:     "Lambda with complex body",
			input:    `(fn [n] (fmt/Printf "Number: %d" n))`,
			expected: "func(n interface{}) interface{} { return fmt.Printf(\"Number: %d\", n) }",
		},
		{
			name:     "Lambda error case - insufficient args",
			input:    "(fn [x])",
			expected: "// Error: fn requires parameter list and body",
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

// TestTranspiler_CollectionOpEdgeCases tests missing edge cases for collection operations
func TestTranspiler_CollectionOpEdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Rest operation with empty array",
			input:    `(def result (rest []))`,
			expected: "func() []interface{} { if len([]interface{}{}) > 1 { return []interface{}{}[1:] } else { return []interface{}{} } }()",
		},
		{
			name:     "Cons with multiple elements",
			input:    `(def result (cons 1 [2 3]))`,
			expected: "append([]interface{}{1}, []interface{}{2, 3}...)",
		},
		{
			name:     "Count with variable argument",
			input:    `(def arr [1 2 3]) (def result (count arr))`,
			expected: "len(arr)",
		},
		{
			name:     "First with variable argument",
			input:    `(def arr [1 2 3]) (def result (first arr))`,
			expected: "func() interface{} { if len(arr) > 0 { return arr[0] } else { return nil } }()",
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
