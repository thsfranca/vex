use crate::ast;
use crate::builtins;
use crate::diagnostics::Diagnostic;
use crate::hir;
use crate::source::Span;
use crate::types::{TypeEnv, VexType};

struct Checker {
    env: TypeEnv,
    diagnostics: Vec<Diagnostic>,
}

impl Checker {
    fn new() -> Self {
        let mut env = TypeEnv::new();
        for builtin in builtins::all_builtins() {
            env.define(builtin.name.to_string(), builtin.ty);
        }
        Self {
            env,
            diagnostics: Vec::new(),
        }
    }

    fn check_expr(&mut self, expr: &ast::Expr) -> Option<hir::Expr> {
        match expr {
            ast::Expr::Int(n, span) => Some(hir::Expr::Int(*n, *span)),
            ast::Expr::Float(f, span) => Some(hir::Expr::Float(*f, *span)),
            ast::Expr::String(s, span) => Some(hir::Expr::String(s.clone(), *span)),
            ast::Expr::Bool(b, span) => Some(hir::Expr::Bool(*b, *span)),
            ast::Expr::Nil(span) => Some(hir::Expr::Nil(*span)),

            ast::Expr::Symbol(name, span) => self.check_symbol(name, *span),
            ast::Expr::Keyword(name, span) => {
                self.diagnostics.push(Diagnostic::error(
                    format!("unexpected keyword ':{}'", name),
                    *span,
                ));
                None
            }

            ast::Expr::Call { func, args, span } => self.check_call(func, args, *span),
            ast::Expr::If {
                test,
                then_branch,
                else_branch,
                span,
            } => self.check_if(test, then_branch, else_branch, *span),

            ast::Expr::Let { .. } | ast::Expr::Cond { .. } | ast::Expr::Lambda { .. } => {
                self.diagnostics.push(Diagnostic::error(
                    "not yet implemented in type checker",
                    expr.span(),
                ));
                None
            }
        }
    }

    fn check_symbol(&mut self, name: &str, span: Span) -> Option<hir::Expr> {
        if let Some(ty) = self.env.lookup(name) {
            Some(hir::Expr::Var {
                name: name.to_string(),
                span,
                ty: ty.clone(),
            })
        } else {
            self.diagnostics.push(Diagnostic::error(
                format!("undefined symbol '{}'", name),
                span,
            ));
            None
        }
    }

    fn check_call(
        &mut self,
        func: &ast::Expr,
        args: &[ast::Expr],
        span: Span,
    ) -> Option<hir::Expr> {
        let checked_func = self.check_expr(func)?;
        let mut checked_args = Vec::new();
        for arg in args {
            checked_args.push(self.check_expr(arg)?);
        }

        let func_name = if let ast::Expr::Symbol(name, _) = func {
            Some(name.as_str())
        } else {
            None
        };

        let func_ty = self.resolve_call_type(func_name, &checked_func, &checked_args, span)?;

        let (ret_ty, checked_func) = match func_ty {
            VexType::Fn { params, ret } => {
                if let Some(name) = func_name
                    && builtins::lookup(name).is_some_and(|b| b.variadic)
                {
                    (*ret, checked_func)
                } else {
                    if params.len() != checked_args.len() {
                        self.diagnostics.push(Diagnostic::error(
                            format!(
                                "expected {} argument(s), found {}",
                                params.len(),
                                checked_args.len()
                            ),
                            span,
                        ));
                        return None;
                    }
                    for (i, (param_ty, arg)) in params.iter().zip(checked_args.iter()).enumerate() {
                        let arg_ty = arg.ty();
                        if param_ty != arg_ty {
                            self.diagnostics.push(Diagnostic::error(
                                format!(
                                    "argument {} has type {}, expected {}",
                                    i + 1,
                                    arg_ty,
                                    param_ty
                                ),
                                arg.span(),
                            ));
                            return None;
                        }
                    }
                    (*ret, checked_func)
                }
            }
            _ => {
                self.diagnostics.push(Diagnostic::error(
                    format!("cannot call value of type {}", checked_func.ty()),
                    checked_func.span(),
                ));
                return None;
            }
        };

        Some(hir::Expr::Call {
            func: Box::new(checked_func),
            args: checked_args,
            span,
            ty: ret_ty,
        })
    }

