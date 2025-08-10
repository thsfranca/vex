package transpiler

import (
	"github.com/antlr4-go/antlr/v4"
	"github.com/thsfranca/vex/internal/transpiler/analysis"
	"github.com/thsfranca/vex/internal/transpiler/ast"
	"github.com/thsfranca/vex/internal/transpiler/codegen"
	"github.com/thsfranca/vex/internal/transpiler/macro"
	"github.com/thsfranca/vex/internal/transpiler/parser"
)

// ParserAdapter adapts ast.VexParser to the Parser interface
type ParserAdapter struct {
	parser *ast.VexParser
    exports map[string]map[string]bool
}

func NewParserAdapter() *ParserAdapter {
	return &ParserAdapter{
        parser:  ast.NewParser(),
        exports: make(map[string]map[string]bool),
	}
}

func (pa *ParserAdapter) Parse(input string) (AST, error) {
	astResult, err := pa.parser.Parse(input)
	if err != nil {
		return nil, err
	}
	
	// Convert ast.AST to our AST interface
	return NewConcreteAST(astResult.Root()), nil
}

func (pa *ParserAdapter) ParseFile(filename string) (AST, error) {
	astResult, err := pa.parser.ParseFile(filename)
	if err != nil {
		return nil, err
	}
	
	return NewConcreteAST(astResult.Root()), nil
}

// AnalyzerAdapter adapts analysis.AnalyzerImpl to the Analyzer interface
type AnalyzerAdapter struct {
	analyzer *analysis.AnalyzerImpl
}

func NewAnalyzerAdapter() *AnalyzerAdapter {
	return &AnalyzerAdapter{
		analyzer: analysis.NewAnalyzer(),
	}
}

func (aa *AnalyzerAdapter) Analyze(ast AST) (SymbolTable, error) {
	// Create a compatible AST for the analyzer
	analysisAST := &AnalysisAST{ast: ast}
	
	symbolTable, err := aa.analyzer.Analyze(analysisAST)
	if err != nil {
		return nil, err
	}
	
	// Wrap the symbol table
	return &SymbolTableAdapter{table: symbolTable}, nil
}

func (aa *AnalyzerAdapter) SetErrorReporter(reporter ErrorReporter) {
	// Convert ErrorReporter to analysis.ErrorReporter
	analysisReporter := &ErrorReporterAdapter{reporter: reporter}
	aa.analyzer.SetErrorReporter(analysisReporter)
}

// CodeGeneratorAdapter adapts codegen.GoCodeGenerator to the CodeGenerator interface
type CodeGeneratorAdapter struct {
	generator *codegen.GoCodeGenerator
}

func NewCodeGeneratorAdapter(config codegen.Config) *CodeGeneratorAdapter {
	return &CodeGeneratorAdapter{
		generator: codegen.NewGoCodeGenerator(config),
	}
}

func (cga *CodeGeneratorAdapter) Generate(ast AST, symbols SymbolTable) (string, error) {
	// Create compatible types
	codegenAST := &CodegenAST{ast: ast}
	codegenSymbols := &CodegenSymbolTable{table: symbols}
	
	return cga.generator.Generate(codegenAST, codegenSymbols)
}

func (cga *CodeGeneratorAdapter) AddImport(importPath string) {
	cga.generator.AddImport(importPath)
}

func (cga *CodeGeneratorAdapter) SetPackageName(name string) {
	cga.generator.SetPackageName(name)
}

// Adapter types for bridging interfaces

type AnalysisAST struct {
	ast AST
}

func (aast *AnalysisAST) Accept(visitor analysis.ASTVisitor) error {
	// Bridge the visitor interface
	bridgeVisitor := &AnalysisVisitorBridge{visitor: visitor}
	return aast.ast.Accept(bridgeVisitor)
}

type AnalysisVisitorBridge struct {
	visitor analysis.ASTVisitor
}

func (avb *AnalysisVisitorBridge) VisitProgram(ctx *parser.ProgramContext) error {
	return avb.visitor.VisitProgram(ctx)
}

func (avb *AnalysisVisitorBridge) VisitList(ctx *parser.ListContext) (Value, error) {
	value, err := avb.visitor.VisitList(ctx)
	if err != nil {
		return nil, err
	}
	return &ValueAdapter{value: value}, nil
}

