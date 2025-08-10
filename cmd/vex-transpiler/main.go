package main

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/thsfranca/vex/internal/transpiler"
	"github.com/thsfranca/vex/internal/transpiler/packages"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]
	switch command {
	case "transpile":
		transpileCommand(os.Args[2:])
	case "run":
		runCommand(os.Args[2:])
	case "build":
		buildCommand(os.Args[2:])
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n\n", command)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Fprintf(os.Stderr, "Vex - A statically-typed functional programming language\n\n")
	fmt.Fprintf(os.Stderr, "Usage:\n")
	fmt.Fprintf(os.Stderr, "  vex transpile -input <file.vex> [-output <file.go>] [-verbose]\n")
	fmt.Fprintf(os.Stderr, "  vex run -input <file.vex> [-verbose]\n")
	fmt.Fprintf(os.Stderr, "  vex build -input <file.vex> [-output <binary>] [-verbose]\n\n")
	fmt.Fprintf(os.Stderr, "Commands:\n")
	fmt.Fprintf(os.Stderr, "  transpile  Transpile Vex source code to Go\n")
	fmt.Fprintf(os.Stderr, "  run        Compile and execute Vex source code\n")
	fmt.Fprintf(os.Stderr, "  build      Build Vex source code to binary executable\n\n")
	fmt.Fprintf(os.Stderr, "Examples:\n")
	fmt.Fprintf(os.Stderr, "  vex transpile -input example.vex -output example.go\n")
	fmt.Fprintf(os.Stderr, "  vex run -input example.vex\n")
	fmt.Fprintf(os.Stderr, "  vex build -input example.vex -output example\n")
}

func transpileCommand(args []string) {
	transpileFlags := flag.NewFlagSet("transpile", flag.ExitOnError)
	var (
		inputFile  = transpileFlags.String("input", "", "Input .vex file to transpile")
		outputFile = transpileFlags.String("output", "", "Output .go file (optional, defaults to stdout)")
		verbose    = transpileFlags.Bool("verbose", false, "Enable verbose output")
	)
	transpileFlags.Parse(args)

	if *inputFile == "" {
		fmt.Fprintf(os.Stderr, "Error: -input flag is required\n\n")
		printUsage()
		os.Exit(1)
	}

	if *verbose {
		fmt.Fprintf(os.Stderr, "🔄 Transpiling: %s\n", *inputFile)
	}

    // Read input file (sanity check)
    if _, err := os.ReadFile(*inputFile); err != nil {
        fmt.Fprintf(os.Stderr, "❌ Error reading file %s: %v\n", *inputFile, err)
        os.Exit(1)
    }

    // Load core.vx if present
    coreContent := loadCoreVex(*verbose)

    // Resolve packages and build full program
    moduleRoot, _ := filepath.Abs(".")
    resolver := packages.NewResolver(moduleRoot)
    res, err := resolver.BuildProgramFromEntry(*inputFile)
    if err != nil {
        fmt.Fprintf(os.Stderr, "❌ Package resolution error: %v\n", err)
        os.Exit(1)
    }
    fullProgram := coreContent + "\n" + res.CombinedSource

    // Create transpiler with local package imports ignored in Go output
    tCfg := transpiler.TranspilerConfig{
        EnableMacros:     true,
        CoreMacroPath:    "",
        PackageName:      "main",
        GenerateComments: true,
        IgnoreImports:    res.IgnoreImports,
        Exports:          res.Exports,
    }
    tImpl, err := transpiler.NewTranspilerWithConfig(tCfg)
    if err != nil {
        fmt.Fprintf(os.Stderr, "❌ Transpiler init error: %v\n", err)
        os.Exit(1)
    }

    // Transpile
    goCode, err := tImpl.TranspileFromInput(fullProgram)
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Transpilation error: %v\n", err)
		os.Exit(1)
	}

	// Output result
	if *outputFile != "" {
        err = os.WriteFile(*outputFile, []byte(goCode), 0644)
		if err != nil {
			fmt.Fprintf(os.Stderr, "❌ Error writing output file %s: %v\n", *outputFile, err)
			os.Exit(1)
		}
		if *verbose {
			fmt.Fprintf(os.Stderr, "✅ Transpilation complete: %s\n", *outputFile)
		}
	} else {
		// Output to stdout
		fmt.Print(goCode)
	}
}

