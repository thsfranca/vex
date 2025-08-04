// Package transpiler provides the Vex type system implementation
package transpiler

import (
	"fmt"
	"reflect"
	"strings"
)

// VexType represents a type in the Vex type system
type VexType interface {
	// String returns the string representation of the type
	String() string
	// GoType returns the corresponding Go type string for transpilation
	GoType() string
	// IsAssignableFrom checks if this type can accept values of another type
	IsAssignableFrom(other VexType) bool
	// Equals checks if two types are equivalent
	Equals(other VexType) bool
}

// TypeKind represents the kind of a Vex type
type TypeKind int

const (
	TypeKindPrimitive TypeKind = iota
	TypeKindList
	TypeKindMap
	TypeKindFunction
	TypeKindSymbol
	TypeKindUnknown
	TypeKindGeneric
)

// PrimitiveType represents primitive types in Vex
type PrimitiveType struct {
	Kind   string // "int", "float", "string", "bool"
	goType string
}

// Built-in primitive types
var (
	IntType    = &PrimitiveType{Kind: "int", goType: "int64"}
	FloatType  = &PrimitiveType{Kind: "float", goType: "float64"}
	StringType = &PrimitiveType{Kind: "string", goType: "string"}
	BoolType   = &PrimitiveType{Kind: "bool", goType: "bool"}
	SymbolType = &PrimitiveType{Kind: "symbol", goType: "Symbol"}
)

// String returns the string representation of the primitive type
func (pt *PrimitiveType) String() string {
	return pt.Kind
}

// GoType returns the corresponding Go type
func (pt *PrimitiveType) GoType() string {
	return pt.goType
}

// IsAssignableFrom checks if this primitive type can accept values of another type
func (pt *PrimitiveType) IsAssignableFrom(other VexType) bool {
	if otherPrim, ok := other.(*PrimitiveType); ok {
		return pt.Kind == otherPrim.Kind
	}
	return false
}

// Equals checks if two primitive types are equal
func (pt *PrimitiveType) Equals(other VexType) bool {
	if otherPrim, ok := other.(*PrimitiveType); ok {
		return pt.Kind == otherPrim.Kind
	}
	return false
}

// ListType represents homogeneous list types [T]
type ListType struct {
	ElementType VexType
}

// NewListType creates a new list type with the specified element type
func NewListType(elementType VexType) *ListType {
	return &ListType{ElementType: elementType}
}

// String returns the string representation of the list type
func (lt *ListType) String() string {
	return fmt.Sprintf("[%s]", lt.ElementType.String())
}

// GoType returns the corresponding Go type (slice)
func (lt *ListType) GoType() string {
	return fmt.Sprintf("[]%s", lt.ElementType.GoType())
}

// IsAssignableFrom checks if this list type can accept values of another type
func (lt *ListType) IsAssignableFrom(other VexType) bool {
	if otherList, ok := other.(*ListType); ok {
		return lt.ElementType.IsAssignableFrom(otherList.ElementType)
	}
	return false
}

// Equals checks if two list types are equal
func (lt *ListType) Equals(other VexType) bool {
	if otherList, ok := other.(*ListType); ok {
		return lt.ElementType.Equals(otherList.ElementType)
	}
	return false
}

// MapType represents immutable map types {K: V}
type MapType struct {
	KeyType   VexType
	ValueType VexType
}

// NewMapType creates a new map type with the specified key and value types
func NewMapType(keyType, valueType VexType) *MapType {
	return &MapType{KeyType: keyType, ValueType: valueType}
}

// String returns the string representation of the map type
func (mt *MapType) String() string {
	return fmt.Sprintf("{%s: %s}", mt.KeyType.String(), mt.ValueType.String())
}

// GoType returns the corresponding Go type (map)
func (mt *MapType) GoType() string {
	return fmt.Sprintf("map[%s]%s", mt.KeyType.GoType(), mt.ValueType.GoType())
}

