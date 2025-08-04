package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

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
	fmt.Fprintf(os.Stderr, "  vex run -input <file.vex> [-verbose]\n\n")
	fmt.Fprintf(os.Stderr, "Commands:\n")
	fmt.Fprintf(os.Stderr, "  transpile  Transpile Vex source code to Go\n")
	fmt.Fprintf(os.Stderr, "  run        Transpile and run Vex source code\n\n")
	fmt.Fprintf(os.Stderr, "Examples:\n")
	fmt.Fprintf(os.Stderr, "  vex transpile -input example.vex -output example.go\n")
	fmt.Fprintf(os.Stderr, "  vex run -input example.vex\n")
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
