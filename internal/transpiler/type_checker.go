// Package transpiler provides type checking for the Vex language
package transpiler

import (
	"fmt"
	"strings"

	"github.com/antlr4-go/antlr/v4"
	"github.com/thsfranca/vex/internal/transpiler/parser"
)

// TypeChecker performs multi-pass type checking and validation
type TypeChecker struct {
	namespaceManager *NamespaceManager
	typeInference    *TypeInference
	typeUtils        *TypeUtils
	errors           []TypeError
	currentPass      int
}

// TypeError represents a type checking error
type TypeError struct {
	Message  string
	Location string
	Line     int
	Column   int
	Pass     int
}

// String returns the string representation of a type error
func (te *TypeError) String() string {
	if te.Line > 0 {
		return fmt.Sprintf("Type error (pass %d) at line %d:%d (%s): %s", te.Pass, te.Line, te.Column, te.Location, te.Message)
	}
	return fmt.Sprintf("Type error (pass %d) at %s: %s", te.Pass, te.Location, te.Message)
}

// NewTypeChecker creates a new type checker
func NewTypeChecker(namespaceManager *NamespaceManager) *TypeChecker {
	return &TypeChecker{
		namespaceManager: namespaceManager,
		typeInference:    NewTypeInference(namespaceManager),
		typeUtils:        &TypeUtils{},
		errors:           make([]TypeError, 0, 16),
		currentPass:      0,
	}
}

// CheckProgram performs comprehensive type checking on a program
func (tc *TypeChecker) CheckProgram(programCtx *parser.ProgramContext) ([]TypeError, error) {
	tc.errors = tc.errors[:0] // Reset errors

	// Pass 1: Symbol collection and basic binding
	tc.currentPass = 1
	err := tc.collectSymbols(programCtx)
	if err != nil {
		return tc.errors, err
	}

	// Pass 2: Type inference and constraint generation
	tc.currentPass = 2
	err = tc.performTypeInference(programCtx)
	if err != nil {
		return tc.errors, err
	}

	// Pass 3: Type compatibility validation
	tc.currentPass = 3
	err = tc.validateTypeCompatibility(programCtx)
	if err != nil {
		return tc.errors, err
	}

	// Pass 4: Semantic validation (immutability, etc.)
	tc.currentPass = 4
	err = tc.validateSemantics(programCtx)
	if err != nil {
		return tc.errors, err
	}

	return tc.errors, nil
}

