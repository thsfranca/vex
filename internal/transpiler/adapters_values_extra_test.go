package transpiler

import (
	"testing"

	"github.com/antlr4-go/antlr/v4"
	"github.com/thsfranca/vex/internal/transpiler/macro"
	"github.com/thsfranca/vex/internal/transpiler/parser"
)

type reporterWithErrors struct{}

func (r *reporterWithErrors) ReportError(line, column int, message string)   {}
func (r *reporterWithErrors) ReportWarning(line, column int, message string) {}
func (r *reporterWithErrors) HasErrors() bool                                { return true }
func (r *reporterWithErrors) GetErrors() []Error {
    return []Error{{Line: 1, Column: 2, Message: "m", Type: TypeError}}
}

// Symbol table that returns a value for Lookup to exercise CodegenValueBridge
type lookupSymTable struct{}

func (l *lookupSymTable) Define(name string, value Value) error { return nil }
func (l *lookupSymTable) Lookup(name string) (Value, bool)      { return &simpleValue{val: "ok", typ: "T"}, true }
func (l *lookupSymTable) EnterScope()                            {}
func (l *lookupSymTable) ExitScope()                             {}

// Mutable reporter used to assert delegation behavior
type mutableReporter struct{ errs []Error }
func (m *mutableReporter) ReportError(line, column int, message string) { m.errs = append(m.errs, Error{Line: line, Column: column, Message: message}) }
func (m *mutableReporter) ReportWarning(line, column int, message string) {}
func (m *mutableReporter) HasErrors() bool { return len(m.errs) > 0 }
func (m *mutableReporter) GetErrors() []Error { return m.errs }

func TestAnalysisValueAdapter_StringAndType(t *testing.T) {
    ava := &AnalysisValueAdapter{value: &simpleValue{val: "v", typ: "K"}}
    if ava.String() != "v" || ava.Type() != "K" {
        t.Fatalf("unexpected outputs: %s %s", ava.String(), ava.Type())
    }
}

func TestErrorReporterAdapter_GetErrors_Maps(t *testing.T) {
    era := &ErrorReporterAdapter{reporter: &reporterWithErrors{}}
    errs := era.GetErrors()
    if len(errs) != 1 || errs[0].Line != 1 || errs[0].Column != 2 || errs[0].Message != "m" {
        t.Fatalf("unexpected mapped errors: %#v", errs)
    }
}

func TestMacroExpanderAdapter_ExpandTraversesBridges(t *testing.T) {
    pa := NewParserAdapter()
    ast, err := pa.Parse("(+ 1 2)")
    if err != nil { t.Fatalf("parse: %v", err) }

    reg := macro.NewRegistry(macro.Config{EnableValidation: true})
    exp := macro.NewMacroExpander(reg)
    mea := NewMacroExpanderAdapter(exp)

    out, err := mea.ExpandMacros(ast)
    if err != nil || out == nil || out.Root() == nil {
        t.Fatalf("ExpandMacros failed: %v, out=%#v", err, out)
    }
}

func TestCodegenSymbolTable_LookupBridge(t *testing.T) {
    cst := &CodegenSymbolTable{table: &lookupSymTable{}}
    v, ok := cst.Lookup("x")
    if !ok || v == nil || v.String() != "ok" || v.Type() != "T" {
        t.Fatalf("unexpected lookup: ok=%v v=%#v", ok, v)
    }
}

func TestErrorReporterAdapter_Delegates(t *testing.T) {
    st := &mutableReporter{}
    era := &ErrorReporterAdapter{reporter: st}
    era.ReportError(1, 2, "m")
    if !era.HasErrors() { t.Fatalf("expected HasErrors true") }
    errs := era.GetErrors()
    if len(errs) != 1 || errs[0].Line != 1 || errs[0].Column != 2 || errs[0].Message != "m" {
        t.Fatalf("GetErrors mismatch: %#v", errs)
    }
}

// Macro visitor to assert Accept was invoked on MacroASTAdapter
type macroVisitorFlag struct{ called bool }

func (m *macroVisitorFlag) VisitProgram(ctx *parser.ProgramContext) error { m.called = true; return nil }
func (m *macroVisitorFlag) VisitList(ctx *parser.ListContext) (macro.Value, error) { return &simpleValue{val: "", typ: ""}, nil }
func (m *macroVisitorFlag) VisitArray(ctx *parser.ArrayContext) (macro.Value, error) { return &simpleValue{val: "", typ: ""}, nil }
func (m *macroVisitorFlag) VisitTerminal(node antlr.TerminalNode) (macro.Value, error) { return &simpleValue{val: node.GetText(), typ: ""}, nil }

func TestMacroASTAdapter_AcceptInvokesVisitor(t *testing.T) {
    pa := NewParserAdapter()
    ast, err := pa.Parse("(+ 1 2)")
    if err != nil { t.Fatalf("parse: %v", err) }
    maa := &MacroASTAdapter{ast: ast}
    mv := &macroVisitorFlag{}
    if err := maa.Accept(mv); err != nil { t.Fatalf("accept: %v", err) }
    if !mv.called { t.Fatalf("visitor not called via Accept") }
}


