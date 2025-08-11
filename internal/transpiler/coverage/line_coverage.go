package coverage

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"
)

// LineInfo represents a single line of code with coverage data
type LineInfo struct {
	LineNumber int    `json:"line_number"`
	Content    string `json:"content"`
	IsCode     bool   `json:"is_code"`      // Is this an actual code line?
	IsCovered  bool   `json:"is_covered"`   // Is this line covered by tests?
	HitCount   int    `json:"hit_count"`    // How many tests hit this line
	TestFiles  []string `json:"test_files"` // Which test files cover this line
}

// FileCoverage represents line-by-line coverage for a single file
type FileCoverage struct {
	FilePath       string     `json:"file_path"`
	Package        string     `json:"package"`
	Lines          []LineInfo `json:"lines"`
	TotalLines     int        `json:"total_lines"`
	CodeLines      int        `json:"code_lines"`
	CoveredLines   int        `json:"covered_lines"`
	CoveragePercent float64   `json:"coverage_percent"`
	UncoveredLines []int      `json:"uncovered_lines"`
}

// LineCoverageAnalyzer performs line-level coverage analysis
type LineCoverageAnalyzer struct {
	// Patterns to identify code vs comments/whitespace
	codePattern    *regexp.Regexp
	commentPattern *regexp.Regexp
	importPattern  *regexp.Regexp
}

// NewLineCoverageAnalyzer creates a new line coverage analyzer
func NewLineCoverageAnalyzer() *LineCoverageAnalyzer {
	return &LineCoverageAnalyzer{
		// Match lines with actual code (function calls, definitions, etc.)
		codePattern: regexp.MustCompile(`\(\s*([a-zA-Z][a-zA-Z0-9\-_?*/+<>=!]*)`),
		
		// Match comment lines
		commentPattern: regexp.MustCompile(`^\s*(;;|$)`),
		
		// Match import statements
		importPattern: regexp.MustCompile(`\(\s*import\s+`),
	}
}

// AnalyzeFile performs detailed line-by-line coverage analysis
func (lca *LineCoverageAnalyzer) AnalyzeFile(filePath string, packageName string, testedFunctions map[string]*TestedFunction) (*FileCoverage, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file %s: %v", filePath, err)
	}
	defer file.Close()
	
	coverage := &FileCoverage{
		FilePath: filePath,
		Package:  packageName,
		Lines:    []LineInfo{},
	}
	
	scanner := bufio.NewScanner(file)
	lineNumber := 1
	
	for scanner.Scan() {
		line := scanner.Text()
		lineInfo := lca.analyzeLine(line, lineNumber, testedFunctions)
		coverage.Lines = append(coverage.Lines, lineInfo)
		
		// Count totals
		coverage.TotalLines++
		if lineInfo.IsCode {
			coverage.CodeLines++
			if lineInfo.IsCovered {
				coverage.CoveredLines++
			} else {
				coverage.UncoveredLines = append(coverage.UncoveredLines, lineNumber)
			}
		}
		
		lineNumber++
	}
	
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading file %s: %v", filePath, err)
	}
	
	// Calculate coverage percentage
	if coverage.CodeLines > 0 {
		coverage.CoveragePercent = float64(coverage.CoveredLines) / float64(coverage.CodeLines) * 100
	}
	
	return coverage, nil
}

// analyzeLine determines if a line is code and if it's covered
func (lca *LineCoverageAnalyzer) analyzeLine(line string, lineNumber int, testedFunctions map[string]*TestedFunction) LineInfo {
	lineInfo := LineInfo{
		LineNumber: lineNumber,
		Content:    line,
		TestFiles:  []string{},
	}
	
	trimmed := strings.TrimSpace(line)
	
	// Skip empty lines and comments
	if trimmed == "" || lca.commentPattern.MatchString(trimmed) {
		return lineInfo
	}
	
	// Skip import statements (considered infrastructure, not testable code)
	if lca.importPattern.MatchString(trimmed) {
		return lineInfo
	}
	
	// Check if this line contains code
	if lca.codePattern.MatchString(trimmed) {
		lineInfo.IsCode = true
		
		// Determine coverage by checking if any functions called on this line are tested
		matches := lca.codePattern.FindAllStringSubmatch(trimmed, -1)
		for _, match := range matches {
			if len(match) > 1 {
				funcName := match[1]
				
				// Check if this function is tested
				if tested, exists := testedFunctions[funcName]; exists && len(tested.TestFiles) > 0 {
					lineInfo.IsCovered = true
					lineInfo.HitCount++
					
					// Add unique test files
					for _, testFile := range tested.TestFiles {
						if !contains(lineInfo.TestFiles, testFile) {
							lineInfo.TestFiles = append(lineInfo.TestFiles, testFile)
						}
					}
				}
			}
		}
	}
	
	return lineInfo
}

// contains checks if a slice contains a string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// GenerateLineCoverageReport creates a detailed line-by-line coverage report
func (lca *LineCoverageAnalyzer) GenerateLineCoverageReport(fileCoverage *FileCoverage) string {
	var report strings.Builder
	
	report.WriteString(fmt.Sprintf("📄 Line Coverage: %s\n", fileCoverage.FilePath))
	report.WriteString(fmt.Sprintf("   Coverage: %.1f%% (%d/%d lines)\n", 
		fileCoverage.CoveragePercent, fileCoverage.CoveredLines, fileCoverage.CodeLines))
	
	if len(fileCoverage.UncoveredLines) > 0 {
		report.WriteString(fmt.Sprintf("   Uncovered lines: %v\n", fileCoverage.UncoveredLines))
		
		// Show a few uncovered line samples
		report.WriteString("   Samples:\n")
		count := 0
		for _, lineNum := range fileCoverage.UncoveredLines {
			if count >= 3 {
				break
			}
			if lineNum <= len(fileCoverage.Lines) {
				line := fileCoverage.Lines[lineNum-1]
				report.WriteString(fmt.Sprintf("     L%d: %s\n", lineNum, strings.TrimSpace(line.Content)))
				count++
			}
		}
	}
	
	return report.String()
}
