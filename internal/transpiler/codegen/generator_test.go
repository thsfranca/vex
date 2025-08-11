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

// extra: TestGenerator_extractImportPair_ListAndArray moved from generator_extra_test.go
func TestGenerator_extractImportPair_ListAndArray(t *testing.T) {
    cfg := Config{PackageName: "main", IndentSize: 2}
    g := NewGoCodeGenerator(cfg)
    // List form: ("fmt" alias)
    list := parser.NewVexParser(antlr.NewCommonTokenStream(parser.NewVexLexer(antlr.NewInputStream("( \"fmt\" alias )")), 0)).List()
    p, a := g.extractImportPair(list)
    if p != "fmt" || a != "alias" { t.Fatalf("list pair mismatch: %q %q", p, a) }
    // Array form: ["os" a]
    arr := parser.NewVexParser(antlr.NewCommonTokenStream(parser.NewVexLexer(antlr.NewInputStream("[ \"os\" a ]")), 0)).Array()
    p2, a2 := g.extractImportPair(arr)
    if p2 != "os" || a2 != "a" { t.Fatalf("array pair mismatch: %q %q", p2, a2) }
}

// extra: TestGenerator_visitNode_DefaultAndNodeToString moved from generator_extra_test.go
func TestGenerator_visitNode_DefaultAndNodeToString(t *testing.T) {
    cfg := Config{PackageName: "main", IndentSize: 2}
    g := NewGoCodeGenerator(cfg)
    // Unknown node returns default value
    type unknownTree struct{ antlr.Tree }
    v, err := g.visitNode(&unknownTree{})
    if err != nil || v == nil || v.String() == "" { t.Fatalf("visitNode default failed: %v %#v", err, v) }
    // nodeToString on rule node should use GetText
    list := parser.NewVexParser(antlr.NewCommonTokenStream(parser.NewVexLexer(antlr.NewInputStream("(+ 1 2)")), 0)).List()
    if s := g.nodeToString(list); s == "" { t.Fatalf("nodeToString empty for rule node") }
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

// Additional coverage: missing-args paths and package variable/function export enforcement
func TestGeneratePrimitiveOp_MissingArgs(t *testing.T) {
    g := NewGoCodeGenerator(Config{PackageName: "main"})
    // len with no args
    v, _ := g.generatePrimitiveOp("len", []string{})
    if v.String() != "0" {
        t.Fatalf("expected default 0 for len with no args, got %s", v.String())
    }
    // append with no args
    v, _ = g.generatePrimitiveOp("append", []string{})
    if v.String() != "[]interface{}{}" {
        t.Fatalf("expected empty slice for append with no args, got %s", v.String())
    }
}

func TestGenerateFunctionCall_PackageVariable(t *testing.T) {
    g := NewGoCodeGenerator(Config{PackageName: "main"})
    v, err := g.generateFunctionCall("os.Args", []string{})
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    if v.String() != "os.Args" {
        t.Fatalf("expected package variable access, got %s", v.String())
    }
}

func TestGenerateFunctionCall_LocalPackageExportEnforcement(t *testing.T) {
    cfg := Config{PackageName: "main", IgnoreImports: map[string]bool{"local/pkg": true}, Exports: map[string]map[string]bool{"local/pkg": {"public": true}}}
    g := NewGoCodeGenerator(cfg)
    if _, err := g.generateFunctionCall("local/pkg/hidden", nil); err == nil {
        t.Fatalf("expected error for non-exported symbol")
    }
}

func TestGenerator_ImportWithAlias_And_IfOptimization(t *testing.T) {
    // Import alias via pair and optimize if condition when len([]) == 0
    prog := parseProgram(t, "(import [[\"fmt\" f]])\n(if (= (len []) 0) 1 2)")
    g := NewGoCodeGenerator(Config{PackageName: "main"})
    if err := g.VisitProgram(prog); err != nil {
        t.Fatalf("VisitProgram error: %v", err)
    }
    out := g.buildFinalCode()
    if !strings.Contains(out, "import f \"fmt\"") {
        t.Fatalf("expected alias import, got:\n%s", out)
    }
    if !strings.Contains(out, "if true") {
        t.Fatalf("expected if condition optimization to true, got:\n%s", out)
    }
}

func TestGenerateDo_TransformsDefAndReturnsLast(t *testing.T) {
    g := NewGoCodeGenerator(Config{PackageName: "main"})
    v, err := g.generateDo([]string{"x := 10", "(+ x 1)"})
    if err != nil {
        t.Fatalf("generateDo error: %v", err)
    }
    s := v.String()
    if !strings.Contains(s, "def(x, 10)") {
        t.Fatalf("expected def transform inside do, got: %s", s)
    }
    if !strings.Contains(s, "return (+ x 1)") {
        t.Fatalf("expected last expression as return, got: %s", s)
    }
}

func TestGenerateArithmetic_InsufficientArgs(t *testing.T) {
    g := NewGoCodeGenerator(Config{PackageName: "main"})
    v, err := g.generateArithmetic("+", []string{"1"})
    if err != nil {
        t.Fatalf("generateArithmetic error: %v", err)
    }
    if v.String() != "0" {
        t.Fatalf("expected default 0 for insufficient args, got %s", v.String())
    }
}

func TestGenerateFunctionCall_FallbackDotsConversion(t *testing.T) {
    g := NewGoCodeGenerator(Config{PackageName: "main"})
    v, err := g.generateFunctionCall("pkg/sub/Fn", []string{"1"})
    if err != nil {
        t.Fatalf("generateFunctionCall error: %v", err)
    }
    if v.String() != "pkg.sub.Fn(1)" {
        t.Fatalf("expected fallback dot conversion, got %s", v.String())
    }
}

func TestGenerateImport_StringArgs(t *testing.T) {
	g := NewGoCodeGenerator(Config{PackageName: "main"})
	
	// Test import with quoted string argument
	result, err := g.generateImport([]string{"\"fmt\""})
	if err != nil {
		t.Fatalf("generateImport failed: %v", err)
	}
	
	if result.String() == "" {
		t.Fatalf("generateImport should return non-empty result")
	}
	
	// Test import with insufficient args
	_, err = g.generateImport([]string{})
	if err == nil {
		t.Fatalf("generateImport should error with no args")
	}
}
