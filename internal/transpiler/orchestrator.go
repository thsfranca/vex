package transpiler

import (
	"fmt"
	"os"
	"strings"

	"github.com/thsfranca/vex/internal/transpiler/codegen"
	"github.com/thsfranca/vex/internal/transpiler/macro"
)

// VexTranspiler creates a transpiler with clean architecture
type VexTranspiler struct {
	parser          Parser
	macroSystem     MacroExpander
	analyzer        Analyzer
	codeGen         CodeGenerator
	config          TranspilerConfig
	detectedModules map[string]string
}

// TranspilerBuilder builds a transpiler with configurable components
type TranspilerBuilder struct {
	config TranspilerConfig
}

// NewBuilder creates a new transpiler builder
func NewBuilder() *TranspilerBuilder {
	return &TranspilerBuilder{
		config: TranspilerConfig{
			EnableMacros:     true,
			CoreMacroPath:    "", // Will be determined dynamically
			PackageName:      "main",
			GenerateComments: true,
            IgnoreImports:    make(map[string]bool),
            Exports:          make(map[string]map[string]bool),
		},
	}
}

// WithConfig sets the transpiler configuration
func (b *TranspilerBuilder) WithConfig(config TranspilerConfig) *TranspilerBuilder {
	b.config = config
	return b
}

// WithMacros enables or disables macro support
func (b *TranspilerBuilder) WithMacros(enabled bool) *TranspilerBuilder {
	b.config.EnableMacros = enabled
	return b
}

// WithCoreMacroPath sets the path to core macro definitions
func (b *TranspilerBuilder) WithCoreMacroPath(path string) *TranspilerBuilder {
	b.config.CoreMacroPath = path
	return b
}

// WithPackageName sets the package name for generated code
func (b *TranspilerBuilder) WithPackageName(name string) *TranspilerBuilder {
	b.config.PackageName = name
	return b
}

// Build creates the configured transpiler
func (b *TranspilerBuilder) Build() (*VexTranspiler, error) {
	// Create parser
	parser := NewParserAdapter()

	// Create macro system
	var macroSystem MacroExpander
	if b.config.EnableMacros {
		macroConfig := macro.Config{
			CoreMacroPath:    b.config.CoreMacroPath,
			EnableValidation: true,
		}
		registry := macro.NewRegistry(macroConfig)
		
		// Load core macros
		if err := registry.LoadCoreMacros(); err != nil {
			return nil, fmt.Errorf("failed to load core macros: %v", err)
		}
		
		macroSystem = NewMacroExpanderAdapter(macro.NewMacroExpander(registry))
	}

    // Create analyzer
    analyzer := NewAnalyzerAdapter()
    // Provide package environment for analyzer (package exports/ignore/schemes)
    analyzer.SetPackageEnv(b.config.IgnoreImports, b.config.Exports, b.config.PkgSchemes)

	// Create code generator
	codeGenConfig := codegen.Config{
		PackageName:      b.config.PackageName,
		GenerateComments: b.config.GenerateComments,
		IndentSize:       4,
        IgnoreImports:    b.config.IgnoreImports,
	}
	codeGen := NewCodeGeneratorAdapter(codeGenConfig)

	return &VexTranspiler{
		parser:          parser,
		macroSystem:     macroSystem,
		analyzer:        analyzer,
		codeGen:         codeGen,
		config:          b.config,
		detectedModules: make(map[string]string),
	}, nil
}

