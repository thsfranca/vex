package analysis

import (
	"fmt"
	"strings"
	"testing"

	"github.com/antlr4-go/antlr/v4"
	"github.com/thsfranca/vex/internal/transpiler/parser"
)

func TestNewAnalyzer(t *testing.T) {
	analyzer := NewAnalyzer()
	
	if analyzer == nil {
		t.Fatal("NewAnalyzer should return non-nil analyzer")
	}
	
	if analyzer.symbolTable == nil {
		t.Error("Analyzer should initialize symbol table")
	}
	
	if analyzer.errorReporter == nil {
		t.Error("Analyzer should initialize error reporter")
	}
}

func TestAnalyzer_SetErrorReporter(t *testing.T) {
	analyzer := NewAnalyzer()
	reporter := NewErrorReporter()
	
	analyzer.SetErrorReporter(reporter)
	
	if analyzer.errorReporter != reporter {
		t.Error("SetErrorReporter should set the provided reporter")
	}
}

func TestAnalyzer_GetErrorReporter(t *testing.T) {
	analyzer := NewAnalyzer()
	
	reporter := analyzer.GetErrorReporter()
	if reporter == nil {
		t.Error("GetErrorReporter should return non-nil reporter")
	}
}

func TestAnalyzer_Analyze(t *testing.T) {
	analyzer := NewAnalyzer()
	
	// Create test AST with simple expression that won't cause semantic errors
	ast := &MockAST{
		root: createMockProgramNode("42"),
	}
	
	symbolTable, err := analyzer.Analyze(ast)
	if err != nil {
		t.Errorf("Analyze should succeed for simple input: %v", err)
		return
	}
	
	if symbolTable == nil {
		t.Error("Analyze should return non-nil symbol table")
	}
}

func TestAnalyzer_Analyze_WithErrors(t *testing.T) {
	analyzer := NewAnalyzer()
	
	// Create AST that will cause errors (incomplete def)
	ast := &MockAST{
		root: createMockProgramNode("(def)"),
	}
	
	_, err := analyzer.Analyze(ast)
	if err == nil {
		t.Error("Analyze should return error for invalid input")
	}
	
	if !strings.Contains(err.Error(), "analysis failed with errors") {
		t.Errorf("Error should mention analysis failure, got: %v", err.Error())
	}
}

func TestAnalyzer_VisitProgram(t *testing.T) {
	analyzer := NewAnalyzer()
	
	// Create mock program context
	programCtx := createMockProgramNode("(def x 42)")
	
	err := analyzer.VisitProgram(programCtx)
	if err != nil {
		t.Errorf("VisitProgram should succeed: %v", err)
	}
}

func TestAnalyzer_VisitList_DefExpression(t *testing.T) {
	analyzer := NewAnalyzer()
	
	// For this test, let's manually call analyzeDef with proper arguments
	args := []Value{
		NewBasicValue("x", "symbol"),
		NewBasicValue("42", "number"),
	}
	
	listCtx := createMockListNode("def", "x", "42")
	
	value, err := analyzer.analyzeDef(listCtx, args)
	if err != nil {
		t.Errorf("analyzeDef should succeed for valid def: %v", err)
		return
	}
	
	if value == nil {
		t.Error("analyzeDef should return non-nil value")
	}
	
	// Check that symbol was defined
	_, exists := analyzer.symbolTable.Lookup("x")
	if !exists {
		t.Error("Symbol 'x' should be defined after def expression")
	}
}

func TestAnalyzer_VisitList_EmptyExpression(t *testing.T) {
	analyzer := NewAnalyzer()
	
	// Create empty list context
	listCtx := createEmptyMockListNode()
	
	_, err := analyzer.VisitList(listCtx)
	if err == nil {
		t.Error("VisitList should return error for empty expression")
	}
	
	// Should report error
	if !analyzer.errorReporter.HasErrors() {
		t.Error("Should report error for empty expression")
	}
}

func TestAnalyzer_VisitArray(t *testing.T) {
	analyzer := NewAnalyzer()
	
	// Create mock array context
	arrayCtx := createMockArrayNode("1", "2", "3")
	
	value, err := analyzer.VisitArray(arrayCtx)
	if err != nil {
		t.Errorf("VisitArray should succeed: %v", err)
		return
	}
	
	if value == nil {
		t.Error("VisitArray should return non-nil value")
	}
	
	if value.Type() != "[]interface{}" {
		t.Errorf("Array value type = %v, want %v", value.Type(), "[]interface{}")
	}
}

