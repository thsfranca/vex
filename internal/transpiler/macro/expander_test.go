package macro

import (
	"strings"
	"testing"
)

func TestNewExpander(t *testing.T) {
	registry := NewRegistry(Config{EnableValidation: false})
	expander := NewExpander(registry)
	
	if expander == nil {
		t.Fatal("NewExpander should return non-nil expander")
	}
	
	if expander.registry != registry {
		t.Error("Expander should store the provided registry")
	}
}

// extra: from expander_extra_test.go
func TestExpander_SetParent_NoPanic(t *testing.T) {
    // Ensure substitutedTerminal.SetParent is called without panicking
    st := &substitutedTerminal{text: "x"}
    st.SetParent(nil)
}

func TestExpander_ExpandMacro(t *testing.T) {
	registry := NewRegistry(Config{EnableValidation: false})
	
	// Register test macros
	simpleMacro := &Macro{
		Name:   "simple",
		Params: []string{"x"},
		Body:   "x",
	}
	registry.RegisterMacro("simple", simpleMacro)
	
	addMacro := &Macro{
		Name:   "add",
		Params: []string{"a", "b"},
		Body:   "(+ a b)",
	}
	registry.RegisterMacro("add", addMacro)
	
	expander := NewExpander(registry)
	
	tests := []struct {
		name      string
		macroName string
		args      []string
		want      string
		wantErr   bool
	}{
		{
			name:      "Simple substitution",
			macroName: "simple",
			args:      []string{"42"},
			want:      "42",
			wantErr:   false,
		},
		{
			name:      "Multiple parameter substitution",
			macroName: "add",
			args:      []string{"10", "20"},
			want:      "(+ 10 20)", // Properly spaced syntax
			wantErr:   false,
		},
		{
			name:      "Non-existent macro",
			macroName: "nonexistent",
			args:      []string{"x"},
			want:      "",
			wantErr:   true,
		},
		{
			name:      "Wrong argument count",
			macroName: "add",
			args:      []string{"10"},
			want:      "",
			wantErr:   true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := expander.ExpandMacro(tt.macroName, tt.args)
			
			if (err != nil) != tt.wantErr {
				t.Errorf("ExpandMacro() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			
			if !tt.wantErr && got != tt.want {
				t.Errorf("ExpandMacro() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestExpander_substituteParameters(t *testing.T) {
	registry := NewRegistry(Config{EnableValidation: false})
	expander := NewExpander(registry)
	
	tests := []struct {
		name   string
		body   string
		params []string
		args   []string
		want   string
	}{
		{
			name:   "Single parameter",
			body:   "x",
			params: []string{"x"},
			args:   []string{"42"},
			want:   "42",
		},
		{
			name:   "Multiple parameters",
			body:   "(+ x y)",
			params: []string{"x", "y"},
			args:   []string{"10", "20"},
			want:   "(+ 10 20)", // Properly spaced syntax
		},
		{
			name:   "No parameters",
			body:   "42",
			params: []string{},
			args:   []string{},
			want:   "42",
		},
		{
			name:   "Parameter in nested expression",
			body:   "(* (+ x 1) y)",
			params: []string{"x", "y"},
			args:   []string{"5", "3"},
			want:   "(* (+ 5 1) 3)", // Properly spaced syntax
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := expander.substituteParameters(tt.body, tt.params, tt.args)
			if got != tt.want {
				t.Errorf("substituteParameters() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestExpander_replaceWholeWord(t *testing.T) {
	registry := NewRegistry(Config{EnableValidation: false})
	expander := NewExpander(registry)
	
	tests := []struct {
		name  string
		text  string
		param string
		arg   string
		want  string
	}{
		{
			name:  "Simple replacement",
			text:  "x",
			param: "x",
			arg:   "42",
			want:  "42",
		},
		{
			name:  "Word boundary - start",
			text:  "x + y",
			param: "x",
			arg:   "42",
			want:  "42 + y",
		},
		{
			name:  "Word boundary - middle",
			text:  "a + x + b",
			param: "x",
			arg:   "42",
			want:  "a + 42 + b",
		},
		{
			name:  "Word boundary - end",
			text:  "y + x",
			param: "x",
			arg:   "42",
			want:  "y + 42",
		},
		{
			name:  "No replacement when part of word",
			text:  "xhr",
			param: "x",
			arg:   "42",
			want:  "xhr",
		},
		{
			name:  "Multiple occurrences",
			text:  "x + x",
			param: "x",
			arg:   "42",
			want:  "42 + 42",
		},
		{
			name:  "With parentheses",
			text:  "(+ x y)",
			param: "x",
			arg:   "42",
			want:  "(+ 42 y)",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := expander.replaceWholeWord(tt.text, tt.param, tt.arg)
			if got != tt.want {
				t.Errorf("replaceWholeWord() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestExpander_isIdentifierChar(t *testing.T) {
	registry := NewRegistry(Config{EnableValidation: false})
	expander := NewExpander(registry)
	
	tests := []struct {
		char rune
		want bool
	}{
		{'a', true},
		{'Z', true},
		{'0', true},
		{'9', true},
		{'_', true},
		{'-', true},
		{'?', true},
		{'!', true},
		{' ', false},
		{'(', false},
		{')', false},
		{'+', false},
		{'*', false},
	}
	
	for _, tt := range tests {
		t.Run(string(tt.char), func(t *testing.T) {
			got := expander.isIdentifierChar(tt.char)
			if got != tt.want {
				t.Errorf("isIdentifierChar(%q) = %v, want %v", tt.char, got, tt.want)
			}
		})
	}
}

func TestExpander_recursivelyExpandMacros(t *testing.T) {
	registry := NewRegistry(Config{EnableValidation: false})
	
	// Register nested macros
	innerMacro := &Macro{
		Name:   "inner",
		Params: []string{"x"},
		Body:   "(+ x 1)",
	}
	registry.RegisterMacro("inner", innerMacro)
	
	outerMacro := &Macro{
		Name:   "outer",
		Params: []string{"y"},
		Body:   "(inner y)",
	}
	registry.RegisterMacro("outer", outerMacro)
	
	expander := NewExpander(registry)
	
	tests := []struct {
		name    string
		expr    string
		want    string
		wantErr bool
	}{
		{
			name:    "No macro call",
			expr:    "42",
			want:    "42",
			wantErr: false,
		},
		{
			name:    "Simple macro call",
			expr:    "(inner 5)",
			want:    "(+ 5 1)", // Properly spaced syntax
			wantErr: false,
		},
		{
			name:    "Nested macro call",
			expr:    "(outer 10)",
			want:    "(+ 10 1)", // Shows full expansion with proper spacing
			wantErr: false,
		},
		{
			name:    "Non-macro function call",
			expr:    "(unknown-func 1 2)",
			want:    "(unknown-func 1 2)",
			wantErr: false,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := expander.recursivelyExpandMacros(tt.expr)
			
			if (err != nil) != tt.wantErr {
				t.Errorf("recursivelyExpandMacros() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			
			if got != tt.want {
				t.Errorf("recursivelyExpandMacros() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestExpander_reconstructVexSyntax(t *testing.T) {
	registry := NewRegistry(Config{EnableValidation: false})
	expander := NewExpander(registry)
	
	// Test with substituted terminal
	terminal := &substitutedTerminal{text: "42"}
	result := expander.reconstructVexSyntax(terminal)
	
	if result != "42" {
		t.Errorf("reconstructVexSyntax() = %v, want %v", result, "42")
	}
}

func TestSubstitutedTerminal(t *testing.T) {
	terminal := &substitutedTerminal{text: "test-value"}
	
	// Test all methods
	if terminal.GetText() != "test-value" {
		t.Errorf("GetText() = %v, want %v", terminal.GetText(), "test-value")
	}
	
	if terminal.GetSymbol() != nil {
		t.Error("GetSymbol() should return nil")
	}
	
	if terminal.Accept(nil) != nil {
		t.Error("Accept() should return nil")
	}
	
	if terminal.GetChild(0) != nil {
		t.Error("GetChild() should return nil")
	}
	
	if terminal.GetChildCount() != 0 {
		t.Error("GetChildCount() should return 0")
	}
	
	if terminal.GetChildren() != nil {
		t.Error("GetChildren() should return nil")
	}
	
	if terminal.GetParent() != nil {
		t.Error("GetParent() should return nil")
	}
	
	if terminal.GetPayload() != "test-value" {
		t.Errorf("GetPayload() = %v, want %v", terminal.GetPayload(), "test-value")
	}
	
	interval := terminal.GetSourceInterval()
	if interval.Start != -1 || interval.Stop != -1 {
		t.Errorf("GetSourceInterval() = %v, want {-1, -1}", interval)
	}
	
	// Test SetParent (should not panic)
	terminal.SetParent(nil)
	
	result := terminal.ToStringTree(nil, nil)
	if result != "test-value" {
		t.Errorf("ToStringTree() = %v, want %v", result, "test-value")
	}
}

func TestExpander_ComplexMacroExpansion(t *testing.T) {
	registry := NewRegistry(Config{EnableValidation: false})
	
	// Register defn macro (like in the real system)
	defnMacro := &Macro{
		Name:   "defn",
		Params: []string{"name", "params", "body"},
		Body:   "(def name (fn params body))",
	}
	registry.RegisterMacro("defn", defnMacro)
	
	expander := NewExpander(registry)
	
	// Test complex expansion
	args := []string{"add", "[x y]", "(+ x y)"}
	result, err := expander.ExpandMacro("defn", args)
	
	if err != nil {
		t.Errorf("Complex macro expansion failed: %v", err)
		return
	}
	
	expected := "(def add (fn [ x y ] (+ x y)))" // Properly spaced syntax
	if result != expected {
		t.Errorf("Complex expansion = %v, want %v", result, expected)
	}
}

func TestExpander_ErrorCases(t *testing.T) {
	registry := NewRegistry(Config{EnableValidation: false})
	expander := NewExpander(registry)
	
	tests := []struct {
		name      string
		macroName string
		args      []string
		errMsg    string
	}{
		{
			name:      "Macro not found",
			macroName: "nonexistent",
			args:      []string{"x"},
			errMsg:    "macro 'nonexistent' not found",
		},
		{
			name:      "Too few arguments",
			macroName: "test",
			args:      []string{},
			errMsg:    "expects 2 arguments, got 0",
		},
		{
			name:      "Too many arguments",
			macroName: "test",
			args:      []string{"a", "b", "c"},
			errMsg:    "expects 2 arguments, got 3",
		},
	}
	
	// Register a test macro for argument count tests
	testMacro := &Macro{
		Name:   "test",
		Params: []string{"x", "y"},
		Body:   "(+ x y)",
	}
	registry.RegisterMacro("test", testMacro)
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := expander.ExpandMacro(tt.macroName, tt.args)
			
			if err == nil {
				t.Error("Expected error but got none")
				return
			}
			
			if !strings.Contains(err.Error(), tt.errMsg) {
				t.Errorf("Expected error to contain %q, got %q", tt.errMsg, err.Error())
			}
		})
	}
}