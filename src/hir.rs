use crate::source::Span;
use crate::types::{RecordField, UnionVariant, VexType};

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
pub enum Pattern {
    Wildcard(Span),
    Binding {
        name: String,
        ty: VexType,
        span: Span,
    },
    Literal(Box<Expr>),
    Constructor {
        union_name: String,
        variant_name: String,
        bindings: Vec<Pattern>,
        span: Span,
    },
}

#[derive(Debug, Clone, PartialEq)]
pub struct MatchClause {
    pub pattern: Pattern,
    pub body: Expr,
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

    FieldAccess {
        object: Box<Expr>,
        field: String,
        span: Span,
        ty: VexType,
    },

    RecordConstructor {
        name: String,
        args: Vec<Expr>,
        span: Span,
        ty: VexType,
    },

    Match {
        scrutinee: Box<Expr>,
        clauses: Vec<MatchClause>,
        span: Span,
        ty: VexType,
    },

    VariantConstructor {
        union_name: String,
        variant_name: String,
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
            | Expr::Call { span, .. }
            | Expr::FieldAccess { span, .. }
            | Expr::RecordConstructor { span, .. }
            | Expr::Match { span, .. }
            | Expr::VariantConstructor { span, .. } => *span,
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
            | Expr::Call { ty, .. }
            | Expr::FieldAccess { ty, .. }
            | Expr::RecordConstructor { ty, .. }
            | Expr::Match { ty, .. }
            | Expr::VariantConstructor { ty, .. } => ty,
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

    Deftype {
        name: String,
        fields: Vec<RecordField>,
        span: Span,
    },

    Defunion {
        name: String,
        variants: Vec<UnionVariant>,
        span: Span,
    },

    Expr(Expr),
}

impl TopForm {
    pub fn span(&self) -> Span {
        match self {
            TopForm::Defn { span, .. }
            | TopForm::Def { span, .. }
            | TopForm::Deftype { span, .. }
            | TopForm::Defunion { span, .. } => *span,
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
    fn deftype_top_form() {
        let form = TopForm::Deftype {
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
            span: span(0, 35),
        };
        assert_eq!(form.span(), span(0, 35));
        if let TopForm::Deftype { name, fields, .. } = &form {
            assert_eq!(name, "Point");
            assert_eq!(fields.len(), 2);
            assert_eq!(fields[0].name, "x");
            assert_eq!(fields[0].ty, VexType::Float);
            assert_eq!(fields[1].name, "y");
            assert_eq!(fields[1].ty, VexType::Float);
        }
    }

    #[test]
    fn module_with_deftype() {
        let module = Module {
            top_forms: vec![
                TopForm::Deftype {
                    name: "Config".into(),
                    fields: vec![RecordField {
                        name: "name".into(),
                        ty: VexType::String,
                    }],
                    span: span(0, 30),
                },
                TopForm::Defn {
                    name: "main".into(),
                    params: vec![],
                    return_type: VexType::Unit,
                    body: vec![Expr::Nil(span(40, 43))],
                    span: span(32, 44),
                },
            ],
        };
        assert_eq!(module.top_forms.len(), 2);
        assert!(matches!(&module.top_forms[0], TopForm::Deftype { .. }));
        assert!(matches!(&module.top_forms[1], TopForm::Defn { .. }));
    }

    #[test]
    fn field_access_type() {
        let expr = Expr::FieldAccess {
            object: Box::new(Expr::Var {
                name: "point".into(),
                span: span(3, 8),
                ty: VexType::Record {
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
                },
            }),
            field: "x".into(),
            span: span(0, 11),
            ty: VexType::Float,
        };
        assert_eq!(expr.ty(), &VexType::Float);
        assert_eq!(expr.span(), span(0, 11));
    }

    #[test]
    fn record_constructor_type() {
        let point_ty = VexType::Record {
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
        let expr = Expr::RecordConstructor {
            name: "Point".into(),
            args: vec![
                Expr::Float(1.0, span(7, 10)),
                Expr::Float(2.0, span(11, 14)),
            ],
            span: span(0, 15),
            ty: point_ty.clone(),
        };
        assert_eq!(expr.ty(), &point_ty);
        assert_eq!(expr.span(), span(0, 15));
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

    #[test]
    fn defunion_top_form() {
        let form = TopForm::Defunion {
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
            span: span(0, 50),
        };
        assert!(matches!(&form, TopForm::Defunion { name, variants, .. }
            if name == "Shape" && variants.len() == 2));
        assert_eq!(form.span(), span(0, 50));
    }

    #[test]
    fn match_type() {
        let expr = Expr::Match {
            scrutinee: Box::new(Expr::Var {
                name: "x".into(),
                span: span(7, 8),
                ty: VexType::Int,
            }),
            clauses: vec![MatchClause {
                pattern: Pattern::Wildcard(span(9, 10)),
                body: Expr::Int(0, span(11, 12)),
                span: span(9, 12),
            }],
            span: span(0, 13),
            ty: VexType::Int,
        };
        assert_eq!(expr.ty(), &VexType::Int);
        assert_eq!(expr.span(), span(0, 13));
    }

    #[test]
    fn match_constructor_pattern() {
        let pattern = Pattern::Constructor {
            union_name: "Option".into(),
            variant_name: "Some".into(),
            bindings: vec![Pattern::Binding {
                name: "x".into(),
                ty: VexType::Int,
                span: span(6, 7),
            }],
            span: span(0, 8),
        };
        assert!(
            matches!(&pattern, Pattern::Constructor { variant_name, .. } if variant_name == "Some")
        );
    }
}
