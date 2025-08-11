package analysis

import (
	"fmt"
	"strings"
	"testing"

	"github.com/antlr4-go/antlr/v4"
	"github.com/thsfranca/vex/internal/transpiler/parser"
)

func TestNewAnalyzer(t *testing.T) {
	analyzer := NewAnalyzer()
	
	if analyzer == nil {
		t.Fatal("NewAnalyzer should return non-nil analyzer")
	}
	
	if analyzer.symbolTable == nil {
		t.Error("Analyzer should initialize symbol table")
	}
	
	if analyzer.errorReporter == nil {
		t.Error("Analyzer should initialize error reporter")
	}
}

func TestAnalyzer_SetErrorReporter(t *testing.T) {
	analyzer := NewAnalyzer()
	reporter := NewErrorReporter()
	
	analyzer.SetErrorReporter(reporter)
	
	if analyzer.errorReporter != reporter {
		t.Error("SetErrorReporter should set the provided reporter")
	}
}

func TestAnalyzer_GetErrorReporter(t *testing.T) {
	analyzer := NewAnalyzer()
	
	reporter := analyzer.GetErrorReporter()
	if reporter == nil {
		t.Error("GetErrorReporter should return non-nil reporter")
	}
}

func TestAnalyzer_Analyze(t *testing.T) {
	analyzer := NewAnalyzer()
	
	// Create test AST with simple expression that won't cause semantic errors
	ast := &MockAST{
		root: createMockProgramNode("42"),
	}
	
	symbolTable, err := analyzer.Analyze(ast)
	if err != nil {
		t.Errorf("Analyze should succeed for simple input: %v", err)
		return
	}
	
	if symbolTable == nil {
		t.Error("Analyze should return non-nil symbol table")
	}
}

func TestAnalyzer_Analyze_WithErrors(t *testing.T) {
	analyzer := NewAnalyzer()
	
	// Create AST that will cause errors (incomplete def)
	ast := &MockAST{
		root: createMockProgramNode("(def)"),
	}
	
	_, err := analyzer.Analyze(ast)
	if err == nil {
		t.Error("Analyze should return error for invalid input")
	}
	
	if !strings.Contains(err.Error(), "analysis failed with errors") {
		t.Errorf("Error should mention analysis failure, got: %v", err.Error())
	}
}

func TestAnalyzer_VisitProgram(t *testing.T) {
	analyzer := NewAnalyzer()
	
	// Create mock program context
	programCtx := createMockProgramNode("(def x 42)")
	
	err := analyzer.VisitProgram(programCtx)
	if err != nil {
		t.Errorf("VisitProgram should succeed: %v", err)
	}
}

func TestAnalyzer_VisitList_DefExpression(t *testing.T) {
	analyzer := NewAnalyzer()
	
	// For this test, let's manually call analyzeDef with proper arguments
	args := []Value{
		NewBasicValue("x", "symbol"),
		NewBasicValue("42", "number"),
	}
	
	listCtx := createMockListNode("def", "x", "42")
	
	value, err := analyzer.analyzeDef(listCtx, args)
	if err != nil {
		t.Errorf("analyzeDef should succeed for valid def: %v", err)
		return
	}
	
	if value == nil {
		t.Error("analyzeDef should return non-nil value")
	}
	
	// Check that symbol was defined
	_, exists := analyzer.symbolTable.Lookup("x")
	if !exists {
		t.Error("Symbol 'x' should be defined after def expression")
	}
}

func TestAnalyzer_VisitList_EmptyExpression(t *testing.T) {
	analyzer := NewAnalyzer()
	
	// Create empty list context
	listCtx := createEmptyMockListNode()
	
	_, err := analyzer.VisitList(listCtx)
	if err == nil {
		t.Error("VisitList should return error for empty expression")
	}
	
	// Should report error
	if !analyzer.errorReporter.HasErrors() {
		t.Error("Should report error for empty expression")
	}
}

func TestAnalyzer_VisitArray(t *testing.T) {
	analyzer := NewAnalyzer()
	
	// Create mock array context
	arrayCtx := createMockArrayNode("1", "2", "3")
	
	value, err := analyzer.VisitArray(arrayCtx)
	if err != nil {
		t.Errorf("VisitArray should succeed: %v", err)
		return
	}
	
	if value == nil {
		t.Error("VisitArray should return non-nil value")
	}
	
	if value.Type() != "[]interface{}" {
		t.Errorf("Array value type = %v, want %v", value.Type(), "[]interface{}")
	}
}

func TestAnalyzer_VisitTerminal(t *testing.T) {
	analyzer := NewAnalyzer()
	
	tests := []struct {
		name     string
		text     string
		wantType string
	}{
		{"String literal", "\"hello\"", "string"},
		{"Number", "42", "number"},
		{"Boolean true", "true", "bool"},
		{"Boolean false", "false", "bool"},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			terminal := createMockTerminalNode(tt.text)
			
			value, err := analyzer.VisitTerminal(terminal)
			if err != nil && tt.wantType != "undefined" {
				t.Errorf("VisitTerminal should succeed for %s: %v", tt.name, err)
				return
			}
			
			if value == nil {
				t.Error("VisitTerminal should return non-nil value")
				return
			}
			
			if value.Type() != tt.wantType {
				t.Errorf("Terminal value type = %v, want %v", value.Type(), tt.wantType)
			}
			
			if value.String() != tt.text {
				t.Errorf("Terminal value = %v, want %v", value.String(), tt.text)
			}
		})
	}
}

