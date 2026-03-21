use crate::source::Span;

#[derive(Debug, Clone, PartialEq)]
pub enum TypeExpr {
    Named {
        name: String,
        span: Span,
    },
    Function {
        params: Vec<TypeExpr>,
        ret: Box<TypeExpr>,
        span: Span,
    },
    Applied {
        name: String,
        args: Vec<TypeExpr>,
        span: Span,
    },
}

impl TypeExpr {
    pub fn span(&self) -> Span {
        match self {
            TypeExpr::Named { span, .. }
            | TypeExpr::Function { span, .. }
            | TypeExpr::Applied { span, .. } => *span,
        }
    }
}

#[derive(Debug, Clone, PartialEq)]
pub struct Param {
    pub name: String,
    pub type_ann: Option<TypeExpr>,
    pub span: Span,
}

#[derive(Debug, Clone, PartialEq)]
pub struct Binding {
    pub name: String,
    pub value: Expr,
    pub span: Span,
}

#[derive(Debug, Clone, PartialEq)]
pub struct CondClause {
    pub test: Expr,
    pub value: Expr,
    pub span: Span,
}

#[derive(Debug, Clone, PartialEq)]
pub enum Pattern {
    Wildcard(Span),
    Binding(String, Span),
    Literal(Box<Expr>),
    Constructor {
        name: String,
        args: Vec<Pattern>,
        span: Span,
    },
}

impl Pattern {
    pub fn span(&self) -> Span {
        match self {
            Pattern::Wildcard(s) | Pattern::Binding(_, s) => *s,
            Pattern::Literal(expr) => expr.span(),
            Pattern::Constructor { span, .. } => *span,
        }
    }
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

    Symbol(String, Span),
    Keyword(String, Span),

    If {
        test: Box<Expr>,
        then_branch: Box<Expr>,
        else_branch: Box<Expr>,
        span: Span,
    },

    Cond {
        clauses: Vec<CondClause>,
        else_body: Option<Box<Expr>>,
        span: Span,
    },

    Let {
        bindings: Vec<Binding>,
        body: Vec<Expr>,
        span: Span,
    },

    Lambda {
        params: Vec<Param>,
        return_type: Option<TypeExpr>,
        body: Vec<Expr>,
        span: Span,
    },

    Call {
        func: Box<Expr>,
        args: Vec<Expr>,
        span: Span,
    },

    FieldAccess {
        object: Box<Expr>,
        field: String,
        span: Span,
    },

    Match {
        scrutinee: Box<Expr>,
        clauses: Vec<MatchClause>,
        span: Span,
    },

    Spawn {
        body: Box<Expr>,
        span: Span,
    },

    Channel {
        element_type: TypeExpr,
        size: Option<Box<Expr>>,
        span: Span,
    },

    Send {
        channel: Box<Expr>,
        value: Box<Expr>,
        span: Span,
    },

    Recv {
        channel: Box<Expr>,
        span: Span,
    },

    Quote {
        expr: Box<Expr>,
        span: Span,
    },

    Unquote {
        expr: Box<Expr>,
        span: Span,
    },

    Splice {
        expr: Box<Expr>,
        span: Span,
    },
}

impl Expr {
    pub fn span(&self) -> Span {
        match self {
            Expr::Int(_, s)
            | Expr::Float(_, s)
            | Expr::String(_, s)
            | Expr::Bool(_, s)
            | Expr::Nil(s)
            | Expr::Symbol(_, s)
            | Expr::Keyword(_, s) => *s,
            Expr::If { span, .. }
            | Expr::Cond { span, .. }
            | Expr::Let { span, .. }
            | Expr::Lambda { span, .. }
            | Expr::Call { span, .. }
            | Expr::FieldAccess { span, .. }
            | Expr::Match { span, .. }
            | Expr::Spawn { span, .. }
            | Expr::Channel { span, .. }
            | Expr::Send { span, .. }
            | Expr::Recv { span, .. }
            | Expr::Quote { span, .. }
            | Expr::Unquote { span, .. }
            | Expr::Splice { span, .. } => *span,
        }
    }
}

#[derive(Debug, Clone, PartialEq)]
pub struct Field {
    pub name: String,
    pub type_expr: TypeExpr,
    pub span: Span,
}

#[derive(Debug, Clone, PartialEq)]
pub enum TopForm {
    Module {
        name: String,
        span: Span,
    },

    Export {
        symbols: Vec<String>,
        span: Span,
    },

    Import {
        module_path: String,
        symbols: Vec<String>,
        span: Span,
    },

