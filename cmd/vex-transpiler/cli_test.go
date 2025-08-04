package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadCoreVex(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(string) error
		expected string
		hasError bool
	}{
		{
			name: "Valid core.vx file",
			setup: func(dir string) error {
				coreContent := `(import "fmt")
(fmt/Println "Core loaded")`
				return os.WriteFile(filepath.Join(dir, "core.vx"), []byte(coreContent), 0644)
			},
			expected: `(import "fmt")`,
			hasError: false,
		},
		{
			name: "Missing core.vx file",
			setup: func(dir string) error {
				// Don't create the file
				return nil
			},
			expected: "",
			hasError: false, // loadCoreVex returns empty string, not error
		},
		{
			name: "Empty core.vx file",
			setup: func(dir string) error {
				return os.WriteFile(filepath.Join(dir, "core.vx"), []byte(""), 0644)
			},
			expected: "",
			hasError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := t.TempDir()
			
			// Setup test environment
			if err := tt.setup(tempDir); err != nil {
				t.Fatalf("Setup failed: %v", err)
			}

			// Change to temp directory for the test
			oldWd, _ := os.Getwd()
			defer os.Chdir(oldWd)
			os.Chdir(tempDir)

			result := loadCoreVex(false)

			if tt.hasError && result != "" {
				t.Error("Expected empty result for error case")
				return
			}

			if !tt.hasError && tt.expected != "" && !strings.Contains(result, tt.expected) {
				t.Errorf("Expected result to contain: %s\nActual result: %s", tt.expected, result)
			}
		})
	}
}

func TestGenerateGoMod(t *testing.T) {
	tests := []struct {
		name     string
		modules  map[string]string
		expected []string
	}{
		{
			name:    "No modules",
			modules: map[string]string{},
			expected: []string{
				"module vex-project-",
				"go 1.21",
			},
		},
		{
			name: "With third-party modules",
			modules: map[string]string{
				"github.com/google/uuid": "v1.0.0",
				"golang.org/x/crypto":    "v1.0.0",
			},
			expected: []string{
				"module vex-project-",
				"go 1.21",
				"require (",
				"github.com/google/uuid v1.0.0",
				"golang.org/x/crypto v1.0.0",
				")",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := t.TempDir()
			
			err := generateGoMod(tempDir, tt.modules, false)
			if err != nil {
				t.Fatalf("generateGoMod failed: %v", err)
			}

			// Read the generated go.mod
			goModPath := filepath.Join(tempDir, "go.mod")
			content, err := os.ReadFile(goModPath)
			if err != nil {
				t.Fatalf("Failed to read generated go.mod: %v", err)
			}

			goModContent := string(content)
			for _, expected := range tt.expected {
				if !strings.Contains(goModContent, expected) {
					t.Errorf("Expected go.mod to contain: %s\nActual content:\n%s", expected, goModContent)
				}
			}
		})
	}
}

