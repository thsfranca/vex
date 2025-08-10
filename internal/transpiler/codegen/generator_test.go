package codegen

import (
	"strings"
	"testing"

	"github.com/antlr4-go/antlr/v4"
	"github.com/thsfranca/vex/internal/transpiler/parser"
)

func parseProgram(t *testing.T, src string) *parser.ProgramContext {
	t.Helper()
	input := antlr.NewInputStream(src)
	lex := parser.NewVexLexer(input)
	tokens := antlr.NewCommonTokenStream(lex, 0)
	p := parser.NewVexParser(tokens)
	prog := p.Program()
	if prog == nil {
		t.Fatalf("failed to parse program")
	}
	return prog.(*parser.ProgramContext)
}

func TestGenerator_DefAndImplicitReturn(t *testing.T) {
	prog := parseProgram(t, "(def x 42)")
	g := NewGoCodeGenerator(Config{PackageName: "main"})
	if err := g.VisitProgram(prog); err != nil {
		t.Fatalf("VisitProgram error: %v", err)
	}
	out := g.buildFinalCode()
	if !strings.Contains(out, "x := 42") {
		t.Fatalf("expected assignment in output, got:\n%s", out)
	}
	if !strings.Contains(out, "_ = x // Return last defined value") {
		t.Fatalf("expected implicit return for single def, got:\n%s", out)
	}
}

func TestGenerator_ImportAndPrintln(t *testing.T) {
	prog := parseProgram(t, "(import \"fmt\")\n(fmt/Println \"hi\")")
	g := NewGoCodeGenerator(Config{PackageName: "main"})
	if err := g.VisitProgram(prog); err != nil {
		t.Fatalf("VisitProgram error: %v", err)
	}
	out := g.buildFinalCode()
	if !strings.Contains(out, "import \"fmt\"") {
		t.Fatalf("expected fmt import, got:\n%s", out)
	}
	if !strings.Contains(out, "fmt.Println(\"hi\")") {
		t.Fatalf("expected println call, got:\n%s", out)
	}
}

func TestGenerator_ArithmeticAndComparison(t *testing.T) {
    prog := parseProgram(t, "(+ 1 2)\n(= 1 1)\n(> 2 1)\n(< 1 2)")
	g := NewGoCodeGenerator(Config{PackageName: "main"})
	if err := g.VisitProgram(prog); err != nil {
		t.Fatalf("VisitProgram error: %v", err)
	}
	out := g.buildFinalCode()
	if !strings.Contains(out, "_ = (1 + 2)") {
		t.Fatalf("expected arithmetic in output, got:\n%s", out)
	}
	if !strings.Contains(out, "_ = (1 == 1)") {
		t.Fatalf("expected equality conversion to ==, got:\n%s", out)
	}
    if !strings.Contains(out, "_ = (2 > 1)") {
        t.Fatalf("expected greater-than comparison, got:\n%s", out)
    }
    if !strings.Contains(out, "_ = (1 < 2)") {
        t.Fatalf("expected less-than comparison, got:\n%s", out)
    }
}

func TestGenerator_ArraysAndDoAndFn(t *testing.T) {
	// Use list forms at top-level so the generator visits them
	src := "(def arr [1 2 3])\n(do (+ 1 2) (* 3 4))\n(def f (fn [x y] (+ x y)))"
	prog := parseProgram(t, src)
	g := NewGoCodeGenerator(Config{PackageName: "main"})
	if err := g.VisitProgram(prog); err != nil {
		t.Fatalf("VisitProgram error: %v", err)
	}
	out := g.buildFinalCode()
	if !strings.Contains(out, "arr := []interface{}{1, 2, 3}") {
		t.Fatalf("expected array literal conversion, got:\n%s", out)
	}
	if !strings.Contains(out, "_ = arr") {
		t.Fatalf("expected use of defined variable to satisfy compiler, got:\n%s", out)
	}
	if !strings.Contains(out, "_ = func() interface{} { (1 + 2); return (3 * 4) }()") {
		t.Fatalf("expected do block lowering, got:\n%s", out)
	}
	if !strings.Contains(out, "f := func(x interface{}, y interface{}) interface{} { return (x + y) }") {
		t.Fatalf("expected fn lowering inside def, got:\n%s", out)
	}
}

func TestGenerator_PrimitiveOps(t *testing.T) {
    prog := parseProgram(t, "(get [1 2 3] 1)\n(slice [1 2 3] 1)\n(len [1 2])\n(append [1] [2 3])\n(get [] 0)\n(slice [] 0)")
	g := NewGoCodeGenerator(Config{PackageName: "main"})
	if err := g.VisitProgram(prog); err != nil {
		t.Fatalf("VisitProgram error: %v", err)
	}
	out := g.buildFinalCode()
	if !strings.Contains(out, "_ = func() interface{} { if len([]interface{}{1, 2, 3}) > 1 { return []interface{}{1, 2, 3}[1] } else { return nil } }()") {
		t.Fatalf("expected get lowering with bounds check, got:\n%s", out)
	}
	if !strings.Contains(out, "_ = func() []interface{} { if len([]interface{}{1, 2, 3}) > 1 { return []interface{}{1, 2, 3}[1:] } else { return []interface{}{} } }()") {
		t.Fatalf("expected slice lowering, got:\n%s", out)
	}
	if !strings.Contains(out, "_ = len([]interface{}{1, 2})") {
		t.Fatalf("expected len lowering, got:\n%s", out)
	}
	if !strings.Contains(out, "_ = append([]interface{}{1}, []interface{}{2, 3}...)") {
		t.Fatalf("expected append lowering, got:\n%s", out)
	}
    if !strings.Contains(out, "_ = func() interface{} { if len([]interface{}{}) > 0 {") {
        t.Fatalf("expected get empty-array branch to be covered, got:\n%s", out)
    }
    if !strings.Contains(out, "_ = func() []interface{} { if len([]interface{}{}) > 0 {") {
        t.Fatalf("expected slice empty-array branch to be covered, got:\n%s", out)
    }
}

func TestGenerator_NestedFunctionNameAndPackageVar(t *testing.T) {
	// ((fn [x] x) 10) and os.Args access
	prog := parseProgram(t, "((fn [x] x) 10)\n(def args (os/Args))")
	g := NewGoCodeGenerator(Config{PackageName: "main"})
	if err := g.VisitProgram(prog); err != nil {
		t.Fatalf("VisitProgram error: %v", err)
	}
	out := g.buildFinalCode()
	if !strings.Contains(out, "_ = func(x interface{}) interface{} { return x }(10)") {
		t.Fatalf("expected nested fn call lowering, got:\n%s", out)
	}
	if !strings.Contains(out, "args := os.Args") {
		t.Fatalf("expected package variable access, got:\n%s", out)
	}
}

func TestGenerator_SymbolHyphenToUnderscore(t *testing.T) {
	prog := parseProgram(t, "(def program-name 1)")
	g := NewGoCodeGenerator(Config{PackageName: "main"})
	if err := g.VisitProgram(prog); err != nil {
		t.Fatalf("VisitProgram error: %v", err)
	}
	out := g.buildFinalCode()
	if !strings.Contains(out, "program_name := 1") {
		t.Fatalf("expected symbol hyphen conversion, got:\n%s", out)
	}
}
