package codegen

import (
	"testing"

	"github.com/antlr4-go/antlr/v4"
	"github.com/thsfranca/vex/internal/transpiler/parser"
)

// benchAST adapts a parsed ProgramContext to the local AST interface.
type benchAST struct{ prog *parser.ProgramContext }

func (a *benchAST) Accept(v ASTVisitor) error { return v.VisitProgram(a.prog) }

// stubSymbols satisfies SymbolTable; generator does not depend on it.
type stubSymbols struct{}

func (s *stubSymbols) Define(name string, value Value) error { return nil }
func (s *stubSymbols) Lookup(name string) (Value, bool) { return nil, false }
func (s *stubSymbols) EnterScope()                     {}
func (s *stubSymbols) ExitScope()                      {}

func parseProgramBench(src string) *parser.ProgramContext {
    input := antlr.NewInputStream(src)
    lex := parser.NewVexLexer(input)
    tokens := antlr.NewCommonTokenStream(lex, 0)
    p := parser.NewVexParser(tokens)
    return p.Program().(*parser.ProgramContext)
}

func BenchmarkCodegen_Simple(b *testing.B) {
    src := "(def x 42)\n(fmt/Println x)"
    prog := parseProgramBench(src)
    ast := &benchAST{prog: prog}
    gen := NewGoCodeGenerator(Config{PackageName: "main", GenerateComments: false, IndentSize: 4})
    syms := &stubSymbols{}
    b.ReportAllocs()
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        _, err := gen.Generate(ast, syms)
        if err != nil {
            b.Fatal(err)
        }
    }
}

func BenchmarkCodegen_NestedArithmetic(b *testing.B) {
    src := "(+ (* 10 5) (- 20 (/ 100 5)))"
    prog := parseProgramBench(src)
    ast := &benchAST{prog: prog}
    gen := NewGoCodeGenerator(Config{PackageName: "main", GenerateComments: false, IndentSize: 4})
    syms := &stubSymbols{}
    b.ReportAllocs()
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        _, err := gen.Generate(ast, syms)
        if err != nil {
            b.Fatal(err)
        }
    }
}

func BenchmarkCodegen_ImportsAndAliases(b *testing.B) {
    src := "(import [[\"net/http\" http] [\"encoding/json\" json]])\n(http/Get \"http://example.com\")\n(json/Marshal \"x\")"
    prog := parseProgramBench(src)
    ast := &benchAST{prog: prog}
    gen := NewGoCodeGenerator(Config{PackageName: "main", GenerateComments: false, IndentSize: 4})
    syms := &stubSymbols{}
    b.ReportAllocs()
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        _, err := gen.Generate(ast, syms)
        if err != nil {
            b.Fatal(err)
        }
    }
}


