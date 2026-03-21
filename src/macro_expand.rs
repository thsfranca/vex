use crate::ast::{self, Binding, Expr, TopForm};
use crate::source::Span;

pub fn expand(program: Vec<TopForm>) -> Vec<TopForm> {
    program.into_iter().map(expand_top_form).collect()
}

fn expand_top_form(form: TopForm) -> TopForm {
    match form {
        TopForm::Defn {
            name,
            params,
            return_type,
            body,
            span,
        } => TopForm::Defn {
            name,
            params,
            return_type,
            body: body.into_iter().map(expand_expr).collect(),
            span,
        },
        TopForm::Def {
            name,
            type_ann,
            value,
            span,
        } => TopForm::Def {
            name,
            type_ann,
            value: expand_expr(value),
            span,
        },
        TopForm::Expr(expr) => TopForm::Expr(expand_expr(expr)),
        other => other,
    }
}

fn expand_expr(expr: Expr) -> Expr {
    match expr {
        Expr::Cond {
            clauses,
            else_body,
            span,
        } => expand_cond(clauses, else_body, span),

        Expr::Call { func, args, span } => {
            if let Expr::Symbol(ref name, _) = *func {
                match name.as_str() {
                    "and" if args.len() == 2 => return expand_and(args, span),
                    "or" if args.len() == 2 => return expand_or(args, span),
                    _ => {}
                }
            }
            Expr::Call {
                func: Box::new(expand_expr(*func)),
                args: args.into_iter().map(expand_expr).collect(),
                span,
            }
        }

        Expr::If {
            test,
            then_branch,
            else_branch,
            span,
        } => Expr::If {
            test: Box::new(expand_expr(*test)),
            then_branch: Box::new(expand_expr(*then_branch)),
            else_branch: Box::new(expand_expr(*else_branch)),
            span,
        },

        Expr::Let {
            bindings,
            body,
            span,
        } => Expr::Let {
            bindings: bindings
                .into_iter()
                .map(|b| Binding {
                    name: b.name,
                    value: expand_expr(b.value),
                    span: b.span,
                })
                .collect(),
            body: body.into_iter().map(expand_expr).collect(),
            span,
        },

        Expr::Lambda {
            params,
            return_type,
            body,
            span,
        } => Expr::Lambda {
            params,
            return_type,
            body: body.into_iter().map(expand_expr).collect(),
            span,
        },

        Expr::FieldAccess {
            object,
            field,
            span,
        } => Expr::FieldAccess {
            object: Box::new(expand_expr(*object)),
            field,
            span,
        },

        Expr::Match {
            scrutinee,
            clauses,
            span,
        } => Expr::Match {
            scrutinee: Box::new(expand_expr(*scrutinee)),
            clauses: clauses
                .into_iter()
                .map(|c| ast::MatchClause {
                    pattern: c.pattern,
                    body: expand_expr(c.body),
                    span: c.span,
                })
                .collect(),
            span,
        },

        Expr::Spawn { body, span } => Expr::Spawn {
            body: Box::new(expand_expr(*body)),
            span,
        },

        Expr::Send {
            channel,
            value,
            span,
        } => Expr::Send {
            channel: Box::new(expand_expr(*channel)),
            value: Box::new(expand_expr(*value)),
            span,
        },

        Expr::Recv { channel, span } => Expr::Recv {
            channel: Box::new(expand_expr(*channel)),
            span,
        },

        Expr::Quote { expr, span } => Expr::Quote {
            expr: Box::new(expand_expr(*expr)),
            span,
        },

        Expr::Unquote { expr, span } => Expr::Unquote {
            expr: Box::new(expand_expr(*expr)),
            span,
        },

        Expr::Splice { expr, span } => Expr::Splice {
            expr: Box::new(expand_expr(*expr)),
            span,
        },

        other @ (Expr::Int(..)
        | Expr::Float(..)
        | Expr::String(..)
        | Expr::Bool(..)
        | Expr::Nil(..)
        | Expr::Symbol(..)
        | Expr::Keyword(..)
        | Expr::Channel { .. }) => other,
    }
}

