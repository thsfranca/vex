package tests

import (
	"strings"
	"testing"

	"github.com/thsfranca/vex/internal/transpiler"
)

func TestNewCodeGenerator(t *testing.T) {
	cg := transpiler.NewCodeGenerator()
	if cg == nil {
		t.Error("Expected NewCodeGenerator() to return a non-nil code generator")
	}
	
	// Should start with empty code
	if code := cg.GetCode(); code != "" {
		t.Errorf("Expected empty code initially, got: %s", code)
	}
	
	// Should start with no imports
	if imports := cg.GetImports(); len(imports) != 0 {
		t.Errorf("Expected no imports initially, got: %v", imports)
	}
}

func TestCodeGenerator_EmitNumber(t *testing.T) {
	cg := transpiler.NewCodeGenerator()
	cg.EmitNumber("42")
	
	code := cg.GetCode()
	expected := "_ = 42\n"
	
	if code != expected {
		t.Errorf("Expected: %q, got: %q", expected, code)
	}
}

func TestCodeGenerator_EmitString(t *testing.T) {
	cg := transpiler.NewCodeGenerator()
	cg.EmitString(`"hello world"`)
	
	code := cg.GetCode()
	expected := `_ = "hello world"` + "\n"
	
	if code != expected {
		t.Errorf("Expected: %q, got: %q", expected, code)
	}
}

func TestCodeGenerator_EmitSymbol(t *testing.T) {
	cg := transpiler.NewCodeGenerator()
	cg.EmitSymbol("myVariable")
	
	code := cg.GetCode()
	expected := "_ = myVariable\n"
	
	if code != expected {
		t.Errorf("Expected: %q, got: %q", expected, code)
	}
}

func TestCodeGenerator_EmitVariableDefinition(t *testing.T) {
	cg := transpiler.NewCodeGenerator()
	cg.EmitVariableDefinition("x", "42")
	
	code := cg.GetCode()
	expected := "x := 42\n"
	
	if code != expected {
		t.Errorf("Expected: %q, got: %q", expected, code)
	}
}

func TestCodeGenerator_EmitExpressionStatement(t *testing.T) {
	cg := transpiler.NewCodeGenerator()
	cg.EmitExpressionStatement("1 + 2")
	
	code := cg.GetCode()
	expected := "_ = 1 + 2\n"
	
	if code != expected {
		t.Errorf("Expected: %q, got: %q", expected, code)
	}
}

func TestCodeGenerator_EmitArithmeticExpression(t *testing.T) {
	testCases := []struct {
		name      string
		operator  string
		operands  []string
		expected  string
	}{
		{
			name:     "Addition with two operands",
			operator: "+",
			operands: []string{"1", "2"},
			expected: "_ = 1 + 2\n",
		},
		{
			name:     "Subtraction",
			operator: "-",
			operands: []string{"10", "3"},
			expected: "_ = 10 - 3\n",
		},
		{
			name:     "Multiplication",
			operator: "*",
			operands: []string{"4", "5"},
			expected: "_ = 4 * 5\n",
		},
		{
			name:     "Division",
			operator: "/",
			operands: []string{"12", "3"},
			expected: "_ = 12 / 3\n",
		},
		{
			name:     "Multiple operands",
			operator: "+",
			operands: []string{"1", "2", "3", "4"},
			expected: "_ = 1 + 2 + 3 + 4\n",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cg := transpiler.NewCodeGenerator()
			cg.EmitArithmeticExpression(tc.operator, tc.operands)
			
			code := cg.GetCode()
			if code != tc.expected {
				t.Errorf("Expected: %q, got: %q", tc.expected, code)
			}
		})
	}
}

func TestCodeGenerator_EmitArithmeticExpression_InvalidOperands(t *testing.T) {
	cg := transpiler.NewCodeGenerator()
	cg.EmitArithmeticExpression("+", []string{"1"}) // Only one operand
	
	code := cg.GetCode()
	if !strings.Contains(code, "Invalid arithmetic expression") {
		t.Errorf("Expected error message for invalid operands, got: %q", code)
	}
}

