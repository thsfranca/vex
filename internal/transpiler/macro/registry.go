package macro

import (
	"fmt"
	"os"
	"strings"

	"github.com/antlr4-go/antlr/v4"
	"github.com/thsfranca/vex/internal/transpiler/parser"
)

// Macro represents a registered macro with its parameters and body
type Macro struct {
	Name   string
	Params []string
	Body   string
}

// Registry manages macro registration and expansion
type Registry struct {
	macros     map[string]*Macro
	coreLoaded bool
	config     Config
}

// Config holds macro system configuration
type Config struct {
	CoreMacroPath    string
	EnableFallback   bool
	EnableValidation bool
}

// NewRegistry creates a new macro registry
func NewRegistry(config Config) *Registry {
	return &Registry{
		macros: make(map[string]*Macro),
		config: config,
	}
}

// RegisterMacro registers a new macro
func (r *Registry) RegisterMacro(name string, macro *Macro) error {
	if r.config.EnableValidation {
		if err := r.validateMacro(name, macro); err != nil {
			return err
		}
	}
	
	r.macros[name] = macro
	return nil
}

// HasMacro checks if a macro is registered
func (r *Registry) HasMacro(name string) bool {
	_, exists := r.macros[name]
	return exists
}

// GetMacro retrieves a registered macro
func (r *Registry) GetMacro(name string) (*Macro, bool) {
	macro, exists := r.macros[name]
	return macro, exists
}

// LoadCoreMacros loads core macros from file or fallback
func (r *Registry) LoadCoreMacros() error {
	if r.coreLoaded {
		return nil
	}
	r.coreLoaded = true

	// Try multiple possible paths for core.vx
	corePaths := []string{
		r.config.CoreMacroPath,
		"core/core.vx",           // From project root
		"../../core/core.vx",     // From test directory
		"../../../core/core.vx",  // From deeper test directories
	}

	for _, path := range corePaths {
		if path != "" {
			if content, err := os.ReadFile(path); err == nil {
				return r.loadMacrosFromContent(string(content))
			}
		}
	}

	// Fallback to built-in macros
	if r.config.EnableFallback {
		return r.loadFallbackMacros()
	}

	return fmt.Errorf("no core macros available - tried paths: %v", corePaths)
}

// loadMacrosFromContent parses macro definitions from Vex code
func (r *Registry) loadMacrosFromContent(content string) error {
	// Temporarily disable validation for core macros
	originalValidation := r.config.EnableValidation
	r.config.EnableValidation = false
	defer func() {
		r.config.EnableValidation = originalValidation
	}()
	
	// Create a minimal parser for macro definitions
	inputStream := antlr.NewInputStream(content)
	lexer := parser.NewVexLexer(inputStream)
	tokenStream := antlr.NewCommonTokenStream(lexer, 0)
	vexParser := parser.NewVexParser(tokenStream)

	tree := vexParser.Program()
	if tree == nil {
		return fmt.Errorf("failed to parse macro definitions")
	}

	// Extract macro definitions
	return r.extractMacrosFromAST(tree.(*parser.ProgramContext))
}

