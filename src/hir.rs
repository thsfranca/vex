use crate::source::Span;
use crate::types::VexType;

#[derive(Debug, Clone, PartialEq)]
pub struct Param {
    pub name: String,
    pub ty: VexType,
    pub span: Span,
}

#[derive(Debug, Clone, PartialEq)]
pub struct Binding {
    pub name: String,
    pub ty: VexType,
    pub value: Expr,
    pub span: Span,
}

#[derive(Debug, Clone, PartialEq)]
pub enum Expr {
    Int(i64, Span),
    Float(f64, Span),
    String(String, Span),
    Bool(bool, Span),
    Nil(Span),

    Var {
        name: String,
        span: Span,
        ty: VexType,
    },

    If {
        test: Box<Expr>,
        then_branch: Box<Expr>,
        else_branch: Box<Expr>,
        span: Span,
        ty: VexType,
    },

    Let {
        bindings: Vec<Binding>,
        body: Vec<Expr>,
        span: Span,
        ty: VexType,
    },

    Lambda {
        params: Vec<Param>,
        return_type: VexType,
        body: Vec<Expr>,
        span: Span,
        ty: VexType,
    },

    Call {
        func: Box<Expr>,
        args: Vec<Expr>,
        span: Span,
        ty: VexType,
    },
}

impl Expr {
    pub fn span(&self) -> Span {
        match self {
            Expr::Int(_, s)
            | Expr::Float(_, s)
            | Expr::String(_, s)
            | Expr::Bool(_, s)
            | Expr::Nil(s) => *s,
            Expr::Var { span, .. }
            | Expr::If { span, .. }
            | Expr::Let { span, .. }
            | Expr::Lambda { span, .. }
            | Expr::Call { span, .. } => *span,
        }
    }

    pub fn ty(&self) -> &VexType {
        match self {
            Expr::Int(..) => &VexType::Int,
            Expr::Float(..) => &VexType::Float,
            Expr::String(..) => &VexType::String,
            Expr::Bool(..) => &VexType::Bool,
            Expr::Nil(..) => &VexType::Unit,
            Expr::Var { ty, .. }
            | Expr::If { ty, .. }
            | Expr::Let { ty, .. }
            | Expr::Lambda { ty, .. }
            | Expr::Call { ty, .. } => ty,
        }
    }
}

#[derive(Debug, Clone, PartialEq)]
pub enum TopForm {
    Defn {
        name: String,
        params: Vec<Param>,
        return_type: VexType,
        body: Vec<Expr>,
        span: Span,
    },

    Def {
        name: String,
        ty: VexType,
        value: Expr,
        span: Span,
    },

    Expr(Expr),
}

impl TopForm {
    pub fn span(&self) -> Span {
        match self {
            TopForm::Defn { span, .. } | TopForm::Def { span, .. } => *span,
            TopForm::Expr(expr) => expr.span(),
        }
    }
}

#[derive(Debug, Clone, PartialEq)]
pub struct Module {
    pub top_forms: Vec<TopForm>,
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::source::FileId;

    fn span(start: u32, end: u32) -> Span {
        Span::new(FileId::new(0), start, end)
    }

    #[test]
    fn literal_types() {
        assert_eq!(Expr::Int(42, span(0, 2)).ty(), &VexType::Int);
        assert_eq!(Expr::Float(3.14, span(0, 4)).ty(), &VexType::Float);
        assert_eq!(Expr::String("hi".into(), span(0, 4)).ty(), &VexType::String);
        assert_eq!(Expr::Bool(true, span(0, 4)).ty(), &VexType::Bool);
        assert_eq!(Expr::Nil(span(0, 3)).ty(), &VexType::Unit);
    }

    #[test]
    fn var_type() {
        let expr = Expr::Var {
            name: "x".into(),
            span: span(0, 1),
            ty: VexType::Int,
        };
        assert_eq!(expr.ty(), &VexType::Int);
        assert_eq!(expr.span(), span(0, 1));
    }

    #[test]
    fn if_type() {
        let expr = Expr::If {
            test: Box::new(Expr::Bool(true, span(4, 8))),
            then_branch: Box::new(Expr::Int(1, span(9, 10))),
            else_branch: Box::new(Expr::Int(0, span(11, 12))),
            span: span(0, 13),
            ty: VexType::Int,
        };
        assert_eq!(expr.ty(), &VexType::Int);
        assert_eq!(expr.span(), span(0, 13));
    }

