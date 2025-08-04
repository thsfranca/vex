// Package transpiler provides dynamic macro registration
package transpiler

import (
	"fmt"
	"strings"

	"github.com/antlr4-go/antlr/v4"
)

// MacroDefinition represents a user-defined macro
type MacroDefinition struct {
	Name       string
	Parameters []string
	Template   antlr.Tree // The macro body template
}

// MacroRegistry stores dynamically registered macros
type MacroRegistry struct {
	macros map[string]*MacroDefinition
}

// NewMacroRegistry creates a new macro registry
func NewMacroRegistry() *MacroRegistry {
	return &MacroRegistry{
		macros: make(map[string]*MacroDefinition, 8), // Pre-allocate for common user macros
	}
}

// RegisterMacro adds a new macro to the registry
func (mr *MacroRegistry) RegisterMacro(name string, params []string, template antlr.Tree) {
	mr.macros[name] = &MacroDefinition{
		Name:       name,
		Parameters: params,
		Template:   template,
	}
}

// GetMacro retrieves a macro by name
func (mr *MacroRegistry) GetMacro(name string) (*MacroDefinition, bool) {
	macro, exists := mr.macros[name]
	return macro, exists
}

// IsMacro checks if a name is a registered macro
func (mr *MacroRegistry) IsMacro(name string) bool {
	_, exists := mr.macros[name]
	return exists
}

// ExpandMacro expands a macro call with given arguments
func (mr *MacroRegistry) ExpandMacro(name string, args []antlr.Tree) (string, error) {
	macro, exists := mr.macros[name]
	if !exists {
		return "", fmt.Errorf("macro %s not found", name)
	}

	// Create substitution map
	substitutions := make(map[string]string, len(macro.Parameters)) // Pre-allocate for exact parameter count
	
	// Map arguments to parameters
	for i, param := range macro.Parameters {
		if i < len(args) {
			// Convert argument to string representation
			argStr := nodeToString(args[i])
			substitutions[param] = argStr
		}
	}

	// Expand the template with substitutions
	expanded := expandTemplate(macro.Template, substitutions)
	
	return expanded, nil
}

// nodeToString converts an AST node to its string representation
func nodeToString(node antlr.Tree) string {
	if terminalNode, ok := node.(*antlr.TerminalNodeImpl); ok {
		return terminalNode.GetText()
	}
	
	// For complex nodes, reconstruct the source
	return reconstructSource(node)
}

// reconstructSource reconstructs source code from AST node
func reconstructSource(node antlr.Tree) string {
	switch ctx := node.(type) {
	case *antlr.TerminalNodeImpl:
		return ctx.GetText()
	default:
		// For lists and arrays, reconstruct with proper brackets
		children := ctx.GetChildren()
		if len(children) == 0 {
			return ""
		}
		
		var result strings.Builder
		lastWasBracket := false
		
		for i, child := range children {
			// Check if child is a terminal node before calling GetText
			childText := ""
			if terminalChild, ok := child.(*antlr.TerminalNodeImpl); ok {
				childText = terminalChild.GetText()
			}
			
			// Add space before non-bracket tokens, but not if:
			// - This is the first token
			// - The previous token was an opening bracket
			// - This token is a closing bracket
			isBracket := childText == "(" || childText == ")" || childText == "[" || childText == "]"
			
			if i > 0 && !lastWasBracket && !isBracket {
				result.WriteString(" ")
			}
			
			result.WriteString(reconstructSource(child))
			lastWasBracket = (childText == "(" || childText == "[")
		}
		return result.String()
	}
}

// expandTemplate expands a template with parameter substitutions
func expandTemplate(template antlr.Tree, substitutions map[string]string) string {
	// For now, simple string-based substitution
	// A full implementation would handle proper AST transformation
	templateStr := reconstructSource(template)
	
	// Apply substitutions
	for param, value := range substitutions {
		// Handle template syntax: ~param becomes the value
		templateStr = strings.ReplaceAll(templateStr, "~"+param, value)
	}
	
	return templateStr
}