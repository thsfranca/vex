package transpiler

import (
	"testing"

	"github.com/antlr4-go/antlr/v4"
	"github.com/thsfranca/vex/internal/transpiler/parser"
)

// Moved from interfaces_extra_test.go
// Verify ConcreteAST.Accept covers invalid root path
func TestConcreteAST_Accept_InvalidRoot(t *testing.T) {
    // Build a List without wrapping in Program
    list := parser.NewVexParser(antlr.NewCommonTokenStream(parser.NewVexLexer(antlr.NewInputStream("(+ 1 2)")), 0)).List()
    ast := &ConcreteAST{root: list}
    if err := ast.Accept(nil); err == nil {
        t.Fatalf("expected error on invalid root type")
    }
}


