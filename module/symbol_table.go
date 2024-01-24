package module

type SymbolTable interface {
	Insert(Symbol)
	Lookup(string) Symbol
}

type Symbol struct {
	Name  string
	Value interface{}
}

type symbolTable struct {
	Table map[string]Symbol
}

func NewSymbolTable() SymbolTable {
	return &symbolTable{Table: make(map[string]Symbol)}
}

func (s *symbolTable) Insert(symbol Symbol) {
	s.Table[symbol.Name] = symbol
}

func (s *symbolTable) Lookup(name string) Symbol {
	return s.Table[name]
}
