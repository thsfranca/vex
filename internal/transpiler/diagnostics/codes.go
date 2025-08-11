package diagnostics

// Code is a stable identifier for a diagnostic.
// Follows the convention documented in docs/error-messages.md (e.g., TYPE-UNDEFINED).
type Code string

const (
    // Syntax / forms
    CodeSynEmpty      Code = "SYNTAX-EMPTY"
    CodeSyntaxEmpty   Code = "SYNTAX-EMPTY" // Alias for compatibility

    // Typing
    CodeTypUndef       Code = "TYPE-UNDEFINED"
    CodeTypCond        Code = "TYPE-CONDITION"
    CodeTypIfMismatch  Code = "TYPE-IF-MISMATCH"
    CodeTypArrayElem   Code = "TYPE-ARRAY-ELEMENT"
    CodeTypMapKey      Code = "TYPE-MAP-KEY"
    CodeTypMapVal      Code = "TYPE-MAP-VALUE"
    CodeTypNum         Code = "TYPE-NUMBER"
    CodeTypEq          Code = "TYPE-EQUALITY"
    CodeTypNot         Code = "TYPE-NOT"
    CodeTypBoolArgs    Code = "TYPE-BOOLEAN-ARGS"

    // Arity/shape
    CodeAriArgs        Code = "ARITY-ARGUMENTS"

    // Macros
    CodeMacArgs        Code = "MACRO-ARGUMENTS"
    CodeMacReserved    Code = "MACRO-RESERVED"
    CodeMacUndef       Code = "MACRO-UNDEFINED"

    // Records
    CodeRecField       Code = "RECORD-FIELD"
    CodeRecFieldType   Code = "RECORD-FIELD-TYPE"
    CodeRecNominal     Code = "RECORD-NOMINAL"
    CodeRecArgs        Code = "RECORD-ARGUMENTS"
    CodeRecName        Code = "RECORD-NAME"
    CodeRecFields      Code = "RECORD-FIELDS"
    CodeRecConstruct   Code = "RECORD-CONSTRUCT"

    // Special forms
    CodeIfArgs         Code = "IF-ARGUMENTS"
    CodeDefArgs        Code = "DEF-ARGUMENTS"
    CodeFnArgs         Code = "FUNCTION-ARGUMENTS"
    CodeFnParams       Code = "FUNCTION-PARAMETERS"
    CodeFnRetType      Code = "FUNCTION-RETURN-TYPE"
    CodeExportArgs     Code = "EXPORT-ARGUMENTS"
    
    // Compatibility aliases for longer names
    CodeExportArguments     Code = "EXPORT-ARGUMENTS"
    CodeFunctionNaming      Code = "FUNCTION-NAMING"
    CodeFunctionParameters  Code = "FUNCTION-PARAMETERS"
    CodeFunctionReturnType  Code = "FUNCTION-RETURN-TYPE"

    // Imports / packages
    CodeImpSyntax      Code = "IMPORT-SYNTAX"
    CodePkgNotExported Code = "PACKAGE-NOT-EXPORTED"
    CodePkgCycle       Code = "PACKAGE-CYCLE"

    // Collections / maps
    CodeMapArgs        Code = "MAP-ARGUMENTS"
    CodeTypIndex       Code = "TYPE-INDEX"
    
    // General
    CodeGenUndef       Code = "GENERAL-UNDEFINED"
    
    // Symbol table
    CodeSymDup         Code = "SYMBOL-DUPLICATE"
    
    // Type/argument mismatch
    CodeTypArg         Code = "TYPE-ARGUMENT"
    
    // Additional compatibility aliases
    CodeTypeIndex         Code = "TYPE-INDEX"
    CodeRecordArguments   Code = "RECORD-ARGUMENTS"
    CodeRecordName        Code = "RECORD-NAME"
    CodeRecordFields      Code = "RECORD-FIELDS"
    CodeTypeArrayElement  Code = "TYPE-ARRAY-ELEMENT"
    CodeTypeUndefined     Code = "TYPE-UNDEFINED"
    CodeDefArguments      Code = "DEF-ARGUMENTS"
    CodeIfArguments       Code = "IF-ARGUMENTS"
    CodeTypeCondition     Code = "TYPE-CONDITION"
    CodeRecordNominal     Code = "RECORD-NOMINAL"
    CodeTypeIfMismatch    Code = "TYPE-IF-MISMATCH"
    CodeRecordConstruct   Code = "RECORD-CONSTRUCT"
    CodeFunctionArguments Code = "FUNCTION-ARGUMENTS"
    CodeMacroArguments    Code = "MACRO-ARGUMENTS"
    CodeMacroReserved     Code = "MACRO-RESERVED"
    CodeTypeEquality      Code = "TYPE-EQUALITY"
    CodePackageNotExported Code = "PACKAGE-NOT-EXPORTED"
    CodeArityArguments    Code = "ARITY-ARGUMENTS"
    CodeMapArguments      Code = "MAP-ARGUMENTS"
    CodeTypeMapKey        Code = "TYPE-MAP-KEY"
    CodeTypeMapValue      Code = "TYPE-MAP-VALUE"
)

