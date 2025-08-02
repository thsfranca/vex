package transpiler

import (
	parser "fugo/tools/gen/go"
	"strings"

	"github.com/antlr4-go/antlr/v4"
)

// Transpiler converts Fugo AST to Go code
type Transpiler struct {
	output strings.Builder
}

// New creates a new transpiler instance
func New() *Transpiler {
	return &Transpiler{}
}

// TranspileToGo converts a Fugo program to Go code
func (t *Transpiler) TranspileToGo(input string) (string, error) {
	// Parse the input using ANTLR
	lexer := parser.NewFugoLexer(antlr.NewInputStream(input))
	stream := antlr.NewCommonTokenStream(lexer, 0)
	p := parser.NewFugoParser(stream)
	
	// Parse starting from the root rule
	tree := p.Sp()
	
	// For now, just acknowledge we parsed successfully
	t.output.Reset()
	t.output.WriteString("package main\n\n")
	t.output.WriteString("func main() {\n")
	t.output.WriteString("\t// Transpiled from Fugo: " + input + "\n")
	t.output.WriteString("\t// Parse tree created successfully\n")
	t.output.WriteString("}\n")
	
	// Prevent unused variable warning
	_ = tree
	
	return t.output.String(), nil
}