package tests

import (
	"testing"

	"github.com/thsfranca/vex/internal/transpiler"
)

func TestTranspiler_Reset_Coverage(t *testing.T) {
	// Test the Reset method to improve coverage
	tr := transpiler.New()
	
	// Call Reset - it should not crash
	tr.Reset()
	
	// Should still be able to use transpiler after reset
	result, err := tr.TranspileFromInput("(def x 10)")
	if err != nil {
		t.Fatalf("Unexpected error after reset: %v", err)
	}
	
	if result == "" {
		t.Error("Expected non-empty result after reset")
	}
	
	// Call Reset again
	tr.Reset()
	
	// Should still work
	result2, err := tr.TranspileFromInput("(def y 20)")
	if err != nil {
		t.Fatalf("Unexpected error after second reset: %v", err)
	}
	
	if result2 == "" {
		t.Error("Expected non-empty result after second reset")
	}
}
