package coverage

import (
	"os"
	"regexp"
	"strings"
)

// TestCorrelation handles mapping test cases to functions being tested
type TestCorrelation struct {
	// Patterns to detect function calls in test files
	functionCallPattern *regexp.Regexp
}

// TestedFunction represents a function that has test coverage
type TestedFunction struct {
	FunctionName string   `json:"function_name"`
	TestFiles    []string `json:"test_files"`
	TestCases    []string `json:"test_cases"`
}

// NewTestCorrelation creates a new test correlation engine
func NewTestCorrelation() *TestCorrelation {
	return &TestCorrelation{
		// Match function calls: (function-name args...)
		functionCallPattern: regexp.MustCompile(`\(\s*([a-zA-Z][a-zA-Z0-9\-_?]*)\s`),
	}
}

// AnalyzeTestFile parses a test file to find function calls
func (tc *TestCorrelation) AnalyzeTestFile(testFilePath string) ([]string, error) {
	content, err := os.ReadFile(testFilePath)
	if err != nil {
		return nil, err
	}
	
	var calledFunctions []string
	functionCalls := make(map[string]bool)
	
	lines := strings.Split(string(content), "\n")
	
	for _, line := range lines {
		// Skip comments and empty lines
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, ";;") {
			continue
		}
		
		// Find all function calls in this line
		matches := tc.functionCallPattern.FindAllStringSubmatch(line, -1)
		for _, match := range matches {
			if len(match) > 1 {
				funcName := match[1]
				
				// Filter out built-in forms and common test utilities
				if !tc.isBuiltinOrTestUtility(funcName) {
					functionCalls[funcName] = true
				}
			}
		}
	}
	
	// Convert map to slice
	for funcName := range functionCalls {
		calledFunctions = append(calledFunctions, funcName)
	}
	
	return calledFunctions, nil
}

// isBuiltinOrTestUtility checks if a function name is a built-in or test utility
func (tc *TestCorrelation) isBuiltinOrTestUtility(funcName string) bool {
	builtins := map[string]bool{
		// Core language forms
		"def": true, "defn": true, "defmacro": true, "macro": true,
		"if": true, "do": true, "fn": true, "let": true, "when": true, "unless": true,
		"import": true, "export": true,
		
		// Arithmetic and comparison
		"+": true, "-": true, "*": true, "/": true,
		">": true, "<": true, "=": true, ">=": true, "<=": true,
		
		// Array/collection operations
		"get": true, "slice": true, "len": true, "append": true,
		
		// Test framework
		"deftest": true, "assert-eq": true, "assert-true": true, "assert-false": true,
		
		// Go interop (common patterns)
		"fmt/Println": true, "fmt/Printf": true, "fmt/Print": true,
	}
	
	// Also filter out namespace calls (contain /)
	if strings.Contains(funcName, "/") {
		return true
	}
	
	return builtins[funcName]
}

// CorrelateTestsWithFunctions maps test files to the functions they test
func (tc *TestCorrelation) CorrelateTestsWithFunctions(pkg *PackageFunctions) map[string]*TestedFunction {
	testedFunctions := make(map[string]*TestedFunction)
	
	// Initialize all functions as untested
	for _, function := range pkg.Functions {
		testedFunctions[function.Name] = &TestedFunction{
			FunctionName: function.Name,
			TestFiles:    []string{},
			TestCases:    []string{},
		}
	}
	
	// Analyze each test file
	for _, testFile := range pkg.TestFiles {
		calledFunctions, err := tc.AnalyzeTestFile(testFile)
		if err != nil {
			continue // Skip files we can't parse
		}
		
		// Map called functions to this test file
		for _, funcName := range calledFunctions {
			if tested, exists := testedFunctions[funcName]; exists {
				// Add this test file to the function's coverage
				tested.TestFiles = append(tested.TestFiles, testFile)
			}
		}
	}
	
	return testedFunctions
}

// CalculateFunctionCoverage calculates the percentage of functions with tests
func (tc *TestCorrelation) CalculateFunctionCoverage(testedFunctions map[string]*TestedFunction) (int, int, float64) {
	totalFunctions := len(testedFunctions)
	testedCount := 0
	
	for _, tested := range testedFunctions {
		if len(tested.TestFiles) > 0 {
			testedCount++
		}
	}
	
	coverage := 0.0
	if totalFunctions > 0 {
		coverage = float64(testedCount) / float64(totalFunctions) * 100
	}
	
	return testedCount, totalFunctions, coverage
}
