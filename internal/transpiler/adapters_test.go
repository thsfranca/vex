package transpiler

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/thsfranca/vex/internal/transpiler/analysis"
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
func (d *dummySymTable) EnterScope()                           {}
func (d *dummySymTable) ExitScope()                            {}

// extra: consolidated from adapters_values_extra_test.go
func TestAnalysisValueAdapter_StringAndType(t *testing.T) {
	ava := &AnalysisValueAdapter{value: &simpleValue{val: "v", typ: "K"}}
	if ava.String() != "v" || ava.Type() != "K" {
		t.Fatalf("unexpected outputs: %s %s", ava.String(), ava.Type())
	}
}

// Cover CodegenSymbolTable and CodegenValueBridge low-coverage paths
func TestCodegenSymbolTable_DefineLookupScopeAndBridge(t *testing.T) {
	// Backing symbol table to observe calls
	bt := &backingSymTable{m: map[string]Value{}}
	cst := &CodegenSymbolTable{table: bt}

	// Define via codegen interface (wraps value)
	if err := cst.Define("k", &stubCodegenValue{s: "1", t: "int"}); err != nil {
		t.Fatalf("Define error: %v", err)
	}
	// Enter/Exit scope delegate calls
	cst.EnterScope()
	cst.ExitScope()
	// Lookup should return a CodegenValueBridge
	cv, ok := cst.Lookup("k")
	if !ok || cv == nil {
		t.Fatalf("expected lookup to succeed")
	}
	// Bridge String/Type should reflect backing value
	if cv.String() != "1" || cv.Type() != "int" {
		t.Fatalf("bridge mismatch: %s %s", cv.String(), cv.Type())
	}
}

type backingSymTable struct{ m map[string]Value }

func (b *backingSymTable) Define(name string, value Value) error { b.m[name] = value; return nil }
func (b *backingSymTable) Lookup(name string) (Value, bool)      { v, ok := b.m[name]; return v, ok }
func (b *backingSymTable) EnterScope()                           {}
func (b *backingSymTable) ExitScope()                            {}

// Stubs for codegen values
type stubCodegenValue struct{ s, t string }

func (v *stubCodegenValue) String() string { return v.s }
func (v *stubCodegenValue) Type() string   { return v.t }

// Also cover the SymbolTableAdapter wrappers with a minimal backing table
func TestSymbolTableAdapter_Wrappers(t *testing.T) {
	bt := &backingAnalysisTable{m: map[string]analysis.Value{}}
	sta := &SymbolTableAdapter{table: bt}
	if err := sta.Define("z", &simpleValue{val: "2", typ: "int"}); err != nil {
		t.Fatalf("Define: %v", err)
	}
	sta.EnterScope()
	sta.ExitScope()
	v, ok := sta.Lookup("z")
	if !ok || v == nil || v.String() != "2" || v.Type() != "int" {
		t.Fatalf("Lookup mismatch: ok=%v v=%#v", ok, v)
	}
}

type backingAnalysisTable struct{ m map[string]analysis.Value }

func (b *backingAnalysisTable) Define(name string, value analysis.Value) error {
	b.m[name] = value
	return nil
}
func (b *backingAnalysisTable) Lookup(name string) (analysis.Value, bool) {
	v, ok := b.m[name]
	return v, ok
}
func (b *backingAnalysisTable) EnterScope() {}
func (b *backingAnalysisTable) ExitScope()  {}
