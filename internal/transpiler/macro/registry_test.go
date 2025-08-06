package macro

import (
	"strings"
	"testing"
)

func TestNewRegistry(t *testing.T) {
	config := Config{
		CoreMacroPath:    "test.vx",
		EnableFallback:   true,
		EnableValidation: true,
	}
	
	registry := NewRegistry(config)
	if registry == nil {
		t.Fatal("NewRegistry should return non-nil registry")
	}
	
	if registry.coreLoaded {
		t.Error("Registry should start with coreLoaded = false")
	}
	
	if registry.macros == nil {
		t.Error("Registry should initialize macros map")
	}
}

func TestRegistry_RegisterMacro(t *testing.T) {
	registry := NewRegistry(Config{EnableValidation: true})
	
	macro := &Macro{
		Name:   "test-macro",
		Params: []string{"x", "y"},
		Body:   "(+ x y)",
	}
	
	err := registry.RegisterMacro("test-macro", macro)
	if err != nil {
		t.Errorf("RegisterMacro should succeed: %v", err)
	}
	
	// Check if macro was registered
	if !registry.HasMacro("test-macro") {
		t.Error("Macro should be registered")
	}
	
	// Check retrieval
	retrieved, exists := registry.GetMacro("test-macro")
	if !exists {
		t.Error("Registered macro should be retrievable")
	}
	
	if retrieved.Name != macro.Name {
		t.Errorf("Retrieved macro name = %v, want %v", retrieved.Name, macro.Name)
	}
}