func (avb *AnalysisVisitorBridge) VisitArray(ctx *parser.ArrayContext) (Value, error) {
	value, err := avb.visitor.VisitArray(ctx)
	if err != nil {
		return nil, err
	}
	return &ValueAdapter{value: value}, nil
}

func (avb *AnalysisVisitorBridge) VisitTerminal(node antlr.TerminalNode) (Value, error) {
	value, err := avb.visitor.VisitTerminal(node)
	if err != nil {
		return nil, err
	}
	return &ValueAdapter{value: value}, nil
}

type ValueAdapter struct {
	value analysis.Value
}

func (va *ValueAdapter) String() string {
	return va.value.String()
}

func (va *ValueAdapter) Type() string {
	return va.value.Type()
}

type SymbolTableAdapter struct {
	table analysis.SymbolTable
}

func (sta *SymbolTableAdapter) Define(name string, value Value) error {
	// Convert Value to analysis.Value
	analysisValue := &AnalysisValueAdapter{value: value}
	return sta.table.Define(name, analysisValue)
}

func (sta *SymbolTableAdapter) Lookup(name string) (Value, bool) {
	value, exists := sta.table.Lookup(name)
	if !exists {
		return nil, false
	}
	return &ValueAdapter{value: value}, true
}

func (sta *SymbolTableAdapter) EnterScope() {
	sta.table.EnterScope()
}

func (sta *SymbolTableAdapter) ExitScope() {
	sta.table.ExitScope()
}

type AnalysisValueAdapter struct {
	value Value
}

func (ava *AnalysisValueAdapter) String() string {
	return ava.value.String()
}

func (ava *AnalysisValueAdapter) Type() string {
	return ava.value.Type()
}

// CodeGenerator adapters

type CodegenAST struct {
	ast AST
}

func (cast *CodegenAST) Accept(visitor codegen.ASTVisitor) error {
	bridgeVisitor := &CodegenVisitorBridge{visitor: visitor}
	return cast.ast.Accept(bridgeVisitor)
}

type CodegenVisitorBridge struct {
	visitor codegen.ASTVisitor
}

func (cvb *CodegenVisitorBridge) VisitProgram(ctx *parser.ProgramContext) error {
	return cvb.visitor.VisitProgram(ctx)
}

func (cvb *CodegenVisitorBridge) VisitList(ctx *parser.ListContext) (Value, error) {
	value, err := cvb.visitor.VisitList(ctx)
	if err != nil {
		return nil, err
	}
	return &CodegenValueAdapter{value: value}, nil
}

func (cvb *CodegenVisitorBridge) VisitArray(ctx *parser.ArrayContext) (Value, error) {
	value, err := cvb.visitor.VisitArray(ctx)
	if err != nil {
		return nil, err
	}
	return &CodegenValueAdapter{value: value}, nil
}

func (cvb *CodegenVisitorBridge) VisitTerminal(node antlr.TerminalNode) (Value, error) {
	value, err := cvb.visitor.VisitTerminal(node)
	if err != nil {
		return nil, err
	}
	return &CodegenValueAdapter{value: value}, nil
}

type CodegenValueAdapter struct {
	value codegen.Value
}

func (cva *CodegenValueAdapter) String() string {
	return cva.value.String()
}

func (cva *CodegenValueAdapter) Type() string {
	return cva.value.Type()
}

type CodegenSymbolTable struct {
	table SymbolTable
}

func (cst *CodegenSymbolTable) Define(name string, value codegen.Value) error {
	// Convert codegen.Value to our Value
	ourValue := &CodegenValueAdapter{value: value}
	return cst.table.Define(name, ourValue)
}

func (cst *CodegenSymbolTable) Lookup(name string) (codegen.Value, bool) {
	value, exists := cst.table.Lookup(name)
	if !exists {
		return nil, false
	}
	return &CodegenValueBridge{value: value}, true
}

func (cst *CodegenSymbolTable) EnterScope() {
	cst.table.EnterScope()
}

func (cst *CodegenSymbolTable) ExitScope() {
	cst.table.ExitScope()
}

type CodegenValueBridge struct {
	value Value
}

