package diagnostics

// Code is a stable identifier for a diagnostic.
// Follows the convention documented in docs/error-messages.md (e.g., TYPE-UNDEFINED).
type Code string

const (
	// Syntax / forms
	CodeSynEmpty    Code = "SYNTAX-EMPTY"
	CodeSyntaxEmpty Code = "SYNTAX-EMPTY" // Alias for compatibility

	// Typing
	CodeTypUndef       Code = "TYPE-UNDEFINED"
	CodeTypeUndefined  Code = "TYPE-UNDEFINED" // Alias for analyzer compatibility
	CodeTypCond        Code = "TYPE-CONDITION"
	CodeTypeCondition  Code = "TYPE-CONDITION" // Alias for analyzer compatibility
	CodeTypIfMismatch  Code = "TYPE-IF-MISMATCH"
	CodeTypeIfMismatch Code = "TYPE-IF-MISMATCH" // Alias for analyzer compatibility
	CodeTypArrayElem   Code = "TYPE-ARRAY-ELEMENT"
	CodeTypMapKey      Code = "TYPE-MAP-KEY"
	CodeTypeMapKey     Code = "TYPE-MAP-KEY" // Alias for analyzer compatibility
	CodeTypMapVal      Code = "TYPE-MAP-VALUE"
	CodeTypeMapValue   Code = "TYPE-MAP-VALUE" // Alias for analyzer compatibility
	CodeTypNum         Code = "TYPE-NUMBER"
	CodeTypEq          Code = "TYPE-EQUALITY"
	CodeTypNot         Code = "TYPE-NOT"
	CodeTypBoolArgs    Code = "TYPE-BOOLEAN-ARGS"

	// Arity/shape
	CodeAriArgs Code = "ARITY-ARGUMENTS"

	// Macros
	CodeMacArgs     Code = "MACRO-ARGUMENTS"
	CodeMacReserved Code = "MACRO-RESERVED"
	CodeMacUndef    Code = "MACRO-UNDEFINED"

	// Records
	CodeRecField        Code = "RECORD-FIELD"
	CodeRecFieldType    Code = "RECORD-FIELD-TYPE"
	CodeRecNominal      Code = "RECORD-NOMINAL"
	CodeRecordNominal   Code = "RECORD-NOMINAL" // Alias for analyzer compatibility
	CodeRecArgs         Code = "RECORD-ARGUMENTS"
	CodeRecName         Code = "RECORD-NAME"
	CodeRecordName      Code = "RECORD-NAME" // Alias for analyzer compatibility
	CodeRecFields       Code = "RECORD-FIELDS"
	CodeRecordFields    Code = "RECORD-FIELDS" // Alias for analyzer compatibility
	CodeRecConstruct    Code = "RECORD-CONSTRUCT"
	CodeRecordConstruct Code = "RECORD-CONSTRUCT" // Alias for analyzer compatibility

	// Special forms
	CodeIfArgs             Code = "IF-ARGUMENTS"
	CodeIfArguments        Code = "IF-ARGUMENTS" // Alias for analyzer compatibility
	CodeDefArgs            Code = "DEF-ARGUMENTS"
	CodeDefArguments       Code = "DEF-ARGUMENTS" // Alias for analyzer compatibility
	CodeFnArgs             Code = "FUNCTION-ARGUMENTS"
	CodeFunctionArguments  Code = "FUNCTION-ARGUMENTS" // Alias for analyzer compatibility
	CodeFnParams           Code = "FUNCTION-PARAMETERS"
	CodeFunctionParameters Code = "FUNCTION-PARAMETERS" // Alias for analyzer compatibility
	CodeFnRetType          Code = "FUNCTION-RETURN-TYPE"
	CodeFunctionReturnType Code = "FUNCTION-RETURN-TYPE" // Alias for analyzer compatibility
	CodeExportArgs         Code = "EXPORT-ARGUMENTS"

	// Imports / packages
	CodeImpSyntax      Code = "IMPORT-SYNTAX"
	CodePkgNotExported Code = "PACKAGE-NOT-EXPORTED"
	CodePkgCycle       Code = "PACKAGE-CYCLE"

	// Collections / maps
	CodeMapArgs      Code = "MAP-ARGUMENTS"
	CodeMapArguments Code = "MAP-ARGUMENTS" // Alias for analyzer compatibility
	CodeTypIndex     Code = "TYPE-INDEX"

	// General
	CodeGenUndef Code = "GENERAL-UNDEFINED"

	// Symbol table
	CodeSymDup    Code = "SYMBOL-DUPLICATE"
	CodeSymNaming Code = "SYMBOL-NAMING"

	// Type/argument mismatch
	CodeTypArg Code = "TYPE-ARGUMENT"

	// Additional codes still in use
	CodeExportArguments  Code = "EXPORT-ARGUMENTS"
	CodeTypeIndex        Code = "TYPE-INDEX"
	CodeTypeArrayElement Code = "TYPE-ARRAY-ELEMENT"
)
