package analysis

import (
	"testing"
)

func TestAnalyzer_FunctionWithTypeAnnotations(t *testing.T) {
    tests := []struct {
        name                string
        paramList           string
        returnTypeAnnotation string
        bodyExpression      string
        expectedParamCount  int
    }{
        {
            name:                "Function with explicit parameter types and return type",
            paramList:           "[x: int y: int]",
            returnTypeAnnotation: "->",
            bodyExpression:      "int",
            expectedParamCount:  2,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            analyzer := NewAnalyzer()
            
            // Create arguments for analyzeFn
            var args []Value
            args = append(args, NewBasicValue(tt.paramList, "array"))
            
            if tt.returnTypeAnnotation != "" {
                args = append(args, NewBasicValue(tt.returnTypeAnnotation, "symbol"))
                args = append(args, NewBasicValue("int", "symbol")) // return type
                args = append(args, NewBasicValue(tt.bodyExpression, "expression"))
            } else {
                args = append(args, NewBasicValue(tt.bodyExpression, "expression"))
            }
            
            // Create mock context
            ctx := createMockListNode("fn")
            
            // Analyze the function
            result, err := analyzer.analyzeFn(ctx, args)
            
            if err != nil {
                t.Fatalf("analyzeFn failed: %v", err)
            }
            
            // Verify the result is a function
            if result == nil {
                t.Fatal("analyzeFn returned nil result")
            }

            if result.Type() != "func" {
                t.Errorf("Expected function type, got %s", result.Type())
            }

            // Get the internal type and verify structure
            if bv, ok := result.(*BasicValue); ok && bv.getType() != nil {
                if fnType, ok := bv.getType().(*TypeFunction); ok {
                    // Check parameter count
                    if len(fnType.Params) != tt.expectedParamCount {
                        t.Errorf("Expected %d parameters, got %d", tt.expectedParamCount, len(fnType.Params))
                    }
                    
                    // Check explicit return type (all return types must be explicit now)
                    if _, ok := fnType.Result.(*TypeConstant); !ok {
                        t.Errorf("Expected explicit return type (TypeConstant), got %T", fnType.Result)
                    }
                } else {
                    t.Errorf("Expected TypeFunction, got %T", bv.getType())
                }
            } else {
                t.Error("Could not get internal type from result")
            }
            
            // Check for errors
            if analyzer.errorReporter.HasErrors() {
                errors := analyzer.errorReporter.GetErrors()
                t.Errorf("Unexpected analysis errors: %v", errors)
            }
        })
    }
}


