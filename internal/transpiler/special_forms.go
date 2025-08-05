package transpiler

import (
	"fmt"
	"strings"
)

// handleDefinition processes variable definitions: (def name value)
func (t *Transpiler) handleDefinition(args []string) {
	if len(args) < 2 {
		t.output.WriteString("// Error: def requires name and value\n")
		return
	}

	name := args[0]
	value := args[1]
	
	// Check if this is defining a variable that was called as a function
	t.output.WriteString(fmt.Sprintf("var %s = %s\n", name, value))
	t.output.WriteString(fmt.Sprintf("_ = %s // Use variable to satisfy Go compiler\n", name))
}

// handleImport processes import statements: (import "package")
func (t *Transpiler) handleImport(args []string) {
	if len(args) < 1 {
		t.output.WriteString("// Error: import requires package name\n")
		return
	}

	importPath := strings.Trim(args[0], "\"")
	t.imports[importPath] = true
	
	// Check if this is a third-party module
	if t.isThirdPartyModule(importPath) {
		// For now, we'll detect common modules. In a real implementation,
		// you might want to use go mod to resolve dependencies
		t.goModules[importPath] = "latest"
	}
}

// isThirdPartyModule checks if an import is a third-party module
func (t *Transpiler) isThirdPartyModule(importPath string) bool {
	// Standard library packages don't need to be added to go.mod
	standardLibs := map[string]bool{
		"fmt": true, "os": true, "strings": true, "strconv": true,
		"time": true, "net/http": true, "encoding/json": true,
		"io": true, "bufio": true, "regexp": true, "sort": true,
		"math": true, "crypto": true, "errors": true, "context": true,
	}
	
	return !standardLibs[importPath] && !strings.HasPrefix(importPath, "github.com/thsfranca/vex")
}

// handleArithmetic processes arithmetic operations: (+ 1 2) (- 3 1) etc.
func (t *Transpiler) handleArithmetic(op string, args []string) {
	if len(args) < 2 {
		t.output.WriteString(fmt.Sprintf("// Error: %s requires at least 2 arguments\n", op))
		return
	}

	// Left-associative chaining for multiple operands
	result := "(" + args[0] + " " + op + " " + args[1] + ")"
	for i := 2; i < len(args); i++ {
		result = "(" + result + " " + op + " " + args[i] + ")"
	}

	t.output.WriteString("_ = " + result + "\n")
}

// handleFunctionCall processes regular function calls
func (t *Transpiler) handleFunctionCall(funcName string, args []string) {
	// Check if this is a macro call
	if macro, exists := t.macros[funcName]; exists {
		t.expandMacro(macro, args)
		return
	}

	// Handle package/function notation (fmt/Println -> fmt.Println)
	if strings.Contains(funcName, "/") {
		funcName = strings.ReplaceAll(funcName, "/", ".")
	}

	argStr := strings.Join(args, ", ")
	
	// For functions that return multiple values or don't need assignment, just call them
	if strings.Contains(funcName, "Println") || strings.Contains(funcName, "Printf") {
		t.output.WriteString(fmt.Sprintf("%s(%s)\n", funcName, argStr))
	} else {
		t.output.WriteString(fmt.Sprintf("_ = %s(%s)\n", funcName, argStr))
	}
}

// handleIf processes if conditionals: (if condition then-expr else-expr)
func (t *Transpiler) handleIf(args []string) {
	if len(args) < 2 {
		t.output.WriteString("// Error: if requires at least condition and then-expr\n")
		return
	}

	condition := args[0]
	thenExpr := args[1]
	var elseExpr string
	if len(args) >= 3 {
		elseExpr = args[2]
	} else {
		elseExpr = "nil"
	}

	// Generate an IIFE (Immediately Invoked Function Expression) for the if
	result := fmt.Sprintf("func() interface{} { if %s { return %s } else { return %s } }()", condition, thenExpr, elseExpr)
	t.output.WriteString("_ = " + result + "\n")
}

// handleDo processes do blocks: (do expr1 expr2 expr3)
func (t *Transpiler) handleDo(args []string) {
	for _, expr := range args {
		t.output.WriteString(expr + "\n")
	}
}

// handleLambda processes lambda function definitions: (fn [params] body)
func (t *Transpiler) handleLambda(args []string) {
	if len(args) < 2 {
		t.output.WriteString("// Error: fn requires parameter list and body\n")
		return
	}

	// The first argument is the parameter list which comes as transpiled array syntax
	paramList := args[0]
	body := args[1]

	// Extract parameter names from the transpiled array syntax
	params := t.parseParameterList(paramList)
	
	// Generate Go function with interface{} parameters
	paramDecls := make([]string, len(params))
	for i, param := range params {
		paramDecls[i] = fmt.Sprintf("%s interface{}", param)
	}
	
	paramString := strings.Join(paramDecls, ", ")
	
	// Generate Go function literal
	functionLiteral := fmt.Sprintf("func(%s) interface{} { return %s }", paramString, body)
	
	t.output.WriteString("_ = " + functionLiteral + "\n")
}