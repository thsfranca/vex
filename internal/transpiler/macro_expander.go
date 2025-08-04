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

// RegisterStringMacro registers a user-defined macro with string template
func (me *MacroExpander) RegisterStringMacro(name string, params []string, template string) {
	me.macros[name] = func(match string, args []string) (string, error) {
		return me.expandStringMacro(name, params, template, args)
	}
}

// expandStringMacro expands a string-based user-defined macro
func (me *MacroExpander) expandStringMacro(name string, params []string, template string, args []string) (string, error) {
	// Debug output
	// Create substitution map
	substitutions := make(map[string]string, len(params))
	
	// Map arguments to parameters
	for i, param := range params {
		if i < len(args) {
			substitutions[param] = args[i]
		}
	}
	
	// Apply substitutions to template
	expanded := template
	for param, value := range substitutions {
		// Handle template syntax: ~param becomes the value
		expanded = strings.ReplaceAll(expanded, "~"+param, value)
	}
	
	// Macro expansion completed successfully
	return expanded, nil
}

// ExpandMacros expands macros in Vex source code using multi-pass expansion
func (me *MacroExpander) ExpandMacros(source string) (string, error) {
	expanded := source
	maxPasses := 10 // Prevent infinite loops
	
	// Start multi-pass expansion
	for pass := 1; pass <= maxPasses; pass++ {
		// Look for new macro definitions in the current expanded text
		newMacros := me.collectNewMacroDefinitions(expanded)
		
		if len(newMacros) > 0 {
			// Register new macros found in this pass
			for name, macroDef := range newMacros {
				me.RegisterStringMacro(name, macroDef.Params, macroDef.Template)
			}
		}
		
		// Remove all macro definitions before expansion
		beforeRemoval := expanded
		expanded = me.removeMacroDefinitions(expanded)
		
		// Expand all registered macros
		hadExpansion := false
		for macroName, macroFunc := range me.macros {
			before := expanded
			var err error
			expanded, err = me.expandMacroInSource(expanded, macroName, macroFunc)
			if err != nil {
				return "", fmt.Errorf("error expanding macro %s in pass %d: %w", macroName, pass, err)
			}
			if expanded != before {
				hadExpansion = true
			}
		}
		
		// If we found no new macros and had no expansions, we're done
		if len(newMacros) == 0 && !hadExpansion && beforeRemoval == expanded {
			break
		}
		
		// Safety check for infinite loops
		if pass == maxPasses {
			return "", fmt.Errorf("macro expansion exceeded maximum passes (%d), possible infinite recursion", maxPasses)
		}
	}
	
	// Multi-pass expansion completed
	return expanded, nil
}

// removeMacroDefinitions removes all (macro ...) definitions from source
func (me *MacroExpander) removeMacroDefinitions(source string) string {
	// Use a simple but effective approach: find (macro and match balanced parentheses
	result := source
	
	for {
		start := strings.Index(result, "(macro ")
		if start == -1 {
			break // No more macro definitions found
		}
		
		// Find the matching closing parenthesis
		depth := 0
		end := start
		for i := start; i < len(result); i++ {
			if result[i] == '(' {
				depth++
			} else if result[i] == ')' {
				depth--
				if depth == 0 {
					end = i + 1
					break
				}
			}
		}
		
		// Remove the macro definition
		result = result[:start] + result[end:]
	}
	
	// Clean up multiple newlines
	result = regexp.MustCompile(`\n\s*\n\s*\n`).ReplaceAllString(result, "\n\n")
	
	// Macro definitions removed successfully
	return result
}

// collectNewMacroDefinitions finds and extracts new macro definitions from expanded text
func (me *MacroExpander) collectNewMacroDefinitions(source string) map[string]*ParsedMacroDefinition {
	newMacros := make(map[string]*ParsedMacroDefinition)
	
	for {
		start := strings.Index(source, "(macro ")
		if start == -1 {
			break // No more macro definitions found
		}
		
		// Find the matching closing parenthesis
		depth := 0
		end := start
		for i := start; i < len(source); i++ {
			if source[i] == '(' {
				depth++
			} else if source[i] == ')' {
				depth--
				if depth == 0 {
					end = i + 1
					break
				}
			}
		}
		
		// Extract and parse the macro definition
		macroDefText := source[start:end]
		if macroDef := me.parseMacroDefinition(macroDefText); macroDef != nil {
			newMacros[macroDef.Name] = macroDef
		}
		
		// Move past this macro definition
		source = source[end:]
	}
	
	return newMacros
}

// ParsedMacroDefinition represents a parsed macro definition from text
type ParsedMacroDefinition struct {
	Name     string
	Params   []string
	Template string
}

// parseMacroDefinition parses a single macro definition from text
func (me *MacroExpander) parseMacroDefinition(macroText string) *ParsedMacroDefinition {
	// Simple parsing: (macro name [params] template)
	// Remove outer parentheses
	if len(macroText) < 8 || !strings.HasPrefix(macroText, "(macro ") {
		return nil
	}
	
	content := strings.TrimSpace(macroText[7 : len(macroText)-1]) // Remove "(macro " and ")"
	
	// Split into parts
	parts := strings.Fields(content)
	if len(parts) < 3 {
		return nil
	}
	
	name := parts[0]
	
	// Find parameter list [...]
	paramStart := strings.Index(content, "[")
	paramEnd := strings.Index(content, "]")
	if paramStart == -1 || paramEnd == -1 || paramEnd <= paramStart {
		return nil
	}
	
	paramStr := content[paramStart+1 : paramEnd]
	params := []string{}
	if strings.TrimSpace(paramStr) != "" {
		params = strings.Fields(paramStr)
	}
	
	// Template is everything after the parameter list
	template := strings.TrimSpace(content[paramEnd+1:])
	
	return &ParsedMacroDefinition{
		Name:     name,
		Params:   params,
		Template: template,
	}
}

// expandMacroInSource expands a specific macro in the source code
func (me *MacroExpander) expandMacroInSource(source, macroName string, macroFunc MacroExpanderFunc) (string, error) {
	// Check if we have a cached regex for this macro
	re, exists := me.regexCache[macroName]
	if !exists {
		// Compile and cache the regex pattern for this macro
		// Pattern to match: (macroName ... with balanced parentheses/brackets ...)
		pattern := `\(` + regexp.QuoteMeta(macroName) + `(?:[^()]|\([^()]*\)|\[[^\]]*\])*\)`
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
	
	// Parse arguments respecting nested parentheses and brackets
	return me.parseArguments(content)
}

// parseArguments parses macro arguments respecting nested structures
func (me *MacroExpander) parseArguments(content string) []string {
	var args []string
	var current strings.Builder
	depth := 0
	inString := false
	
	for i, char := range content {
		switch char {
		case '"':
			// Toggle string state (ignoring escaped quotes for simplicity)
			if i == 0 || content[i-1] != '\\' {
				inString = !inString
			}
			current.WriteRune(char)
			
		case '(', '[':
			if !inString {
				depth++
			}
			current.WriteRune(char)
			
		case ')', ']':
			if !inString {
				depth--
			}
			current.WriteRune(char)
			
		case ' ', '\t', '\n':
			if !inString && depth == 0 {
				// We're at the top level and hit whitespace - end current argument
				if current.Len() > 0 {
					args = append(args, strings.TrimSpace(current.String()))
					current.Reset()
				}
			} else {
				current.WriteRune(char)
			}
			
		default:
			current.WriteRune(char)
		}
	}
	
	// Add the last argument
	if current.Len() > 0 {
		args = append(args, strings.TrimSpace(current.String()))
	}
	
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