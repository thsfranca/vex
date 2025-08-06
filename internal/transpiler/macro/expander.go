package macro

import (
	"fmt"
	"strings"

	"github.com/antlr4-go/antlr/v4"
	"github.com/thsfranca/vex/internal/transpiler/parser"
)

// Expander handles macro expansion in AST
type Expander struct {
	registry *Registry
}

// NewExpander creates a new macro expander
func NewExpander(registry *Registry) *Expander {
	return &Expander{
		registry: registry,
	}
}

// ExpandMacro expands a single macro call
func (e *Expander) ExpandMacro(macroName string, args []string) (string, error) {
	macro, exists := e.registry.GetMacro(macroName)
	if !exists {
		return "", fmt.Errorf("macro '%s' not found", macroName)
	}

	if len(args) != len(macro.Params) {
		return "", fmt.Errorf("macro '%s' expects %d arguments, got %d", 
			macroName, len(macro.Params), len(args))
	}

	// Substitute parameters in macro body
	expanded := e.substituteParameters(macro.Body, macro.Params, args)
	
	// Recursively expand any nested macros
	return e.recursivelyExpandMacros(expanded)
}

// substituteParameters substitutes macro parameters with arguments
func (e *Expander) substituteParameters(body string, params []string, args []string) string {
	// Create parameter mapping with recursive expansion of arguments
	paramMap := make(map[string]string)
	for i, param := range params {
		if i < len(args) {
			// Recursively expand any macros in the argument
			expandedArg, _ := e.recursivelyExpandMacros(args[i])
			paramMap[param] = expandedArg
		}
	}

	// Handle simple parameter substitution
	if len(params) == 1 && strings.TrimSpace(body) == params[0] {
		return paramMap[params[0]]
	}

	// Try AST-based substitution for complex expressions
	return e.substituteParametersInAST(body, paramMap)
}

// substituteParametersInAST performs AST-aware parameter substitution
func (e *Expander) substituteParametersInAST(body string, paramMap map[string]string) string {
	// Try to parse the body as an expression
	inputStream := antlr.NewInputStream(body)
	lexer := parser.NewVexLexer(inputStream)
	tokenStream := antlr.NewCommonTokenStream(lexer, 0)
	vexParser := parser.NewVexParser(tokenStream)

	// Try parsing as different constructs
	if tree := vexParser.Program(); tree != nil {
		if programCtx, ok := tree.(*parser.ProgramContext); ok {
			if len(programCtx.GetChildren()) > 0 {
				substituted := e.substituteInNode(programCtx.GetChildren()[0], paramMap)
				return e.reconstructVexSyntax(substituted)
			}
		}
	}

	// Fallback to word-boundary replacement
	result := body
	for param, arg := range paramMap {
		result = e.replaceWholeWord(result, param, arg)
	}
	
	return result
}

// substituteInNode recursively substitutes parameters in an AST node
func (e *Expander) substituteInNode(node antlr.Tree, paramMap map[string]string) antlr.Tree {
	switch n := node.(type) {
	case *parser.ListContext:
		return e.substituteInList(n, paramMap)
	case *parser.ArrayContext:
		return e.substituteInArray(n, paramMap)
	case *antlr.TerminalNodeImpl:
		if replacement, exists := paramMap[n.GetText()]; exists {
			return &substitutedTerminal{text: replacement}
		}
		return n
	case antlr.TerminalNode:
		if replacement, exists := paramMap[n.GetText()]; exists {
			return &substitutedTerminal{text: replacement}
		}
		return n
	default:
		return n
	}
}

// substituteInList handles substitution in list contexts
func (e *Expander) substituteInList(ctx *parser.ListContext, paramMap map[string]string) antlr.Tree {
	// Reconstruct the list with substituted children
	var parts []string
	for _, child := range ctx.GetChildren() {
		if child != nil {
			substituted := e.substituteInNode(child, paramMap)
			parts = append(parts, e.reconstructVexSyntax(substituted))
		}
	}
	
	// Use smart spacing like in reconstructVexSyntax
	reconstructed := ""
	for i, part := range parts {
		if i == 0 {
			reconstructed += part // First part (usually '(')
		} else if part == ")" {
			reconstructed += part // Last part, no space before
		} else if parts[i-1] == "(" {
			reconstructed += part // No space after opening parenthesis
		} else {
			reconstructed += " " + part // Add space before other parts
		}
	}
	
	// Parse the reconstructed expression back into an AST node
	inputStream := antlr.NewInputStream(reconstructed)
	lexer := parser.NewVexLexer(inputStream)
	tokenStream := antlr.NewCommonTokenStream(lexer, 0)
	vexParser := parser.NewVexParser(tokenStream)
	
	if listTree := vexParser.List(); listTree != nil {
		return listTree.(*parser.ListContext)
	}
	
	return ctx // Fallback
}

// substituteInArray handles substitution in array contexts
func (e *Expander) substituteInArray(ctx *parser.ArrayContext, paramMap map[string]string) antlr.Tree {
	// Similar to substituteInList but for arrays
	var parts []string
	for _, child := range ctx.GetChildren() {
		if child != nil {
			substituted := e.substituteInNode(child, paramMap)
			parts = append(parts, e.reconstructVexSyntax(substituted))
		}
	}
	
	reconstructed := strings.Join(parts, " ")
	
	// Parse back into array
	inputStream := antlr.NewInputStream(reconstructed)
	lexer := parser.NewVexLexer(inputStream)
	tokenStream := antlr.NewCommonTokenStream(lexer, 0)
	vexParser := parser.NewVexParser(tokenStream)
	
	if arrayTree := vexParser.Array(); arrayTree != nil {
		return arrayTree.(*parser.ArrayContext)
	}
	
	return ctx // Fallback
}

