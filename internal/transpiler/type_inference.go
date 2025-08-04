// Package transpiler provides type inference for the Vex language
package transpiler

import (
	"fmt"

	"github.com/antlr4-go/antlr/v4"
	"github.com/thsfranca/vex/internal/transpiler/parser"
)

// TypeInference implements Hindley-Milner style type inference
type TypeInference struct {
	namespaceManager *NamespaceManager
	typeUtils        *TypeUtils
	substitutions    map[int]VexType // Unknown type ID -> concrete type mappings
	nextUnknownID    int
	constraints      []*TypeConstraint
}

// TypeConstraint represents a constraint between two types that must be unified
type TypeConstraint struct {
	Left    VexType
	Right   VexType
	Context string // For error reporting
}

// NewTypeInference creates a new type inference engine
func NewTypeInference(namespaceManager *NamespaceManager) *TypeInference {
	return &TypeInference{
		namespaceManager: namespaceManager,
		typeUtils:        &TypeUtils{},
		substitutions:    make(map[int]VexType, 64),
		nextUnknownID:    1,
		constraints:      make([]*TypeConstraint, 0, 32),
	}
}

// InferTypes performs type inference on an AST node
func (ti *TypeInference) InferTypes(node antlr.Tree) (VexType, error) {
	// Reset state for new inference
	ti.constraints = ti.constraints[:0]

	// Generate constraints from the AST
	inferredType, err := ti.generateConstraints(node)
	if err != nil {
		return nil, err
	}

	// Solve constraints through unification
	err = ti.solveConstraints()
	if err != nil {
		return nil, err
	}

	// Apply substitutions to get final type
	finalType := ti.applySubstitutions(inferredType)
	return finalType, nil
}

// generateConstraints traverses the AST and generates type constraints
func (ti *TypeInference) generateConstraints(node antlr.Tree) (VexType, error) {
	switch ctx := node.(type) {
	case *antlr.TerminalNodeImpl:
		return ti.inferLiteralType(ctx.GetText())

	case *parser.ListContext:
		return ti.inferListType(ctx)

	case *parser.ArrayContext:
		return ti.inferArrayType(ctx)

	case *parser.ProgramContext:
		// For programs, infer the type of the last expression
		children := ctx.GetChildren()
		if len(children) > 0 {
			return ti.generateConstraints(children[len(children)-1])
		}
		return NewUnknownType(ti.nextUnknownID), nil

	default:
		return NewUnknownType(ti.nextUnknownID), fmt.Errorf("unsupported node type for inference: %T", ctx)
	}
}

// inferLiteralType infers the type of a literal value
func (ti *TypeInference) inferLiteralType(value string) (VexType, error) {
	return ti.typeUtils.InferLiteralType(value), nil
}

// inferListType infers the type of a list (S-expression)
func (ti *TypeInference) inferListType(ctx *parser.ListContext) (VexType, error) {
	children := ctx.GetChildren()
	if len(children) < 3 { // '(' ... ')'
		return NewUnknownType(ti.getNextUnknownID()), nil
	}

	content := children[1 : len(children)-1]
	if len(content) == 0 {
		return NewUnknownType(ti.getNextUnknownID()), nil
	}

	// Get the first element to determine the operation
	firstChild := content[0]
	var firstElement string
	if symbolCtx, ok := firstChild.(*antlr.TerminalNodeImpl); ok {
		firstElement = symbolCtx.GetText()
	}

	switch firstElement {
	case "def":
		return ti.inferDefinitionType(content)
	case "fn":
		return ti.inferFunctionType(content[1:]) // Skip "fn"
	case "+", "-", "*", "/":
		return ti.inferArithmeticType(firstElement, content[1:])
	case "if":
		return ti.inferConditionalType(content[1:])
	default:
		// Check if it's a function call
		if binding, exists := ti.namespaceManager.GetCurrentNamespace().Resolve(firstElement); exists {
			if binding.IsFunction {
				return ti.inferFunctionCallType(binding, content[1:])
			}
		}
		// Default to unknown type
		return NewUnknownType(ti.getNextUnknownID()), nil
	}
}

// inferArrayType infers the type of an array literal
func (ti *TypeInference) inferArrayType(ctx *parser.ArrayContext) (VexType, error) {
	children := ctx.GetChildren()
	if len(children) < 3 { // '[' ... ']'
		return NewListType(NewUnknownType(ti.getNextUnknownID())), nil
	}

	content := children[1 : len(children)-1]
	if len(content) == 0 {
		return NewListType(NewUnknownType(ti.getNextUnknownID())), nil
	}

	// Infer the type of the first element
	elementType, err := ti.generateConstraints(content[0])
	if err != nil {
		return nil, err
	}

	// Add constraints that all elements must have the same type
	for i := 1; i < len(content); i++ {
		otherElementType, err := ti.generateConstraints(content[i])
		if err != nil {
			return nil, err
		}

		ti.addConstraint(elementType, otherElementType, fmt.Sprintf("array element %d", i))
	}

	return NewListType(elementType), nil
}

