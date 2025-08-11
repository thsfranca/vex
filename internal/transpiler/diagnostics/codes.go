package diagnostics

// Code is a stable identifier for a diagnostic.
// Follows the convention documented in docs/error-messages.md (e.g., VEX-TYP-UNDEF).
type Code string

const (
    // Syntax / forms
    CodeSynEmpty      Code = "VEX-SYN-EMPTY"

    // Typing
    CodeTypUndef       Code = "VEX-TYP-UNDEF"
    CodeTypCond        Code = "VEX-TYP-COND"
    CodeTypIfMismatch  Code = "VEX-TYP-IF-MISMATCH"
    CodeTypArrayElem   Code = "VEX-TYP-ARRAY-ELEM"
    CodeTypMapKey      Code = "VEX-TYP-MAP-KEY"
    CodeTypMapVal      Code = "VEX-TYP-MAP-VAL"
    CodeTypNum         Code = "VEX-TYP-NUM"
    CodeTypEq          Code = "VEX-TYP-EQ"
    CodeTypNot         Code = "VEX-TYP-NOT"
    CodeTypBoolArgs    Code = "VEX-TYP-BOOL-ARGS"

    // Arity/shape
    CodeAriArgs        Code = "VEX-ARI-ARGS"

    // Macros
    CodeMacArgs        Code = "VEX-MAC-ARGS"
    CodeMacReserved    Code = "VEX-MAC-RESERVED"

    // Records
    CodeRecField       Code = "VEX-REC-FIELD"
    CodeRecFieldType   Code = "VEX-REC-FIELD-TYPE"
    CodeRecNominal     Code = "VEX-TYP-REC-NOMINAL"

    // Special forms
    CodeIfArgs         Code = "VEX-IF-ARGS"
    CodeDefArgs        Code = "VEX-DEF-ARGS"
    CodeFnArgs         Code = "VEX-FN-ARGS"

    // Imports / packages
    CodeImpSyntax      Code = "VEX-IMP-SYNTAX"
    CodePkgNotExported Code = "VEX-PKG-NOT-EXPORTED"
    CodePkgCycle       Code = "VEX-PKG-CYCLE"

    // Collections / maps
    CodeMapArgs        Code = "VEX-MAP-ARGS"
)


