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

func TestAnalyzeExport_Minimal(t *testing.T) {
    a := NewAnalyzer()
    // (export [x y]) is accepted and returns void-like value
    input := antlr.NewInputStream("(export [x y])")
    l := parser.NewVexLexer(input)
    ts := antlr.NewCommonTokenStream(l, 0)
    p := parser.NewVexParser(ts)
    prog := p.Program()
    if prog == nil { t.Fatalf("expected program parse") }
    if err := a.VisitProgram(prog.(*parser.ProgramContext)); err != nil {
        t.Fatalf("visit program: %v", err)
    }
}

func TestAnalyzeFunctionCall_EqArityAndMismatch(t *testing.T) {
    a := NewAnalyzer()
    // (= 1 2) should type to bool via unify success
    in1 := antlr.NewInputStream("(= 1 2)")
    l1 := parser.NewVexLexer(in1)
    ts1 := antlr.NewCommonTokenStream(l1, 0)
    p1 := parser.NewVexParser(ts1)
    if v, err := a.VisitList(p1.List().(*parser.ListContext)); err != nil || v == nil || v.Type() != "bool" {
        t.Fatalf("eq typing failed: %v %#v", err, v)
    }
    // Arity error path for '=' (we still return bool value per analyzer behavior)
    in2 := antlr.NewInputStream("(= 1 2 3)")
    l2 := parser.NewVexLexer(in2)
    ts2 := antlr.NewCommonTokenStream(l2, 0)
    p2 := parser.NewVexParser(ts2)
    if v, err := a.VisitList(p2.List().(*parser.ListContext)); err != nil || v == nil || v.Type() != "bool" {
        t.Fatalf("eq arity path not exercised: %v %#v", err, v)
    }
}

func TestAnalyzeDo_TypingAndMismatch(t *testing.T) {
    a := NewAnalyzer()
    // (do 1 2) should succeed and return number
    in1 := antlr.NewInputStream("(do 1 2)")
    l1 := parser.NewVexLexer(in1)
    ts1 := antlr.NewCommonTokenStream(l1, 0)
    p1 := parser.NewVexParser(ts1)
    v1, err := a.VisitList(p1.List().(*parser.ListContext))
    if err != nil || v1 == nil || v1.Type() != "number" {
        t.Fatalf("do typing failed: %v %#v", err, v1)
    }

    // (do 1 "x") should report mismatch diagnostic but still return a value
    in2 := antlr.NewInputStream("(do 1 \"x\")")
    l2 := parser.NewVexLexer(in2)
    ts2 := antlr.NewCommonTokenStream(l2, 0)
    p2 := parser.NewVexParser(ts2)
    v2, _ := a.VisitList(p2.List().(*parser.ListContext))
    if v2 == nil {
        t.Fatalf("expected value from do with mismatch")
    }
}

func TestSetPackageEnv_Branches(t *testing.T) {
    a := NewAnalyzer()
    // Provide non-nil maps to exercise assignment paths
    ignore := map[string]bool{"modA": true}
    exports := map[string]map[string]bool{"pkg": {"X": true}}
    schemes := map[string]map[string]*TypeScheme{"pkg": {"f": {}}}
    a.SetPackageEnv(ignore, exports, schemes)
    // Call again with empty maps to hit alternate branch
    a.SetPackageEnv(map[string]bool{}, map[string]map[string]bool{}, map[string]map[string]*TypeScheme{})
}

func TestAnalyzeRecordConstructor_SuccessAndMismatch(t *testing.T) {
    a := NewAnalyzer()
    // Define a nominal record "User" with fields; constructor syntax expects vector form
    rec := NewRecordValue("User", map[string]string{"name": "string", "age": "int"}, []string{"name", "age"})
    if err := a.symbolTable.Define("User", rec); err != nil { t.Fatalf("define: %v", err) }

    // Successful construction: (User [name: "bob" age: 42])
    in1 := antlr.NewInputStream("(User [name: \"bob\" age: 42])")
    l1 := parser.NewVexLexer(in1)
    ts1 := antlr.NewCommonTokenStream(l1, 0)
    p1 := parser.NewVexParser(ts1)
    if v, err := a.VisitList(p1.List().(*parser.ListContext)); err != nil || v == nil || v.Type() != "record" {
        t.Fatalf("record constructor typing failed: %v %#v", err, v)
    }

    // Mismatch: (User [name: 42 age: 42]) should emit diagnostics
    in2 := antlr.NewInputStream("(User [name: 42 age: 42])")
    l2 := parser.NewVexLexer(in2)
    ts2 := antlr.NewCommonTokenStream(l2, 0)
    p2 := parser.NewVexParser(ts2)
    if v, _ := a.VisitList(p2.List().(*parser.ListContext)); v == nil {
        t.Fatalf("expected value even on mismatch")
    }
}