func TestCodeGenerator_EmitImport(t *testing.T) {
	cg := transpiler.NewCodeGenerator()
	cg.EmitImport(`"fmt"`)
	
	imports := cg.GetImports()
	if len(imports) != 1 {
		t.Errorf("Expected 1 import, got %d", len(imports))
	}
	
	if imports[0] != `"fmt"` {
		t.Errorf("Expected import %q, got %q", `"fmt"`, imports[0])
	}
}

func TestCodeGenerator_EmitImport_DuplicateImports(t *testing.T) {
	cg := transpiler.NewCodeGenerator()
	cg.EmitImport(`"fmt"`)
	cg.EmitImport(`"fmt"`) // Duplicate
	cg.EmitImport(`"net/http"`)
	
	imports := cg.GetImports()
	if len(imports) != 2 {
		t.Errorf("Expected 2 unique imports, got %d: %v", len(imports), imports)
	}
}

func TestCodeGenerator_EmitImport_CleanPath(t *testing.T) {
	cg := transpiler.NewCodeGenerator()
	cg.EmitImport("fmt") // Without quotes
	
	imports := cg.GetImports()
	if len(imports) != 1 {
		t.Errorf("Expected 1 import, got %d", len(imports))
	}
	
	if imports[0] != `"fmt"` {
		t.Errorf("Expected import with quotes %q, got %q", `"fmt"`, imports[0])
	}
}

func TestCodeGenerator_EmitMethodCall(t *testing.T) {
	testCases := []struct {
		name     string
		receiver string
		method   string
		args     []string
		expected string
	}{
		{
			name:     "Method with no args",
			receiver: "conn",
			method:   "Close",
			args:     []string{},
			expected: "_ = conn.Close()\n",
		},
		{
			name:     "Method with one arg",
			receiver: "w",
			method:   "Write",
			args:     []string{`"data"`},
			expected: `_ = w.Write("data")` + "\n",
		},
		{
			name:     "Method with multiple args",
			receiver: "router",
			method:   "HandleFunc",
			args:     []string{`"/path"`, "handler"},
			expected: `_ = router.HandleFunc("/path", handler)` + "\n",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cg := transpiler.NewCodeGenerator()
			cg.EmitMethodCall(tc.receiver, tc.method, tc.args)
			
			code := cg.GetCode()
			if code != tc.expected {
				t.Errorf("Expected: %q, got: %q", tc.expected, code)
			}
		})
	}
}

func TestCodeGenerator_EmitSlashNotationCall(t *testing.T) {
	testCases := []struct {
		name        string
		packageName string
		funcName    string
		args        []string
		expected    string
	}{
		{
			name:        "Function with no args",
			packageName: "os",
			funcName:    "Exit",
			args:        []string{},
			expected:    "os.Exit()\n",
		},
		{
			name:        "Function with one arg",
			packageName: "fmt",
			funcName:    "Println",
			args:        []string{`"hello"`},
			expected:    `fmt.Println("hello")` + "\n",
		},
		{
			name:        "Function with multiple args",
			packageName: "fmt",
			funcName:    "Printf",
			args:        []string{`"%s %d"`, `"hello"`, "42"},
			expected:    `fmt.Printf("%s %d", "hello", 42)` + "\n",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cg := transpiler.NewCodeGenerator()
			cg.EmitSlashNotationCall(tc.packageName, tc.funcName, tc.args)
			
			code := cg.GetCode()
			if code != tc.expected {
				t.Errorf("Expected: %q, got: %q", tc.expected, code)
			}
		})
	}
}

