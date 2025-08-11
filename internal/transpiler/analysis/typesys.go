package analysis

// Type represents a type in the type system.
// Implementations include TypeVariable, TypeConstant, TypeFunction, TypeArray.
type Type interface {
    apply(s Subst) Type
    freeTypeVars() map[int]struct{}
}

// Subst is a substitution mapping type variable IDs to types.
type Subst map[int]Type

func (s Subst) compose(other Subst) Subst {
    // Apply s to all in other, then merge
    out := make(Subst, len(other)+len(s))
    for k, v := range other {
        out[k] = v.apply(s)
    }
    for k, v := range s {
        out[k] = v
    }
    return out
}

// TypeVariable is a unification variable.
type TypeVariable struct {
    ID int
}

func (t *TypeVariable) apply(s Subst) Type {
    if rep, ok := s[t.ID]; ok {
        return rep
    }
    return t
}

func (t *TypeVariable) freeTypeVars() map[int]struct{} {
    return map[int]struct{}{t.ID: {}}
}

// TypeConstant is a named concrete type (e.g., number, string, bool, record name).
type TypeConstant struct {
    Name string
}

func (t *TypeConstant) apply(_ Subst) Type { return t }
func (t *TypeConstant) freeTypeVars() map[int]struct{} { return map[int]struct{}{} }

// TypeFunction represents a function type arg1 -> arg2 -> ... -> result.
type TypeFunction struct {
    Params []Type
    Result Type
}

func (t *TypeFunction) apply(s Subst) Type {
    ps := make([]Type, len(t.Params))
    for i, p := range t.Params {
        ps[i] = p.apply(s)
    }
    return &TypeFunction{Params: ps, Result: t.Result.apply(s)}
}

func (t *TypeFunction) freeTypeVars() map[int]struct{} {
    out := make(map[int]struct{})
    for _, p := range t.Params {
        for id := range p.freeTypeVars() {
            out[id] = struct{}{}
        }
    }
    for id := range t.Result.freeTypeVars() {
        out[id] = struct{}{}
    }
    return out
}

// TypeArray represents homogeneous arrays.
type TypeArray struct {
    Elem Type
}

func (t *TypeArray) apply(s Subst) Type { return &TypeArray{Elem: t.Elem.apply(s)} }
func (t *TypeArray) freeTypeVars() map[int]struct{} { return t.Elem.freeTypeVars() }

// TypeMap represents homogeneous map types.
type TypeMap struct {
    Key Type
    Val Type
}

func (t *TypeMap) apply(s Subst) Type { return &TypeMap{Key: t.Key.apply(s), Val: t.Val.apply(s)} }
func (t *TypeMap) freeTypeVars() map[int]struct{} {
    out := t.Key.freeTypeVars()
    for id := range t.Val.freeTypeVars() { out[id] = struct{}{} }
    return out
}