    ImportGo {
        go_package: String,
        symbols: Vec<String>,
        span: Span,
    },

    Defn {
        name: String,
        params: Vec<Param>,
        return_type: Option<TypeExpr>,
        body: Vec<Expr>,
        span: Span,
    },

    Def {
        name: String,
        type_ann: Option<TypeExpr>,
        value: Expr,
        span: Span,
    },

    Deftype {
        name: String,
        fields: Vec<Field>,
        span: Span,
    },

    Defunion {
        name: String,
        variants: Vec<Variant>,
        span: Span,
    },

    DefMacro {
        name: String,
        params: Vec<Param>,
        rest_param: Option<String>,
        body: Vec<Expr>,
        span: Span,
    },

    Expr(Expr),
}

impl TopForm {
    pub fn span(&self) -> Span {
        match self {
            TopForm::Module { span, .. }
            | TopForm::Export { span, .. }
            | TopForm::Import { span, .. }
            | TopForm::ImportGo { span, .. }
            | TopForm::Defn { span, .. }
            | TopForm::Def { span, .. }
            | TopForm::Deftype { span, .. }
            | TopForm::Defunion { span, .. }
            | TopForm::DefMacro { span, .. } => *span,
            TopForm::Expr(expr) => expr.span(),
        }
    }
}

#[derive(Debug, Clone, PartialEq)]
pub struct Variant {
    pub name: String,
    pub types: Vec<TypeExpr>,
    pub span: Span,
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::source::FileId;

    fn span(start: u32, end: u32) -> Span {
        Span::new(FileId::new(0), start, end)
    }

    #[test]
    fn literal_spans() {
        assert_eq!(Expr::Int(42, span(0, 2)).span(), span(0, 2));
        assert_eq!(Expr::Float(3.14, span(0, 4)).span(), span(0, 4));
        assert_eq!(Expr::String("hi".into(), span(0, 4)).span(), span(0, 4));
        assert_eq!(Expr::Bool(true, span(0, 4)).span(), span(0, 4));
        assert_eq!(Expr::Nil(span(0, 3)).span(), span(0, 3));
    }

    #[test]
    fn symbol_and_keyword_spans() {
        assert_eq!(Expr::Symbol("x".into(), span(0, 1)).span(), span(0, 1));
        assert_eq!(Expr::Keyword("name".into(), span(0, 5)).span(), span(0, 5));
    }

    #[test]
    fn if_expr() {
        let expr = Expr::If {
            test: Box::new(Expr::Bool(true, span(4, 8))),
            then_branch: Box::new(Expr::Int(1, span(9, 10))),
            else_branch: Box::new(Expr::Int(0, span(11, 12))),
            span: span(0, 13),
        };
        assert_eq!(expr.span(), span(0, 13));
        assert!(matches!(expr, Expr::If { .. }));
    }

    #[test]
    fn cond_expr() {
        let expr = Expr::Cond {
            clauses: vec![CondClause {
                test: Expr::Bool(true, span(6, 10)),
                value: Expr::Int(1, span(11, 12)),
                span: span(6, 12),
            }],
            else_body: Some(Box::new(Expr::Int(0, span(19, 20)))),
            span: span(0, 21),
        };
        assert_eq!(expr.span(), span(0, 21));
        if let Expr::Cond {
            clauses, else_body, ..
        } = &expr
        {
            assert_eq!(clauses.len(), 1);
            assert!(else_body.is_some());
        }
    }

    #[test]
    fn let_expr() {
        let expr = Expr::Let {
            bindings: vec![Binding {
                name: "x".into(),
                value: Expr::Int(42, span(6, 8)),
                span: span(5, 8),
            }],
            body: vec![Expr::Symbol("x".into(), span(10, 11))],
            span: span(0, 12),
        };
        assert_eq!(expr.span(), span(0, 12));
        if let Expr::Let { bindings, body, .. } = &expr {
            assert_eq!(bindings.len(), 1);
            assert_eq!(bindings[0].name, "x");
            assert_eq!(body.len(), 1);
        }
    }

    #[test]
    fn lambda_expr() {
        let expr = Expr::Lambda {
            params: vec![Param {
                name: "n".into(),
                type_ann: Some(TypeExpr::Named {
                    name: "Int".into(),
                    span: span(8, 11),
                }),
                span: span(5, 11),
            }],
            return_type: Some(TypeExpr::Named {
                name: "Int".into(),
                span: span(14, 17),
            }),
            body: vec![Expr::Symbol("n".into(), span(18, 19))],
            span: span(0, 20),
        };
        assert_eq!(expr.span(), span(0, 20));
        if let Expr::Lambda { params, body, .. } = &expr {
            assert_eq!(params.len(), 1);
            assert_eq!(params[0].name, "n");
            assert!(params[0].type_ann.is_some());
            assert_eq!(body.len(), 1);
        }
    }