// IsAssignableFrom checks if this map type can accept values of another type
func (mt *MapType) IsAssignableFrom(other VexType) bool {
	if otherMap, ok := other.(*MapType); ok {
		return mt.KeyType.IsAssignableFrom(otherMap.KeyType) &&
			mt.ValueType.IsAssignableFrom(otherMap.ValueType)
	}
	return false
}

// Equals checks if two map types are equal
func (mt *MapType) Equals(other VexType) bool {
	if otherMap, ok := other.(*MapType); ok {
		return mt.KeyType.Equals(otherMap.KeyType) &&
			mt.ValueType.Equals(otherMap.ValueType)
	}
	return false
}

// FunctionType represents function types with parameters and return type
type FunctionType struct {
	Parameters []VexType
	ReturnType VexType
}

// NewFunctionType creates a new function type
func NewFunctionType(parameters []VexType, returnType VexType) *FunctionType {
	return &FunctionType{Parameters: parameters, ReturnType: returnType}
}

// String returns the string representation of the function type
func (ft *FunctionType) String() string {
	var params []string
	for _, param := range ft.Parameters {
		params = append(params, param.String())
	}
	return fmt.Sprintf("(%s) -> %s", strings.Join(params, ", "), ft.ReturnType.String())
}

// GoType returns the corresponding Go function type
func (ft *FunctionType) GoType() string {
	var params []string
	for _, param := range ft.Parameters {
		params = append(params, param.GoType())
	}
	return fmt.Sprintf("func(%s) %s", strings.Join(params, ", "), ft.ReturnType.GoType())
}

// IsAssignableFrom checks if this function type can accept values of another type
func (ft *FunctionType) IsAssignableFrom(other VexType) bool {
	if otherFunc, ok := other.(*FunctionType); ok {
		if len(ft.Parameters) != len(otherFunc.Parameters) {
			return false
		}
		for i, param := range ft.Parameters {
			if !param.IsAssignableFrom(otherFunc.Parameters[i]) {
				return false
			}
		}
		return ft.ReturnType.IsAssignableFrom(otherFunc.ReturnType)
	}
	return false
}

// Equals checks if two function types are equal
func (ft *FunctionType) Equals(other VexType) bool {
	if otherFunc, ok := other.(*FunctionType); ok {
		if len(ft.Parameters) != len(otherFunc.Parameters) {
			return false
		}
		for i, param := range ft.Parameters {
			if !param.Equals(otherFunc.Parameters[i]) {
				return false
			}
		}
		return ft.ReturnType.Equals(otherFunc.ReturnType)
	}
	return false
}

// GenericType represents generic type parameters (e.g., T in List[T])
type GenericType struct {
	Name string
}

// NewGenericType creates a new generic type parameter
func NewGenericType(name string) *GenericType {
	return &GenericType{Name: name}
}

// String returns the string representation of the generic type
func (gt *GenericType) String() string {
	return gt.Name
}

// GoType returns the interface{} type for Go (will be refined during inference)
func (gt *GenericType) GoType() string {
	return "interface{}"
}

// IsAssignableFrom checks if this generic type can accept values of another type
func (gt *GenericType) IsAssignableFrom(other VexType) bool {
	// Generic types are assignable from any type during inference
	return true
}

// Equals checks if two generic types are equal
func (gt *GenericType) Equals(other VexType) bool {
	if otherGeneric, ok := other.(*GenericType); ok {
		return gt.Name == otherGeneric.Name
	}
	return false
}

// UnknownType represents types that haven't been inferred yet
type UnknownType struct {
	ID int // Unique identifier for this unknown type
}

// NewUnknownType creates a new unknown type with a unique ID
func NewUnknownType(id int) *UnknownType {
	return &UnknownType{ID: id}
}

// String returns the string representation of the unknown type
func (ut *UnknownType) String() string {
	return fmt.Sprintf("?%d", ut.ID)
}

// GoType returns interface{} for unknown types
func (ut *UnknownType) GoType() string {
	return "interface{}"
}