func TestAnalyzer_VisitTerminal(t *testing.T) {
	analyzer := NewAnalyzer()
	
	tests := []struct {
		name     string
		text     string
		wantType string
	}{
		{"String literal", "\"hello\"", "string"},
		{"Number", "42", "number"},
		{"Boolean true", "true", "bool"},
		{"Boolean false", "false", "bool"},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			terminal := createMockTerminalNode(tt.text)
			
			value, err := analyzer.VisitTerminal(terminal)
			if err != nil && tt.wantType != "undefined" {
				t.Errorf("VisitTerminal should succeed for %s: %v", tt.name, err)
				return
			}
			
			if value == nil {
				t.Error("VisitTerminal should return non-nil value")
				return
			}
			
			if value.Type() != tt.wantType {
				t.Errorf("Terminal value type = %v, want %v", value.Type(), tt.wantType)
			}
			
			if value.String() != tt.text {
				t.Errorf("Terminal value = %v, want %v", value.String(), tt.text)
			}
		})
	}
}

func TestAnalyzer_VisitTerminal_UndefinedSymbol(t *testing.T) {
	analyzer := NewAnalyzer()
	
	terminal := createMockTerminalNode("undefined_var")
	
	value, err := analyzer.VisitTerminal(terminal)
	if err == nil {
		t.Error("VisitTerminal should return error for undefined symbol")
	}
	
	if value == nil {
		t.Error("VisitTerminal should return value even for error case")
		return
	}
	
	if value.Type() != "undefined" {
		t.Errorf("Undefined symbol type = %v, want %v", value.Type(), "undefined")
	}
	
	// Should report error
	if !analyzer.errorReporter.HasErrors() {
		t.Error("Should report error for undefined symbol")
	}
}

