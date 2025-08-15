package transpiler

import (
	"testing"

	"github.com/antlr4-go/antlr/v4"
	a "github.com/thsfranca/vex/internal/transpiler/analysis"
	cg "github.com/thsfranca/vex/internal/transpiler/codegen"
	"github.com/thsfranca/vex/internal/transpiler/macro"
	"github.com/thsfranca/vex/internal/transpiler/parser"
)

// Stub analysis visitor returning fixed values
type stubAVisitor struct{}

func (s *stubAVisitor) VisitProgram(ctx *parser.ProgramContext) error { return nil }
func (s *stubAVisitor) VisitList(ctx *parser.ListContext) (a.Value, error) {
	return a.NewBasicValue("L", "list"), nil
}
func (s *stubAVisitor) VisitArray(ctx *parser.ArrayContext) (a.Value, error) {
	return a.NewBasicValue("A", "array"), nil
}
func (s *stubAVisitor) VisitTerminal(node antlr.TerminalNode) (a.Value, error) {
	return a.NewBasicValue(node.GetText(), "symbol"), nil
}

// Stub codegen visitor returning fixed values
type stubCGVisitor struct{}

func (s *stubCGVisitor) VisitProgram(ctx *parser.ProgramContext) error { return nil }
func (s *stubCGVisitor) VisitList(ctx *parser.ListContext) (cg.Value, error) {
	return a.NewBasicValue("L", "list"), nil
}
func (s *stubCGVisitor) VisitArray(ctx *parser.ArrayContext) (cg.Value, error) {
	return a.NewBasicValue("A", "array"), nil
}
func (s *stubCGVisitor) VisitTerminal(node antlr.TerminalNode) (cg.Value, error) {
	return a.NewBasicValue(node.GetText(), "symbol"), nil
}

// Ensure adapters for macro visitor bridging are exercised too
type stubMacroVisitor struct{}

func (s *stubMacroVisitor) VisitProgram(ctx *parser.ProgramContext) error { return nil }
func (s *stubMacroVisitor) VisitList(ctx *parser.ListContext) (macro.Value, error) {
	return a.NewBasicValue("ML", "list"), nil
}
func (s *stubMacroVisitor) VisitArray(ctx *parser.ArrayContext) (macro.Value, error) {
	return a.NewBasicValue("MA", "array"), nil
}
func (s *stubMacroVisitor) VisitTerminal(node antlr.TerminalNode) (macro.Value, error) {
	return a.NewBasicValue(node.GetText(), "symbol"), nil
}

func TestAnalysisVisitorBridge_Methods(t *testing.T) {
	input := antlr.NewInputStream("(+ 1 2)")
	l := parser.NewVexLexer(input)
	ts := antlr.NewCommonTokenStream(l, 0)
	p := parser.NewVexParser(ts)
	list := p.List()
	if list == nil {
		t.Fatalf("expected list parse")
	}

	br := &AnalysisVisitorBridge{visitor: &stubAVisitor{}}
	if v, err := br.VisitList(list.(*parser.ListContext)); err != nil || v == nil || v.String() != "L" {
		t.Fatalf("VisitList bridge failed: %v, %v", err, v)
	}
	if v, _ := br.VisitList(list.(*parser.ListContext)); v == nil || v.Type() != "list" {
		t.Fatalf("VisitList Type() mismatch: %v", v)
	}
	// Parse an array using a fresh parser
	arrInput := antlr.NewInputStream("[1 2]")
	arrLex := parser.NewVexLexer(arrInput)
	arrTs := antlr.NewCommonTokenStream(arrLex, 0)
	arrParser := parser.NewVexParser(arrTs)
	arr := arrParser.Array()
	if arr == nil {
		t.Fatalf("expected array parse")
	}
	if v, err := br.VisitArray(arr.(*parser.ArrayContext)); err != nil || v == nil || v.String() != "A" {
		t.Fatalf("VisitArray bridge failed: %v, %v", err, v)
	}
	if v, _ := br.VisitArray(arr.(*parser.ArrayContext)); v == nil || v.Type() != "array" {
		t.Fatalf("VisitArray Type() mismatch: %v", v)
	}
	// Create a terminal node from a simple input
	termInput := antlr.NewInputStream("x")
	termLex := parser.NewVexLexer(termInput)
	termTs := antlr.NewCommonTokenStream(termLex, 0)
	termTok := termTs.Get(0)
	term := antlr.NewTerminalNodeImpl(termTok)
	if v, err := br.VisitTerminal(term); err != nil || v == nil || v.String() != "x" {
		t.Fatalf("VisitTerminal bridge failed: %v, %v", err, v)
	}
	if v, _ := br.VisitTerminal(term); v == nil || v.Type() != "symbol" {
		t.Fatalf("VisitTerminal Type() mismatch: %v", v)
	}
}

