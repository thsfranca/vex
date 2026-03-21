use std::collections::HashMap;
use std::fmt;

#[derive(Debug, Clone, PartialEq, Eq)]
pub struct RecordField {
    pub name: std::string::String,
    pub ty: VexType,
}

#[derive(Debug, Clone, PartialEq, Eq)]
pub struct UnionVariant {
    pub name: std::string::String,
    pub types: Vec<VexType>,
}

#[derive(Debug, Clone, PartialEq, Eq)]
pub enum VexType {
    Int,
    Float,
    Bool,
    String,
    Unit,
    Fn {
        params: Vec<VexType>,
        ret: Box<VexType>,
    },
    Record {
        name: std::string::String,
        fields: Vec<RecordField>,
    },
    Union {
        name: std::string::String,
        variants: Vec<UnionVariant>,
    },
    List(Box<VexType>),
    Map {
        key: Box<VexType>,
        value: Box<VexType>,
    },
    Channel(Box<VexType>),
    Option(Box<VexType>),
    Result {
        ok: Box<VexType>,
        err: Box<VexType>,
    },
    TypeVar(u32),
}

impl VexType {
    pub fn is_numeric(&self) -> bool {
        matches!(self, VexType::Int | VexType::Float)
    }

    pub fn types_compatible(a: &VexType, b: &VexType) -> Option<VexType> {
        if a == b {
            return Some(a.clone());
        }
        match (a, b) {
            (VexType::TypeVar(_), other) | (other, VexType::TypeVar(_)) => Some(other.clone()),
            (VexType::List(inner_a), VexType::List(inner_b)) => {
                let merged = VexType::types_compatible(inner_a, inner_b)?;
                Some(VexType::List(Box::new(merged)))
            }
            (VexType::Map { key: ka, value: va }, VexType::Map { key: kb, value: vb }) => {
                let key = VexType::types_compatible(ka, kb)?;
                let value = VexType::types_compatible(va, vb)?;
                Some(VexType::Map {
                    key: Box::new(key),
                    value: Box::new(value),
                })
            }
            (VexType::Channel(inner_a), VexType::Channel(inner_b)) => {
                let merged = VexType::types_compatible(inner_a, inner_b)?;
                Some(VexType::Channel(Box::new(merged)))
            }
            (VexType::Option(inner_a), VexType::Option(inner_b)) => {
                let merged = VexType::types_compatible(inner_a, inner_b)?;
                Some(VexType::Option(Box::new(merged)))
            }
            (
                VexType::Result {
                    ok: ok_a,
                    err: err_a,
                },
                VexType::Result {
                    ok: ok_b,
                    err: err_b,
                },
            ) => {
                let ok = VexType::types_compatible(ok_a, ok_b)?;
                let err = VexType::types_compatible(err_a, err_b)?;
                Some(VexType::Result {
                    ok: Box::new(ok),
                    err: Box::new(err),
                })
            }
            _ => None,
        }
    }

    pub fn resolve_vars(&self, target: &VexType) -> VexType {
        match (self, target) {
            (VexType::TypeVar(_), concrete) => concrete.clone(),
            (VexType::List(a), VexType::List(b)) => VexType::List(Box::new(a.resolve_vars(b))),
            (VexType::Map { key: ka, value: va }, VexType::Map { key: kb, value: vb }) => {
                VexType::Map {
                    key: Box::new(ka.resolve_vars(kb)),
                    value: Box::new(va.resolve_vars(vb)),
                }
            }
            (VexType::Channel(a), VexType::Channel(b)) => {
                VexType::Channel(Box::new(a.resolve_vars(b)))
            }
            (VexType::Option(a), VexType::Option(b)) => {
                VexType::Option(Box::new(a.resolve_vars(b)))
            }
            (
                VexType::Result {
                    ok: oa, err: ea, ..
                },
                VexType::Result {
                    ok: ob, err: eb, ..
                },
            ) => VexType::Result {
                ok: Box::new(oa.resolve_vars(ob)),
                err: Box::new(ea.resolve_vars(eb)),
            },
            _ => self.clone(),
        }
    }

    pub fn field_type(&self, field_name: &str) -> Option<&VexType> {
        match self {
            VexType::Record { fields, .. } => {
                fields.iter().find(|f| f.name == field_name).map(|f| &f.ty)
            }
            _ => None,
        }
    }
}