    #[test]
    fn field_access_expr() {
        let expr = Expr::FieldAccess {
            object: Box::new(Expr::Symbol("point".into(), span(3, 8))),
            field: "x".into(),
            span: span(0, 11),
        };
        assert_eq!(expr.span(), span(0, 11));
        if let Expr::FieldAccess { object, field, .. } = &expr {
            assert!(matches!(object.as_ref(), Expr::Symbol(s, _) if s == "point"));
            assert_eq!(field, "x");
        }
    }

    #[test]
    fn call_expr() {
        let expr = Expr::Call {
            func: Box::new(Expr::Symbol("+".into(), span(1, 2))),
            args: vec![Expr::Int(1, span(3, 4)), Expr::Int(2, span(5, 6))],
            span: span(0, 7),
        };
        assert_eq!(expr.span(), span(0, 7));
        if let Expr::Call { func, args, .. } = &expr {
            assert!(matches!(func.as_ref(), Expr::Symbol(name, _) if name == "+"));
            assert_eq!(args.len(), 2);
        }
    }

    #[test]
    fn defn_top_form() {
        let form = TopForm::Defn {
            name: "main".into(),
            params: vec![],
            return_type: None,
            body: vec![Expr::Call {
                func: Box::new(Expr::Symbol("println".into(), span(16, 23))),
                args: vec![Expr::String("Hello, World!".into(), span(24, 39))],
                span: span(15, 40),
            }],
            span: span(0, 41),
        };
        assert_eq!(form.span(), span(0, 41));
        if let TopForm::Defn {
            name, params, body, ..
        } = &form
        {
            assert_eq!(name, "main");
            assert!(params.is_empty());
            assert_eq!(body.len(), 1);
        }
    }

    #[test]
    fn def_top_form() {
        let form = TopForm::Def {
            name: "pi".into(),
            type_ann: Some(TypeExpr::Named {
                name: "Float".into(),
                span: span(9, 14),
            }),
            value: Expr::Float(3.14, span(15, 19)),
            span: span(0, 20),
        };
        assert_eq!(form.span(), span(0, 20));
        assert!(matches!(&form, TopForm::Def { name, .. } if name == "pi"));
    }

    #[test]
    fn expr_top_form() {
        let form = TopForm::Expr(Expr::Int(42, span(0, 2)));
        assert_eq!(form.span(), span(0, 2));
        assert!(matches!(&form, TopForm::Expr(Expr::Int(42, _))));
    }

    #[test]
    fn type_expr_named() {
        let t = TypeExpr::Named {
            name: "Int".into(),
            span: span(0, 3),
        };
        assert_eq!(t.span(), span(0, 3));
    }

    #[test]
    fn type_expr_function() {
        let t = TypeExpr::Function {
            params: vec![
                TypeExpr::Named {
                    name: "Int".into(),
                    span: span(5, 8),
                },
                TypeExpr::Named {
                    name: "Int".into(),
                    span: span(9, 12),
                },
            ],
            ret: Box::new(TypeExpr::Named {
                name: "Int".into(),
                span: span(14, 17),
            }),
            span: span(0, 18),
        };
        assert_eq!(t.span(), span(0, 18));
        if let TypeExpr::Function { params, .. } = &t {
            assert_eq!(params.len(), 2);
        }
    }

    #[test]
    fn type_expr_applied() {
        let t = TypeExpr::Applied {
            name: "List".into(),
            args: vec![TypeExpr::Named {
                name: "Int".into(),
                span: span(6, 9),
            }],
            span: span(0, 10),
        };
        assert_eq!(t.span(), span(0, 10));
        if let TypeExpr::Applied { name, args, .. } = &t {
            assert_eq!(name, "List");
            assert_eq!(args.len(), 1);
        }
    }

    #[test]
    fn field_struct() {
        let field = Field {
            name: "name".into(),
            type_expr: TypeExpr::Named {
                name: "String".into(),
                span: span(6, 12),
            },
            span: span(1, 13),
        };
        assert_eq!(field.name, "name");
        assert_eq!(field.span, span(1, 13));
    }

