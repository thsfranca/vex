use std::collections::HashMap;
use std::fmt;

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
    TypeVar(u32),
}

impl VexType {
    pub fn is_numeric(&self) -> bool {
        matches!(self, VexType::Int | VexType::Float)
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
}
