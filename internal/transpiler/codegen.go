package transpiler

import (
	"strconv"
	"strings"
	"sync"

	"github.com/antlr4-go/antlr/v4"
	"github.com/thsfranca/vex/internal/transpiler/parser"
)

// stringSlicePool provides pooled string slices to reduce allocations
var stringSlicePool = sync.Pool{
	New: func() interface{} {
		return make([]string, 0, 8) // Pre-allocate capacity for common use cases
	},
}

// CodeGenerator handles the generation of Go code from AST nodes
type CodeGenerator struct {
	indentLevel   int
	buffer        strings.Builder
	imports       map[string]bool // Track unique imports
	indentCache   []string        // Cache for indent strings
	operatorCache map[string]string // Cache for operator conversions
}

// NewCodeGenerator creates a new code generator instance
func NewCodeGenerator() *CodeGenerator {
	cg := &CodeGenerator{
		indentLevel:   0,
		buffer:        strings.Builder{},
		imports:       make(map[string]bool, 8),              // Pre-allocate for common imports
		indentCache:   make([]string, 0, 10),                // Pre-allocate for common indentation levels
		operatorCache: make(map[string]string, 4),          // Cache for +, -, *, /
	}
	cg.buffer.Grow(1024) // Pre-allocate capacity for typical Go code generation
	return cg
}

// EmitNumber generates Go code for a numeric literal
func (cg *CodeGenerator) EmitNumber(value string) {
	cg.writeIndented("_ = " + value + "\n")
}

// EmitString generates Go code for a string literal
func (cg *CodeGenerator) EmitString(value string) {
	cg.writeIndented("_ = " + value + "\n")
}

// EmitSymbol generates Go code for a symbol
func (cg *CodeGenerator) EmitSymbol(symbol string) {
	cg.writeIndented("_ = " + symbol + "\n")
}

// EmitVariableDefinition generates Go code for variable definition
// (def x 10) -> x := 10
// TODO: This should only be called if semantic analysis determines the variable is used later
// Otherwise, Go will produce "declared and not used" errors, which is correct behavior
func (cg *CodeGenerator) EmitVariableDefinition(name, value string) {
	cg.writeIndented(name + " := " + value + "\n")
}

// EmitExpressionStatement generates Go code for standalone expressions
// (42) or (+ 1 2) -> _ = 42 or _ = 1 + 2 (discarded result)
func (cg *CodeGenerator) EmitExpressionStatement(expression string) {
	cg.writeIndented("_ = " + expression + "\n")
}

// EmitArithmeticExpression generates Go code for arithmetic operations
// (+ 1 2) -> 1 + 2
func (cg *CodeGenerator) EmitArithmeticExpression(operator string, operands []string) {
	if len(operands) < 2 {
		cg.writeIndented("// Invalid arithmetic expression with " + strconv.Itoa(len(operands)) + " operands\n")
		return
	}

	// Convert Lisp prefix notation to Go infix notation
	goOperator := cg.convertOperator(operator)
	expression := strings.Join(operands, " " + goOperator + " ")
	cg.writeIndented("_ = " + expression + "\n")
}

// EmitImport collects import statements for later generation
// (import "net/http") -> adds to imports collection
func (cg *CodeGenerator) EmitImport(importPath string) {
	// Clean the import path (remove quotes if they exist, then add them back)
	cleanPath := strings.Trim(importPath, "\"")
	cg.imports["\"" + cleanPath + "\""] = true
}

// EmitMethodCall generates Go method calls
// (.HandleFunc router "/path" handler) -> router.HandleFunc("/path", handler)
func (cg *CodeGenerator) EmitMethodCall(receiver, methodName string, args []string) {
	argsStr := strings.Join(args, ", ")
	cg.writeIndented("_ = " + receiver + "." + methodName + "(" + argsStr + ")\n")
}

