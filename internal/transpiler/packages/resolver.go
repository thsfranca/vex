package packages

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/antlr4-go/antlr/v4"
	"github.com/thsfranca/vex/internal/transpiler/analysis"
	"github.com/thsfranca/vex/internal/transpiler/parser"
)

// Resolver builds a combined Vex program from an entry file by discovering local packages,
// resolving dependencies, ordering them, and detecting circular dependencies.
// Resolver discovers local Vex packages, validates dependencies, and orders compilation.
type Resolver struct {
	moduleRoot string
	edgeLoc    map[string]map[string]string // fromPkg -> toPkg -> file path where import declared (first seen)
}

// NewResolver creates a new package resolver.
func NewResolver(moduleRoot string) *Resolver {
	return &Resolver{moduleRoot: moduleRoot, edgeLoc: make(map[string]map[string]string)}
}

// Result of resolution
// Result carries the combined program and metadata for code generation.
type Result struct {
	CombinedSource string
	IgnoreImports  map[string]bool                            // import paths that should not be emitted as Go imports (local Vex packages)
	Exports        map[string]map[string]bool                 // package path -> set of exported symbols
	PkgSchemes     map[string]map[string]*analysis.TypeScheme // package path -> symbol -> scheme
}

// BuildProgramFromEntry resolves dependencies starting from entryFile and returns combined source in topo order.
func (r *Resolver) BuildProgramFromEntry(entryFile string) (*Result, error) {
	// Determine module root via vex.pkg if present
	r.moduleRoot = r.detectModuleRoot(filepath.Dir(entryFile))

	// Graph: node = import path (relative to module root), edges to its local imports
	graph := make(map[string][]string)
	visited := make(map[string]bool)
	temp := make(map[string]bool)
	order := make([]string, 0)
	ignoreImports := make(map[string]bool)
	exports := make(map[string]map[string]bool)
	stack := make([]string, 0)

	// Discover initial local imports from entry file content
	entryContent, err := os.ReadFile(entryFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read entry file: %w", err)
	}

	entryImports := r.collectImports(string(entryContent))
	// Seed graph with entry pseudo-node mapping to local imports
	const entryNode = "@entry"
	graph[entryNode] = make([]string, 0)
	for _, imp := range entryImports {
		if r.isLocalPackage(imp) {
			graph[entryNode] = append(graph[entryNode], imp)
			ignoreImports[imp] = true
			// record edge location
			if _, ok := r.edgeLoc[entryNode]; !ok {
				r.edgeLoc[entryNode] = make(map[string]string)
			}
			r.edgeLoc[entryNode][imp] = entryFile
		}
	}

	// DFS over local packages to build graph
	var visit func(node string) error
	visit = func(node string) error {
		if visited[node] {
			return nil
		}
		if temp[node] {
			// cycle detected; build chain from stack
			cycle := buildCycle(stack, node)
			return fmt.Errorf("[PACKAGE-CYCLE]: %s", formatCycleError(cycle, r.edgeLoc))
		}
		temp[node] = true
		stack = append(stack, node)

		// Ensure node has import list
		if _, ok := graph[node]; !ok {
			// Build imports for this node if it's a package path
			if node != entryNode {
				imports, filesByDep, err := r.collectPackageImportsWithFiles(node)
				if err != nil {
					return err
				}
				for _, dep := range imports {
					if r.isLocalPackage(dep) {
						graph[node] = append(graph[node], dep)
						ignoreImports[dep] = true
						// record edge location
						if _, ok := r.edgeLoc[node]; !ok {
							r.edgeLoc[node] = make(map[string]string)
						}
						if p, ok := filesByDep[dep]; ok {
							r.edgeLoc[node][dep] = p
						}
					}
				}
				// Collect package exports
				pkgExports, err := r.collectPackageExports(node)
				if err == nil && len(pkgExports) > 0 {
					exports[node] = pkgExports
				}
			} else {
				graph[node] = []string{}
			}
		}

		// Visit dependencies
		for _, dep := range graph[node] {
			if err := visit(dep); err != nil {
				return err
			}
		}

		temp[node] = false
		visited[node] = true
		// pop stack
		if len(stack) > 0 {
			stack = stack[:len(stack)-1]
		}
		if node != entryNode {
			order = append(order, node)
		}
		return nil
	}

	if err := visit(entryNode); err != nil {
		// Expand cycle error message
		return nil, fmt.Errorf("package resolution failed: %w", err)
	}

	// Build combined source: concatenate all .vx files from ordered packages, then the entry file content
	var combined strings.Builder
	for _, pkgPath := range order {
		dir := filepath.Join(r.moduleRoot, filepath.FromSlash(pkgPath))
		files, _ := r.findVexFiles(dir)
		for _, f := range files {
			data, err := os.ReadFile(f)
			if err != nil {
				return nil, fmt.Errorf("failed to read %s: %w", f, err)
			}
			combined.WriteString(string(data))
			if !strings.HasSuffix(string(data), "\n") {
				combined.WriteString("\n")
			}
		}
	}
	// Finally append entry file content
	combined.WriteString(string(entryContent))
	if !strings.HasSuffix(string(entryContent), "\n") {
		combined.WriteString("\n")
	}

	// Compute type schemes for exported symbols in each discovered local package
	pkgSchemes := make(map[string]map[string]*analysis.TypeScheme)
	for _, pkgPath := range order {
		if ex, ok := exports[pkgPath]; ok && len(ex) > 0 {
			if sch, err := r.collectPackageSchemes(pkgPath, ex); err == nil && len(sch) > 0 {
				pkgSchemes[pkgPath] = sch
			} else if err != nil {
				// TODO: Remove debug - temporarily expose error
				return nil, fmt.Errorf("failed to collect schemes for package %s: %w", pkgPath, err)
			}
		}
	}

	return &Result{CombinedSource: combined.String(), IgnoreImports: ignoreImports, Exports: exports, PkgSchemes: pkgSchemes}, nil
}

