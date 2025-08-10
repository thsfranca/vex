package transpiler

import (
	"strings"
	"testing"
)

func TestTranspilerBuilder_WithersAndBuild(t *testing.T) {
    cfg := TranspilerConfig{
        EnableMacros:     false,
        CoreMacroPath:    "unused",
        PackageName:      "initial",
        GenerateComments: true,
        IgnoreImports:    make(map[string]bool),
        Exports:          make(map[string]map[string]bool),
    }

    b := NewBuilder().WithConfig(cfg).WithPackageName("main").WithCoreMacroPath("still-unused").WithMacros(false)
    vt, err := b.Build()
    if err != nil {
        t.Fatalf("Build failed: %v", err)
    }
    if vt == nil {
        t.Fatalf("expected non-nil transpiler")
    }
}

func TestOrchestrator_TranspileFromInput_ThirdPartyDetection(t *testing.T) {
    vt, err := NewBuilder().WithMacros(false).WithPackageName("main").Build()
    if err != nil {
        t.Fatalf("build: %v", err)
    }

    src := "(import \"golang.org/x/mod\")\n(def x 1)"
    out, err := vt.TranspileFromInput(src)
    if err != nil {
        t.Fatalf("TranspileFromInput error: %v", err)
    }
    if !strings.Contains(out, "import \"golang.org/x/mod\"") {
        t.Fatalf("expected go import for third-party module, got:\n%s", out)
    }
    mods := vt.GetDetectedModules()
    if _, ok := mods["golang.org/x/mod"]; !ok {
        t.Fatalf("expected third-party module to be detected")
    }
}


