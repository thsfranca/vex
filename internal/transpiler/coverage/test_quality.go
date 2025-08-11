package coverage

import (
	"fmt"
	"os"
	"regexp"
	"strings"
)

// TestQualityMetrics represents the quality of tests for a package/function
type TestQualityMetrics struct {
	AssertionDensity     float64    `json:"assertion_density"`     // Assertions per test
	EdgeCaseScore        float64    `json:"edge_case_score"`       // How well edge cases are tested
	TestMethodDiversity  float64    `json:"test_method_diversity"` // Variety of testing approaches
	NamingQuality        float64    `json:"naming_quality"`        // Quality of test names
	OverallQualityScore  float64    `json:"overall_quality_score"` // Composite score 0-100
	Suggestions          []string   `json:"suggestions"`           // Improvement suggestions
	RedFlags             []string   `json:"red_flags"`             // Potential issues
}

// TestQualityAnalyzer evaluates the quality of test code
type TestQualityAnalyzer struct {
	assertPattern      *regexp.Regexp
	edgeCasePatterns   []*regexp.Regexp
	testNamePattern    *regexp.Regexp
	badNamePatterns    []*regexp.Regexp
}

// NewTestQualityAnalyzer creates a new test quality analyzer
func NewTestQualityAnalyzer() *TestQualityAnalyzer {
	return &TestQualityAnalyzer{
		assertPattern: regexp.MustCompile(`\(\s*assert-\w+`),
		
		// Patterns that suggest edge case testing
		edgeCasePatterns: []*regexp.Regexp{
			regexp.MustCompile(`\b(empty|nil|null|zero|negative|boundary|max|min|overflow|underflow)\b`),
			regexp.MustCompile(`\[\s*\]`),                    // Empty arrays
			regexp.MustCompile(`""\s*\)`),                    // Empty strings
			regexp.MustCompile(`\b-?\d+\b`),                  // Numeric values
			regexp.MustCompile(`\b(true|false)\b`),           // Boolean values
		},
		
		testNamePattern: regexp.MustCompile(`\(\s*deftest\s+([a-zA-Z][a-zA-Z0-9\-_]*)`),
		
		// Patterns that suggest poor test naming
		badNamePatterns: []*regexp.Regexp{
			regexp.MustCompile(`^test\d+$`),                  // test1, test2, etc.
			regexp.MustCompile(`^(basic|simple|test)$`),      // Generic names
			regexp.MustCompile(`^.{1,3}$`),                   // Too short
		},
	}
}

// AnalyzeTestQuality evaluates the quality of test files
func (tqa *TestQualityAnalyzer) AnalyzeTestQuality(testFiles []string, packageName string) (*TestQualityMetrics, error) {
	metrics := &TestQualityMetrics{
		Suggestions: []string{},
		RedFlags:    []string{},
	}
	
	if len(testFiles) == 0 {
		metrics.Suggestions = append(metrics.Suggestions, "No test files found - consider adding tests")
		return metrics, nil
	}
	
	var totalLines, totalAssertions, totalTests int
	var edgeCaseScore, namingScore float64
	testMethods := make(map[string]bool)
	
	for _, testFile := range testFiles {
		lines, assertions, tests, edgeScore, nameScore, methods, err := tqa.analyzeTestFile(testFile)
		if err != nil {
			continue // Skip files we can't analyze
		}
		
		totalLines += lines
		totalAssertions += assertions
		totalTests += tests
		edgeCaseScore += edgeScore
		namingScore += nameScore
		
		// Track test method diversity
		for method := range methods {
			testMethods[method] = true
		}
	}
	
	// Calculate metrics
	if totalTests > 0 {
		metrics.AssertionDensity = float64(totalAssertions) / float64(totalTests)
		metrics.EdgeCaseScore = edgeCaseScore / float64(len(testFiles)) // Average across files
		metrics.NamingQuality = namingScore / float64(len(testFiles))   // Average across files
	}
	
	metrics.TestMethodDiversity = float64(len(testMethods)) / 5.0 * 100 // Out of 5 common methods
	if metrics.TestMethodDiversity > 100 {
		metrics.TestMethodDiversity = 100
	}
	
	// Calculate overall quality score (weighted average)
	metrics.OverallQualityScore = (
		metrics.AssertionDensity * 30.0 +      // 30% weight
		metrics.EdgeCaseScore * 25.0 +         // 25% weight  
		metrics.TestMethodDiversity * 25.0 +   // 25% weight
		metrics.NamingQuality * 20.0)          // 20% weight
	
	// Cap at 100
	if metrics.OverallQualityScore > 100 {
		metrics.OverallQualityScore = 100
	}
	
	// Generate suggestions and red flags
	tqa.generateSuggestions(metrics, totalTests, packageName)
	
	return metrics, nil
}

