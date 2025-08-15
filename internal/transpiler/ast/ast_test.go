package ast

import (
	"strings"
	"testing"

	"github.com/antlr4-go/antlr/v4"
	"github.com/thsfranca/vex/internal/transpiler/parser"
)

func TestNewParser(t *testing.T) {
	parser := NewParser()
	if parser == nil {
		t.Fatal("NewParser should return a non-nil parser")
	}
}

func TestVexParser_Parse(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantErr  bool
		validate func(t *testing.T, ast AST)
	}{
		{
			name:    "Simple expression",
			input:   "(def x 42)",
			wantErr: false,
			validate: func(t *testing.T, ast AST) {
				if ast == nil {
					t.Error("AST should not be nil")
				}
				if ast.Root() == nil {
					t.Error("AST root should not be nil")
				}
			},
		},
		{
			name:    "Function definition",
			input:   "(defn add [x y] (+ x y))",
			wantErr: false,
			validate: func(t *testing.T, ast AST) {
				if ast == nil {
					t.Error("AST should not be nil")
				}
				if ast.Root() == nil {
					t.Error("AST root should not be nil")
				}
			},
		},
		{
			name:    "Multiple expressions",
			input:   "(def x 42)\n(def y 100)",
			wantErr: false,
			validate: func(t *testing.T, ast AST) {
				if ast == nil {
					t.Error("AST should not be nil")
				}
				if ast.Root() == nil {
					t.Error("AST root should not be nil")
				}
			},
		},
		{
			name:    "Array literal",
			input:   "(def arr [1 2 3])",
			wantErr: false,
			validate: func(t *testing.T, ast AST) {
				if ast == nil {
					t.Error("AST should not be nil")
				}
			},
		},
		{
			name:    "String literal",
			input:   "(def greeting \"Hello, World!\")",
			wantErr: false,
			validate: func(t *testing.T, ast AST) {
				if ast == nil {
					t.Error("AST should not be nil")
				}
			},
		},
		{
			name:    "Empty input",
			input:   "",
			wantErr: false,
			validate: func(t *testing.T, ast AST) {
				if ast == nil {
					t.Error("AST should not be nil even for empty input")
				}
			},
		},
		{
			name:    "Nested expressions",
			input:   "(def result (+ (* 2 3) (- 10 5)))",
			wantErr: false,
			validate: func(t *testing.T, ast AST) {
				if ast == nil {
					t.Error("AST should not be nil")
				}
			},
		},
	}

	parser := NewParser()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ast, err := parser.Parse(tt.input)

			if (err != nil) != tt.wantErr {
				t.Errorf("VexParser.Parse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.validate != nil {
				tt.validate(t, ast)
			}
		})
	}
}

func TestVexParser_ParseFile(t *testing.T) {
	parser := NewParser()

	// Test parsing existing example file
	ast, err := parser.ParseFile("../../../examples/valid/hello-world.vx")
	if err != nil {
		t.Errorf("ParseFile should parse hello-world.vx successfully: %v", err)
		return
	}

	if ast == nil {
		t.Error("AST should not be nil")
		return
	}

	if ast.Root() == nil {
		t.Error("AST root should not be nil")
	}
}

func TestVexParser_ParseFile_NonExistent(t *testing.T) {
	parser := NewParser()

	// Test parsing non-existent file
	_, err := parser.ParseFile("non-existent-file.vx")
	if err == nil {
		t.Error("ParseFile should return error for non-existent file")
	}
}

func TestNewVexAST(t *testing.T) {
	// Create a mock AST node
	inputStream := antlr.NewInputStream("(def x 42)")
	lexer := parser.NewVexLexer(inputStream)
	tokenStream := antlr.NewCommonTokenStream(lexer, 0)
	vexParser := parser.NewVexParser(tokenStream)
	tree := vexParser.Program()

	ast := NewVexAST(tree)
	if ast == nil {
		t.Fatal("NewVexAST should return non-nil AST")
	}

	if ast.Root() != tree {
		t.Error("AST root should be the provided tree")
	}
}

func TestVexAST_Accept(t *testing.T) {
	// Create a test AST
	inputStream := antlr.NewInputStream("(def x 42)")
	lexer := parser.NewVexLexer(inputStream)
	tokenStream := antlr.NewCommonTokenStream(lexer, 0)
	vexParser := parser.NewVexParser(tokenStream)
	tree := vexParser.Program()

	ast := NewVexAST(tree)

	// Create a mock visitor
	visitor := &MockASTVisitor{}

	err := ast.Accept(visitor)
	if err != nil {
		t.Errorf("Accept should not return error: %v", err)
	}

	if !visitor.visitProgramCalled {
		t.Error("Visitor's VisitProgram should have been called")
	}
}

func TestVexAST_Accept_InvalidRoot(t *testing.T) {
	// Create AST with invalid root type
	ast := NewVexAST(&mockInvalidTree{})

	visitor := &MockASTVisitor{}
	err := ast.Accept(visitor)

	if err == nil {
		t.Error("Accept should return error for invalid root type")
	}

	if !strings.Contains(err.Error(), "invalid AST root type") {
		t.Errorf("Error should mention invalid root type, got: %v", err)
	}
}







// Mock implementations for testing

type MockASTVisitor struct {
	visitProgramCalled  bool
	visitListCalled     bool
	visitArrayCalled    bool
	visitTerminalCalled bool
}

func (m *MockASTVisitor) VisitProgram(ctx *parser.ProgramContext) error {
	m.visitProgramCalled = true
	return nil
}

func (m *MockASTVisitor) VisitList(ctx *parser.ListContext) (Value, error) {
	m.visitListCalled = true
	return &BasicValue{value: "list", typ: "list"}, nil
}

func (m *MockASTVisitor) VisitArray(ctx *parser.ArrayContext) (Value, error) {
	m.visitArrayCalled = true
	return &BasicValue{value: "array", typ: "array"}, nil
}

func (m *MockASTVisitor) VisitTerminal(node antlr.TerminalNode) (Value, error) {
	m.visitTerminalCalled = true
	return &BasicValue{value: node.GetText(), typ: "terminal"}, nil
}

type mockInvalidTree struct{}

func (m *mockInvalidTree) GetChild(i int) antlr.Tree                                      { return nil }
func (m *mockInvalidTree) GetChildCount() int                                             { return 0 }
func (m *mockInvalidTree) GetChildren() []antlr.Tree                                      { return nil }
func (m *mockInvalidTree) GetParent() antlr.Tree                                          { return nil }
func (m *mockInvalidTree) GetPayload() interface{}                                        { return nil }
func (m *mockInvalidTree) GetSourceInterval() antlr.Interval                              { return antlr.Interval{} }
func (m *mockInvalidTree) SetParent(parent antlr.Tree)                                    {}
func (m *mockInvalidTree) ToStringTree(ruleNames []string, recog antlr.Recognizer) string { return "" }
func (m *mockInvalidTree) Accept(visitor antlr.ParseTreeVisitor) interface{}              { return nil }
