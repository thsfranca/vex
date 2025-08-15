package analysis

import (
	"testing"

	"github.com/antlr4-go/antlr/v4"
	"github.com/thsfranca/vex/internal/transpiler/parser"
)

func TestAnalyzeExport_Error_NoArgs(t *testing.T) {
	a := NewAnalyzer()
	in := antlr.NewInputStream("(export)")
	l := parser.NewVexLexer(in)
	ts := antlr.NewCommonTokenStream(l, 0)
	p := parser.NewVexParser(ts)
	list := p.List().(*parser.ListContext)
	if _, err := a.VisitList(list); err == nil {
		t.Fatalf("expected error for export with no args")
	}
}

func TestAnalyzeRecordFieldCall_UnknownField(t *testing.T) {
	a := NewAnalyzer()
	rec := NewRecordValue("User", map[string]string{"name": "string"}, []string{"name"})
	if err := a.symbolTable.Define("User", rec); err != nil {
		t.Fatalf("define: %v", err)
	}
	in := antlr.NewInputStream("(User :age)")
	l := parser.NewVexLexer(in)
	ts := antlr.NewCommonTokenStream(l, 0)
	p := parser.NewVexParser(ts)
	if v, _ := a.VisitList(p.List().(*parser.ListContext)); v == nil {
		t.Fatalf("expected value even on unknown field")
	}
	if a.GetErrorReporter().GetErrorCount() == 0 {
		t.Fatalf("expected error reported for unknown field")
	}
}