func runCommand(args []string) {
	runFlags := flag.NewFlagSet("run", flag.ExitOnError)
	var (
		inputFile = runFlags.String("input", "", "Input .vex file to run")
		verbose   = runFlags.Bool("verbose", false, "Enable verbose output")
	)
	runFlags.Parse(args)

	if *inputFile == "" {
		fmt.Fprintf(os.Stderr, "Error: -input flag is required\n\n")
		printUsage()
		os.Exit(1)
	}

	if *verbose {
		fmt.Fprintf(os.Stderr, "🚀 Running Vex file: %s\n", *inputFile)
	}

	// Load core.vx if it exists
	coreContent := loadCoreVex(*verbose)

    // Read user input file (kept for future use; resolution does its own reading)
    _, err := os.ReadFile(*inputFile)
    if err != nil {
        fmt.Fprintf(os.Stderr, "❌ Error reading file %s: %v\n", *inputFile, err)
        os.Exit(1)
    }

    // Resolve packages and build full program
    moduleRoot, _ := filepath.Abs(".")
    resolver := packages.NewResolver(moduleRoot)
    res, err := resolver.BuildProgramFromEntry(*inputFile)
    if err != nil {
        fmt.Fprintf(os.Stderr, "❌ Package resolution error: %v\n", err)
        os.Exit(1)
    }
    fullProgram := coreContent + "\n" + res.CombinedSource

    // Create transpiler with local package imports ignored in Go output
    tCfg := transpiler.TranspilerConfig{
        EnableMacros:     true,
        CoreMacroPath:    "",
        PackageName:      "main",
        GenerateComments: true,
        IgnoreImports:    res.IgnoreImports,
        Exports:          res.Exports,
    }
    tImpl, err := transpiler.NewTranspilerWithConfig(tCfg)
    if err != nil {
        fmt.Fprintf(os.Stderr, "❌ Transpiler init error: %v\n", err)
        os.Exit(1)
    }

	// Transpile combined program
    goCode, err := tImpl.TranspileFromInput(fullProgram)
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Transpilation error: %v\n", err)
		os.Exit(1)
	}

	if *verbose {
		fmt.Fprintf(os.Stderr, "🔄 Transpilation complete, executing...\n")
	}

	// Create temporary file for Go code
	tmpDir := os.TempDir()
	baseName := strings.TrimSuffix(filepath.Base(*inputFile), filepath.Ext(*inputFile))
	tmpGoFile := filepath.Join(tmpDir, baseName+"_temp.go")

	// Write Go code to temporary file
    err = os.WriteFile(tmpGoFile, []byte(goCode), 0644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Error writing temporary Go file: %v\n", err)
		os.Exit(1)
	}

	// Clean up temporary file when done
	defer func() {
		if err := os.Remove(tmpGoFile); err != nil && *verbose {
			fmt.Fprintf(os.Stderr, "⚠️  Warning: could not remove temporary file %s: %v\n", tmpGoFile, err)
		}
	}()

	// Build the Go code
	executable := strings.TrimSuffix(tmpGoFile, ".go")
	cmd := exec.Command("go", "build", "-o", executable, tmpGoFile)
	buildOutput, buildErr := cmd.CombinedOutput()

	if buildErr != nil {
		fmt.Fprintf(os.Stderr, "❌ Build error: %v\n%s", buildErr, string(buildOutput))
		os.Exit(1)
	}

	// Clean up executable when done
	defer os.Remove(executable)

	// Run the executable
	runCmd := exec.Command(executable)
	runCmd.Stdout = os.Stdout
	runCmd.Stderr = os.Stderr

	err = runCmd.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Execution error: %v\n", err)
		os.Exit(1)
	}

	if *verbose {
		fmt.Fprintf(os.Stderr, "✅ Execution complete\n")
	}
}