// EmitSlashNotationCall generates Go package function calls from slash notation
// (fmt/Println "message") -> fmt.Println("message")
func (cg *CodeGenerator) EmitSlashNotationCall(packageName, functionName string, args []string) {
	argsStr := strings.Join(args, ", ")
	cg.writeIndented(packageName + "." + functionName + "(" + argsStr + ")\n")
}

// EmitFunctionLiteral generates Go function literals
// (fn [w r] body) -> func(w http.ResponseWriter, r *http.Request) { body }
func (cg *CodeGenerator) EmitFunctionLiteral(params []string, bodyElements []antlr.Tree, visitor *ASTVisitor) string {
	var result strings.Builder
	result.Grow(128) // Pre-allocate capacity for typical function literals
	
	// Start function literal
	result.WriteString("func(")
	
	// Add parameters with types
	if len(params) == 2 {
		// Assume HTTP handler signature for 2 parameters
		result.WriteString(params[0] + " http.ResponseWriter, " + params[1] + " *http.Request")
		// Add http import
		cg.EmitImport("\"net/http\"")
	} else {
		// Generic parameters - types will need to be inferred or specified
		for i, param := range params {
			if i > 0 {
				result.WriteString(", ")
			}
			result.WriteString(param + " interface{}")
		}
	}
	
	result.WriteString(") {\n")
	
	// Generate function body by processing each body element
	for _, bodyElement := range bodyElements {
		bodyCode := cg.processFunctionBodyElement(bodyElement, visitor)
		if bodyCode != "" {
			result.WriteString("\t" + bodyCode + "\n")
		}
	}
	
	result.WriteString("}")
	
	return result.String()
}

// getStringSlice returns a pooled string slice
func getStringSlice() []string {
	slice := stringSlicePool.Get().([]string)
	return slice[:0] // Reset length but keep capacity
}

// putStringSlice returns a string slice to the pool
func putStringSlice(slice []string) {
	if cap(slice) > 64 { // Avoid keeping very large slices in pool
		return
	}
	stringSlicePool.Put(slice)
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
			args := getStringSlice()
			defer putStringSlice(args)
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
					return receiver + "." + firstElement[1:] + "(" + argsStr + ")"
				}
			} else {
				// Function call: (fmt.Println "Hello")
				argsStr := strings.Join(args, ", ")
				return firstElement + "(" + argsStr + ")"
			}
		}
		
	default:
		// Use the visitor's evaluateExpression for everything else
		return visitor.evaluateExpression(element)
	}
	
	return ""
}



// convertOperator converts Vex operators to Go operators (with caching)
func (cg *CodeGenerator) convertOperator(vexOp string) string {
	// Check cache first
	if goOp, exists := cg.operatorCache[vexOp]; exists {
		return goOp
	}
	
	// Compute and cache result
	var goOp string
	switch vexOp {
	case "+":
		goOp = "+"
	case "-":
		goOp = "-"
	case "*":
		goOp = "*"
	case "/":
		goOp = "/"
	default:
		goOp = vexOp // fallback
	}
	
	cg.operatorCache[vexOp] = goOp
	return goOp
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
	indent := cg.getIndent()
	cg.buffer.WriteString(indent + line)
}

// getIndent returns cached indentation string for current level
func (cg *CodeGenerator) getIndent() string {
	// Extend cache if needed
	for len(cg.indentCache) <= cg.indentLevel {
		level := len(cg.indentCache)
		cg.indentCache = append(cg.indentCache, strings.Repeat("\t", level))
	}
	return cg.indentCache[cg.indentLevel]
}

// GetCode returns the generated code
func (cg *CodeGenerator) GetCode() string {
	return cg.buffer.String()
}

// GetImports returns all collected imports
func (cg *CodeGenerator) GetImports() []string {
	imports := getStringSlice()
	for importPath := range cg.imports {
		imports = append(imports, importPath)
	}
	// Don't defer putStringSlice since we're returning the slice
	return imports
}

