package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s <debug-info|skip-simulation|build-only|test-only|lint-only>\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "\nCommands:\n")
		fmt.Fprintf(os.Stderr, "  debug-info       - Show debug information about repository and environment\n")
		fmt.Fprintf(os.Stderr, "  skip-simulation  - Simulate skip scenario for performance testing\n")
		fmt.Fprintf(os.Stderr, "  build-only       - Build Go packages only\n")
		fmt.Fprintf(os.Stderr, "  test-only        - Run tests only\n")
		fmt.Fprintf(os.Stderr, "  lint-only        - Run linting only\n")
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "debug-info":
		showDebugInfo()
	case "skip-simulation":
		simulateSkip()
	case "build-only":
		buildOnly()
	case "test-only":
		testOnly()
	case "lint-only":
		lintOnly()
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", command)
		os.Exit(1)
	}
}

func showDebugInfo() {
	fmt.Println("🔍 DEBUG INFORMATION")
	fmt.Println("===================")

	// Environment info
	fmt.Printf("Test type: %s\n", getEnv("TEST_TYPE", "unknown"))
	fmt.Printf("Debug mode: %s\n", getEnv("DEBUG_MODE", "false"))
	fmt.Printf("Event: %s\n", getEnv("GITHUB_EVENT_NAME", "unknown"))
	fmt.Printf("Ref: %s\n", getEnv("GITHUB_REF", "unknown"))
	fmt.Printf("SHA: %s\n", getEnv("GITHUB_SHA", "unknown"))
	fmt.Println("")

	// Repository structure
	fmt.Println("📂 Repository structure:")
	cmd := exec.Command("find", ".", "-type", "f", "-name", "*.go")
	output, err := cmd.Output()
	if err == nil {
		lines := strings.Split(strings.TrimSpace(string(output)), "\n")
		for i, line := range lines {
			if i >= 10 {
				break
			}
			if line != "" {
				fmt.Println(line)
			}
		}
	}
	fmt.Println("")

	// Go modules
	fmt.Println("📦 Go modules:")
	if content, err := os.ReadFile("go.mod"); err == nil {
		fmt.Println(string(content))
	} else {
		fmt.Println("No go.mod found")
	}
}

func simulateSkip() {
	fmt.Println("✅ Simulating skip scenario for non-Go changes")
	fmt.Println("This is what happens when only docs/VSCode extension files change")
	fmt.Println("Total time: ~10-15 seconds")
	fmt.Println("vs. old workflow: ~8-10 minutes")
}

func buildOnly() {
	fmt.Println("🔨 Building Go packages...")

	// Check if there are any main.go files
	cmd := exec.Command("find", ".", "-name", "main.go")
	output, err := cmd.Output()

	if err != nil || strings.TrimSpace(string(output)) == "" {
		fmt.Println("ℹ️ No main.go files found - nothing to build")
		return
	}

	// Build all packages
	fmt.Println("Building all Go packages...")
	cmd = exec.Command("go", "build", "./...")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		fmt.Printf("❌ Build failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("✅ Build completed successfully")
}

func testOnly() {
	fmt.Println("🧪 Running tests...")

	// Check if there are any test files
	cmd := exec.Command("find", ".", "-name", "*_test.go")
	output, err := cmd.Output()

	if err != nil || strings.TrimSpace(string(output)) == "" {
		fmt.Println("ℹ️ No test files found - nothing to test")
		return
	}

	// Run tests
	fmt.Println("Running Go tests...")
	cmd = exec.Command("go", "test", "-v", "./...")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		fmt.Printf("❌ Tests failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("✅ Tests completed successfully")
}

func lintOnly() {
	fmt.Println("🔍 Running linting...")

	// Run go vet
	fmt.Println("Running go vet...")
	cmd := exec.Command("go", "vet", "./...")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		fmt.Printf("❌ Linting failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("✅ Linting completed successfully")
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
