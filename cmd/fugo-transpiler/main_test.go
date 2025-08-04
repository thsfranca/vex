package main

import (
	"bytes"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// TestMainFunction tests the main function with various command line arguments
func TestMainFunction(t *testing.T) {
	// Build the binary for testing
	binaryPath := buildTestBinary(t)
	defer os.Remove(binaryPath)
	
	testCases := []struct {
		name           string
		args           []string
		inputFile      string
		inputContent   string
		expectError    bool
		expectedStdout string
		expectedStderr string
	}{
		{
			name:           "No arguments",
			args:           []string{},
			expectError:    true,
			expectedStderr: "Usage:",
		},
		{
			name:           "Unknown command",
			args:           []string{"unknown"},
			expectError:    true,
			expectedStderr: "Unknown command:",
		},
		{
			name:           "Transpile without input",
			args:           []string{"transpile"},
			expectError:    true,
			expectedStderr: "Error: -input flag is required",
		},
		{
			name:         "Simple transpilation",
			args:         []string{"transpile", "-input", "test.vex"},
			inputFile:    "test.vex",
			inputContent: "(def x 42)",
			expectError:  false,
			expectedStdout: "var x int64 = 42",
		},
		{
			name:         "Verbose transpilation",
			args:         []string{"transpile", "-input", "test.vex", "-verbose"},
			inputFile:    "test.vex",
			inputContent: "(def x 42)",
			expectError:  false,
			expectedStdout: "var x int64 = 42",
			expectedStderr: "Transpiling:",
		},
		{
			name:           "Run without input",
			args:           []string{"run"},
			expectError:    true,
			expectedStderr: "Error: -input flag is required",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create temporary directory for this test
			tempDir := t.TempDir()
			
			// Create input file if needed
			var inputPath string
			if tc.inputFile != "" {
				inputPath = filepath.Join(tempDir, tc.inputFile)
				err := os.WriteFile(inputPath, []byte(tc.inputContent), 0644)
				if err != nil {
					t.Fatalf("Failed to create input file: %v", err)
				}
				
				// Update args to use the full path
				for i, arg := range tc.args {
					if arg == tc.inputFile {
						tc.args[i] = inputPath
					}
				}
			}
			
			// Run the command
			cmd := exec.Command(binaryPath, tc.args...)
			var stdout, stderr bytes.Buffer
			cmd.Stdout = &stdout
			cmd.Stderr = &stderr
			
			err := cmd.Run()
			
			// Check error expectation
			if tc.expectError && err == nil {
				t.Error("Expected command to fail, but it succeeded")
			}
			if !tc.expectError && err != nil {
				t.Errorf("Expected command to succeed, but it failed: %v\nStderr: %s", err, stderr.String())
			}
			
			// Check output expectations
			if tc.expectedStdout != "" {
				if !strings.Contains(stdout.String(), tc.expectedStdout) {
					t.Errorf("Expected stdout to contain: %s\nActual stdout: %s", tc.expectedStdout, stdout.String())
				}
			}
			
			if tc.expectedStderr != "" {
				if !strings.Contains(stderr.String(), tc.expectedStderr) {
					t.Errorf("Expected stderr to contain: %s\nActual stderr: %s", tc.expectedStderr, stderr.String())
				}
			}
		})
	}
}

func TestMainFunction_OutputFile(t *testing.T) {
	// Build the binary for testing
	binaryPath := buildTestBinary(t)
	defer os.Remove(binaryPath)
	
	tempDir := t.TempDir()
	
	// Create input file
	inputFile := filepath.Join(tempDir, "input.vex")
	inputContent := `(def greeting "Hello World")
(fmt/Println greeting)`
	
	err := os.WriteFile(inputFile, []byte(inputContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create input file: %v", err)
	}
	
	// Output file
	outputFile := filepath.Join(tempDir, "output.go")
	
	// Run transpiler
	cmd := exec.Command(binaryPath, "transpile", "-input", inputFile, "-output", outputFile)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	
	err = cmd.Run()
	if err != nil {
		t.Fatalf("Command failed: %v\nStderr: %s", err, stderr.String())
	}
	
	// Check that output file was created
	if _, err := os.Stat(outputFile); os.IsNotExist(err) {
		t.Error("Expected output file to be created")
	}
	
	// Check output file content
	content, err := os.ReadFile(outputFile)
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}
	
	expectedParts := []string{
		"package main",
		`var greeting string = "Hello World"`,
		"_ = greeting",
		"fmt.Println(greeting)",
	}
	
	for _, part := range expectedParts {
		if !strings.Contains(string(content), part) {
			t.Errorf("Expected output file to contain: %s\nActual content: %s", part, string(content))
		}
	}
}

func TestMainFunction_NonExistentFile(t *testing.T) {
	// Build the binary for testing
	binaryPath := buildTestBinary(t)
	defer os.Remove(binaryPath)
	
	// Try to transpile a non-existent file
	cmd := exec.Command(binaryPath, "transpile", "-input", "nonexistent.vex")
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	
	err := cmd.Run()
	if err == nil {
		t.Error("Expected command to fail for non-existent file")
	}
	
	stderrStr := stderr.String()
	if !strings.Contains(stderrStr, "Error reading file") {
		t.Errorf("Expected file read error, got: %s", stderrStr)
	}
}

