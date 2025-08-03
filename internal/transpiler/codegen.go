package transpiler

import (
	"fmt"
	"strings"
)

// CodeGenerator handles the generation of Go code from AST nodes
type CodeGenerator struct {
	indentLevel int
	buffer      strings.Builder
	imports     map[string]bool // Track unique imports
}

// NewCodeGenerator creates a new code generator instance
func NewCodeGenerator() *CodeGenerator {
	return &CodeGenerator{
		indentLevel: 0,
		buffer:      strings.Builder{},
		imports:     make(map[string]bool),
	}
}

// EmitNumber generates Go code for a numeric literal
func (cg *CodeGenerator) EmitNumber(value string) {
	cg.writeIndented(fmt.Sprintf("_ = %s\n", value))
}

// EmitString generates Go code for a string literal
func (cg *CodeGenerator) EmitString(value string) {
	cg.writeIndented(fmt.Sprintf("_ = %s\n", value))
}

// EmitSymbol generates Go code for a symbol
func (cg *CodeGenerator) EmitSymbol(symbol string) {
	cg.writeIndented(fmt.Sprintf("_ = %s\n", symbol))
}

// EmitVariableDefinition generates Go code for variable definition
// (def x 10) -> x := 10
// TODO: This should only be called if semantic analysis determines the variable is used later
// Otherwise, Go will produce "declared and not used" errors, which is correct behavior
func (cg *CodeGenerator) EmitVariableDefinition(name, value string) {
	cg.writeIndented(fmt.Sprintf("%s := %s\n", name, value))
}

// EmitExpressionStatement generates Go code for standalone expressions
// (42) or (+ 1 2) -> _ = 42 or _ = 1 + 2 (discarded result)
func (cg *CodeGenerator) EmitExpressionStatement(expression string) {
	cg.writeIndented(fmt.Sprintf("_ = %s\n", expression))
}

// EmitArithmeticExpression generates Go code for arithmetic operations
// (+ 1 2) -> 1 + 2
func (cg *CodeGenerator) EmitArithmeticExpression(operator string, operands []string) {
	if len(operands) < 2 {
		cg.writeIndented(fmt.Sprintf("// Invalid arithmetic expression with %d operands\n", len(operands)))
		return
	}

	// Convert Lisp prefix notation to Go infix notation
	goOperator := convertOperator(operator)
	expression := strings.Join(operands, fmt.Sprintf(" %s ", goOperator))
	cg.writeIndented(fmt.Sprintf("_ = %s\n", expression))
}

// EmitImport collects import statements for later generation
// (import "net/http") -> adds to imports collection
func (cg *CodeGenerator) EmitImport(importPath string) {
	// Clean the import path (remove quotes if they exist, then add them back)
	cleanPath := strings.Trim(importPath, "\"")
	cg.imports[fmt.Sprintf("\"%s\"", cleanPath)] = true
}

// EmitMethodCall generates Go method calls
// (.HandleFunc router "/path" handler) -> router.HandleFunc("/path", handler)
func (cg *CodeGenerator) EmitMethodCall(receiver, methodName string, args []string) {
	argsStr := strings.Join(args, ", ")
	cg.writeIndented(fmt.Sprintf("_ = %s.%s(%s)\n", receiver, methodName, argsStr))
}

// convertOperator converts Vex operators to Go operators
func convertOperator(vexOp string) string {
	switch vexOp {
	case "+":
		return "+"
	case "-":
		return "-"
	case "*":
		return "*"
	case "/":
		return "/"
	default:
		return vexOp // fallback
	}
}

// IncreaseIndent increases the current indentation level
func (cg *CodeGenerator) IncreaseIndent() {
	cg.indentLevel++
}

// DecreaseIndent decreases the current indentation level
func (cg *CodeGenerator) DecreaseIndent() {
	if cg.indentLevel > 0 {
		cg.indentLevel--
	}
}

// writeIndented writes a line with proper indentation
func (cg *CodeGenerator) writeIndented(line string) {
	indent := strings.Repeat("\t", cg.indentLevel)
	cg.buffer.WriteString(indent + line)
}

// GetCode returns the generated code
func (cg *CodeGenerator) GetCode() string {
	return cg.buffer.String()
}

// GetImports returns all collected imports
func (cg *CodeGenerator) GetImports() []string {
	var imports []string
	for importPath := range cg.imports {
		imports = append(imports, importPath)
	}
	return imports
}

// Reset clears the code generator state
func (cg *CodeGenerator) Reset() {
	cg.buffer.Reset()
	cg.indentLevel = 0
	cg.imports = make(map[string]bool)
}