// Reset clears the code generator state
func (cg *CodeGenerator) Reset() {
	cg.buffer.Reset()
	cg.indentLevel = 0
	cg.imports = make(map[string]bool, 8)                   // Pre-allocate for common imports
	// Keep caches but clear them to avoid memory leaks
	cg.indentCache = cg.indentCache[:0]                     // Keep capacity, reset length
	cg.operatorCache = make(map[string]string, 4)          // Cache for +, -, *, /
}

// Typed code generation methods for semantic visitor

// EmitTypedVariableDefinition generates Go code for typed variable definition
func (cg *CodeGenerator) EmitTypedVariableDefinition(name, value string, vexType VexType) {
	goType := vexType.GoType()
	if goType == "interface{}" {
		// Use type inference if type is unknown
		cg.writeIndented(name + " := " + value + "\n")
	} else {
		// Use explicit type declaration
		cg.writeIndented("var " + name + " " + goType + " = " + value + "\n")
	}
	// Add a dummy usage to prevent "declared and not used" errors
	cg.writeIndented("_ = " + name + "\n")
}

// EmitTypedFunction generates Go code for typed function definition
func (cg *CodeGenerator) EmitTypedFunction(name string, paramNames []string, paramTypes []VexType, returnType VexType) {
	cg.writeIndented("func " + name + "(")
	
	// Generate parameter list with types
	for i, paramName := range paramNames {
		if i > 0 {
			cg.buffer.WriteString(", ")
		}
		paramType := "interface{}"
		if i < len(paramTypes) && paramTypes[i].GoType() != "interface{}" {
			paramType = paramTypes[i].GoType()
		}
		cg.buffer.WriteString(paramName + " " + paramType)
	}
	
	cg.buffer.WriteString(") ")
	
	// Add return type if not unknown
	if returnType.GoType() != "interface{}" {
		cg.buffer.WriteString(returnType.GoType() + " ")
	}
	
	cg.buffer.WriteString("{\n")
	cg.IncreaseIndent()
	// Function body will be added by subsequent calls
}

// EmitTypedFunctionLiteral generates Go code for typed function literal
func (cg *CodeGenerator) EmitTypedFunctionLiteral(paramNames []string, paramTypes []VexType, returnType VexType) {
	cg.writeIndented("func(")
	
	// Generate parameter list with types
	for i, paramName := range paramNames {
		if i > 0 {
			cg.buffer.WriteString(", ")
		}
		paramType := "interface{}"
		if i < len(paramTypes) && paramTypes[i].GoType() != "interface{}" {
			paramType = paramTypes[i].GoType()
		}
		cg.buffer.WriteString(paramName + " " + paramType)
	}
	
	cg.buffer.WriteString(") ")
	
	// Add return type if not unknown
	if returnType.GoType() != "interface{}" {
		cg.buffer.WriteString(returnType.GoType() + " ")
	}
	
	cg.buffer.WriteString("{\n")
	// Function body will be handled separately
}

// EmitTypedArithmeticExpression generates Go code for typed arithmetic
func (cg *CodeGenerator) EmitTypedArithmeticExpression(operator string, operands []string, resultType VexType) {
	if len(operands) < 2 {
		cg.writeIndented("// Invalid arithmetic expression\n")
		return
	}
	
	// Build expression with proper type casting if needed
	var expr string
	if resultType.Equals(FloatType) {
		// Ensure float operations
		expr = cg.buildFloatArithmetic(operator, operands)
	} else {
		// Regular integer arithmetic
		expr = cg.buildIntArithmetic(operator, operands)
	}
	
	cg.writeIndented("_ = " + expr + "\n")
}