func buildCommand(args []string) {
	buildFlags := flag.NewFlagSet("build", flag.ExitOnError)
	var (
		inputFile  = buildFlags.String("input", "", "Input .vex file to build")
		outputFile = buildFlags.String("output", "", "Output binary file (optional, defaults to input filename)")
		verbose    = buildFlags.Bool("verbose", false, "Enable verbose output")
	)
	buildFlags.Parse(args)

	if *inputFile == "" {
		fmt.Fprintf(os.Stderr, "Error: -input flag is required\n\n")
		printUsage()
		os.Exit(1)
	}

	// Determine output filename
	var outputBinary string
	if *outputFile == "" {
		// Default to input filename without extension
		baseName := strings.TrimSuffix(filepath.Base(*inputFile), filepath.Ext(*inputFile))
		outputBinary = baseName
	} else {
		outputBinary = *outputFile
	}

	if *verbose {
		fmt.Fprintf(os.Stderr, "🔨 Building Vex file: %s -> %s\n", *inputFile, outputBinary)
	}

	// Load core.vx if it exists
	coreContent := loadCoreVex(*verbose)

    // Read user input file (kept for future use; resolution does its own reading)
    _, err := os.ReadFile(*inputFile)
    if err != nil {
        fmt.Fprintf(os.Stderr, "❌ Error reading file %s: %v\n", *inputFile, err)
        os.Exit(1)
    }

    // Resolve packages and build full program
    moduleRoot, _ := filepath.Abs(".")
    resolver := packages.NewResolver(moduleRoot)
    res, err := resolver.BuildProgramFromEntry(*inputFile)
    if err != nil {
        fmt.Fprintf(os.Stderr, "❌ Package resolution error: %v\n", err)
        os.Exit(1)
    }
    fullProgram := coreContent + "\n" + res.CombinedSource

    // Create transpiler with local package imports ignored in Go output
    tCfg := transpiler.TranspilerConfig{
        EnableMacros:     true,
        CoreMacroPath:    "",
        PackageName:      "main",
        GenerateComments: true,
        IgnoreImports:    res.IgnoreImports,
    }
    tImpl, err := transpiler.NewTranspilerWithConfig(tCfg)
    if err != nil {
        fmt.Fprintf(os.Stderr, "❌ Transpiler init error: %v\n", err)
        os.Exit(1)
    }

	// Transpile combined program
    goCode, err := tImpl.TranspileFromInput(fullProgram)
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Transpilation error: %v\n", err)
		os.Exit(1)
	}

	if *verbose {
		fmt.Fprintf(os.Stderr, "🔄 Transpilation complete, building binary...\n")
	}

	// Create temporary directory for build
	tmpDir := os.TempDir()
	buildDir := filepath.Join(tmpDir, "vex-build-"+strings.ReplaceAll(time.Now().Format("20060102-150405"), ":", ""))
    if err := os.MkdirAll(buildDir, 0o755); err != nil {
		fmt.Fprintf(os.Stderr, "❌ Error creating build directory: %v\n", err)
		os.Exit(1)
	}

	// Clean up temporary directory when done
	defer func() {
		if err := os.RemoveAll(buildDir); err != nil && *verbose {
			fmt.Fprintf(os.Stderr, "⚠️  Warning: could not remove temporary build directory %s: %v\n", buildDir, err)
		}
	}()

	// Generate go.mod with detected dependencies
    detectedModules := tImpl.GetDetectedModules()
	if err := generateGoMod(buildDir, detectedModules, *verbose); err != nil {
		fmt.Fprintf(os.Stderr, "❌ Error generating go.mod: %v\n", err)
		os.Exit(1)
	}

	// Write Go code to build directory
	mainGoFile := filepath.Join(buildDir, "main.go")
    if err := os.WriteFile(mainGoFile, []byte(goCode), 0o644); err != nil {
		fmt.Fprintf(os.Stderr, "❌ Error writing Go file: %v\n", err)
		os.Exit(1)
	}

	// Download dependencies if needed
	if len(detectedModules) > 0 {
		if *verbose {
			fmt.Fprintf(os.Stderr, "📦 Downloading dependencies...\n")
		}
		if err := downloadDependencies(buildDir, *verbose); err != nil {
			fmt.Fprintf(os.Stderr, "❌ Error downloading dependencies: %v\n", err)
			os.Exit(1)
		}
	}

	// Build binary to final location
	absOutputPath, err := filepath.Abs(outputBinary)
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Error getting absolute output path: %v\n", err)
		os.Exit(1)
	}

	cmd := exec.Command("go", "build", "-o", absOutputPath, "main.go")
	cmd.Dir = buildDir
	
	buildOutput, buildErr := cmd.CombinedOutput()
	if buildErr != nil {
		fmt.Fprintf(os.Stderr, "❌ Build error: %v\n%s", buildErr, string(buildOutput))
		os.Exit(1)
	}

	if *verbose {
		fmt.Fprintf(os.Stderr, "✅ Binary built successfully: %s\n", absOutputPath)
	} else {
		fmt.Printf("Binary built: %s\n", outputBinary)
	}
}

