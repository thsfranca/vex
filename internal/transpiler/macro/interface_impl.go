package macro

import (
	"fmt"
	"strings"

	"github.com/antlr4-go/antlr/v4"
	"github.com/thsfranca/vex/internal/transpiler/parser"
)

// MacroExpanderImpl implements the MacroExpander interface
// MacroExpanderImpl is a concrete MacroExpander that walks and rewrites Vex AST.
type MacroExpanderImpl struct {
	*Expander
	registry *Registry
}

// NewMacroExpander creates a new macro expander that implements the interface
// NewMacroExpander constructs a MacroExpanderImpl backed by a Registry.
func NewMacroExpander(registry *Registry) *MacroExpanderImpl {
	return &MacroExpanderImpl{
		Expander: NewExpander(registry),
		registry: registry,
	}
}

// ExpandMacros expands all macros in an AST
func (me *MacroExpanderImpl) ExpandMacros(ast AST) (AST, error) {
	// Get the root node
	root := ast.Root()
	
	// Expand macros in the tree
	expandedRoot, err := me.expandMacrosInTree(root)
	if err != nil {
		return nil, err
	}
	
	// Return new AST with expanded tree
	return NewVexAST(expandedRoot), nil
}

// RegisterMacro registers a macro
func (me *MacroExpanderImpl) RegisterMacro(name string, macro *Macro) error {
	return me.registry.RegisterMacro(name, macro)
}

// HasMacro checks if a macro exists
func (me *MacroExpanderImpl) HasMacro(name string) bool {
	return me.registry.HasMacro(name)
}

// GetMacro retrieves a macro
func (me *MacroExpanderImpl) GetMacro(name string) (*Macro, bool) {
	return me.registry.GetMacro(name)
}

// LoadStdlibModule loads a specific stdlib module
func (me *MacroExpanderImpl) LoadStdlibModule(moduleName string) error {
	return me.registry.LoadStdlibModule(moduleName)
}

// expandMacrosInTree recursively expands macros in an AST tree
func (me *MacroExpanderImpl) expandMacrosInTree(node antlr.Tree) (antlr.Tree, error) {
	switch n := node.(type) {
	case *parser.ProgramContext:
		return me.expandMacrosInProgram(n)
	case *parser.ListContext:
		return me.expandMacrosInList(n)
	case *parser.ArrayContext:
		return me.expandMacrosInArray(n)
	default:
		return node, nil // Terminal nodes don't need expansion
	}
}

// expandMacrosInProgram expands macros in a program context
func (me *MacroExpanderImpl) expandMacrosInProgram(ctx *parser.ProgramContext) (antlr.Tree, error) {
	// Collect all expanded children
	var expandedChildren []antlr.Tree
	hasChanges := false
	
	for _, child := range ctx.GetChildren() {
		if listCtx, ok := child.(*parser.ListContext); ok {
			expanded, err := me.expandMacrosInList(listCtx)
			if err != nil {
				return nil, err
			}
			expandedChildren = append(expandedChildren, expanded)
			// Check if the expansion actually changed the node
			if expanded != listCtx {
				hasChanges = true
			}
		} else {
			expandedChildren = append(expandedChildren, child)
		}
	}
	
	// If no changes were made, return the original context
	if !hasChanges {
		return ctx, nil
	}
	
	// Create a new program by re-parsing the expanded content
	return me.reconstructProgramFromChildren(expandedChildren)
}

