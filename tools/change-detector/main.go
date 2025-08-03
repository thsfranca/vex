package main

import (
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s <makefile-analysis|final-decision> [basic-go-files] [makefile-go-related]\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "\nCommands:\n")
		fmt.Fprintf(os.Stderr, "  makefile-analysis  - Analyze if Makefile changes are Go-related\n")
		fmt.Fprintf(os.Stderr, "  final-decision     - Make final decision about running Go tests\n")
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "makefile-analysis":
		analyzeMarkfileChanges()
	case "final-decision":
		makeFinalDecision()
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", command)
		os.Exit(1)
	}
}

func analyzeMarkfileChanges() {
	fmt.Println("🔍 Analyzing Makefile changes for Go relevance...")

	// Get git diff for Makefile
	cmd := exec.Command("git", "diff", "origin/main...HEAD", "Makefile")
	output, err := cmd.Output()
	if err != nil {
		fmt.Printf("go-related=false\n")
		return
	}

	diff := string(output)

	// Check if Makefile diff contains Go-related patterns
	goPattern := regexp.MustCompile(`(go (build|test|mod|generate)|GOPATH|GOOS|GOARCH|\.go|antlr.*-Dlanguage=Go)`)

	if goPattern.MatchString(diff) {
		fmt.Println("✅ Found Go-related Makefile changes:")

		// Show matching lines
		lines := strings.Split(diff, "\n")
		for _, line := range lines {
			if goPattern.MatchString(line) {
				fmt.Println(line)
			}
		}

		fmt.Println("go-related=true")
	} else {
		fmt.Println("⚠️ Makefile changes appear to be non-Go related (tooling, documentation, etc.)")
		fmt.Println("Changed lines preview:")

		// Show first 20 lines of diff
		lines := strings.Split(diff, "\n")
		for i, line := range lines {
			if i >= 20 {
				break
			}
			fmt.Println(line)
		}

		fmt.Println("go-related=false")
	}
}

func makeFinalDecision() {
	basicGo := getEnvOrArg("BASIC_GO", 2, "false")
	makefileGo := getEnvOrArg("MAKEFILE_GO", 3, "false")

	if basicGo == "true" || makefileGo == "true" {
		fmt.Println("🎯 Final decision: GO TESTS NEEDED")
		fmt.Printf("  - Basic Go files: %s\n", basicGo)
		fmt.Printf("  - Go-related Makefile: %s\n", makefileGo)
		fmt.Println("go-files=true")
	} else {
		fmt.Println("⚡ Final decision: SKIP GO TESTS")
		fmt.Printf("  - Basic Go files: %s\n", basicGo)
		fmt.Printf("  - Go-related Makefile: %s\n", makefileGo)
		fmt.Println("go-files=false")
	}
}

func getEnvOrArg(envVar string, argIndex int, defaultValue string) string {
	// First try environment variable
	if value := os.Getenv(envVar); value != "" {
		return value
	}

	// Then try command line argument
	if len(os.Args) > argIndex {
		return os.Args[argIndex]
	}

	return defaultValue
}
