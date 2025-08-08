package analysis

import (
	"testing"
)

func TestNewBasicValue(t *testing.T) {
	value := NewBasicValue("42", "int")
	
	if value.String() != "42" {
		t.Errorf("BasicValue.String() = %v, want %v", value.String(), "42")
	}
	
	if value.Type() != "int" {
		t.Errorf("BasicValue.Type() = %v, want %v", value.Type(), "int")
	}
}

func TestBasicValue_StringType(t *testing.T) {
	value := NewBasicValue("\"hello\"", "string")
	
	if value.String() != "\"hello\"" {
		t.Errorf("BasicValue.String() = %v, want %v", value.String(), "\"hello\"")
	}
	
	if value.Type() != "string" {
		t.Errorf("BasicValue.Type() = %v, want %v", value.Type(), "string")
	}
}

func TestNewSymbolTable(t *testing.T) {
	st := NewSymbolTable()
	
	if st == nil {
		t.Fatal("NewSymbolTable should return non-nil symbol table")
	}
	
	if st.currentScope != 0 {
		t.Errorf("Initial scope should be 0, got %v", st.currentScope)
	}
	
	if len(st.scopes) != 1 {
		t.Errorf("Should start with 1 scope, got %v", len(st.scopes))
	}
	
	if st.symbols == nil {
		t.Error("Symbols map should be initialized")
	}
}

func TestSymbolTable_Define(t *testing.T) {
	st := NewSymbolTable()
	value := NewBasicValue("42", "int")
	
	// Test successful definition
	err := st.Define("x", value)
	if err != nil {
		t.Errorf("Define should succeed: %v", err)
	}
	
	// Test retrieving defined symbol
	retrieved, exists := st.Lookup("x")
	if !exists {
		t.Error("Defined symbol should be retrievable")
	}
	
	if retrieved.String() != value.String() {
		t.Errorf("Retrieved value = %v, want %v", retrieved.String(), value.String())
	}
	
	if retrieved.Type() != value.Type() {
		t.Errorf("Retrieved type = %v, want %v", retrieved.Type(), value.Type())
	}
}

func TestSymbolTable_Define_EmptyName(t *testing.T) {
	st := NewSymbolTable()
	value := NewBasicValue("42", "int")
	
	err := st.Define("", value)
	if err == nil {
		t.Error("Define should fail for empty name")
	}
	
	if err.Error() != "symbol name cannot be empty" {
		t.Errorf("Expected specific error message, got: %v", err.Error())
	}
}

func TestSymbolTable_Define_Duplicate(t *testing.T) {
	st := NewSymbolTable()
	value1 := NewBasicValue("42", "int")
	value2 := NewBasicValue("100", "int")
	
	// First definition should succeed
	err := st.Define("x", value1)
	if err != nil {
		t.Errorf("First Define should succeed: %v", err)
	}
	
	// Second definition in same scope should fail
	err = st.Define("x", value2)
	if err == nil {
		t.Error("Define should fail for duplicate name in same scope")
	}
	
	if err.Error() != "symbol 'x' already defined in current scope" {
		t.Errorf("Expected specific error message, got: %v", err.Error())
	}
}

func TestSymbolTable_Lookup_NotFound(t *testing.T) {
	st := NewSymbolTable()
	
	_, exists := st.Lookup("nonexistent")
	if exists {
		t.Error("Lookup should return false for non-existent symbol")
	}
}

func TestSymbolTable_ScopeManagement(t *testing.T) {
	st := NewSymbolTable()
	
	// Define in global scope
	globalValue := NewBasicValue("global", "string")
	st.Define("x", globalValue)
	
	// Enter new scope
	st.EnterScope()
	if st.currentScope != 1 {
		t.Errorf("Current scope should be 1, got %v", st.currentScope)
	}
	
	// Define in inner scope (should shadow global)
	innerValue := NewBasicValue("inner", "string")
	st.Define("x", innerValue)
	
	// Should find inner value
	retrieved, exists := st.Lookup("x")
	if !exists {
		t.Error("Should find symbol in inner scope")
	}
	
	if retrieved.String() != "inner" {
		t.Errorf("Should find inner value, got %v", retrieved.String())
	}
	
	// Exit scope
	st.ExitScope()
	if st.currentScope != 0 {
		t.Errorf("Current scope should be 0 after exit, got %v", st.currentScope)
	}
	
	// Should now find global value
	retrieved, exists = st.Lookup("x")
	if !exists {
		t.Error("Should find symbol in global scope after exit")
	}
	
	if retrieved.String() != "global" {
		t.Errorf("Should find global value after scope exit, got %v", retrieved.String())
	}
}

