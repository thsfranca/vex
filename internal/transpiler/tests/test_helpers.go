// Package tests provides test utilities and helpers for the transpiler
package tests

import (
	"strings"
	"testing"

	"github.com/thsfranca/vex/internal/transpiler"
)

// TestCase represents a test case for transpiler testing
type TestCase struct {
	Name     string
	Input    string
	Expected string
	Error    bool
}

// AssertTranspileSuccess tests that input transpiles to expected output
func AssertTranspileSuccess(t *testing.T, input, expected string) {
	t.Helper()
	
	tr := transpiler.New()
	result, err := tr.TranspileFromInput(input)
	
	if err != nil {
		t.Fatalf("Expected successful transpilation, got error: %v", err)
	}
	
	if !strings.Contains(result, expected) {
		t.Errorf("Expected output to contain:\n%s\n\nActual output:\n%s", expected, result)
	}
}

// AssertTranspileError tests that input produces an error during transpilation
func AssertTranspileError(t *testing.T, input string) {
	t.Helper()
	
	tr := transpiler.New()
	_, err := tr.TranspileFromInput(input)
	
	if err == nil {
		t.Error("Expected transpilation error, but got success")
	}
}

// AssertContainsAll checks that result contains all expected strings
func AssertContainsAll(t *testing.T, result string, expected ...string) {
	t.Helper()
	
	for _, exp := range expected {
		if !strings.Contains(result, exp) {
			t.Errorf("Expected result to contain: %s\nActual result:\n%s", exp, result)
		}
	}
}

// AssertNotContains checks that result does not contain any of the strings
func AssertNotContains(t *testing.T, result string, unexpected ...string) {
	t.Helper()
	
	for _, unexp := range unexpected {
		if strings.Contains(result, unexp) {
			t.Errorf("Expected result NOT to contain: %s\nActual result:\n%s", unexp, result)
		}
	}
}

// NormalizeWhitespace normalizes whitespace for comparison
func NormalizeWhitespace(s string) string {
	lines := strings.Split(s, "\n")
	var normalized []string
	
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" {
			normalized = append(normalized, trimmed)
		}
	}
	
	return strings.Join(normalized, "\n")
}

// RunTestCases runs a slice of test cases
func RunTestCases(t *testing.T, cases []TestCase) {
	t.Helper()
	
	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			tr := transpiler.New()
			result, err := tr.TranspileFromInput(tc.Input)
			
			if tc.Error {
				if err == nil {
					t.Error("Expected error, but got success")
				}
				return
			}
			
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
			
			if tc.Expected != "" && !strings.Contains(result, tc.Expected) {
				t.Errorf("Expected output to contain:\n%s\n\nActual output:\n%s", tc.Expected, result)
			}
		})
	}
}