    #[test]
    fn deftype_top_form() {
        let form = TopForm::Deftype {
            name: "Point".into(),
            fields: vec![
                Field {
                    name: "x".into(),
                    type_expr: TypeExpr::Named {
                        name: "Float".into(),
                        span: span(18, 23),
                    },
                    span: span(15, 24),
                },
                Field {
                    name: "y".into(),
                    type_expr: TypeExpr::Named {
                        name: "Float".into(),
                        span: span(28, 33),
                    },
                    span: span(25, 34),
                },
            ],
            span: span(0, 35),
        };
        assert_eq!(form.span(), span(0, 35));
        if let TopForm::Deftype { name, fields, .. } = &form {
            assert_eq!(name, "Point");
            assert_eq!(fields.len(), 2);
            assert_eq!(fields[0].name, "x");
            assert_eq!(fields[1].name, "y");
        }
    }

    #[test]
    fn deftype_no_fields() {
        let form = TopForm::Deftype {
            name: "Empty".into(),
            fields: vec![],
            span: span(0, 15),
        };
        assert_eq!(form.span(), span(0, 15));
        assert!(
            matches!(&form, TopForm::Deftype { name, fields, .. } if name == "Empty" && fields.is_empty())
        );
    }

    #[test]
    fn hello_world_ast() {
        let program = vec![TopForm::Defn {
            name: "main".into(),
            params: vec![],
            return_type: None,
            body: vec![Expr::Call {
                func: Box::new(Expr::Symbol("println".into(), span(16, 23))),
                args: vec![Expr::String("Hello, World!".into(), span(24, 39))],
                span: span(15, 40),
            }],
            span: span(0, 41),
        }];
        assert_eq!(program.len(), 1);
        assert!(matches!(&program[0], TopForm::Defn { name, .. } if name == "main"));
    }

    #[test]
    fn fibonacci_ast() {
        let fib_body = vec![Expr::If {
            test: Box::new(Expr::Call {
                func: Box::new(Expr::Symbol("<=".into(), span(10, 12))),
                args: vec![
                    Expr::Symbol("n".into(), span(13, 14)),
                    Expr::Int(1, span(15, 16)),
                ],
                span: span(9, 17),
            }),
            then_branch: Box::new(Expr::Symbol("n".into(), span(22, 23))),
            else_branch: Box::new(Expr::Call {
                func: Box::new(Expr::Symbol("+".into(), span(29, 30))),
                args: vec![
                    Expr::Call {
                        func: Box::new(Expr::Symbol("fib".into(), span(32, 35))),
                        args: vec![Expr::Call {
                            func: Box::new(Expr::Symbol("-".into(), span(37, 38))),
                            args: vec![
                                Expr::Symbol("n".into(), span(39, 40)),
                                Expr::Int(1, span(41, 42)),
                            ],
                            span: span(36, 43),
                        }],
                        span: span(31, 44),
                    },
                    Expr::Call {
                        func: Box::new(Expr::Symbol("fib".into(), span(46, 49))),
                        args: vec![Expr::Call {
                            func: Box::new(Expr::Symbol("-".into(), span(51, 52))),
                            args: vec![
                                Expr::Symbol("n".into(), span(53, 54)),
                                Expr::Int(2, span(55, 56)),
                            ],
                            span: span(50, 57),
                        }],
                        span: span(45, 58),
                    },
                ],
                span: span(28, 59),
            }),
            span: span(5, 60),
        }];

        let program = vec![TopForm::Defn {
            name: "fib".into(),
            params: vec![Param {
                name: "n".into(),
                type_ann: Some(TypeExpr::Named {
                    name: "Int".into(),
                    span: span(0, 3),
                }),
                span: span(0, 3),
            }],
            return_type: Some(TypeExpr::Named {
                name: "Int".into(),
                span: span(0, 3),
            }),
            body: fib_body,
            span: span(0, 61),
        }];

        assert_eq!(program.len(), 1);
        if let TopForm::Defn { name, params, .. } = &program[0] {
            assert_eq!(name, "fib");
            assert_eq!(params.len(), 1);
            assert_eq!(params[0].name, "n");
        }
    }

    #[test]
    fn variant_struct() {
        let v = Variant {
            name: "Some".into(),
            types: vec![TypeExpr::Named {
                name: "Int".into(),
                span: span(6, 9),
            }],
            span: span(0, 10),
        };
        assert_eq!(v.name, "Some");
        assert_eq!(v.types.len(), 1);
    }

