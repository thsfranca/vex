package analysis

import (
	"testing"

	"github.com/antlr4-go/antlr/v4"
	"github.com/thsfranca/vex/internal/transpiler/parser"
)

func parseProgram(src string) *parser.ProgramContext {
    input := antlr.NewInputStream(src)
    lex := parser.NewVexLexer(input)
    tokens := antlr.NewCommonTokenStream(lex, 0)
    p := parser.NewVexParser(tokens)
    return p.Program().(*parser.ProgramContext)
}

type astAdapter struct{ prog *parser.ProgramContext }

func (a *astAdapter) Accept(v ASTVisitor) error { return v.VisitProgram(a.prog) }

func BenchmarkAnalyzer_Simple(b *testing.B) {
    src := "(def x 1) (+ x 2)"
    prog := parseProgram(src)
    ast := &astAdapter{prog: prog}
    b.ReportAllocs()
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        a := NewAnalyzer()
        if _, err := a.Analyze(ast); err != nil {
            b.Fatal(err)
        }
    }
}

func BenchmarkAnalyzer_Mixed(b *testing.B) {
    // Keep analyzer happy: avoid undefined symbols by not calling functions
    src := "(def a 10) (def b 20) (if (> a 0) a b)"
    prog := parseProgram(src)
    ast := &astAdapter{prog: prog}
    b.ReportAllocs()
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        a := NewAnalyzer()
        if _, err := a.Analyze(ast); err != nil {
            b.Fatal(err)
        }
    }
}