func (cvb *CodegenValueBridge) String() string {
	return cvb.value.String()
}

func (cvb *CodegenValueBridge) Type() string {
	return cvb.value.Type()
}

// MacroExpanderAdapter adapts macro.MacroExpanderImpl to the MacroExpander interface
type MacroExpanderAdapter struct {
	expander *macro.MacroExpanderImpl
}

func NewMacroExpanderAdapter(expander *macro.MacroExpanderImpl) *MacroExpanderAdapter {
	return &MacroExpanderAdapter{expander: expander}
}

func (mea *MacroExpanderAdapter) ExpandMacros(ast AST) (AST, error) {
	// Convert AST to macro.AST
	macroAST := &MacroASTAdapter{ast: ast}
	
	expandedAST, err := mea.expander.ExpandMacros(macroAST)
	if err != nil {
		return nil, err
	}
	
	// Convert back to our AST interface
	return NewConcreteAST(expandedAST.Root()), nil
}

func (mea *MacroExpanderAdapter) RegisterMacro(name string, ourMacro *macro.Macro) error {
	return mea.expander.RegisterMacro(name, ourMacro)
}

func (mea *MacroExpanderAdapter) HasMacro(name string) bool {
	return mea.expander.HasMacro(name)
}

func (mea *MacroExpanderAdapter) GetMacro(name string) (*macro.Macro, bool) {
	return mea.expander.GetMacro(name)
}

// MacroASTAdapter adapts our AST to macro.AST
type MacroASTAdapter struct {
	ast AST
}

func (maa *MacroASTAdapter) Root() antlr.Tree {
	return maa.ast.Root()
}

func (maa *MacroASTAdapter) Accept(visitor macro.ASTVisitor) error {
	// Bridge visitor interfaces
	bridgeVisitor := &MacroVisitorBridge{visitor: visitor}
	return maa.ast.Accept(bridgeVisitor)
}

type MacroVisitorBridge struct {
	visitor macro.ASTVisitor
}

func (mvb *MacroVisitorBridge) VisitProgram(ctx *parser.ProgramContext) error {
	return mvb.visitor.VisitProgram(ctx)
}

func (mvb *MacroVisitorBridge) VisitList(ctx *parser.ListContext) (Value, error) {
	value, err := mvb.visitor.VisitList(ctx)
	if err != nil {
		return nil, err
	}
	return &MacroValueAdapter{value: value}, nil
}

func (mvb *MacroVisitorBridge) VisitArray(ctx *parser.ArrayContext) (Value, error) {
	value, err := mvb.visitor.VisitArray(ctx)
	if err != nil {
		return nil, err
	}
	return &MacroValueAdapter{value: value}, nil
}

func (mvb *MacroVisitorBridge) VisitTerminal(node antlr.TerminalNode) (Value, error) {
	value, err := mvb.visitor.VisitTerminal(node)
	if err != nil {
		return nil, err
	}
	return &MacroValueAdapter{value: value}, nil
}

type MacroValueAdapter struct {
	value macro.Value
}

func (mva *MacroValueAdapter) String() string {
	return mva.value.String()
}

func (mva *MacroValueAdapter) Type() string {
	return mva.value.Type()
}

// ErrorReporterAdapter adapts our ErrorReporter to analysis.ErrorReporter
type ErrorReporterAdapter struct {
	reporter ErrorReporter
}

func (era *ErrorReporterAdapter) ReportError(line, column int, message string) {
	era.reporter.ReportError(line, column, message)
}

func (era *ErrorReporterAdapter) ReportWarning(line, column int, message string) {
	era.reporter.ReportWarning(line, column, message)
}

func (era *ErrorReporterAdapter) HasErrors() bool {
	return era.reporter.HasErrors()
}

func (era *ErrorReporterAdapter) GetErrors() []analysis.CompilerError {
	errors := era.reporter.GetErrors()
	analysisErrors := make([]analysis.CompilerError, len(errors))
	
	for i, err := range errors {
		analysisErrors[i] = analysis.CompilerError{
			Line:    err.Line,
			Column:  err.Column,
			Message: err.Message,
			Type:    analysis.ErrorType(err.Type), // Assuming compatible enum values
		}
	}
	
	return analysisErrors
}