func TestMainFunction_InvalidVexSyntax(t *testing.T) {
	// Build the binary for testing
	binaryPath := buildTestBinary(t)
	defer os.Remove(binaryPath)
	
	tempDir := t.TempDir()
	
	// Create input file with invalid syntax
	inputFile := filepath.Join(tempDir, "invalid.vex")
	invalidContent := "(def x" // Missing closing parenthesis
	
	err := os.WriteFile(inputFile, []byte(invalidContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create input file: %v", err)
	}
	
	// Run transpiler
	cmd := exec.Command(binaryPath, "transpile", "-input", inputFile)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	
	err = cmd.Run()
	if err == nil {
		t.Error("Expected command to fail for invalid syntax")
	}
	
	stderrStr := stderr.String()
	if !strings.Contains(stderrStr, "Transpilation error") {
		t.Errorf("Expected transpilation error, got: %s", stderrStr)
	}
}

func TestMainFunction_OutputFileWriteError(t *testing.T) {
	// Build the binary for testing
	binaryPath := buildTestBinary(t)
	defer os.Remove(binaryPath)
	
	tempDir := t.TempDir()
	
	// Create input file
	inputFile := filepath.Join(tempDir, "input.vex")
	err := os.WriteFile(inputFile, []byte("(def x 42)"), 0644)
	if err != nil {
		t.Fatalf("Failed to create input file: %v", err)
	}
	
	// Try to write to a directory (should fail)
	outputFile := filepath.Join(tempDir, "nonexistent_dir", "output.go")
	
	cmd := exec.Command(binaryPath, "transpile", "-input", inputFile, "-output", outputFile)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	
	err = cmd.Run()
	if err == nil {
		t.Error("Expected command to fail when output directory doesn't exist")
	}
	
	stderrStr := stderr.String()
	if !strings.Contains(stderrStr, "Error writing output file") {
		t.Errorf("Expected output write error, got: %s", stderrStr)
	}
}

func TestMainFunction_VerboseOutput(t *testing.T) {
	// Build the binary for testing
	binaryPath := buildTestBinary(t)
	defer os.Remove(binaryPath)
	
	tempDir := t.TempDir()
	
	// Create input and output files
	inputFile := filepath.Join(tempDir, "input.vex")
	outputFile := filepath.Join(tempDir, "output.go")
	
	err := os.WriteFile(inputFile, []byte("(def x 42)"), 0644)
	if err != nil {
		t.Fatalf("Failed to create input file: %v", err)
	}
	
	// Run with verbose flag
	cmd := exec.Command(binaryPath, "transpile", "-input", inputFile, "-output", outputFile, "-verbose")
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	
	err = cmd.Run()
	if err != nil {
		t.Fatalf("Command failed: %v", err)
	}
	
	stderrStr := stderr.String()
	
	// Should contain both start and completion messages
	if !strings.Contains(stderrStr, "Transpiling:") {
		t.Error("Expected verbose start message")
	}
	if !strings.Contains(stderrStr, "Transpilation complete:") {
		t.Error("Expected verbose completion message")
	}
}

func TestMainFunction_StdoutOutput(t *testing.T) {
	// Build the binary for testing
	binaryPath := buildTestBinary(t)
	defer os.Remove(binaryPath)
	
	tempDir := t.TempDir()
	
	// Create input file
	inputFile := filepath.Join(tempDir, "input.vex")
	inputContent := "(def message \"Hello stdout\")"
	
	err := os.WriteFile(inputFile, []byte(inputContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create input file: %v", err)
	}
	
	// Run without output file (should write to stdout)
	cmd := exec.Command(binaryPath, "transpile", "-input", inputFile)
	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	
	err = cmd.Run()
	if err != nil {
		t.Fatalf("Command failed: %v", err)
	}
	
	stdoutStr := stdout.String()
	
	// Check that Go code was written to stdout
	expectedParts := []string{
		"package main",
		`var message string = "Hello stdout"`,
		"_ = message",
		"func main() {",
	}
	
	for _, part := range expectedParts {
		if !strings.Contains(stdoutStr, part) {
			t.Errorf("Expected stdout to contain: %s\nActual stdout: %s", part, stdoutStr)
		}
	}
}

// buildTestBinary builds the main binary for testing
func buildTestBinary(t *testing.T) string {
	t.Helper()
	
	// Create temporary binary
	tmpDir := t.TempDir()
	binaryPath := filepath.Join(tmpDir, "vex-transpiler-test")
	
	// Build the binary
	cmd := exec.Command("go", "build", "-o", binaryPath, ".")
	cmd.Dir = "." // Build in the current directory (cmd/vex-transpiler)
	
	// Capture output for debugging
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	
	err := cmd.Run()
	if err != nil {
		t.Fatalf("Failed to build test binary: %v\nStderr: %s", err, stderr.String())
	}
	
	return binaryPath
}

// TestFlagParsing tests the flag parsing logic directly
func TestFlagParsing(t *testing.T) {
	// Test default values
	// Note: We can't easily test flag parsing without refactoring main()
	// This is a limitation of testing main functions with global flags
	// In a production environment, we might refactor to make this more testable
	
	// For now, we test through the CLI interface
	// This is an integration test rather than a unit test
	
	t.Skip("Flag parsing tested through CLI integration tests")
}

// TestMainErrorHandling tests error handling scenarios
func TestMainErrorHandling(t *testing.T) {
	testCases := []struct {
		name        string
		args        []string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "Missing command",
			args:        []string{},
			expectError: true,
			errorMsg:    "Usage:",
		},
		{
			name:        "Unknown command",
			args:        []string{"badcommand"},
			expectError: true,
			errorMsg:    "Unknown command:",
		},
		{
			name:        "Transpile without input",
			args:        []string{"transpile"},
			expectError: true,
			errorMsg:    "Error: -input flag is required",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// We test error handling through the built binary
			// since testing main() directly would exit the test process
			
			binaryPath := buildTestBinary(t)
			defer os.Remove(binaryPath)
			
			cmd := exec.Command(binaryPath, tc.args...)
			var stderr bytes.Buffer
			cmd.Stderr = &stderr
			
			err := cmd.Run()
			
			if tc.expectError && err == nil {
				t.Error("Expected command to fail")
			}
			
			if tc.errorMsg != "" {
				stderrStr := stderr.String()
				if !strings.Contains(stderrStr, tc.errorMsg) {
					t.Errorf("Expected error message to contain: %s\nActual stderr: %s", tc.errorMsg, stderrStr)
				}
			}
		})
	}
}

