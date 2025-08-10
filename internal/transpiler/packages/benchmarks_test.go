package packages

import (
	"os"
	"path/filepath"
	"testing"
)

func BenchmarkResolver_SmallProject(b *testing.B) {
    // Use the repo's examples/valid/project-multi as a realistic small project
    // From repo root one level up from this package dir
    // Walk up until we find go.mod
    dir, _ := os.Getwd()
    for {
        if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
            break
        }
        parent := filepath.Dir(dir)
        if parent == dir {
            b.Fatal("could not locate repository root")
        }
        dir = parent
    }
    entry := filepath.Join(dir, "examples/valid/project-multi/main.vx")
    r := NewResolver(dir)
    b.ReportAllocs()
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        if _, err := r.BuildProgramFromEntry(entry); err != nil {
            b.Fatal(err)
        }
    }
}


