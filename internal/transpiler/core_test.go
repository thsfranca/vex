package transpiler

import (
	"os"
	"path/filepath"
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
			expected: "x := 10",
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
			transpiler := New()
			result, err := transpiler.TranspileFromInput(tt.input)

			if tt.hasError && err == nil {
				t.Errorf("Expected an error but got none")
				return
			}

			if !tt.hasError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
				return
			}

			if !strings.Contains(result, tt.expected) {
				t.Errorf("Expected output to contain %q, but got %q", tt.expected, result)
			}
		})
	}
}

func TestTranspiler_FullProgram(t *testing.T) {
	input := `(import "fmt")
(def greeting "Hello, World!")
(fmt/Println greeting)`

	transpiler := New()
	result, err := transpiler.TranspileFromInput(input)

	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	expectedParts := []string{
		"package main",
		`import "fmt"`,
		"func main()",
		"greeting := \"Hello, World!\"",
		"fmt.Println(greeting)",
	}

	for _, part := range expectedParts {
		if !strings.Contains(result, part) {
			t.Errorf("Expected output to contain %q, but got %q", part, result)
		}
	}
}

func TestTranspiler_ImplicitReturns(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:  "Variable definition return",
			input: "(def x 42)",
			expected: []string{
				"x := 42",
				"_ = x // Return last defined value",
			},
		},
		{
			name:  "Import return",
			input: `(import "fmt")`,
			expected: []string{
				`import "fmt"`,
				"_ = \"import completed\" // Import statement result",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			transpiler := New()
			result, err := transpiler.TranspileFromInput(tt.input)

			if err != nil {
				t.Errorf("Expected no error, got: %v", err)
			}

			for _, expected := range tt.expected {
				if !strings.Contains(result, expected) {
					t.Errorf("Expected output to contain %q, but got %q", expected, result)
				}
			}
		})
	}
}

func TestTranspiler_DependencyDetection(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		expectedImport string
	}{
		{
			name:           "Standard library import",
			input:          `(import "fmt")`,
			expectedImport: "fmt",
		},
		{
			name:           "Multiple imports",
			input:          `(import "fmt")\n(import "os")`,
			expectedImport: "fmt",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			transpiler := New()
			result, err := transpiler.TranspileFromInput(tt.input)

			if err != nil {
				t.Errorf("Expected no error, got: %v", err)
			}

			expectedImportLine := `import "` + tt.expectedImport + `"`
			if !strings.Contains(result, expectedImportLine) {
				t.Errorf("Expected output to contain %q, but got %q", expectedImportLine, result)
			}
		})
	}
}

func TestTranspiler_GoModulesDetection(t *testing.T) {
	tests := []struct {
		name                string
		input               string
		expectedThirdParty  bool
		expectedModuleName  string
	}{
		{
			name:               "Standard library - no third party",
			input:              `(import "fmt")`,
			expectedThirdParty: false,
		},
		{
			name:               "Third party module",
			input:              `(import "github.com/example/package")`,
			expectedThirdParty: true,
			expectedModuleName: "github.com/example/package",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			transpiler := New()
			_, err := transpiler.TranspileFromInput(tt.input)

			if err != nil {
				t.Errorf("Expected no error, got: %v", err)
			}

			modules := transpiler.GetDetectedModules()
			
			if tt.expectedThirdParty {
				if _, exists := modules[tt.expectedModuleName]; !exists {
					t.Errorf("Expected module %q to be detected as third-party", tt.expectedModuleName)
				}
			} else {
				if len(modules) > 0 {
					t.Errorf("Expected no third-party modules, but found: %v", modules)
				}
			}
		})
	}
}

func TestTranspiler_ImportDeduplication(t *testing.T) {
	input := `(import "fmt")
(import "fmt")
(import "os")`

	transpiler := New()
	result, err := transpiler.TranspileFromInput(input)

	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	// Count occurrences of import statements
	fmtCount := strings.Count(result, `import "fmt"`)
	osCount := strings.Count(result, `import "os"`)

	if fmtCount != 1 {
		t.Errorf("Expected exactly 1 fmt import, got %d", fmtCount)
	}

	if osCount != 1 {
		t.Errorf("Expected exactly 1 os import, got %d", osCount)
	}
}