// expandMacrosInList expands macros in a list context
func (me *MacroExpanderImpl) expandMacrosInList(ctx *parser.ListContext) (antlr.Tree, error) {
	childCount := ctx.GetChildCount()
	if childCount < 3 { // Need at least: '(', function, ')'
		return ctx, nil
	}

	// Get function name (first child after '(')
	funcNameNode := ctx.GetChild(1)
	funcName := me.nodeToString(funcNameNode)

	// Special case: macro definitions should not be expanded
	if funcName == "macro" {
		// This is a macro definition, not a call - don't expand it
		// Just register the macro and return the original context
		return me.processMacroDefinition(ctx)
	}

	// Check if this is a macro call
	if me.registry.HasMacro(funcName) {
		// Extract arguments
		args := make([]string, 0)
		for i := 2; i < childCount-1; i++ { // Skip '(' and ')'
			child := ctx.GetChild(i)
			if child != nil {
				// Recursively expand macros in arguments first
				expandedChild, err := me.expandMacrosInTree(child)
				if err != nil {
					return nil, err
				}
				args = append(args, me.reconstructVexSyntax(expandedChild))
			}
		}

		// Expand the macro
		expanded, err := me.ExpandMacro(funcName, args)
		if err != nil {
			return nil, err
		}

		// Parse the expanded result back into an AST node
		return me.parseExpandedCode(expanded)
	}

	// Not a macro call, but still need to recursively process children for nested macros
	hasChanges := false
	newChildren := make([]antlr.Tree, childCount)
	
	// Copy all children, recursively expanding any that need it
	for i := 0; i < childCount; i++ {
		child := ctx.GetChild(i)
		if child != nil {
			expandedChild, err := me.expandMacrosInTree(child)
			if err != nil {
				return nil, err
			}
			newChildren[i] = expandedChild
			// Check if this child was actually changed
			if expandedChild != child {
				hasChanges = true
			}
		} else {
			newChildren[i] = child
		}
	}
	
	// If no changes were made, return original context
	if !hasChanges {
		return ctx, nil
	}
	
	// Create a new list context with expanded children
	return me.reconstructListFromChildren(newChildren)
}

// reconstructListFromChildren creates a new list context from expanded children
func (me *MacroExpanderImpl) reconstructListFromChildren(children []antlr.Tree) (antlr.Tree, error) {
	// For now, we'll reconstruct by creating Vex syntax and re-parsing
	// This is not the most efficient but ensures correctness
	
	var parts []string
	parts = append(parts, "(")
	
	for i := 1; i < len(children)-1; i++ { // Skip opening and closing parentheses
		if children[i] != nil {
			parts = append(parts, me.reconstructVexSyntax(children[i]))
		}
	}
	
	parts = append(parts, ")")
	vexCode := strings.Join(parts, " ")
	
	// Parse the reconstructed code back into an AST
	return me.parseExpandedCode(vexCode)
}

// expandMacrosInArray expands macros in an array context
func (me *MacroExpanderImpl) expandMacrosInArray(ctx *parser.ArrayContext) (antlr.Tree, error) {
	// Arrays typically don't contain macro calls, but we should check elements
	for _, child := range ctx.GetChildren() {
		if listCtx, ok := child.(*parser.ListContext); ok {
			_, err := me.expandMacrosInList(listCtx)
			if err != nil {
				return nil, err
			}
		}
	}
	
	return ctx, nil
}

// parseExpandedCode parses expanded macro code back into AST
func (me *MacroExpanderImpl) parseExpandedCode(code string) (antlr.Tree, error) {
	inputStream := antlr.NewInputStream(code)
	lexer := parser.NewVexLexer(inputStream)
	tokenStream := antlr.NewCommonTokenStream(lexer, 0)
	vexParser := parser.NewVexParser(tokenStream)

	// Try to parse as a list first
	if listTree := vexParser.List(); listTree != nil {
		return listTree.(*parser.ListContext), nil
	}

	// Reset and try as a program
	inputStream = antlr.NewInputStream(code)
	lexer = parser.NewVexLexer(inputStream)
	tokenStream = antlr.NewCommonTokenStream(lexer, 0)
	vexParser = parser.NewVexParser(tokenStream)

	if programTree := vexParser.Program(); programTree != nil {
		return programTree.(*parser.ProgramContext), nil
	}

	return nil, fmt.Errorf("failed to parse expanded macro code: %s", code)
}

// Helper methods
func (me *MacroExpanderImpl) nodeToString(node antlr.Tree) string {
	// Always use reconstructVexSyntax for consistent spacing
	return me.Expander.reconstructVexSyntax(node)
}