// isLocalPackage determines if an import path refers to a local directory with .vx files under module root.
func (r *Resolver) isLocalPackage(importPath string) bool {
	dir := filepath.Join(r.moduleRoot, filepath.FromSlash(importPath))
	// Must exist and contain at least one .vx file
	files, err := r.findVexFiles(dir)
	return err == nil && len(files) > 0
}

func (r *Resolver) findVexFiles(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	var files []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if strings.HasSuffix(name, ".vx") || strings.HasSuffix(name, ".vex") {
			files = append(files, filepath.Join(dir, name))
		}
	}
	return files, nil
}

// collectPackageImports scans all .vx files in a local package directory and aggregates their local import paths.
func (r *Resolver) collectPackageImportsWithFiles(importPath string) ([]string, map[string]string, error) {
	dir := filepath.Join(r.moduleRoot, filepath.FromSlash(importPath))
	var imports []string
	filesByDep := make(map[string]string)
	var firstErr error

	walkFn := func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			if firstErr == nil {
				firstErr = err
			}
			return nil
		}
		if d.IsDir() {
			if path != dir {
				// Do not recurse into subpackages
				return filepath.SkipDir
			}
			return nil
		}
		if strings.HasSuffix(d.Name(), ".vx") || strings.HasSuffix(d.Name(), ".vex") {
			data, readErr := os.ReadFile(path)
			if readErr != nil {
				if firstErr == nil {
					firstErr = readErr
				}
				return nil
			}
			imps := r.collectImports(string(data))
			imports = append(imports, imps...)
			for _, dep := range imps {
				if _, ok := filesByDep[dep]; !ok {
					filesByDep[dep] = path
				}
			}
		}
		return nil
	}

	_ = filepath.WalkDir(dir, walkFn)
	if firstErr != nil {
		return nil, nil, firstErr
	}

	// Deduplicate
	uniq := make(map[string]bool)
	var result []string
	for _, imp := range imports {
		if !uniq[imp] {
			uniq[imp] = true
			result = append(result, imp)
		}
	}
	return result, filesByDep, nil
}

