package transpiler

import (
	"fmt"

	"github.com/antlr4-go/antlr/v4"
	"github.com/thsfranca/vex/internal/transpiler/analysis"
	"github.com/thsfranca/vex/internal/transpiler/macro"
	"github.com/thsfranca/vex/internal/transpiler/parser"
)

// Value represents a value in the Vex language
type Value interface {
	String() string
	Type() string
}

// AST represents an abstract syntax tree
type AST interface {
	Root() antlr.Tree
	Accept(visitor ASTVisitor) error
}

// ConcreteAST is a concrete implementation that bridges different AST types
type ConcreteAST struct {
	root antlr.Tree
}

// NewConcreteAST wraps an ANTLR root node into a transpiler-compatible AST.
func NewConcreteAST(root antlr.Tree) *ConcreteAST {
	return &ConcreteAST{root: root}
}

func (ast *ConcreteAST) Root() antlr.Tree {
	return ast.root
}

func (ast *ConcreteAST) Accept(visitor ASTVisitor) error {
	if programCtx, ok := ast.root.(*parser.ProgramContext); ok {
		return visitor.VisitProgram(programCtx)
	}
	return fmt.Errorf("invalid AST root type")
}

// ASTVisitor defines the interface for visiting AST nodes
type ASTVisitor interface {
	VisitProgram(ctx *parser.ProgramContext) error
	VisitList(ctx *parser.ListContext) (Value, error)
	VisitArray(ctx *parser.ArrayContext) (Value, error)
	VisitTerminal(node antlr.TerminalNode) (Value, error)
}

// MacroExpander handles macro expansion
type MacroExpander interface {
	ExpandMacros(ast AST) (AST, error)
	RegisterMacro(name string, macro *macro.Macro) error
	HasMacro(name string) bool
	GetMacro(name string) (*macro.Macro, bool)
}

// SymbolTable manages variable and function definitions
type SymbolTable interface {
	Define(name string, value Value) error
	Lookup(name string) (Value, bool)
	EnterScope() 
	ExitScope()
}

// ErrorReporter handles error reporting and collection
type ErrorReporter interface {
	ReportError(line, column int, message string)
	ReportWarning(line, column int, message string)
	HasErrors() bool
	GetErrors() []Error
}

// CodeGenerator generates target code from AST
type CodeGenerator interface {
	Generate(ast AST, symbols SymbolTable) (string, error)
	AddImport(importPath string)
	SetPackageName(name string)
}

// Parser handles parsing of source code into AST
type Parser interface {
	Parse(input string) (AST, error)
	ParseFile(filename string) (AST, error)
}

// Analyzer performs semantic analysis
type Analyzer interface {
	Analyze(ast AST) (SymbolTable, error)
	SetErrorReporter(reporter ErrorReporter)
    // SetPackageEnv informs the analyzer about local Vex packages, their exports,
    // and provided type schemes to enable package-boundary typing.
    SetPackageEnv(ignore map[string]bool, exports map[string]map[string]bool, schemes map[string]map[string]*analysis.TypeScheme)
}

// Error represents a compilation error
type Error struct {
	Line    int
	Column  int
	Message string
	Type    ErrorType
}

// ErrorType classifies compiler errors reported by the transpiler.
type ErrorType int

const (
	SyntaxError ErrorType = iota
	SemanticError
	MacroError
	TypeError
)

// TranspilerConfig holds configuration for the transpiler
type TranspilerConfig struct {
	EnableMacros     bool
	CoreMacroPath    string
	PackageName      string
	GenerateComments bool
    IgnoreImports    map[string]bool
    Exports          map[string]map[string]bool
    PkgSchemes       map[string]map[string]*analysis.TypeScheme
}