// Helper function to capture output from a function that writes to stdout/stderr
func captureOutput(f func()) (stdout, stderr string) {
	// Save original
	origStdout := os.Stdout
	origStderr := os.Stderr
	
	// Create pipes
	rOut, wOut, _ := os.Pipe()
	rErr, wErr, _ := os.Pipe()
	
	// Replace stdout/stderr
	os.Stdout = wOut
	os.Stderr = wErr
	
	// Channel to capture output
	outCh := make(chan string)
	errCh := make(chan string)
	
	// Goroutines to read from pipes
	go func() {
		var buf bytes.Buffer
		io.Copy(&buf, rOut)
		outCh <- buf.String()
	}()
	
	go func() {
		var buf bytes.Buffer
		io.Copy(&buf, rErr)
		errCh <- buf.String()
	}()
	
	// Execute function
	f()
	
	// Close writers
	wOut.Close()
	wErr.Close()
	
	// Restore stdout/stderr
	os.Stdout = origStdout
	os.Stderr = origStderr
	
	// Get output
	stdout = <-outCh
	stderr = <-errCh
	
	return stdout, stderr
}

// TestRunCommand tests the vex run command
func TestRunCommand(t *testing.T) {
	// Build the binary for testing
	binaryPath := buildTestBinary(t)
	defer os.Remove(binaryPath)
	
	tempDir := t.TempDir()
	
	testCases := []struct {
		name           string
		inputContent   string
		expectError    bool
		expectedStdout string
		expectedStderr string
	}{
		{
			name:         "Simple constant definition",
			inputContent: "(def x 42)",
			expectError:  false,
		},
		{
			name:           "Hello world",
			inputContent:   "(import \"fmt\")\n(def message \"Hello from Vex!\")\n(fmt/Println message)",
			expectError:    false,
			expectedStdout: "Hello from Vex!",
		},
		{
			name:           "Simple macro",
			inputContent:   "(import \"fmt\")\n(macro say [text] (fmt/Println ~text))\n(say \"Macro works!\")",
			expectError:    false,
			expectedStdout: "Macro works!",
		},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create input file
			inputFile := filepath.Join(tempDir, "test.vx")
			err := os.WriteFile(inputFile, []byte(tc.inputContent), 0644)
			if err != nil {
				t.Fatalf("Failed to create input file: %v", err)
			}
			
			// Run the command
			cmd := exec.Command(binaryPath, "run", "-input", inputFile)
			var stdout, stderr bytes.Buffer
			cmd.Stdout = &stdout
			cmd.Stderr = &stderr
			
			err = cmd.Run()
			
			// Check error expectation
			if tc.expectError && err == nil {
				t.Error("Expected command to fail, but it succeeded")
			}
			if !tc.expectError && err != nil {
				t.Errorf("Expected command to succeed, but it failed: %v\nStderr: %s", err, stderr.String())
			}
			
			// Check output expectations
			if tc.expectedStdout != "" {
				if !strings.Contains(stdout.String(), tc.expectedStdout) {
					t.Errorf("Expected stdout to contain: %s\nActual stdout: %s", tc.expectedStdout, stdout.String())
				}
			}
			
			if tc.expectedStderr != "" {
				if !strings.Contains(stderr.String(), tc.expectedStderr) {
					t.Errorf("Expected stderr to contain: %s\nActual stderr: %s", tc.expectedStderr, stderr.String())
				}
			}
		})
	}
}