// collectImports parses Vex source and returns all import paths mentioned, ignoring aliases.
func (r *Resolver) collectImports(source string) []string {
	input := antlr.NewInputStream(source)
	lexer := parser.NewVexLexer(input)
	tokens := antlr.NewCommonTokenStream(lexer, 0)
	p := parser.NewVexParser(tokens)
	tree := p.Program()

	var imports []string

	// Walk program children to find (import ...)
	for _, child := range tree.GetChildren() {
		list, ok := child.(*parser.ListContext)
		if !ok || list.GetChildCount() < 3 {
			continue
		}
		// First symbol after '('
		if fn, ok := list.GetChild(1).(antlr.TerminalNode); ok && fn.GetText() == "import" {
			argNode := list.GetChild(2)
			// Single string
			if term, ok := argNode.(antlr.TerminalNode); ok {
				t := term.GetText()
				if strings.HasPrefix(t, "\"") && strings.HasSuffix(t, "\"") {
					imports = append(imports, strings.Trim(t, "\""))
				}
				continue
			}
			// Array form
			if arr, ok := argNode.(*parser.ArrayContext); ok {
				for i := 1; i < arr.GetChildCount()-1; i++ {
					el := arr.GetChild(i)
					if el == nil {
						continue
					}
					if term, ok := el.(antlr.TerminalNode); ok {
						tt := term.GetText()
						if strings.HasPrefix(tt, "\"") && strings.HasSuffix(tt, "\"") {
							imports = append(imports, strings.Trim(tt, "\""))
						}
						continue
					}
					if pairArr, ok := el.(*parser.ArrayContext); ok {
						if pth, _ := r.extractPathAlias(pairArr); pth != "" {
							imports = append(imports, pth)
						}
						continue
					}
					if pairList, ok := el.(*parser.ListContext); ok {
						if pth, _ := r.extractPathAlias(pairList); pth != "" {
							imports = append(imports, pth)
						}
						continue
					}
				}
			}
		}
	}

	return imports
}

// collectPackageSchemes analyzes a local package and returns type schemes for its exported symbols.
func (r *Resolver) collectPackageSchemes(importPath string, exported map[string]bool) (map[string]*analysis.TypeScheme, error) {
	dir := filepath.Join(r.moduleRoot, filepath.FromSlash(importPath))
	files, err := r.findVexFiles(dir)
	if err != nil {
		return nil, err
	}
	if len(files) == 0 {
		return map[string]*analysis.TypeScheme{}, nil
	}

	// Concatenate all files in package to a single source
	var b strings.Builder
	for _, f := range files {
		data, readErr := os.ReadFile(f)
		if readErr != nil {
			return nil, readErr
		}
		b.WriteString(string(data))
		if !strings.HasSuffix(string(data), "\n") {
			b.WriteString("\n")
		}
	}

	input := antlr.NewInputStream(b.String())
	lexer := parser.NewVexLexer(input)
	tokens := antlr.NewCommonTokenStream(lexer, 0)
	p := parser.NewVexParser(tokens)
	prog := p.Program()
	if prog == nil {
		return map[string]*analysis.TypeScheme{}, nil
	}

	// Macro expansion removed from type discovery to enforce explicit imports
	// Type discovery should work on raw AST without automatic macro loading

	// Run analyzer to compute schemes
	a := analysis.NewAnalyzer()
	ast := &resolverAnalysisAST{root: prog}
	if _, err := a.Analyze(ast); err != nil {
		// Return empty on analysis error; caller may ignore
		// TODO: Remove debug - temporarily expose error
		return map[string]*analysis.TypeScheme{}, fmt.Errorf("analysis failed: %w", err)
	}

	out := make(map[string]*analysis.TypeScheme)
	for sym := range exported {
		if sch, ok := a.GetTypeScheme(sym); ok {
			out[sym] = sch
		}
	}
	return out, nil
}

// resolverAnalysisAST adapts a parser.ProgramContext to the analysis.AST interface.
type resolverAnalysisAST struct{ root antlr.Tree }

func (a *resolverAnalysisAST) Accept(visitor analysis.ASTVisitor) error {
	if prog, ok := a.root.(*parser.ProgramContext); ok {
		return visitor.VisitProgram(prog)
	}
	return fmt.Errorf("invalid AST root type for analysis")
}

// collectPackageExports scans all .vx files in a package directory to find export declarations.
func (r *Resolver) collectPackageExports(importPath string) (map[string]bool, error) {
	dir := filepath.Join(r.moduleRoot, filepath.FromSlash(importPath))
	exports := make(map[string]bool)
	var firstErr error

	walkFn := func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			if firstErr == nil {
				firstErr = err
			}
			return nil
		}
		if d.IsDir() {
			if path != dir {
				return filepath.SkipDir
			}
			return nil
		}
		if strings.HasSuffix(d.Name(), ".vx") || strings.HasSuffix(d.Name(), ".vex") {
			data, readErr := os.ReadFile(path)
			if readErr != nil {
				if firstErr == nil {
					firstErr = readErr
				}
				return nil
			}
			for _, sym := range r.findExportsInSource(string(data)) {
				exports[sym] = true
			}
		}
		return nil
	}

	_ = filepath.WalkDir(dir, walkFn)
	if firstErr != nil {
		return nil, firstErr
	}
	return exports, nil
}