// Interface definitions (to avoid import cycles)
type AST interface {
	Root() antlr.Tree
	Accept(visitor ASTVisitor) error
}

type ASTVisitor interface {
	VisitProgram(ctx *parser.ProgramContext) error
	VisitList(ctx *parser.ListContext) (Value, error)
	VisitArray(ctx *parser.ArrayContext) (Value, error)
	VisitTerminal(node antlr.TerminalNode) (Value, error)
}

type Value interface {
	String() string
	Type() string
}

// VexAST implementation for macro expansion
type VexAST struct {
	root antlr.Tree
}

func NewVexAST(root antlr.Tree) *VexAST {
	return &VexAST{root: root}
}

func (ast *VexAST) Root() antlr.Tree {
	return ast.root
}

func (ast *VexAST) Accept(visitor ASTVisitor) error {
	if programCtx, ok := ast.root.(*parser.ProgramContext); ok {
		return visitor.VisitProgram(programCtx)
	}
	return fmt.Errorf("invalid AST root type")
}

// processMacroDefinition handles macro definitions during expansion
func (me *MacroExpanderImpl) processMacroDefinition(ctx *parser.ListContext) (antlr.Tree, error) {
	childCount := ctx.GetChildCount()
	if childCount < 5 { // Need at least: '(', 'macro', name, params, body, ')'
		return ctx, nil // Let analyzer handle the error
	}

	// Extract macro name
	nameNode := ctx.GetChild(2)
	macroName := me.nodeToString(nameNode)

	// Extract parameters
	paramsNode := ctx.GetChild(3)
	paramsStr := me.nodeToString(paramsNode)

	// Extract body
	bodyNode := ctx.GetChild(4)
	bodyStr := me.nodeToString(bodyNode)

	// Register the macro
	macro := &Macro{
		Name:   macroName,
		Params: me.parseParameterList(paramsStr),
		Body:   bodyStr,
	}
	me.registry.RegisterMacro(macroName, macro)

	// Return the original context unchanged - macro definitions don't get expanded
	return ctx, nil
}

// parseParameterList parses parameter list from string
func (me *MacroExpanderImpl) parseParameterList(paramStr string) []string {
	trimmed := strings.TrimSpace(paramStr)
	if strings.HasPrefix(trimmed, "[") && strings.HasSuffix(trimmed, "]") {
		inner := strings.TrimSpace(trimmed[1 : len(trimmed)-1])
		if inner == "" {
			return []string{}
		}
		return strings.Fields(inner)
	}
	return []string{}
}

// reconstructProgramFromChildren creates a new program AST from expanded children
func (me *MacroExpanderImpl) reconstructProgramFromChildren(children []antlr.Tree) (antlr.Tree, error) {
	// Build the complete source code from all children, excluding EOF tokens
	var parts []string
	for _, child := range children {
		if child != nil {
			// Skip EOF tokens and error nodes
			if terminal, ok := child.(antlr.TerminalNode); ok {
				if terminal.GetText() == "<EOF>" {
					continue
				}
			}
			if _, ok := child.(*antlr.ErrorNodeImpl); ok {
				continue
			}
			
			childText := me.nodeToString(child)
			if strings.TrimSpace(childText) != "" && childText != "<EOF>" {
				parts = append(parts, childText)
			}
		}
	}
	
	// Join with spaces for proper program structure
	programText := strings.Join(parts, " ")
	
	// Parse the reconstructed program
	inputStream := antlr.NewInputStream(programText)
	lexer := parser.NewVexLexer(inputStream)
	tokenStream := antlr.NewCommonTokenStream(lexer, 0)
	vexParser := parser.NewVexParser(tokenStream)

	// Parse as a complete program
	if programTree := vexParser.Program(); programTree != nil {
		return programTree.(*parser.ProgramContext), nil
	}
	
	return nil, fmt.Errorf("failed to reconstruct program AST from: %s", programText)
}