package transpiler

import (
	"fmt"
	"strings"

	"github.com/antlr4-go/antlr/v4"
	"github.com/thsfranca/vex/internal/transpiler/parser"
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

// EmitSlashNotationCall generates Go package function calls from slash notation
// (fmt/Println "message") -> fmt.Println("message")
func (cg *CodeGenerator) EmitSlashNotationCall(packageName, functionName string, args []string) {
	argsStr := strings.Join(args, ", ")
	cg.writeIndented(fmt.Sprintf("%s.%s(%s)\n", packageName, functionName, argsStr))
}

// EmitFunctionLiteral generates Go function literals
// (fn [w r] body) -> func(w http.ResponseWriter, r *http.Request) { body }
func (cg *CodeGenerator) EmitFunctionLiteral(params []string, bodyElements []antlr.Tree, visitor *ASTVisitor) string {
	var result strings.Builder
	
	// Start function literal
	result.WriteString("func(")
	
	// Add parameters with types
	if len(params) == 2 {
		// Assume HTTP handler signature for 2 parameters
		result.WriteString(fmt.Sprintf("%s http.ResponseWriter, %s *http.Request", params[0], params[1]))
		// Add http import
		cg.EmitImport("\"net/http\"")
	} else {
		// Generic parameters - types will need to be inferred or specified
		for i, param := range params {
			if i > 0 {
				result.WriteString(", ")
			}
			result.WriteString(fmt.Sprintf("%s interface{}", param))
		}
	}
	
	result.WriteString(") {\n")
	
	// Generate function body by processing each body element
	for _, bodyElement := range bodyElements {
		bodyCode := cg.processFunctionBodyElement(bodyElement, visitor)
		if bodyCode != "" {
			result.WriteString(fmt.Sprintf("\t%s\n", bodyCode))
		}
	}
	
	result.WriteString("}")
	
	return result.String()
}

// processFunctionBodyElement processes a single element in a function body
func (cg *CodeGenerator) processFunctionBodyElement(element antlr.Tree, visitor *ASTVisitor) string {
	switch ctx := element.(type) {
	case *parser.ListContext:
		// Method calls, function calls, etc.
		children := ctx.GetChildren()
		if len(children) < 3 { // Need at least ( symbol )
			return ""
		}
		
		// Get the first element (function/method name)
		firstChild := children[1] // Skip opening parenthesis
		if terminalNode, ok := firstChild.(*antlr.TerminalNodeImpl); ok {
			firstElement := terminalNode.GetText()
			
			// Extract arguments (skip opening paren, function name, closing paren)
			var args []string
			for i := 2; i < len(children)-1; i++ {
				argValue := visitor.evaluateExpression(children[i])
				args = append(args, argValue)
			}
			
			if strings.HasPrefix(firstElement, ".") {
				// Method call: (.WriteString w "Hello!")
				if len(args) >= 1 {
					receiver := args[0]
					methodArgs := args[1:]
					argsStr := strings.Join(methodArgs, ", ")
					return fmt.Sprintf("%s.%s(%s)", receiver, firstElement[1:], argsStr)
				}
			} else {
				// Function call: (fmt.Println "Hello")
				argsStr := strings.Join(args, ", ")
				return fmt.Sprintf("%s(%s)", firstElement, argsStr)
			}
		}
		
	default:
		// Use the visitor's evaluateExpression for everything else
		return visitor.evaluateExpression(element)
	}
	
	return ""
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
