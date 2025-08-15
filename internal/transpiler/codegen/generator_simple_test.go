package codegen

import (
	"testing"
)

// Test SetPackageName method directly
func TestGoCodeGenerator_SetPackageName_Direct(t *testing.T) {
	generator := NewGoCodeGenerator(Config{})
	
	// Test setting package name
	generator.SetPackageName("mypackage")
	
	// Verify it was set by checking internal state
	if generator.packageName != "mypackage" {
		t.Errorf("Expected package name to be 'mypackage', got %q", generator.packageName)
	}

	// Test setting another name
	generator.SetPackageName("testpackage")
	if generator.packageName != "testpackage" {
		t.Errorf("Expected package name to be 'testpackage', got %q", generator.packageName)
	}
}

// Test generateCollectionOp method directly
func TestGoCodeGenerator_generateCollectionOp_Direct(t *testing.T) {
	tests := []struct {
		name     string
		op       string
		args     []string
		expected string
		wantType string
	}{
		{
			name:     "empty? with empty array",
			op:       "empty?",
			args:     []string{"[]interface{}{}"},
			expected: "true",
			wantType: "bool",
		},
		{
			name:     "empty? with variable",
			op:       "empty?",
			args:     []string{"arr"},
			expected: "(len(arr) == 0)",
			wantType: "bool",
		},
		{
			name:     "empty? with wrong args",
			op:       "empty?",
			args:     []string{"arr1", "arr2"},
			expected: "false",
			wantType: "bool",
		},
		{
			name:     "count operation",
			op:       "count",
			args:     []string{"myarray"},
			expected: "len(myarray)",
			wantType: "int",
		},
		{
			name:     "count with no args",
			op:       "count",
			args:     []string{},
			expected: "0",
			wantType: "int",
		},
		{
			name:     "unknown operation",
			op:       "unknown-op",
			args:     []string{"arg"},
			expected: "nil",
			wantType: "interface{}",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			generator := NewGoCodeGenerator(Config{PackageName: "main"})

			result, err := generator.generateCollectionOp(tt.op, tt.args)
			if err != nil {
				t.Fatalf("generateCollectionOp failed: %v", err)
			}

			if result.String() != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result.String())
			}

			if result.Type() != tt.wantType {
				t.Errorf("Expected type %q, got %q", tt.wantType, result.Type())
			}
		})
	}
}

// Test generateMacro method directly
func TestGoCodeGenerator_generateMacro_Direct(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		expected string
		wantType string
	}{
		{
			name:     "Macro with arguments",
			args:     []string{"my-macro", "[x]", "body"},
			expected: "",
			wantType: "void",
		},
		{
			name:     "Macro with no arguments",
			args:     []string{},
			expected: "",
			wantType: "void",
		},
		{
			name:     "Complex macro definition",
			args:     []string{"complex-macro", "[a: int b: string]", "(+ a 1)"},
			expected: "",
			wantType: "void",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			generator := NewGoCodeGenerator(Config{PackageName: "main"})

			result, err := generator.generateMacro(tt.args)
			if err != nil {
				t.Fatalf("generateMacro failed: %v", err)
			}

			if result.String() != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result.String())
			}

			if result.Type() != tt.wantType {
				t.Errorf("Expected type %q, got %q", tt.wantType, result.Type())
			}
		})
	}
}

// Simple mock symbol table
type SimpleMockSymbolTable struct{}

func (smst *SimpleMockSymbolTable) Define(name string, value Value) error { return nil }
func (smst *SimpleMockSymbolTable) Lookup(name string) (Value, bool)      { return nil, false }
func (smst *SimpleMockSymbolTable) EnterScope()                            {}
func (smst *SimpleMockSymbolTable) ExitScope()                             {}
