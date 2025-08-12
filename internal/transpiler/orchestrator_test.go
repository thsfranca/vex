package transpiler

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	packages "github.com/thsfranca/vex/internal/transpiler/packages"
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

// Moved from orchestrator_extras_test.go
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

func TestOrchestrator_isThirdPartyModule_Heuristics(t *testing.T) {
    vt, err := NewBuilder().WithMacros(false).WithPackageName("main").Build()
    if err != nil {
        t.Fatalf("build: %v", err)
    }
    if !vt.isThirdPartyModule("golang.org/x/mod") {
        t.Fatalf("expected golang.org/x/mod to be treated as third-party")
    }
    if vt.isThirdPartyModule("fmt") {
        t.Fatalf("expected fmt to not be treated as third-party")
    }
}

func TestOrchestrator_astToString_Placeholder(t *testing.T) {
    vt, err := NewBuilder().WithMacros(false).WithPackageName("main").Build()
    if err != nil { t.Fatalf("build: %v", err) }
    pa := NewParserAdapter()
    ast, err := pa.Parse("(+ 1 2)")
    if err != nil { t.Fatalf("parse: %v", err) }
    if s := vt.astToString(ast); s == "" {
        t.Fatalf("astToString placeholder should not be empty")
    }
}

func writeFileOrch(t *testing.T, path, content string) {
    t.Helper()
    if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
        t.Fatalf("mkdir: %v", err)
    }
    if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
        t.Fatalf("write: %v", err)
    }
}

func TestEndToEnd_CrossPackageSchemes_Transpile(t *testing.T) {
    dir := t.TempDir()
    // Package a exporting id function
    writeFileOrch(t, filepath.Join(dir, "a", "a.vx"), "(import \"vex.core\")\n(export [id])\n(defn id [x: interface{}] -> interface{} x)\n")
    // Entry that calls a/id with int and string
    entry := filepath.Join(dir, "main.vx")
    writeFileOrch(t, entry, "(import [\"a\"])\n(def main (do (a/id 1) (a/id \"s\")))\n")

    r := packages.NewResolver(dir)
    res, err := r.BuildProgramFromEntry(entry)
    if err != nil { t.Fatalf("resolve error: %v", err) }

    cfg := TranspilerConfig{
        EnableMacros:     true,
        PackageName:      "main",
        GenerateComments: false,
        IgnoreImports:    res.IgnoreImports,
        Exports:          res.Exports,
        PkgSchemes:       res.PkgSchemes,
    }
    tImpl, err := NewTranspilerWithConfig(cfg)
    if err != nil { t.Fatalf("builder error: %v", err) }
    if _, err := tImpl.TranspileFromInput(res.CombinedSource); err != nil {
        t.Fatalf("transpile should succeed using cross-package schemes: %v", err)
    }

    // Now assert that non-exported symbol errors at analyzer
    writeFileOrch(t, entry, "(import [\"a\"])\n(def main (a/hidden))\n")
    res2, err := r.BuildProgramFromEntry(entry)
    if err != nil { t.Fatalf("resolve error: %v", err) }
    cfg.IgnoreImports = res2.IgnoreImports
    cfg.Exports = res2.Exports
    cfg.PkgSchemes = res2.PkgSchemes
    tImpl2, err := NewTranspilerWithConfig(cfg)
    if err != nil { t.Fatalf("builder2 error: %v", err) }
    if _, err := tImpl2.TranspileFromInput(res2.CombinedSource); err == nil {
        t.Fatalf("expected error when calling non-exported symbol a/hidden")
    }
}


