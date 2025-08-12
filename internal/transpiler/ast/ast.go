package ast

import (
	"fmt"
	"os"

	"github.com/antlr4-go/antlr/v4"
	"github.com/thsfranca/vex/internal/transpiler/parser"
)

// VexAST implements the AST interface
type VexAST struct {
	root antlr.Tree
}

// NewVexAST creates a new VexAST
func NewVexAST(root antlr.Tree) *VexAST {
	return &VexAST{root: root}
}

// Root returns the root node of the AST
func (ast *VexAST) Root() antlr.Tree {
	return ast.root
}

// Accept applies a visitor to the AST
func (ast *VexAST) Accept(visitor ASTVisitor) error {
	if programCtx, ok := ast.root.(*parser.ProgramContext); ok {
		return visitor.VisitProgram(programCtx)
	}
	return fmt.Errorf("invalid AST root type")
}

// ASTVisitor interface for visiting AST nodes
type ASTVisitor interface {
	VisitProgram(ctx *parser.ProgramContext) error
	VisitList(ctx *parser.ListContext) (Value, error)
	VisitArray(ctx *parser.ArrayContext) (Value, error)
	VisitTerminal(node antlr.TerminalNode) (Value, error)
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
}

func NewBasicValue(value, typ string) *BasicValue {
	return &BasicValue{value: value, typ: typ}
}

func (v *BasicValue) String() string {
	return v.value
}

func (v *BasicValue) Type() string {
	return v.typ
}

// VexParser implements the Parser interface
type VexParser struct{}

// NewParser creates a new VexParser
func NewParser() *VexParser {
	return &VexParser{}
}

// Parse parses source code into an AST
func (p *VexParser) Parse(input string) (AST, error) {
	inputStream := antlr.NewInputStream(input)
	lexer := parser.NewVexLexer(inputStream)
	tokenStream := antlr.NewCommonTokenStream(lexer, 0)
	vexParser := parser.NewVexParser(tokenStream)

	// Parse the program
	tree := vexParser.Program()
	if tree == nil {
		return nil, fmt.Errorf("failed to parse input")
	}

	return NewVexAST(tree.(*parser.ProgramContext)), nil
}

// ParseFile parses a file into an AST
func (p *VexParser) ParseFile(filename string) (AST, error) {
	content, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %v", filename, err)
	}

	return p.Parse(string(content))
}

// AST interface
type AST interface {
	Root() antlr.Tree
	Accept(visitor ASTVisitor) error
}