impl fmt::Display for VexType {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        match self {
            VexType::Int => write!(f, "Int"),
            VexType::Float => write!(f, "Float"),
            VexType::Bool => write!(f, "Bool"),
            VexType::String => write!(f, "String"),
            VexType::Unit => write!(f, "Unit"),
            VexType::Fn { params, ret } => {
                write!(f, "(Fn [")?;
                for (i, p) in params.iter().enumerate() {
                    if i > 0 {
                        write!(f, " ")?;
                    }
                    write!(f, "{}", p)?;
                }
                write!(f, "] {})", ret)
            }
            VexType::Record { name, .. } => write!(f, "{}", name),
            VexType::Union { name, .. } => write!(f, "{}", name),
            VexType::List(inner) => write!(f, "(List {})", inner),
            VexType::Map { key, value } => write!(f, "(Map {} {})", key, value),
            VexType::Channel(inner) => write!(f, "(Channel {})", inner),
            VexType::Option(inner) => write!(f, "(Option {})", inner),
            VexType::Result { ok, err } => write!(f, "(Result {} {})", ok, err),
            VexType::TypeVar(id) => write!(f, "?T{}", id),
        }
    }
}

pub struct TypeEnv {
    scopes: Vec<HashMap<std::string::String, VexType>>,
    next_type_var: u32,
}

impl Default for TypeEnv {
    fn default() -> Self {
        Self::new()
    }
}

impl TypeEnv {
    pub fn new() -> Self {
        Self {
            scopes: vec![HashMap::new()],
            next_type_var: 0,
        }
    }

    pub fn push_scope(&mut self) {
        self.scopes.push(HashMap::new());
    }

    pub fn pop_scope(&mut self) {
        self.scopes.pop();
    }

    pub fn define(&mut self, name: std::string::String, ty: VexType) {
        if let Some(scope) = self.scopes.last_mut() {
            scope.insert(name, ty);
        }
    }

    pub fn lookup(&self, name: &str) -> Option<&VexType> {
        for scope in self.scopes.iter().rev() {
            if let Some(ty) = scope.get(name) {
                return Some(ty);
            }
        }
        None
    }