    #[test]
    fn defunion_top_form() {
        let form = TopForm::Defunion {
            name: "Option".into(),
            variants: vec![
                Variant {
                    name: "Some".into(),
                    types: vec![TypeExpr::Named {
                        name: "Int".into(),
                        span: span(22, 25),
                    }],
                    span: span(16, 26),
                },
                Variant {
                    name: "None".into(),
                    types: vec![],
                    span: span(27, 33),
                },
            ],
            span: span(0, 34),
        };
        assert!(matches!(&form, TopForm::Defunion { name, variants, .. }
            if name == "Option" && variants.len() == 2));
        assert_eq!(form.span(), span(0, 34));
    }

    #[test]
    fn pattern_wildcard() {
        let p = Pattern::Wildcard(span(0, 1));
        assert_eq!(p.span(), span(0, 1));
    }

    #[test]
    fn pattern_binding() {
        let p = Pattern::Binding("x".into(), span(0, 1));
        assert_eq!(p.span(), span(0, 1));
    }

    #[test]
    fn pattern_constructor() {
        let p = Pattern::Constructor {
            name: "Some".into(),
            args: vec![Pattern::Binding("x".into(), span(6, 7))],
            span: span(0, 8),
        };
        assert_eq!(p.span(), span(0, 8));
        if let Pattern::Constructor { name, args, .. } = &p {
            assert_eq!(name, "Some");
            assert_eq!(args.len(), 1);
        }
    }

    #[test]
    fn match_expr() {
        let expr = Expr::Match {
            scrutinee: Box::new(Expr::Symbol("x".into(), span(7, 8))),
            clauses: vec![
                MatchClause {
                    pattern: Pattern::Constructor {
                        name: "Some".into(),
                        args: vec![Pattern::Binding("v".into(), span(15, 16))],
                        span: span(10, 17),
                    },
                    body: Expr::Symbol("v".into(), span(18, 19)),
                    span: span(10, 19),
                },
                MatchClause {
                    pattern: Pattern::Wildcard(span(20, 21)),
                    body: Expr::Nil(span(22, 25)),
                    span: span(20, 25),
                },
            ],
            span: span(0, 26),
        };
        assert_eq!(expr.span(), span(0, 26));
        if let Expr::Match { clauses, .. } = &expr {
            assert_eq!(clauses.len(), 2);
        }
    }

    #[test]
    fn module_top_form() {
        let form = TopForm::Module {
            name: "vex.http".into(),
            span: span(0, 16),
        };
        assert_eq!(form.span(), span(0, 16));
        assert!(matches!(&form, TopForm::Module { name, .. } if name == "vex.http"));
    }

    #[test]
    fn export_top_form() {
        let form = TopForm::Export {
            symbols: vec!["foo".into(), "bar".into()],
            span: span(0, 18),
        };
        assert_eq!(form.span(), span(0, 18));
        if let TopForm::Export { symbols, .. } = &form {
            assert_eq!(symbols, &["foo", "bar"]);
        }
    }

    #[test]
    fn import_top_form() {
        let form = TopForm::Import {
            module_path: "math.core".into(),
            symbols: vec!["add".into(), "sub".into()],
            span: span(0, 28),
        };
        assert_eq!(form.span(), span(0, 28));
        if let TopForm::Import {
            module_path,
            symbols,
            ..
        } = &form
        {
            assert_eq!(module_path, "math.core");
            assert_eq!(symbols, &["add", "sub"]);
        }
    }

    #[test]
    fn import_go_top_form() {
        let form = TopForm::ImportGo {
            go_package: "net/http".into(),
            symbols: vec!["Get".into(), "Post".into()],
            span: span(0, 30),
        };
        assert_eq!(form.span(), span(0, 30));
        if let TopForm::ImportGo {
            go_package,
            symbols,
            ..
        } = &form
        {
            assert_eq!(go_package, "net/http");
            assert_eq!(symbols, &["Get", "Post"]);
        }
    }

    #[test]
    fn spawn_expr() {
        let expr = Expr::Spawn {
            body: Box::new(Expr::Call {
                func: Box::new(Expr::Symbol("println".into(), span(7, 14))),
                args: vec![Expr::String("hi".into(), span(15, 19))],
                span: span(6, 20),
            }),
            span: span(0, 21),
        };
        assert_eq!(expr.span(), span(0, 21));
        assert!(matches!(expr, Expr::Spawn { .. }));
    }