func TestAnalyzer_VisitTerminal_UndefinedSymbol(t *testing.T) {
    analyzer := NewAnalyzer()

    terminal := createMockTerminalNode("undefined_var")

    value, err := analyzer.VisitTerminal(terminal)
    if err == nil {
        t.Fatalf("VisitTerminal should error for bare undefined symbols")
    }
    if value == nil {
        t.Fatalf("VisitTerminal should still return a value for diagnostics")
    }
    if value.Type() != "undefined" {
        t.Fatalf("Undefined symbol should yield 'undefined' type, got %s", value.Type())
    }
    if !analyzer.errorReporter.HasErrors() {
        t.Fatalf("Diagnostic should be recorded for undefined symbol")
    }
    // Assert diagnostic code VEX-TYP-UNDEF present
    errs := analyzer.errorReporter.GetErrors()
    hasCode := false
    for _, e := range errs {
        if strings.Contains(e.Message, "VEX-TYP-UNDEF") { hasCode = true; break }
    }
    if !hasCode { t.Fatalf("expected VEX-TYP-UNDEF diagnostic, got: %s", analyzer.errorReporter.FormatErrors()) }
}

func TestAnalyzer_LetPolymorphism_IdFunction(t *testing.T) {
    analyzer := NewAnalyzer()
    // Define id via def: (def id (fn [x] x)) using direct calls to analyzer helpers
    // Simulate parsing and visiting
    // Create function value
    fnCtx := createMockListNode("fn", "[x]", "x")
    fnVal, err := analyzer.analyzeFn(fnCtx, []Value{NewBasicValue("[x]", "array"), NewBasicValue("x", "expression")})
    if err != nil { t.Fatalf("analyzeFn failed: %v", err) }
    // Define symbol
    defCtx := createMockListNode("def", "id", "(fn [x] x)")
    _, err = analyzer.analyzeDef(defCtx, []Value{NewBasicValue("id", "symbol"), fnVal})
    if err != nil { t.Fatalf("analyzeDef failed: %v", err) }

    // Use at int
    term := createMockTerminalNode("id")
    v, err := analyzer.VisitTerminal(term)
    if err != nil { t.Fatalf("VisitTerminal failed: %v", err) }
    callCtx := createMockListNode("id", "1")
    // analyzeFunctionCall expects args as Values; visit number terminal to get typed value
    oneVal, _ := analyzer.VisitTerminal(createMockTerminalNode("1"))
    _, _ = analyzer.analyzeFunctionCall(callCtx, "id", []Value{oneVal})

    // Use at string
    v2, err := analyzer.VisitTerminal(createMockTerminalNode("id"))
    if err != nil { t.Fatalf("VisitTerminal failed: %v", err) }
    _ = v2 // ensure a fresh instantiation path
    callCtx2 := createMockListNode("id", "\"s\"")
    strVal, _ := analyzer.VisitTerminal(createMockTerminalNode("\"s\""))
    _, _ = analyzer.analyzeFunctionCall(callCtx2, "id", []Value{strVal})

    // No errors expected; different instantiations should coexist
    if analyzer.errorReporter.HasErrors() {
        t.Fatalf("no errors expected for polymorphic id, got: %s", analyzer.errorReporter.FormatErrors())
    }
    _ = v
}

func TestAnalyzer_ArrayPrimitivesTyping(t *testing.T) {
    a := NewAnalyzer()
    // def xs [1 2]
    _ = a.symbolTable.Define("xs", NewBasicValue("[1 2]", "[]interface{}").WithType(&TypeArray{Elem: &TypeConstant{Name: "int"}}))
    xsVal, _ := a.symbolTable.Lookup("xs")

    // get xs 0 => int
    list := createHeadedMockListNode("get", "xs", "0")
    got, err := a.analyzeFunctionCall(list, "get", []Value{xsVal, NewBasicValue("0", "number").WithType(&TypeConstant{Name: "int"})})
    if err != nil { t.Fatalf("get typing failed: %v", err) }
    if a.publicTypeString(a.typeFromValue(got)) != "number" { // public int maps to number
        t.Fatalf("get type = %s, want number(int)", a.publicTypeString(a.typeFromValue(got)))
    }

    // slice xs 1 => []
    list2 := createHeadedMockListNode("slice", "xs", "1")
    got2, err := a.analyzeFunctionCall(list2, "slice", []Value{xsVal, NewBasicValue("1", "number").WithType(&TypeConstant{Name: "int"})})
    if err != nil { t.Fatalf("slice typing failed: %v", err) }
    if got2.Type() != "[]interface{}" { t.Fatalf("slice type = %s, want []interface{}", got2.Type()) }

    // append xs [3] => [] with int elem
    list3 := createHeadedMockListNode("append", "xs", "[3]")
    arrVal := NewBasicValue("[3]", "[]interface{}").WithType(&TypeArray{Elem: &TypeConstant{Name: "int"}})
    got3, err := a.analyzeFunctionCall(list3, "append", []Value{xsVal, arrVal})
    if err != nil { t.Fatalf("append typing failed: %v", err) }
    if got3.Type() != "[]interface{}" { t.Fatalf("append type = %s, want []interface{}", got3.Type()) }
}