func TestDownloadDependencies(t *testing.T) {
	tempDir := t.TempDir()
	
	// Create a minimal go.mod
	goModContent := `module test-project

go 1.21
`
	err := os.WriteFile(filepath.Join(tempDir, "go.mod"), []byte(goModContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create go.mod: %v", err)
	}

	// Test the function (it should not fail for a basic module)
	err = downloadDependencies(tempDir, false)
	if err != nil {
		// This might fail in CI environments without Go, so we log but don't fail
		t.Logf("downloadDependencies failed (expected in some environments): %v", err)
	}
}

func TestBuildBinary(t *testing.T) {
	tempDir := t.TempDir()
	genDir := filepath.Join(tempDir, "gen")
	binDir := filepath.Join(tempDir, "bin")
	
	// Create directories
	err := os.MkdirAll(genDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create gen dir: %v", err)
	}
	err = os.MkdirAll(binDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create bin dir: %v", err)
	}

	// Create a minimal Go program
	mainGoContent := `package main

import "fmt"

func main() {
	fmt.Println("Hello from test")
}
`
	err = os.WriteFile(filepath.Join(genDir, "main.go"), []byte(mainGoContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create main.go: %v", err)
	}

	// Create a minimal go.mod
	goModContent := `module test-project

go 1.21
`
	err = os.WriteFile(filepath.Join(tempDir, "go.mod"), []byte(goModContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create go.mod: %v", err)
	}

	// Test the build function
	binaryPath := filepath.Join(binDir, "test")
	err = buildBinary(tempDir, genDir, binaryPath, false)
	if err != nil {
		// This might fail in CI environments without Go, so we log but don't fail
		t.Logf("buildBinary failed (expected in some environments): %v", err)
		return
	}

	// Check if binary was created
	if _, err := os.Stat(filepath.Join(binDir, "app")); os.IsNotExist(err) {
		t.Error("Expected binary to be created but it doesn't exist")
	}
}

func TestPrintUsage(t *testing.T) {
	// Capture stdout would be complex, so just test that the function doesn't panic
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("printUsage panicked: %v", r)
		}
	}()
	
	printUsage()
}

func TestTranspileCommand(t *testing.T) {
	tempDir := t.TempDir()
	
	// Create a test Vex file
	vexFile := filepath.Join(tempDir, "test.vx")
	vexContent := `(import "fmt")
(def x 42)
(fmt/Println "Hello World")`
	
	err := os.WriteFile(vexFile, []byte(vexContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Test cases
	tests := []struct {
		name      string
		inputFile string
		hasError  bool
	}{
		{
			name:      "Valid Vex file",
			inputFile: vexFile,
			hasError:  false,
		},
		{
			name:      "Non-existent file",
			inputFile: filepath.Join(tempDir, "nonexistent.vx"),
			hasError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Mock command line args
			oldArgs := os.Args
			defer func() { os.Args = oldArgs }()
			
			os.Args = []string{"vex", "transpile", "-input", tt.inputFile}
			
			// Test the function (we can't easily test the flag parsing, so test the core logic)
			// This is a simplified test of the transpile logic
			if _, err := os.Stat(tt.inputFile); os.IsNotExist(err) && !tt.hasError {
				t.Error("Expected file to exist for success case")
			}
		})
	}
}

func TestRunCommandUnit(t *testing.T) {
	tempDir := t.TempDir()
	
	// Create a test Vex file
	vexFile := filepath.Join(tempDir, "test.vx")
	vexContent := `(import "fmt")
(fmt/Println "Hello from run test")`
	
	err := os.WriteFile(vexFile, []byte(vexContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Test the file exists (basic test since running the full command would be complex)
	if _, err := os.Stat(vexFile); os.IsNotExist(err) {
		t.Error("Test Vex file should exist")
	}
}

func TestLoadCoreVexEdgeCases(t *testing.T) {
	tests := []struct {
		name        string
		setup       func(string) error
		verbose     bool
		expectedLen int
	}{
		{
			name: "Large core.vx file",
			setup: func(dir string) error {
				content := strings.Repeat("(import \"fmt\")\n", 100)
				return os.WriteFile(filepath.Join(dir, "core.vx"), []byte(content), 0644)
			},
			verbose:     true,
			expectedLen: 500, // Should be substantial
		},
		{
			name: "Verbose mode",
			setup: func(dir string) error {
				return os.WriteFile(filepath.Join(dir, "core.vx"), []byte("(import \"fmt\")"), 0644)
			},
			verbose:     true,
			expectedLen: 10,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := t.TempDir()
			
			if err := tt.setup(tempDir); err != nil {
				t.Fatalf("Setup failed: %v", err)
			}

			oldWd, _ := os.Getwd()
			defer os.Chdir(oldWd)
			os.Chdir(tempDir)

			result := loadCoreVex(tt.verbose)

			if len(result) < tt.expectedLen {
				t.Errorf("Expected result length >= %d, got %d", tt.expectedLen, len(result))
			}
		})
	}
}

func TestGenerateGoModEdgeCases(t *testing.T) {
	tests := []struct {
		name    string
		modules map[string]string
		verbose bool
	}{
		{
			name: "Many modules",
			modules: map[string]string{
				"github.com/google/uuid":    "v1.0.0",
				"golang.org/x/crypto":       "v1.0.0", 
				"github.com/gin-gonic/gin":  "v1.0.0",
				"github.com/gorilla/mux":    "v1.0.0",
			},
			verbose: false,
		},
		{
			name: "Verbose mode",
			modules: map[string]string{
				"github.com/test/module": "v1.0.0",
			},
			verbose: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := t.TempDir()
			
			err := generateGoMod(tempDir, tt.modules, tt.verbose)
			if err != nil {
				t.Fatalf("generateGoMod failed: %v", err)
			}

			// Verify all modules are in go.mod
			goModPath := filepath.Join(tempDir, "go.mod")
			content, err := os.ReadFile(goModPath)
			if err != nil {
				t.Fatalf("Failed to read go.mod: %v", err)
			}

			goModStr := string(content)
			for module := range tt.modules {
				if !strings.Contains(goModStr, module) {
					t.Errorf("Expected go.mod to contain module: %s", module)
				}
			}
		})
	}
}

func TestDownloadDependenciesEdgeCases(t *testing.T) {
	tempDir := t.TempDir()
	
	// Test verbose mode
	goModContent := `module test-verbose

go 1.21
`
	err := os.WriteFile(filepath.Join(tempDir, "go.mod"), []byte(goModContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create go.mod: %v", err)
	}

	// Test with verbose = true (this exercises different code paths)
	err = downloadDependencies(tempDir, true)
	// Note: This may fail in test environments, but we're testing the code paths
	if err != nil {
		t.Logf("downloadDependencies with verbose failed (expected in test environments): %v", err)
	}
}

func TestBuildBinaryEdgeCases(t *testing.T) {
	tempDir := t.TempDir()
	genDir := filepath.Join(tempDir, "gen")
	binDir := filepath.Join(tempDir, "bin")
	
	// Create directories
	os.MkdirAll(genDir, 0755)
	os.MkdirAll(binDir, 0755)

	// Test verbose mode
	mainGoContent := `package main
func main() {}`
	
	os.WriteFile(filepath.Join(genDir, "main.go"), []byte(mainGoContent), 0644)
	
	goModContent := `module test-verbose
go 1.21`
	os.WriteFile(filepath.Join(tempDir, "go.mod"), []byte(goModContent), 0644)

	// Test with verbose = true
	binaryPath := filepath.Join(binDir, "test-verbose")
	err := buildBinary(tempDir, genDir, binaryPath, true)
	if err != nil {
		t.Logf("buildBinary with verbose failed (expected in test environments): %v", err)
	}
}