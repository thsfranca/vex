// Package transpiler provides macro collection for user-defined macros
package transpiler

import (
	"fmt"

	"github.com/antlr4-go/antlr/v4"
	"github.com/thsfranca/vex/internal/transpiler/parser"
)

// MacroCollector finds and registers user-defined macros before macro expansion
type MacroCollector struct {
	expander *MacroExpander
}

// NewMacroCollector creates a new macro collector
func NewMacroCollector(expander *MacroExpander) *MacroCollector {
	return &MacroCollector{
		expander: expander,
	}
}

// CollectMacros walks the AST to find and register macro definitions
func (mc *MacroCollector) CollectMacros(programCtx *parser.ProgramContext) error {
	for _, child := range programCtx.GetChildren() {
		if listCtx, ok := child.(*parser.ListContext); ok {
			err := mc.collectFromList(listCtx)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// collectFromList processes a list expression looking for macro definitions
func (mc *MacroCollector) collectFromList(listCtx *parser.ListContext) error {
	children := listCtx.GetChildren()
	if len(children) < 3 { // '(' ... ')'
		return nil
	}

	content := children[1 : len(children)-1]
	if len(content) == 0 {
		return nil
	}

	// Check if this is a macro definition
	firstChild := content[0]
	var firstElement string
	if symbolCtx, ok := firstChild.(*antlr.TerminalNodeImpl); ok {
		firstElement = symbolCtx.GetText()
	}

	if firstElement == "macro" {
		return mc.processMacroDefinition(content[1:]) // Skip "macro" keyword
	}

	// Recursively process nested lists
	for _, child := range content {
		if nestedList, ok := child.(*parser.ListContext); ok {
			err := mc.collectFromList(nestedList)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// processMacroDefinition processes a macro definition and registers it with the expander
func (mc *MacroCollector) processMacroDefinition(content []antlr.Tree) error {
	if len(content) < 3 {
		// Invalid macro definition, but don't fail - let the semantic analyzer handle the error
		return nil
	}

	// Get macro name
	var macroName string
	if symbolCtx, ok := content[0].(*antlr.TerminalNodeImpl); ok {
		macroName = symbolCtx.GetText()
	} else {
		return nil // Invalid, but continue processing
	}

	// Get parameters from array
	paramList := content[1]
	var params []string
	if arrayCtx, ok := paramList.(*parser.ArrayContext); ok {
		children := arrayCtx.GetChildren()
		// Skip [ and ] brackets
		for i := 1; i < len(children)-1; i++ {
			if terminalNode, ok := children[i].(*antlr.TerminalNodeImpl); ok {
				params = append(params, terminalNode.GetText())
			}
		}
	} else {
		return nil // Invalid parameters, continue processing
	}

	// Get macro body template
	if len(content) > 2 {
		template := content[2]
		templateStr := mc.reconstructTemplate(template)

		// Debug: print macro registration
		fmt.Printf("DEBUG: Registering macro '%s' with params %v and template: %s\n", macroName, params, templateStr)

		// Register the macro with the expander using a string-based expansion function
		mc.expander.RegisterStringMacro(macroName, params, templateStr)
	}

	return nil
}

// reconstructTemplate converts an AST node back to its string representation
func (mc *MacroCollector) reconstructTemplate(node antlr.Tree) string {
	return reconstructSource(node) // Reuse the existing function from macro_registry.go
}

// RegisterStringMacro is a method we need to add to MacroExpander for string-based macros
// This will be added to the MacroExpander struct
