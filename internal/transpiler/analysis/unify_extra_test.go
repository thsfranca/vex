package analysis

import "testing"

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


