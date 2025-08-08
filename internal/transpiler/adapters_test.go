package transpiler

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/thsfranca/vex/internal/transpiler/codegen"
)

func TestParserAdapter_ParseAndParseFile(t *testing.T) {
	pa := NewParserAdapter()
	// Parse string
	ast, err := pa.Parse("(+ 1 2)")
	if err != nil || ast == nil {
		t.Fatalf("Parse failed: %v", err)
	}
	// Parse file
	dir := t.TempDir()
	file := filepath.Join(dir, "t.vx")
	if err := os.WriteFile(file, []byte("(+ 1 2)"), 0644); err != nil {
		t.Fatalf("write temp file: %v", err)
	}
	ast2, err := pa.ParseFile(file)
	if err != nil || ast2 == nil {
		t.Fatalf("ParseFile failed: %v", err)
	}
}

func TestAnalyzerAdapter_Analyze_SetReporter(t *testing.T) {
	aa := NewAnalyzerAdapter()
	pa := NewParserAdapter()
	ast, err := pa.Parse("(+ 1 2)")
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	// minimal reporter
	aa.SetErrorReporter(&testReporter{})
	st, err := aa.Analyze(ast)
	if err != nil || st == nil {
		t.Fatalf("Analyze failed: %v", err)
	}
	// exercise SymbolTableAdapter methods
	_ = st.Define("x", &simpleValue{val: "1", typ: "int"})
	_, _ = st.Lookup("x")
	st.EnterScope()
	st.ExitScope()
}

type testReporter struct{}

func (t *testReporter) ReportError(line, column int, message string)   {}
func (t *testReporter) ReportWarning(line, column int, message string) {}
func (t *testReporter) HasErrors() bool                                { return false }
func (t *testReporter) GetErrors() []Error                             { return nil }

type simpleValue struct{ val, typ string }

func (s *simpleValue) String() string { return s.val }
func (s *simpleValue) Type() string   { return s.typ }

func TestCodeGeneratorAdapter_Generate_AddImport_SetPackage(t *testing.T) {
	pa := NewParserAdapter()
	ast, err := pa.Parse("(def x 1)")
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	cga := NewCodeGeneratorAdapter(codegen.Config{PackageName: "main"})
	cga.AddImport("fmt")
	cga.SetPackageName("main")
	st := &dummySymTable{}
	code, err := cga.Generate(ast, st)
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	if !strings.Contains(code, "package main") {
		t.Fatalf("expected package declaration, got:\n%s", code)
	}
}

type dummySymTable struct{}

func (d *dummySymTable) Define(name string, value Value) error { return nil }
func (d *dummySymTable) Lookup(name string) (Value, bool)      { return nil, false }
func (d *dummySymTable) EnterScope()                            {}
func (d *dummySymTable) ExitScope()                             {}