    #[test]
    fn let_type() {
        let expr = Expr::Let {
            bindings: vec![Binding {
                name: "x".into(),
                ty: VexType::Int,
                value: Expr::Int(42, span(6, 8)),
                span: span(5, 8),
            }],
            body: vec![Expr::Var {
                name: "x".into(),
                span: span(10, 11),
                ty: VexType::Int,
            }],
            span: span(0, 12),
            ty: VexType::Int,
        };
        assert_eq!(expr.ty(), &VexType::Int);
    }

    #[test]
    fn lambda_type() {
        let fn_ty = VexType::Fn {
            params: vec![VexType::Int],
            ret: Box::new(VexType::Int),
        };
        let expr = Expr::Lambda {
            params: vec![Param {
                name: "n".into(),
                ty: VexType::Int,
                span: span(5, 6),
            }],
            return_type: VexType::Int,
            body: vec![Expr::Var {
                name: "n".into(),
                span: span(8, 9),
                ty: VexType::Int,
            }],
            span: span(0, 10),
            ty: fn_ty.clone(),
        };
        assert_eq!(expr.ty(), &fn_ty);
    }

    #[test]
    fn call_type() {
        let expr = Expr::Call {
            func: Box::new(Expr::Var {
                name: "+".into(),
                span: span(1, 2),
                ty: VexType::Fn {
                    params: vec![VexType::Int, VexType::Int],
                    ret: Box::new(VexType::Int),
                },
            }),
            args: vec![Expr::Int(1, span(3, 4)), Expr::Int(2, span(5, 6))],
            span: span(0, 7),
            ty: VexType::Int,
        };
        assert_eq!(expr.ty(), &VexType::Int);
    }

    #[test]
    fn defn_top_form() {
        let form = TopForm::Defn {
            name: "main".into(),
            params: vec![],
            return_type: VexType::Unit,
            body: vec![Expr::Call {
                func: Box::new(Expr::Var {
                    name: "println".into(),
                    span: span(16, 23),
                    ty: VexType::Fn {
                        params: vec![VexType::String],
                        ret: Box::new(VexType::Unit),
                    },
                }),
                args: vec![Expr::String("Hello, World!".into(), span(24, 39))],
                span: span(15, 40),
                ty: VexType::Unit,
            }],
            span: span(0, 41),
        };
        assert_eq!(form.span(), span(0, 41));
        assert!(matches!(&form, TopForm::Defn { name, .. } if name == "main"));
    }

    #[test]
    fn def_top_form() {
        let form = TopForm::Def {
            name: "pi".into(),
            ty: VexType::Float,
            value: Expr::Float(3.14, span(15, 19)),
            span: span(0, 20),
        };
        assert_eq!(form.span(), span(0, 20));
        assert!(matches!(&form, TopForm::Def { name, .. } if name == "pi"));
    }

    #[test]
    fn module_structure() {
        let module = Module {
            top_forms: vec![
                TopForm::Def {
                    name: "x".into(),
                    ty: VexType::Int,
                    value: Expr::Int(42, span(5, 7)),
                    span: span(0, 8),
                },
                TopForm::Defn {
                    name: "main".into(),
                    params: vec![],
                    return_type: VexType::Unit,
                    body: vec![Expr::Var {
                        name: "x".into(),
                        span: span(20, 21),
                        ty: VexType::Int,
                    }],
                    span: span(10, 22),
                },
            ],
        };
        assert_eq!(module.top_forms.len(), 2);
        assert!(matches!(&module.top_forms[0], TopForm::Def { .. }));
        assert!(matches!(&module.top_forms[1], TopForm::Defn { .. }));
    }

    #[test]
    fn no_cond_in_hir() {
        let desugared = Expr::If {
            test: Box::new(Expr::Bool(true, span(0, 4))),
            then_branch: Box::new(Expr::String("yes".into(), span(5, 10))),
            else_branch: Box::new(Expr::If {
                test: Box::new(Expr::Bool(false, span(11, 16))),
                then_branch: Box::new(Expr::String("no".into(), span(17, 21))),
                else_branch: Box::new(Expr::String("maybe".into(), span(22, 29))),
                span: span(11, 30),
                ty: VexType::String,
            }),
            span: span(0, 30),
            ty: VexType::String,
        };
        assert_eq!(desugared.ty(), &VexType::String);
    }
}
