use crate::ast;
use crate::builtins;
use crate::diagnostics::Diagnostic;
use crate::hir;
use crate::source::Span;
use crate::types::{RecordField, TypeEnv, VexType};

struct Checker {
    env: TypeEnv,
    type_defs: std::collections::HashMap<String, VexType>,
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
            type_defs: std::collections::HashMap::new(),
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
            ast::Expr::FieldAccess {
                object,
                field,
                span,
            } => self.check_field_access(object, field, *span),
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
        if let ast::Expr::Symbol(name, _) = func
            && let Some(record_ty) = self.type_defs.get(name).cloned()
        {
            return self.check_record_constructor(name, &record_ty, args, span);
        }

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
                    if let Some(ty) = self.type_defs.get(name) {
                        Some(ty.clone())
                    } else {
                        self.diagnostics
                            .push(Diagnostic::error(format!("unknown type '{}'", name), *span));
                        None
                    }
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

    fn check_field_access(
        &mut self,
        object: &ast::Expr,
        field: &str,
        span: Span,
    ) -> Option<hir::Expr> {
        let checked_object = self.check_expr(object)?;
        let object_ty = checked_object.ty().clone();

        match object_ty.field_type(field) {
            Some(field_ty) => {
                let ty = field_ty.clone();
                Some(hir::Expr::FieldAccess {
                    object: Box::new(checked_object),
                    field: field.to_string(),
                    span,
                    ty,
                })
            }
            None => {
                self.diagnostics.push(Diagnostic::error(
                    format!("type {} has no field '{}'", object_ty, field),
                    span,
                ));
                None
            }
        }
    }

    fn check_record_constructor(
        &mut self,
        name: &str,
        record_ty: &VexType,
        args: &[ast::Expr],
        span: Span,
    ) -> Option<hir::Expr> {
        let fields = match record_ty {
            VexType::Record { fields, .. } => fields,
            _ => return None,
        };

        if fields.len() != args.len() {
            self.diagnostics.push(Diagnostic::error(
                format!(
                    "{} has {} field(s), but {} argument(s) were provided",
                    name,
                    fields.len(),
                    args.len()
                ),
                span,
            ));
            return None;
        }

        let mut checked_args = Vec::new();
        for (field, arg) in fields.iter().zip(args.iter()) {
            let checked_arg = self.check_expr(arg)?;
            if checked_arg.ty() != &field.ty {
                self.diagnostics.push(Diagnostic::error(
                    format!(
                        "field '{}' expects type {}, got {}",
                        field.name,
                        field.ty,
                        checked_arg.ty()
                    ),
                    checked_arg.span(),
                ));
                return None;
            }
            checked_args.push(checked_arg);
        }

        Some(hir::Expr::RecordConstructor {
            name: name.to_string(),
            args: checked_args,
            span,
            ty: record_ty.clone(),
        })
    }

    fn check_defn(
        &mut self,
        name: &str,
        params: &[ast::Param],
        return_type: Option<&ast::TypeExpr>,
        body: &[ast::Expr],
        span: Span,
    ) -> Option<hir::TopForm> {
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

        let ret_ty_from_ann = if let Some(ret_ann) = return_type {
            Some(self.resolve_type(ret_ann)?)
        } else {
            None
        };

        let fn_ty = VexType::Fn {
            params: param_types,
            ret: Box::new(ret_ty_from_ann.clone().unwrap_or(VexType::Unit)),
        };
        self.env.define(name.to_string(), fn_ty);

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

        let ret_ty = if let Some(declared) = ret_ty_from_ann {
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
            body_ty.clone()
        };

        let final_fn_ty = VexType::Fn {
            params: checked_params.iter().map(|p| p.ty.clone()).collect(),
            ret: Box::new(ret_ty.clone()),
        };
        self.env.define(name.to_string(), final_fn_ty);

        Some(hir::TopForm::Defn {
            name: name.to_string(),
            params: checked_params,
            return_type: ret_ty,
            body: checked_body,
            span,
        })
    }

    fn check_def(
        &mut self,
        name: &str,
        type_ann: Option<&ast::TypeExpr>,
        value: &ast::Expr,
        span: Span,
    ) -> Option<hir::TopForm> {
        let checked_value = self.check_expr(value)?;
        let value_ty = checked_value.ty().clone();

        let ty = if let Some(ann) = type_ann {
            let declared = self.resolve_type(ann)?;
            if declared != value_ty {
                self.diagnostics.push(Diagnostic::error(
                    format!(
                        "declared type {} doesn't match value type {}",
                        declared, value_ty
                    ),
                    span,
                ));
                return None;
            }
            declared
        } else {
            value_ty
        };

        self.env.define(name.to_string(), ty.clone());

        Some(hir::TopForm::Def {
            name: name.to_string(),
            ty,
            value: checked_value,
            span,
        })
    }