// inferDefinitionType infers the type of a variable definition
func (ti *TypeInference) inferDefinitionType(content []antlr.Tree) (VexType, error) {
	if len(content) < 3 {
		return nil, fmt.Errorf("invalid definition: expected (def name value)")
	}

	// Get variable name
	var varName string
	if symbolCtx, ok := content[1].(*antlr.TerminalNodeImpl); ok {
		varName = symbolCtx.GetText()
	} else {
		return nil, fmt.Errorf("invalid definition: name must be a symbol")
	}

	// Infer the type of the value
	valueType, err := ti.generateConstraints(content[2])
	if err != nil {
		return nil, err
	}

	// Bind the variable in the current namespace
	ti.namespaceManager.GetCurrentNamespace().Bind(varName, valueType, false)

	// Definition expressions don't have a return value (void)
	return NewUnknownType(ti.getNextUnknownID()), nil
}

// inferFunctionType infers the type of a function literal
func (ti *TypeInference) inferFunctionType(content []antlr.Tree) (VexType, error) {
	if len(content) < 2 {
		return nil, fmt.Errorf("invalid function: expected (fn [params] body)")
	}

	// Parse parameters
	paramTypes, err := ti.inferParameterTypes(content[0])
	if err != nil {
		return nil, err
	}

	// Create a new scope for the function
	_ = ti.namespaceManager.CreateNamespace("func", ti.namespaceManager.GetCurrentNamespace())
	oldNamespace := ti.namespaceManager.GetCurrentNamespace()
	ti.namespaceManager.SetCurrentNamespace("func")

	// Infer the return type from the body
	returnType, err := ti.generateConstraints(content[1])
	if err != nil {
		ti.namespaceManager.SetCurrentNamespace(oldNamespace.Name)
		return nil, err
	}

	// Restore the previous namespace
	ti.namespaceManager.SetCurrentNamespace(oldNamespace.Name)

	return NewFunctionType(paramTypes, returnType), nil
}

// inferParameterTypes infers the types of function parameters
func (ti *TypeInference) inferParameterTypes(paramNode antlr.Tree) ([]VexType, error) {
	arrayCtx, ok := paramNode.(*parser.ArrayContext)
	if !ok {
		return nil, fmt.Errorf("function parameters must be an array")
	}

	children := arrayCtx.GetChildren()
	if len(children) < 3 { // '[' ... ']'
		return []VexType{}, nil
	}

	content := children[1 : len(children)-1]
	paramTypes := make([]VexType, len(content))

	for i, paramNode := range content {
		// For now, parameters have unknown types (will be inferred from usage)
		paramTypes[i] = NewUnknownType(ti.getNextUnknownID())

		// If it's a terminal node (parameter name), bind it in the function scope
		if symbolCtx, ok := paramNode.(*antlr.TerminalNodeImpl); ok {
			paramName := symbolCtx.GetText()
			ti.namespaceManager.GetCurrentNamespace().Bind(paramName, paramTypes[i], false)
		}
	}

	return paramTypes, nil
}

// inferArithmeticType infers the type of arithmetic expressions
func (ti *TypeInference) inferArithmeticType(operator string, operands []antlr.Tree) (VexType, error) {
	if len(operands) < 2 {
		return nil, fmt.Errorf("arithmetic operation %s requires at least 2 operands", operator)
	}

	// For arithmetic operations, all operands should be numbers
	// Start with the first operand
	firstType, err := ti.generateConstraints(operands[0])
	if err != nil {
		return nil, err
	}

	// Add constraints that all operands must be the same numeric type
	for i := 1; i < len(operands); i++ {
		operandType, err := ti.generateConstraints(operands[i])
		if err != nil {
			return nil, err
		}

		ti.addConstraint(firstType, operandType, fmt.Sprintf("arithmetic operand %d", i))
	}

	// Result type is the same as operand types for basic arithmetic
	return firstType, nil
}

// inferConditionalType infers the type of conditional expressions
func (ti *TypeInference) inferConditionalType(content []antlr.Tree) (VexType, error) {
	if len(content) < 3 {
		return nil, fmt.Errorf("conditional requires condition, then-branch, and else-branch")
	}

	// Infer condition type (should be boolean)
	conditionType, err := ti.generateConstraints(content[0])
	if err != nil {
		return nil, err
	}
	ti.addConstraint(conditionType, BoolType, "conditional condition")

	// Infer then-branch type
	thenType, err := ti.generateConstraints(content[1])
	if err != nil {
		return nil, err
	}

	// Infer else-branch type
	elseType, err := ti.generateConstraints(content[2])
	if err != nil {
		return nil, err
	}

	// Then and else branches must have the same type
	ti.addConstraint(thenType, elseType, "conditional branches")

	return thenType, nil
}