func TestTranspiler_VariableUsage(t *testing.T) {
	input := `(def message "Hello")
(def count 42)`

	transpiler := New()
	result, err := transpiler.TranspileFromInput(input)

	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	expectedParts := []string{
		"message := \"Hello\"",
		"_ = message // Use variable to satisfy Go compiler",
		"count := 42",
		"_ = count // Use variable to satisfy Go compiler",
	}

	for _, part := range expectedParts {
		if !strings.Contains(result, part) {
			t.Errorf("Expected output to contain %q, but got %q", part, result)
		}
	}
}

func TestTranspiler_ComplexDependencyProgram(t *testing.T) {
	input := `(import "fmt")
(import "os")
(import "strings")
(def program-name "VexTranspiler")
(def version "1.0.0")
(fmt/Printf "Program: %s, Version: %s\n" program-name version)
(def args (os/Args))
(def arg-count (count args))
(fmt/Printf "Arguments: %d\n" arg-count)`

	transpiler := New()
	result, err := transpiler.TranspileFromInput(input)

	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	expectedImports := []string{
		`import "fmt"`,
		`import "os"`,
		`import "strings"`,
	}

	for _, imp := range expectedImports {
		if !strings.Contains(result, imp) {
			t.Errorf("Expected output to contain %q", imp)
		}
	}

	expectedCode := []string{
		"program_name := \"VexTranspiler\"",
		"version := \"1.0.0\"",
		"fmt.Printf(\"Program: %s, Version: %s\\n\", program_name, version)",
		"args := os.Args",
		"arg_count := len(args)",
		"fmt.Printf(\"Arguments: %d\\n\", arg_count)",
	}

	for _, code := range expectedCode {
		if !strings.Contains(result, code) {
			t.Errorf("Expected output to contain %q", code)
		}
	}

	// Check that detected modules are empty for standard library
	modules := transpiler.GetDetectedModules()
	if len(modules) > 0 {
		t.Errorf("Expected no third-party modules for standard library imports, got: %v", modules)
	}
}

func TestTranspiler_TranspileFromFile(t *testing.T) {
	// Create a temporary file
	tempDir := t.TempDir()
	tempFile := filepath.Join(tempDir, "test.vx")

	content := `(import "fmt")
(def message "Hello from file!")
(fmt/Println message)`

	err := os.WriteFile(tempFile, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}

	transpiler := New()
	result, err := transpiler.TranspileFromFile(tempFile)

	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
		return
	}

	expectedParts := []string{
		"package main",
		`import "fmt"`,
		"message := \"Hello from file!\"",
		"fmt.Println(message)",
	}

	for _, part := range expectedParts {
		if !strings.Contains(result, part) {
			t.Errorf("Expected output to contain %q", part)
		}
	}
}

func TestTranspiler_TranspileFromFile_NonExistent(t *testing.T) {
	transpiler := New()
	_, err := transpiler.TranspileFromFile("nonexistent.vx")

	if err == nil {
		t.Error("Expected an error for non-existent file")
	}
}

func TestTranspiler_ArrayHandling(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Empty array",
			input:    "(def arr [])",
			expected: "arr := []interface{}{}",
		},
		{
			name:     "Array with elements",
			input:    "(def nums [1 2 3])",
			expected: "nums := []interface{}{1, 2, 3}",
		},
		{
			name:     "Array with strings",
			input:    `(def words ["hello" "world"])`,
			expected: `words := []interface{}{"hello", "world"}`,
		},
		{
			name:     "Mixed array",
			input:    `(def mixed [1 "hello" 42])`,
			expected: `mixed := []interface{}{1, "hello", 42}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			transpiler := New()
			result, err := transpiler.TranspileFromInput(tt.input)

			if err != nil {
				t.Errorf("Expected no error, got: %v", err)
			}

			if !strings.Contains(result, tt.expected) {
				t.Errorf("Expected output to contain %q, but got %q", tt.expected, result)
			}
		})
	}
}

func TestTranspiler_VisitNodeComprehensive(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:  "String literals preserved",
			input: `(def msg "Hello World")`,
			expected: []string{
				`msg := "Hello World"`,
			},
		},
		{
			name:  "Numeric literals",
			input: "(def num 42)",
			expected: []string{
				"num := 42",
			},
		},
		{
			name:  "Symbol handling",
			input: "(def x y)",
			expected: []string{
				"x := y",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			transpiler := New()
			result, err := transpiler.TranspileFromInput(tt.input)

			if err != nil {
				t.Errorf("Expected no error, got: %v", err)
			}

			for _, expected := range tt.expected {
				if !strings.Contains(result, expected) {
					t.Errorf("Expected output to contain %q, but got %q", expected, result)
				}
			}
		})
	}
}