// EmitTypedConditional generates Go code for typed conditional expression
func (cg *CodeGenerator) EmitTypedConditional(condition, thenBranch, elseBranch string, resultType VexType) {
	cg.writeIndented("func() " + resultType.GoType() + " {\n")
	cg.IncreaseIndent()
	cg.writeIndented("if " + condition + " {\n")
	cg.IncreaseIndent()
	cg.writeIndented("return " + thenBranch + "\n")
	cg.DecreaseIndent()
	cg.writeIndented("} else {\n")
	cg.IncreaseIndent()
	cg.writeIndented("return " + elseBranch + "\n")
	cg.DecreaseIndent()
	cg.writeIndented("}\n")
	cg.DecreaseIndent()
	cg.writeIndented("}()\n")
}

// EmitTypedMethodCall generates Go code for typed method call
func (cg *CodeGenerator) EmitTypedMethodCall(receiver, method string, args []string, receiverType VexType, argTypes []VexType) {
	// Add type assertions if needed
	typedReceiver := receiver
	if receiverType.GoType() != "interface{}" {
		// Could add type assertion here if needed
		typedReceiver = receiver
	}
	
	cg.writeIndented("_ = " + typedReceiver + "." + method + "(")
	for i, arg := range args {
		if i > 0 {
			cg.buffer.WriteString(", ")
		}
		cg.buffer.WriteString(arg)
	}
	cg.buffer.WriteString(")\n")
}

// EmitTypedSlashNotationCall generates Go code for typed package function call
func (cg *CodeGenerator) EmitTypedSlashNotationCall(packageName, funcName string, args []string, argTypes []VexType) {
	cg.EmitImport("\"" + packageName + "\"")
	
	// Don't assign function calls to _ - just execute them as statements
	cg.writeIndented(packageName + "." + funcName + "(")
	for i, arg := range args {
		if i > 0 {
			cg.buffer.WriteString(", ")
		}
		cg.buffer.WriteString(arg)
	}
	cg.buffer.WriteString(")\n")
}

// EmitTypedFunctionCall generates Go code for typed function call
func (cg *CodeGenerator) EmitTypedFunctionCall(functionName string, args []string, returnType VexType) {
	var expr string
	if returnType.GoType() == "interface{}" {
		expr = functionName + "("
	} else {
		expr = functionName + "("
	}
	
	for i, arg := range args {
		if i > 0 {
			expr += ", "
		}
		expr += arg
	}
	expr += ")"
	
	cg.writeIndented("_ = " + expr + "\n")
}

// EmitTypedArray generates Go code for typed array literal
func (cg *CodeGenerator) EmitTypedArray(listType *ListType) {
	elementType := listType.ElementType.GoType()
	if elementType == "interface{}" {
		cg.writeIndented("_ = []interface{}{}\n")
	} else {
		cg.writeIndented("_ = []" + elementType + "{}\n")
	}
}

// EmitArray generates Go code for generic array literal (fallback)
func (cg *CodeGenerator) EmitArray() {
	cg.writeIndented("_ = []interface{}{}\n")
}

// Helper methods for typed arithmetic

// buildFloatArithmetic builds float arithmetic expressions
func (cg *CodeGenerator) buildFloatArithmetic(operator string, operands []string) string {
	goOperator := cg.convertOperator(operator)
	
	// Ensure all operands are float64
	floatOperands := make([]string, len(operands))
	for i, operand := range operands {
		if strings.Contains(operand, ".") {
			floatOperands[i] = operand
		} else {
			floatOperands[i] = "float64(" + operand + ")"
		}
	}
	
	// Build left-associative expression
	result := floatOperands[0]
	for i := 1; i < len(floatOperands); i++ {
		result = "(" + result + " " + goOperator + " " + floatOperands[i] + ")"
	}
	
	return result
}

// buildIntArithmetic builds integer arithmetic expressions
func (cg *CodeGenerator) buildIntArithmetic(operator string, operands []string) string {
	goOperator := cg.convertOperator(operator)
	
	// Build left-associative expression
	result := operands[0]
	for i := 1; i < len(operands); i++ {
		result = "(" + result + " " + goOperator + " " + operands[i] + ")"
	}
	
	return result
}
