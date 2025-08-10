package codegen

import (
	"strings"
	"testing"

	"github.com/antlr4-go/antlr/v4"
	"github.com/thsfranca/vex/internal/transpiler/parser"
)

func parseProg(t *testing.T, src string) *parser.ProgramContext {
    t.Helper()
    in := antlr.NewInputStream(src)
    lex := parser.NewVexLexer(in)
    toks := antlr.NewCommonTokenStream(lex, 0)
    p := parser.NewVexParser(toks)
    return p.Program().(*parser.ProgramContext)
}

func TestExports_AllowedSymbol(t *testing.T) {
    cfg := Config{PackageName: "main", Exports: map[string]map[string]bool{
        "a": {"add": true},
    }}
    g := NewGoCodeGenerator(cfg)
    prog := parseProg(t, "(import \"a\")\n(a/add 1 2)")
    if err := g.VisitProgram(prog); err != nil {
        t.Fatalf("VisitProgram error: %v", err)
    }
    out := g.buildFinalCode()
    if !strings.Contains(out, "import \"a\"") {
        t.Fatalf("expected import for a, got:\n%s", out)
    }
    if !strings.Contains(out, "a.add(1, 2)") {
        t.Fatalf("expected call to a.add, got:\n%s", out)
    }
}

func TestExports_NonExportedSymbolErrors(t *testing.T) {
    cfg := Config{PackageName: "main", Exports: map[string]map[string]bool{
        "a": {"add": true}, // 'hidden' not exported
    }}
    g := NewGoCodeGenerator(cfg)
    prog := parseProg(t, "(import \"a\")\n(a/hidden)")
    err := g.VisitProgram(prog)
    if err == nil {
        t.Fatalf("expected error for non-exported symbol, got nil")
    }
    if !strings.Contains(err.Error(), "not exported") {
        t.Fatalf("expected not exported error, got: %v", err)
    }
}


