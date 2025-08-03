package tests

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/thsfranca/vex/internal/transpiler"
)

func TestTranspiler_New(t *testing.T) {
	tr := transpiler.New()
	if tr == nil {
		t.Error("Expected New() to return a non-nil transpiler")
	}
}

func TestTranspiler_TranspileFromInput_SimpleExpressions(t *testing.T) {
	testCases := []TestCase{
		{
			Name:     "Number literal",
			Input:    "(42)",
			Expected: "_ = 42",
			Error:    false,
		},
		{
			Name:     "String literal", 
			Input:    `("hello")`,
			Expected: `_ = "hello"`,
			Error:    false,
		},
		{
			Name:     "Simple arithmetic",
			Input:    "(+ 1 2)",
			Expected: "_ = 1 + 2",
			Error:    false,
		},
		{
			Name:     "Variable definition",
			Input:    "(def x 10)",
			Expected: "x := 10",
			Error:    false,
		},
		{
			Name:     "Multiple arithmetic operands",
			Input:    "(+ 1 2 3)",
			Expected: "_ = 1 + 2 + 3",
			Error:    false,
		},
	}

	RunTestCases(t, testCases)
}

func TestTranspiler_TranspileFromInput_ImportStatements(t *testing.T) {
	input := `(import "fmt")`
	tr := transpiler.New()
	
	result, err := tr.TranspileFromInput(input)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	
	AssertContainsAll(t, result, 
		`package main`,
		`import "fmt"`,
		`func main() {`,
	)
}

func TestTranspiler_TranspileFromInput_SlashNotation(t *testing.T) {
	input := `(fmt/Println "Hello World")`
	tr := transpiler.New()
	
	result, err := tr.TranspileFromInput(input)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	
	AssertContainsAll(t, result,
		`fmt.Println("Hello World")`,
	)
}

func TestTranspiler_TranspileFromInput_MethodCalls(t *testing.T) {
	input := `(.HandleFunc router "/hello" handler)`
	tr := transpiler.New()
	
	result, err := tr.TranspileFromInput(input)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	
	AssertContainsAll(t, result,
		`_ = router.HandleFunc("/hello", handler)`,
	)
}

func TestTranspiler_TranspileFromInput_FunctionLiterals(t *testing.T) {
	input := `(fn [w r] (.WriteString w "Hello"))`
	tr := transpiler.New()
	
	result, err := tr.TranspileFromInput(input)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	
	AssertContainsAll(t, result,
		`func(w http.ResponseWriter, r *http.Request) {`,
		`w.WriteString("Hello")`,
		`import "net/http"`,
	)
}

func TestTranspiler_TranspileFromInput_ComplexExample(t *testing.T) {
	input := `
(import "net/http")
(import "github.com/gorilla/mux")
(def router (.NewRouter mux))
(def server (.NewServer http ":8080" router))
`
	
	tr := transpiler.New()
	result, err := tr.TranspileFromInput(input)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	
	AssertContainsAll(t, result,
		`package main`,
		`import "net/http"`,
		`import "github.com/gorilla/mux"`,
		`router := mux.NewRouter()`,
		`server := http.NewServer(":8080", router)`,
		`func main() {`,
	)
}

func TestTranspiler_TranspileFromInput_SyntaxErrors(t *testing.T) {
	errorCases := []TestCase{
		{
			Name:  "Unclosed parenthesis",
			Input: "(+ 1 2",
			Error: true,
		},
		{
			Name:  "Extra closing parenthesis",
			Input: "(+ 1 2))",
			Error: true,
		},
		{
			Name:  "Empty input",
			Input: "",
			Error: true, // Empty input produces syntax error in current implementation
		},
	}

	RunTestCases(t, errorCases)
}

func TestTranspiler_TranspileFromFile(t *testing.T) {
	// Create a temporary file
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test.vex")
	
	content := `(def x 42)
(+ x 10)`
	
	err := os.WriteFile(testFile, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	
	tr := transpiler.New()
	result, err := tr.TranspileFromFile(testFile)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	
	AssertContainsAll(t, result,
		"x := 42",
		"_ = x + 10",
	)
}

func TestTranspiler_TranspileFromFile_NonExistentFile(t *testing.T) {
	tr := transpiler.New()
	_, err := tr.TranspileFromFile("nonexistent.vex")
	
	if err == nil {
		t.Error("Expected error for non-existent file, but got success")
	}
	
	if !strings.Contains(err.Error(), "failed to read file") {
		t.Errorf("Expected 'failed to read file' error, got: %v", err)
	}
}

func TestTranspiler_Reset(t *testing.T) {
	tr := transpiler.New()
	
	// Transpile something first
	_, err := tr.TranspileFromInput("(def x 10)")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	
	// Reset should not cause issues
	tr.Reset()
	
	// Should be able to transpile again
	result, err := tr.TranspileFromInput("(def y 20)")
	if err != nil {
		t.Fatalf("Unexpected error after reset: %v", err)
	}
	
	AssertContainsAll(t, result, "y := 20")
}

func TestTranspiler_GeneratedCodeStructure(t *testing.T) {
	tr := transpiler.New()
	result, err := tr.TranspileFromInput("(def x 10)")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	
	// Check basic Go code structure
	lines := strings.Split(result, "\n")
	
	var foundPackage, foundMain, foundClosingBrace bool
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "package main" {
			foundPackage = true
		}
		if trimmed == "func main() {" {
			foundMain = true
		}
		if trimmed == "}" {
			foundClosingBrace = true
		}
	}
	
	if !foundPackage {
		t.Error("Generated code should have 'package main'")
	}
	if !foundMain {
		t.Error("Generated code should have 'func main() {'")
	}
	if !foundClosingBrace {
		t.Error("Generated code should have closing brace '}'")
	}
}