    #[test]
    fn channel_expr() {
        let expr = Expr::Channel {
            element_type: TypeExpr::Named {
                name: "Int".into(),
                span: span(9, 12),
            },
            size: Some(Box::new(Expr::Int(10, span(13, 15)))),
            span: span(0, 16),
        };
        assert_eq!(expr.span(), span(0, 16));
        if let Expr::Channel {
            element_type, size, ..
        } = &expr
        {
            assert!(matches!(element_type, TypeExpr::Named { name, .. } if name == "Int"));
            assert!(matches!(size.as_deref(), Some(Expr::Int(10, _))));
        }
    }

    #[test]
    fn channel_expr_unbuffered() {
        let expr = Expr::Channel {
            element_type: TypeExpr::Named {
                name: "String".into(),
                span: span(9, 15),
            },
            size: None,
            span: span(0, 16),
        };
        assert_eq!(expr.span(), span(0, 16));
        if let Expr::Channel { size, .. } = &expr {
            assert!(size.is_none());
        }
    }

    #[test]
    fn send_expr() {
        let expr = Expr::Send {
            channel: Box::new(Expr::Symbol("ch".into(), span(6, 8))),
            value: Box::new(Expr::Int(42, span(9, 11))),
            span: span(0, 12),
        };
        assert_eq!(expr.span(), span(0, 12));
        if let Expr::Send { channel, value, .. } = &expr {
            assert!(matches!(channel.as_ref(), Expr::Symbol(s, _) if s == "ch"));
            assert!(matches!(value.as_ref(), Expr::Int(42, _)));
        }
    }

    #[test]
    fn recv_expr() {
        let expr = Expr::Recv {
            channel: Box::new(Expr::Symbol("ch".into(), span(6, 8))),
            span: span(0, 9),
        };
        assert_eq!(expr.span(), span(0, 9));
        if let Expr::Recv { channel, .. } = &expr {
            assert!(matches!(channel.as_ref(), Expr::Symbol(s, _) if s == "ch"));
        }
    }

    #[test]
    fn quote_expr() {
        let expr = Expr::Quote {
            expr: Box::new(Expr::Symbol("x".into(), span(7, 8))),
            span: span(0, 9),
        };
        assert_eq!(expr.span(), span(0, 9));
        if let Expr::Quote { expr: inner, .. } = &expr {
            assert!(matches!(inner.as_ref(), Expr::Symbol(s, _) if s == "x"));
        }
    }

    #[test]
    fn unquote_expr() {
        let expr = Expr::Unquote {
            expr: Box::new(Expr::Symbol("val".into(), span(9, 12))),
            span: span(0, 13),
        };
        assert_eq!(expr.span(), span(0, 13));
        if let Expr::Unquote { expr: inner, .. } = &expr {
            assert!(matches!(inner.as_ref(), Expr::Symbol(s, _) if s == "val"));
        }
    }

    #[test]
    fn splice_expr() {
        let expr = Expr::Splice {
            expr: Box::new(Expr::Symbol("args".into(), span(8, 12))),
            span: span(0, 13),
        };
        assert_eq!(expr.span(), span(0, 13));
        if let Expr::Splice { expr: inner, .. } = &expr {
            assert!(matches!(inner.as_ref(), Expr::Symbol(s, _) if s == "args"));
        }
    }

    #[test]
    fn defmacro_top_form() {
        let form = TopForm::DefMacro {
            name: "unless".into(),
            params: vec![
                Param {
                    name: "test".into(),
                    type_ann: None,
                    span: span(17, 21),
                },
                Param {
                    name: "body".into(),
                    type_ann: None,
                    span: span(22, 26),
                },
            ],
            rest_param: None,
            body: vec![Expr::Call {
                func: Box::new(Expr::Symbol("list".into(), span(30, 34))),
                args: vec![
                    Expr::Quote {
                        expr: Box::new(Expr::Symbol("if".into(), span(42, 44))),
                        span: span(35, 45),
                    },
                    Expr::Symbol("test".into(), span(46, 50)),
                    Expr::Quote {
                        expr: Box::new(Expr::Nil(span(58, 61))),
                        span: span(51, 62),
                    },
                    Expr::Symbol("body".into(), span(63, 67)),
                ],
                span: span(29, 68),
            }],
            span: span(0, 69),
        };
        assert_eq!(form.span(), span(0, 69));
        if let TopForm::DefMacro {
            name, params, body, ..
        } = &form
        {
            assert_eq!(name, "unless");
            assert_eq!(params.len(), 2);
            assert_eq!(params[0].name, "test");
            assert_eq!(params[1].name, "body");
            assert_eq!(body.len(), 1);
        }
    }
}
