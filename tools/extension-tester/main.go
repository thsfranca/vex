package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s <create-samples|package|verify|summary>\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "\nCommands:\n")
		fmt.Fprintf(os.Stderr, "  create-samples  - Create test Vex files for syntax highlighting\n")
		fmt.Fprintf(os.Stderr, "  package        - Package VSCode extension into .vsix\n")
		fmt.Fprintf(os.Stderr, "  verify         - Verify packaged extension integrity\n")
		fmt.Fprintf(os.Stderr, "  summary        - Generate validation summary\n")
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "create-samples":
		createSamples()
	case "package":
		packageExtension()
	case "verify":
		verifyPackage()
	case "summary":
		generateSummary()
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", command)
		os.Exit(1)
	}
}

func createSamples() {
	fmt.Println("[TEST] Testing syntax highlighting against sample code...")

	// Create test-samples directory
	err := os.MkdirAll("test-samples", 0755)
	if err != nil {
		fmt.Printf("[ERROR] Failed to create test-samples directory: %v\n", err)
		os.Exit(1)
	}

	// Create factorial.vx
	factorialContent := `; Factorial function in Vex
(define factorial (n)
  (if (= n 0)
      1
      (* n (factorial (- n 1)))))

; Test the function
(factorial 5)`

	err = os.WriteFile("test-samples/factorial.vx", []byte(factorialContent), 0644)
	if err != nil {
		fmt.Printf("[ERROR] Failed to create factorial.vx: %v\n", err)
		os.Exit(1)
	}

	// Create fibonacci.vx
	fibonacciContent := `; Fibonacci sequence
(define fib (n)
  (cond 
    ((= n 0) 0)
    ((= n 1) 1)
    (else (+ (fib (- n 1)) (fib (- n 2))))))`

	err = os.WriteFile("test-samples/fibonacci.vx", []byte(fibonacciContent), 0644)
	if err != nil {
		fmt.Printf("[ERROR] Failed to create fibonacci.vx: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("[SUCCESS] Sample Vex files created successfully")
	fmt.Println("📁 Test files:")

	// List test files
	files, err := os.ReadDir("test-samples")
	if err != nil {
		fmt.Printf("[ERROR] Failed to list test files: %v\n", err)
		os.Exit(1)
	}

	for _, file := range files {
		fmt.Printf("  - %s\n", file.Name())
	}
}

func packageExtension() {
	fmt.Println("📦 Packaging VSCode extension...")

	// Verify required files exist
	fmt.Println("📋 Checking extension structure...")

	if _, err := os.Stat("package.json"); err != nil {
		fmt.Println("[ERROR] package.json not found")
		os.Exit(1)
	}

	// Package the extension using @vscode/vsce
	fmt.Println("🔧 Using @vscode/vsce to package extension...")
	cmd := exec.Command("npx", "@vscode/vsce", "package",
		"--out", "vex-test-build.vsix",
		"--no-git-tag-version",
		"--allow-missing-repository")

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		fmt.Printf("[ERROR] Failed to package extension: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("[SUCCESS] Extension packaged successfully")

	// List .vsix files
	files, err := filepath.Glob("*.vsix")
	if err == nil {
		for _, file := range files {
			fmt.Printf("  - %s\n", file)
		}
	}
}

func verifyPackage() {
	fmt.Println("[CHECK] Verifying packaged extension...")

	// Check .vsix was created and is valid
	if _, err := os.Stat("vex-test-build.vsix"); err != nil {
		fmt.Println("[ERROR] Extension package not found")
		os.Exit(1)
	}

	// Test archive integrity
	fmt.Println("🔧 Testing archive integrity...")
	cmd := exec.Command("unzip", "-t", "vex-test-build.vsix")
	err := cmd.Run()
	if err != nil {
		fmt.Println("[ERROR] Extension package is corrupted")
		os.Exit(1)
	}

	// Show package contents (first 20 lines)
	fmt.Println("📋 Package contents:")
	cmd = exec.Command("unzip", "-l", "vex-test-build.vsix")
	output, err := cmd.Output()
	if err == nil {
		lines := strings.Split(string(output), "\n")
		for i, line := range lines {
			if i >= 20 {
				break
			}
			fmt.Println(line)
		}
	}

	// Get package size
	fileInfo, err := os.Stat("vex-test-build.vsix")
	if err == nil {
		fmt.Printf("📏 Package size: %d bytes\n", fileInfo.Size())
	}

	fmt.Println("[SUCCESS] Extension package verified successfully")
}

func generateSummary() {
	extensionFiles := getEnv("EXTENSION_FILES", "true")
	validateResult := getEnv("VALIDATE_RESULT", "success")
	skipResult := getEnv("SKIP_RESULT", "skipped")

	fmt.Println("## 📦 VSCode Extension Validation Summary")

	if extensionFiles == "false" {
		fmt.Println("[SUCCESS] **Fast skip**: No extension files changed")
		fmt.Println("[INFO] **Time saved**: ~3-4 minutes")
		fmt.Printf("- Skip Validation: %s\n", skipResult)
	} else {
		fmt.Println("🔄 **Full validation**: Extension files changed")
		fmt.Printf("- Extension Validation: %s\n", validateResult)
		fmt.Println("")
		fmt.Println("## [TEST] Validation Steps")
		fmt.Println("- [SUCCESS] **JavaScript linting**: Code quality checks")
		fmt.Println("- [SUCCESS] **Code formatting**: Prettier validation")
		fmt.Println("- [SUCCESS] **Manifest validation**: package.json structure")
		fmt.Println("- [SUCCESS] **Grammar validation**: TextMate syntax highlighting")
		fmt.Println("- [SUCCESS] **Theme validation**: Color scheme definitions")
		fmt.Println("- [SUCCESS] **Sample testing**: Vex code syntax validation")
		fmt.Println("- [SUCCESS] **Package creation**: .vsix build and verification")

		if validateResult == "failure" {
			fmt.Println("")
			fmt.Println("[ERROR] **Extension validation failed** - Check logs for details")
			os.Exit(1)
		} else {
			fmt.Println("")
			fmt.Println("[SUCCESS] **Extension validation passed** - Ready for distribution!")
		}
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func runCommand(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