func loadCoreVex(verbose bool) string {
	// Try to load core.vx from current directory
	coreContent, err := os.ReadFile("core.vx")
	if err != nil {
		if verbose {
			fmt.Fprintf(os.Stderr, "ℹ️  core.vx not found, using minimal bootstrap\n")
		}
		// Return minimal bootstrap if core.vx doesn't exist
		return `; Minimal Vex bootstrap`
	}

	if verbose {
		fmt.Fprintf(os.Stderr, "📚 Loaded core.vx standard library\n")
	}
	return string(coreContent)
}

func generateGoMod(vexDir string, modules map[string]string, verbose bool) error {
	goModPath := filepath.Join(vexDir, "go.mod")
	
	// Create basic go.mod
	content := fmt.Sprintf("module vex-project-%d\n\ngo 1.21\n", time.Now().Unix())
	
	// Add detected dependencies
	if len(modules) > 0 {
		content += "\nrequire (\n"
		for module, version := range modules {
			// For "latest", let go mod tidy resolve the version
			content += fmt.Sprintf("\t%s %s\n", module, version)
		}
		content += ")\n"
		
		if verbose {
			fmt.Fprintf(os.Stderr, "📦 Added dependencies: %v\n", modules)
		}
	}
	
	return os.WriteFile(goModPath, []byte(content), 0o644)
}

func downloadDependencies(vexDir string, verbose bool) error {
	cmd := exec.Command("go", "mod", "tidy")
	cmd.Dir = vexDir
	
	if verbose {
		cmd.Stdout = os.Stderr
		cmd.Stderr = os.Stderr
	}
	
	return cmd.Run()
}

func buildBinary(vexDir, genDir, binaryPath string, verbose bool) error {
	// Build from the module directory using the relative path to main.go
	relativeMainPath := "./gen/main.go"
	relativeBinaryPath := "./bin/app" // Use relative path for output too
	
	cmd := exec.Command("go", "build", "-o", relativeBinaryPath, relativeMainPath)
	cmd.Dir = vexDir
	
	if verbose {
		cmd.Stdout = os.Stderr
		cmd.Stderr = os.Stderr
		fmt.Fprintf(os.Stderr, "🔨 Building binary: go build -o %s %s (in %s)\n", relativeBinaryPath, relativeMainPath, vexDir)
	}
	
	// Always capture stderr for error reporting
	var stderr bytes.Buffer
	if !verbose {
		cmd.Stderr = &stderr
	}
	
	if err := cmd.Run(); err != nil {
		if verbose {
			return fmt.Errorf("build failed: %v", err)
		} else {
			return fmt.Errorf("build failed: %v\n%s", err, stderr.String())
		}
	}
	
	return nil
}
