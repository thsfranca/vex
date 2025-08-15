package diagnostics

import (
	"bytes"
	"text/template"
)

// Catalog maps codes to short, canonical message templates.
// Templates receive the Diagnostic as dot (fields + Params map).
var Catalog = map[Code]*template.Template{
	// Syntax errors
	CodeSynEmpty: mustParse("empty expression"),

	// Type system errors
	CodeTypUndef:      mustParse("undefined symbol '{{index .Params \"Name\"}}'"),
	CodeTypIfMismatch: mustParse("branch types differ"),
	CodeTypCond:       mustParse("condition is not bool"),
	CodeTypArrayElem:  mustParse("array elements must share a type"),
	CodeTypMapKey:     mustParse("map keys have incompatible types"),
	CodeTypMapVal:     mustParse("map values have incompatible types"),
	CodeTypNum:        mustParse("{{index .Params \"Op\"}} expects number arguments{{if .Params.Got}}; got {{index .Params \"Got\"}}{{end}}"),
	CodeTypEq:         mustParse("= expects both arguments to have the same type"),
	CodeTypNot:        mustParse("not expects a single bool argument"),
	CodeTypBoolArgs:   mustParse("{{index .Params \"Op\"}} expects bool arguments"),
	CodeTypIndex:      mustParse("index must be number"),

	// Arity and argument errors
	CodeAriArgs: mustParse("{{index .Params \"Op\"}} expects {{index .Params \"Expected\"}} arguments; got {{index .Params \"Got\"}}"),

	// Macro system errors
	CodeMacArgs:     mustParse("macro requires name, parameters, and body"),
	CodeMacReserved: mustParse("'{{index .Params \"Name\"}}' is reserved and cannot be used as macro name"),
	CodeMacUndef:    mustParse("macro '{{index .Params \"Name\"}}' not found"),

	// Record type errors
	CodeRecField:     mustParse("{{index .Params \"Message\"}}"),
	CodeRecFieldType: mustParse("{{index .Params \"Message\"}}"),
	CodeRecArgs:      mustParse("record requires name and fields"),
	CodeRecName:      mustParse("invalid record name"),
	CodeRecFields:    mustParse("record expects a field vector [field: Type ...]"),
	CodeRecConstruct: mustParse("record construction requires field vector"),

	// Special form errors
	CodeIfArgs:     mustParse("if requires condition and then-branch"),
	CodeDefArgs:    mustParse("def requires name and value"),
	CodeFnArgs:     mustParse("fn requires parameter list and body"),
	CodeFnParams:   mustParse("all function parameters require explicit type annotations (param: type)"),
	CodeFnRetType:  mustParse("function requires explicit return type annotation (-> type)"),
	CodeExportArgs: mustParse("export requires a list of symbols"),

	// Package and import errors
	CodeImpSyntax:      mustParse("import requires package path"),
	CodePkgNotExported: mustParse("symbol '{{index .Params \"Symbol\"}}' is not exported from package '{{index .Params \"Package\"}}'"),
	CodePkgCycle:       mustParse("{{index .Params \"Chain\"}}"),

	// Collection errors
	CodeMapArgs: mustParse("map requires a [key value ...] vector"),

	// General errors
	CodeGenUndef: mustParse("undefined function: {{index .Params \"Name\"}}"),

	// Symbol table errors
	CodeSymDup: mustParse("symbol '{{index .Params \"Name\"}}' already defined in current scope"),

	// Type/argument mismatch
	CodeTypArg: mustParse("argument {{index .Params \"Position\"}} type mismatch"),

	// Record nominal type
	CodeRecNominal: mustParse("{{index .Params \"Message\"}}"),
}

func mustParse(t string) *template.Template {
	return template.Must(template.New("").Parse(t))
}

func renderFromCatalog(d Diagnostic) string {
	tmpl, ok := Catalog[d.Code]
	if !ok {
		return ""
	}
	var buf bytes.Buffer
	_ = tmpl.Execute(&buf, d)
	return buf.String()
}
