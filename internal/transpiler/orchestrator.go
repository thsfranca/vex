package transpiler

import (
	"fmt"
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







// Build creates the configured transpiler
func (b *TranspilerBuilder) Build() (*VexTranspiler, error) {
	// Create parser
	parser := NewParserAdapter()

	// Create macro system
	var macroSystem MacroExpander
	if b.config.EnableMacros {
		macroConfig := macro.Config{
			CoreMacroPath:    b.config.CoreMacroPath,
			StdlibPath:       b.config.StdlibPath,
			EnableValidation: true,
		}
		registry := macro.NewRegistry(macroConfig)

		// Core macros are now loaded only via explicit imports
		// Auto-loading removed to enforce explicit stdlib imports

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

	// Phase 0: Extract and load stdlib imports before macro expansion
	if vt.config.EnableMacros && vt.macroSystem != nil {
		if err := vt.loadStdlibImports(input); err != nil {
			return "", fmt.Errorf("stdlib import error: %v", err)
		}
	}

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

// loadStdlibImports extracts import statements from source and loads required stdlib modules
func (vt *VexTranspiler) loadStdlibImports(input string) error {
	// Parse import statements from the source using a simple regex approach
	// This is a lightweight pre-processing step before full parsing
	lines := strings.Split(input, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Look for import statements: (import vex.module) or (import "vex.module")
		if strings.HasPrefix(line, "(import ") && strings.HasSuffix(line, ")") {
			// Extract module name between "import " and ")"
			importContent := strings.TrimSpace(line[8 : len(line)-1]) // Remove "(import " and ")"

			// Remove quotes if present
			moduleName := strings.Trim(importContent, "\"")

			// Only handle stdlib modules (starting with "vex.")
			if strings.HasPrefix(moduleName, "vex.") {
				if err := vt.macroSystem.LoadStdlibModule(moduleName); err != nil {
					return fmt.Errorf("failed to load stdlib module '%s': %v", moduleName, err)
				}
				// Add to IgnoreImports so it doesn't get output as a Go import
				if vt.config.IgnoreImports == nil {
					vt.config.IgnoreImports = make(map[string]bool)
				}
				vt.config.IgnoreImports[moduleName] = true
			}
		}
	}

	return nil
}









// GetDetectedModules returns detected third-party modules
func (vt *VexTranspiler) GetDetectedModules() map[string]string {
	// Basic implementation of third-party module detection
	// This should be enhanced to work with the actual import tracking in the future
	return vt.detectedModules
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