// inferFunctionCallType infers the type of a function call
func (ti *TypeInference) inferFunctionCallType(funcBinding *Binding, args []antlr.Tree) (VexType, error) {
	funcType, ok := funcBinding.Type.(*FunctionType)
	if !ok {
		return nil, fmt.Errorf("symbol %s is not a function", funcBinding.Symbol.Name)
	}

	if len(args) != len(funcType.Parameters) {
		return nil, fmt.Errorf("function %s expects %d arguments, got %d",
			funcBinding.Symbol.Name, len(funcType.Parameters), len(args))
	}

	// Add constraints for each argument
	for i, arg := range args {
		argType, err := ti.generateConstraints(arg)
		if err != nil {
			return nil, err
		}

		ti.addConstraint(argType, funcType.Parameters[i],
			fmt.Sprintf("function %s argument %d", funcBinding.Symbol.Name, i))
	}

	return funcType.ReturnType, nil
}

// addConstraint adds a type constraint to be solved later
func (ti *TypeInference) addConstraint(left, right VexType, context string) {
	constraint := &TypeConstraint{
		Left:    left,
		Right:   right,
		Context: context,
	}
	ti.constraints = append(ti.constraints, constraint)
}

// solveConstraints solves all accumulated constraints through unification
func (ti *TypeInference) solveConstraints() error {
	for _, constraint := range ti.constraints {
		err := ti.unifyConstraint(constraint)
		if err != nil {
			return fmt.Errorf("type error in %s: %w", constraint.Context, err)
		}
	}
	return nil
}

// unifyConstraint unifies a single constraint
func (ti *TypeInference) unifyConstraint(constraint *TypeConstraint) error {
	left := ti.applySubstitutions(constraint.Left)
	right := ti.applySubstitutions(constraint.Right)

	return ti.unify(left, right)
}

// unify attempts to unify two types, adding substitutions as needed
func (ti *TypeInference) unify(type1, type2 VexType) error {
	// Apply existing substitutions
	type1 = ti.applySubstitutions(type1)
	type2 = ti.applySubstitutions(type2)

	// If types are already equal, no unification needed
	if type1.Equals(type2) {
		return nil
	}

	// Handle unknown types
	if unknown1, ok := type1.(*UnknownType); ok {
		ti.substitutions[unknown1.ID] = type2
		return nil
	}

	if unknown2, ok := type2.(*UnknownType); ok {
		ti.substitutions[unknown2.ID] = type1
		return nil
	}

	// Handle structured types
	unified, err := ti.typeUtils.UnifyTypes(type1, type2)
	if err != nil {
		return err
	}

	// If unification succeeded but types weren't equal, update substitutions
	if !type1.Equals(unified) {
		if unknown1, ok := type1.(*UnknownType); ok {
			ti.substitutions[unknown1.ID] = unified
		}
	}

	if !type2.Equals(unified) {
		if unknown2, ok := type2.(*UnknownType); ok {
			ti.substitutions[unknown2.ID] = unified
		}
	}

	return nil
}

// applySubstitutions applies all current substitutions to a type
func (ti *TypeInference) applySubstitutions(vexType VexType) VexType {
	switch t := vexType.(type) {
	case *UnknownType:
		if substitution, exists := ti.substitutions[t.ID]; exists {
			// Recursively apply substitutions to avoid chains
			return ti.applySubstitutions(substitution)
		}
		return t

	case *ListType:
		elementType := ti.applySubstitutions(t.ElementType)
		if !elementType.Equals(t.ElementType) {
			return NewListType(elementType)
		}
		return t

	case *MapType:
		keyType := ti.applySubstitutions(t.KeyType)
		valueType := ti.applySubstitutions(t.ValueType)
		if !keyType.Equals(t.KeyType) || !valueType.Equals(t.ValueType) {
			return NewMapType(keyType, valueType)
		}
		return t

	case *FunctionType:
		paramTypes := make([]VexType, len(t.Parameters))
		changed := false

		for i, param := range t.Parameters {
			paramTypes[i] = ti.applySubstitutions(param)
			if !paramTypes[i].Equals(param) {
				changed = true
			}
		}

		returnType := ti.applySubstitutions(t.ReturnType)
		if !returnType.Equals(t.ReturnType) {
			changed = true
		}

		if changed {
			return NewFunctionType(paramTypes, returnType)
		}
		return t

	default:
		return vexType
	}
}

// getNextUnknownID returns the next unique ID for unknown types
func (ti *TypeInference) getNextUnknownID() int {
	id := ti.nextUnknownID
	ti.nextUnknownID++
	return id
}

// GetSubstitutions returns a copy of current substitutions (for debugging)
func (ti *TypeInference) GetSubstitutions() map[int]VexType {
	result := make(map[int]VexType, len(ti.substitutions))
	for id, vexType := range ti.substitutions {
		result[id] = vexType
	}
	return result
}

// GetConstraints returns a copy of current constraints (for debugging)
func (ti *TypeInference) GetConstraints() []*TypeConstraint {
	result := make([]*TypeConstraint, len(ti.constraints))
	copy(result, ti.constraints)
	return result
}
