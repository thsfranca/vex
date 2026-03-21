use std::collections::HashMap;
use std::sync::atomic::{AtomicUsize, Ordering};

use crate::ast::{self, Binding, Expr, TopForm};
use crate::diagnostics::Diagnostic;
use crate::lexer;
use crate::parser;
use crate::source::FileId;
use crate::types::{SyntaxValue, expr_to_syntax, syntax_to_expr};

const MAX_EXPANSION_DEPTH: usize = 64;
const PRELUDE_SOURCE: &str = include_str!("../stdlib/prelude.vx");

static GENSYM_COUNTER: AtomicUsize = AtomicUsize::new(0);

fn gensym(base: &str) -> String {
    let id = GENSYM_COUNTER.fetch_add(1, Ordering::Relaxed);
    format!("__{}_{}", base, id)
}

struct MacroDef {
    params: Vec<String>,
    rest_param: Option<String>,
    body: Vec<Expr>,
}

#[derive(Debug, Clone)]
enum MacroVal {
    Intro(SyntaxValue),
    CallSite(SyntaxValue),
    MList(Vec<MacroVal>),
}

fn macro_val_to_syntax(val: &MacroVal) -> SyntaxValue {
    match val {
        MacroVal::Intro(s) | MacroVal::CallSite(s) => s.clone(),
        MacroVal::MList(items) => {
            SyntaxValue::List(items.iter().map(macro_val_to_syntax).collect())
        }
    }
}

fn load_prelude(registry: &mut HashMap<String, MacroDef>) {
    let prelude_file = FileId::new(u32::MAX);
    let (tokens, _) = lexer::lex(PRELUDE_SOURCE, prelude_file);
    let (forms, _) = parser::parse(&tokens);
    for form in forms {
        if let TopForm::DefMacro {
            name,
            params,
            rest_param,
            body,
            ..
        } = form
        {
            registry.insert(
                name,
                MacroDef {
                    params: params.into_iter().map(|p| p.name).collect(),
                    rest_param,
                    body,
                },
            );
        }
    }
}