    pub fn fresh_type_var(&mut self) -> VexType {
        let id = self.next_type_var;
        self.next_type_var += 1;
        VexType::TypeVar(id)
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn display_primitives() {
        assert_eq!(VexType::Int.to_string(), "Int");
        assert_eq!(VexType::Float.to_string(), "Float");
        assert_eq!(VexType::Bool.to_string(), "Bool");
        assert_eq!(VexType::String.to_string(), "String");
        assert_eq!(VexType::Unit.to_string(), "Unit");
    }

    #[test]
    fn display_function_type() {
        let ty = VexType::Fn {
            params: vec![VexType::Int, VexType::Int],
            ret: Box::new(VexType::Int),
        };
        assert_eq!(ty.to_string(), "(Fn [Int Int] Int)");
    }

    #[test]
    fn display_function_no_params() {
        let ty = VexType::Fn {
            params: vec![],
            ret: Box::new(VexType::Unit),
        };
        assert_eq!(ty.to_string(), "(Fn [] Unit)");
    }

    #[test]
    fn display_higher_order_function() {
        let callback = VexType::Fn {
            params: vec![VexType::Int],
            ret: Box::new(VexType::Bool),
        };
        let ty = VexType::Fn {
            params: vec![callback],
            ret: Box::new(VexType::Unit),
        };
        assert_eq!(ty.to_string(), "(Fn [(Fn [Int] Bool)] Unit)");
    }

    #[test]
    fn display_type_var() {
        assert_eq!(VexType::TypeVar(0).to_string(), "?T0");
        assert_eq!(VexType::TypeVar(42).to_string(), "?T42");
    }

    #[test]
    fn is_numeric() {
        assert!(VexType::Int.is_numeric());
        assert!(VexType::Float.is_numeric());
        assert!(!VexType::Bool.is_numeric());
        assert!(!VexType::String.is_numeric());
        assert!(!VexType::Unit.is_numeric());
    }

    #[test]
    fn type_equality() {
        assert_eq!(VexType::Int, VexType::Int);
        assert_ne!(VexType::Int, VexType::Float);

        let fn1 = VexType::Fn {
            params: vec![VexType::Int],
            ret: Box::new(VexType::Bool),
        };
        let fn2 = VexType::Fn {
            params: vec![VexType::Int],
            ret: Box::new(VexType::Bool),
        };
        let fn3 = VexType::Fn {
            params: vec![VexType::Float],
            ret: Box::new(VexType::Bool),
        };
        assert_eq!(fn1, fn2);
        assert_ne!(fn1, fn3);
    }

    #[test]
    fn display_record() {
        let ty = VexType::Record {
            name: "Point".into(),
            fields: vec![
                RecordField {
                    name: "x".into(),
                    ty: VexType::Float,
                },
                RecordField {
                    name: "y".into(),
                    ty: VexType::Float,
                },
            ],
        };
        assert_eq!(ty.to_string(), "Point");
    }

    #[test]
    fn record_field_lookup() {
        let ty = VexType::Record {
            name: "Point".into(),
            fields: vec![
                RecordField {
                    name: "x".into(),
                    ty: VexType::Float,
                },
                RecordField {
                    name: "y".into(),
                    ty: VexType::Int,
                },
            ],
        };
        assert_eq!(ty.field_type("x"), Some(&VexType::Float));
        assert_eq!(ty.field_type("y"), Some(&VexType::Int));
        assert_eq!(ty.field_type("z"), None);
    }

    #[test]
    fn record_field_lookup_non_record() {
        assert_eq!(VexType::Int.field_type("x"), None);
        assert_eq!(VexType::String.field_type("x"), None);
    }

    #[test]
    fn record_equality() {
        let r1 = VexType::Record {
            name: "Point".into(),
            fields: vec![RecordField {
                name: "x".into(),
                ty: VexType::Int,
            }],
        };
        let r2 = VexType::Record {
            name: "Point".into(),
            fields: vec![RecordField {
                name: "x".into(),
                ty: VexType::Int,
            }],
        };
        let r3 = VexType::Record {
            name: "Other".into(),
            fields: vec![RecordField {
                name: "x".into(),
                ty: VexType::Int,
            }],
        };
        assert_eq!(r1, r2);
        assert_ne!(r1, r3);
    }

    #[test]
    fn env_define_and_lookup() {
        let mut env = TypeEnv::new();
        env.define("x".into(), VexType::Int);
        assert_eq!(env.lookup("x"), Some(&VexType::Int));
        assert_eq!(env.lookup("y"), None);
    }

    #[test]
    fn env_scoping() {
        let mut env = TypeEnv::new();
        env.define("x".into(), VexType::Int);

        env.push_scope();
        env.define("x".into(), VexType::Float);
        assert_eq!(env.lookup("x"), Some(&VexType::Float));

        env.pop_scope();
        assert_eq!(env.lookup("x"), Some(&VexType::Int));
    }

    #[test]
    fn env_inner_scope_sees_outer() {
        let mut env = TypeEnv::new();
        env.define("x".into(), VexType::Int);

        env.push_scope();
        env.define("y".into(), VexType::Bool);
        assert_eq!(env.lookup("x"), Some(&VexType::Int));
        assert_eq!(env.lookup("y"), Some(&VexType::Bool));

        env.pop_scope();
        assert_eq!(env.lookup("y"), None);
    }

    #[test]
    fn env_fresh_type_vars() {
        let mut env = TypeEnv::new();
        assert_eq!(env.fresh_type_var(), VexType::TypeVar(0));
        assert_eq!(env.fresh_type_var(), VexType::TypeVar(1));
        assert_eq!(env.fresh_type_var(), VexType::TypeVar(2));
    }

    #[test]
    fn display_union() {
        let ty = VexType::Union {
            name: "Shape".into(),
            variants: vec![
                UnionVariant {
                    name: "Circle".into(),
                    types: vec![VexType::Float],
                },
                UnionVariant {
                    name: "Rect".into(),
                    types: vec![VexType::Float, VexType::Float],
                },
            ],
        };
        assert_eq!(format!("{}", ty), "Shape");
    }

    #[test]
    fn display_option() {
        let ty = VexType::Option(Box::new(VexType::Int));
        assert_eq!(format!("{}", ty), "(Option Int)");
    }

    #[test]
    fn display_option_nested() {
        let ty = VexType::Option(Box::new(VexType::Option(Box::new(VexType::String))));
        assert_eq!(format!("{}", ty), "(Option (Option String))");
    }

    #[test]
    fn display_result() {
        let ty = VexType::Result {
            ok: Box::new(VexType::Int),
            err: Box::new(VexType::String),
        };
        assert_eq!(format!("{}", ty), "(Result Int String)");
    }

    #[test]
    fn option_equality() {
        let a = VexType::Option(Box::new(VexType::Int));
        let b = VexType::Option(Box::new(VexType::Int));
        let c = VexType::Option(Box::new(VexType::String));
        assert_eq!(a, b);
        assert_ne!(a, c);
    }

    #[test]
    fn result_equality() {
        let a = VexType::Result {
            ok: Box::new(VexType::Int),
            err: Box::new(VexType::String),
        };
        let b = VexType::Result {
            ok: Box::new(VexType::Int),
            err: Box::new(VexType::String),
        };
        let c = VexType::Result {
            ok: Box::new(VexType::Float),
            err: Box::new(VexType::String),
        };
        assert_eq!(a, b);
        assert_ne!(a, c);
    }

    #[test]
    fn types_compatible_equal() {
        assert_eq!(
            VexType::types_compatible(&VexType::Int, &VexType::Int),
            Some(VexType::Int)
        );
    }

    #[test]
    fn types_compatible_mismatch() {
        assert_eq!(
            VexType::types_compatible(&VexType::Int, &VexType::String),
            None
        );
    }

    #[test]
    fn types_compatible_typevar_resolves() {
        assert_eq!(
            VexType::types_compatible(&VexType::TypeVar(0), &VexType::Int),
            Some(VexType::Int)
        );
        assert_eq!(
            VexType::types_compatible(&VexType::Int, &VexType::TypeVar(0)),
            Some(VexType::Int)
        );
    }

    #[test]
    fn types_compatible_option_with_typevar() {
        let a = VexType::Option(Box::new(VexType::TypeVar(0)));
        let b = VexType::Option(Box::new(VexType::Int));
        assert_eq!(
            VexType::types_compatible(&a, &b),
            Some(VexType::Option(Box::new(VexType::Int)))
        );
    }

    #[test]
    fn types_compatible_result_with_typevar() {
        let a = VexType::Result {
            ok: Box::new(VexType::Int),
            err: Box::new(VexType::TypeVar(0)),
        };
        let b = VexType::Result {
            ok: Box::new(VexType::Int),
            err: Box::new(VexType::String),
        };
        assert_eq!(
            VexType::types_compatible(&a, &b),
            Some(VexType::Result {
                ok: Box::new(VexType::Int),
                err: Box::new(VexType::String),
            })
        );
    }

    #[test]
    fn display_list() {
        let ty = VexType::List(Box::new(VexType::Int));
        assert_eq!(format!("{}", ty), "(List Int)");
    }

    #[test]
    fn display_list_nested() {
        let ty = VexType::List(Box::new(VexType::List(Box::new(VexType::String))));
        assert_eq!(format!("{}", ty), "(List (List String))");
    }

    #[test]
    fn display_map() {
        let ty = VexType::Map {
            key: Box::new(VexType::String),
            value: Box::new(VexType::Int),
        };
        assert_eq!(format!("{}", ty), "(Map String Int)");
    }

    #[test]
    fn list_equality() {
        let a = VexType::List(Box::new(VexType::Int));
        let b = VexType::List(Box::new(VexType::Int));
        let c = VexType::List(Box::new(VexType::String));
        assert_eq!(a, b);
        assert_ne!(a, c);
    }

    #[test]
    fn map_equality() {
        let a = VexType::Map {
            key: Box::new(VexType::String),
            value: Box::new(VexType::Int),
        };
        let b = VexType::Map {
            key: Box::new(VexType::String),
            value: Box::new(VexType::Int),
        };
        let c = VexType::Map {
            key: Box::new(VexType::Int),
            value: Box::new(VexType::Int),
        };
        assert_eq!(a, b);
        assert_ne!(a, c);
    }

    #[test]
    fn types_compatible_list() {
        let a = VexType::List(Box::new(VexType::TypeVar(0)));
        let b = VexType::List(Box::new(VexType::Int));
        assert_eq!(
            VexType::types_compatible(&a, &b),
            Some(VexType::List(Box::new(VexType::Int)))
        );
    }

    #[test]
    fn types_compatible_map() {
        let a = VexType::Map {
            key: Box::new(VexType::String),
            value: Box::new(VexType::TypeVar(0)),
        };
        let b = VexType::Map {
            key: Box::new(VexType::String),
            value: Box::new(VexType::Int),
        };
        assert_eq!(
            VexType::types_compatible(&a, &b),
            Some(VexType::Map {
                key: Box::new(VexType::String),
                value: Box::new(VexType::Int),
            })
        );
    }

    #[test]
    fn union_equality() {
        let a = VexType::Union {
            name: "Opt".into(),
            variants: vec![
                UnionVariant {
                    name: "Some".into(),
                    types: vec![VexType::Int],
                },
                UnionVariant {
                    name: "None".into(),
                    types: vec![],
                },
            ],
        };
        let b = a.clone();
        assert_eq!(a, b);
    }

    #[test]
    fn display_channel() {
        let ty = VexType::Channel(Box::new(VexType::Int));
        assert_eq!(format!("{}", ty), "(Channel Int)");
    }

    #[test]
    fn channel_equality() {
        let a = VexType::Channel(Box::new(VexType::Int));
        let b = VexType::Channel(Box::new(VexType::Int));
        let c = VexType::Channel(Box::new(VexType::String));
        assert_eq!(a, b);
        assert_ne!(a, c);
    }

    #[test]
    fn types_compatible_channel() {
        let a = VexType::Channel(Box::new(VexType::TypeVar(0)));
        let b = VexType::Channel(Box::new(VexType::Int));
        assert_eq!(
            VexType::types_compatible(&a, &b),
            Some(VexType::Channel(Box::new(VexType::Int)))
        );
    }
}
