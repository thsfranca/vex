package codegen

import (
	"strings"
	"testing"

	"github.com/antlr4-go/antlr/v4"
	"github.com/thsfranca/vex/internal/transpiler/parser"
)

func parseProgramHelper(t *testing.T, src string) *parser.ProgramContext {
	t.Helper()
	input := antlr.NewInputStream(src)
	lex := parser.NewVexLexer(input)
	tokens := antlr.NewCommonTokenStream(lex, 0)
	p := parser.NewVexParser(tokens)
	prog := p.Program()
	if prog == nil {
		t.Fatalf("failed to parse program")
	}
	return prog.(*parser.ProgramContext)
}

func TestImportArrayWithAliasesAndCalls(t *testing.T) {
	src := "(import [\"fmt\" [\"net/http\" http] [\"encoding/json\" json]])\n" +
		"(http/Get \"http://example.com\")\n" +
		"(json/Marshal \"x\")\n" +
		"(fmt/Println \"ok\")\n"
	prog := parseProgramHelper(t, src)
	g := NewGoCodeGenerator(Config{PackageName: "main"})
	if err := g.VisitProgram(prog); err != nil {
		t.Fatalf("VisitProgram error: %v", err)
	}
	out := g.buildFinalCode()

	// Imports with aliases
	if !strings.Contains(out, "import http \"net/http\"") {
		t.Fatalf("expected alias import for net/http, got:\n%s", out)
	}
	if !strings.Contains(out, "import json \"encoding/json\"") {
		t.Fatalf("expected alias import for encoding/json, got:\n%s", out)
	}
	if !strings.Contains(out, "import \"fmt\"") {
		t.Fatalf("expected import for fmt, got:\n%s", out)
	}

	// Calls use aliases or package names
	if !strings.Contains(out, "http.Get(\"http://example.com\")") {
		t.Fatalf("expected http.Get call, got:\n%s", out)
	}
	if !strings.Contains(out, "json.Marshal(\"x\")") {
		t.Fatalf("expected json.Marshal call, got:\n%s", out)
	}
	if !strings.Contains(out, "fmt.Println(\"ok\")") {
		t.Fatalf("expected fmt.Println call, got:\n%s", out)
	}
}

func TestLocalPackageCall_DirectIdentifier(t *testing.T) {
	// Configure generator to treat package "a" as local and export symbol "add"
	g := NewGoCodeGenerator(Config{
		PackageName:   "main",
		IgnoreImports: map[string]bool{"a": true},
		Exports:       map[string]map[string]bool{"a": {"add": true}},
	})

	// Program calls a/add without any import statement
	prog := parseProgramHelper(t, "(a/add 1 2)")
	if err := g.VisitProgram(prog); err != nil {
		t.Fatalf("VisitProgram error: %v", err)
	}
	out := g.buildFinalCode()

	if strings.Contains(out, "import \"a\"") {
		t.Fatalf("should not emit Go import for local package 'a':\n%s", out)
	}
	if !strings.Contains(out, "add(1, 2)") {
		t.Fatalf("expected direct call to add(1, 2), got:\n%s", out)
	}
}
