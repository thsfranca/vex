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

// Expression represents a Fugo expression for transpilation
type Expression struct {
	Type     string   // "literal", "arithmetic"
	Value    string   // for literals
	Operator string   // for arithmetic
	Operands []string // for arithmetic
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
		if expr := t.parseArithmetic(ctx); expr != nil {
			t.output.WriteString("\tvar _ int = " + expr.Operands[0] + " " + expr.Operator + " " + expr.Operands[1] + "\n")
		} else {
			for i := 0; i < ctx.GetChildCount(); i++ {
				child := ctx.GetChild(i)
				t.transpileNode(child)
			}
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

// parseArithmetic checks if a list represents arithmetic operation
func (t *Transpiler) parseArithmetic(ctx *parser.ListContext) *Expression {
	children := make([]string, 0)
	
	// Extract text content from list children
	for i := 0; i < ctx.GetChildCount(); i++ {
		child := ctx.GetChild(i)
		if terminal, ok := child.(antlr.TerminalNode); ok {
			if terminal.GetSymbol().GetTokenType() == parser.FugoLexerSYMBOL {
				children = append(children, terminal.GetText())
			}
		}
	}
	
	// Check for simple addition: (+ operand1 operand2)
	if len(children) == 3 && children[0] == "+" {
		if t.isNumber(children[1]) && t.isNumber(children[2]) {
			return &Expression{
				Type:     "arithmetic",
				Operator: "+",
				Operands: []string{children[1], children[2]},
			}
		}
	}
	
	return nil
}