package transpiler

import (
	"fmt"
	"strings"
)

// CodeGenerator handles the generation of Go code from AST nodes
type CodeGenerator struct {
	indentLevel int
	buffer      strings.Builder
}

// NewCodeGenerator creates a new code generator instance
func NewCodeGenerator() *CodeGenerator {
	return &CodeGenerator{
		indentLevel: 0,
		buffer:      strings.Builder{},
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
func (cg *CodeGenerator) EmitVariableDefinition(name, value string) {
	cg.writeIndented(fmt.Sprintf("%s := %s\n", name, value))
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

// Reset clears the code generator state
func (cg *CodeGenerator) Reset() {
	cg.buffer.Reset()
	cg.indentLevel = 0
}