package macro

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRegistry_LoadFromPath_FileAndDir(t *testing.T) {
    // Create a temporary .vx file with a simple macro
    dir := t.TempDir()
    file := filepath.Join(dir, "m.vx")
    if err := os.WriteFile(file, []byte("(macro inc [x] (+ x 1))\n"), 0644); err != nil {
        t.Fatalf("write: %v", err)
    }

    // Load by file path
    reg := NewRegistry(Config{EnableValidation: true, CoreMacroPath: file})
    if err := reg.LoadCoreMacros(); err != nil { t.Fatalf("LoadCoreMacros(file): %v", err) }
    if !reg.HasMacro("inc") { t.Fatalf("expected macro 'inc' after file load") }

    // Load by directory path
    reg2 := NewRegistry(Config{EnableValidation: true, CoreMacroPath: dir})
    if err := reg2.LoadCoreMacros(); err != nil { t.Fatalf("LoadCoreMacros(dir): %v", err) }
    if !reg2.HasMacro("inc") { t.Fatalf("expected macro 'inc' after dir load") }
}

func TestRegistry_ParseParameterList_VexAndGoFormats(t *testing.T) {
    reg := NewRegistry(Config{})
    // Vex format
    params, err := reg.parseParameterList("[x y]")
    if err != nil || len(params) != 2 || params[0] != "x" || params[1] != "y" {
        t.Fatalf("vex params parse failed: %v %#v", err, params)
    }
    // Go-format guard should return error
    if _, err := reg.parseParameterList("[]interface{}{}"); err == nil {
        t.Fatalf("expected error for empty go-format list")
    }
}


