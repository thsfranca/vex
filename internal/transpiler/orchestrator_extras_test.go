package transpiler

import "testing"

func TestOrchestrator_GettersAndAstToString(t *testing.T) {
    vt, err := NewBuilder().WithMacros(false).WithPackageName("main").Build()
    if err != nil { t.Fatalf("build: %v", err) }

    if vt.GetMacroSystem() != nil {
        t.Fatalf("expected no macro system when disabled")
    }
    if vt.GetAnalyzer() == nil || vt.GetCodeGenerator() == nil {
        t.Fatalf("expected analyzer and code generator")
    }
    if vt.GetDetectedModules() == nil {
        t.Fatalf("expected detected modules map")
    }

    pa := NewParserAdapter()
    ast, err := pa.Parse("(+ 1 2)")
    if err != nil { t.Fatalf("parse: %v", err) }
    if s := vt.astToString(ast); s == "" {
        t.Fatalf("astToString should return placeholder, got empty string")
    }
}