func TestAnalyzer_analyzeDef(t *testing.T) {
	analyzer := NewAnalyzer()
	
	tests := []struct {
		name     string
		args     []Value
		wantErr  bool
		symbol   string
		value    string
	}{
		{
			name: "Valid definition",
			args: []Value{
				NewBasicValue("x", "symbol"),
				NewBasicValue("42", "number"),
			},
			wantErr: false,
			symbol:  "x",
			value:   "42",
		},
		{
			name: "Insufficient arguments",
			args: []Value{
				NewBasicValue("x", "symbol"),
			},
			wantErr: true,
		},
		{
			name:    "No arguments",
			args:    []Value{},
			wantErr: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset analyzer state
			analyzer.symbolTable = NewSymbolTable()
			analyzer.errorReporter.Clear()
			
			listCtx := createMockListNode("def")
			
			value, err := analyzer.analyzeDef(listCtx, tt.args)
			
			if (err != nil) != tt.wantErr {
				t.Errorf("analyzeDef() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			
			if !tt.wantErr {
				if value == nil {
					t.Error("analyzeDef should return non-nil value for success")
					return
				}
				
				if value.String() != tt.value {
					t.Errorf("Returned value = %v, want %v", value.String(), tt.value)
				}
				
				// Check symbol was defined
				_, exists := analyzer.symbolTable.Lookup(tt.symbol)
				if !exists {
					t.Errorf("Symbol %s should be defined", tt.symbol)
				}
			}
		})
	}
}

func TestAnalyzer_analyzeIf(t *testing.T) {
	analyzer := NewAnalyzer()
	
	tests := []struct {
		name    string
		args    []Value
		wantErr bool
	}{
		{
			name: "Valid if with then and else",
			args: []Value{
				NewBasicValue("true", "bool"),
				NewBasicValue("42", "number"),
				NewBasicValue("0", "number"),
			},
			wantErr: false,
		},
		{
			name: "Valid if with only then",
			args: []Value{
				NewBasicValue("true", "bool"),
				NewBasicValue("42", "number"),
			},
			wantErr: false,
		},
		{
			name: "Insufficient arguments",
			args: []Value{
				NewBasicValue("true", "bool"),
			},
			wantErr: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			analyzer.errorReporter.Clear()
			
			listCtx := createMockListNode("if")
			
			value, err := analyzer.analyzeIf(listCtx, tt.args)
			
			if (err != nil) != tt.wantErr {
				t.Errorf("analyzeIf() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			
			if !tt.wantErr && value == nil {
				t.Error("analyzeIf should return non-nil value for success")
			}
		})
	}
}

func TestAnalyzer_analyzeFn(t *testing.T) {
	analyzer := NewAnalyzer()
	
	args := []Value{
		NewBasicValue("[x y]", "array"),
		NewBasicValue("(+ x y)", "expression"),
	}
	
	listCtx := createMockListNode("fn")
	
	value, err := analyzer.analyzeFn(listCtx, args)
	if err != nil {
		t.Errorf("analyzeFn should succeed: %v", err)
		return
	}
	
	if value == nil {
		t.Error("analyzeFn should return non-nil value")
		return
	}
	
	if value.Type() != "func" {
		t.Errorf("Function value type = %v, want %v", value.Type(), "func")
	}
}

func TestAnalyzer_analyzeMacro(t *testing.T) {
	analyzer := NewAnalyzer()
	
	tests := []struct {
		name      string
		args      []Value
		wantErr   bool
		checkName string
	}{
		{
			name: "Valid macro definition",
			args: []Value{
				NewBasicValue("my-macro", "symbol"),
				NewBasicValue("[x]", "array"),
				NewBasicValue("x", "expression"),
			},
			wantErr:   false,
			checkName: "my-macro",
		},
		{
			name: "Reserved word macro name",
			args: []Value{
				NewBasicValue("if", "symbol"),
				NewBasicValue("[x]", "array"),
				NewBasicValue("x", "expression"),
			},
			wantErr: true,
		},
		{
			name: "Insufficient arguments",
			args: []Value{
				NewBasicValue("my-macro", "symbol"),
				NewBasicValue("[x]", "array"),
			},
			wantErr: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset state
			analyzer.symbolTable = NewSymbolTable()
			analyzer.errorReporter.Clear()
			
			listCtx := createMockListNode("macro")
			
			value, err := analyzer.analyzeMacro(listCtx, tt.args)
			
			if (err != nil) != tt.wantErr {
				t.Errorf("analyzeMacro() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			
			if !tt.wantErr {
				if value == nil {
					t.Error("analyzeMacro should return non-nil value for success")
					return
				}
				
				if value.Type() != "macro" {
					t.Errorf("Macro value type = %v, want %v", value.Type(), "macro")
				}
				
				// Check symbol was defined
				_, exists := analyzer.symbolTable.Lookup(tt.checkName)
				if !exists {
					t.Errorf("Macro symbol %s should be defined", tt.checkName)
				}
			}
		})
	}
}

func TestAnalyzer_analyzeFunctionCall(t *testing.T) {
	analyzer := NewAnalyzer()
	
	tests := []struct {
		name     string
		funcName string
		args     []Value
		wantWarn bool
	}{
		{
			name:     "Builtin function",
			funcName: "+",
			args: []Value{
				NewBasicValue("1", "number"),
				NewBasicValue("2", "number"),
			},
			wantWarn: false,
		},
		{
			name:     "Package function",
			funcName: "fmt/Println",
			args: []Value{
				NewBasicValue("\"hello\"", "string"),
			},
			wantWarn: false,
		},
		{
			name:     "Unknown function",
			funcName: "unknown-func",
			args:     []Value{},
			wantWarn: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			analyzer.errorReporter.Clear()
			
			listCtx := createMockListNode(tt.funcName)
			
			value, err := analyzer.analyzeFunctionCall(listCtx, tt.funcName, tt.args)
			if err != nil {
				t.Errorf("analyzeFunctionCall should not error: %v", err)
				return
			}
			
			if value == nil {
				t.Error("analyzeFunctionCall should return non-nil value")
				return
			}
			
			warnings := analyzer.errorReporter.GetWarnings()
			hasWarning := len(warnings) > 0
			
			if hasWarning != tt.wantWarn {
				t.Errorf("Expected warning = %v, got warning = %v", tt.wantWarn, hasWarning)
			}
		})
	}
}

// Helper functions for testing

func TestHelper_isNumber(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"42", true},
		{"0", true},
		{"123", true},
		{"", false},
		{"abc", false},
		{"12a", false},
		{"a12", false},
	}
	
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := isNumber(tt.input)
			if got != tt.want {
				t.Errorf("isNumber(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestHelper_isReservedWord(t *testing.T) {
	tests := []struct {
		word string
		want bool
	}{
		{"if", true},
		{"def", true},
		{"fn", true},
		{"let", true},
		{"do", true},
		{"when", true},
		{"unless", true},
		{"import", true},
		{"custom-word", false},
		{"my-function", false},
	}
	
	for _, tt := range tests {
		t.Run(tt.word, func(t *testing.T) {
			got := isReservedWord(tt.word)
			if got != tt.want {
				t.Errorf("isReservedWord(%q) = %v, want %v", tt.word, got, tt.want)
			}
		})
	}
}

func TestHelper_isBuiltinFunction(t *testing.T) {
	tests := []struct {
		name string
		want bool
	}{
		{"+", true},
		{"-", true},
		{"*", true},
		{"/", true},
		{">", true},
		{"<", true},
		{"=", true},
		{"not", true},
		{"first", true},
		{"rest", true},
		{"println", true},
		{"fmt/Println", true},
		{"os/Exit", true},
		{"custom-func", false},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isBuiltinFunction(tt.name)
			if got != tt.want {
				t.Errorf("isBuiltinFunction(%q) = %v, want %v", tt.name, got, tt.want)
			}
		})
	}
}