// TranspileFromInput transpiles Vex source code to Go
func (vt *VexTranspiler) TranspileFromInput(input string) (string, error) {
	// Reset detected modules for each transpilation
	vt.detectedModules = make(map[string]string)
	
	// Detect third-party modules from imports
	vt.detectThirdPartyModules(input)
	
	// Phase 1: Parse
	ast, err := vt.parser.Parse(input)
	if err != nil {
		return "", fmt.Errorf("parse error: %v", err)
	}

	// Phase 2: Macro expansion (if enabled)
	if vt.config.EnableMacros && vt.macroSystem != nil {
		ast, err = vt.macroSystem.ExpandMacros(ast)
		if err != nil {
			return "", fmt.Errorf("macro expansion error: %v", err)
		}
	}

	// Phase 3: Semantic analysis
	symbolTable, err := vt.analyzer.Analyze(ast)
	if err != nil {
		return "", fmt.Errorf("analysis error: %v", err)
	}

	// Phase 4: Code generation
	code, err := vt.codeGen.Generate(ast, symbolTable)
	if err != nil {
		return "", fmt.Errorf("code generation error: %v", err)
	}

	return code, nil
}

// TranspileFromFile transpiles a Vex file to Go
func (vt *VexTranspiler) TranspileFromFile(filename string) (string, error) {
	// Read file content and use TranspileFromInput
	content, err := os.ReadFile(filename)
	if err != nil {
		return "", fmt.Errorf("failed to read file %s: %v", filename, err)
	}
	
	return vt.TranspileFromInput(string(content))
}

// GetMacroSystem returns the macro system (for testing/debugging)
func (vt *VexTranspiler) GetMacroSystem() MacroExpander {
	return vt.macroSystem
}

// GetAnalyzer returns the analyzer (for testing/debugging)
func (vt *VexTranspiler) GetAnalyzer() Analyzer {
	return vt.analyzer
}

// GetCodeGenerator returns the code generator (for testing/debugging)
func (vt *VexTranspiler) GetCodeGenerator() CodeGenerator {
	return vt.codeGen
}

// GetDetectedModules returns detected third-party modules
func (vt *VexTranspiler) GetDetectedModules() map[string]string {
	// Basic implementation of third-party module detection
	// This should be enhanced to work with the actual import tracking in the future
	return vt.detectedModules
}

// Helper method to convert AST back to string (for macro processing)
func (vt *VexTranspiler) astToString(ast AST) string {
	// This is a simplified implementation
	// In a real system, you'd want a proper AST-to-string converter
	return "/* AST conversion not implemented yet */"
}

// NewVexTranspiler builds a VexTranspiler with default configuration.
func NewVexTranspiler() (*VexTranspiler, error) {
	return NewBuilder().Build()
}

// NewTranspilerWithConfig builds a VexTranspiler using the provided configuration.
func NewTranspilerWithConfig(config TranspilerConfig) (*VexTranspiler, error) {
	return NewBuilder().WithConfig(config).Build()
}

// Example usage demonstrating the clean architecture:
//
// transpiler, err := NewBuilder().
//     WithMacros(true).
//     WithCoreMacroPath("custom/macros.vx").
//     WithPackageName("mypackage").
//     Build()
// if err != nil {
//     return err
// }
//
// result, err := transpiler.TranspileFromInput(vexCode)
// if err != nil {
//     return err
// }
//
// fmt.Println(result)

// Note: Legacy adapter removed; tests migrated to use current APIs.

// detectThirdPartyModules scans input for import statements and identifies third-party modules
func (vt *VexTranspiler) detectThirdPartyModules(input string) {
	// Simple regex-based detection for now - this could be enhanced to use the actual parser
	lines := strings.Split(input, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "(import \"") && strings.HasSuffix(line, "\")") {
			// Extract the import path
			start := strings.Index(line, "\"") + 1
			end := strings.LastIndex(line, "\"")
			if start < end {
				importPath := line[start:end]
				if vt.isThirdPartyModule(importPath) {
					vt.detectedModules[importPath] = "third-party"
				}
			}
		}
	}
}

// isThirdPartyModule determines if an import path is a third-party module
func (vt *VexTranspiler) isThirdPartyModule(importPath string) bool {
	// Standard library packages are typically short names without dots
	// Third-party packages usually have domain names
	return strings.Contains(importPath, ".") || strings.HasPrefix(importPath, "golang.org/")
}