pub fn expand(program: Vec<TopForm>) -> (Vec<TopForm>, Vec<Diagnostic>) {
    let mut registry: HashMap<String, MacroDef> = HashMap::new();
    let mut diagnostics = Vec::new();

    load_prelude(&mut registry);

    for form in &program {
        if let TopForm::DefMacro {
            name,
            params,
            rest_param,
            body,
            ..
        } = form
        {
            registry.insert(
                name.clone(),
                MacroDef {
                    params: params.iter().map(|p| p.name.clone()).collect(),
                    rest_param: rest_param.clone(),
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
        Expr::Call { func, args, span } => {
            if let Expr::Symbol(ref name, _) = *func
                && let Some(macro_def) = registry.get(name.as_str())
            {
                let min_args = macro_def.params.len();
                let is_variadic = macro_def.rest_param.is_some();

                if is_variadic {
                    if args.len() < min_args {
                        diagnostics.push(Diagnostic::error(
                            format!(
                                "macro '{}' expects at least {} arguments, got {}",
                                name,
                                min_args,
                                args.len()
                            ),
                            span,
                        ));
                        return Expr::Nil(span);
                    }
                } else if args.len() != min_args {
                    diagnostics.push(Diagnostic::error(
                        format!(
                            "macro '{}' expects {} arguments, got {}",
                            name,
                            min_args,
                            args.len()
                        ),
                        span,
                    ));
                    return Expr::Nil(span);
                }

                let mut env: HashMap<String, MacroVal> = HashMap::new();
                for (param, arg) in macro_def.params.iter().zip(args.iter()) {
                    env.insert(param.clone(), MacroVal::CallSite(expr_to_syntax(arg)));
                }
                if let Some(rest_name) = &macro_def.rest_param {
                    let rest_vals: Vec<SyntaxValue> =
                        args[min_args..].iter().map(expr_to_syntax).collect();
                    env.insert(
                        rest_name.clone(),
                        MacroVal::CallSite(SyntaxValue::List(rest_vals)),
                    );
                }

                let mut result = MacroVal::Intro(SyntaxValue::Nil);
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

                let renamed = rename_in_macro_val(result);
                let result_syntax = macro_val_to_syntax(&renamed);
                let expanded_ast = syntax_to_expr(&result_syntax, span);
                return expand_expr(expanded_ast, registry, diagnostics, depth + 1);
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

fn eval_macro_expr(expr: &Expr, env: &HashMap<String, MacroVal>) -> Result<MacroVal, String> {
    match expr {
        Expr::Int(n, _) => Ok(MacroVal::Intro(SyntaxValue::Int(*n))),
        Expr::Float(n, _) => Ok(MacroVal::Intro(SyntaxValue::Float(*n))),
        Expr::String(s, _) => Ok(MacroVal::Intro(SyntaxValue::Str(s.clone()))),
        Expr::Bool(b, _) => Ok(MacroVal::Intro(SyntaxValue::Bool(*b))),
        Expr::Nil(_) => Ok(MacroVal::Intro(SyntaxValue::Nil)),
        Expr::Keyword(s, _) => Ok(MacroVal::Intro(SyntaxValue::Kw(s.clone()))),

        Expr::Symbol(name, _) => env
            .get(name)
            .cloned()
            .ok_or_else(|| format!("undefined variable in macro body: {}", name)),

        Expr::Quote { expr, .. } => Ok(MacroVal::Intro(expr_to_syntax(expr))),

        Expr::If {
            test,
            then_branch,
            else_branch,
            ..
        } => {
            let cond = eval_macro_expr(test, env)?;
            match macro_val_to_syntax(&cond) {
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
            let mut result = MacroVal::Intro(SyntaxValue::Nil);
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

fn mval_to_list_items(val: &MacroVal) -> Option<Vec<MacroVal>> {
    match val {
        MacroVal::MList(items) => Some(items.clone()),
        MacroVal::Intro(SyntaxValue::List(items)) => {
            Some(items.iter().map(|s| MacroVal::Intro(s.clone())).collect())
        }
        MacroVal::CallSite(SyntaxValue::List(items)) => Some(
            items
                .iter()
                .map(|s| MacroVal::CallSite(s.clone()))
                .collect(),
        ),
        _ => None,
    }
}

fn eval_macro_builtin(name: &str, args: Vec<MacroVal>) -> Result<MacroVal, String> {
    match name {
        "list" => Ok(MacroVal::MList(args)),

        "cons" => {
            if args.len() != 2 {
                return Err("syntax-cons requires 2 arguments".into());
            }
            match mval_to_list_items(&args[1]) {
                Some(mut items) => {
                    items.insert(0, args[0].clone());
                    Ok(MacroVal::MList(items))
                }
                None => Err("syntax-cons: second argument must be a list".into()),
            }
        }

        "first" => {
            if args.len() != 1 {
                return Err("syntax-first requires 1 argument".into());
            }
            match mval_to_list_items(&args[0]) {
                Some(items) if !items.is_empty() => Ok(items[0].clone()),
                _ => Err("syntax-first: argument must be a non-empty list".into()),
            }
        }

        "rest" => {
            if args.len() != 1 {
                return Err("syntax-rest requires 1 argument".into());
            }
            match mval_to_list_items(&args[0]) {
                Some(items) if !items.is_empty() => Ok(MacroVal::MList(items[1..].to_vec())),
                _ => Err("syntax-rest: argument must be a non-empty list".into()),
            }
        }

        "symbol?" => {
            if args.len() != 1 {
                return Err("syntax-symbol? requires 1 argument".into());
            }
            let s = macro_val_to_syntax(&args[0]);
            Ok(MacroVal::Intro(SyntaxValue::Bool(matches!(
                s,
                SyntaxValue::Sym(_)
            ))))
        }

        "list?" => {
            if args.len() != 1 {
                return Err("syntax-list? requires 1 argument".into());
            }
            let s = macro_val_to_syntax(&args[0]);
            Ok(MacroVal::Intro(SyntaxValue::Bool(matches!(
                s,
                SyntaxValue::List(_)
            ))))
        }

        "concat" => {
            if args.len() != 2 {
                return Err("syntax-concat requires 2 arguments".into());
            }
            match (mval_to_list_items(&args[0]), mval_to_list_items(&args[1])) {
                (Some(mut a), Some(b)) => {
                    a.extend(b);
                    Ok(MacroVal::MList(a))
                }
                _ => Err("syntax-concat: both arguments must be lists".into()),
            }
        }

        "empty?" => {
            if args.len() != 1 {
                return Err("empty? requires 1 argument".into());
            }
            match mval_to_list_items(&args[0]) {
                Some(items) => Ok(MacroVal::Intro(SyntaxValue::Bool(items.is_empty()))),
                None => Ok(MacroVal::Intro(SyntaxValue::Bool(matches!(
                    macro_val_to_syntax(&args[0]),
                    SyntaxValue::Nil
                )))),
            }
        }

        "keyword?" => {
            if args.len() != 1 {
                return Err("keyword? requires 1 argument".into());
            }
            let s = macro_val_to_syntax(&args[0]);
            Ok(MacroVal::Intro(SyntaxValue::Bool(matches!(
                s,
                SyntaxValue::Kw(_)
            ))))
        }

        _ => Err(format!("unknown function in macro body: {}", name)),
    }
}

fn rename_in_macro_val(val: MacroVal) -> MacroVal {
    let mut renames: HashMap<String, String> = HashMap::new();
    rename_mval(val, &mut renames)
}

fn rename_mval(val: MacroVal, renames: &mut HashMap<String, String>) -> MacroVal {
    match val {
        MacroVal::CallSite(_) => val,
        MacroVal::Intro(s) => MacroVal::Intro(rename_syntax(s, renames)),
        MacroVal::MList(items) => {
            if is_mlist_let_form(&items) {
                rename_mlist_let(items, renames)
            } else if is_mlist_fn_form(&items) {
                rename_mlist_fn(items, renames)
            } else {
                MacroVal::MList(items.into_iter().map(|v| rename_mval(v, renames)).collect())
            }
        }
    }
}

fn is_mlist_let_form(items: &[MacroVal]) -> bool {
    if items.len() < 3 {
        return false;
    }
    matches!(macro_val_to_syntax(&items[0]), SyntaxValue::Sym(ref s) if s == "let")
}

fn is_mlist_fn_form(items: &[MacroVal]) -> bool {
    if items.len() < 3 {
        return false;
    }
    matches!(macro_val_to_syntax(&items[0]), SyntaxValue::Sym(ref s) if s == "fn")
}

fn rename_mlist_let(items: Vec<MacroVal>, renames: &mut HashMap<String, String>) -> MacroVal {
    let mut result = vec![items[0].clone()];
    let mut body_renames = renames.clone();

    if let Some(binding_items) = mval_to_list_items(&items[1]) {
        let mut new_bindings = Vec::new();
        let mut i = 0;
        while i + 1 < binding_items.len() {
            let name_val = &binding_items[i];
            let value_val = &binding_items[i + 1];
            if let MacroVal::Intro(SyntaxValue::Sym(name)) = name_val {
                let new_name = gensym(name);
                body_renames.insert(name.clone(), new_name.clone());
                new_bindings.push(MacroVal::Intro(SyntaxValue::Sym(new_name)));
                new_bindings.push(rename_mval(value_val.clone(), renames));
            } else {
                new_bindings.push(rename_mval(name_val.clone(), renames));
                new_bindings.push(rename_mval(value_val.clone(), renames));
            }
            i += 2;
        }
        result.push(MacroVal::MList(new_bindings));
    } else {
        result.push(rename_mval(items[1].clone(), renames));
    }

    for item in items.into_iter().skip(2) {
        result.push(rename_mval(item, &mut body_renames));
    }

    MacroVal::MList(result)
}

fn rename_mlist_fn(items: Vec<MacroVal>, renames: &mut HashMap<String, String>) -> MacroVal {
    let mut result = vec![items[0].clone()];
    let mut body_renames = renames.clone();

    if let Some(param_items) = mval_to_list_items(&items[1]) {
        let new_params: Vec<MacroVal> = param_items
            .into_iter()
            .map(|p| {
                if let MacroVal::Intro(SyntaxValue::Sym(name)) = &p {
                    let new_name = gensym(name);
                    body_renames.insert(name.clone(), new_name.clone());
                    MacroVal::Intro(SyntaxValue::Sym(new_name))
                } else {
                    rename_mval(p, renames)
                }
            })
            .collect();
        result.push(MacroVal::MList(new_params));
    } else {
        result.push(rename_mval(items[1].clone(), renames));
    }

    for item in items.into_iter().skip(2) {
        result.push(rename_mval(item, &mut body_renames));
    }

    MacroVal::MList(result)
}

fn rename_syntax(syntax: SyntaxValue, renames: &HashMap<String, String>) -> SyntaxValue {
    match syntax {
        SyntaxValue::Sym(ref name) => {
            if let Some(new_name) = renames.get(name) {
                SyntaxValue::Sym(new_name.clone())
            } else {
                syntax
            }
        }
        SyntaxValue::List(items) => {
            if is_syntax_let_form(&items) {
                rename_syntax_let(items, renames)
            } else if is_syntax_fn_form(&items) {
                rename_syntax_fn(items, renames)
            } else {
                SyntaxValue::List(
                    items
                        .into_iter()
                        .map(|s| rename_syntax(s, renames))
                        .collect(),
                )
            }
        }
        other => other,
    }
}

fn is_syntax_let_form(items: &[SyntaxValue]) -> bool {
    items.len() >= 3 && matches!(&items[0], SyntaxValue::Sym(s) if s == "let")
}

fn is_syntax_fn_form(items: &[SyntaxValue]) -> bool {
    items.len() >= 3 && matches!(&items[0], SyntaxValue::Sym(s) if s == "fn")
}

fn rename_syntax_let(items: Vec<SyntaxValue>, renames: &HashMap<String, String>) -> SyntaxValue {
    let mut result = vec![items[0].clone()];
    let mut body_renames = renames.clone();

    if let SyntaxValue::List(bindings) = &items[1] {
        let mut new_bindings = Vec::new();
        let mut i = 0;
        while i + 1 < bindings.len() {
            if let SyntaxValue::Sym(name) = &bindings[i] {
                let new_name = gensym(name);
                body_renames.insert(name.clone(), new_name.clone());
                new_bindings.push(SyntaxValue::Sym(new_name));
                new_bindings.push(rename_syntax(bindings[i + 1].clone(), renames));
            } else {
                new_bindings.push(rename_syntax(bindings[i].clone(), renames));
                new_bindings.push(rename_syntax(bindings[i + 1].clone(), renames));
            }
            i += 2;
        }
        result.push(SyntaxValue::List(new_bindings));
    } else {
        result.push(rename_syntax(items[1].clone(), renames));
    }

    for item in items.into_iter().skip(2) {
        result.push(rename_syntax(item, &body_renames));
    }

    SyntaxValue::List(result)
}

fn rename_syntax_fn(items: Vec<SyntaxValue>, renames: &HashMap<String, String>) -> SyntaxValue {
    let mut result = vec![items[0].clone()];
    let mut body_renames = renames.clone();

    if let SyntaxValue::List(params) = &items[1] {
        let new_params: Vec<SyntaxValue> = params
            .iter()
            .map(|p| {
                if let SyntaxValue::Sym(name) = p {
                    let new_name = gensym(name);
                    body_renames.insert(name.clone(), new_name.clone());
                    SyntaxValue::Sym(new_name)
                } else {
                    rename_syntax(p.clone(), renames)
                }
            })
            .collect();
        result.push(SyntaxValue::List(new_params));
    } else {
        result.push(rename_syntax(items[1].clone(), renames));
    }

    for item in items.into_iter().skip(2) {
        result.push(rename_syntax(item, &body_renames));
    }

    SyntaxValue::List(result)
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::source::{FileId, Span};

    fn span(start: u32, end: u32) -> Span {
        Span::new(FileId::new(0), start, end)
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
            assert!(
                bindings[0].name.starts_with("__tmp_"),
                "expected hygienic tmp name, got: {}",
                bindings[0].name
            );
            let tmp_name = &bindings[0].name;
            assert!(matches!(&bindings[0].value, Expr::Symbol(s, _) if s == "a"));

            assert_eq!(body.len(), 1);
            if let Expr::If {
                test,
                then_branch,
                else_branch,
                ..
            } = &body[0]
            {
                assert!(matches!(test.as_ref(), Expr::Symbol(s, _) if s == tmp_name));
                assert!(matches!(then_branch.as_ref(), Expr::Symbol(s, _) if s == tmp_name));
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
            body: vec![Expr::Call {
                func: Box::new(Expr::Symbol("and".into(), span(1, 4))),
                args: vec![
                    Expr::Bool(true, span(5, 9)),
                    Expr::Bool(false, span(10, 15)),
                ],
                span: span(0, 16),
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
                rest_param: None,
                body: vec![Expr::Call {
                    func: Box::new(Expr::Symbol("list".into(), s)),
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
                rest_param: None,
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
                rest_param: None,
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
                rest_param: None,
                body: vec![Expr::Let {
                    bindings: vec![Binding {
                        name: "result".into(),
                        value: Expr::Call {
                            func: Box::new(Expr::Symbol("list".into(), s)),
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
                rest_param: None,
                body: vec![Expr::Call {
                    func: Box::new(Expr::Symbol("list".into(), s)),
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

    #[test]
    fn hygiene_renames_let_bindings() {
        let s = span(0, 1);
        let input = vec![
            TopForm::DefMacro {
                name: "with-temp".into(),
                params: vec![ast::Param {
                    name: "body".into(),
                    type_ann: None,
                    span: s,
                }],
                rest_param: None,
                body: vec![Expr::Call {
                    func: Box::new(Expr::Symbol("list".into(), s)),
                    args: vec![
                        Expr::Quote {
                            expr: Box::new(Expr::Symbol("let".into(), s)),
                            span: s,
                        },
                        Expr::Call {
                            func: Box::new(Expr::Symbol("list".into(), s)),
                            args: vec![
                                Expr::Quote {
                                    expr: Box::new(Expr::Symbol("tmp".into(), s)),
                                    span: s,
                                },
                                Expr::Quote {
                                    expr: Box::new(Expr::Int(0, s)),
                                    span: s,
                                },
                            ],
                            span: s,
                        },
                        Expr::Symbol("body".into(), s),
                    ],
                    span: s,
                }],
                span: s,
            },
            TopForm::Expr(Expr::Call {
                func: Box::new(Expr::Symbol("with-temp".into(), s)),
                args: vec![Expr::Symbol("tmp".into(), s)],
                span: s,
            }),
        ];

        let (result, diags) = expand(input);
        assert!(diags.is_empty(), "{:?}", diags);
        assert_eq!(result.len(), 1);

        if let TopForm::Expr(Expr::Let { bindings, body, .. }) = &result[0] {
            assert_eq!(bindings.len(), 1);
            assert!(
                bindings[0].name.starts_with("__tmp_"),
                "expected renamed binding, got: {}",
                bindings[0].name
            );
            assert_ne!(bindings[0].name, "tmp");
            if let Expr::Symbol(body_name, _) = &body[0] {
                assert_eq!(
                    body_name, "tmp",
                    "call-site reference should keep original name"
                );
            } else {
                panic!("expected symbol in body");
            }
        } else {
            panic!("expected let expression");
        }
    }

    #[test]
    fn hygiene_renames_lambda_params() {
        let s = span(0, 1);
        let input = vec![
            TopForm::DefMacro {
                name: "make-fn".into(),
                params: vec![],
                rest_param: None,
                body: vec![Expr::Call {
                    func: Box::new(Expr::Symbol("list".into(), s)),
                    args: vec![
                        Expr::Quote {
                            expr: Box::new(Expr::Symbol("fn".into(), s)),
                            span: s,
                        },
                        Expr::Call {
                            func: Box::new(Expr::Symbol("list".into(), s)),
                            args: vec![Expr::Quote {
                                expr: Box::new(Expr::Symbol("x".into(), s)),
                                span: s,
                            }],
                            span: s,
                        },
                        Expr::Quote {
                            expr: Box::new(Expr::Symbol("x".into(), s)),
                            span: s,
                        },
                    ],
                    span: s,
                }],
                span: s,
            },
            TopForm::Expr(Expr::Call {
                func: Box::new(Expr::Symbol("make-fn".into(), s)),
                args: vec![],
                span: s,
            }),
        ];

        let (result, diags) = expand(input);
        assert!(diags.is_empty(), "{:?}", diags);
        assert_eq!(result.len(), 1);

        if let TopForm::Expr(Expr::Lambda { params, body, .. }) = &result[0] {
            assert_eq!(params.len(), 1);
            assert!(
                params[0].name.starts_with("__x_"),
                "expected renamed param, got: {}",
                params[0].name
            );
            if let Expr::Symbol(body_name, _) = &body[0] {
                assert_eq!(
                    body_name, &params[0].name,
                    "reference in body should match renamed param"
                );
            } else {
                panic!("expected symbol in body");
            }
        } else {
            panic!("expected lambda, got: {:?}", result[0]);
        }
    }

    #[test]
    fn variadic_defmacro_collects_rest_args() {
        let s = span(0, 1);
        let input = vec![
            TopForm::DefMacro {
                name: "wrap-all".into(),
                params: vec![],
                rest_param: Some("args".into()),
                body: vec![Expr::Call {
                    func: Box::new(Expr::Symbol("cons".into(), s)),
                    args: vec![
                        Expr::Quote {
                            expr: Box::new(Expr::Symbol("println".into(), s)),
                            span: s,
                        },
                        Expr::Symbol("args".into(), s),
                    ],
                    span: s,
                }],
                span: s,
            },
            TopForm::Expr(Expr::Call {
                func: Box::new(Expr::Symbol("wrap-all".into(), s)),
                args: vec![Expr::String("a".into(), s), Expr::String("b".into(), s)],
                span: s,
            }),
        ];

        let (result, diags) = expand(input);
        assert!(diags.is_empty(), "{:?}", diags);
        assert_eq!(result.len(), 1);
        if let TopForm::Expr(Expr::Call { func, args, .. }) = &result[0] {
            assert!(matches!(func.as_ref(), Expr::Symbol(s, _) if s == "println"));
            assert_eq!(args.len(), 2);
            assert!(matches!(&args[0], Expr::String(s, _) if s == "a"));
            assert!(matches!(&args[1], Expr::String(s, _) if s == "b"));
        } else {
            panic!("expected call expression, got: {:?}", result[0]);
        }
    }

    #[test]
    fn variadic_defmacro_empty_rest() {
        let s = span(0, 1);
        let input = vec![
            TopForm::DefMacro {
                name: "id-list".into(),
                params: vec![],
                rest_param: Some("args".into()),
                body: vec![Expr::Symbol("args".into(), s)],
                span: s,
            },
            TopForm::Expr(Expr::Call {
                func: Box::new(Expr::Symbol("id-list".into(), s)),
                args: vec![],
                span: s,
            }),
        ];

        let (result, diags) = expand(input);
        assert!(diags.is_empty(), "{:?}", diags);
        assert_eq!(result.len(), 1);
        assert!(matches!(&result[0], TopForm::Expr(Expr::Nil(_))));
    }

    #[test]
    fn variadic_defmacro_with_fixed_params() {
        let s = span(0, 1);
        let input = vec![
            TopForm::DefMacro {
                name: "if-else".into(),
                params: vec![
                    ast::Param {
                        name: "test".into(),
                        type_ann: None,
                        span: s,
                    },
                    ast::Param {
                        name: "then-val".into(),
                        type_ann: None,
                        span: s,
                    },
                ],
                rest_param: Some("rest".into()),
                body: vec![Expr::Call {
                    func: Box::new(Expr::Symbol("list".into(), s)),
                    args: vec![
                        Expr::Quote {
                            expr: Box::new(Expr::Symbol("if".into(), s)),
                            span: s,
                        },
                        Expr::Symbol("test".into(), s),
                        Expr::Symbol("then-val".into(), s),
                        Expr::Quote {
                            expr: Box::new(Expr::Nil(s)),
                            span: s,
                        },
                    ],
                    span: s,
                }],
                span: s,
            },
            TopForm::Expr(Expr::Call {
                func: Box::new(Expr::Symbol("if-else".into(), s)),
                args: vec![
                    Expr::Bool(true, s),
                    Expr::Int(1, s),
                    Expr::Int(2, s),
                    Expr::Int(3, s),
                ],
                span: s,
            }),
        ];

        let (result, diags) = expand(input);
        assert!(diags.is_empty(), "{:?}", diags);
        assert_eq!(result.len(), 1);
        assert!(matches!(&result[0], TopForm::Expr(Expr::If { .. })));
    }

    #[test]
    fn variadic_defmacro_too_few_args() {
        let s = span(0, 1);
        let input = vec![
            TopForm::DefMacro {
                name: "needs-two".into(),
                params: vec![
                    ast::Param {
                        name: "a".into(),
                        type_ann: None,
                        span: s,
                    },
                    ast::Param {
                        name: "b".into(),
                        type_ann: None,
                        span: s,
                    },
                ],
                rest_param: Some("rest".into()),
                body: vec![Expr::Quote {
                    expr: Box::new(Expr::Nil(s)),
                    span: s,
                }],
                span: s,
            },
            TopForm::Expr(Expr::Call {
                func: Box::new(Expr::Symbol("needs-two".into(), s)),
                args: vec![Expr::Int(1, s)],
                span: s,
            }),
        ];

        let (_, diags) = expand(input);
        assert!(!diags.is_empty());
    }

    #[test]
    fn empty_helper() {
        let s = span(0, 1);
        let input = vec![
            TopForm::DefMacro {
                name: "is-empty".into(),
                params: vec![ast::Param {
                    name: "xs".into(),
                    type_ann: None,
                    span: s,
                }],
                rest_param: None,
                body: vec![Expr::Call {
                    func: Box::new(Expr::Symbol("list".into(), s)),
                    args: vec![
                        Expr::Quote {
                            expr: Box::new(Expr::Symbol("if".into(), s)),
                            span: s,
                        },
                        Expr::Call {
                            func: Box::new(Expr::Symbol("empty?".into(), s)),
                            args: vec![Expr::Symbol("xs".into(), s)],
                            span: s,
                        },
                        Expr::Quote {
                            expr: Box::new(Expr::Bool(true, s)),
                            span: s,
                        },
                        Expr::Quote {
                            expr: Box::new(Expr::Bool(false, s)),
                            span: s,
                        },
                    ],
                    span: s,
                }],
                span: s,
            },
            TopForm::Expr(Expr::Call {
                func: Box::new(Expr::Symbol("is-empty".into(), s)),
                args: vec![Expr::Quote {
                    expr: Box::new(Expr::Call {
                        func: Box::new(Expr::Symbol("x".into(), s)),
                        args: vec![],
                        span: s,
                    }),
                    span: s,
                }],
                span: s,
            }),
        ];

        let (result, diags) = expand(input);
        assert!(diags.is_empty(), "{:?}", diags);
        if let TopForm::Expr(Expr::If {
            test,
            then_branch,
            else_branch,
            ..
        }) = &result[0]
        {
            assert!(matches!(test.as_ref(), Expr::Bool(false, _)));
            assert!(matches!(then_branch.as_ref(), Expr::Bool(true, _)));
            assert!(matches!(else_branch.as_ref(), Expr::Bool(false, _)));
        } else {
            panic!("expected if expression");
        }
    }

    #[test]
    fn keyword_helper() {
        let s = span(0, 1);
        let input = vec![
            TopForm::DefMacro {
                name: "check-kw".into(),
                params: vec![ast::Param {
                    name: "x".into(),
                    type_ann: None,
                    span: s,
                }],
                rest_param: None,
                body: vec![Expr::Call {
                    func: Box::new(Expr::Symbol("list".into(), s)),
                    args: vec![
                        Expr::Quote {
                            expr: Box::new(Expr::Symbol("if".into(), s)),
                            span: s,
                        },
                        Expr::Call {
                            func: Box::new(Expr::Symbol("keyword?".into(), s)),
                            args: vec![Expr::Symbol("x".into(), s)],
                            span: s,
                        },
                        Expr::Quote {
                            expr: Box::new(Expr::Bool(true, s)),
                            span: s,
                        },
                        Expr::Quote {
                            expr: Box::new(Expr::Bool(false, s)),
                            span: s,
                        },
                    ],
                    span: s,
                }],
                span: s,
            },
            TopForm::Expr(Expr::Call {
                func: Box::new(Expr::Symbol("check-kw".into(), s)),
                args: vec![Expr::Keyword("else".into(), s)],
                span: s,
            }),
        ];

        let (result, diags) = expand(input);
        assert!(diags.is_empty(), "{:?}", diags);
        if let TopForm::Expr(Expr::If { test, .. }) = &result[0] {
            assert!(matches!(test.as_ref(), Expr::Bool(true, _)));
        } else {
            panic!("expected if expression");
        }
    }

    #[test]
    fn parser_defmacro_rest_param() {
        let source = "(defmacro my-list [& items] items)";
        let file = FileId::new(0);
        let (tokens, _) = lexer::lex(source, file);
        let (forms, diags) = parser::parse(&tokens);
        assert!(diags.is_empty(), "{:?}", diags);
        assert_eq!(forms.len(), 1);
        if let TopForm::DefMacro {
            name,
            params,
            rest_param,
            ..
        } = &forms[0]
        {
            assert_eq!(name, "my-list");
            assert!(params.is_empty());
            assert_eq!(rest_param.as_deref(), Some("items"));
        } else {
            panic!("expected DefMacro");
        }
    }

    #[test]
    fn parser_defmacro_fixed_and_rest_params() {
        let source = "(defmacro my-macro [a b & rest] (list a b))";
        let file = FileId::new(0);
        let (tokens, _) = lexer::lex(source, file);
        let (forms, diags) = parser::parse(&tokens);
        assert!(diags.is_empty(), "{:?}", diags);
        if let TopForm::DefMacro {
            params, rest_param, ..
        } = &forms[0]
        {
            assert_eq!(params.len(), 2);
            assert_eq!(params[0].name, "a");
            assert_eq!(params[1].name, "b");
            assert_eq!(rest_param.as_deref(), Some("rest"));
        } else {
            panic!("expected DefMacro");
        }
    }
}
