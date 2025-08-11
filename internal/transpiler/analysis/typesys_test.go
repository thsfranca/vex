package analysis

import "testing"

// Moved from typesys_extra_test.go
func TestTypeMap_ApplyAndFreeVars(t *testing.T) {
    tvK := &TypeVariable{ID: 10}
    tvV := &TypeVariable{ID: 20}
    tm := &TypeMap{Key: tvK, Val: tvV}
    fvs := tm.freeTypeVars()
    if _, ok := fvs[10]; !ok { t.Fatalf("missing key var in free vars") }
    if _, ok := fvs[20]; !ok { t.Fatalf("missing val var in free vars") }

    s := Subst{10: &TypeConstant{Name: "int"}, 20: &TypeConstant{Name: "string"}}
    applied := tm.apply(s).(*TypeMap)
    if _, ok := applied.Key.(*TypeConstant); !ok { t.Fatalf("key not substituted") }
    if _, ok := applied.Val.(*TypeConstant); !ok { t.Fatalf("val not substituted") }
}