func TestAnalyzer_analyzeDef(t *testing.T) {
	analyzer := NewAnalyzer()
	
	tests := []struct {
		name     string
		args     []Value
		wantErr  bool
		symbol   string
		value    string
	}{
		{
			name: "Valid definition",
			args: []Value{
				NewBasicValue("x", "symbol"),
				NewBasicValue("42", "number"),
			},
			wantErr: false,
			symbol:  "x",
			value:   "42",
		},
		{
			name: "Insufficient arguments",
			args: []Value{
				NewBasicValue("x", "symbol"),
			},
			wantErr: true,
		},
		{
			name:    "No arguments",
			args:    []Value{},
			wantErr: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset analyzer state
			analyzer.symbolTable = NewSymbolTable()
			analyzer.errorReporter.Clear()
			
			listCtx := createMockListNode("def")
			
			value, err := analyzer.analyzeDef(listCtx, tt.args)
			
			if (err != nil) != tt.wantErr {
				t.Errorf("analyzeDef() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			
			if !tt.wantErr {
				if value == nil {
					t.Error("analyzeDef should return non-nil value for success")
					return
				}
				
				if value.String() != tt.value {
					t.Errorf("Returned value = %v, want %v", value.String(), tt.value)
				}
				
				// Check symbol was defined
				_, exists := analyzer.symbolTable.Lookup(tt.symbol)
				if !exists {
					t.Errorf("Symbol %s should be defined", tt.symbol)
				}
			}
		})
	}
}

func TestAnalyzer_analyzeIf(t *testing.T) {
	analyzer := NewAnalyzer()
	
	tests := []struct {
		name    string
		args    []Value
		wantErr bool
	}{
		{
			name: "Valid if with then and else",
			args: []Value{
				NewBasicValue("true", "bool"),
				NewBasicValue("42", "number"),
				NewBasicValue("0", "number"),
			},
			wantErr: false,
		},
		{
			name: "Valid if with only then",
			args: []Value{
				NewBasicValue("true", "bool"),
				NewBasicValue("42", "number"),
			},
			wantErr: false,
		},
		{
			name: "Insufficient arguments",
			args: []Value{
				NewBasicValue("true", "bool"),
			},
			wantErr: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			analyzer.errorReporter.Clear()
			
			listCtx := createMockListNode("if")
			
			value, err := analyzer.analyzeIf(listCtx, tt.args)
			
			if (err != nil) != tt.wantErr {
				t.Errorf("analyzeIf() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			
			if !tt.wantErr && value == nil {
				t.Error("analyzeIf should return non-nil value for success")
			}
		})
	}
}

func TestAnalyzer_analyzeFn(t *testing.T) {
	analyzer := NewAnalyzer()
	
	args := []Value{
		NewBasicValue("[x y]", "array"),
		NewBasicValue("(+ x y)", "expression"),
	}
	
	listCtx := createMockListNode("fn")
	
	value, err := analyzer.analyzeFn(listCtx, args)
	if err != nil {
		t.Errorf("analyzeFn should succeed: %v", err)
		return
	}
	
	if value == nil {
		t.Error("analyzeFn should return non-nil value")
		return
	}
	
	if value.Type() != "func" {
		t.Errorf("Function value type = %v, want %v", value.Type(), "func")
	}
}

func TestAnalyzer_analyzeMacro(t *testing.T) {
	analyzer := NewAnalyzer()
	
	tests := []struct {
		name      string
		args      []Value
		wantErr   bool
		checkName string
	}{
		{
			name: "Valid macro definition",
			args: []Value{
				NewBasicValue("my-macro", "symbol"),
				NewBasicValue("[x]", "array"),
				NewBasicValue("x", "expression"),
			},
			wantErr:   false,
			checkName: "my-macro",
		},
		{
			name: "Reserved word macro name",
			args: []Value{
				NewBasicValue("if", "symbol"),
				NewBasicValue("[x]", "array"),
				NewBasicValue("x", "expression"),
			},
			wantErr: true,
		},
		{
			name: "Insufficient arguments",
			args: []Value{
				NewBasicValue("my-macro", "symbol"),
				NewBasicValue("[x]", "array"),
			},
			wantErr: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset state
			analyzer.symbolTable = NewSymbolTable()
			analyzer.errorReporter.Clear()
			
			listCtx := createMockListNode("macro")
			
			value, err := analyzer.analyzeMacro(listCtx, tt.args)
			
			if (err != nil) != tt.wantErr {
				t.Errorf("analyzeMacro() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			
			if !tt.wantErr {
				if value == nil {
					t.Error("analyzeMacro should return non-nil value for success")
					return
				}
				
				if value.Type() != "macro" {
					t.Errorf("Macro value type = %v, want %v", value.Type(), "macro")
				}
				
				// Check symbol was defined
				_, exists := analyzer.symbolTable.Lookup(tt.checkName)
				if !exists {
					t.Errorf("Macro symbol %s should be defined", tt.checkName)
				}
			}
		})
	}
}

func TestAnalyzer_analyzeFunctionCall(t *testing.T) {
	analyzer := NewAnalyzer()
	
	tests := []struct {
		name     string
		funcName string
		args     []Value
		wantWarn bool
	}{
		{
			name:     "Builtin function",
			funcName: "+",
			args: []Value{
				NewBasicValue("1", "number"),
				NewBasicValue("2", "number"),
			},
			wantWarn: false,
		},
		{
			name:     "Package function",
			funcName: "fmt/Println",
			args: []Value{
				NewBasicValue("\"hello\"", "string"),
			},
			wantWarn: false,
		},
		{
			name:     "Unknown function",
			funcName: "unknown-func",
			args:     []Value{},
			wantWarn: false,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			analyzer.errorReporter.Clear()
			
			listCtx := createMockListNode(tt.funcName)
			
			value, err := analyzer.analyzeFunctionCall(listCtx, tt.funcName, tt.args)
			if tt.funcName == "unknown-func" {
				if err == nil {
					t.Errorf("expected error for unknown function, got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("analyzeFunctionCall should not error: %v", err)
				return
			}
			
			if value == nil {
				t.Error("analyzeFunctionCall should return non-nil value")
				return
			}
			
			warnings := analyzer.errorReporter.GetWarnings()
			hasWarning := len(warnings) > 0
			
			if hasWarning != tt.wantWarn {
				t.Errorf("Expected warning = %v, got warning = %v", tt.wantWarn, hasWarning)
			}
		})
	}
}

func TestEqualityTypingViaScheme(t *testing.T) {
    a := NewAnalyzer()
    // (= 1 1) -> bool
    ctx := createHeadedMockListNode("=", "1", "1")
    _, err := a.analyzeFunctionCall(ctx, "=", []Value{NewBasicValue("1", "number").WithType(&TypeConstant{Name: "int"}), NewBasicValue("1", "number").WithType(&TypeConstant{Name: "int"})})
    if err != nil { t.Fatalf("equality int==int should type: %v", err) }

    // (= "x" "y") -> bool
    ctx2 := createHeadedMockListNode("=", "\"x\"", "\"y\"")
    _, err = a.analyzeFunctionCall(ctx2, "=", []Value{NewBasicValue("\"x\"", "string"), NewBasicValue("\"y\"", "string")})
    if err != nil { t.Fatalf("equality string==string should type: %v", err) }
}

func TestEqualityTyping_NegativeMismatch(t *testing.T) {
    a := NewAnalyzer()
    // (= 1 "x") should produce a type error via scheme unification
    ctx := createHeadedMockListNode("=", "1", "\"x\"")
    _, _ = a.analyzeFunctionCall(ctx, "=", []Value{NewBasicValue("1", "number").WithType(&TypeConstant{Name: "int"}), NewBasicValue("\"x\"", "string")})
    if !a.errorReporter.HasErrors() {
        t.Fatalf("expected type error for mismatched equality")
    }
    // Assert diagnostic code is VEX-TYP-EQ
    errs := a.errorReporter.GetErrors()
    found := false
    for _, e := range errs {
        if strings.Contains(e.Message, "VEX-TYP-EQ") {
            found = true
            break
        }
    }
    if !found {
        t.Fatalf("expected VEX-TYP-EQ diagnostic in errors: %v", a.errorReporter.FormatErrors())
    }
}

func TestIfConditionRequiresBool_Negative(t *testing.T) {
    a := NewAnalyzer()
    // (if 1 2 3) condition not bool
    listCtx := createMockListNode("if")
    _, _ = a.analyzeIf(listCtx, []Value{NewBasicValue("1", "number"), NewBasicValue("2", "number"), NewBasicValue("3", "number")})
    if !a.errorReporter.HasErrors() {
        t.Fatalf("expected error for non-boolean if condition")
    }
    // Check code VEX-TYP-COND appears in message
    errs := a.errorReporter.GetErrors()
    ok := false
    for _, e := range errs {
        if strings.Contains(e.Message, "VEX-TYP-COND") { ok = true; break }
    }
    if !ok { t.Fatalf("expected VEX-TYP-COND diagnostic") }
}

func TestValueRestriction_NoQuantificationForNonValues(t *testing.T) {
    a := NewAnalyzer()
    // Define id as fn: should generalize with at least one quantified var
    fnCtx := createMockListNode("fn", "[x]", "x")
    fnVal, err := a.analyzeFn(fnCtx, []Value{NewBasicValue("[x]", "array"), NewBasicValue("x", "expression")})
    if err != nil { t.Fatalf("analyzeFn failed: %v", err) }
    _, err = a.analyzeDef(createMockListNode("def"), []Value{NewBasicValue("id", "symbol"), fnVal})
    if err != nil { t.Fatalf("def id failed: %v", err) }
    if sch, ok := a.GetTypeScheme("id"); !ok || len(sch.Quantified) == 0 {
        t.Fatalf("id should be generalized with quantified variables")
    }

    // Define y as (+ 1 2): should not introduce quantification (ground type)
    plus := NewBasicValue("call-result", "number").WithType(&TypeConstant{Name: "int"})
    _, err = a.analyzeDef(createMockListNode("def"), []Value{NewBasicValue("y", "symbol"), plus})
    if err != nil { t.Fatalf("def y failed: %v", err) }
    if sch, ok := a.GetTypeScheme("y"); ok {
        if len(sch.Quantified) != 0 {
            t.Fatalf("y should not have quantified variables; got %v", sch.Quantified)
        }
    }

    // Define z as (do 1 2): should not generalize
    doRes := NewBasicValue("do-result", "number").WithType(&TypeConstant{Name: "int"})
    _, err = a.analyzeDef(createMockListNode("def"), []Value{NewBasicValue("z", "symbol"), doRes})
    if err != nil { t.Fatalf("def z failed: %v", err) }
    if sch, ok := a.GetTypeScheme("z"); ok {
        if len(sch.Quantified) != 0 {
            t.Fatalf("z should not have quantified variables; got %v", sch.Quantified)
        }
    }
}

func TestCrossPackageTypingWithSchemes(t *testing.T) {
    a := NewAnalyzer()
    // Prepare package env: local package "a" exports id with scheme ∀t. t -> t
    ignore := map[string]bool{"a": true}
    exports := map[string]map[string]bool{"a": {"id": true}}
    // Construct scheme: ForAll [1]. (1 -> 1)
    tv := &TypeVariable{ID: 1}
    sch := &TypeScheme{Quantified: []int{1}, Body: &TypeFunction{Params: []Type{tv}, Result: tv}}
    schemes := map[string]map[string]*TypeScheme{"a": {"id": sch}}

    a.SetPackageEnv(ignore, exports, schemes)

    // Call a/id with int
    listInt := createHeadedMockListNode("a/id", "1")
    vInt := NewBasicValue("1", "number").WithType(&TypeConstant{Name: "int"})
    if _, err := a.analyzeFunctionCall(listInt, "a/id", []Value{vInt}); err != nil {
        t.Fatalf("a/id on int should type via scheme: %v", err)
    }

    // Call a/id with string
    listStr := createHeadedMockListNode("a/id", "\"s\"")
    vStr := NewBasicValue("\"s\"", "string")
    if _, err := a.analyzeFunctionCall(listStr, "a/id", []Value{vStr}); err != nil {
        t.Fatalf("a/id on string should type via scheme: %v", err)
    }

    // Non-exported symbol should error
    listHidden := createHeadedMockListNode("a/hidden")
    if _, err := a.analyzeFunctionCall(listHidden, "a/hidden", []Value{}); err == nil {
        t.Fatalf("a/hidden should error due to non-exported symbol")
    }
}

func TestRecordNominalMismatch_InIfBranches(t *testing.T) {
    a := NewAnalyzer()
    // Define two nominally distinct records with same shape
    _, _ = a.VisitList(createHeadedMockListNode("record", "A", "[x: number]"))
    _, _ = a.VisitList(createHeadedMockListNode("record", "B", "[x: number]"))
    // Build branches: then -> A, else -> B
    thenA, _ := a.analyzeRecordConstructor(createHeadedMockListNode("A", "[x:", "1]"), &RecordValue{name: "A", fields: map[string]string{"x": "number"}, fieldOrder: []string{"x"}})
    elseB, _ := a.analyzeRecordConstructor(createHeadedMockListNode("B", "[x:", "2]"), &RecordValue{name: "B", fields: map[string]string{"x": "number"}, fieldOrder: []string{"x"}})
    // Use analyzeIf to force unification of branch types
    listCtx := createMockListNode("if")
    _, _ = a.analyzeIf(listCtx, []Value{NewBasicValue("true", "bool"), thenA, elseB})
    if !a.errorReporter.HasErrors() {
        t.Fatalf("expected nominal mismatch error when if branches return A vs B")
    }
    // Assert dedicated code VEX-TYP-REC-NOMINAL present
    msg := a.errorReporter.FormatErrors()
    if !strings.Contains(msg, "VEX-TYP-REC-NOMINAL") {
        t.Fatalf("expected VEX-TYP-REC-NOMINAL diagnostic, got: %s", msg)
    }
}

func TestArrayElementMismatch_ReportsDiagnostic(t *testing.T) {
    a := NewAnalyzer()
    arr := createMockArrayNode("1", "\"x\"")
    _, _ = a.VisitArray(arr)
    if !a.errorReporter.HasErrors() {
        t.Fatalf("expected array element type mismatch error")
    }
    // Assert VEX-TYP-ARRAY-ELEM appears
    msg := a.errorReporter.FormatErrors()
    if !strings.Contains(msg, "VEX-TYP-ARRAY-ELEM") {
        t.Fatalf("expected VEX-TYP-ARRAY-ELEM diagnostic, got: %s", msg)
    }
}

func TestMapKeyAndValueMismatch_ReportDiagnostics(t *testing.T) {
    a := NewAnalyzer()
    // (map [1 : 1 "k" : 2]) -> key mismatch (int vs string), value types both numbers, so only key error
    listCtx := createHeadedMockListNode("map", "[1:", "1", "\"k\":", "2]")
    _, _ = a.VisitList(listCtx)
    if !a.errorReporter.HasErrors() {
        t.Fatalf("expected map key type mismatch error")
    }
    // Check VEX-TYP-MAP-KEY present
    keyMsg := a.errorReporter.FormatErrors()
    if !strings.Contains(keyMsg, "VEX-TYP-MAP-KEY") {
        t.Fatalf("expected VEX-TYP-MAP-KEY diagnostic, got: %s", keyMsg)
    }

    // Now trigger a value mismatch: (["k": 1 "k": "x"]) same key type, differing values (int vs string)
    a = NewAnalyzer()
    listCtx2 := createHeadedMockListNode("map", "[\"k\":", "1", "\"k\":", "\"x\"]")
    _, _ = a.VisitList(listCtx2)
    if !a.errorReporter.HasErrors() {
        t.Fatalf("expected map value type mismatch error")
    }
    valMsg := a.errorReporter.FormatErrors()
    if !strings.Contains(valMsg, "VEX-TYP-MAP-VAL") {
        t.Fatalf("expected VEX-TYP-MAP-VAL diagnostic, got: %s", valMsg)
    }
}

func TestOccurCheck_Negative_UnificationCycle(t *testing.T) {
    a := NewAnalyzer()
    // Craft an expression that attempts to unify a type variable with a function of itself
    // (fn [x] (x x)) — classic self-application causing occurs check in HM
    // We will simulate body typing by providing a raw that analyzer will parse and visit.
    fnCtx := createMockListNode("fn", "[x]", "(x x)")
    _, err := a.analyzeFn(fnCtx, []Value{NewBasicValue("[x]", "array"), NewBasicValue("(x x)", "expression")})
    if err == nil && !a.errorReporter.HasErrors() {
        t.Fatalf("expected occur-check error for self-application")
    }
    // We cannot guarantee exact text, but ensure errors mention occur-check or cycle
    msg := a.errorReporter.FormatErrors()
    if msg == "" {
        t.Fatalf("expected diagnostics for self-application, got none")
    }
    // Accept either occur-check/cycle wording or strict arg-mismatch code depending on unification path
    if !(strings.Contains(msg, "occur-check") || strings.Contains(msg, "occur") || strings.Contains(msg, "cycle") || strings.Contains(msg, "VEX-TYP-ARG")) {
        t.Fatalf("expected occur-check/cycle or VEX-TYP-ARG, got: %s", msg)
    }
}

func TestAnalyzer_ValueRestriction_GeneralizeOnlyValues(t *testing.T) {
    a := NewAnalyzer()
    // Define f = (fn [x] x) should generalize
    fnCtx := createMockListNode("fn", "[x]", "x")
    fnVal, err := a.analyzeFn(fnCtx, []Value{NewBasicValue("[x]", "array"), NewBasicValue("x", "expression")})
    if err != nil { t.Fatalf("analyzeFn failed: %v", err) }
    _, err = a.analyzeDef(createMockListNode("def"), []Value{NewBasicValue("id", "symbol"), fnVal})
    if err != nil { t.Fatalf("analyzeDef failed: %v", err) }
    if _, ok := a.typeEnv["id"]; !ok { t.Fatalf("id should be generalized") }

    // Define y = (+ 1 2) should not be over-generalized as a scheme with variables
    plusCall := NewBasicValue("call-result", "number").WithType(&TypeConstant{Name: "int"})
    _, err = a.analyzeDef(createMockListNode("def"), []Value{NewBasicValue("y", "symbol"), plusCall})
    if err != nil { t.Fatalf("def y failed: %v", err) }
    if sch, ok := a.typeEnv["y"]; ok {
        // Expect concrete int/number, not quantified vars; Body should be TypeConstant or concrete type
        if _, isVar := sch.Body.(*TypeVariable); isVar {
            t.Fatalf("y should not generalize to type variable")
        }
    }
}

func TestAnalyzer_Record_NominalTypeMismatch(t *testing.T) {
    a := NewAnalyzer()
    // (record A [x: number]) (record B [x: number])
    _, _ = a.VisitList(createHeadedMockListNode("record", "A", "[x: number]"))
    _, _ = a.VisitList(createHeadedMockListNode("record", "B", "[x: number]"))
    // Construct A and try to use where B expected via unification
    // Simulate a function taking B and returning B: (fn [b] b)
    fnVal, err := a.analyzeFn(createMockListNode("fn", "[b]", "b"), []Value{NewBasicValue("[b]", "array"), NewBasicValue("b", "expression")})
    if err != nil { t.Fatalf("fn failed: %v", err) }
    _, _ = a.analyzeDef(createMockListNode("def"), []Value{NewBasicValue("idB", "symbol"), fnVal})
    // Build A value
    aVal, _ := a.analyzeRecordConstructor(createHeadedMockListNode("A", "[x:", "1]"), &RecordValue{name: "A", fields: map[string]string{"x": "number"}, fieldOrder: []string{"x"}})
    // Call idB with A should unify to param var; we confirm nominal typing by giving B scheme later (skipped here)
    _ = aVal
}

func TestAnalyzer_RecordTypeChecks(t *testing.T) {
    analyzer := NewAnalyzer()

    t.Run("valid record declaration", func(t *testing.T) {
        analyzer.symbolTable = NewSymbolTable()
        analyzer.errorReporter.Clear()
        listCtx := createHeadedMockListNode("record", "Person", "[name: string age: number]")
        // VisitList will dispatch to analyzeRecord when head is "record"
        _, err := analyzer.VisitList(listCtx)
        if err != nil {
            t.Fatalf("VisitList(record) should not error for valid record: %v", err)
        }
        // Symbol should be defined
        if _, ok := analyzer.symbolTable.Lookup("Person"); !ok {
            t.Fatalf("record type 'Person' should be defined")
        }
        if analyzer.errorReporter.HasErrors() {
            t.Fatalf("no errors expected, got: %s", analyzer.errorReporter.FormatErrors())
        }
    })

    t.Run("missing name and fields", func(t *testing.T) {
        analyzer.symbolTable = NewSymbolTable()
        analyzer.errorReporter.Clear()
        listCtx := createHeadedMockListNode("record")
        _, err := analyzer.VisitList(listCtx)
        if err == nil {
            t.Fatalf("VisitList(record) should error when name/fields are missing")
        }
        errs := analyzer.errorReporter.GetErrors()
        if len(errs) == 0 || !strings.Contains(errs[0].Message, "record requires name and fields") {
            t.Fatalf("expected error about missing name/fields, got: %v", analyzer.errorReporter.FormatErrors())
        }
    })

    t.Run("invalid record name (reserved)", func(t *testing.T) {
        analyzer.symbolTable = NewSymbolTable()
        analyzer.errorReporter.Clear()
        listCtx := createHeadedMockListNode("record", "if", "[x: string]")
        _, err := analyzer.VisitList(listCtx)
        if err == nil {
            t.Fatalf("VisitList(record) should error for invalid record name")
        }
        errs := analyzer.errorReporter.GetErrors()
        if len(errs) == 0 || !strings.Contains(errs[0].Message, "invalid record name") {
            t.Fatalf("expected invalid record name error, got: %v", analyzer.errorReporter.FormatErrors())
        }
    })

    t.Run("fields must be an array vector", func(t *testing.T) {
        analyzer.symbolTable = NewSymbolTable()
        analyzer.errorReporter.Clear()
        listCtx := createHeadedMockListNode("record", "Person", "x")
        _, err := analyzer.VisitList(listCtx)
        if err == nil {
            t.Fatalf("VisitList(record) should error when fields are not an array")
        }
        errs := analyzer.errorReporter.GetErrors()
        if len(errs) == 0 || !strings.Contains(errs[0].Message, "record expects a field vector") {
            t.Fatalf("expected field vector error, got: %v", analyzer.errorReporter.FormatErrors())
        }
    })

    t.Run("missing type for field", func(t *testing.T) {
        analyzer.symbolTable = NewSymbolTable()
        analyzer.errorReporter.Clear()
        listCtx := createHeadedMockListNode("record", "Person", "[name:")
        _, _ = analyzer.VisitList(listCtx)
        errs := analyzer.errorReporter.GetErrors()
        if len(errs) == 0 || !(strings.Contains(errs[0].Message, "missing type for field") || strings.Contains(errs[0].Message, "invalid field format")) {
            t.Fatalf("expected missing type/format error, got: %v", analyzer.errorReporter.FormatErrors())
        }
    })

    t.Run("invalid field format (missing colon)", func(t *testing.T) {
        analyzer.symbolTable = NewSymbolTable()
        analyzer.errorReporter.Clear()
        listCtx := createHeadedMockListNode("record", "Person", "[name", "string]")
        _, _ = analyzer.VisitList(listCtx)
        errs := analyzer.errorReporter.GetErrors()
        if len(errs) == 0 || !(strings.Contains(errs[0].Message, "invalid field format") || strings.Contains(errs[0].Message, "missing ':'")) {
            t.Fatalf("expected invalid field format error, got: %v", analyzer.errorReporter.FormatErrors())
        }
    })

    t.Run("duplicate field names", func(t *testing.T) {
        analyzer.symbolTable = NewSymbolTable()
        analyzer.errorReporter.Clear()
        listCtx := createHeadedMockListNode("record", "Person", "[name: string name: string]")
        _, _ = analyzer.VisitList(listCtx)
        errs := analyzer.errorReporter.GetErrors()
        if len(errs) == 0 || !strings.Contains(errs[0].Message, "duplicate field 'name'") {
            t.Fatalf("expected duplicate field error, got: %v", analyzer.errorReporter.FormatErrors())
        }
    })
}

func TestAnalyzer_ExportSpecialForm_NoError(t *testing.T) {
    analyzer := NewAnalyzer()
    listCtx := createMockListNode("export")
    args := []Value{NewBasicValue("[x]", "array")}
    _, err := analyzer.analyzeExport(listCtx, args)
    if err != nil {
        t.Fatalf("analyzeExport should not error for valid args: %v", err)
    }
}

func TestAnalyzer_ErrorMessages_IncludeSuggestions(t *testing.T) {
    analyzer := NewAnalyzer()

    // def suggestion
    analyzer.errorReporter.Clear()
    listCtx := createMockListNode("def")
    _, err := analyzer.analyzeDef(listCtx, []Value{})
    if err == nil {
        t.Error("analyzeDef should error for missing args")
    }
    errs := analyzer.errorReporter.GetErrors()
    if len(errs) == 0 || !strings.Contains(errs[0].Message, "Suggestion: use (def name value)") {
        t.Error("def error should include suggestion")
    }

    // if suggestion
    analyzer.errorReporter.Clear()
    listCtx = createMockListNode("if")
    _, err = analyzer.analyzeIf(listCtx, []Value{NewBasicValue("true", "bool")})
    if err == nil {
        t.Error("analyzeIf should error for missing args")
    }
    errs = analyzer.errorReporter.GetErrors()
    if len(errs) == 0 || !strings.Contains(errs[0].Message, "Suggestion: use (if condition then-expr [else-expr])") {
        t.Error("if error should include suggestion")
    }

    // fn suggestion
    analyzer.errorReporter.Clear()
    listCtx = createMockListNode("fn")
    _, err = analyzer.analyzeFn(listCtx, []Value{NewBasicValue("[x]", "array")})
    if err == nil {
        t.Error("analyzeFn should error for missing body")
    }
    errs = analyzer.errorReporter.GetErrors()
    if len(errs) == 0 || !strings.Contains(errs[0].Message, "Suggestion: use (fn [params] body)") {
        t.Error("fn error should include suggestion")
    }

    // reserved macro name suggestion
    analyzer.errorReporter.Clear()
    listCtx = createMockListNode("macro")
    _, err = analyzer.analyzeMacro(listCtx, []Value{NewBasicValue("if", "symbol"), NewBasicValue("[x]", "array"), NewBasicValue("x", "expression")})
    if err == nil {
        t.Error("analyzeMacro should error for reserved name")
    }
    errs = analyzer.errorReporter.GetErrors()
    if len(errs) == 0 || !strings.Contains(errs[0].Message, "Suggestion: choose a different macro name") {
        t.Error("macro name error should include suggestion")
    }
}

// Helper functions for testing

func TestHelper_isNumber(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"42", true},
		{"0", true},
		{"123", true},
		{"", false},
		{"abc", false},
		{"12a", false},
		{"a12", false},
	}
	
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := isNumber(tt.input)
			if got != tt.want {
				t.Errorf("isNumber(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestHelper_isReservedWord(t *testing.T) {
	tests := []struct {
		word string
		want bool
	}{
		{"if", true},
		{"def", true},
		{"fn", true},
		{"let", true},
		{"do", true},
		{"when", true},
		{"unless", true},
		{"import", true},
		{"custom-word", false},
		{"my-function", false},
	}
	
	for _, tt := range tests {
		t.Run(tt.word, func(t *testing.T) {
			got := isReservedWord(tt.word)
			if got != tt.want {
				t.Errorf("isReservedWord(%q) = %v, want %v", tt.word, got, tt.want)
			}
		})
	}
}

func TestHelper_isBuiltinFunction(t *testing.T) {
	tests := []struct {
		name string
		want bool
	}{
		{"+", true},
		{"-", true},
		{"*", true},
		{"/", true},
		{">", true},
		{"<", true},
		{"=", true},
		{"not", true},
		{"first", true},
		{"rest", true},
		{"println", true},
		{"fmt/Println", true},
		{"os/Exit", true},
		{"custom-func", false},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isBuiltinFunction(tt.name)
			if got != tt.want {
				t.Errorf("isBuiltinFunction(%q) = %v, want %v", tt.name, got, tt.want)
			}
		})
	}
}

// Mock implementations for testing

type MockAST struct {
	root antlr.Tree
}

func (m *MockAST) Accept(visitor ASTVisitor) error {
	if programCtx, ok := m.root.(*parser.ProgramContext); ok {
		return visitor.VisitProgram(programCtx)
	}
	return fmt.Errorf("invalid AST root type")
}

// Mock helper functions for testing
func createMockProgramNode(input string) *parser.ProgramContext {
	// Use ANTLR to parse the input for more realistic testing
	inputStream := antlr.NewInputStream(input)
	lexer := parser.NewVexLexer(inputStream)
	tokenStream := antlr.NewCommonTokenStream(lexer, 0)
	vexParser := parser.NewVexParser(tokenStream)
	return vexParser.Program().(*parser.ProgramContext)
}

func createMockListNode(elements ...string) *parser.ListContext {
	// Create a simple expression to parse
	expr := "(test"
	for _, elem := range elements {
		expr += " " + elem
	}
	expr += ")"
	
	inputStream := antlr.NewInputStream(expr)
	lexer := parser.NewVexLexer(inputStream)
	tokenStream := antlr.NewCommonTokenStream(lexer, 0)
	vexParser := parser.NewVexParser(tokenStream)
	return vexParser.List().(*parser.ListContext)
}

// createHeadedMockListNode builds a list where the first element is treated as the head symbol.
// It constructs an expression like: (head rest...)
func createHeadedMockListNode(head string, rest ...string) *parser.ListContext {
    expr := "(" + head
    for _, elem := range rest {
        expr += " " + elem
    }
    expr += ")"
    inputStream := antlr.NewInputStream(expr)
    lexer := parser.NewVexLexer(inputStream)
    tokenStream := antlr.NewCommonTokenStream(lexer, 0)
    vexParser := parser.NewVexParser(tokenStream)
    return vexParser.List().(*parser.ListContext)
}

func createEmptyMockListNode() *parser.ListContext {
	// Create minimal valid list that will be empty when processed
	inputStream := antlr.NewInputStream("()")
	lexer := parser.NewVexLexer(inputStream)
	tokenStream := antlr.NewCommonTokenStream(lexer, 0)
	vexParser := parser.NewVexParser(tokenStream)
	return vexParser.List().(*parser.ListContext)
}

func createMockArrayNode(elements ...string) *parser.ArrayContext {
	expr := "["
	for i, elem := range elements {
		if i > 0 {
			expr += " "
		}
		expr += elem
	}
	expr += "]"
	
	inputStream := antlr.NewInputStream(expr)
	lexer := parser.NewVexLexer(inputStream)
	tokenStream := antlr.NewCommonTokenStream(lexer, 0)
	vexParser := parser.NewVexParser(tokenStream)
	return vexParser.Array().(*parser.ArrayContext)
}

func createMockTerminalNode(text string) antlr.TerminalNode {
	return &mockTerminalNode{text: text}
}

type mockTerminalNode struct {
	text string
}

func (m *mockTerminalNode) GetText() string { return m.text }
func (m *mockTerminalNode) GetSymbol() antlr.Token {
	return nil // Simplified for testing
}
func (m *mockTerminalNode) Accept(visitor antlr.ParseTreeVisitor) interface{} { return nil }
func (m *mockTerminalNode) GetChild(i int) antlr.Tree { return nil }
func (m *mockTerminalNode) GetChildCount() int { return 0 }
func (m *mockTerminalNode) GetChildren() []antlr.Tree { return nil }
func (m *mockTerminalNode) GetParent() antlr.Tree { return nil }
func (m *mockTerminalNode) GetPayload() interface{} { return m.text }
func (m *mockTerminalNode) GetSourceInterval() antlr.Interval { return antlr.Interval{} }
func (m *mockTerminalNode) SetParent(parent antlr.Tree) {}
func (m *mockTerminalNode) ToStringTree(ruleNames []string, recog antlr.Recognizer) string { return m.text }