func TestCodegenVisitorBridge_Methods(t *testing.T) {
	input := antlr.NewInputStream("(+ 1 2)")
	l := parser.NewVexLexer(input)
	ts := antlr.NewCommonTokenStream(l, 0)
	p := parser.NewVexParser(ts)
	list := p.List()
	if list == nil {
		t.Fatalf("expected list parse")
	}
	br := &CodegenVisitorBridge{visitor: &stubCGVisitor{}}
	if v, err := br.VisitList(list.(*parser.ListContext)); err != nil || v == nil || v.String() != "L" {
		t.Fatalf("VisitList bridge failed: %v, %v", err, v)
	}
	if v, _ := br.VisitList(list.(*parser.ListContext)); v == nil || v.Type() != "list" {
		t.Fatalf("VisitList Type() mismatch: %v", v)
	}
	arrInput := antlr.NewInputStream("[1 2]")
	arrLex := parser.NewVexLexer(arrInput)
	arrTs := antlr.NewCommonTokenStream(arrLex, 0)
	arrParser := parser.NewVexParser(arrTs)
	arr := arrParser.Array()
	if arr == nil {
		t.Fatalf("expected array parse")
	}
	if v, err := br.VisitArray(arr.(*parser.ArrayContext)); err != nil || v == nil || v.String() != "A" {
		t.Fatalf("VisitArray bridge failed: %v, %v", err, v)
	}
	if v, _ := br.VisitArray(arr.(*parser.ArrayContext)); v == nil || v.Type() != "array" {
		t.Fatalf("VisitArray Type() mismatch: %v", v)
	}
	termInput := antlr.NewInputStream("y")
	termLex := parser.NewVexLexer(termInput)
	termTs := antlr.NewCommonTokenStream(termLex, 0)
	termTok := termTs.Get(0)
	term := antlr.NewTerminalNodeImpl(termTok)
	if v, err := br.VisitTerminal(term); err != nil || v == nil || v.String() != "y" {
		t.Fatalf("VisitTerminal bridge failed: %v, %v", err, v)
	}
	if v, _ := br.VisitTerminal(term); v == nil || v.Type() != "symbol" {
		t.Fatalf("VisitTerminal Type() mismatch: %v", v)
	}
}

func TestMacroVisitorBridge_Methods(t *testing.T) {
	input := antlr.NewInputStream("(do [1 2] 3)")
	l := parser.NewVexLexer(input)
	ts := antlr.NewCommonTokenStream(l, 0)
	p := parser.NewVexParser(ts)
	prog := p.Program()
	if prog == nil {
		t.Fatalf("expected program parse")
	}

	// Adapt ConcreteAST to macro visitor bridge
	ast := NewConcreteAST(prog)
	br := &MacroVisitorBridge{visitor: &stubMacroVisitor{}}
	if err := ast.Accept(br); err != nil {
		t.Fatalf("macro bridge VisitProgram failed: %v", err)
	}
	// Visit children to touch list/array/terminal paths
	if list, ok := prog.GetChild(0).(*parser.ListContext); ok {
		if v, err := br.VisitList(list); err != nil || v == nil || v.String() != "ML" {
			t.Fatalf("VisitList: %v %v", err, v)
		}
		if v, _ := br.VisitList(list); v == nil || v.Type() != "list" {
			t.Fatalf("VisitList Type(): %v", v)
		}
	}
	arrInput := antlr.NewInputStream("[1 2]")
	arrLex := parser.NewVexLexer(arrInput)
	arrTs := antlr.NewCommonTokenStream(arrLex, 0)
	arrParser := parser.NewVexParser(arrTs)
	if arr := arrParser.Array(); arr != nil {
		if v, err := br.VisitArray(arr.(*parser.ArrayContext)); err != nil || v == nil || v.String() != "MA" {
			t.Fatalf("VisitArray: %v %v", err, v)
		}
		if v, _ := br.VisitArray(arr.(*parser.ArrayContext)); v == nil || v.Type() != "array" {
			t.Fatalf("VisitArray Type(): %v", v)
		}
	}
	tlex := parser.NewVexLexer(antlr.NewInputStream("z"))
	tts := antlr.NewCommonTokenStream(tlex, 0)
	ttok := tts.Get(0)
	term := antlr.NewTerminalNodeImpl(ttok)
	if v, err := br.VisitTerminal(term); err != nil || v == nil || v.String() != "z" {
		t.Fatalf("VisitTerminal: %v %v", err, v)
	}
	if v, _ := br.VisitTerminal(term); v == nil || v.Type() != "symbol" {
		t.Fatalf("VisitTerminal Type(): %v", v)
	}
}

