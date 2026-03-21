use std::collections::HashMap;

use crate::ast::{self, Binding, Expr, TopForm};
use crate::diagnostics::Diagnostic;
use crate::source::Span;
use crate::types::{SyntaxValue, expr_to_syntax, syntax_to_expr};

const MAX_EXPANSION_DEPTH: usize = 64;

struct MacroDef {
    params: Vec<String>,
    body: Vec<Expr>,
}

pub fn expand(program: Vec<TopForm>) -> (Vec<TopForm>, Vec<Diagnostic>) {
    let mut registry: HashMap<String, MacroDef> = HashMap::new();
    let mut diagnostics = Vec::new();

    for form in &program {
        if let TopForm::DefMacro {
            name, params, body, ..
        } = form
        {
            registry.insert(
                name.clone(),
                MacroDef {
                    params: params.iter().map(|p| p.name.clone()).collect(),
                    body: body.clone(),
                },
            );
        }
    }

    let expanded: Vec<TopForm> = program
        .into_iter()
        .filter(|form| !matches!(form, TopForm::DefMacro { .. }))
        .map(|form| expand_top_form(form, &registry, &mut diagnostics))
        .collect();

    (expanded, diagnostics)
}

fn expand_top_form(
    form: TopForm,
    registry: &HashMap<String, MacroDef>,
    diagnostics: &mut Vec<Diagnostic>,
) -> TopForm {
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
            body: body
                .into_iter()
                .map(|e| expand_expr(e, registry, diagnostics, 0))
                .collect(),
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
            value: expand_expr(value, registry, diagnostics, 0),
            span,
        },
        TopForm::Expr(expr) => TopForm::Expr(expand_expr(expr, registry, diagnostics, 0)),
        other => other,
    }
}

