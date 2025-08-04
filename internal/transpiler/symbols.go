// Package transpiler provides the symbol system for Vex
package transpiler

import (
	"fmt"
	"sync"
)

// Symbol represents an interned symbol in Vex
type Symbol struct {
	Name string
	ID   int
}

// String returns the string representation of the symbol
func (s *Symbol) String() string {
	return s.Name
}

// Equals checks if two symbols are equal (by ID for interned symbols)
func (s *Symbol) Equals(other *Symbol) bool {
	return s.ID == other.ID
}

// SymbolTable manages symbol interning and lookup
type SymbolTable struct {
	mu          sync.RWMutex
	symbols     map[string]*Symbol // name -> symbol mapping
	symbolsByID map[int]*Symbol    // ID -> symbol mapping
	nextID      int
}

// NewSymbolTable creates a new symbol table
func NewSymbolTable() *SymbolTable {
	return &SymbolTable{
		symbols:     make(map[string]*Symbol, 256), // Pre-allocate for common symbols
		symbolsByID: make(map[int]*Symbol, 256),
		nextID:      1, // Start from 1, reserve 0 for special cases
	}
}

// Intern interns a symbol name, returning an existing symbol if it exists
func (st *SymbolTable) Intern(name string) *Symbol {
	st.mu.RLock()
	if existing, exists := st.symbols[name]; exists {
		st.mu.RUnlock()
		return existing
	}
	st.mu.RUnlock()

	st.mu.Lock()
	defer st.mu.Unlock()

	// Double-check after acquiring write lock
	if existing, exists := st.symbols[name]; exists {
		return existing
	}

	// Create new symbol
	symbol := &Symbol{
		Name: name,
		ID:   st.nextID,
	}
	st.nextID++

	st.symbols[name] = symbol
	st.symbolsByID[symbol.ID] = symbol

	return symbol
}

// Lookup looks up a symbol by name without interning
func (st *SymbolTable) Lookup(name string) (*Symbol, bool) {
	st.mu.RLock()
	defer st.mu.RUnlock()

	symbol, exists := st.symbols[name]
	return symbol, exists
}

// GetByID retrieves a symbol by its ID
func (st *SymbolTable) GetByID(id int) (*Symbol, bool) {
	st.mu.RLock()
	defer st.mu.RUnlock()

	symbol, exists := st.symbolsByID[id]
	return symbol, exists
}

// Size returns the number of interned symbols
func (st *SymbolTable) Size() int {
	st.mu.RLock()
	defer st.mu.RUnlock()
	return len(st.symbols)
}

// AllSymbols returns all interned symbols (for debugging/introspection)
func (st *SymbolTable) AllSymbols() []*Symbol {
	st.mu.RLock()
	defer st.mu.RUnlock()

	symbols := make([]*Symbol, 0, len(st.symbols))
	for _, symbol := range st.symbols {
		symbols = append(symbols, symbol)
	}
	return symbols
}

// Namespace represents a namespace for symbol resolution
type Namespace struct {
	Name        string
	Parent      *Namespace
	bindings    map[string]*Binding
	symbolTable *SymbolTable
	mu          sync.RWMutex
}

// Binding represents a binding of a symbol to a type and value information
type Binding struct {
	Symbol     *Symbol
	Type       VexType
	IsMutable  bool
	IsFunction bool
	Namespace  *Namespace
}

// NewNamespace creates a new namespace
func NewNamespace(name string, parent *Namespace, symbolTable *SymbolTable) *Namespace {
	return &Namespace{
		Name:        name,
		Parent:      parent,
		bindings:    make(map[string]*Binding, 64), // Pre-allocate for common bindings
		symbolTable: symbolTable,
	}
}

// Bind binds a symbol to a type in this namespace
func (ns *Namespace) Bind(name string, vexType VexType, isMutable bool) *Binding {
	ns.mu.Lock()
	defer ns.mu.Unlock()

	symbol := ns.symbolTable.Intern(name)
	binding := &Binding{
		Symbol:    symbol,
		Type:      vexType,
		IsMutable: isMutable,
		Namespace: ns,
	}

	ns.bindings[name] = binding
	return binding
}

// BindFunction binds a function symbol to a function type
func (ns *Namespace) BindFunction(name string, funcType *FunctionType) *Binding {
	ns.mu.Lock()
	defer ns.mu.Unlock()

	symbol := ns.symbolTable.Intern(name)
	binding := &Binding{
		Symbol:     symbol,
		Type:       funcType,
		IsMutable:  false, // Functions are immutable
		IsFunction: true,
		Namespace:  ns,
	}

	ns.bindings[name] = binding
	return binding
}

// Resolve resolves a symbol in this namespace, walking up the parent chain
func (ns *Namespace) Resolve(name string) (*Binding, bool) {
	current := ns
	for current != nil {
		current.mu.RLock()
		if binding, exists := current.bindings[name]; exists {
			current.mu.RUnlock()
			return binding, true
		}
		current.mu.RUnlock()
		current = current.Parent
	}
	return nil, false
}

// LocalResolve resolves a symbol only in this namespace (not parent)
func (ns *Namespace) LocalResolve(name string) (*Binding, bool) {
	ns.mu.RLock()
	defer ns.mu.RUnlock()

	binding, exists := ns.bindings[name]
	return binding, exists
}

// IsBound checks if a symbol is bound in this namespace or parents
func (ns *Namespace) IsBound(name string) bool {
	_, exists := ns.Resolve(name)
	return exists
}

