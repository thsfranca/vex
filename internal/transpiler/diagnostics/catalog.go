package diagnostics

import (
	"bytes"
	"text/template"
)

// Catalog maps codes to short, canonical message templates.
// Templates receive the Diagnostic as dot (fields + Params map).
var Catalog = map[Code]*template.Template{
    // Syntax / forms
    CodeSynEmpty:       mustParse("empty expression"),

    // Typing
    CodeTypUndef:       mustParse("undefined symbol '{{index .Params \"Name\"}}'") ,
    CodeTypIfMismatch:  mustParse("branch types differ"),
    CodeTypCond:        mustParse("condition is not bool"),
    CodeTypArrayElem:   mustParse("array elements must share a type"),
    CodeTypMapKey:      mustParse("map keys have incompatible types"),
    CodeTypMapVal:      mustParse("map values have incompatible types"),
    CodeTypNum:         mustParse("{{index .Params \"Op\"}} expects number arguments{{if .Params.Got}}; got {{index .Params \"Got\"}}{{end}}"),
    CodeTypEq:          mustParse("= expects both arguments to have the same type"),
    CodeTypNot:         mustParse("not expects a single bool argument"),
    CodeTypBoolArgs:    mustParse("{{index .Params \"Op\"}} expects bool arguments"),

    // Arity/shape
    CodeAriArgs:        mustParse("{{index .Params \"Op\"}} expects {{index .Params \"Expected\"}} arguments; got {{index .Params \"Got\"}}"),

    // Macros
    CodeMacArgs:        mustParse("macro requires name, parameters, and body"),
    CodeMacReserved:    mustParse("'{{index .Params \"Name\"}}' is reserved and cannot be used as macro name"),

    // Records
    CodeRecField:       mustParse("{{index .Params \"Message\"}}"),
    CodeRecFieldType:   mustParse("{{index .Params \"Message\"}}"),

    // Special forms
    CodeIfArgs:         mustParse("if requires condition and then-branch"),
    CodeDefArgs:        mustParse("def requires name and value"),
    CodeFnArgs:         mustParse("fn requires parameter list and body"),

    // Imports / packages
    CodeImpSyntax:      mustParse("import requires package path"),
    CodePkgNotExported: mustParse("symbol '{{index .Params \"Symbol\"}}' is not exported from package '{{index .Params \"Package\"}}'") ,
    CodePkgCycle:       mustParse("{{index .Params \"Chain\"}}"),

    // Collections / maps
    CodeMapArgs:        mustParse("map requires a [key value ...] vector"),
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


