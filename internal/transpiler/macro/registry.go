package macro

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
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
// Registry stores macros and loads core macros from disk.
type Registry struct {
	macros         map[string]*Macro
	coreLoaded     bool
	loadedModules  map[string]bool // Track which stdlib modules have been loaded
	config         Config
}

// Config holds macro system configuration
// Config controls core macro loading and validation behavior.
type Config struct {
	CoreMacroPath    string
	StdlibPath       string // Absolute path to stdlib directory (for tests)
	EnableValidation bool
}

// NewRegistry creates a new macro registry
// NewRegistry constructs a macro registry with the given configuration.
func NewRegistry(config Config) *Registry {
	return &Registry{
		macros:        make(map[string]*Macro),
		loadedModules: make(map[string]bool),
		config:        config,
	}
}

// RegisterMacro registers a new macro
// RegisterMacro adds a macro to the registry with optional validation.
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

// LoadStdlibModule loads a specific stdlib module (e.g., "vex.core", "vex.collections")
func (r *Registry) LoadStdlibModule(moduleName string) error {
	// Check if already loaded
	if r.loadedModules[moduleName] {
		return nil
	}
	
	// Convert module name to path (e.g., "vex.core" -> "vex/core")
	modulePath := strings.ReplaceAll(moduleName, ".", "/")
	
	// Try different stdlib locations
	var stdlibDirs []string
	
	// If explicit stdlib path is provided, use it first
	if r.config.StdlibPath != "" {
		stdlibDirs = append(stdlibDirs, filepath.Join(r.config.StdlibPath, modulePath))
	}
	
	// Add relative paths as fallback
	stdlibDirs = append(stdlibDirs,
		filepath.Join("stdlib", modulePath),
		filepath.Join("../../stdlib", modulePath),
		filepath.Join("../../../stdlib", modulePath),
	)
	
	// Try to load from one of the directories
	for _, dir := range stdlibDirs {
		if err := r.loadFromDirectory(dir); err == nil {
			r.loadedModules[moduleName] = true
			return nil
		}
	}
	
	return fmt.Errorf("stdlib module '%s' not found in any of the search paths", moduleName)
}

// LoadCoreMacros loads core macros from file (legacy auto-loading)
func (r *Registry) LoadCoreMacros() error {
	if r.coreLoaded {
		return nil
	}
	r.coreLoaded = true

    // 1) Explicit path (file or directory)
    if r.config.CoreMacroPath != "" {
        if err := r.loadFromPath(r.config.CoreMacroPath); err == nil {
            return nil
        }
    }

    // 2) Stdlib package directories (load all that exist)
    var stdlibDirs []string
    
    // If explicit stdlib path is provided, use it first
    if r.config.StdlibPath != "" {
        stdlibDirs = append(stdlibDirs,
            filepath.Join(r.config.StdlibPath, "vex/core"),
            filepath.Join(r.config.StdlibPath, "vex/conditions"),
            filepath.Join(r.config.StdlibPath, "vex/collections"),
            filepath.Join(r.config.StdlibPath, "vex/bindings"),
            filepath.Join(r.config.StdlibPath, "vex/flow"),
            filepath.Join(r.config.StdlibPath, "vex/threading"),
            filepath.Join(r.config.StdlibPath, "vex/test"),
        )
    }
    
    // Add relative paths as fallback
    stdlibDirs = append(stdlibDirs,
        "stdlib/vex/core",
        "stdlib/vex/conditions",
        "stdlib/vex/collections",
        "stdlib/vex/bindings",
        "stdlib/vex/flow",
        "stdlib/vex/threading",
        "stdlib/vex/test",
        "../../stdlib/vex/core",
        "../../stdlib/vex/conditions",
        "../../stdlib/vex/collections",
        "../../stdlib/vex/bindings",
        "../../stdlib/vex/flow",
        "../../stdlib/vex/threading",
        "../../stdlib/vex/test",
        "../../../stdlib/vex/core",
        "../../../stdlib/vex/conditions",
        "../../../stdlib/vex/collections",
        "../../../stdlib/vex/bindings",
        "../../../stdlib/vex/flow",
        "../../../stdlib/vex/threading",
        "../../../stdlib/vex/test",
    )
    loadedAny := false
    for _, dir := range stdlibDirs {
        if err := r.loadFromDirectory(dir); err == nil {
            loadedAny = true
        }
    }
    if loadedAny {
        return nil
    }

    return fmt.Errorf("no core macros available - tried stdlib paths")
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

func (r *Registry) loadFromPath(path string) error {
    info, err := os.Stat(path)
    if err != nil {
        return err
    }
    if info.IsDir() {
        return r.loadFromDirectory(path)
    }
    return r.loadFromFile(path)
}

func (r *Registry) loadFromFile(path string) error {
    if !strings.HasSuffix(path, ".vx") && !strings.HasSuffix(path, ".vex") {
        return fmt.Errorf("not a vex file: %s", path)
    }
    content, err := os.ReadFile(path)
    if err != nil {
        return err
    }
    return r.loadMacrosFromContent(string(content))
}

func (r *Registry) loadFromDirectory(dir string) error {
    entries, err := os.ReadDir(dir)
    if err != nil {
        return err
    }
    had := false
    for _, e := range entries {
        if e.IsDir() {
            continue
        }
        name := e.Name()
        if strings.HasSuffix(name, ".vx") || strings.HasSuffix(name, ".vex") {
            had = true
            full := filepath.Join(dir, name)
            if err := r.loadFromFile(full); err != nil {
                return err
            }
        }
    }
    if !had {
        return fs.ErrNotExist
    }
    return nil
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



// validateMacro validates a macro definition
func (r *Registry) validateMacro(name string, macro *Macro) error {
	if name == "" {
		return fmt.Errorf("macro name cannot be empty")
	}

	// Check for reserved words
	reservedWords := []string{"if", "def", "fn", "let", "do", "when", "unless"}
	for _, reserved := range reservedWords {
		if name == reserved {
			return fmt.Errorf("[MACRO-RESERVED]: '%s' is a reserved word", name)
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