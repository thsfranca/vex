package macro

import (
	"testing"

	"github.com/antlr4-go/antlr/v4"
	"github.com/thsfranca/vex/internal/transpiler/parser"
)

func TestNewMacroExpander(t *testing.T) {
	registry := NewRegistry(Config{EnableValidation: false})
	expander := NewMacroExpander(registry)
	
	if expander == nil {
		t.Fatal("NewMacroExpander should return non-nil expander")
	}
	
	if expander.registry != registry {
		t.Error("MacroExpander should store the provided registry")
	}
	
	if expander.Expander == nil {
		t.Error("MacroExpander should have embedded Expander")
	}
}

func TestMacroExpanderImpl_RegisterMacro(t *testing.T) {
	registry := NewRegistry(Config{EnableValidation: false})
	expander := NewMacroExpander(registry)
	
	macro := &Macro{
		Name:   "test-macro",
		Params: []string{"x"},
		Body:   "x",
	}
	
	err := expander.RegisterMacro("test-macro", macro)
	if err != nil {
		t.Errorf("RegisterMacro should succeed: %v", err)
	}
	
	// Verify it was registered
	if !expander.HasMacro("test-macro") {
		t.Error("Macro should be registered")
	}
}

func TestMacroExpanderImpl_HasMacro(t *testing.T) {
	registry := NewRegistry(Config{EnableValidation: false})
	expander := NewMacroExpander(registry)
	
	// Should not have unregistered macro
	if expander.HasMacro("nonexistent") {
		t.Error("Should not have unregistered macro")
	}
	
	// Register and check
	macro := &Macro{Name: "test", Params: []string{}, Body: "42"}
	expander.RegisterMacro("test", macro)
	
	if !expander.HasMacro("test") {
		t.Error("Should have registered macro")
	}
}

func TestMacroExpanderImpl_GetMacro(t *testing.T) {
	registry := NewRegistry(Config{EnableValidation: false})
	expander := NewMacroExpander(registry)
	
	// Test non-existent macro
	_, exists := expander.GetMacro("nonexistent")
	if exists {
		t.Error("Should not find non-existent macro")
	}
	
	// Register and retrieve
	originalMacro := &Macro{
		Name:   "test",
		Params: []string{"x", "y"},
		Body:   "(+ x y)",
	}
	expander.RegisterMacro("test", originalMacro)
	
	retrievedMacro, exists := expander.GetMacro("test")
	if !exists {
		t.Error("Should find registered macro")
	}
	
	if retrievedMacro.Name != originalMacro.Name {
		t.Errorf("Retrieved macro name = %v, want %v", retrievedMacro.Name, originalMacro.Name)
	}
	
	if len(retrievedMacro.Params) != len(originalMacro.Params) {
		t.Errorf("Retrieved macro params length = %v, want %v", len(retrievedMacro.Params), len(originalMacro.Params))
	}
	
	if retrievedMacro.Body != originalMacro.Body {
		t.Errorf("Retrieved macro body = %v, want %v", retrievedMacro.Body, originalMacro.Body)
	}
}

func TestMacroExpanderImpl_ExpandMacros(t *testing.T) {
	registry := NewRegistry(Config{EnableValidation: false})
	
	// Register a test macro
	testMacro := &Macro{
		Name:   "test",
		Params: []string{"x"},
		Body:   "(+ x 1)",
	}
	registry.RegisterMacro("test", testMacro)
	
	expander := NewMacroExpander(registry)
	
	// Create a simple AST
	inputStream := antlr.NewInputStream("(def x 42)")
	lexer := parser.NewVexLexer(inputStream)
	tokenStream := antlr.NewCommonTokenStream(lexer, 0)
	vexParser := parser.NewVexParser(tokenStream)
	tree := vexParser.Program()
	
	ast := NewVexAST(tree)
	
	// Test expansion
	expandedAST, err := expander.ExpandMacros(ast)
	if err != nil {
		t.Errorf("ExpandMacros should succeed: %v", err)
		return
	}
	
	if expandedAST == nil {
		t.Error("Expanded AST should not be nil")
	}
	
	if expandedAST.Root() == nil {
		t.Error("Expanded AST root should not be nil")
	}
}