func TestCodeGenerator_EmitFunctionLiteral_HTTPHandler(t *testing.T) {
	cg := transpiler.NewCodeGenerator()
	visitor := transpiler.NewASTVisitor() // We need a visitor for body processing
	
	params := []string{"w", "r"}
	
	// This is a simplified test with empty body elements (normally AST nodes)
	result := cg.EmitFunctionLiteral(params, nil, visitor)
	
	expectedSubstrings := []string{
		"func(w http.ResponseWriter, r *http.Request) {",
		"}",
	}
	
	for _, expected := range expectedSubstrings {
		if !strings.Contains(result, expected) {
			t.Errorf("Expected result to contain: %q\nActual result: %q", expected, result)
		}
	}
	
	// Check that http import was added
	imports := cg.GetImports()
	found := false
	for _, imp := range imports {
		if strings.Contains(imp, "net/http") {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected net/http import to be added")
	}
}

func TestCodeGenerator_EmitFunctionLiteral_Generic(t *testing.T) {
	cg := transpiler.NewCodeGenerator()
	visitor := transpiler.NewASTVisitor()
	
	params := []string{"x", "y", "z"}
	
	result := cg.EmitFunctionLiteral(params, nil, visitor)
	
	expectedSubstrings := []string{
		"func(x interface{}, y interface{}, z interface{}) {",
		"}",
	}
	
	for _, expected := range expectedSubstrings {
		if !strings.Contains(result, expected) {
			t.Errorf("Expected result to contain: %q\nActual result: %q", expected, result)
		}
	}
}

func TestCodeGenerator_IncreaseDecreaseIndent(t *testing.T) {
	cg := transpiler.NewCodeGenerator()
	
	// Initial level (no indentation)
	cg.EmitNumber("1")
	
	// Increase indent
	cg.IncreaseIndent()
	cg.EmitNumber("2")
	
	// Increase again
	cg.IncreaseIndent()
	cg.EmitNumber("3")
	
	// Decrease once
	cg.DecreaseIndent()
	cg.EmitNumber("4")
	
	// Decrease to zero
	cg.DecreaseIndent()
	cg.EmitNumber("5")
	
	code := cg.GetCode()
	lines := strings.Split(strings.TrimSpace(code), "\n")
	
	if len(lines) != 5 {
		t.Errorf("Expected 5 lines, got %d", len(lines))
	}
	
	expectedIndents := []int{0, 1, 2, 1, 0}
	for i, line := range lines {
		actualIndent := 0
		for _, char := range line {
			if char == '\t' {
				actualIndent++
			} else {
				break
			}
		}
		
		if actualIndent != expectedIndents[i] {
			t.Errorf("Line %d: expected %d tabs, got %d. Line: %q", i+1, expectedIndents[i], actualIndent, line)
		}
	}
}

func TestCodeGenerator_DecreaseIndent_BelowZero(t *testing.T) {
	cg := transpiler.NewCodeGenerator()
	
	// Try to decrease below zero
	cg.DecreaseIndent()
	cg.DecreaseIndent()
	
	// Should still emit at level 0
	cg.EmitNumber("42")
	
	code := cg.GetCode()
	if strings.HasPrefix(code, "\t") {
		t.Error("Expected no indentation when decreasing below zero")
	}
}

func TestCodeGenerator_Reset(t *testing.T) {
	cg := transpiler.NewCodeGenerator()
	
	// Add some content
	cg.EmitNumber("42")
	cg.EmitImport(`"fmt"`)
	cg.IncreaseIndent()
	
	// Verify content exists
	if code := cg.GetCode(); code == "" {
		t.Error("Expected non-empty code before reset")
	}
	if imports := cg.GetImports(); len(imports) == 0 {
		t.Error("Expected imports before reset")
	}
	
	// Reset
	cg.Reset()
	
	// Verify everything is cleared
	if code := cg.GetCode(); code != "" {
		t.Errorf("Expected empty code after reset, got: %q", code)
	}
	if imports := cg.GetImports(); len(imports) != 0 {
		t.Errorf("Expected no imports after reset, got: %v", imports)
	}
	
	// Should be able to use normally after reset
	cg.EmitNumber("100")
	if code := cg.GetCode(); !strings.Contains(code, "100") {
		t.Error("Expected to work normally after reset")
	}
}
