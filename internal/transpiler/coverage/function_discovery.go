package coverage

import (
	"fmt"
	"os"
	"regexp"
	"strings"
)

// FunctionInfo represents a discovered function or macro
type FunctionInfo struct {
	Name     string `json:"name"`
	Type     string `json:"type"` // "defn", "defmacro", "def"
	File     string `json:"file"`
	Line     int    `json:"line"`
	Package  string `json:"package"`
}

// PackageFunctions holds all functions discovered in a package
type PackageFunctions struct {
	Package   string          `json:"package"`
	Functions []FunctionInfo  `json:"functions"`
	TestFiles []string        `json:"test_files"`
}

// FunctionDiscovery handles parsing Vex source files to find function definitions
type FunctionDiscovery struct {
	// Regex patterns for different function types
	defnPattern     *regexp.Regexp
	defmacroPattern *regexp.Regexp
	defPattern      *regexp.Regexp
	macroPattern    *regexp.Regexp
}

// NewFunctionDiscovery creates a new function discovery engine
func NewFunctionDiscovery() *FunctionDiscovery {
	return &FunctionDiscovery{
		// Match (defn function-name [params] -> returnType body)
		defnPattern: regexp.MustCompile(`\(\s*defn\s+([a-zA-Z][a-zA-Z0-9\-_?]*)`),
		
		// Match (defmacro macro-name [params] body)
		defmacroPattern: regexp.MustCompile(`\(\s*defmacro\s+([a-zA-Z][a-zA-Z0-9\-_?]*)`),
		
		// Match (def variable-name value) but only for function values
		defPattern: regexp.MustCompile(`\(\s*def\s+([a-zA-Z][a-zA-Z0-9\-_?]*)\s+\(\s*fn\s`),
		
		// Match (macro macro-name [params] body)
		macroPattern: regexp.MustCompile(`\(\s*macro\s+([a-zA-Z][a-zA-Z0-9\-_?]*)`),
	}
}

// DiscoverFunctions parses a Vex source file and extracts function definitions
func (fd *FunctionDiscovery) DiscoverFunctions(filePath string, packageName string) ([]FunctionInfo, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %v", filePath, err)
	}
	
	var functions []FunctionInfo
	lines := strings.Split(string(content), "\n")
	
	for lineNum, line := range lines {
		// Skip comments and empty lines
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, ";;") {
			continue
		}
		
		// Check for defn functions
		if matches := fd.defnPattern.FindStringSubmatch(line); len(matches) > 1 {
			functions = append(functions, FunctionInfo{
				Name:    matches[1],
				Type:    "defn",
				File:    filePath,
				Line:    lineNum + 1,
				Package: packageName,
			})
		}
		
		// Check for defmacro macros
		if matches := fd.defmacroPattern.FindStringSubmatch(line); len(matches) > 1 {
			functions = append(functions, FunctionInfo{
				Name:    matches[1],
				Type:    "defmacro",
				File:    filePath,
				Line:    lineNum + 1,
				Package: packageName,
			})
		}
		
		// Check for def + fn combinations
		if matches := fd.defPattern.FindStringSubmatch(line); len(matches) > 1 {
			functions = append(functions, FunctionInfo{
				Name:    matches[1],
				Type:    "def+fn",
				File:    filePath,
				Line:    lineNum + 1,
				Package: packageName,
			})
		}
		
		// Check for macro definitions
		if matches := fd.macroPattern.FindStringSubmatch(line); len(matches) > 1 {
			functions = append(functions, FunctionInfo{
				Name:    matches[1],
				Type:    "macro",
				File:    filePath,
				Line:    lineNum + 1,
				Package: packageName,
			})
		}
	}
	
	return functions, nil
}

// DiscoverPackageFunctions scans a package directory for all functions
func (fd *FunctionDiscovery) DiscoverPackageFunctions(packageDir string, packageName string) (*PackageFunctions, error) {
	pkg := &PackageFunctions{
		Package:   packageName,
		Functions: []FunctionInfo{},
		TestFiles: []string{},
	}
	
	entries, err := os.ReadDir(packageDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read package directory %s: %v", packageDir, err)
	}
	
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		
		fileName := entry.Name()
		filePath := packageDir + "/" + fileName
		
		// Check if it's a Vex source file
		if strings.HasSuffix(fileName, ".vx") || strings.HasSuffix(fileName, ".vex") {
			if strings.HasSuffix(fileName, "_test.vx") || strings.HasSuffix(fileName, "_test.vex") {
				// Test file
				pkg.TestFiles = append(pkg.TestFiles, filePath)
			} else {
				// Source file - discover functions
				functions, err := fd.DiscoverFunctions(filePath, packageName)
				if err != nil {
					// Log error but continue with other files
					fmt.Printf("Warning: failed to parse %s: %v\n", filePath, err)
					continue
				}
				pkg.Functions = append(pkg.Functions, functions...)
			}
		}
	}
	
	return pkg, nil
}
