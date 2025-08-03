package main

import (
	"fmt"

	"github.com/antlr4-go/antlr/v4"
	// Adjust import path as needed
)

func main() {
	// Example Flux code to parse
	input := `
		; Example Flux program
		(package main)
		(def greet [name] (print "Hello" name))
		[1 2 3 "hello"]
	`

	// Create input stream
	inputStream := antlr.NewInputStream(input)

	// Create lexer
	lexer := parser.NewFluxLexer(inputStream)

	// Create token stream
	tokenStream := antlr.NewCommonTokenStream(lexer, 0)

	// Create parser
	p := parser.NewFluxParser(tokenStream)

	// Add error listener
	p.AddErrorListener(antlr.NewDiagnosticErrorListener(true))

	// Parse starting from the 'sp' rule (root rule)
	tree := p.Sp()

	// Print the parse tree
	fmt.Println("Parse tree:")
	fmt.Println(tree.ToStringTree(p.GetRuleNames(), p))

	// Create and use a custom listener
	listener := &FluxListener{}
	antlr.ParseTreeWalkerDefault.Walk(listener, tree)
}

// FluxListener is a custom listener for Flux parse events
type FluxListener struct {
	*parser.BaseFluxListener
}

// EnterSp is called when entering the root rule
func (l *FluxListener) EnterSp(ctx *parser.SpContext) {
	fmt.Printf("Entering program with %d expressions\n", len(ctx.AllList()))
}

// EnterList is called when entering a list expression
func (l *FluxListener) EnterList(ctx *parser.ListContext) {
	fmt.Printf("Found list with %d elements\n", ctx.GetChildCount()-2) // -2 for parentheses
}

// EnterArray is called when entering an array expression
func (l *FluxListener) EnterArray(ctx *parser.ArrayContext) {
	fmt.Printf("Found array with %d elements\n", ctx.GetChildCount()-2) // -2 for brackets
}