    fn resolve_call_type(
        &mut self,
        func_name: Option<&str>,
        checked_func: &hir::Expr,
        checked_args: &[hir::Expr],
        _span: Span,
    ) -> Option<VexType> {
        if let Some(name) = func_name
            && let Some(builtin) = builtins::lookup(name)
            && !builtin.variadic
            && self.is_numeric_op(name)
            && !checked_args.is_empty()
        {
            let first_arg_ty = checked_args[0].ty();
            if first_arg_ty == &VexType::Float {
                let ret = if let VexType::Fn { ret, .. } = &builtin.ty {
                    if **ret == VexType::Int {
                        VexType::Float
                    } else {
                        (**ret).clone()
                    }
                } else {
                    return Some(checked_func.ty().clone());
                };
                return Some(VexType::Fn {
                    params: vec![VexType::Float; checked_args.len()],
                    ret: Box::new(ret),
                });
            }
        }

        Some(checked_func.ty().clone())
    }

    fn is_numeric_op(&self, name: &str) -> bool {
        matches!(
            name,
            "+" | "-" | "*" | "/" | "mod" | "<" | ">" | "<=" | ">=" | "=" | "!="
        )
    }

    fn check_if(
        &mut self,
        test: &ast::Expr,
        then_branch: &ast::Expr,
        else_branch: &ast::Expr,
        span: Span,
    ) -> Option<hir::Expr> {
        let checked_test = self.check_expr(test)?;

        if checked_test.ty() != &VexType::Bool {
            self.diagnostics.push(Diagnostic::error(
                format!("if condition must be Bool, found {}", checked_test.ty()),
                checked_test.span(),
            ));
            return None;
        }

        let checked_then = self.check_expr(then_branch)?;
        let checked_else = self.check_expr(else_branch)?;

        if checked_then.ty() != checked_else.ty() {
            self.diagnostics.push(Diagnostic::error(
                format!(
                    "if branches have different types: {} and {}",
                    checked_then.ty(),
                    checked_else.ty()
                ),
                span,
            ));
            return None;
        }

        let ty = checked_then.ty().clone();

        Some(hir::Expr::If {
            test: Box::new(checked_test),
            then_branch: Box::new(checked_then),
            else_branch: Box::new(checked_else),
            span,
            ty,
        })
    }

    fn check_top_form(&mut self, form: &ast::TopForm) -> Option<hir::TopForm> {
        match form {
            ast::TopForm::Expr(expr) => {
                let checked = self.check_expr(expr)?;
                Some(hir::TopForm::Expr(checked))
            }
            ast::TopForm::Defn { span, .. } | ast::TopForm::Def { span, .. } => {
                self.diagnostics.push(Diagnostic::error(
                    "not yet implemented in type checker",
                    *span,
                ));
                None
            }
        }
    }
}

