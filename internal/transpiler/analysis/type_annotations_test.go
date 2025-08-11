package analysis

import (
	"testing"
)

func TestAnalyzer_ParameterParsing(t *testing.T) {
    tests := []struct {
        name                string
        paramListInner      string
        expectedParamNames  []string
        expectedParamTypes  []string
    }{

        {
            name:               "All parameters with type annotations",
            paramListInner:     "x: int y: string z: bool",
            expectedParamNames: []string{"x", "y", "z"},
            expectedParamTypes: []string{"int", "string", "bool"},
        },

        {
            name:               "Single parameter with type",
            paramListInner:     "name: string",
            expectedParamNames: []string{"name"},
            expectedParamTypes: []string{"string"},
        },
        {
            name:               "Empty parameter list",
            paramListInner:     "",
            expectedParamNames: []string{},
            expectedParamTypes: []string{},
        },
        {
            name:               "Parameters without type annotations (should fail)",
            paramListInner:     "x y z",
            expectedParamNames: nil, // Should fail
            expectedParamTypes: nil,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            analyzer := NewAnalyzer()
            
            // Enter scope to collect defined symbols
            analyzer.symbolTable.EnterScope()
            defer analyzer.symbolTable.ExitScope()

            // Parse parameters
            paramTypes := analyzer.parseParameterListWithTypes(tt.paramListInner)

            // Handle expected failure case
            if tt.expectedParamTypes == nil {
                if paramTypes != nil {
                    t.Errorf("Expected parsing to fail for %s, but got %d parameters", tt.paramListInner, len(paramTypes))
                }
                return
            }

            // Verify parameter count
            if len(paramTypes) != len(tt.expectedParamTypes) {
                t.Errorf("Expected %d parameters, got %d", len(tt.expectedParamTypes), len(paramTypes))
                return
            }

            // Verify parameter types
            for i, expectedType := range tt.expectedParamTypes {
                actualType := paramTypes[i]
                if !verifyTypeMatch(actualType, expectedType) {
                    t.Errorf("Parameter %d: expected type %s, got %T", i, expectedType, actualType)
                }
            }

            // Verify symbols were defined in symbol table
            for _, paramName := range tt.expectedParamNames {
                if _, exists := analyzer.symbolTable.Lookup(paramName); !exists {
                    t.Errorf("Parameter %s was not defined in symbol table", paramName)
                }
            }
        })
    }
}

// verifyTypeMatch checks if an actual Type matches the expected type description
func verifyTypeMatch(actual Type, expected string) bool {
    switch expected {
    case "TypeVariable":
        _, ok := actual.(*TypeVariable)
        return ok
    case "int", "string", "bool", "number":
        if tc, ok := actual.(*TypeConstant); ok {
            return tc.Name == expected
        }
        return false
    default:
        // For other types, check if it's a TypeConstant with the expected name
        if tc, ok := actual.(*TypeConstant); ok {
            return tc.Name == expected
        }
        return false
    }
}


