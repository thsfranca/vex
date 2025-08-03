package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/thsfranca/vex/internal/transpiler"
)

func main() {
	var (
		inputFile  = flag.String("input", "", "Input .vex file to transpile")
		outputFile = flag.String("output", "", "Output .go file (optional, defaults to stdout)")
		verbose    = flag.Bool("verbose", false, "Enable verbose output")
	)
	flag.Parse()

	if *inputFile == "" {
		fmt.Fprintf(os.Stderr, "Usage: %s -input <file.vex> [-output <file.go>] [-verbose]\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "\nTranspiles Vex source code to Go\n")
		fmt.Fprintf(os.Stderr, "\nExample:\n")
		fmt.Fprintf(os.Stderr, "  %s -input example.vex -output example.go\n", os.Args[0])
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