// findExportsInSource parses a Vex source to extract exported symbol names from (export [ ... ]) forms.
func (r *Resolver) findExportsInSource(source string) []string {
	input := antlr.NewInputStream(source)
	lexer := parser.NewVexLexer(input)
	tokens := antlr.NewCommonTokenStream(lexer, 0)
	p := parser.NewVexParser(tokens)
	tree := p.Program()

	var out []string
	for _, child := range tree.GetChildren() {
		list, ok := child.(*parser.ListContext)
		if !ok || list.GetChildCount() < 3 {
			continue
		}
		if fn, ok := list.GetChild(1).(antlr.TerminalNode); ok && fn.GetText() == "export" {
			argNode := list.GetChild(2)
			if arr, ok := argNode.(*parser.ArrayContext); ok {
				for i := 1; i < arr.GetChildCount()-1; i++ {
					if t, ok := arr.GetChild(i).(antlr.TerminalNode); ok {
						text := t.GetText()
						if text == "," {
							continue
						}
						// Symbols may appear without quotes as TERMINAL tokens
						cleaned := strings.Trim(text, "\"")
						if cleaned != "[" && cleaned != "]" && cleaned != "(" && cleaned != ")" {
							out = append(out, cleaned)
						}
					}
				}
			}
		}
	}
	return out
}

// buildCycle builds a slice representing the cycle path ending where 'node' was found on the stack
func buildCycle(stack []string, node string) []string {
	// find node in stack
	idx := -1
	for i := len(stack) - 1; i >= 0; i-- {
		if stack[i] == node {
			idx = i
			break
		}
	}
	if idx == -1 {
		return append([]string{}, node)
	}
	return append([]string{}, stack[idx:]...)
}

// formatCycleError produces a readable cycle error string with file hints when available
func formatCycleError(cycle []string, edgeLoc map[string]map[string]string) string {
	if len(cycle) == 0 {
		return "circular dependency detected"
	}
	// Build chain a -> b -> c -> a
	var parts []string
	for i := 0; i < len(cycle); i++ {
		curr := cycle[i]
		next := cycle[(i+1)%len(cycle)]
		if locs, ok := edgeLoc[curr]; ok {
			if p, ok := locs[next]; ok && p != "" {
				parts = append(parts, fmt.Sprintf("%s -(import at %s)-> %s", curr, p, next))
				continue
			}
		}
		parts = append(parts, fmt.Sprintf("%s -> %s", curr, next))
	}
	return "circular dependency detected: " + strings.Join(parts, " | ")
}

// detectModuleRoot walks up from startDir to find a vex.pkg file; returns dir containing it or original moduleRoot.
func (r *Resolver) detectModuleRoot(startDir string) string {
	dir := startDir
	for {
		candidate := filepath.Join(dir, "vex.pkg")
		if _, err := os.Stat(candidate); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir { // reached filesystem root
			break
		}
		dir = parent
	}
	if r.moduleRoot != "" {
		return r.moduleRoot
	}
	return startDir
}

func (r *Resolver) extractPathAlias(node antlr.Tree) (string, string) {
	var parts []string
	switch n := node.(type) {
	case *parser.ListContext:
		for i := 1; i < n.GetChildCount()-1; i++ {
			if t, ok := n.GetChild(i).(antlr.TerminalNode); ok {
				parts = append(parts, t.GetText())
			}
		}
	case *parser.ArrayContext:
		for i := 1; i < n.GetChildCount()-1; i++ {
			if t, ok := n.GetChild(i).(antlr.TerminalNode); ok {
				parts = append(parts, t.GetText())
			}
		}
	default:
		return "", ""
	}
	if len(parts) == 0 {
		return "", ""
	}
	path := parts[0]
	if strings.HasPrefix(path, "\"") && strings.HasSuffix(path, "\"") {
		path = strings.Trim(path, "\"")
	}
	alias := ""
	if len(parts) > 1 {
		alias = strings.Trim(parts[1], "\"")
	}
	return path, alias
}