// recursivelyExpandMacros expands nested macros in an expression
func (e *Expander) recursivelyExpandMacros(expr string) (string, error) {
	// Check if the expression contains macro calls
	if !strings.Contains(expr, "(") {
		return expr, nil // No function calls, return as-is
	}

	// For simple cases, check if the entire expression is a known macro
	trimmed := strings.TrimSpace(expr)
	if strings.HasPrefix(trimmed, "(") && strings.HasSuffix(trimmed, ")") {
		// Extract the parts manually for simple cases
		inner := strings.TrimSpace(trimmed[1 : len(trimmed)-1])
		parts := strings.Fields(inner)

		if len(parts) >= 1 {
			macroName := parts[0]
			if e.registry.HasMacro(macroName) {
				// Extract arguments
				args := parts[1:]

				// Recursively expand arguments first
				expandedArgs := make([]string, len(args))
				for i, arg := range args {
					expanded, err := e.recursivelyExpandMacros(arg)
					if err != nil {
						return expr, err
					}
					expandedArgs[i] = expanded
				}

				// Expand this macro
				result, err := e.ExpandMacro(macroName, expandedArgs)
				if err != nil {
					return expr, err
				}

				// Check if the result contains more macros
				if result != expr && strings.Contains(result, "(") {
					return e.recursivelyExpandMacros(result)
				}

				return result, nil
			}
		}
	}

	return expr, nil
}

// replaceWholeWord replaces whole words with word boundaries
func (e *Expander) replaceWholeWord(text, param, arg string) string {
	result := ""
	i := 0
	for i < len(text) {
		// Find the next occurrence of the parameter
		index := strings.Index(text[i:], param)
		if index == -1 {
			result += text[i:]
			break
		}
		
		// Adjust index to absolute position
		index += i
		
		// Check if it's a whole word (not part of another identifier)
		isWholeWord := true
		
		// Check character before
		if index > 0 {
			prevChar := rune(text[index-1])
			if e.isIdentifierChar(prevChar) {
				isWholeWord = false
			}
		}
		
		// Check character after
		if index+len(param) < len(text) {
			nextChar := rune(text[index+len(param)])
			if e.isIdentifierChar(nextChar) {
				isWholeWord = false
			}
		}
		
		if isWholeWord {
			// Replace the parameter
			result += text[i:index] + arg
			i = index + len(param)
		} else {
			// Not a whole word, skip this occurrence
			result += text[i:index+1]
			i = index + 1
		}
	}
	
	return result
}

// isIdentifierChar checks if a character can be part of an identifier
func (e *Expander) isIdentifierChar(r rune) bool {
	return (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || 
		   (r >= '0' && r <= '9') || r == '_' || r == '-' || r == '?' || r == '!'
}

// reconstructVexSyntax reconstructs Vex syntax from an AST node
func (e *Expander) reconstructVexSyntax(node antlr.Tree) string {
	switch n := node.(type) {
	case *parser.ListContext:
		var parts []string
		for _, child := range n.GetChildren() {
			if child != nil {
				parts = append(parts, e.reconstructVexSyntax(child))
			}
		}
		// Smart spacing: spaces between tokens but not around parentheses
		result := ""
		for i, part := range parts {
			if i == 0 {
				result += part // First part (usually '(')
			} else if part == ")" {
				result += part // Last part, no space before
			} else if parts[i-1] == "(" {
				result += part // No space after opening parenthesis
			} else {
				result += " " + part // Add space before other parts
			}
		}
		return result
	case *parser.ArrayContext:
		var parts []string
		for _, child := range n.GetChildren() {
			if child != nil {
				parts = append(parts, e.reconstructVexSyntax(child))
			}
		}
		return strings.Join(parts, " ")
	case *parser.ProgramContext:
		if len(n.GetChildren()) > 0 {
			return e.reconstructVexSyntax(n.GetChildren()[0])
		}
		return ""
	case *substitutedTerminal:
		return n.GetText()
	case *antlr.TerminalNodeImpl:
		return n.GetText()
	case antlr.TerminalNode:
		return n.GetText()
	default:
		return ""
	}
}

// substitutedTerminal represents a terminal node with substituted text
type substitutedTerminal struct {
	text string
}

func (s *substitutedTerminal) GetText() string { return s.text }
func (s *substitutedTerminal) GetSymbol() antlr.Token { return nil }
func (s *substitutedTerminal) Accept(visitor antlr.ParseTreeVisitor) interface{} { return nil }
func (s *substitutedTerminal) GetChild(i int) antlr.Tree { return nil }
func (s *substitutedTerminal) GetChildCount() int { return 0 }
func (s *substitutedTerminal) GetChildren() []antlr.Tree { return nil }
func (s *substitutedTerminal) GetParent() antlr.Tree { return nil }
func (s *substitutedTerminal) GetPayload() interface{} { return s.text }
func (s *substitutedTerminal) GetSourceInterval() antlr.Interval { return antlr.Interval{Start: -1, Stop: -1} }
func (s *substitutedTerminal) SetParent(parent antlr.Tree) {}
func (s *substitutedTerminal) ToStringTree(ruleNames []string, recog antlr.Recognizer) string { return s.text }