func TestSymbolTable_EnterScope(t *testing.T) {
	st := NewSymbolTable()
	
	initialScope := st.currentScope
	initialScopeCount := len(st.scopes)
	
	st.EnterScope()
	
	if st.currentScope != initialScope+1 {
		t.Errorf("Current scope should increase by 1, got %v, want %v", st.currentScope, initialScope+1)
	}
	
	if len(st.scopes) != initialScopeCount+1 {
		t.Errorf("Scope count should increase by 1, got %v, want %v", len(st.scopes), initialScopeCount+1)
	}
}

func TestSymbolTable_ExitScope_AtGlobal(t *testing.T) {
	st := NewSymbolTable()
	
	// Try to exit global scope (should not panic or change scope)
	st.ExitScope()
	
	if st.currentScope != 0 {
		t.Errorf("Should remain at global scope, got %v", st.currentScope)
	}
}

func TestSymbolTable_GetCurrentScope(t *testing.T) {
	st := NewSymbolTable()
	
	if st.GetCurrentScope() != 0 {
		t.Errorf("Initial scope should be 0, got %v", st.GetCurrentScope())
	}
	
	st.EnterScope()
	if st.GetCurrentScope() != 1 {
		t.Errorf("After entering scope should be 1, got %v", st.GetCurrentScope())
	}
}

func TestSymbolTable_GetSymbol(t *testing.T) {
	st := NewSymbolTable()
	value := NewBasicValue("42", "int")
	
	// Test non-existent symbol
	_, exists := st.GetSymbol("nonexistent")
	if exists {
		t.Error("GetSymbol should return false for non-existent symbol")
	}
	
	// Define and retrieve symbol
	st.Define("x", value)
	symbol, exists := st.GetSymbol("x")
	if !exists {
		t.Error("GetSymbol should return true for existing symbol")
	}
	
	if symbol.Name != "x" {
		t.Errorf("Symbol name = %v, want %v", symbol.Name, "x")
	}
	
	if symbol.Type != "int" {
		t.Errorf("Symbol type = %v, want %v", symbol.Type, "int")
	}
	
	if symbol.Scope != 0 {
		t.Errorf("Symbol scope = %v, want %v", symbol.Scope, 0)
	}
}

func TestSymbolTable_GetAllSymbols(t *testing.T) {
	st := NewSymbolTable()
	
	// Initially should be empty
	symbols := st.GetAllSymbols()
	if len(symbols) != 0 {
		t.Errorf("Initially should have 0 symbols, got %v", len(symbols))
	}
	
	// Add symbols
	st.Define("x", NewBasicValue("42", "int"))
	st.Define("y", NewBasicValue("hello", "string"))
	
	symbols = st.GetAllSymbols()
	if len(symbols) != 2 {
		t.Errorf("Should have 2 symbols, got %v", len(symbols))
	}
	
	// Check symbols exist
	if _, exists := symbols["x"]; !exists {
		t.Error("Should contain symbol 'x'")
	}
	
	if _, exists := symbols["y"]; !exists {
		t.Error("Should contain symbol 'y'")
	}
}

func TestSymbolTable_GetAllSymbols_MultipleScopes(t *testing.T) {
	st := NewSymbolTable()
	
	// Global scope
	st.Define("global", NewBasicValue("global_val", "string"))
	
	// Inner scope
	st.EnterScope()
	st.Define("inner", NewBasicValue("inner_val", "string"))
	
	symbols := st.GetAllSymbols()
	if len(symbols) != 2 {
		t.Errorf("Should have 2 symbols across scopes, got %v", len(symbols))
	}
	
	// Should contain both symbols
	if _, exists := symbols["global"]; !exists {
		t.Error("Should contain global symbol")
	}
	
	if _, exists := symbols["inner"]; !exists {
		t.Error("Should contain inner symbol")
	}
}

func TestSymbolTable_ScopeCleanup(t *testing.T) {
	st := NewSymbolTable()
	
	// Define in global scope
	st.Define("global", NewBasicValue("global_val", "string"))
	
	// Enter scope and define
	st.EnterScope()
	st.Define("inner", NewBasicValue("inner_val", "string"))
	
	// Both should be accessible
	_, globalExists := st.Lookup("global")
	_, innerExists := st.Lookup("inner")
	
	if !globalExists || !innerExists {
		t.Error("Both symbols should be accessible in inner scope")
	}
	
	// Exit scope
	st.ExitScope()
	
	// Global should still exist, inner should not
	_, globalExists = st.Lookup("global")
	_, innerExists = st.Lookup("inner")
	
	if !globalExists {
		t.Error("Global symbol should still exist after scope exit")
	}
	
	if innerExists {
		t.Error("Inner symbol should not exist after scope exit")
	}
}