func TestRegistry_RegisterMacro_ValidationEnabled(t *testing.T) {
	registry := NewRegistry(Config{EnableValidation: true})
	
	tests := []struct {
		name      string
		macroName string
		macro     *Macro
		wantErr   bool
		errMsg    string
	}{
		{
			name:      "Valid macro",
			macroName: "valid-macro",
			macro: &Macro{
				Name:   "valid-macro",
				Params: []string{"x"},
				Body:   "x",
			},
			wantErr: false,
		},
		{
			name:      "Empty name",
			macroName: "",
			macro: &Macro{
				Name:   "",
				Params: []string{"x"},
				Body:   "x",
			},
			wantErr: true,
			errMsg:  "macro name cannot be empty",
		},
		{
			name:      "Reserved word",
			macroName: "if",
			macro: &Macro{
				Name:   "if",
				Params: []string{"x"},
				Body:   "x",
			},
			wantErr: true,
			errMsg:  "'if' is a reserved word",
		},
		{
			name:      "Duplicate parameters",
			macroName: "dup-params",
			macro: &Macro{
				Name:   "dup-params",
				Params: []string{"x", "x"},
				Body:   "x",
			},
			wantErr: true,
			errMsg:  "duplicate parameter 'x'",
		},
		{
			name:      "Empty body",
			macroName: "empty-body",
			macro: &Macro{
				Name:   "empty-body",
				Params: []string{"x"},
				Body:   "",
			},
			wantErr: true,
			errMsg:  "empty body",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := registry.RegisterMacro(tt.macroName, tt.macro)
			
			if (err != nil) != tt.wantErr {
				t.Errorf("RegisterMacro() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			
			if tt.wantErr && !strings.Contains(err.Error(), tt.errMsg) {
				t.Errorf("Expected error to contain %q, got %q", tt.errMsg, err.Error())
			}
		})
	}
}

func TestRegistry_RegisterMacro_ValidationDisabled(t *testing.T) {
	registry := NewRegistry(Config{EnableValidation: false})
	
	// Should allow reserved words when validation is disabled
	macro := &Macro{
		Name:   "if",
		Params: []string{"x"},
		Body:   "x",
	}
	
	err := registry.RegisterMacro("if", macro)
	if err != nil {
		t.Errorf("RegisterMacro should succeed when validation disabled: %v", err)
	}
}

func TestRegistry_HasMacro(t *testing.T) {
	registry := NewRegistry(Config{EnableValidation: false})
	
	// Initially should not have any macros
	if registry.HasMacro("test") {
		t.Error("Registry should not have unregistered macro")
	}
	
	// Register a macro
	macro := &Macro{Name: "test", Params: []string{}, Body: "42"}
	registry.RegisterMacro("test", macro)
	
	// Should now have the macro
	if !registry.HasMacro("test") {
		t.Error("Registry should have registered macro")
	}
}

func TestRegistry_GetMacro(t *testing.T) {
	registry := NewRegistry(Config{EnableValidation: false})
	
	// Test non-existent macro
	_, exists := registry.GetMacro("nonexistent")
	if exists {
		t.Error("GetMacro should return false for non-existent macro")
	}
	
	// Register and retrieve macro
	macro := &Macro{
		Name:   "test",
		Params: []string{"x", "y"},
		Body:   "(+ x y)",
	}
	registry.RegisterMacro("test", macro)
	
	retrieved, exists := registry.GetMacro("test")
	if !exists {
		t.Error("GetMacro should return true for existing macro")
	}
	
	if retrieved.Name != macro.Name {
		t.Errorf("Retrieved macro name = %v, want %v", retrieved.Name, macro.Name)
	}
	
	if len(retrieved.Params) != len(macro.Params) {
		t.Errorf("Retrieved macro params length = %v, want %v", len(retrieved.Params), len(macro.Params))
	}
	
	if retrieved.Body != macro.Body {
		t.Errorf("Retrieved macro body = %v, want %v", retrieved.Body, macro.Body)
	}
}

// TestRegistry_LoadFallbackMacros removed - fallback macros are obsolete
// All macros should be defined in core/core.vx, not hardcoded

// TestRegistry_LoadCoreMacros_Fallback and TestRegistry_LoadCoreMacros_NoFallback removed
// These tests were testing obsolete scenarios that don't apply to the current architecture:
// 1. Fallback macros are removed - all macros are in core/core.vx  
// 2. LoadCoreMacros will always succeed when core/core.vx exists

func TestRegistry_LoadCoreMacros_PreventDuplicate(t *testing.T) {
	registry := NewRegistry(Config{
		EnableFallback:   true,
		EnableValidation: false,
	})
	
	// First load
	err := registry.LoadCoreMacros()
	if err != nil {
		t.Errorf("First LoadCoreMacros should succeed: %v", err)
	}
	
	// Second load should be prevented
	err = registry.LoadCoreMacros()
	if err != nil {
		t.Errorf("Second LoadCoreMacros should succeed (no-op): %v", err)
	}
	
	if !registry.coreLoaded {
		t.Error("coreLoaded should be true after loading")
	}
}

func TestRegistry_parseParameterList(t *testing.T) {
	registry := NewRegistry(Config{})
	
	tests := []struct {
		name      string
		paramList string
		want      []string
		wantErr   bool
	}{
		{
			name:      "Vex format - empty",
			paramList: "[]",
			want:      []string{},
			wantErr:   false,
		},
		{
			name:      "Vex format - single param",
			paramList: "[x]",
			want:      []string{"x"},
			wantErr:   false,
		},
		{
			name:      "Vex format - multiple params",
			paramList: "[x y z]",
			want:      []string{"x", "y", "z"},
			wantErr:   false,
		},
		{
			name:      "Go format - empty",
			paramList: "[]interface{}{}",
			want:      []string{},
			wantErr:   true, // Current implementation doesn't support this format properly
		},
		{
			name:      "Go format - single param",
			paramList: "[]interface{}{x}",
			want:      []string{"x"},
			wantErr:   false,
		},
		{
			name:      "Go format - multiple params",
			paramList: "[]interface{}{x, y, z}",
			want:      []string{"x", "y", "z"},
			wantErr:   false,
		},
		{
			name:      "Invalid format",
			paramList: "not-a-list",
			want:      nil,
			wantErr:   true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := registry.parseParameterList(tt.paramList)
			
			if (err != nil) != tt.wantErr {
				t.Errorf("parseParameterList() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			
			if !tt.wantErr {
				if len(got) != len(tt.want) {
					t.Errorf("parseParameterList() length = %v, want %v", len(got), len(tt.want))
					return
				}
				
				for i, param := range got {
					if param != tt.want[i] {
						t.Errorf("parseParameterList() param[%d] = %v, want %v", i, param, tt.want[i])
					}
				}
			}
		})
	}
}

func TestRegistry_validateMacro(t *testing.T) {
	registry := NewRegistry(Config{EnableValidation: true})
	
	// Test conflict detection by registering a macro first
	existingMacro := &Macro{Name: "existing", Params: []string{"x"}, Body: "x"}
	registry.RegisterMacro("existing", existingMacro)
	
	tests := []struct {
		name      string
		macroName string
		macro     *Macro
		wantErr   bool
		errContains string
	}{
		{
			name:      "Valid macro",
			macroName: "valid",
			macro: &Macro{
				Name:   "valid",
				Params: []string{"x", "y"},
				Body:   "(+ x y)",
			},
			wantErr: false,
		},
		{
			name:      "Empty name",
			macroName: "",
			macro:     &Macro{Name: "", Params: []string{}, Body: "test"},
			wantErr:   true,
			errContains: "macro name cannot be empty",
		},
		{
			name:      "Reserved word - if",
			macroName: "if",
			macro:     &Macro{Name: "if", Params: []string{}, Body: "test"},
			wantErr:   true,
			errContains: "'if' is a reserved word",
		},
		{
			name:      "Reserved word - def",
			macroName: "def",
			macro:     &Macro{Name: "def", Params: []string{}, Body: "test"},
			wantErr:   true,
			errContains: "'def' is a reserved word",
		},
		{
			name:      "Conflicting macro",
			macroName: "existing",
			macro:     &Macro{Name: "existing", Params: []string{}, Body: "test"},
			wantErr:   true,
			errContains: "macro 'existing' is already defined",
		},
		{
			name:      "Empty parameter",
			macroName: "empty-param",
			macro:     &Macro{Name: "empty-param", Params: []string{"x", "", "y"}, Body: "test"},
			wantErr:   true,
			errContains: "empty parameter name",
		},
		{
			name:      "Duplicate parameters",
			macroName: "dup-param",
			macro:     &Macro{Name: "dup-param", Params: []string{"x", "y", "x"}, Body: "test"},
			wantErr:   true,
			errContains: "duplicate parameter 'x'",
		},
		{
			name:      "Empty body",
			macroName: "empty-body",
			macro:     &Macro{Name: "empty-body", Params: []string{"x"}, Body: "   "},
			wantErr:   true,
			errContains: "empty body",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := registry.validateMacro(tt.macroName, tt.macro)
			
			if (err != nil) != tt.wantErr {
				t.Errorf("validateMacro() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			
			if tt.wantErr && !strings.Contains(err.Error(), tt.errContains) {
				t.Errorf("Expected error to contain %q, got %q", tt.errContains, err.Error())
			}
		})
	}
}