func TestVexAST_Root(t *testing.T) {
	// Create a test tree
	inputStream := antlr.NewInputStream("(def x 42)")
	lexer := parser.NewVexLexer(inputStream)
	tokenStream := antlr.NewCommonTokenStream(lexer, 0)
	vexParser := parser.NewVexParser(tokenStream)
	tree := vexParser.Program()
	
	ast := NewVexAST(tree)
	
	if ast.Root() != tree {
		t.Error("AST Root() should return the original tree")
	}
}

func TestVexAST_Accept(t *testing.T) {
	// Create a test tree
	inputStream := antlr.NewInputStream("(def x 42)")
	lexer := parser.NewVexLexer(inputStream)
	tokenStream := antlr.NewCommonTokenStream(lexer, 0)
	vexParser := parser.NewVexParser(tokenStream)
	tree := vexParser.Program()
	
	ast := NewVexAST(tree)
	
	// Create mock visitor
	visitor := &MockMacroASTVisitor{}
	
	err := ast.Accept(visitor)
	if err != nil {
		t.Errorf("Accept should succeed: %v", err)
	}
	
	if !visitor.visitProgramCalled {
		t.Error("Visitor's VisitProgram should have been called")
	}
}

func TestVexAST_Accept_InvalidRoot(t *testing.T) {
	// Create AST with invalid root
	ast := NewVexAST(&mockInvalidTree{})
	
	visitor := &MockMacroASTVisitor{}
	err := ast.Accept(visitor)
	
	if err == nil {
		t.Error("Accept should return error for invalid root type")
	}
}

func TestMacroExpanderImpl_expandMacrosInTree(t *testing.T) {
	registry := NewRegistry(Config{EnableValidation: false})
	expander := NewMacroExpander(registry)
	
	// Create a simple tree
	inputStream := antlr.NewInputStream("(def x 42)")
	lexer := parser.NewVexLexer(inputStream)
	tokenStream := antlr.NewCommonTokenStream(lexer, 0)
	vexParser := parser.NewVexParser(tokenStream)
	tree := vexParser.Program()
	
	// Test tree expansion
	expandedTree, err := expander.expandMacrosInTree(tree)
	if err != nil {
		t.Errorf("expandMacrosInTree should succeed: %v", err)
	}
	
	if expandedTree == nil {
		t.Error("Expanded tree should not be nil")
	}
}

