package transpiler

import (
	"testing"

	"github.com/thsfranca/vex/internal/transpiler/codegen"
)

// This exercises bridge visitor paths that were previously uncovered.
func TestAdapters_BridgeVisitors_TraverseAll(t *testing.T) {
    pa := NewParserAdapter()
    // Safe program for analyzer traversal
    astForAnalysis, err := pa.Parse("(do (def x 1) (+ x 2) (+ x 3))")
    if err != nil { t.Fatalf("parse: %v", err) }

    aa := NewAnalyzerAdapter()
    aa.SetErrorReporter(&testReporter{})
    if _, err := aa.Analyze(astForAnalysis); err != nil {
        t.Fatalf("analyze: %v", err)
    }

    // Separate program including an array to traverse array visitor paths during codegen
    astForCodegen, err := pa.Parse("(do (def x 1) [x 2] (+ x 3))")
    if err != nil { t.Fatalf("parse for codegen: %v", err) }
    cfg := codegen.Config{PackageName: "main", GenerateComments: false, IndentSize: 4, IgnoreImports: map[string]bool{}}
    cga := NewCodeGeneratorAdapter(cfg)
    if code, err := cga.Generate(astForCodegen, &dummySymTable{}); err != nil || code == "" {
        t.Fatalf("generate failed: %v, code=%q", err, code)
    }
}


