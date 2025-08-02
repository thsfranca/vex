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
	
	// Use valid Fugo syntax that our grammar supports
	result, err := transpiler.TranspileToGo("(hello world)")
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

func TestTranspileToGo_IntegerLiteral(t *testing.T) {
	transpiler := New()
	
	// Test transpiling a simple integer literal
	result, err := transpiler.TranspileToGo("(42)")
	if err != nil {
		t.Fatalf("TranspileToGo failed: %v", err)
	}
	
	// Check for integer literal transpilation
	if !strings.Contains(result, "var _ int = 42") {
		t.Errorf("Expected 'var _ int = 42' in output, got:\n%s", result)
	}
	
	t.Logf("Generated Go code:\n%s", result)
}

func TestTranspileToGo_StringLiteral(t *testing.T) {
	transpiler := New()
	
	result, err := transpiler.TranspileToGo("(\"hello\")")
	if err != nil {
		t.Fatalf("TranspileToGo failed: %v", err)
	}
	
	if !strings.Contains(result, "var _ string = \"hello\"") {
		t.Errorf("Expected 'var _ string = \"hello\"' in output, got:\n%s", result)
	}
	
	t.Logf("Generated Go code:\n%s", result)
}

func TestTranspileToGo_SimpleAddition(t *testing.T) {
	transpiler := New()
	
	// For now, test that we handle multiple literals correctly
	// TODO: Add arithmetic when grammar supports operators
	result, err := transpiler.TranspileToGo("(add 1 2)")
	if err != nil {
		t.Fatalf("TranspileToGo failed: %v", err)
	}
	
	// Should parse as symbols, not arithmetic yet
	if !strings.Contains(result, "package main") {
		t.Error("Expected valid Go package")
	}
	
	t.Logf("Generated Go code:\n%s", result)
}