// Mock implementations for testing

type MockAST struct {
	root antlr.Tree
}

func (m *MockAST) Accept(visitor ASTVisitor) error {
	if programCtx, ok := m.root.(*parser.ProgramContext); ok {
		return visitor.VisitProgram(programCtx)
	}
	return fmt.Errorf("invalid AST root type")
}

// Mock helper functions for testing
func createMockProgramNode(input string) *parser.ProgramContext {
	// Use ANTLR to parse the input for more realistic testing
	inputStream := antlr.NewInputStream(input)
	lexer := parser.NewVexLexer(inputStream)
	tokenStream := antlr.NewCommonTokenStream(lexer, 0)
	vexParser := parser.NewVexParser(tokenStream)
	return vexParser.Program().(*parser.ProgramContext)
}

func createMockListNode(elements ...string) *parser.ListContext {
	// Create a simple expression to parse
	expr := "(test"
	for _, elem := range elements {
		expr += " " + elem
	}
	expr += ")"
	
	inputStream := antlr.NewInputStream(expr)
	lexer := parser.NewVexLexer(inputStream)
	tokenStream := antlr.NewCommonTokenStream(lexer, 0)
	vexParser := parser.NewVexParser(tokenStream)
	return vexParser.List().(*parser.ListContext)
}

func createEmptyMockListNode() *parser.ListContext {
	// Create minimal valid list that will be empty when processed
	inputStream := antlr.NewInputStream("()")
	lexer := parser.NewVexLexer(inputStream)
	tokenStream := antlr.NewCommonTokenStream(lexer, 0)
	vexParser := parser.NewVexParser(tokenStream)
	return vexParser.List().(*parser.ListContext)
}

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

func createMockTerminalNode(text string) antlr.TerminalNode {
	return &mockTerminalNode{text: text}
}

type mockTerminalNode struct {
	text string
}

func (m *mockTerminalNode) GetText() string { return m.text }
func (m *mockTerminalNode) GetSymbol() antlr.Token {
	return nil // Simplified for testing
}
func (m *mockTerminalNode) Accept(visitor antlr.ParseTreeVisitor) interface{} { return nil }
func (m *mockTerminalNode) GetChild(i int) antlr.Tree { return nil }
func (m *mockTerminalNode) GetChildCount() int { return 0 }
func (m *mockTerminalNode) GetChildren() []antlr.Tree { return nil }
func (m *mockTerminalNode) GetParent() antlr.Tree { return nil }
func (m *mockTerminalNode) GetPayload() interface{} { return m.text }
func (m *mockTerminalNode) GetSourceInterval() antlr.Interval { return antlr.Interval{} }
func (m *mockTerminalNode) SetParent(parent antlr.Tree) {}
func (m *mockTerminalNode) ToStringTree(ruleNames []string, recog antlr.Recognizer) string { return m.text }

