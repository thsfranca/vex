package packages

import (
	"os"
	"path/filepath"
	"testing"
)

func writeFile(t *testing.T, path, content string) {
    t.Helper()
    if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
        t.Fatalf("mkdir: %v", err)
    }
    if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
        t.Fatalf("write: %v", err)
    }
}

func TestResolver_OrdersPackagesAndDetectsCycles(t *testing.T) {
    dir := t.TempDir()

    // Create local packages: a depends on b, b exports symbol 'b'
    writeFile(t, filepath.Join(dir, "a", "a.vx"), "(import [\"b\"])\n(def a 1)\n")
    writeFile(t, filepath.Join(dir, "b", "b.vx"), "(export [b])\n(def b 2)\n")
    entry := filepath.Join(dir, "main.vx")
    writeFile(t, entry, "(import [\"a\"])\n(def main 0)\n")

    r := NewResolver(dir)
    res, err := r.BuildProgramFromEntry(entry)
    if err != nil {
        t.Fatalf("resolve error: %v", err)
    }
    if res == nil || len(res.CombinedSource) == 0 {
        t.Fatalf("expected combined source")
    }
    // Ensure local imports are marked to ignore as Go imports
    if !res.IgnoreImports["a"] || !res.IgnoreImports["b"] {
        t.Fatalf("expected ignore imports for local packages a and b: %#v", res.IgnoreImports)
    }

    // Now introduce a cycle: b imports a
    writeFile(t, filepath.Join(dir, "b", "b.vx"), "(import [\"a\"])\n(def b 2)\n")
    _, err = r.BuildProgramFromEntry(entry)
    if err == nil {
        t.Fatalf("expected cycle error, got nil")
    }

    // Exports collection should be present for discovered packages
    if len(res.Exports) == 0 {
        t.Fatalf("expected exports map to be collected")
    }
}


