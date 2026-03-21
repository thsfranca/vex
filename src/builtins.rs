use crate::types::VexType;

#[derive(Debug, Clone, PartialEq)]
pub enum GoTranslation {
    Infix(&'static str),
    Prefix(&'static str),
    FuncCall {
        go_name: &'static str,
        go_import: &'static str,
    },
}

#[derive(Debug, Clone, PartialEq)]
pub struct Builtin {
    pub name: &'static str,
    pub ty: VexType,
    pub go: GoTranslation,
    pub variadic: bool,
}

fn int_bin_op(name: &'static str, ret: VexType, go_op: &'static str) -> Builtin {
    Builtin {
        name,
        ty: VexType::Fn {
            params: vec![VexType::Int, VexType::Int],
            ret: Box::new(ret),
        },
        go: GoTranslation::Infix(go_op),
        variadic: false,
    }
}

fn bool_bin_op(name: &'static str, go_op: &'static str) -> Builtin {
    Builtin {
        name,
        ty: VexType::Fn {
            params: vec![VexType::Bool, VexType::Bool],
            ret: Box::new(VexType::Bool),
        },
        go: GoTranslation::Infix(go_op),
        variadic: false,
    }
}

pub fn all_builtins() -> Vec<Builtin> {
    vec![
        int_bin_op("+", VexType::Int, "+"),
        int_bin_op("-", VexType::Int, "-"),
        int_bin_op("*", VexType::Int, "*"),
        int_bin_op("/", VexType::Int, "/"),
        int_bin_op("mod", VexType::Int, "%"),
        int_bin_op("<", VexType::Bool, "<"),
        int_bin_op(">", VexType::Bool, ">"),
        int_bin_op("<=", VexType::Bool, "<="),
        int_bin_op(">=", VexType::Bool, ">="),
        int_bin_op("=", VexType::Bool, "=="),
        int_bin_op("!=", VexType::Bool, "!="),
        bool_bin_op("&&", "&&"),
        bool_bin_op("||", "||"),
        Builtin {
            name: "not",
            ty: VexType::Fn {
                params: vec![VexType::Bool],
                ret: Box::new(VexType::Bool),
            },
            go: GoTranslation::Prefix("!"),
            variadic: false,
        },
        Builtin {
            name: "println",
            ty: VexType::Fn {
                params: vec![VexType::String],
                ret: Box::new(VexType::Unit),
            },
            go: GoTranslation::FuncCall {
                go_name: "fmt.Println",
                go_import: "fmt",
            },
            variadic: false,
        },
        Builtin {
            name: "str",
            ty: VexType::Fn {
                params: vec![],
                ret: Box::new(VexType::String),
            },
            go: GoTranslation::FuncCall {
                go_name: "fmt.Sprint",
                go_import: "fmt",
            },
            variadic: true,
        },
        Builtin {
            name: "range",
            ty: VexType::Fn {
                params: vec![VexType::Int, VexType::Int],
                ret: Box::new(VexType::List(Box::new(VexType::Int))),
            },
            go: GoTranslation::FuncCall {
                go_name: "vexrt.Range",
                go_import: "vex_out/vexrt",
            },
            variadic: false,
        },
    ]
}

pub fn lookup(name: &str) -> Option<&'static Builtin> {
    use std::sync::LazyLock;

    static BUILTINS: LazyLock<Vec<Builtin>> = LazyLock::new(all_builtins);

    BUILTINS.iter().find(|b| b.name == name)
}

pub fn is_builtin(name: &str) -> bool {
    lookup(name).is_some()
}

pub fn go_imports(names: &[&str]) -> Vec<&'static str> {
    use std::collections::BTreeSet;

    let mut imports = BTreeSet::new();
    for name in names {
        if let Some(builtin) = lookup(name)
            && let GoTranslation::FuncCall { go_import, .. } = &builtin.go
        {
            imports.insert(*go_import);
        }
    }
    imports.into_iter().collect()
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn all_builtins_count() {
        let builtins = all_builtins();
        assert_eq!(builtins.len(), 17);
    }

    #[test]
    fn lookup_arithmetic() {
        let add = lookup("+").unwrap();
        assert_eq!(add.name, "+");
        assert_eq!(
            add.ty,
            VexType::Fn {
                params: vec![VexType::Int, VexType::Int],
                ret: Box::new(VexType::Int),
            }
        );
        assert_eq!(add.go, GoTranslation::Infix("+"));
        assert!(!add.variadic);
    }

    #[test]
    fn lookup_comparison() {
        let le = lookup("<=").unwrap();
        assert_eq!(
            le.ty,
            VexType::Fn {
                params: vec![VexType::Int, VexType::Int],
                ret: Box::new(VexType::Bool),
            }
        );
        assert_eq!(le.go, GoTranslation::Infix("<="));
    }

    #[test]
    fn lookup_mod() {
        let m = lookup("mod").unwrap();
        assert_eq!(m.go, GoTranslation::Infix("%"));
    }

    #[test]
    fn lookup_logical() {
        let and = lookup("&&").unwrap();
        assert_eq!(
            and.ty,
            VexType::Fn {
                params: vec![VexType::Bool, VexType::Bool],
                ret: Box::new(VexType::Bool),
            }
        );

        let not = lookup("not").unwrap();
        assert_eq!(not.go, GoTranslation::Prefix("!"));
        assert_eq!(
            not.ty,
            VexType::Fn {
                params: vec![VexType::Bool],
                ret: Box::new(VexType::Bool),
            }
        );
    }

    #[test]
    fn lookup_println() {
        let p = lookup("println").unwrap();
        assert_eq!(
            p.ty,
            VexType::Fn {
                params: vec![VexType::String],
                ret: Box::new(VexType::Unit),
            }
        );
        assert_eq!(
            p.go,
            GoTranslation::FuncCall {
                go_name: "fmt.Println",
                go_import: "fmt",
            }
        );
        assert!(!p.variadic);
    }

    #[test]
    fn lookup_str() {
        let s = lookup("str").unwrap();
        assert!(s.variadic);
        assert_eq!(
            s.go,
            GoTranslation::FuncCall {
                go_name: "fmt.Sprint",
                go_import: "fmt",
            }
        );
    }

    #[test]
    fn lookup_range() {
        let r = lookup("range").unwrap();
        assert_eq!(r.name, "range");
        assert_eq!(
            r.ty,
            VexType::Fn {
                params: vec![VexType::Int, VexType::Int],
                ret: Box::new(VexType::List(Box::new(VexType::Int))),
            }
        );
        assert_eq!(
            r.go,
            GoTranslation::FuncCall {
                go_name: "vexrt.Range",
                go_import: "vex_out/vexrt",
            }
        );
    }

    #[test]
    fn lookup_nonexistent() {
        assert!(lookup("nonexistent").is_none());
    }

    #[test]
    fn is_builtin_check() {
        assert!(is_builtin("+"));
        assert!(is_builtin("println"));
        assert!(!is_builtin("my_func"));
    }

    #[test]
    fn go_imports_deduplicates() {
        let imports = go_imports(&["println", "str"]);
        assert_eq!(imports, vec!["fmt"]);
    }

    #[test]
    fn go_imports_skips_operators() {
        let imports = go_imports(&["+", "-", "println"]);
        assert_eq!(imports, vec!["fmt"]);
    }

    #[test]
    fn go_imports_empty() {
        let imports = go_imports(&["+", "<=", "not"]);
        assert!(imports.is_empty());
    }
}
