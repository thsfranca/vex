package analysis

import "testing"

// Merged from unify_extra_test.go and unify_more_extra_test.go
func TestUnify_TypeConstantMismatch_Error(t *testing.T) {
    if _, err := unify(&TypeConstant{Name: "int"}, &TypeConstant{Name: "bool"}); err == nil {
        t.Fatalf("expected mismatch error for int vs bool")
    }
}

func TestUnify_FunctionArityMismatch_Error(t *testing.T) {
    f1 := &TypeFunction{Params: []Type{&TypeConstant{Name: "int"}}, Result: &TypeConstant{Name: "int"}}
    f2 := &TypeFunction{Params: []Type{&TypeConstant{Name: "int"}, &TypeConstant{Name: "int"}}, Result: &TypeConstant{Name: "int"}}
    if _, err := unify(f1, f2); err == nil {
        t.Fatalf("expected arity mismatch error")
    }
}

func TestUnify_FunctionUnify_Success(t *testing.T) {
    f1 := &TypeFunction{Params: []Type{&TypeConstant{Name: "int"}}, Result: &TypeConstant{Name: "int"}}
    f2 := &TypeFunction{Params: []Type{&TypeConstant{Name: "number"}}, Result: &TypeConstant{Name: "number"}}
    if _, err := unify(f1, f2); err != nil {
        t.Fatalf("expected function unify success, got: %v", err)
    }
}

func TestUnify_FunctionWithArray_Error(t *testing.T) {
    f := &TypeFunction{Params: []Type{}, Result: &TypeConstant{Name: "int"}}
    arr := &TypeArray{Elem: &TypeConstant{Name: "int"}}
    if _, err := unify(f, arr); err == nil {
        t.Fatalf("expected error when unifying function with array")
    }
}

func TestUnify_MapUnify_SuccessAndMismatch(t *testing.T) {
    m1 := &TypeMap{Key: &TypeConstant{Name: "int"}, Val: &TypeConstant{Name: "string"}}
    m2 := &TypeMap{Key: &TypeConstant{Name: "number"}, Val: &TypeConstant{Name: "string"}}
    if _, err := unify(m1, m2); err != nil {
        t.Fatalf("expected map unify success on numeric key family: %v", err)
    }
    // Mismatch on value type
    m3 := &TypeMap{Key: &TypeConstant{Name: "int"}, Val: &TypeConstant{Name: "bool"}}
    if _, err := unify(m1, m3); err == nil {
        t.Fatalf("expected map value mismatch error")
    }
}

func TestUnify_SymmetricVariableBinding(t *testing.T) {
    tv := &TypeVariable{ID: 7}
    if s, err := unify(&TypeConstant{Name: "int"}, tv); err != nil || s[7] == nil {
        t.Fatalf("expected symmetric bind of var to int, got: %v %#v", err, s)
    }
}

func TestUnify_OccurCheckAndArray(t *testing.T) {
    // occur-check: a ~ [a] should fail
    a := &TypeVariable{ID: 1}
    arr := &TypeArray{Elem: a}
    if _, err := unify(a, arr); err == nil {
        t.Fatalf("expected occur-check failure when unifying a with [a]")
    }

    // Arrays unify element-wise
    s, err := unify(&TypeArray{Elem: &TypeConstant{Name: "int"}}, &TypeArray{Elem: &TypeConstant{Name: "number"}})
    if err != nil { t.Fatalf("array unify should succeed via numeric family: %v", err) }
    // Apply substitution
    got := (&TypeArray{Elem: &TypeConstant{Name: "int"}}).apply(s)
    if _, ok := got.(*TypeArray); !ok {
        t.Fatalf("expected array after applying substitution")
    }
}


