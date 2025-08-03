// Package transpiler provides macro expansion capabilities
package transpiler

import (
	"fmt"
	"regexp"
	"strings"
)

// MacroExpander handles macro expansion for Vex source code
type MacroExpander struct {
	macros      map[string]MacroExpanderFunc
	regexCache  map[string]*regexp.Regexp
}

// MacroExpanderFunc represents a macro that transforms source code
type MacroExpanderFunc func(match string, args []string) (string, error)

// NewMacroExpander creates a new macro expander with built-in macros
func NewMacroExpander() *MacroExpander {
	expander := &MacroExpander{
		macros:     make(map[string]MacroExpanderFunc, 4),  // Pre-allocate for common built-ins
		regexCache: make(map[string]*regexp.Regexp, 4),     // Pre-allocate for common macros
	}
	
	// Register built-in macros
	expander.registerBuiltinMacros()
	
	return expander
}

// registerBuiltinMacros registers the built-in macros
func (me *MacroExpander) registerBuiltinMacros() {
	me.macros["http-server"] = me.expandHttpServer
}

// ExpandMacros expands macros in Vex source code
func (me *MacroExpander) ExpandMacros(source string) (string, error) {
	expanded := source
	
	// Process each registered macro
	for macroName, macroFunc := range me.macros {
		var err error
		expanded, err = me.expandMacroInSource(expanded, macroName, macroFunc)
		if err != nil {
			return "", fmt.Errorf("error expanding macro %s: %w", macroName, err)
		}
	}
	
	return expanded, nil
}

// expandMacroInSource expands a specific macro in the source code
func (me *MacroExpander) expandMacroInSource(source, macroName string, macroFunc MacroExpanderFunc) (string, error) {
	// Check if we have a cached regex for this macro
	re, exists := me.regexCache[macroName]
	if !exists {
		// Compile and cache the regex pattern for this macro
		pattern := `\(` + regexp.QuoteMeta(macroName) + `[^)]*\([^)]*\)[^)]*\)`
		re = regexp.MustCompile(pattern)
		me.regexCache[macroName] = re
	}
	
	// Find all matches
	return re.ReplaceAllStringFunc(source, func(match string) string {
		// Extract arguments from the match
		args := me.extractArguments(match, macroName)
		
		// Expand the macro
		expanded, err := macroFunc(match, args)
		if err != nil {
			// In a real implementation, we'd handle this error properly
			return fmt.Sprintf("/* ERROR: %s */", err.Error())
		}
		
		return expanded
	}), nil
}

// extractArguments extracts arguments from a macro call
func (me *MacroExpander) extractArguments(match, macroName string) []string {
	// Remove the opening paren and macro name
	content := strings.TrimPrefix(match, "("+macroName)
	content = strings.TrimSuffix(content, ")")
	content = strings.TrimSpace(content)
	
	if content == "" {
		return []string{}
	}
	
	// This is a simplified argument parser
	// A full implementation would handle nested s-expressions properly
	args := strings.Fields(content)
	return args
}

// expandHttpServer expands the http-server macro
func (me *MacroExpander) expandHttpServer(match string, args []string) (string, error) {
	// Generate simpler expanded Vex code
	expanded := `(import "net/http")
(import "github.com/gorilla/mux")
(def router (.NewRouter mux))
(.HandleFunc router "/hello" hello-handler)
(.ListenAndServe http ":8080" router)`
	
	return expanded, nil
}