// IsAssignableFrom checks if this unknown type can accept values of another type
func (ut *UnknownType) IsAssignableFrom(other VexType) bool {
	// Unknown types can be unified with any type
	return true
}

// Equals checks if two unknown types are equal
func (ut *UnknownType) Equals(other VexType) bool {
	if otherUnknown, ok := other.(*UnknownType); ok {
		return ut.ID == otherUnknown.ID
	}
	return false
}

// TypeUtils provides utility functions for type operations
type TypeUtils struct{}

// InferLiteralType infers the type of a literal value
func (tu *TypeUtils) InferLiteralType(value string) VexType {
	// Remove quotes from strings
	if strings.HasPrefix(value, "\"") && strings.HasSuffix(value, "\"") {
		return StringType
	}

	// Check if it's a number
	if isInteger(value) {
		return IntType
	}

	if isFloat(value) {
		return FloatType
	}

	// Check boolean
	if value == "true" || value == "false" {
		return BoolType
	}

	// Default to symbol type for identifiers
	return SymbolType
}

// isInteger checks if a string represents an integer
func isInteger(s string) bool {
	if s == "" {
		return false
	}

	// Handle negative numbers
	start := 0
	if s[0] == '-' || s[0] == '+' {
		start = 1
		if len(s) == 1 {
			return false
		}
	}

	for i := start; i < len(s); i++ {
		if s[i] < '0' || s[i] > '9' {
			return false
		}
	}
	return true
}

// isFloat checks if a string represents a floating-point number
func isFloat(s string) bool {
	if s == "" {
		return false
	}

	dotCount := 0
	start := 0

	// Handle negative numbers
	if s[0] == '-' || s[0] == '+' {
		start = 1
		if len(s) == 1 {
			return false
		}
	}

	for i := start; i < len(s); i++ {
		if s[i] == '.' {
			dotCount++
			if dotCount > 1 {
				return false
			}
		} else if s[i] < '0' || s[i] > '9' {
			return false
		}
	}

	return dotCount == 1 // Must have exactly one dot to be a float
}

// GetTypeKind returns the kind of a given type
func (tu *TypeUtils) GetTypeKind(vexType VexType) TypeKind {
	switch vexType.(type) {
	case *PrimitiveType:
		return TypeKindPrimitive
	case *ListType:
		return TypeKindList
	case *MapType:
		return TypeKindMap
	case *FunctionType:
		return TypeKindFunction
	case *GenericType:
		return TypeKindGeneric
	case *UnknownType:
		return TypeKindUnknown
	default:
		return TypeKindUnknown
	}
}

// UnifyTypes attempts to unify two types, returning the most specific common type
func (tu *TypeUtils) UnifyTypes(type1, type2 VexType) (VexType, error) {
	// If either type is unknown, return the other
	if _, ok := type1.(*UnknownType); ok {
		return type2, nil
	}
	if _, ok := type2.(*UnknownType); ok {
		return type1, nil
	}

	// If types are equal, return either one
	if type1.Equals(type2) {
		return type1, nil
	}

	// Handle specific unification cases
	if reflect.TypeOf(type1) == reflect.TypeOf(type2) {
		switch t1 := type1.(type) {
		case *ListType:
			t2 := type2.(*ListType)
			unifiedElement, err := tu.UnifyTypes(t1.ElementType, t2.ElementType)
			if err != nil {
				return nil, err
			}
			return NewListType(unifiedElement), nil

		case *MapType:
			t2 := type2.(*MapType)
			unifiedKey, err := tu.UnifyTypes(t1.KeyType, t2.KeyType)
			if err != nil {
				return nil, err
			}
			unifiedValue, err := tu.UnifyTypes(t1.ValueType, t2.ValueType)
			if err != nil {
				return nil, err
			}
			return NewMapType(unifiedKey, unifiedValue), nil
		}
	}

	return nil, fmt.Errorf("cannot unify types %s and %s", type1.String(), type2.String())
}
