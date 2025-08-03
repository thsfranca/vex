package main

import (
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

type Component struct {
	Name        string
	Target      int
	TestPath    string
	CoverFile   string
	DirPath     string
	Implemented bool
	Coverage    float64
}

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s <generate-coverage|update-readme>\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "\nCommands:\n")
		fmt.Fprintf(os.Stderr, "  generate-coverage  - Generate Go coverage reports\n")
		fmt.Fprintf(os.Stderr, "  update-readme      - Update README with coverage table\n")
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "generate-coverage":
		generateCoverage()
	case "update-readme":
		updateReadme()
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", command)
		os.Exit(1)
	}
}

func generateCoverage() {
	fmt.Println("[COVERAGE] Generating Go coverage reports...")

	// Create coverage directory
	os.MkdirAll("coverage", 0755)

	components := []Component{
		{"Parser", 95, "./internal/frontend/parser/...", "coverage/parser.out", "internal/frontend/parser", false, 0},
		{"Transpiler", 90, "./internal/transpiler/...", "coverage/transpiler.out", "internal/transpiler", false, 0},
		{"Types", 85, "./internal/types/...", "coverage/types.out", "internal/types", false, 0},
		{"Stdlib", 80, "./stdlib/...", "coverage/stdlib.out", "stdlib", false, 0},
	}

	// Run total coverage
	fmt.Println("Running total coverage analysis...")
	runCommand("go", "test", "-v", "-coverprofile=coverage/total.out", "-coverpkg=./...", "./...")

	// Run component coverage
	for i := range components {
		comp := &components[i]
		fmt.Printf("Running %s coverage analysis...\n", comp.Name)

		// Check if component is implemented
		if _, err := os.Stat(comp.DirPath); err == nil {
			comp.Implemented = true
			runCommand("go", "test", "-v", "-coverprofile="+comp.CoverFile, comp.TestPath)
			comp.Coverage = calculateCoverage(comp.CoverFile)
		} else {
			// Create empty coverage file for non-implemented components
			os.WriteFile(comp.CoverFile+"_coverage.txt", []byte("0"), 0644)
		}
	}

	// Calculate total coverage
	totalCoverage := calculateCoverage("coverage/total.out")

	// Output results for environment variables
	for _, comp := range components {
		status := generateStatus(comp.Coverage, comp.Target, comp.Implemented)
		fmt.Printf("%s_STATUS=%s\n", strings.ToUpper(comp.Name), status)
	}

	// Handle total coverage status
	var totalStatus string
	if totalCoverage == 0 {
		totalStatus = "⏳ *No tests yet*"
	} else if totalCoverage >= 75 {
		totalStatus = fmt.Sprintf("[SUCCESS] %.1f%%", totalCoverage)
	} else {
		totalStatus = fmt.Sprintf("[ERROR] %.1f%% (below 75%%)", totalCoverage)
	}
	fmt.Printf("TOTAL_STATUS=%s\n", totalStatus)
}

func updateReadme() {
	fmt.Println("[UPDATE] Updating README with coverage table...")

	// Read environment variables for coverage status
	parserStatus := getEnv("PARSER_STATUS", "⏳ *Not implemented yet*")
	transpilerStatus := getEnv("TRANSPILER_STATUS", "⏳ *Not implemented yet*")
	typesStatus := getEnv("TYPES_STATUS", "⏳ *Not implemented yet*")
	stdlibStatus := getEnv("STDLIB_STATUS", "⏳ *Not implemented yet*")
	totalStatus := getEnv("TOTAL_STATUS", "⏳ *No tests yet*")

	// Create coverage table
	coverageTable := fmt.Sprintf(`| Component | Target | Status | Purpose |
|-----------|--------|--------|---------|
| **Parser** | 95%%+ | %s | Critical language component |
| **Transpiler** | 90%%+ | %s | Core functionality |
| **Type System** | 85%%+ | %s | Type safety |
| **Standard Library** | 80%%+ | %s | User-facing features |
| **Overall Project** | 75%%+ | %s | Quality baseline |

> **Quality Philosophy**: Higher coverage requirements for more critical components. PRs that reduce coverage below these thresholds are automatically blocked.`,
		parserStatus, transpilerStatus, typesStatus, stdlibStatus, totalStatus)

	// Read README.md
	readmeContent, err := os.ReadFile("README.md")
	if err != nil {
		fmt.Printf("Error reading README.md: %v\n", err)
		os.Exit(1)
	}

	// Remove existing coverage table (between Component.*Target.*Status.*Purpose and Quality Philosophy)
	re := regexp.MustCompile(`(?s)\| Component.*?\| Target.*?\| Status.*?\| Purpose.*?\n.*?> \*\*Quality Philosophy\*\*.*?\n`)
	cleanedContent := re.ReplaceAllString(string(readmeContent), "")

	// Find ## Project Status section and insert coverage table
	projectStatusRe := regexp.MustCompile(`(## Project Status\n)`)
	updatedContent := projectStatusRe.ReplaceAllString(cleanedContent, "${1}\n"+coverageTable+"\n\n")

	// Write updated README
	err = os.WriteFile("README.md", []byte(updatedContent), 0644)
	if err != nil {
		fmt.Printf("Error writing README.md: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("[SUCCESS] README.md updated successfully")
}

func runCommand(name string, args ...string) {
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Run() // Ignore errors for coverage commands that might fail on missing packages
}

func calculateCoverage(filename string) float64 {
	if _, err := os.Stat(filename); err != nil {
		return 0
	}

	cmd := exec.Command("go", "tool", "cover", "-func="+filename)
	output, err := cmd.Output()
	if err != nil {
		return 0
	}

	// Parse total coverage from output
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.Contains(line, "total:") {
			parts := strings.Fields(line)
			if len(parts) >= 3 {
				coverageStr := strings.TrimSuffix(parts[2], "%")
				if coverage, err := strconv.ParseFloat(coverageStr, 64); err == nil {
					return coverage
				}
			}
		}
	}
	return 0
}

func generateStatus(coverage float64, target int, implemented bool) string {
	if !implemented {
		return "⏳ *Not implemented yet*"
	}

	if coverage >= float64(target) {
		return fmt.Sprintf("[SUCCESS] %.1f%%", coverage)
	}

	return fmt.Sprintf("[ERROR] %.1f%% (below %d%%)", coverage, target)
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
