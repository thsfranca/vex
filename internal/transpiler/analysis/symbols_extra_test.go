package analysis

import "testing"

func TestBasicValue_RawAndTypeHelpers(t *testing.T) {
    v := NewBasicValue("x", "symbol").MarkRaw()
    if !v.isRaw() { t.Fatalf("expected raw to be true") }
    if v.getType() != nil { t.Fatalf("unexpected non-nil type before WithType") }
    v.WithType(&TypeConstant{Name: "int"})
    if v.getType() == nil { t.Fatalf("expected type after WithType") }
}

func TestRecordValue_FieldOrderAndCopy(t *testing.T) {
    r := NewRecordValue("User", map[string]string{"name": "string", "age": "int"}, []string{"name", "age"})
    order := r.GetFieldOrder()
    if len(order) != 2 || order[0] != "name" || order[1] != "age" {
        t.Fatalf("unexpected order: %#v", order)
    }
    f := r.GetFields()
    f["name"] = "hacked"
    if r.GetFields()["name"] != "string" { t.Fatalf("GetFields should return copy") }
}