    fn check_deftype(
        &mut self,
        name: &str,
        fields: &[ast::Field],
        span: Span,
    ) -> Option<hir::TopForm> {
        let mut record_fields = Vec::new();
        for field in fields {
            let ty = self.resolve_type(&field.type_expr)?;
            record_fields.push(RecordField {
                name: field.name.clone(),
                ty,
            });
        }

        let record_type = VexType::Record {
            name: name.to_string(),
            fields: record_fields.clone(),
        };

        self.type_defs.insert(name.to_string(), record_type);

        Some(hir::TopForm::Deftype {
            name: name.to_string(),
            fields: record_fields,
            span,
        })
    }

    fn check_top_form(&mut self, form: &ast::TopForm) -> Option<hir::TopForm> {
        match form {
            ast::TopForm::Expr(expr) => {
                let checked = self.check_expr(expr)?;
                Some(hir::TopForm::Expr(checked))
            }
            ast::TopForm::Defn {
                name,
                params,
                return_type,
                body,
                span,
            } => self.check_defn(name, params, return_type.as_ref(), body, *span),
            ast::TopForm::Def {
                name,
                type_ann,
                value,
                span,
            } => self.check_def(name, type_ann.as_ref(), value, *span),
            ast::TopForm::Deftype { name, fields, span } => self.check_deftype(name, fields, *span),
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

    #[test]
    fn defn_simple() {
        let (module, diags) = check_source("(defn add [x : Int y : Int] : Int (+ x y))");
        assert!(diags.is_empty());
        assert_eq!(module.top_forms.len(), 1);
        assert!(matches!(&module.top_forms[0], hir::TopForm::Defn { name, .. } if name == "add"));
    }

    #[test]
    fn defn_no_return_type() {
        let (module, diags) = check_source(r#"(defn greet [name : String] (println name))"#);
        assert!(diags.is_empty());
        if let hir::TopForm::Defn { return_type, .. } = &module.top_forms[0] {
            assert_eq!(return_type, &VexType::Unit);
        } else {
            panic!("expected defn");
        }
    }

    #[test]
    fn defn_recursive() {
        let source = "(defn fib [n : Int] : Int (if (<= n 1) n (+ (fib (- n 1)) (fib (- n 2)))))";
        let (module, diags) = check_source(source);
        assert!(diags.is_empty());
        if let hir::TopForm::Defn {
            name, return_type, ..
        } = &module.top_forms[0]
        {
            assert_eq!(name, "fib");
            assert_eq!(return_type, &VexType::Int);
        } else {
            panic!("expected defn");
        }
    }

    #[test]
    fn defn_available_after_definition() {
        let source = r#"(defn double [x : Int] : Int (* x 2))
                        (double 5)"#;
        let (module, diags) = check_source(source);
        assert!(diags.is_empty());
        assert_eq!(module.top_forms.len(), 2);
        if let hir::TopForm::Expr(expr) = &module.top_forms[1] {
            assert_eq!(expr.ty(), &VexType::Int);
        } else {
            panic!("expected expression");
        }
    }

    #[test]
    fn error_defn_return_type_mismatch() {
        let (_, diags) = check_source(r#"(defn f [] : Int "hello")"#);
        assert_eq!(diags.len(), 1);
        assert!(diags[0].message.contains("doesn't match"));
    }

    #[test]
    fn error_defn_missing_param_type() {
        let (_, diags) = check_source("(defn f [x] x)");
        assert_eq!(diags.len(), 1);
        assert!(diags[0].message.contains("requires a type annotation"));
    }

    #[test]
    fn def_simple() {
        let (module, diags) = check_source("(def x 42)");
        assert!(diags.is_empty());
        if let hir::TopForm::Def { name, ty, .. } = &module.top_forms[0] {
            assert_eq!(name, "x");
            assert_eq!(ty, &VexType::Int);
        } else {
            panic!("expected def");
        }
    }

    #[test]
    fn def_with_type_annotation() {
        let (module, diags) = check_source("(def pi : Float 3.14)");
        assert!(diags.is_empty());
        if let hir::TopForm::Def { name, ty, .. } = &module.top_forms[0] {
            assert_eq!(name, "pi");
            assert_eq!(ty, &VexType::Float);
        } else {
            panic!("expected def");
        }
    }

    #[test]
    fn def_available_after_definition() {
        let source = "(def x 10) (+ x 1)";
        let (module, diags) = check_source(source);
        assert!(diags.is_empty());
        assert_eq!(module.top_forms.len(), 2);
        if let hir::TopForm::Expr(expr) = &module.top_forms[1] {
            assert_eq!(expr.ty(), &VexType::Int);
        } else {
            panic!("expected expression");
        }
    }

    #[test]
    fn error_def_type_mismatch() {
        let (_, diags) = check_source(r#"(def x : Int "hello")"#);
        assert_eq!(diags.len(), 1);
        assert!(diags[0].message.contains("doesn't match"));
    }

    #[test]
    fn integration_hello_world() {
        let source = r#"(defn main [] (println "Hello, World!"))"#;
        let (module, diags) = check_source(source);
        assert!(diags.is_empty());
        assert_eq!(module.top_forms.len(), 1);
        if let hir::TopForm::Defn {
            name, return_type, ..
        } = &module.top_forms[0]
        {
            assert_eq!(name, "main");
            assert_eq!(return_type, &VexType::Unit);
        } else {
            panic!("expected defn");
        }
    }

    #[test]
    fn integration_fibonacci() {
        let source = r#"
            (defn fib [n : Int] : Int
              (if (<= n 1)
                n
                (+ (fib (- n 1)) (fib (- n 2)))))

            (defn main []
              (println (str (fib 10))))
        "#;
        let (module, diags) = check_source(source);
        assert!(diags.is_empty());
        assert_eq!(module.top_forms.len(), 2);
    }

    #[test]
    fn integration_def_and_defn() {
        let source = r#"
            (def greeting : String "Hello")
            (defn greet [name : String] : String
              (str greeting ", " name "!"))
            (defn main []
              (println (greet "World")))
        "#;
        let (module, diags) = check_source(source);
        assert!(diags.is_empty());
        assert_eq!(module.top_forms.len(), 3);
    }

    #[test]
    fn integration_cond_and_let() {
        let source = r#"
            (defn classify [n : Int] : String
              (let [abs_n (if (< n 0) (* n -1) n)]
                (cond
                  (= abs_n 0) "zero"
                  (<= abs_n 10) "small"
                  :else "large")))
        "#;
        let (module, diags) = check_source(source);
        assert!(diags.is_empty());
        assert_eq!(module.top_forms.len(), 1);
        if let hir::TopForm::Defn { return_type, .. } = &module.top_forms[0] {
            assert_eq!(return_type, &VexType::String);
        } else {
            panic!("expected defn");
        }
    }

    #[test]
    fn deftype_simple() {
        let source = "(deftype Point (x Float) (y Float))";
        let (module, diags) = check_source(source);
        assert!(diags.is_empty());
        assert_eq!(module.top_forms.len(), 1);
        if let hir::TopForm::Deftype { name, fields, .. } = &module.top_forms[0] {
            assert_eq!(name, "Point");
            assert_eq!(fields.len(), 2);
            assert_eq!(fields[0].name, "x");
            assert_eq!(fields[0].ty, VexType::Float);
            assert_eq!(fields[1].name, "y");
            assert_eq!(fields[1].ty, VexType::Float);
        } else {
            panic!("expected deftype");
        }
    }

    #[test]
    fn deftype_multiple_field_types() {
        let source = "(deftype Config (name String) (count Int) (active Bool))";
        let (module, diags) = check_source(source);
        assert!(diags.is_empty());
        if let hir::TopForm::Deftype { fields, .. } = &module.top_forms[0] {
            assert_eq!(fields.len(), 3);
            assert_eq!(fields[0].ty, VexType::String);
            assert_eq!(fields[1].ty, VexType::Int);
            assert_eq!(fields[2].ty, VexType::Bool);
        } else {
            panic!("expected deftype");
        }
    }

    #[test]
    fn deftype_no_fields() {
        let source = "(deftype Empty)";
        let (module, diags) = check_source(source);
        assert!(diags.is_empty());
        if let hir::TopForm::Deftype { name, fields, .. } = &module.top_forms[0] {
            assert_eq!(name, "Empty");
            assert!(fields.is_empty());
        } else {
            panic!("expected deftype");
        }
    }

    #[test]
    fn deftype_unknown_field_type() {
        let source = "(deftype Bad (x Unknown))";
        let (_, diags) = check_source(source);
        assert_eq!(diags.len(), 1);
        assert!(diags[0].message.contains("unknown type"));
    }

    #[test]
    fn deftype_registered_as_type() {
        let source = r#"
            (deftype Point (x Float) (y Float))
            (deftype Line (start Point) (end Point))
        "#;
        let (module, diags) = check_source(source);
        assert!(diags.is_empty());
        assert_eq!(module.top_forms.len(), 2);
        if let hir::TopForm::Deftype { fields, .. } = &module.top_forms[1] {
            assert_eq!(fields.len(), 2);
            assert!(matches!(&fields[0].ty, VexType::Record { name, .. } if name == "Point"));
            assert!(matches!(&fields[1].ty, VexType::Record { name, .. } if name == "Point"));
        } else {
            panic!("expected deftype");
        }
    }

    #[test]
    fn field_access_simple() {
        let source = r#"
            (deftype Point (x Float) (y Float))
            (defn get-x [p : Point] : Float
              (. p x))
        "#;
        let (module, diags) = check_source(source);
        assert!(diags.is_empty());
        assert_eq!(module.top_forms.len(), 2);
        if let hir::TopForm::Defn { body, .. } = &module.top_forms[1] {
            assert_eq!(body[0].ty(), &VexType::Float);
            assert!(matches!(&body[0], hir::Expr::FieldAccess { field, .. } if field == "x"));
        } else {
            panic!("expected defn");
        }
    }

    #[test]
    fn field_access_non_record() {
        let source = r#"
            (defn bad [n : Int]
              (. n x))
        "#;
        let (_, diags) = check_source(source);
        assert_eq!(diags.len(), 1);
        assert!(diags[0].message.contains("has no field"));
    }

    #[test]
    fn field_access_unknown_field() {
        let source = r#"
            (deftype Point (x Float) (y Float))
            (defn bad [p : Point]
              (. p z))
        "#;
        let (_, diags) = check_source(source);
        assert_eq!(diags.len(), 1);
        assert!(diags[0].message.contains("has no field 'z'"));
    }

    #[test]
    fn field_access_nested() {
        let source = r#"
            (deftype Point (x Float) (y Float))
            (deftype Line (start Point) (end Point))
            (defn get-start-x [l : Line] : Float
              (. (. l start) x))
        "#;
        let (module, diags) = check_source(source);
        assert!(diags.is_empty());
        if let hir::TopForm::Defn { body, .. } = &module.top_forms[2] {
            assert_eq!(body[0].ty(), &VexType::Float);
        } else {
            panic!("expected defn");
        }
    }

    #[test]
    fn record_constructor_simple() {
        let source = r#"
            (deftype Point (x Float) (y Float))
            (defn make-point [] : Point
              (Point 1.0 2.0))
        "#;
        let (module, diags) = check_source(source);
        assert!(diags.is_empty());
        if let hir::TopForm::Defn { body, .. } = &module.top_forms[1] {
            assert!(
                matches!(&body[0], hir::Expr::RecordConstructor { name, .. } if name == "Point")
            );
            assert!(matches!(body[0].ty(), VexType::Record { name, .. } if name == "Point"));
        } else {
            panic!("expected defn");
        }
    }

    #[test]
    fn record_constructor_wrong_arity() {
        let source = r#"
            (deftype Point (x Float) (y Float))
            (defn bad [] (Point 1.0))
        "#;
        let (_, diags) = check_source(source);
        assert_eq!(diags.len(), 1);
        assert!(diags[0].message.contains("2 field(s)"));
        assert!(diags[0].message.contains("1 argument(s)"));
    }

    #[test]
    fn record_constructor_wrong_field_type() {
        let source = r#"
            (deftype Point (x Float) (y Float))
            (defn bad [] (Point 1 2.0))
        "#;
        let (_, diags) = check_source(source);
        assert_eq!(diags.len(), 1);
        assert!(diags[0].message.contains("field 'x'"));
        assert!(diags[0].message.contains("Float"));
    }

    #[test]
    fn record_constructor_and_field_access() {
        let source = r#"
            (deftype Point (x Float) (y Float))
            (defn get-x [] : Float
              (. (Point 1.0 2.0) x))
        "#;
        let (module, diags) = check_source(source);
        assert!(diags.is_empty());
        if let hir::TopForm::Defn { body, .. } = &module.top_forms[1] {
            assert_eq!(body[0].ty(), &VexType::Float);
        } else {
            panic!("expected defn");
        }
    }
}
