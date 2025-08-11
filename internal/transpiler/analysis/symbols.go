package analysis

import (
	"fmt"
)

// Symbol represents a defined symbol (variable, function, etc.)
type Symbol struct {
	Name  string
	Type  string
	Value Value
	Scope int
}

// Value represents a value in the Vex language
type Value interface {
	String() string
	Type() string
}

// BasicValue implements Value for basic types
type BasicValue struct {
	value string
	typ   string
    ty    Type
    raw   bool
}

// NewBasicValue constructs a simple Value carrying a printable form and outward type.
func NewBasicValue(value, typ string) *BasicValue { return &BasicValue{value: value, typ: typ} }

func (v *BasicValue) String() string {
	return v.value
}

func (v *BasicValue) Type() string { return v.typ }

// WithType attaches internal type information without changing the public type string.
func (v *BasicValue) WithType(t Type) *BasicValue { v.ty = t; return v }

// getType returns the internal type if present; otherwise nil.
func (v *BasicValue) getType() Type { return v.ty }

// MarkRaw marks this value as originating from a raw macro/body fragment.
func (v *BasicValue) MarkRaw() *BasicValue { v.raw = true; return v }

// isRaw reports whether this value is raw.
func (v *BasicValue) isRaw() bool { return v.raw }

// RecordValue represents a record type declaration
type RecordValue struct {
    name       string
    fields     map[string]string
    fieldOrder []string
}

// NewRecordValue creates a record declaration with a field map and order.
func NewRecordValue(name string, fields map[string]string, order []string) *RecordValue {
    return &RecordValue{name: name, fields: fields, fieldOrder: order}
}

func (r *RecordValue) String() string { return r.name }
func (r *RecordValue) Type() string   { return "record" }

func (r *RecordValue) GetFields() map[string]string {
    cp := make(map[string]string, len(r.fields))
    for k, v := range r.fields { cp[k] = v }
    return cp
}

func (r *RecordValue) GetFieldOrder() []string {
    out := make([]string, len(r.fieldOrder))
    copy(out, r.fieldOrder)
    return out
}

// SymbolTableImpl implements the SymbolTable interface
type SymbolTableImpl struct {
	symbols     map[string]*Symbol
	scopes      []map[string]*Symbol
	currentScope int
}

// NewSymbolTable creates a new symbol table
// NewSymbolTable initializes a scoped symbol table with a global scope.
func NewSymbolTable() *SymbolTableImpl {
	st := &SymbolTableImpl{
		symbols:      make(map[string]*Symbol),
		scopes:       make([]map[string]*Symbol, 0),
		currentScope: 0,
	}
	
	// Add global scope
	st.scopes = append(st.scopes, make(map[string]*Symbol))
	
	return st
}

// Define adds a new symbol to the current scope
func (st *SymbolTableImpl) Define(name string, value Value) error {
	if name == "" {
		return fmt.Errorf("symbol name cannot be empty")
	}
	
	// Check if already defined in current scope
	currentScopeSymbols := st.scopes[st.currentScope]
	if _, exists := currentScopeSymbols[name]; exists {
		return fmt.Errorf("symbol '%s' already defined in current scope", name)
	}
	
	// Create new symbol
	symbol := &Symbol{
		Name:  name,
		Type:  value.Type(),
		Value: value,
		Scope: st.currentScope,
	}
	
	// Add to current scope and global table
	currentScopeSymbols[name] = symbol
	st.symbols[name] = symbol
	
	return nil
}

// Lookup finds a symbol by name, searching from current scope upward
func (st *SymbolTableImpl) Lookup(name string) (Value, bool) {
	// Search from current scope back to global scope
	for i := st.currentScope; i >= 0; i-- {
		if symbol, exists := st.scopes[i][name]; exists {
			return symbol.Value, true
		}
	}
	
	return nil, false
}

// EnterScope creates a new scope level
func (st *SymbolTableImpl) EnterScope() {
	st.currentScope++
	st.scopes = append(st.scopes, make(map[string]*Symbol))
}

// ExitScope returns to the previous scope level
func (st *SymbolTableImpl) ExitScope() {
	if st.currentScope > 0 {
		// Remove symbols from global table that were defined in this scope
		scopeToRemove := st.scopes[st.currentScope]
		for name := range scopeToRemove {
			delete(st.symbols, name)
		}
		
		// Remove the scope
		st.scopes = st.scopes[:len(st.scopes)-1]
		st.currentScope--
	}
}

// GetCurrentScope returns the current scope level
func (st *SymbolTableImpl) GetCurrentScope() int {
	return st.currentScope
}

// GetSymbol retrieves symbol information by name
func (st *SymbolTableImpl) GetSymbol(name string) (*Symbol, bool) {
	// Search from current scope back to global scope
	for i := st.currentScope; i >= 0; i-- {
		if symbol, exists := st.scopes[i][name]; exists {
			return symbol, true
		}
	}
	
	return nil, false
}

// GetAllSymbols returns all symbols in the current scope
func (st *SymbolTableImpl) GetAllSymbols() map[string]*Symbol {
	result := make(map[string]*Symbol)
	
	// Collect all symbols from global to current scope
	for i := 0; i <= st.currentScope; i++ {
		for name, symbol := range st.scopes[i] {
			result[name] = symbol
		}
	}
	
	return result
}