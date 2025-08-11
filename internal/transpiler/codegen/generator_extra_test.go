package codegen

import (
	"strings"
	"testing"
)

func TestGenerateImport_Direct(t *testing.T) {
    g := NewGoCodeGenerator(Config{PackageName: "main"})
    v, err := g.generateImport([]string{"\"fmt\""})
    if err != nil { t.Fatalf("generateImport error: %v", err) }
    if v == nil || !strings.Contains(v.String(), "import completed") {
        t.Fatalf("unexpected import value: %#v", v)
    }
    // Final code should include import when generated
    ast := &dummyAST{}
    // No crash when building final code; we just ensure it runs
    if _, err := g.Generate(ast, &dummySymTab{}); err != nil {
        t.Fatalf("Generate failed: %v", err)
    }
}

// Minimal AST and symtab to drive generator
type dummyAST struct{}
func (d *dummyAST) Accept(v ASTVisitor) error { return nil }

type dummySymTab struct{}
func (d *dummySymTab) Define(name string, value Value) error { return nil }
func (d *dummySymTab) Lookup(name string) (Value, bool)      { return nil, false }
func (d *dummySymTab) EnterScope()                           {}
func (d *dummySymTab) ExitScope()                            {}


