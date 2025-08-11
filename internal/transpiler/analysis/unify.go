package analysis

import "fmt"

// unify performs type unification with occur-check.
func unify(a, b Type) (Subst, error) {
    switch ta := a.(type) {
    case *TypeVariable:
        return bindVar(ta, b)
    case *TypeConstant:
        switch tb := b.(type) {
        case *TypeVariable:
            return bindVar(tb, a)
        case *TypeConstant:
            if ta.Name == tb.Name {
                return Subst{}, nil
            }
            // Numeric family: allow int/float to unify with number
            if (ta.Name == "number" && (tb.Name == "int" || tb.Name == "float")) ||
               (tb.Name == "number" && (ta.Name == "int" || ta.Name == "float")) {
                return Subst{}, nil
            }
            return nil, fmt.Errorf("cannot unify %v with %v", ta.Name, tb.Name)
        case *TypeFunction, *TypeArray:
            return nil, fmt.Errorf("cannot unify %T with %T", a, b)
        }
    case *TypeFunction:
        switch tb := b.(type) {
        case *TypeVariable:
            return bindVar(tb, a)
        case *TypeFunction:
            if len(ta.Params) != len(tb.Params) {
                return nil, fmt.Errorf("function arity mismatch")
            }
            s := Subst{}
            for i := range ta.Params {
                si, err := unify(ta.Params[i].apply(s), tb.Params[i].apply(s))
                if err != nil {
                    return nil, err
                }
                s = s.compose(si)
            }
            sr, err := unify(ta.Result.apply(s), tb.Result.apply(s))
            if err != nil {
                return nil, err
            }
            return s.compose(sr), nil
        case *TypeConstant, *TypeArray:
            return nil, fmt.Errorf("cannot unify %T with %T", a, b)
        }
    case *TypeArray:
        switch tb := b.(type) {
        case *TypeVariable:
            return bindVar(tb, a)
        case *TypeArray:
            return unify(ta.Elem, tb.Elem)
        default:
            return nil, fmt.Errorf("cannot unify %T with %T", a, b)
        }
    case *TypeMap:
        switch tb := b.(type) {
        case *TypeVariable:
            return bindVar(tb, a)
        case *TypeMap:
            s1, err := unify(ta.Key, tb.Key)
            if err != nil { return nil, err }
            s2, err := unify(ta.Val.apply(s1), tb.Val.apply(s1))
            if err != nil { return nil, err }
            return s1.compose(s2), nil
        default:
            return nil, fmt.Errorf("cannot unify %T with %T", a, b)
        }
    }
    // Fallback: try symmetric
    if _, ok := b.(*TypeVariable); ok {
        return unify(b, a)
    }
    return nil, fmt.Errorf("cannot unify %T with %T", a, b)
}

func bindVar(v *TypeVariable, t Type) (Subst, error) {
    if tv, ok := t.(*TypeVariable); ok && tv.ID == v.ID {
        return Subst{}, nil
    }
    if occurs(v.ID, t) {
        return nil, fmt.Errorf("occur-check failed: %v in %T", v.ID, t)
    }
    return Subst{v.ID: t}, nil
}

func occurs(id int, t Type) bool {
    _, ok := t.freeTypeVars()[id]
    return ok
}