pub fn check(program: &[ast::TopForm]) -> (hir::Module, Vec<Diagnostic>) {
    let mut checker = Checker::new();
    let mut top_forms = Vec::new();

    for form in program {
        if let Some(checked) = checker.check_top_form(form) {
            top_forms.push(checked);
        }
    }

    (hir::Module { top_forms }, checker.diagnostics)
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::lexer::lex;
    use crate::parser::parse;
    use crate::source::FileId;

    fn check_source(source: &str) -> (hir::Module, Vec<Diagnostic>) {
        let (tokens, _) = lex(source, FileId::new(0));
        let (ast, _) = parse(&tokens);
        check(&ast)
    }

    #[test]
    fn literal_int() {
        let (module, diags) = check_source("42");
        assert!(diags.is_empty());
        assert_eq!(module.top_forms.len(), 1);
        if let hir::TopForm::Expr(expr) = &module.top_forms[0] {
            assert_eq!(expr.ty(), &VexType::Int);
        } else {
            panic!("expected expression");
        }
    }

    #[test]
    fn literal_float() {
        let (module, diags) = check_source("3.14");
        assert!(diags.is_empty());
        if let hir::TopForm::Expr(expr) = &module.top_forms[0] {
            assert_eq!(expr.ty(), &VexType::Float);
        } else {
            panic!("expected expression");
        }
    }

    #[test]
    fn literal_string() {
        let (module, diags) = check_source(r#""hello""#);
        assert!(diags.is_empty());
        if let hir::TopForm::Expr(expr) = &module.top_forms[0] {
            assert_eq!(expr.ty(), &VexType::String);
        } else {
            panic!("expected expression");
        }
    }

    #[test]
    fn literal_bool() {
        let (module, diags) = check_source("true");
        assert!(diags.is_empty());
        if let hir::TopForm::Expr(expr) = &module.top_forms[0] {
            assert_eq!(expr.ty(), &VexType::Bool);
        } else {
            panic!("expected expression");
        }
    }

    #[test]
    fn literal_nil() {
        let (module, diags) = check_source("nil");
        assert!(diags.is_empty());
        if let hir::TopForm::Expr(expr) = &module.top_forms[0] {
            assert_eq!(expr.ty(), &VexType::Unit);
        } else {
            panic!("expected expression");
        }
    }

    #[test]
    fn builtin_symbol() {
        let (module, diags) = check_source("+");
        assert!(diags.is_empty());
        if let hir::TopForm::Expr(hir::Expr::Var { name, ty, .. }) = &module.top_forms[0] {
            assert_eq!(name, "+");
            assert!(matches!(ty, VexType::Fn { .. }));
        } else {
            panic!("expected var");
        }
    }

    #[test]
    fn error_undefined_symbol() {
        let (_, diags) = check_source("undefined_var");
        assert_eq!(diags.len(), 1);
        assert!(diags[0].message.contains("undefined symbol"));
    }

    #[test]
    fn call_println() {
        let (module, diags) = check_source(r#"(println "hi")"#);
        assert!(diags.is_empty());
        if let hir::TopForm::Expr(expr) = &module.top_forms[0] {
            assert_eq!(expr.ty(), &VexType::Unit);
        } else {
            panic!("expected expression");
        }
    }

    #[test]
    fn call_arithmetic() {
        let (module, diags) = check_source("(+ 1 2)");
        assert!(diags.is_empty());
        if let hir::TopForm::Expr(expr) = &module.top_forms[0] {
            assert_eq!(expr.ty(), &VexType::Int);
        } else {
            panic!("expected expression");
        }
    }

    #[test]
    fn call_nested_arithmetic() {
        let (module, diags) = check_source("(+ (* 2 3) (- 10 1))");
        assert!(diags.is_empty());
        if let hir::TopForm::Expr(expr) = &module.top_forms[0] {
            assert_eq!(expr.ty(), &VexType::Int);
        } else {
            panic!("expected expression");
        }
    }

    #[test]
    fn call_comparison() {
        let (module, diags) = check_source("(<= 1 2)");
        assert!(diags.is_empty());
        if let hir::TopForm::Expr(expr) = &module.top_forms[0] {
            assert_eq!(expr.ty(), &VexType::Bool);
        } else {
            panic!("expected expression");
        }
    }

    #[test]
    fn call_float_arithmetic() {
        let (module, diags) = check_source("(+ 1.0 2.0)");
        assert!(diags.is_empty());
        if let hir::TopForm::Expr(expr) = &module.top_forms[0] {
            assert_eq!(expr.ty(), &VexType::Float);
        } else {
            panic!("expected expression");
        }
    }

    #[test]
    fn call_float_comparison() {
        let (module, diags) = check_source("(< 1.0 2.0)");
        assert!(diags.is_empty());
        if let hir::TopForm::Expr(expr) = &module.top_forms[0] {
            assert_eq!(expr.ty(), &VexType::Bool);
        } else {
            panic!("expected expression");
        }
    }

    #[test]
    fn call_str_variadic() {
        let (module, diags) = check_source(r#"(str 1 "x" true)"#);
        assert!(diags.is_empty());
        if let hir::TopForm::Expr(expr) = &module.top_forms[0] {
            assert_eq!(expr.ty(), &VexType::String);
        } else {
            panic!("expected expression");
        }
    }

    #[test]
    fn error_wrong_arg_type() {
        let (_, diags) = check_source("(+ 1 true)");
        assert_eq!(diags.len(), 1);
        assert!(diags[0].message.contains("argument 2"));
    }

    #[test]
    fn error_wrong_arity() {
        let (_, diags) = check_source("(+ 1)");
        assert_eq!(diags.len(), 1);
        assert!(diags[0].message.contains("expected 2"));
    }

    #[test]
    fn error_call_non_function() {
        let (_, diags) = check_source("(42 1 2)");
        assert_eq!(diags.len(), 1);
        assert!(diags[0].message.contains("cannot call"));
    }

    #[test]
    fn if_expression() {
        let (module, diags) = check_source("(if true 1 2)");
        assert!(diags.is_empty());
        if let hir::TopForm::Expr(expr) = &module.top_forms[0] {
            assert_eq!(expr.ty(), &VexType::Int);
        } else {
            panic!("expected expression");
        }
    }

    #[test]
    fn if_nested() {
        let (module, diags) = check_source("(if (<= 1 2) (+ 1 1) (- 3 1))");
        assert!(diags.is_empty());
        if let hir::TopForm::Expr(expr) = &module.top_forms[0] {
            assert_eq!(expr.ty(), &VexType::Int);
        } else {
            panic!("expected expression");
        }
    }

    #[test]
    fn error_if_non_bool_test() {
        let (_, diags) = check_source("(if 42 1 2)");
        assert_eq!(diags.len(), 1);
        assert!(diags[0].message.contains("must be Bool"));
    }

    #[test]
    fn error_if_branch_mismatch() {
        let (_, diags) = check_source(r#"(if true 1 "hello")"#);
        assert_eq!(diags.len(), 1);
        assert!(diags[0].message.contains("different types"));
    }
}
