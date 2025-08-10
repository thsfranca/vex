package transpiler

import (
	"testing"

	a "github.com/thsfranca/vex/internal/transpiler/analysis"
	cg "github.com/thsfranca/vex/internal/transpiler/codegen"
)

// Stubs for analysis and codegen values
type stubAnalysisValue struct{ s, t string }
func (v *stubAnalysisValue) String() string { return v.s }
func (v *stubAnalysisValue) Type() string   { return v.t }

type stubCodegenValue struct{ s, t string }
func (v *stubCodegenValue) String() string { return v.s }
func (v *stubCodegenValue) Type() string   { return v.t }

// SymbolTable stub for wrapping by CodegenSymbolTable
type stubSymTable struct{}
func (s *stubSymTable) Define(name string, value Value) error { return nil }
func (s *stubSymTable) Lookup(name string) (Value, bool)      { return nil, false }
func (s *stubSymTable) EnterScope()                            {}
func (s *stubSymTable) ExitScope()                             {}

func TestValueAdapters_StringType(t *testing.T) {
    // analysis.Value -> ValueAdapter
    av := &ValueAdapter{value: &stubAnalysisValue{s: "x", t: "T"}}
    if av.String() != "x" || av.Type() != "T" {
        t.Fatalf("unexpected ValueAdapter outputs: %s %s", av.String(), av.Type())
    }
    // codegen.Value -> CodegenValueAdapter
    cga := &CodegenValueAdapter{value: &stubCodegenValue{s: "y", t: "U"}}
    if cga.String() != "y" || cga.Type() != "U" {
        t.Fatalf("unexpected CodegenValueAdapter outputs: %s %s", cga.String(), cga.Type())
    }
    // transpiler.Value -> CodegenValueBridge
    cvb := &CodegenValueBridge{value: &simpleValue{val: "z", typ: "V"}}
    if cvb.String() != "z" || cvb.Type() != "V" {
        t.Fatalf("unexpected CodegenValueBridge outputs: %s %s", cvb.String(), cvb.Type())
    }
}

func TestCodegenSymbolTable_Wrappers(t *testing.T) {
    st := &stubSymTable{}
    cst := &CodegenSymbolTable{table: st}
    if err := cst.Define("a", &stubCodegenValue{s: "1", t: "int"}); err != nil {
        t.Fatalf("Define returned error: %v", err)
    }
    _, _ = cst.Lookup("a")
    cst.EnterScope()
    cst.ExitScope()
}

func TestErrorReporterAdapter_Methods(t *testing.T) {
    era := &ErrorReporterAdapter{reporter: &testReporter{}}
    era.ReportError(1, 2, "msg")
    era.ReportWarning(3, 4, "warn")
    _ = era.HasErrors()
    _ = era.GetErrors()
}

// Ensure the adapter types compile against the analysis/codegen interfaces
var _ a.Value = (*stubAnalysisValue)(nil)
var _ cg.Value = (*stubCodegenValue)(nil)


