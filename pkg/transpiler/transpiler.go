package transpiler

import (
	"regexp"
	"strings"

	parser "fugo/tools/gen/go"

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

func (t *Transpiler) TranspileToGo(input string) (string, error) {
	lexer := parser.NewFugoLexer(antlr.NewInputStream(input))
	stream := antlr.NewCommonTokenStream(lexer, 0)
	p := parser.NewFugoParser(stream)
	
	tree := p.Sp()
	
	t.output.Reset()
	t.output.WriteString("package main\n\n")
	t.output.WriteString("func main() {\n")
	
	t.transpileNode(tree)
	
	t.output.WriteString("}\n")
	
	return t.output.String(), nil
}

// transpileNode walks AST using visitor pattern
func (t *Transpiler) transpileNode(node antlr.Tree) {
	switch ctx := node.(type) {
	case *parser.SpContext:
		for i := 0; i < ctx.GetChildCount(); i++ {
			t.transpileNode(ctx.GetChild(i))
		}
	case *parser.ListContext:
		for i := 0; i < ctx.GetChildCount(); i++ {
			child := ctx.GetChild(i)
			t.transpileNode(child)
		}
	case antlr.TerminalNode:
		token := ctx.GetSymbol()
		if token.GetTokenType() == parser.FugoLexerSYMBOL {
			text := token.GetText()
			if t.isNumber(text) {
				t.output.WriteString("\tvar _ int = " + text + "\n")
			}
		} else if token.GetTokenType() == parser.FugoLexerSTRING {
			text := token.GetText()
			t.output.WriteString("\tvar _ string = " + text + "\n")
		}
	}
}


func (t *Transpiler) isNumber(s string) bool {
	matched, _ := regexp.MatchString("^[0-9]+$", s)
	return matched
}