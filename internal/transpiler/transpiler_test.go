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
		"_ = x // Use variable to satisfy Go compiler",
		"fmt.Println(\"Hello World\") // Last expression",
		"}",
	}

	for _, expected := range expectedParts {
		if !strings.Contains(result, expected) {
			t.Errorf("Expected output to contain: %s\nActual output:\n%s", expected, result)
		}
	}
}

func TestTranspiler_ImplicitReturns(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Last expression with function call",
			input:    "(import \"fmt\")\n(fmt/Println \"Hello\")",
			expected: `fmt.Println("Hello") // Last expression`,
		},
		{
			name:     "Last expression with variable definition",
			input:    `(def x 42)`,
			expected: `_ = x // Return last defined value`,
		},
		{
			name:     "Last expression with arithmetic",
			input:    `(+ 1 2)`,
			expected: `_ = (1 + 2)`,
		},
		{
			name:     "Multiple expressions - only last is implicit return",  
			input:    "(def x 10)\n(def y 20)\n(+ x y)",
			expected: `_ = (x + y)`,
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

func TestTranspiler_DependencyDetection(t *testing.T) {
	tr := New()
	
	tests := []struct {
		name           string
		importPath     string
		expectedThirdParty bool
	}{
		{
			name:           "Standard library - fmt",
			importPath:     "fmt",
			expectedThirdParty: false,
		},
		{
			name:           "Standard library - strings",
			importPath:     "strings",
			expectedThirdParty: false,
		},
		{
			name:           "Third party - github.com",
			importPath:     "github.com/google/uuid",
			expectedThirdParty: true,
		},
		{
			name:           "Third party - golang.org",
			importPath:     "golang.org/x/crypto/bcrypt",
			expectedThirdParty: true,
		},
		{
			name:           "Third party - other domain",
			importPath:     "example.com/some/package",
			expectedThirdParty: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tr.isThirdPartyModule(tt.importPath)
			if result != tt.expectedThirdParty {
				t.Errorf("isThirdPartyModule(%q) = %v, expected %v", tt.importPath, result, tt.expectedThirdParty)
			}
		})
	}
}

func TestTranspiler_GoModulesDetection(t *testing.T) {
	input := `(import "fmt")
(import "github.com/google/uuid")
(import "golang.org/x/crypto/bcrypt")
(def id (uuid/New))`

	tr := New()
	_, err := tr.TranspileFromInput(input)
	
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	modules := tr.GetDetectedModules()
	
	expectedModules := map[string]string{
		"github.com/google/uuid":       "v1.0.0",
		"golang.org/x/crypto/bcrypt":   "v1.0.0",
	}

	if len(modules) != len(expectedModules) {
		t.Errorf("Expected %d modules, got %d: %v", len(expectedModules), len(modules), modules)
	}

	for expectedModule, expectedVersion := range expectedModules {
		if version, exists := modules[expectedModule]; !exists {
			t.Errorf("Expected module %s not found in detected modules", expectedModule)
		} else if version != expectedVersion {
			t.Errorf("Module %s: expected version %s, got %s", expectedModule, expectedVersion, version)
		}
	}

	// Standard library should not be in modules
	for module := range modules {
		if module == "fmt" {
			t.Errorf("Standard library module 'fmt' should not be in detected modules")
		}
	}
}

func TestTranspiler_ImportDeduplication(t *testing.T) {
	input := `(import "fmt")
(import "fmt")
(import "fmt")
(fmt/Println "Hello")`

	tr := New()
	result, err := tr.TranspileFromInput(input)
	
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Count occurrences of import "fmt"
	importCount := strings.Count(result, `import "fmt"`)
	if importCount != 1 {
		t.Errorf("Expected import \"fmt\" to appear exactly once, but appeared %d times in:\n%s", importCount, result)
	}
}

func TestTranspiler_VariableUsage(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Variable definition includes usage",
			input:    `(def x 42)`,
			expected: `_ = x // Use variable to satisfy Go compiler`,
		},
		{
			name:     "Multiple variables all used",
			input:    "(def x 10)\n(def y 20)",
			expected: `_ = y // Return last defined value`,
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

func TestTranspiler_ComplexDependencyProgram(t *testing.T) {
	input := `(import "fmt")
(import "strings")
(import "github.com/google/uuid")
(import "golang.org/x/crypto/bcrypt")

(def message "Hello Vex!")
(def id (uuid/New))
(def upperMsg (strings/ToUpper message))
(fmt/Println "Message:" upperMsg "ID:" id)`

	tr := New()
	result, err := tr.TranspileFromInput(input)
	
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Check that all imports are present and deduplicated
	expectedImports := []string{
		`import "fmt"`,
		`import "strings"`,
		`import "github.com/google/uuid"`,
		`import "golang.org/x/crypto/bcrypt"`,
	}

	for _, expectedImport := range expectedImports {
		count := strings.Count(result, expectedImport)
		if count != 1 {
			t.Errorf("Expected %s to appear exactly once, appeared %d times", expectedImport, count)
		}
	}

	// Check Go modules detection
	modules := tr.GetDetectedModules()
	expectedModules := []string{
		"github.com/google/uuid",
		"golang.org/x/crypto/bcrypt",
	}

	for _, expectedModule := range expectedModules {
		if _, exists := modules[expectedModule]; !exists {
			t.Errorf("Expected module %s not found in detected modules: %v", expectedModule, modules)
		}
	}

	// Standard library should not be in modules
	stdLibModules := []string{"fmt", "strings"}
	for _, stdModule := range stdLibModules {
		if _, exists := modules[stdModule]; exists {
			t.Errorf("Standard library module %s should not be in detected modules", stdModule)
		}
	}

	// Check variable usage
	expectedUsage := []string{
		"_ = message // Use variable to satisfy Go compiler",
		"_ = id // Use variable to satisfy Go compiler", 
		"_ = upperMsg // Use variable to satisfy Go compiler",
	}

	for _, expected := range expectedUsage {
		if !strings.Contains(result, expected) {
			t.Errorf("Expected variable usage: %s\nNot found in:\n%s", expected, result)
		}
	}

	// Check last expression
	if !strings.Contains(result, "// Last expression") {
		t.Error("Expected last expression to be marked with comment")
	}
}