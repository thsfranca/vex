package transpiler

import (
	"fmt"
	"strings"

	"github.com/thsfranca/vex/internal/transpiler/parser"
)

// evaluateExpression evaluates list expressions that return values
func (t *Transpiler) evaluateExpression(ctx *parser.ListContext) string {
	childCount := ctx.GetChildCount()
	if childCount < 3 { // Need at least: '(', function, ')'
		return "nil"
	}

	// Get function name (first child after '(')
	funcNameNode := ctx.GetChild(1)
	funcName := t.visitNode(funcNameNode)

	// Extract arguments
	args := make([]string, 0)
	for i := 2; i < childCount-1; i++ { // Skip '(' and ')'
		child := ctx.GetChild(i)
		if child != nil {
			args = append(args, t.visitNode(child))
		}
	}

	// Handle special expressions
	switch funcName {
	case "if":
		return t.evaluateIf(args)
	case "do":
		return t.evaluateDo(args)
	case "fn":
		return t.evaluateLambda(args)
	case "first", "rest", "cons", "count", "empty?":
		return t.evaluateCollectionOp(funcName, args)
	case "+", "-", "*", "/":
		if len(args) < 2 {
			return "0"
		}
		// Left-associative chaining for multiple operands
		result := "(" + args[0] + " " + funcName + " " + args[1] + ")"
		for i := 2; i < len(args); i++ {
			result = "(" + result + " " + funcName + " " + args[i] + ")"
		}
		return result
	default:
		// Check if this is a macro call
		if _, exists := t.macros[funcName]; exists {
			// For macro calls in expressions, we need to handle them differently
			// This is a simplified approach - ideally we'd evaluate the macro to an expression
			return fmt.Sprintf("/* macro %s expansion not supported in expressions yet */", funcName)
		}

		// Handle package/function notation
		if strings.Contains(funcName, "/") {
			funcName = strings.ReplaceAll(funcName, "/", ".")
		}
		
		argStr := strings.Join(args, ", ")
		return fmt.Sprintf("%s(%s)", funcName, argStr)
	}
}

// evaluateIf evaluates if expressions: (if condition then-expr else-expr)
func (t *Transpiler) evaluateIf(args []string) string {
	if len(args) < 2 {
		return "nil"
	}

	condition := args[0]
	thenExpr := args[1]
	var elseExpr string
	if len(args) >= 3 {
		elseExpr = args[2]
	} else {
		elseExpr = "nil"
	}

	return fmt.Sprintf("func() interface{} { if %s { return %s } else { return %s } }()", condition, thenExpr, elseExpr)
}

// evaluateDo evaluates do blocks as expressions (returns last expression)
func (t *Transpiler) evaluateDo(args []string) string {
	if len(args) == 0 {
		return "nil"
	}
	
	// Generate function that executes all expressions and returns the last one
	var statements []string
	for i, expr := range args {
		if i == len(args)-1 {
			statements = append(statements, "return "+expr)
		} else {
			statements = append(statements, expr)
		}
	}
	
	return fmt.Sprintf("func() interface{} { %s }()", strings.Join(statements, "; "))
}

// evaluateLambda evaluates lambda functions as expressions  
func (t *Transpiler) evaluateLambda(args []string) string {
	if len(args) < 2 {
		return "nil"
	}

	// Parse parameter list from first argument
	paramList := args[0]
	body := args[1]

	// Extract parameter names from [param1 param2] syntax
	params := t.parseParameterList(paramList)
	
	// Generate Go function with interface{} parameters
	paramDecls := make([]string, len(params))
	for i, param := range params {
		paramDecls[i] = fmt.Sprintf("%s interface{}", param)
	}
	
	paramString := strings.Join(paramDecls, ", ")
	
	// Return Go function literal as expression
	return fmt.Sprintf("func(%s) interface{} { return %s }", paramString, body)
}

// parseParameterList extracts parameter names from [param1 param2] syntax
func (t *Transpiler) parseParameterList(paramList string) []string {
	// The paramList comes from visitNode on an array, so it might be in the form:
	// "[]interface{}{param1, param2}" or just "param1 param2" or empty
	
	// Handle array syntax from transpiled arrays
	if strings.HasPrefix(paramList, "[]interface{}") {
		// Find the content braces (after the type declaration)
		// Format is: []interface{}{param1, param2}
		typeEnd := strings.Index(paramList, "}{")
		if typeEnd != -1 {
			// Extract content after }{
			start := typeEnd + 2
			end := strings.LastIndex(paramList, "}")
			if end != -1 && end > start {
				content := paramList[start:end]
				content = strings.TrimSpace(content)
				if content == "" {
					return []string{}
				}
				// Split by comma and clean up
				parts := strings.Split(content, ",")
				params := make([]string, 0, len(parts))
				for _, part := range parts {
					param := strings.TrimSpace(part)
					if param != "" {
						params = append(params, param)
					}
				}
				return params
			}
		}
		return []string{}
	}
	
	// Handle simple bracket syntax like [x y] (if somehow we get this)
	paramList = strings.Trim(paramList, "[]")
	paramList = strings.TrimSpace(paramList)
	
	if paramList == "" {
		return []string{}
	}
	
	// Split by whitespace and filter empty strings
	parts := strings.Fields(paramList)
	params := make([]string, 0, len(parts))
	for _, part := range parts {
		if part != "" {
			params = append(params, part)
		}
	}
	
	return params
}