// collectSymbols (Pass 1) collects all symbol definitions before type inference
func (tc *TypeChecker) collectSymbols(programCtx *parser.ProgramContext) error {
	for _, child := range programCtx.GetChildren() {
		if listCtx, ok := child.(*parser.ListContext); ok {
			err := tc.collectSymbolsFromList(listCtx)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// collectSymbolsFromList collects symbols from a list expression
func (tc *TypeChecker) collectSymbolsFromList(listCtx *parser.ListContext) error {
	children := listCtx.GetChildren()
	if len(children) < 3 {
		return nil
	}

	content := children[1 : len(children)-1]
	if len(content) == 0 {
		return nil
	}

	// Get the first element
	firstChild := content[0]
	var firstElement string
	if symbolCtx, ok := firstChild.(*antlr.TerminalNodeImpl); ok {
		firstElement = symbolCtx.GetText()
	}

	switch firstElement {
	case "def":
		return tc.collectVariableDefinition(content)
	case "deftype":
		return tc.collectTypeDefinition(content)
	default:
		// Recursively process nested lists
		for _, child := range content {
			if nestedList, ok := child.(*parser.ListContext); ok {
				err := tc.collectSymbolsFromList(nestedList)
				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}

// collectVariableDefinition collects variable definitions for symbol table
func (tc *TypeChecker) collectVariableDefinition(content []antlr.Tree) error {
	if len(content) < 3 {
		tc.addError("Invalid variable definition: expected (def name value)", "unknown")
		return nil
	}

	// Get variable name
	var varName string
	if symbolCtx, ok := content[1].(*antlr.TerminalNodeImpl); ok {
		varName = symbolCtx.GetText()
	} else {
		tc.addError("Variable name must be a symbol", "unknown")
		return nil
	}

	// Pre-bind with unknown type (will be inferred later)
	currentNS := tc.namespaceManager.GetCurrentNamespace()
	currentNS.Bind(varName, NewUnknownType(tc.typeInference.getNextUnknownID()), false)

	return nil
}

// collectTypeDefinition collects type definitions for symbol table
func (tc *TypeChecker) collectTypeDefinition(content []antlr.Tree) error {
	if len(content) < 3 {
		tc.addError("Invalid type definition: expected (deftype name definition)", "unknown")
		return nil
	}

	// Get type name
	var typeName string
	if symbolCtx, ok := content[1].(*antlr.TerminalNodeImpl); ok {
		typeName = symbolCtx.GetText()
	} else {
		tc.addError("Type name must be a symbol", "unknown")
		return nil
	}

	// TODO: Implement custom type definitions
	// For now, just register as a symbol type
	currentNS := tc.namespaceManager.GetCurrentNamespace()
	currentNS.Bind(typeName, SymbolType, false)

	return nil
}

// performTypeInference (Pass 2) performs type inference on all expressions
func (tc *TypeChecker) performTypeInference(programCtx *parser.ProgramContext) error {
	for _, child := range programCtx.GetChildren() {
		_, err := tc.typeInference.InferTypes(child)
		if err != nil {
			tc.addError(fmt.Sprintf("Type inference failed: %s", err.Error()), "unknown")
		}
	}
	return nil
}

// validateTypeCompatibility (Pass 3) validates type compatibility in expressions
func (tc *TypeChecker) validateTypeCompatibility(programCtx *parser.ProgramContext) error {
	for _, child := range programCtx.GetChildren() {
		err := tc.validateExpressionTypes(child)
		if err != nil {
			return err
		}
	}
	return nil
}

// validateExpressionTypes validates types in expressions
func (tc *TypeChecker) validateExpressionTypes(node antlr.Tree) error {
	switch ctx := node.(type) {
	case *parser.ListContext:
		return tc.validateListTypes(ctx)
	case *parser.ArrayContext:
		return tc.validateArrayTypes(ctx)
	default:
		return nil
	}
}

// validateListTypes validates types in list expressions
func (tc *TypeChecker) validateListTypes(listCtx *parser.ListContext) error {
	children := listCtx.GetChildren()
	if len(children) < 3 {
		return nil
	}

	content := children[1 : len(children)-1]
	if len(content) == 0 {
		return nil
	}

	// Get the first element
	firstChild := content[0]
	var firstElement string
	if symbolCtx, ok := firstChild.(*antlr.TerminalNodeImpl); ok {
		firstElement = symbolCtx.GetText()
	}

	switch firstElement {
	case "def":
		return tc.validateDefinitionTypes(content)
	case "+", "-", "*", "/":
		return tc.validateArithmeticTypes(firstElement, content[1:])
	case "if":
		return tc.validateConditionalTypes(content[1:])
	default:
		// Check if it's a function call
		if binding, exists := tc.namespaceManager.GetCurrentNamespace().Resolve(firstElement); exists {
			if binding.IsFunction {
				return tc.validateFunctionCallTypes(binding, content[1:])
			}
		}

		// Recursively validate nested expressions
		for _, child := range content {
			err := tc.validateExpressionTypes(child)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// validateArrayTypes validates types in array literals
func (tc *TypeChecker) validateArrayTypes(arrayCtx *parser.ArrayContext) error {
	children := arrayCtx.GetChildren()
	if len(children) < 3 {
		return nil
	}

	content := children[1 : len(children)-1]
	if len(content) == 0 {
		return nil
	}

	// Get the type of the first element
	firstType, err := tc.typeInference.InferTypes(content[0])
	if err != nil {
		tc.addError(fmt.Sprintf("Cannot infer type of array element: %s", err.Error()), "array")
		return nil
	}

	// Validate that all elements have compatible types
	for i := 1; i < len(content); i++ {
		elementType, err := tc.typeInference.InferTypes(content[i])
		if err != nil {
			tc.addError(fmt.Sprintf("Cannot infer type of array element %d: %s", i, err.Error()), "array")
			continue
		}

		if !firstType.IsAssignableFrom(elementType) {
			tc.addError(fmt.Sprintf("Array element type mismatch: expected %s, got %s",
				firstType.String(), elementType.String()), "array")
		}
	}

	return nil
}

// validateDefinitionTypes validates variable definition types
func (tc *TypeChecker) validateDefinitionTypes(content []antlr.Tree) error {
	if len(content) < 3 {
		tc.addError("Invalid definition: expected (def name value)", "definition")
		return nil
	}

	// Get variable name
	var varName string
	if symbolCtx, ok := content[1].(*antlr.TerminalNodeImpl); ok {
		varName = symbolCtx.GetText()
	} else {
		tc.addErrorWithNode("Variable name must be a symbol", "definition", content[1])
		return nil
	}

	// Infer the value type
	valueType, err := tc.typeInference.InferTypes(content[2])
	if err != nil {
		tc.addErrorWithNode(fmt.Sprintf("Cannot infer type of variable %s: %s", varName, err.Error()), "definition", content[2])
		return nil
	}

	// Update the binding with the inferred type
	currentNS := tc.namespaceManager.GetCurrentNamespace()
	currentNS.Bind(varName, valueType, false)

	return nil
}

// validateArithmeticTypes validates arithmetic expression types
func (tc *TypeChecker) validateArithmeticTypes(operator string, operands []antlr.Tree) error {
	if len(operands) < 2 {
		tc.addError(fmt.Sprintf("Arithmetic operation %s requires at least 2 operands", operator), "arithmetic")
		return nil
	}

	for i, operand := range operands {
		operandType, err := tc.typeInference.InferTypes(operand)
		if err != nil {
			tc.addError(fmt.Sprintf("Cannot infer type of operand %d in %s: %s", i, operator, err.Error()), "arithmetic")
			continue
		}

		// Check if operand is numeric
		if !tc.isNumericType(operandType) {
			tc.addError(fmt.Sprintf("Arithmetic operation %s requires numeric operands, got %s",
				operator, operandType.String()), "arithmetic")
		}
	}

	return nil
}

// validateConditionalTypes validates conditional expression types
func (tc *TypeChecker) validateConditionalTypes(content []antlr.Tree) error {
	if len(content) < 3 {
		tc.addError("Conditional requires condition, then-branch, and else-branch", "conditional")
		return nil
	}

	// Validate condition type
	conditionType, err := tc.typeInference.InferTypes(content[0])
	if err != nil {
		tc.addError(fmt.Sprintf("Cannot infer condition type: %s", err.Error()), "conditional")
	} else if !conditionType.Equals(BoolType) {
		tc.addError(fmt.Sprintf("Condition must be boolean, got %s", conditionType.String()), "conditional")
	}

	// Validate branch types
	thenType, err := tc.typeInference.InferTypes(content[1])
	if err != nil {
		tc.addError(fmt.Sprintf("Cannot infer then-branch type: %s", err.Error()), "conditional")
		return nil
	}

	elseType, err := tc.typeInference.InferTypes(content[2])
	if err != nil {
		tc.addError(fmt.Sprintf("Cannot infer else-branch type: %s", err.Error()), "conditional")
		return nil
	}

	if !thenType.IsAssignableFrom(elseType) {
		tc.addError(fmt.Sprintf("Conditional branch type mismatch: then-branch is %s, else-branch is %s",
			thenType.String(), elseType.String()), "conditional")
	}

	return nil
}

// validateFunctionCallTypes validates function call types
func (tc *TypeChecker) validateFunctionCallTypes(funcBinding *Binding, args []antlr.Tree) error {
	funcType, ok := funcBinding.Type.(*FunctionType)
	if !ok {
		tc.addError(fmt.Sprintf("Symbol %s is not a function", funcBinding.Symbol.Name), "function-call")
		return nil
	}

	if len(args) != len(funcType.Parameters) {
		tc.addError(fmt.Sprintf("Function %s expects %d arguments, got %d",
			funcBinding.Symbol.Name, len(funcType.Parameters), len(args)), "function-call")
		return nil
	}

	// Validate each argument type
	for i, arg := range args {
		argType, err := tc.typeInference.InferTypes(arg)
		if err != nil {
			tc.addError(fmt.Sprintf("Cannot infer type of argument %d to function %s: %s",
				i, funcBinding.Symbol.Name, err.Error()), "function-call")
			continue
		}

		expectedType := funcType.Parameters[i]
		if !expectedType.IsAssignableFrom(argType) {
			tc.addError(fmt.Sprintf("Argument %d to function %s: expected %s, got %s",
				i, funcBinding.Symbol.Name, expectedType.String(), argType.String()), "function-call")
		}
	}

	return nil
}

// validateSemantics (Pass 4) validates semantic constraints like immutability
func (tc *TypeChecker) validateSemantics(programCtx *parser.ProgramContext) error {
	for _, child := range programCtx.GetChildren() {
		err := tc.validateSemanticConstraints(child)
		if err != nil {
			return err
		}
	}
	return nil
}

// validateSemanticConstraints validates semantic constraints
func (tc *TypeChecker) validateSemanticConstraints(node antlr.Tree) error {
	switch ctx := node.(type) {
	case *parser.ListContext:
		return tc.validateListSemantics(ctx)
	default:
		return nil
	}
}

// validateListSemantics validates semantic constraints in lists
func (tc *TypeChecker) validateListSemantics(listCtx *parser.ListContext) error {
	children := listCtx.GetChildren()
	if len(children) < 3 {
		return nil
	}

	content := children[1 : len(children)-1]
	if len(content) == 0 {
		return nil
	}

	// Get the first element
	firstChild := content[0]
	var firstElement string
	if symbolCtx, ok := firstChild.(*antlr.TerminalNodeImpl); ok {
		firstElement = symbolCtx.GetText()
	}

	switch firstElement {
	case "set!":
		return tc.validateMutationSemantics(content[1:])
	default:
		// Recursively validate nested expressions
		for _, child := range content {
			err := tc.validateSemanticConstraints(child)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// validateMutationSemantics validates mutation operations (should be restricted)
func (tc *TypeChecker) validateMutationSemantics(content []antlr.Tree) error {
	if len(content) < 2 {
		tc.addError("Invalid mutation: expected (set! var value)", "mutation")
		return nil
	}

	// Get variable name
	var varName string
	if symbolCtx, ok := content[0].(*antlr.TerminalNodeImpl); ok {
		varName = symbolCtx.GetText()
	} else {
		tc.addError("Mutation target must be a symbol", "mutation")
		return nil
	}

	// Check if variable is mutable
	currentNS := tc.namespaceManager.GetCurrentNamespace()
	if binding, exists := currentNS.Resolve(varName); exists {
		if !binding.IsMutable {
			tc.addError(fmt.Sprintf("Cannot mutate immutable variable %s", varName), "mutation")
		}
	} else {
		tc.addError(fmt.Sprintf("Undefined variable %s in mutation", varName), "mutation")
	}

	return nil
}

// Helper methods

// isNumericType checks if a type is numeric (int or float)
func (tc *TypeChecker) isNumericType(vexType VexType) bool {
	return vexType.Equals(IntType) || vexType.Equals(FloatType)
}

// addError adds a type error to the error list
func (tc *TypeChecker) addError(message, location string) {
	tc.addErrorWithNode(message, location, nil)
}

// addErrorWithNode adds a type error with position information from an AST node
func (tc *TypeChecker) addErrorWithNode(message, location string, node antlr.Tree) {
	line, column := tc.getNodePosition(node)
	error := TypeError{
		Message:  message,
		Location: location,
		Line:     line,
		Column:   column,
		Pass:     tc.currentPass,
	}
	tc.errors = append(tc.errors, error)
}

// getNodePosition extracts line and column information from an AST node
func (tc *TypeChecker) getNodePosition(node antlr.Tree) (int, int) {
	if node == nil {
		return 0, 0
	}

	switch ctx := node.(type) {
	case *antlr.TerminalNodeImpl:
		token := ctx.GetSymbol()
		return token.GetLine(), token.GetColumn()
	case *parser.ListContext:
		// Get position from the first token (opening parenthesis)
		if ctx.GetStart() != nil {
			return ctx.GetStart().GetLine(), ctx.GetStart().GetColumn()
		}
	case *parser.ArrayContext:
		// Get position from the first token (opening bracket)
		if ctx.GetStart() != nil {
			return ctx.GetStart().GetLine(), ctx.GetStart().GetColumn()
		}
	case *parser.ProgramContext:
		// Get position from the first token
		if ctx.GetStart() != nil {
			return ctx.GetStart().GetLine(), ctx.GetStart().GetColumn()
		}
	}
	return 0, 0
}

// GetErrors returns all collected type errors
func (tc *TypeChecker) GetErrors() []TypeError {
	return tc.errors
}

// HasErrors returns true if there are any type errors
func (tc *TypeChecker) HasErrors() bool {
	return len(tc.errors) > 0
}

// GetErrorsByPass returns errors from a specific pass
func (tc *TypeChecker) GetErrorsByPass(pass int) []TypeError {
	var result []TypeError
	for _, err := range tc.errors {
		if err.Pass == pass {
			result = append(result, err)
		}
	}
	return result
}

// FormatErrors returns a formatted string of all errors
func (tc *TypeChecker) FormatErrors() string {
	if len(tc.errors) == 0 {
		return "No type errors"
	}

	var builder strings.Builder
	builder.WriteString(fmt.Sprintf("Found %d type error(s):\n", len(tc.errors)))

	for i, err := range tc.errors {
		builder.WriteString(fmt.Sprintf("  %d. %s\n", i+1, err.String()))
	}

	return builder.String()
}
