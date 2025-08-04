package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/thsfranca/vex/internal/transpiler"
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
	case "exec":
		execCommand(os.Args[2:])
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
	fmt.Fprintf(os.Stderr, "  vex exec <file.vex> [-verbose]\n\n")
	fmt.Fprintf(os.Stderr, "Commands:\n")
	fmt.Fprintf(os.Stderr, "  transpile  Transpile Vex source code to Go\n")
	fmt.Fprintf(os.Stderr, "  run        Transpile and run Vex source code\n")
	fmt.Fprintf(os.Stderr, "  exec       Execute Vex file directly (includes core.vx)\n\n")
	fmt.Fprintf(os.Stderr, "Examples:\n")
	fmt.Fprintf(os.Stderr, "  vex transpile -input example.vex -output example.go\n")
	fmt.Fprintf(os.Stderr, "  vex run -input example.vex\n")
	fmt.Fprintf(os.Stderr, "  vex exec example.vex\n")
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

	// Read input file
	content, err := ioutil.ReadFile(*inputFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Error reading file %s: %v\n", *inputFile, err)
		os.Exit(1)
	}

	// Create transpiler
	t := transpiler.New()

	// Transpile
	goCode, err := t.TranspileFromInput(string(content))
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Transpilation error: %v\n", err)
		os.Exit(1)
	}

	// Output result
	if *outputFile != "" {
		err = ioutil.WriteFile(*outputFile, []byte(goCode), 0644)
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

	// Read input file
	content, err := ioutil.ReadFile(*inputFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Error reading file %s: %v\n", *inputFile, err)
		os.Exit(1)
	}

	// Create transpiler
	t := transpiler.New()

	// Transpile
	goCode, err := t.TranspileFromInput(string(content))
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
	err = ioutil.WriteFile(tmpGoFile, []byte(goCode), 0644)
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

	// Run the Go code
	// Note: Go will complain about unused variables, but the code will compile if syntax is correct
	cmd := exec.Command("go", "build", "-o", strings.TrimSuffix(tmpGoFile, ".go"), tmpGoFile)
	var buildErr error
	buildOutput, buildErr := cmd.CombinedOutput()

	if buildErr != nil {
		fmt.Fprintf(os.Stderr, "❌ Build error: %v\n%s", buildErr, string(buildOutput))
		os.Exit(1)
	}

	// If build succeeded, run the executable
	executable := strings.TrimSuffix(tmpGoFile, ".go")
	defer os.Remove(executable) // Clean up executable

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

func execCommand(args []string) {
	execFlags := flag.NewFlagSet("exec", flag.ExitOnError)
	var (
		verbose = execFlags.Bool("verbose", false, "Enable verbose output")
	)
	execFlags.Parse(args)

	// Get input file from remaining args
	remainingArgs := execFlags.Args()
	if len(remainingArgs) == 0 {
		fmt.Fprintf(os.Stderr, "Error: input file is required\n\n")
		printUsage()
		os.Exit(1)
	}

	inputFile := remainingArgs[0]

	if *verbose {
		fmt.Fprintf(os.Stderr, "🚀 Executing Vex file: %s\n", inputFile)
	}

	// Load core.vx if it exists
	coreContent := loadCoreVex(*verbose)

	// Read user input file
	userContent, err := ioutil.ReadFile(inputFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Error reading file %s: %v\n", inputFile, err)
		os.Exit(1)
	}

	// Combine core + user code
	fullProgram := coreContent + "\n" + string(userContent)

	// Create transpiler
	t := transpiler.New()

	// Transpile combined program
	goCode, err := t.TranspileFromInput(fullProgram)
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Transpilation error: %v\n", err)
		os.Exit(1)
	}

	if *verbose {
		fmt.Fprintf(os.Stderr, "🔄 Transpilation complete, executing...\n")
		fmt.Fprintf(os.Stderr, "Generated Go code:\n%s\n", goCode)
	}

	// Create .vex directory structure
	vexDir := ".vex"
	genDir := filepath.Join(vexDir, "gen")
	binDir := filepath.Join(vexDir, "bin")
	
	if err := os.MkdirAll(genDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "❌ Error creating .vex directory: %v\n", err)
		os.Exit(1)
	}
	if err := os.MkdirAll(binDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "❌ Error creating .vex/bin directory: %v\n", err)
		os.Exit(1)
	}

	// Generate go.mod with detected dependencies
	detectedModules := t.GetDetectedModules()
	if err := generateGoMod(vexDir, detectedModules, *verbose); err != nil {
		fmt.Fprintf(os.Stderr, "❌ Error generating go.mod: %v\n", err)
		os.Exit(1)
	}

	// Write Go code to .vex/gen/main.go
	mainGoFile := filepath.Join(genDir, "main.go")
	if err := ioutil.WriteFile(mainGoFile, []byte(goCode), 0644); err != nil {
		fmt.Fprintf(os.Stderr, "❌ Error writing Go file: %v\n", err)
		os.Exit(1)
	}

	if *verbose {
		fmt.Fprintf(os.Stderr, "📁 Generated Go module in %s/\n", vexDir)
	}

	// Download dependencies
	if len(detectedModules) > 0 {
		if *verbose {
			fmt.Fprintf(os.Stderr, "📦 Downloading dependencies...\n")
		}
		if err := downloadDependencies(vexDir, *verbose); err != nil {
			fmt.Fprintf(os.Stderr, "❌ Error downloading dependencies: %v\n", err)
			os.Exit(1)
		}
	}

	// Build binary
	binaryPath := filepath.Join(binDir, "app")
	if err := buildBinary(vexDir, genDir, binaryPath, *verbose); err != nil {
		fmt.Fprintf(os.Stderr, "❌ Build error: %v\n", err)
		os.Exit(1)
	}

	// Execute binary (use absolute path)
	absBinaryPath, err := filepath.Abs(binaryPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Error getting absolute path: %v\n", err)
		os.Exit(1)
	}
	
	cmd := exec.Command(absBinaryPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err = cmd.Run()
	if err != nil {
		if *verbose {
			fmt.Fprintf(os.Stderr, "❌ Execution error: %v\n", err)
		}
		os.Exit(1)
	}

	if *verbose {
		fmt.Fprintf(os.Stderr, "✅ Execution complete\n")
	}
}

func loadCoreVex(verbose bool) string {
	// Try to load core.vx from current directory
	coreContent, err := ioutil.ReadFile("core.vx")
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
	
	return ioutil.WriteFile(goModPath, []byte(content), 0644)
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