// extractMacrosFromAST extracts macro definitions from parsed AST
func (r *Registry) extractMacrosFromAST(ctx *parser.ProgramContext) error {
	for _, child := range ctx.GetChildren() {
		if listCtx, ok := child.(*parser.ListContext); ok {
			if r.isMacroDefinition(listCtx) {
				if err := r.parseMacroDefinition(listCtx); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

// isMacroDefinition checks if a list is a macro definition
func (r *Registry) isMacroDefinition(listCtx *parser.ListContext) bool {
	children := listCtx.GetChildren()
	if len(children) < 2 {
		return false
	}
	
	if terminal, ok := children[1].(*antlr.TerminalNodeImpl); ok {
		return terminal.GetText() == "macro"
	}
	
	return false
}

// parseMacroDefinition parses a macro definition from AST
func (r *Registry) parseMacroDefinition(listCtx *parser.ListContext) error {
	children := listCtx.GetChildren()
	if len(children) < 6 { // ( macro name [params] body )
		return fmt.Errorf("invalid macro definition")
	}

	// Extract macro name
	nameNode := children[2]
	name := r.nodeToString(nameNode)

	// Extract parameters
	paramNode := children[3]
	paramList := r.nodeToString(paramNode)
	params, err := r.parseParameterList(paramList)
	if err != nil {
		return err
	}

	// Extract body
	bodyNode := children[4]
	body := r.reconstructVexSyntax(bodyNode)

	// Register the macro
	macro := &Macro{
		Name:   name,
		Params: params,
		Body:   body,
	}

	return r.RegisterMacro(name, macro)
}

// loadFallbackMacros loads essential built-in macros as a last resort
func (r *Registry) loadFallbackMacros() error {
	// Only provide absolutely essential macros that are needed for the core to work
	// Most macros should be defined in core/core.vx using (macro ...) syntax
	return fmt.Errorf("no core macros file found - please ensure core/core.vx exists with macro definitions")
}

// validateMacro validates a macro definition
func (r *Registry) validateMacro(name string, macro *Macro) error {
	if name == "" {
		return fmt.Errorf("macro name cannot be empty")
	}

	// Check for reserved words
	reservedWords := []string{"if", "def", "fn", "let", "do", "when", "unless"}
	for _, reserved := range reservedWords {
		if name == reserved {
			return fmt.Errorf("'%s' is a reserved word", name)
		}
	}

	// Check for conflicts
	if _, exists := r.macros[name]; exists {
		return fmt.Errorf("macro '%s' is already defined", name)
	}

	// Validate parameters
	paramSet := make(map[string]bool)
	for _, param := range macro.Params {
		if param == "" {
			return fmt.Errorf("empty parameter name in macro '%s'", name)
		}
		if paramSet[param] {
			return fmt.Errorf("duplicate parameter '%s' in macro '%s'", param, name)
		}
		paramSet[param] = true
	}

	// Validate body
	if strings.TrimSpace(macro.Body) == "" {
		return fmt.Errorf("macro '%s' has empty body", name)
	}

	return nil
}

// Helper methods
func (r *Registry) nodeToString(node antlr.Tree) string {
	if terminal, ok := node.(*antlr.TerminalNodeImpl); ok {
		return terminal.GetText()
	}
	return r.reconstructVexSyntax(node)
}

func (r *Registry) parseParameterList(paramList string) ([]string, error) {
	trimmed := strings.TrimSpace(paramList)
	
	// Handle both Vex format [x y] and Go format []interface{}{x, y}
	if strings.HasPrefix(trimmed, "[") && strings.HasSuffix(trimmed, "]") && !strings.Contains(trimmed, "interface{}") {
		// Vex format
		inner := strings.TrimSpace(trimmed[1 : len(trimmed)-1])
		if inner == "" {
			return []string{}, nil
		}
		return strings.Fields(inner), nil
	} else if strings.HasPrefix(trimmed, "[]interface{}") {
		// Go format
		if start := strings.Index(trimmed, "}{"); start != -1 {
			start += 2
			if end := strings.LastIndex(trimmed, "}"); end > start {
				inner := strings.TrimSpace(trimmed[start:end])
				if inner == "" {
					return []string{}, nil
				}
				params := strings.Split(inner, ",")
				for i, param := range params {
					params[i] = strings.TrimSpace(param)
				}
				return params, nil
			}
		}
	}
	
	return nil, fmt.Errorf("invalid parameter list format: %s", paramList)
}

func (r *Registry) reconstructVexSyntax(node antlr.Tree) string {
	switch n := node.(type) {
	case *parser.ListContext:
		var parts []string
		for _, child := range n.GetChildren() {
			if child != nil {
				parts = append(parts, r.reconstructVexSyntax(child))
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
				parts = append(parts, r.reconstructVexSyntax(child))
			}
		}
		return strings.Join(parts, " ")
	case *antlr.TerminalNodeImpl:
		return n.GetText()
	case antlr.TerminalNode:
		return n.GetText()
	default:
		return ""
	}
}