// analyzeTestFile performs detailed analysis of a single test file
func (tqa *TestQualityAnalyzer) analyzeTestFile(testFile string) (lines, assertions, tests int, edgeScore, nameScore float64, methods map[string]bool, err error) {
	content, err := os.ReadFile(testFile)
	if err != nil {
		return 0, 0, 0, 0, 0, nil, err
	}
	
	methods = make(map[string]bool)
	fileContent := string(content)
	contentLines := strings.Split(fileContent, "\n")
	lines = len(contentLines)
	
	// Count assertions
	assertMatches := tqa.assertPattern.FindAllString(fileContent, -1)
	assertions = len(assertMatches)
	
	// Track assertion types for diversity
	for _, match := range assertMatches {
		methods[match] = true
	}
	
	// Count tests
	testMatches := tqa.testNamePattern.FindAllStringSubmatch(fileContent, -1)
	tests = len(testMatches)
	
	// Analyze edge case coverage
	edgeCaseMatches := 0
	for _, pattern := range tqa.edgeCasePatterns {
		matches := pattern.FindAllString(fileContent, -1)
		edgeCaseMatches += len(matches)
	}
	
	if lines > 0 {
		edgeScore = float64(edgeCaseMatches) / float64(lines) * 100
		if edgeScore > 100 {
			edgeScore = 100
		}
	}
	
	// Analyze test naming quality
	goodNames := 0
	for _, match := range testMatches {
		if len(match) > 1 {
			testName := match[1]
			isGoodName := true
			
			// Check against bad naming patterns
			for _, badPattern := range tqa.badNamePatterns {
				if badPattern.MatchString(testName) {
					isGoodName = false
					break
				}
			}
			
			// Good names should be descriptive (reasonable length)
			if len(testName) < 10 || len(testName) > 100 {
				isGoodName = false
			}
			
			if isGoodName {
				goodNames++
			}
		}
	}
	
	if tests > 0 {
		nameScore = float64(goodNames) / float64(tests) * 100
	}
	
	return lines, assertions, tests, edgeScore, nameScore, methods, nil
}

// generateSuggestions creates actionable improvement suggestions
func (tqa *TestQualityAnalyzer) generateSuggestions(metrics *TestQualityMetrics, totalTests int, packageName string) {
	// Assertion density suggestions
	if metrics.AssertionDensity < 1.0 {
		metrics.RedFlags = append(metrics.RedFlags, 
			fmt.Sprintf("Low assertion density (%.1f/test) - tests may not be verifying behavior", metrics.AssertionDensity))
		metrics.Suggestions = append(metrics.Suggestions, 
			"Add more assertions to verify function behavior and edge cases")
	} else if metrics.AssertionDensity > 10.0 {
		metrics.RedFlags = append(metrics.RedFlags, 
			fmt.Sprintf("Very high assertion density (%.1f/test) - consider splitting complex tests", metrics.AssertionDensity))
	}
	
	// Edge case suggestions
	if metrics.EdgeCaseScore < 20.0 {
		metrics.Suggestions = append(metrics.Suggestions, 
			"Add edge case testing: empty inputs, boundary values, nil/zero cases")
	}
	
	// Test method diversity suggestions
	if metrics.TestMethodDiversity < 40.0 {
		metrics.Suggestions = append(metrics.Suggestions, 
			"Use diverse assertion types: assert-eq, assert-true, assert-false, etc.")
	}
	
	// Naming quality suggestions
	if metrics.NamingQuality < 60.0 {
		metrics.Suggestions = append(metrics.Suggestions, 
			"Improve test names: use descriptive names that explain what is being tested")
	}
	
	// Overall quality suggestions
	if metrics.OverallQualityScore < 50.0 {
		metrics.RedFlags = append(metrics.RedFlags, 
			"Overall test quality is low - significant improvement needed")
		metrics.Suggestions = append(metrics.Suggestions, 
			fmt.Sprintf("Focus on %s package: increase assertions, add edge cases, improve naming", packageName))
	} else if metrics.OverallQualityScore > 80.0 {
		metrics.Suggestions = append(metrics.Suggestions, 
			"🏆 Excellent test quality! Consider this package as a template for others")
	}
}

// GenerateQualityReport creates a readable test quality report
func (tqa *TestQualityAnalyzer) GenerateQualityReport(metrics *TestQualityMetrics, packageName string) string {
	var report strings.Builder
	
	report.WriteString(fmt.Sprintf("🎯 Test Quality Score: %.1f/100\n", metrics.OverallQualityScore))
	
	// Quality breakdown
	report.WriteString("   Breakdown:\n")
	report.WriteString(fmt.Sprintf("     Assertion Density: %.1f/test\n", metrics.AssertionDensity))
	report.WriteString(fmt.Sprintf("     Edge Case Coverage: %.1f%%\n", metrics.EdgeCaseScore))
	report.WriteString(fmt.Sprintf("     Method Diversity: %.1f%%\n", metrics.TestMethodDiversity))
	report.WriteString(fmt.Sprintf("     Naming Quality: %.1f%%\n", metrics.NamingQuality))
	
	// Red flags
	if len(metrics.RedFlags) > 0 {
		report.WriteString("   🚩 Red Flags:\n")
		for _, flag := range metrics.RedFlags {
			report.WriteString(fmt.Sprintf("     %s\n", flag))
		}
	}
	
	// Suggestions
	if len(metrics.Suggestions) > 0 {
		report.WriteString("   💡 Suggestions:\n")
		for i, suggestion := range metrics.Suggestions {
			if i >= 3 { // Limit suggestions
				break
			}
			report.WriteString(fmt.Sprintf("     %s\n", suggestion))
		}
	}
	
	return report.String()
}
