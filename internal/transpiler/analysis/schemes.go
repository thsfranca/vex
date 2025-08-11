package analysis

// TypeScheme represents a polymorphic type with quantified variables.
// ForAll vars. Type
type TypeScheme struct {
    Quantified []int
    Body       Type
}

// TypeEnv is a typing environment mapping names to schemes.
type TypeEnv map[string]*TypeScheme

// generalize produces a scheme by quantifying variables not free in the environment.
func generalize(env TypeEnv, t Type) *TypeScheme {
    freeInT := t.freeTypeVars()
    freeInEnv := make(map[int]struct{})
    for _, sch := range env {
        // free vars of a scheme are free vars of the body minus quantified vars
        bodyFree := sch.Body.freeTypeVars()
        for _, q := range sch.Quantified { delete(bodyFree, q) }
        for id := range bodyFree { freeInEnv[id] = struct{}{} }
    }
    vars := make([]int, 0)
    for id := range freeInT {
        if _, ok := freeInEnv[id]; !ok {
            vars = append(vars, id)
        }
    }
    return &TypeScheme{Quantified: vars, Body: t}
}

// instantiate replaces quantified variables by fresh type variables.
func instantiate(sch *TypeScheme, fresh func() int) Type {
    subst := Subst{}
    for _, id := range sch.Quantified {
        subst[id] = &TypeVariable{ID: fresh()}
    }
    return sch.Body.apply(subst)
}


