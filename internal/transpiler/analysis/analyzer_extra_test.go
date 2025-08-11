package analysis

import (
	"testing"

	"github.com/antlr4-go/antlr/v4"
	"github.com/thsfranca/vex/internal/transpiler/parser"
)

func TestAnalyzer_isUnknownType_And_RecordFieldCall(t *testing.T) {
    if !isUnknownType("interface{}") || !isUnknownType("undefined") || isUnknownType("string") {
        t.Fatalf("isUnknownType logic mismatch")
    }

    a := NewAnalyzer()
    // Seed a record symbol
    rec := NewRecordValue("User", map[string]string{"name": "string", "age": "int"}, []string{"name", "age"})
    if err := a.symbolTable.Define("User", rec); err != nil { t.Fatalf("define: %v", err) }

    // Parse a field call: (User :name)
    in := antlr.NewInputStream("(User :name)")
    lx := parser.NewVexLexer(in)
    ts := antlr.NewCommonTokenStream(lx, 0)
    p := parser.NewVexParser(ts)
    list := p.List().(*parser.ListContext)
    if list == nil { t.Fatalf("expected list parse") }

    v, err := a.VisitList(list)
    if err != nil { t.Fatalf("VisitList error: %v", err) }
    if v == nil || v.Type() == "undefined" {
        t.Fatalf("expected field value type, got %#v", v)
    }
}