// extra: consolidated from adapters_visitors_extra2_test.go
type stubMacroVisitor2 struct{}

func (s *stubMacroVisitor2) VisitProgram(ctx *parser.ProgramContext) error { return nil }
func (s *stubMacroVisitor2) VisitList(ctx *parser.ListContext) (macro.Value, error) {
	return a.NewBasicValue("ML", "list"), nil
}
func (s *stubMacroVisitor2) VisitArray(ctx *parser.ArrayContext) (macro.Value, error) {
	return a.NewBasicValue("MA", "array"), nil
}
func (s *stubMacroVisitor2) VisitTerminal(node antlr.TerminalNode) (macro.Value, error) {
	return a.NewBasicValue(node.GetText(), "symbol"), nil
}

type stubCGVisitor2 struct{}

func (s *stubCGVisitor2) VisitProgram(ctx *parser.ProgramContext) error { return nil }
func (s *stubCGVisitor2) VisitList(ctx *parser.ListContext) (cg.Value, error) {
	return a.NewBasicValue("L", "list"), nil
}
func (s *stubCGVisitor2) VisitArray(ctx *parser.ArrayContext) (cg.Value, error) {
	return a.NewBasicValue("A", "array"), nil
}
func (s *stubCGVisitor2) VisitTerminal(node antlr.TerminalNode) (cg.Value, error) {
	return a.NewBasicValue(node.GetText(), "symbol"), nil
}

func TestMacroAndCodegenValueAdapter_TypeAccessors(t *testing.T) {
	// Build list
	l := parser.NewVexLexer(antlr.NewInputStream("(+ 1 2)"))
	ts := antlr.NewCommonTokenStream(l, 0)
	p := parser.NewVexParser(ts)
	list := p.List()
	if list == nil {
		t.Fatalf("expected list parse")
	}

	// Macro visitor bridge should return MacroValueAdapter with proper Type()
	mbr := &MacroVisitorBridge{visitor: &stubMacroVisitor2{}}
	mv, err := mbr.VisitList(list.(*parser.ListContext))
	if err != nil || mv == nil || mv.Type() != "list" {
		t.Fatalf("macro Type mismatch: %v %#v", err, mv)
	}

	// Codegen visitor bridge should return CodegenValueAdapter with proper Type()
	cbr := &CodegenVisitorBridge{visitor: &stubCGVisitor2{}}
	cv, err := cbr.VisitArray(parser.NewVexParser(antlr.NewCommonTokenStream(parser.NewVexLexer(antlr.NewInputStream("[1 2]")), 0)).Array().(*parser.ArrayContext))
	if err != nil || cv == nil || cv.Type() != "array" {
		t.Fatalf("codegen Type mismatch: %v %#v", err, cv)
	}
}

func TestMacroASTAdapter_Accept(t *testing.T) {
	input := antlr.NewInputStream("(+ 1 2)")
	l := parser.NewVexLexer(input)
	ts := antlr.NewCommonTokenStream(l, 0)
	p := parser.NewVexParser(ts)
	prog := p.Program()
	if prog == nil {
		t.Fatalf("expected program parse")
	}

	ast := NewConcreteAST(prog)
	maa := &MacroASTAdapter{ast: ast}
	if err := maa.Accept(&stubMacroVisitor{}); err != nil {
		t.Fatalf("MacroASTAdapter.Accept failed: %v", err)
	}
}

func TestMacroExpanderAdapter_RegisterHasGet(t *testing.T) {
	reg := macro.NewRegistry(macro.Config{EnableValidation: true})
	mea := NewMacroExpanderAdapter(macro.NewMacroExpander(reg))
	m := &macro.Macro{Name: "m", Params: []string{"x"}, Body: "x"}
	if err := mea.RegisterMacro("m", m); err != nil {
		t.Fatalf("register: %v", err)
	}
	if !mea.HasMacro("m") {
		t.Fatalf("expected HasMacro true")
	}
	if got, ok := mea.GetMacro("m"); !ok || got == nil || got.Name != "m" {
		t.Fatalf("GetMacro failed: %#v %v", got, ok)
	}
}