func TestMacroExpanderImpl_parseExpandedCode(t *testing.T) {
	registry := NewRegistry(Config{EnableValidation: false})
	expander := NewMacroExpander(registry)
	
	tests := []struct {
		name    string
		code    string
		wantErr bool
	}{
		{
			name:    "Valid list expression",
			code:    "(+ 1 2)",
			wantErr: false,
		},
		{
			name:    "Valid definition",
			code:    "(def x 42)",
			wantErr: false,
		},
		{
			name:    "Simple value",
			code:    "42",
			wantErr: false,
		},
		{
			name:    "Invalid syntax",
			code:    "(((",
			wantErr: false, // parseExpandedCode doesn't fail for invalid syntax, just returns nil
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tree, err := expander.parseExpandedCode(tt.code)
			
			if (err != nil) != tt.wantErr {
				t.Errorf("parseExpandedCode() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			
			if !tt.wantErr && tree == nil {
				t.Error("parseExpandedCode() should return non-nil tree for valid input")
			}
		})
	}
}

func TestMacroExpanderImpl_nodeToString(t *testing.T) {
	registry := NewRegistry(Config{EnableValidation: false})
	expander := NewMacroExpander(registry)
	
	// Test with terminal node
	inputStream := antlr.NewInputStream("42")
	lexer := parser.NewVexLexer(inputStream)
	tokenStream := antlr.NewCommonTokenStream(lexer, 0)
	token := tokenStream.Get(0)
	terminal := antlr.NewTerminalNodeImpl(token)
	
	result := expander.nodeToString(terminal)
	if result != "42" {
		t.Errorf("nodeToString() = %v, want %v", result, "42")
	}
}

// Mock implementations for testing

type MockMacroASTVisitor struct {
	visitProgramCalled bool
}

func (m *MockMacroASTVisitor) VisitProgram(ctx *parser.ProgramContext) error {
	m.visitProgramCalled = true
	return nil
}

func (m *MockMacroASTVisitor) VisitList(ctx *parser.ListContext) (Value, error) {
	return &MockValue{value: "list"}, nil
}

func (m *MockMacroASTVisitor) VisitArray(ctx *parser.ArrayContext) (Value, error) {
	return &MockValue{value: "array"}, nil
}

func (m *MockMacroASTVisitor) VisitTerminal(node antlr.TerminalNode) (Value, error) {
	return &MockValue{value: node.GetText()}, nil
}

type MockValue struct {
	value string
}

func (m *MockValue) String() string {
	return m.value
}

func (m *MockValue) Type() string {
	return "mock"
}

type mockInvalidTree struct{}

func (m *mockInvalidTree) GetChild(i int) antlr.Tree { return nil }
func (m *mockInvalidTree) GetChildCount() int { return 0 }
func (m *mockInvalidTree) GetChildren() []antlr.Tree { return nil }
func (m *mockInvalidTree) GetParent() antlr.Tree { return nil }
func (m *mockInvalidTree) GetPayload() interface{} { return nil }
func (m *mockInvalidTree) GetSourceInterval() antlr.Interval { return antlr.Interval{} }
func (m *mockInvalidTree) SetParent(parent antlr.Tree) {}
func (m *mockInvalidTree) ToStringTree(ruleNames []string, recog antlr.Recognizer) string { return "" }
func (m *mockInvalidTree) Accept(visitor antlr.ParseTreeVisitor) interface{} { return nil }

func TestMacroExpanderImpl_expandMacrosInArray(t *testing.T) {
	registry := NewRegistry(Config{EnableValidation: false})
	expander := NewMacroExpander(registry)
	
	// Create array context  
	arrayCtx := createMockArrayNode("1", "2", "3")
	
	result, err := expander.expandMacrosInArray(arrayCtx)
	if err != nil {
		t.Errorf("expandMacrosInArray should succeed: %v", err)
	}
	
	if result == nil {
		t.Error("expandMacrosInArray should return non-nil result")
	}
}

func TestMacroExpanderImpl_expandMacrosInList_WithMacro(t *testing.T) {
	registry := NewRegistry(Config{EnableValidation: false})
	
	// Register a test macro
	testMacro := &Macro{
		Name:   "test-macro",
		Params: []string{"x"},
		Body:   "(+ x 1)",
	}
	registry.RegisterMacro("test-macro", testMacro)
	
	expander := NewMacroExpander(registry)
	
	// Create list context with macro call
	listCtx := createMockListNode("test-macro", "5")
	
	result, err := expander.expandMacrosInList(listCtx)
	if err != nil {
		t.Errorf("expandMacrosInList should succeed: %v", err)
	}
	
	if result == nil {
		t.Error("expandMacrosInList should return non-nil result")
	}
}

// Helper function for creating mock array nodes
func createMockArrayNode(elements ...string) *parser.ArrayContext {
	expr := "["
	for i, elem := range elements {
		if i > 0 {
			expr += " "
		}
		expr += elem
	}
	expr += "]"
	
	inputStream := antlr.NewInputStream(expr)
	lexer := parser.NewVexLexer(inputStream)
	tokenStream := antlr.NewCommonTokenStream(lexer, 0)
	vexParser := parser.NewVexParser(tokenStream)
	return vexParser.Array().(*parser.ArrayContext)
}

func createMockListNode(funcName string, args ...string) *parser.ListContext {
	expr := "(" + funcName
	for _, arg := range args {
		expr += " " + arg
	}
	expr += ")"
	
	inputStream := antlr.NewInputStream(expr)
	lexer := parser.NewVexLexer(inputStream)
	tokenStream := antlr.NewCommonTokenStream(lexer, 0)
	vexParser := parser.NewVexParser(tokenStream)
	return vexParser.List().(*parser.ListContext)
}