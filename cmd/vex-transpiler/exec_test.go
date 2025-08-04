package main

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestExecCommand_BasicFunctionality(t *testing.T) {
	// Build the binary for testing
	binaryPath := buildTestBinary(t)
	defer os.Remove(binaryPath)

	// Create test files
	testDir := t.TempDir()
	
	// Create minimal core.vx for testing
	coreVxFile := filepath.Join(testDir, "core.vx")
	coreVxContent := `(import "fmt")
(fmt/Println "✅ Core library loaded!")`
	
	err := os.WriteFile(coreVxFile, []byte(coreVxContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create core.vx: %v", err)
	}
	
	// Simple test without dependencies
	simpleVexFile := filepath.Join(testDir, "simple.vx")
	simpleContent := `(import "fmt")
(fmt/Println "Hello from test!")`
	
	err = os.WriteFile(simpleVexFile, []byte(simpleContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Test exec command
	cmd := exec.Command(binaryPath, "exec", simpleVexFile)
	cmd.Dir = testDir
	
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	
	err = cmd.Run()
	if err != nil {
		t.Fatalf("Exec command failed: %v\nStderr: %s", err, stderr.String())
	}

	// Check output
	output := stdout.String()
	expectedOutputs := []string{
		"✅ Core library loaded!",
		"Hello from test!",
	}

	for _, expected := range expectedOutputs {
		if !strings.Contains(output, expected) {
			t.Errorf("Expected output to contain: %s\nActual output: %s", expected, output)
		}
	}
}

func TestExecCommand_WithStandardLibraryDependencies(t *testing.T) {
	// Build the binary for testing
	binaryPath := buildTestBinary(t)
	defer os.Remove(binaryPath)

	// Create test files
	testDir := t.TempDir()
	
	// Create minimal core.vx for testing
	coreVxFile := filepath.Join(testDir, "core.vx")
	coreVxContent := `(import "fmt")
(fmt/Println "✅ Core library loaded!")`
	
	err := os.WriteFile(coreVxFile, []byte(coreVxContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create core.vx: %v", err)
	}
	
	// Test with standard library dependencies
	depsVexFile := filepath.Join(testDir, "deps.vx")
	depsContent := `(import "fmt")
(import "strings")
(def message "hello world")
(def upperMsg (strings/ToUpper message))
(fmt/Println "Message:" upperMsg)`
	
	err = os.WriteFile(depsVexFile, []byte(depsContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Test exec command with verbose output
	cmd := exec.Command(binaryPath, "exec", "-verbose", depsVexFile)
	cmd.Dir = testDir
	
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	
	// Set a reasonable timeout
	done := make(chan error, 1)
	go func() {
		done <- cmd.Run()
	}()

	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("Exec command failed: %v\nStderr: %s\nStdout: %s", err, stderr.String(), stdout.String())
		}
	case <-time.After(30 * time.Second):
		cmd.Process.Kill()
		t.Fatal("Exec command timed out after 30 seconds")
	}

	// Check output
	output := stdout.String()
	expectedOutputs := []string{
		"✅ Core library loaded!",
		"Message: HELLO WORLD",
	}

	for _, expected := range expectedOutputs {
		if !strings.Contains(output, expected) {
			t.Errorf("Expected output to contain: %s\nActual output: %s", expected, output)
		}
	}

	// Check verbose output in stderr
	verboseOutput := stderr.String()
	expectedVerboseOutputs := []string{
		"🚀 Executing Vex file:",
		"📚 Loaded core.vx standard library",
		"🔄 Transpilation complete, executing...",
		"Generated Go code:",
	}

	for _, expected := range expectedVerboseOutputs {
		if !strings.Contains(verboseOutput, expected) {
			t.Errorf("Expected verbose output to contain: %s\nActual stderr: %s", expected, verboseOutput)
		}
	}
}

func TestExecCommand_ImplicitReturns(t *testing.T) {
	// Build the binary for testing
	binaryPath := buildTestBinary(t)
	defer os.Remove(binaryPath)

	testDir := t.TempDir()
	
	// Create minimal core.vx for testing
	coreVxFile := filepath.Join(testDir, "core.vx")
	coreVxContent := `(import "fmt")
(fmt/Println "✅ Core library loaded!")`
	
	err := os.WriteFile(coreVxFile, []byte(coreVxContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create core.vx: %v", err)
	}
	
	tests := []struct {
		name           string
		vexContent     string
		expectedOutput []string
	}{
		{
			name: "Variable definitions with implicit usage",
			vexContent: `(import "fmt")
(def x 42)
(def y 24)
(fmt/Println "Sum:" (+ x y))`,
			expectedOutput: []string{
				"✅ Core library loaded!",
				"Sum: 66",
			},
		},
		{
			name: "Arithmetic as last expression",
			vexContent: `(def a 10)
(def b 5)
(fmt/Println "Result:" (+ a b))`,
			expectedOutput: []string{
				"Result: 15",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vexFile := filepath.Join(testDir, tt.name+".vx")
			err2 := os.WriteFile(vexFile, []byte(tt.vexContent), 0644)
			if err2 != nil {
				t.Fatalf("Failed to create test file: %v", err2)
			}

			cmd := exec.Command(binaryPath, "exec", vexFile)
			cmd.Dir = testDir
			
			var stdout, stderr bytes.Buffer
			cmd.Stdout = &stdout
			cmd.Stderr = &stderr
			
			// Set timeout
			done := make(chan error, 1)
			go func() {
				done <- cmd.Run()
			}()

			select {
			case err := <-done:
				if err != nil {
					t.Fatalf("Exec command failed: %v\nStderr: %s", err, stderr.String())
				}
			case <-time.After(15 * time.Second):
				cmd.Process.Kill()
				t.Fatal("Exec command timed out")
			}

			output := stdout.String()
			for _, expected := range tt.expectedOutput {
				if !strings.Contains(output, expected) {
					t.Errorf("Expected output to contain: %s\nActual output: %s", expected, output)
				}
			}
		})
	}
}

func TestExecCommand_ErrorHandling(t *testing.T) {
	// Build the binary for testing
	binaryPath := buildTestBinary(t)
	defer os.Remove(binaryPath)

	testDir := t.TempDir()

	tests := []struct {
		name           string
		args           []string
		vexContent     string
		expectError    bool
		expectedStderr string
	}{
		{
			name:           "Missing file argument",
			args:           []string{"exec"},
			expectError:    true,
			expectedStderr: "Error: input file is required",
		},
		{
			name:           "Non-existent file",
			args:           []string{"exec", "nonexistent.vx"},
			expectError:    true,
			expectedStderr: "Error reading file",
		},
		{
			name:        "Invalid Vex syntax",
			args:        []string{"exec", "invalid.vx"},
			vexContent:  `(def x`,
			expectError: true,
			expectedStderr: "Build error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var vexFile string
			if tt.vexContent != "" {
				vexFile = filepath.Join(testDir, "invalid.vx")
				err := os.WriteFile(vexFile, []byte(tt.vexContent), 0644)
				if err != nil {
					t.Fatalf("Failed to create test file: %v", err)
				}
			}

			cmd := exec.Command(binaryPath, tt.args...)
			cmd.Dir = testDir
			
			var stdout, stderr bytes.Buffer
			cmd.Stdout = &stdout
			cmd.Stderr = &stderr
			
			err := cmd.Run()
			
			if tt.expectError && err == nil {
				t.Error("Expected command to fail, but it succeeded")
			} else if !tt.expectError && err != nil {
				t.Errorf("Expected command to succeed, but it failed: %v", err)
			}

			if tt.expectedStderr != "" && !strings.Contains(stderr.String(), tt.expectedStderr) {
				t.Errorf("Expected stderr to contain: %s\nActual stderr: %s", tt.expectedStderr, stderr.String())
			}
		})
	}
}

func TestExecCommand_GoModulesGeneration(t *testing.T) {
	// Build the binary for testing
	binaryPath := buildTestBinary(t)
	defer os.Remove(binaryPath)

	testDir := t.TempDir()
	
	// Create minimal core.vx for testing
	coreVxFile := filepath.Join(testDir, "core.vx")
	coreVxContent := `(import "fmt")
(fmt/Println "✅ Core library loaded!")`
	
	err := os.WriteFile(coreVxFile, []byte(coreVxContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create core.vx: %v", err)
	}
	
	// Test that .vex directory and files are generated correctly
	vexFile := filepath.Join(testDir, "modules.vx")
	vexContent := `(import "fmt")
(fmt/Println "Testing module generation")`
	
	err = os.WriteFile(vexFile, []byte(vexContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Execute with verbose to see the process
	cmd := exec.Command(binaryPath, "exec", "-verbose", vexFile)
	cmd.Dir = testDir
	
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	
	// Set timeout
	done := make(chan error, 1)
	go func() {
		done <- cmd.Run()
	}()

	select {
	case err := <-done:
		// Check if .vex directory structure was created
		vexDir := filepath.Join(testDir, ".vex")
		
		// Check directory structure
		expectedDirs := []string{
			filepath.Join(vexDir, "gen"),
			filepath.Join(vexDir, "bin"),
		}
		
		for _, dir := range expectedDirs {
			if _, err := os.Stat(dir); os.IsNotExist(err) {
				t.Errorf("Expected directory %s was not created", dir)
			}
		}

		// Check go.mod was generated
		goModPath := filepath.Join(vexDir, "go.mod")
		if _, err := os.Stat(goModPath); os.IsNotExist(err) {
			t.Error("Expected go.mod file was not created")
		} else {
			// Read and check go.mod content
			goModContent, err := os.ReadFile(goModPath)
			if err != nil {
				t.Errorf("Failed to read go.mod: %v", err)
			} else {
				goModStr := string(goModContent)
				expectedInGoMod := []string{
					"module vex-project-",
					"go 1.21",
				}
				
				for _, expected := range expectedInGoMod {
					if !strings.Contains(goModStr, expected) {
						t.Errorf("Expected go.mod to contain: %s\nActual go.mod:\n%s", expected, goModStr)
					}
				}
			}
		}

		// Check main.go was generated
		mainGoPath := filepath.Join(vexDir, "gen", "main.go")
		if _, err := os.Stat(mainGoPath); os.IsNotExist(err) {
			t.Error("Expected main.go file was not created")
		} else {
			// Read and check main.go content
			mainGoContent, err := os.ReadFile(mainGoPath)
			if err != nil {
				t.Errorf("Failed to read main.go: %v", err)
			} else {
				mainGoStr := string(mainGoContent)
				expectedInMainGo := []string{
					"package main",
					`import "fmt"`,
					"func main() {",
					"fmt.Println(",
				}
				
				for _, expected := range expectedInMainGo {
					if !strings.Contains(mainGoStr, expected) {
						t.Errorf("Expected main.go to contain: %s\nActual main.go:\n%s", expected, mainGoStr)
					}
				}
			}
		}

		if err != nil {
			t.Logf("Command had an error (this may be expected): %v\nStderr: %s", err, stderr.String())
		}

	case <-time.After(30 * time.Second):
		cmd.Process.Kill()
		t.Fatal("Exec command timed out after 30 seconds")
	}
}