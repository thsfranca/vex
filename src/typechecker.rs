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

            ast::Expr::Let {
                bindings,
                body,
                span,
            } => self.check_let(bindings, body, *span),
            ast::Expr::Cond {
                clauses,
                else_body,
                span,
            } => self.check_cond(clauses, else_body.as_deref(), *span),
            ast::Expr::Lambda {
                params,
                return_type,
                body,
                span,
            } => self.check_lambda(params, return_type.as_ref(), body, *span),
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

    fn resolve_type(&mut self, type_expr: &ast::TypeExpr) -> Option<VexType> {
        match type_expr {
            ast::TypeExpr::Named { name, span } => match name.as_str() {
                "Int" => Some(VexType::Int),
                "Float" => Some(VexType::Float),
                "Bool" => Some(VexType::Bool),
                "String" => Some(VexType::String),
                "Unit" => Some(VexType::Unit),
                _ => {
                    self.diagnostics
                        .push(Diagnostic::error(format!("unknown type '{}'", name), *span));
                    None
                }
            },
            ast::TypeExpr::Function { params, ret, .. } => {
                let mut param_types = Vec::new();
                for p in params {
                    param_types.push(self.resolve_type(p)?);
                }
                let ret_type = self.resolve_type(ret)?;
                Some(VexType::Fn {
                    params: param_types,
                    ret: Box::new(ret_type),
                })
            }
            ast::TypeExpr::Applied { name, span, .. } => {
                self.diagnostics.push(Diagnostic::error(
                    format!("applied type '{}' not supported in MVP", name),
                    *span,
                ));
                None
            }
        }
    }

    fn check_body(&mut self, body: &[ast::Expr]) -> Option<Vec<hir::Expr>> {
        let mut checked = Vec::new();
        for expr in body {
            checked.push(self.check_expr(expr)?);
        }
        Some(checked)
    }

    fn check_let(
        &mut self,
        bindings: &[ast::Binding],
        body: &[ast::Expr],
        span: Span,
    ) -> Option<hir::Expr> {
        self.env.push_scope();

        let mut checked_bindings = Vec::new();
        for binding in bindings {
            let checked_value = self.check_expr(&binding.value)?;
            let ty = checked_value.ty().clone();
            self.env.define(binding.name.clone(), ty.clone());
            checked_bindings.push(hir::Binding {
                name: binding.name.clone(),
                ty,
                value: checked_value,
                span: binding.span,
            });
        }

        let checked_body = self.check_body(body)?;
        let ty = checked_body
            .last()
            .map(|e| e.ty().clone())
            .unwrap_or(VexType::Unit);

        self.env.pop_scope();

        Some(hir::Expr::Let {
            bindings: checked_bindings,
            body: checked_body,
            span,
            ty,
        })
    }

    fn check_cond(
        &mut self,
        clauses: &[ast::CondClause],
        else_body: Option<&ast::Expr>,
        span: Span,
    ) -> Option<hir::Expr> {
        let else_expr = if let Some(e) = else_body {
            self.check_expr(e)?
        } else {
            hir::Expr::Nil(span)
        };

        let mut result = else_expr;

        for clause in clauses.iter().rev() {
            let checked_test = self.check_expr(&clause.test)?;

            if checked_test.ty() != &VexType::Bool {
                self.diagnostics.push(Diagnostic::error(
                    format!("cond test must be Bool, found {}", checked_test.ty()),
                    checked_test.span(),
                ));
                return None;
            }

            let checked_value = self.check_expr(&clause.value)?;

            if checked_value.ty() != result.ty() {
                self.diagnostics.push(Diagnostic::error(
                    format!(
                        "cond branches have different types: {} and {}",
                        checked_value.ty(),
                        result.ty()
                    ),
                    clause.span,
                ));
                return None;
            }

            let ty = checked_value.ty().clone();

            result = hir::Expr::If {
                test: Box::new(checked_test),
                then_branch: Box::new(checked_value),
                else_branch: Box::new(result),
                span: clause.span,
                ty,
            };
        }

        Some(result)
    }

    fn check_lambda(
        &mut self,
        params: &[ast::Param],
        return_type: Option<&ast::TypeExpr>,
        body: &[ast::Expr],
        span: Span,
    ) -> Option<hir::Expr> {
        let mut checked_params = Vec::new();
        let mut param_types = Vec::new();

        for param in params {
            let ty = if let Some(type_ann) = &param.type_ann {
                self.resolve_type(type_ann)?
            } else {
                self.diagnostics.push(Diagnostic::error(
                    format!("parameter '{}' requires a type annotation", param.name),
                    param.span,
                ));
                return None;
            };
            param_types.push(ty.clone());
            checked_params.push(hir::Param {
                name: param.name.clone(),
                ty,
                span: param.span,
            });
        }

        self.env.push_scope();
        for param in &checked_params {
            self.env.define(param.name.clone(), param.ty.clone());
        }

        let checked_body = self.check_body(body)?;
        let body_ty = checked_body
            .last()
            .map(|e| e.ty().clone())
            .unwrap_or(VexType::Unit);

        self.env.pop_scope();

        let ret_ty = if let Some(ret_ann) = return_type {
            let declared = self.resolve_type(ret_ann)?;
            if declared != body_ty {
                self.diagnostics.push(Diagnostic::error(
                    format!(
                        "declared return type {} doesn't match body type {}",
                        declared, body_ty
                    ),
                    span,
                ));
                return None;
            }
            declared
        } else {
            body_ty
        };

        let fn_ty = VexType::Fn {
            params: param_types,
            ret: Box::new(ret_ty.clone()),
        };

        Some(hir::Expr::Lambda {
            params: checked_params,
            return_type: ret_ty,
            body: checked_body,
            span,
            ty: fn_ty,
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

    #[test]
    fn let_simple() {
        let (module, diags) = check_source("(let [x 42] x)");
        assert!(diags.is_empty());
        if let hir::TopForm::Expr(expr) = &module.top_forms[0] {
            assert_eq!(expr.ty(), &VexType::Int);
        } else {
            panic!("expected expression");
        }
    }

    #[test]
    fn let_multiple_bindings() {
        let (module, diags) = check_source("(let [x 1 y 2] (+ x y))");
        assert!(diags.is_empty());
        if let hir::TopForm::Expr(expr) = &module.top_forms[0] {
            assert_eq!(expr.ty(), &VexType::Int);
        } else {
            panic!("expected expression");
        }
    }

    #[test]
    fn let_scoping() {
        let (_, diags) = check_source("(let [x 1] x) x");
        assert_eq!(diags.len(), 1);
        assert!(diags[0].message.contains("undefined symbol"));
    }

    #[test]
    fn let_body_type_is_last() {
        let (module, diags) = check_source(r#"(let [x 1] (println "hi") x)"#);
        assert!(diags.is_empty());
        if let hir::TopForm::Expr(expr) = &module.top_forms[0] {
            assert_eq!(expr.ty(), &VexType::Int);
        } else {
            panic!("expected expression");
        }
    }

    #[test]
    fn cond_basic() {
        let (module, diags) = check_source("(cond (<= 1 2) 10 :else 20)");
        assert!(diags.is_empty());
        if let hir::TopForm::Expr(expr) = &module.top_forms[0] {
            assert_eq!(expr.ty(), &VexType::Int);
            assert!(matches!(expr, hir::Expr::If { .. }));
        } else {
            panic!("expected expression");
        }
    }

    #[test]
    fn cond_multiple_clauses() {
        let (module, diags) = check_source("(cond (< 1 0) 1 (= 1 1) 2 :else 3)");
        assert!(diags.is_empty());
        if let hir::TopForm::Expr(expr) = &module.top_forms[0] {
            assert_eq!(expr.ty(), &VexType::Int);
        } else {
            panic!("expected expression");
        }
    }

    #[test]
    fn cond_no_else() {
        let (module, diags) = check_source("(cond)");
        assert!(diags.is_empty());
        if let hir::TopForm::Expr(expr) = &module.top_forms[0] {
            assert_eq!(expr.ty(), &VexType::Unit);
        } else {
            panic!("expected expression");
        }
    }

    #[test]
    fn error_cond_non_bool_test() {
        let (_, diags) = check_source("(cond 42 1 :else 2)");
        assert_eq!(diags.len(), 1);
        assert!(diags[0].message.contains("must be Bool"));
    }

    #[test]
    fn lambda_simple() {
        let (module, diags) = check_source("(fn [x : Int] (+ x 1))");
        assert!(diags.is_empty());
        if let hir::TopForm::Expr(expr) = &module.top_forms[0] {
            assert_eq!(
                expr.ty(),
                &VexType::Fn {
                    params: vec![VexType::Int],
                    ret: Box::new(VexType::Int),
                }
            );
        } else {
            panic!("expected expression");
        }
    }

    #[test]
    fn lambda_with_return_type() {
        let (module, diags) = check_source("(fn [x : Int] : Int (+ x 1))");
        assert!(diags.is_empty());
        if let hir::TopForm::Expr(expr) = &module.top_forms[0] {
            assert_eq!(
                expr.ty(),
                &VexType::Fn {
                    params: vec![VexType::Int],
                    ret: Box::new(VexType::Int),
                }
            );
        } else {
            panic!("expected expression");
        }
    }

    #[test]
    fn lambda_no_params() {
        let (module, diags) = check_source("(fn [] 42)");
        assert!(diags.is_empty());
        if let hir::TopForm::Expr(expr) = &module.top_forms[0] {
            assert_eq!(
                expr.ty(),
                &VexType::Fn {
                    params: vec![],
                    ret: Box::new(VexType::Int),
                }
            );
        } else {
            panic!("expected expression");
        }
    }

    #[test]
    fn lambda_call() {
        let (module, diags) = check_source("((fn [x : Int] (+ x 1)) 5)");
        assert!(diags.is_empty());
        if let hir::TopForm::Expr(expr) = &module.top_forms[0] {
            assert_eq!(expr.ty(), &VexType::Int);
        } else {
            panic!("expected expression");
        }
    }

    #[test]
    fn error_lambda_missing_type_ann() {
        let (_, diags) = check_source("(fn [x] x)");
        assert_eq!(diags.len(), 1);
        assert!(diags[0].message.contains("requires a type annotation"));
    }

    #[test]
    fn error_lambda_return_type_mismatch() {
        let (_, diags) = check_source(r#"(fn [x : Int] : String (+ x 1))"#);
        assert_eq!(diags.len(), 1);
        assert!(diags[0].message.contains("doesn't match"));
    }
}