fn expand_expr(
    expr: Expr,
    registry: &HashMap<String, MacroDef>,
    diagnostics: &mut Vec<Diagnostic>,
    depth: usize,
) -> Expr {
    if depth > MAX_EXPANSION_DEPTH {
        diagnostics.push(Diagnostic::error(
            "macro expansion depth limit exceeded",
            expr.span(),
        ));
        return expr;
    }

    match expr {
        Expr::Cond {
            clauses,
            else_body,
            span,
        } => expand_cond(clauses, else_body, span, registry, diagnostics, depth),

        Expr::Call { func, args, span } => {
            if let Expr::Symbol(ref name, _) = *func {
                match name.as_str() {
                    "and" if args.len() == 2 => {
                        return expand_and(args, span, registry, diagnostics, depth);
                    }
                    "or" if args.len() == 2 => {
                        return expand_or(args, span, registry, diagnostics, depth);
                    }
                    _ => {
                        if let Some(macro_def) = registry.get(name.as_str()) {
                            if args.len() != macro_def.params.len() {
                                diagnostics.push(Diagnostic::error(
                                    format!(
                                        "macro '{}' expects {} arguments, got {}",
                                        name,
                                        macro_def.params.len(),
                                        args.len()
                                    ),
                                    span,
                                ));
                                return Expr::Nil(span);
                            }

                            let mut env: HashMap<String, SyntaxValue> = HashMap::new();
                            for (param, arg) in macro_def.params.iter().zip(args.iter()) {
                                env.insert(param.clone(), expr_to_syntax(arg));
                            }

                            let mut result = SyntaxValue::Nil;
                            for body_expr in &macro_def.body {
                                match eval_macro_expr(body_expr, &env) {
                                    Ok(val) => result = val,
                                    Err(msg) => {
                                        diagnostics.push(Diagnostic::error(
                                            format!("macro expansion error: {}", msg),
                                            span,
                                        ));
                                        return Expr::Nil(span);
                                    }
                                }
                            }

                            let expanded_ast = syntax_to_expr(&result, span);
                            return expand_expr(expanded_ast, registry, diagnostics, depth + 1);
                        }
                    }
                }
            }
            Expr::Call {
                func: Box::new(expand_expr(*func, registry, diagnostics, depth)),
                args: args
                    .into_iter()
                    .map(|a| expand_expr(a, registry, diagnostics, depth))
                    .collect(),
                span,
            }
        }

        Expr::If {
            test,
            then_branch,
            else_branch,
            span,
        } => Expr::If {
            test: Box::new(expand_expr(*test, registry, diagnostics, depth)),
            then_branch: Box::new(expand_expr(*then_branch, registry, diagnostics, depth)),
            else_branch: Box::new(expand_expr(*else_branch, registry, diagnostics, depth)),
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
                    value: expand_expr(b.value, registry, diagnostics, depth),
                    span: b.span,
                })
                .collect(),
            body: body
                .into_iter()
                .map(|e| expand_expr(e, registry, diagnostics, depth))
                .collect(),
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
            body: body
                .into_iter()
                .map(|e| expand_expr(e, registry, diagnostics, depth))
                .collect(),
            span,
        },

        Expr::FieldAccess {
            object,
            field,
            span,
        } => Expr::FieldAccess {
            object: Box::new(expand_expr(*object, registry, diagnostics, depth)),
            field,
            span,
        },

        Expr::Match {
            scrutinee,
            clauses,
            span,
        } => Expr::Match {
            scrutinee: Box::new(expand_expr(*scrutinee, registry, diagnostics, depth)),
            clauses: clauses
                .into_iter()
                .map(|c| ast::MatchClause {
                    pattern: c.pattern,
                    body: expand_expr(c.body, registry, diagnostics, depth),
                    span: c.span,
                })
                .collect(),
            span,
        },

        Expr::Spawn { body, span } => Expr::Spawn {
            body: Box::new(expand_expr(*body, registry, diagnostics, depth)),
            span,
        },

        Expr::Send {
            channel,
            value,
            span,
        } => Expr::Send {
            channel: Box::new(expand_expr(*channel, registry, diagnostics, depth)),
            value: Box::new(expand_expr(*value, registry, diagnostics, depth)),
            span,
        },

        Expr::Recv { channel, span } => Expr::Recv {
            channel: Box::new(expand_expr(*channel, registry, diagnostics, depth)),
            span,
        },

        Expr::Quote { expr, span } => Expr::Quote {
            expr: Box::new(expand_expr(*expr, registry, diagnostics, depth)),
            span,
        },

        Expr::Unquote { expr, span } => Expr::Unquote {
            expr: Box::new(expand_expr(*expr, registry, diagnostics, depth)),
            span,
        },

        Expr::Splice { expr, span } => Expr::Splice {
            expr: Box::new(expand_expr(*expr, registry, diagnostics, depth)),
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

fn eval_macro_expr(expr: &Expr, env: &HashMap<String, SyntaxValue>) -> Result<SyntaxValue, String> {
    match expr {
        Expr::Int(n, _) => Ok(SyntaxValue::Int(*n)),
        Expr::Float(n, _) => Ok(SyntaxValue::Float(*n)),
        Expr::String(s, _) => Ok(SyntaxValue::Str(s.clone())),
        Expr::Bool(b, _) => Ok(SyntaxValue::Bool(*b)),
        Expr::Nil(_) => Ok(SyntaxValue::Nil),
        Expr::Keyword(s, _) => Ok(SyntaxValue::Kw(s.clone())),

        Expr::Symbol(name, _) => env
            .get(name)
            .cloned()
            .ok_or_else(|| format!("undefined variable in macro body: {}", name)),

        Expr::Quote { expr, .. } => Ok(expr_to_syntax(expr)),

        Expr::If {
            test,
            then_branch,
            else_branch,
            ..
        } => {
            let cond = eval_macro_expr(test, env)?;
            match cond {
                SyntaxValue::Bool(true) => eval_macro_expr(then_branch, env),
                SyntaxValue::Bool(false) => eval_macro_expr(else_branch, env),
                _ => Err("if condition must be Bool in macro body".into()),
            }
        }

        Expr::Let { bindings, body, .. } => {
            let mut local_env = env.clone();
            for binding in bindings {
                let val = eval_macro_expr(&binding.value, &local_env)?;
                local_env.insert(binding.name.clone(), val);
            }
            let mut result = SyntaxValue::Nil;
            for e in body {
                result = eval_macro_expr(e, &local_env)?;
            }
            Ok(result)
        }

        Expr::Call { func, args, .. } => {
            if let Expr::Symbol(name, _) = func.as_ref() {
                let mut arg_vals = Vec::new();
                for arg in args {
                    arg_vals.push(eval_macro_expr(arg, env)?);
                }
                eval_macro_builtin(name, arg_vals)
            } else {
                Err("macro body: only named function calls are supported".into())
            }
        }

        _ => Err(format!(
            "unsupported expression form in macro body: {:?}",
            std::mem::discriminant(expr)
        )),
    }
}

fn eval_macro_builtin(name: &str, args: Vec<SyntaxValue>) -> Result<SyntaxValue, String> {
    match name {
        "syntax-list" => Ok(SyntaxValue::List(args)),

        "syntax-cons" => {
            if args.len() != 2 {
                return Err("syntax-cons requires 2 arguments".into());
            }
            let head = args[0].clone();
            match &args[1] {
                SyntaxValue::List(items) => {
                    let mut result = vec![head];
                    result.extend(items.iter().cloned());
                    Ok(SyntaxValue::List(result))
                }
                _ => Err("syntax-cons: second argument must be a list".into()),
            }
        }

        "syntax-first" => {
            if args.len() != 1 {
                return Err("syntax-first requires 1 argument".into());
            }
            match &args[0] {
                SyntaxValue::List(items) if !items.is_empty() => Ok(items[0].clone()),
                SyntaxValue::List(_) => Err("syntax-first: empty list".into()),
                _ => Err("syntax-first: argument must be a list".into()),
            }
        }

        "syntax-rest" => {
            if args.len() != 1 {
                return Err("syntax-rest requires 1 argument".into());
            }
            match &args[0] {
                SyntaxValue::List(items) if !items.is_empty() => {
                    Ok(SyntaxValue::List(items[1..].to_vec()))
                }
                SyntaxValue::List(_) => Err("syntax-rest: empty list".into()),
                _ => Err("syntax-rest: argument must be a list".into()),
            }
        }

        "syntax-symbol?" => {
            if args.len() != 1 {
                return Err("syntax-symbol? requires 1 argument".into());
            }
            Ok(SyntaxValue::Bool(matches!(&args[0], SyntaxValue::Sym(_))))
        }

        "syntax-list?" => {
            if args.len() != 1 {
                return Err("syntax-list? requires 1 argument".into());
            }
            Ok(SyntaxValue::Bool(matches!(&args[0], SyntaxValue::List(_))))
        }

        "syntax-concat" => {
            if args.len() != 2 {
                return Err("syntax-concat requires 2 arguments".into());
            }
            match (&args[0], &args[1]) {
                (SyntaxValue::List(a), SyntaxValue::List(b)) => {
                    let mut result = a.clone();
                    result.extend(b.iter().cloned());
                    Ok(SyntaxValue::List(result))
                }
                _ => Err("syntax-concat: both arguments must be lists".into()),
            }
        }

        _ => Err(format!("unknown function in macro body: {}", name)),
    }
}

fn expand_cond(
    clauses: Vec<ast::CondClause>,
    else_body: Option<Box<Expr>>,
    span: Span,
    registry: &HashMap<String, MacroDef>,
    diagnostics: &mut Vec<Diagnostic>,
    depth: usize,
) -> Expr {
    let mut result = match else_body {
        Some(e) => expand_expr(*e, registry, diagnostics, depth),
        None => Expr::Nil(span),
    };

    for clause in clauses.into_iter().rev() {
        result = Expr::If {
            test: Box::new(expand_expr(clause.test, registry, diagnostics, depth)),
            then_branch: Box::new(expand_expr(clause.value, registry, diagnostics, depth)),
            else_branch: Box::new(result),
            span,
        };
    }

    result
}

fn expand_and(
    mut args: Vec<Expr>,
    span: Span,
    registry: &HashMap<String, MacroDef>,
    diagnostics: &mut Vec<Diagnostic>,
    depth: usize,
) -> Expr {
    let b = expand_expr(args.pop().unwrap(), registry, diagnostics, depth);
    let a = expand_expr(args.pop().unwrap(), registry, diagnostics, depth);
    Expr::If {
        test: Box::new(a),
        then_branch: Box::new(b),
        else_branch: Box::new(Expr::Bool(false, span)),
        span,
    }
}

fn expand_or(
    mut args: Vec<Expr>,
    span: Span,
    registry: &HashMap<String, MacroDef>,
    diagnostics: &mut Vec<Diagnostic>,
    depth: usize,
) -> Expr {
    let b = expand_expr(args.pop().unwrap(), registry, diagnostics, depth);
    let a = expand_expr(args.pop().unwrap(), registry, diagnostics, depth);
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

        let (result, diags) = expand(input);
        assert!(diags.is_empty());
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

        let (result, diags) = expand(input);
        assert!(diags.is_empty());
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

        let (result, diags) = expand(input);
        assert!(diags.is_empty());
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

        let (result, diags) = expand(input);
        assert!(diags.is_empty());
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

        let (result, diags) = expand(input);
        assert!(diags.is_empty());
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

        let (result, diags) = expand(input);
        assert!(diags.is_empty());
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

        let (result, diags) = expand(input);
        assert!(diags.is_empty());
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

        let (result, diags) = expand(input);
        assert!(diags.is_empty());
        if let TopForm::Expr(Expr::Lambda { body, .. }) = &result[0] {
            assert!(matches!(&body[0], Expr::If { .. }));
        } else {
            panic!("expected lambda");
        }
    }

    #[test]
    fn defmacro_unless_expands() {
        let s = span(0, 1);
        let input = vec![
            TopForm::DefMacro {
                name: "unless".into(),
                params: vec![
                    ast::Param {
                        name: "test".into(),
                        type_ann: None,
                        span: s,
                    },
                    ast::Param {
                        name: "body".into(),
                        type_ann: None,
                        span: s,
                    },
                ],
                body: vec![Expr::Call {
                    func: Box::new(Expr::Symbol("syntax-list".into(), s)),
                    args: vec![
                        Expr::Quote {
                            expr: Box::new(Expr::Symbol("if".into(), s)),
                            span: s,
                        },
                        Expr::Symbol("test".into(), s),
                        Expr::Quote {
                            expr: Box::new(Expr::Nil(s)),
                            span: s,
                        },
                        Expr::Symbol("body".into(), s),
                    ],
                    span: s,
                }],
                span: s,
            },
            TopForm::Expr(Expr::Call {
                func: Box::new(Expr::Symbol("unless".into(), s)),
                args: vec![
                    Expr::Call {
                        func: Box::new(Expr::Symbol(">".into(), s)),
                        args: vec![Expr::Symbol("x".into(), s), Expr::Int(10, s)],
                        span: s,
                    },
                    Expr::Call {
                        func: Box::new(Expr::Symbol("println".into(), s)),
                        args: vec![Expr::String("small".into(), s)],
                        span: s,
                    },
                ],
                span: s,
            }),
        ];

        let (result, diags) = expand(input);
        assert!(diags.is_empty(), "{:?}", diags);
        assert_eq!(result.len(), 1);
        if let TopForm::Expr(Expr::If {
            test,
            then_branch,
            else_branch,
            ..
        }) = &result[0]
        {
            assert!(matches!(test.as_ref(), Expr::Call { .. }));
            assert!(matches!(then_branch.as_ref(), Expr::Nil(_)));
            assert!(matches!(else_branch.as_ref(), Expr::Call { .. }));
        } else {
            panic!("expected if expression, got {:?}", result[0]);
        }
    }

    #[test]
    fn defmacro_stripped_from_output() {
        let s = span(0, 1);
        let input = vec![
            TopForm::DefMacro {
                name: "noop".into(),
                params: vec![],
                body: vec![Expr::Quote {
                    expr: Box::new(Expr::Nil(s)),
                    span: s,
                }],
                span: s,
            },
            TopForm::Expr(Expr::Int(42, s)),
        ];

        let (result, diags) = expand(input);
        assert!(diags.is_empty());
        assert_eq!(result.len(), 1);
        assert!(matches!(&result[0], TopForm::Expr(Expr::Int(42, _))));
    }

    #[test]
    fn defmacro_wrong_arg_count() {
        let s = span(0, 1);
        let input = vec![
            TopForm::DefMacro {
                name: "my-macro".into(),
                params: vec![ast::Param {
                    name: "x".into(),
                    type_ann: None,
                    span: s,
                }],
                body: vec![Expr::Symbol("x".into(), s)],
                span: s,
            },
            TopForm::Expr(Expr::Call {
                func: Box::new(Expr::Symbol("my-macro".into(), s)),
                args: vec![Expr::Int(1, s), Expr::Int(2, s)],
                span: s,
            }),
        ];

        let (_, diags) = expand(input);
        assert!(!diags.is_empty());
    }

    #[test]
    fn defmacro_with_let_in_body() {
        let s = span(0, 1);
        let input = vec![
            TopForm::DefMacro {
                name: "wrap-if".into(),
                params: vec![
                    ast::Param {
                        name: "cond-expr".into(),
                        type_ann: None,
                        span: s,
                    },
                    ast::Param {
                        name: "body-expr".into(),
                        type_ann: None,
                        span: s,
                    },
                ],
                body: vec![Expr::Let {
                    bindings: vec![Binding {
                        name: "result".into(),
                        value: Expr::Call {
                            func: Box::new(Expr::Symbol("syntax-list".into(), s)),
                            args: vec![
                                Expr::Quote {
                                    expr: Box::new(Expr::Symbol("if".into(), s)),
                                    span: s,
                                },
                                Expr::Symbol("cond-expr".into(), s),
                                Expr::Symbol("body-expr".into(), s),
                                Expr::Quote {
                                    expr: Box::new(Expr::Nil(s)),
                                    span: s,
                                },
                            ],
                            span: s,
                        },
                        span: s,
                    }],
                    body: vec![Expr::Symbol("result".into(), s)],
                    span: s,
                }],
                span: s,
            },
            TopForm::Expr(Expr::Call {
                func: Box::new(Expr::Symbol("wrap-if".into(), s)),
                args: vec![Expr::Bool(true, s), Expr::Int(42, s)],
                span: s,
            }),
        ];

        let (result, diags) = expand(input);
        assert!(diags.is_empty(), "{:?}", diags);
        assert_eq!(result.len(), 1);
        assert!(matches!(&result[0], TopForm::Expr(Expr::If { .. })));
    }

    #[test]
    fn defmacro_in_defn_body() {
        let s = span(0, 1);
        let input = vec![
            TopForm::DefMacro {
                name: "double".into(),
                params: vec![ast::Param {
                    name: "x".into(),
                    type_ann: None,
                    span: s,
                }],
                body: vec![Expr::Call {
                    func: Box::new(Expr::Symbol("syntax-list".into(), s)),
                    args: vec![
                        Expr::Quote {
                            expr: Box::new(Expr::Symbol("+".into(), s)),
                            span: s,
                        },
                        Expr::Symbol("x".into(), s),
                        Expr::Symbol("x".into(), s),
                    ],
                    span: s,
                }],
                span: s,
            },
            TopForm::Defn {
                name: "test".into(),
                params: vec![],
                return_type: None,
                body: vec![Expr::Call {
                    func: Box::new(Expr::Symbol("double".into(), s)),
                    args: vec![Expr::Int(21, s)],
                    span: s,
                }],
                span: s,
            },
        ];

        let (result, diags) = expand(input);
        assert!(diags.is_empty(), "{:?}", diags);
        assert_eq!(result.len(), 1);
        if let TopForm::Defn { body, .. } = &result[0] {
            assert!(
                matches!(&body[0], Expr::Call { func, .. } if matches!(func.as_ref(), Expr::Symbol(s, _) if s == "+"))
            );
        } else {
            panic!("expected defn");
        }
    }
}
