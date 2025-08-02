package transpiler

import (
	"strings"
	"testing"
)

func TestNew(t *testing.T) {
	transpiler := New()
	if transpiler == nil {
		t.Fatal("New() returned nil")
	}
}

func TestTranspileToGo_Placeholder(t *testing.T) {
	transpiler := New()
	
	result, err := transpiler.TranspileToGo("(+ 1 2)")
	if err != nil {
		t.Fatalf("TranspileToGo failed: %v", err)
	}
	
	// Verify it generates basic Go code structure
	if !strings.Contains(result, "package main") {
		t.Error("Expected 'package main' in output")
	}
	
	if !strings.Contains(result, "func main()") {
		t.Error("Expected 'func main()' in output")
	}
	
	t.Logf("Generated Go code:\n%s", result)
}