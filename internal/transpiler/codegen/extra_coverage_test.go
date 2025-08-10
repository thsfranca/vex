package codegen

import "testing"

// Removed test for deprecated generateImport simple form; generateImportFromCtx is used instead.

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