fn expand_cond(clauses: Vec<ast::CondClause>, else_body: Option<Box<Expr>>, span: Span) -> Expr {
    let mut result = match else_body {
        Some(e) => expand_expr(*e),
        None => Expr::Nil(span),
    };

    for clause in clauses.into_iter().rev() {
        result = Expr::If {
            test: Box::new(expand_expr(clause.test)),
            then_branch: Box::new(expand_expr(clause.value)),
            else_branch: Box::new(result),
            span,
        };
    }

    result
}

fn expand_and(mut args: Vec<Expr>, span: Span) -> Expr {
    let b = expand_expr(args.pop().unwrap());
    let a = expand_expr(args.pop().unwrap());
    Expr::If {
        test: Box::new(a),
        then_branch: Box::new(b),
        else_branch: Box::new(Expr::Bool(false, span)),
        span,
    }
}

fn expand_or(mut args: Vec<Expr>, span: Span) -> Expr {
    let b = expand_expr(args.pop().unwrap());
    let a = expand_expr(args.pop().unwrap());
    let tmp_name = "__or_tmp".to_string();
    Expr::Let {
        bindings: vec![Binding {
            name: tmp_name.clone(),
            value: a,
            span,
        }],
        body: vec![Expr::If {
            test: Box::new(Expr::Symbol(tmp_name.clone(), span)),
            then_branch: Box::new(Expr::Symbol(tmp_name, span)),
            else_branch: Box::new(b),
            span,
        }],
        span,
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::source::{FileId, Span};

    fn span(start: u32, end: u32) -> Span {
        Span::new(FileId::new(0), start, end)
    }

    #[test]
    fn expand_cond_to_nested_if() {
        let input = vec![TopForm::Expr(Expr::Cond {
            clauses: vec![
                ast::CondClause {
                    test: Expr::Bool(true, span(0, 4)),
                    value: Expr::Int(1, span(5, 6)),
                    span: span(0, 6),
                },
                ast::CondClause {
                    test: Expr::Bool(false, span(7, 12)),
                    value: Expr::Int(2, span(13, 14)),
                    span: span(7, 14),
                },
            ],
            else_body: Some(Box::new(Expr::Int(3, span(21, 22)))),
            span: span(0, 23),
        })];

        let result = expand(input);
        if let TopForm::Expr(Expr::If {
            test,
            then_branch,
            else_branch,
            ..
        }) = &result[0]
        {
            assert!(matches!(test.as_ref(), Expr::Bool(true, _)));
            assert!(matches!(then_branch.as_ref(), Expr::Int(1, _)));
            assert!(matches!(else_branch.as_ref(), Expr::If { .. }));

            if let Expr::If {
                test: inner_test,
                then_branch: inner_then,
                else_branch: inner_else,
                ..
            } = else_branch.as_ref()
            {
                assert!(matches!(inner_test.as_ref(), Expr::Bool(false, _)));
                assert!(matches!(inner_then.as_ref(), Expr::Int(2, _)));
                assert!(matches!(inner_else.as_ref(), Expr::Int(3, _)));
            } else {
                panic!("expected nested if");
            }
        } else {
            panic!("expected if expression");
        }
    }

    #[test]
    fn expand_cond_no_else() {
        let input = vec![TopForm::Expr(Expr::Cond {
            clauses: vec![ast::CondClause {
                test: Expr::Bool(true, span(0, 4)),
                value: Expr::Int(1, span(5, 6)),
                span: span(0, 6),
            }],
            else_body: None,
            span: span(0, 7),
        })];

        let result = expand(input);
        if let TopForm::Expr(Expr::If { else_branch, .. }) = &result[0] {
            assert!(matches!(else_branch.as_ref(), Expr::Nil(_)));
        } else {
            panic!("expected if expression");
        }
    }

    #[test]
    fn expand_and_to_if() {
        let input = vec![TopForm::Expr(Expr::Call {
            func: Box::new(Expr::Symbol("and".into(), span(1, 4))),
            args: vec![
                Expr::Symbol("a".into(), span(5, 6)),
                Expr::Symbol("b".into(), span(7, 8)),
            ],
            span: span(0, 9),
        })];

        let result = expand(input);
        if let TopForm::Expr(Expr::If {
            test,
            then_branch,
            else_branch,
            ..
        }) = &result[0]
        {
            assert!(matches!(test.as_ref(), Expr::Symbol(s, _) if s == "a"));
            assert!(matches!(then_branch.as_ref(), Expr::Symbol(s, _) if s == "b"));
            assert!(matches!(else_branch.as_ref(), Expr::Bool(false, _)));
        } else {
            panic!("expected if expression");
        }
    }

    #[test]
    fn expand_or_to_let_if() {
        let input = vec![TopForm::Expr(Expr::Call {
            func: Box::new(Expr::Symbol("or".into(), span(1, 3))),
            args: vec![
                Expr::Symbol("a".into(), span(4, 5)),
                Expr::Symbol("b".into(), span(6, 7)),
            ],
            span: span(0, 8),
        })];

        let result = expand(input);
        if let TopForm::Expr(Expr::Let { bindings, body, .. }) = &result[0] {
            assert_eq!(bindings.len(), 1);
            assert_eq!(bindings[0].name, "__or_tmp");
            assert!(matches!(&bindings[0].value, Expr::Symbol(s, _) if s == "a"));

            assert_eq!(body.len(), 1);
            if let Expr::If {
                test,
                then_branch,
                else_branch,
                ..
            } = &body[0]
            {
                assert!(matches!(test.as_ref(), Expr::Symbol(s, _) if s == "__or_tmp"));
                assert!(matches!(then_branch.as_ref(), Expr::Symbol(s, _) if s == "__or_tmp"));
                assert!(matches!(else_branch.as_ref(), Expr::Symbol(s, _) if s == "b"));
            } else {
                panic!("expected if in let body");
            }
        } else {
            panic!("expected let expression");
        }
    }

    #[test]
    fn expand_nested_macros() {
        let input = vec![TopForm::Expr(Expr::Call {
            func: Box::new(Expr::Symbol("and".into(), span(1, 4))),
            args: vec![
                Expr::Call {
                    func: Box::new(Expr::Symbol("or".into(), span(6, 8))),
                    args: vec![
                        Expr::Symbol("a".into(), span(9, 10)),
                        Expr::Symbol("b".into(), span(11, 12)),
                    ],
                    span: span(5, 13),
                },
                Expr::Symbol("c".into(), span(14, 15)),
            ],
            span: span(0, 16),
        })];

        let result = expand(input);
        if let TopForm::Expr(Expr::If { test, .. }) = &result[0] {
            assert!(
                matches!(test.as_ref(), Expr::Let { .. }),
                "inner or should expand to let"
            );
        } else {
            panic!("expected if expression");
        }
    }

    #[test]
    fn expand_preserves_non_macro_calls() {
        let input = vec![TopForm::Expr(Expr::Call {
            func: Box::new(Expr::Symbol("+".into(), span(1, 2))),
            args: vec![Expr::Int(1, span(3, 4)), Expr::Int(2, span(5, 6))],
            span: span(0, 7),
        })];

        let result = expand(input);
        assert!(matches!(
            &result[0],
            TopForm::Expr(Expr::Call { func, .. }) if matches!(func.as_ref(), Expr::Symbol(s, _) if s == "+")
        ));
    }

    #[test]
    fn expand_in_defn_body() {
        let input = vec![TopForm::Defn {
            name: "test".into(),
            params: vec![],
            return_type: None,
            body: vec![Expr::Cond {
                clauses: vec![ast::CondClause {
                    test: Expr::Bool(true, span(0, 4)),
                    value: Expr::Int(1, span(5, 6)),
                    span: span(0, 6),
                }],
                else_body: Some(Box::new(Expr::Int(0, span(13, 14)))),
                span: span(0, 15),
            }],
            span: span(0, 20),
        }];

        let result = expand(input);
        if let TopForm::Defn { body, .. } = &result[0] {
            assert!(matches!(&body[0], Expr::If { .. }));
        } else {
            panic!("expected defn");
        }
    }

    #[test]
    fn expand_in_lambda_body() {
        let input = vec![TopForm::Expr(Expr::Lambda {
            params: vec![],
            return_type: None,
            body: vec![Expr::Call {
                func: Box::new(Expr::Symbol("and".into(), span(1, 4))),
                args: vec![
                    Expr::Bool(true, span(5, 9)),
                    Expr::Bool(false, span(10, 15)),
                ],
                span: span(0, 16),
            }],
            span: span(0, 17),
        })];

        let result = expand(input);
        if let TopForm::Expr(Expr::Lambda { body, .. }) = &result[0] {
            assert!(matches!(&body[0], Expr::If { .. }));
        } else {
            panic!("expected lambda");
        }
    }
}