// GetLocalBindings returns all bindings in this namespace (not parents)
func (ns *Namespace) GetLocalBindings() map[string]*Binding {
	ns.mu.RLock()
	defer ns.mu.RUnlock()

	// Return a copy to avoid concurrent access issues
	result := make(map[string]*Binding, len(ns.bindings))
	for name, binding := range ns.bindings {
		result[name] = binding
	}
	return result
}

// NamespaceManager manages multiple namespaces and provides namespace-qualified resolution
type NamespaceManager struct {
	globalSymbolTable *SymbolTable
	namespaces        map[string]*Namespace
	currentNamespace  *Namespace
	mu                sync.RWMutex
}

// NewNamespaceManager creates a new namespace manager with a global namespace
func NewNamespaceManager() *NamespaceManager {
	symbolTable := NewSymbolTable()
	globalNS := NewNamespace("global", nil, symbolTable)

	nm := &NamespaceManager{
		globalSymbolTable: symbolTable,
		namespaces:        make(map[string]*Namespace, 16),
		currentNamespace:  globalNS,
	}

	nm.namespaces["global"] = globalNS

	// Bind built-in symbols and types
	nm.bindBuiltins(globalNS)

	return nm
}

// bindBuiltins binds built-in symbols and functions to the global namespace
func (nm *NamespaceManager) bindBuiltins(globalNS *Namespace) {
	// Built-in type symbols
	globalNS.Bind("int", IntType, false)
	globalNS.Bind("float", FloatType, false)
	globalNS.Bind("string", StringType, false)
	globalNS.Bind("bool", BoolType, false)
	globalNS.Bind("symbol", SymbolType, false)

	// Built-in functions (arithmetic operators)
	globalNS.BindFunction("+", NewFunctionType([]VexType{IntType, IntType}, IntType))
	globalNS.BindFunction("-", NewFunctionType([]VexType{IntType, IntType}, IntType))
	globalNS.BindFunction("*", NewFunctionType([]VexType{IntType, IntType}, IntType))
	globalNS.BindFunction("/", NewFunctionType([]VexType{IntType, IntType}, IntType))

	// Built-in boolean values
	globalNS.Bind("true", BoolType, false)
	globalNS.Bind("false", BoolType, false)
}

// CreateNamespace creates a new namespace
func (nm *NamespaceManager) CreateNamespace(name string, parent *Namespace) *Namespace {
	nm.mu.Lock()
	defer nm.mu.Unlock()

	if parent == nil {
		parent = nm.namespaces["global"]
	}

	ns := NewNamespace(name, parent, nm.globalSymbolTable)
	nm.namespaces[name] = ns
	return ns
}

// GetNamespace retrieves a namespace by name
func (nm *NamespaceManager) GetNamespace(name string) (*Namespace, bool) {
	nm.mu.RLock()
	defer nm.mu.RUnlock()

	ns, exists := nm.namespaces[name]
	return ns, exists
}

// SetCurrentNamespace sets the current working namespace
func (nm *NamespaceManager) SetCurrentNamespace(name string) error {
	nm.mu.RLock()
	defer nm.mu.RUnlock()

	ns, exists := nm.namespaces[name]
	if !exists {
		return fmt.Errorf("namespace %s not found", name)
	}

	nm.currentNamespace = ns
	return nil
}

// GetCurrentNamespace returns the current working namespace
func (nm *NamespaceManager) GetCurrentNamespace() *Namespace {
	nm.mu.RLock()
	defer nm.mu.RUnlock()
	return nm.currentNamespace
}

// ResolveQualified resolves a qualified symbol (namespace/symbol)
func (nm *NamespaceManager) ResolveQualified(qualifiedName string) (*Binding, error) {
	nm.mu.RLock()
	defer nm.mu.RUnlock()

	// Parse qualified name (e.g., "http/Server" -> namespace="http", name="Server")
	parts := parseQualifiedName(qualifiedName)
	if len(parts) == 1 {
		// Unqualified name, resolve in current namespace
		binding, exists := nm.currentNamespace.Resolve(parts[0])
		if !exists {
			return nil, fmt.Errorf("symbol %s not found", parts[0])
		}
		return binding, nil
	}

	if len(parts) == 2 {
		// Qualified name
		nsName, symbolName := parts[0], parts[1]
		ns, exists := nm.namespaces[nsName]
		if !exists {
			return nil, fmt.Errorf("namespace %s not found", nsName)
		}

		binding, exists := ns.Resolve(symbolName)
		if !exists {
			return nil, fmt.Errorf("symbol %s not found in namespace %s", symbolName, nsName)
		}
		return binding, nil
	}

	return nil, fmt.Errorf("invalid qualified name: %s", qualifiedName)
}

// parseQualifiedName parses a qualified name into its components
func parseQualifiedName(qualified string) []string {
	// Simple implementation: split on '/'
	// Could be enhanced for more complex namespace resolution
	result := make([]string, 0, 2)
	slashIndex := -1

	for i, r := range qualified {
		if r == '/' {
			slashIndex = i
			break
		}
	}

	if slashIndex == -1 {
		result = append(result, qualified)
	} else {
		result = append(result, qualified[:slashIndex])
		result = append(result, qualified[slashIndex+1:])
	}

	return result
}

// GetGlobalSymbolTable returns the global symbol table
func (nm *NamespaceManager) GetGlobalSymbolTable() *SymbolTable {
	return nm.globalSymbolTable
}
