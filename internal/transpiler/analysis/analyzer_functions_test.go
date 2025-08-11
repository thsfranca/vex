package analysis

import (
	"testing"

	"github.com/antlr4-go/antlr/v4"
	"github.com/thsfranca/vex/internal/transpiler/parser"
)

func TestParseSingleExpression_Variants(t *testing.T) {
    a := NewAnalyzer()
    // List
    if n := a.parseSingleExpression("(+ 1 2)"); n == nil { t.Fatalf("list parse returned nil") }
    // Program fallback
    if n := a.parseSingleExpression("(+ 1 2) (+ 3 4)"); n == nil { t.Fatalf("program parse returned nil") }
}

func TestAnalyzeFunctionCall_Branches(t *testing.T) {
    a := NewAnalyzer()
    // '+' variadic with float presence
    in := antlr.NewInputStream("(+ 1 2.0 3)")
    l := parser.NewVexLexer(in)
    ts := antlr.NewCommonTokenStream(l, 0)
    p := parser.NewVexParser(ts)
    if v, err := a.VisitList(p.List().(*parser.ListContext)); err != nil || v == nil { t.Fatalf("+ path failed: %v %#v", err, v) }
    // unary '-' path
    in2 := antlr.NewInputStream("(- 1)")
    l2 := parser.NewVexLexer(in2)
    ts2 := antlr.NewCommonTokenStream(l2, 0)
    p2 := parser.NewVexParser(ts2)
    if v, err := a.VisitList(p2.List().(*parser.ListContext)); err != nil || v == nil { t.Fatalf("- path failed: %v %#v", err, v) }

    // namespaced local package call with not exported symbol
    a.SetPackageEnv(map[string]bool{"local/pkg": true}, map[string]map[string]bool{"local/pkg": {"Exported": true}}, map[string]map[string]*TypeScheme{})
    in3 := antlr.NewInputStream("(local/pkg/Hidden 1)")
    l3 := parser.NewVexLexer(in3)
    ts3 := antlr.NewCommonTokenStream(l3, 0)
    p3 := parser.NewVexParser(ts3)
    _ , _ = a.VisitList(p3.List().(*parser.ListContext))

    // unknown function error
    in4 := antlr.NewInputStream("(unknown 1)")
    l4 := parser.NewVexLexer(in4)
    ts4 := antlr.NewCommonTokenStream(l4, 0)
    p4 := parser.NewVexParser(ts4)
    _, _ = a.VisitList(p4.List().(*parser.ListContext))
}

func TestTypeHelpers_IntFloat(t *testing.T) {
    if !isInt("42") || isInt("42.0") { t.Fatalf("isInt checks failed") }
    if !isFloat("1.2") || isFloat("1.") || isFloat(".2") { t.Fatalf("isFloat checks failed") }
}

func TestVisitNode_DefaultUnknown(t *testing.T) {
    a := NewAnalyzer()
    type unknownTree struct{ antlr.Tree }
    if v, err := a.visitNode(&unknownTree{}); err == nil || v == nil || v.Type() != "undefined" { t.Fatalf("expected unknown handling") }
}

func TestTypeFromValue_Branches(t *testing.T) {
    a := NewAnalyzer()
    cases := []Value{
        NewBasicValue("s", "string"),
        NewBasicValue("t", "bool"),
        NewBasicValue("n", "number"),
        NewBasicValue("r", "record"),
        NewBasicValue("f", "func"),
        NewBasicValue("arr", "[]interface{}"),
        NewBasicValue("m", "map[interface{}]interface{}"),
        NewBasicValue("x", "symbol"),
    }
    for _, v := range cases {
        _ = a.typeFromValue(v)
    }
}


