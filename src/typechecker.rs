use crate::ast;
use crate::builtins;
use crate::diagnostics::Diagnostic;
use crate::hir;
use crate::source::Span;
use crate::types::{RecordField, TypeEnv, UnionVariant, VexType};

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
            ast::Expr::Match {
                scrutinee,
                clauses,
                span,
            } => self.check_match(scrutinee, clauses, *span),
        }
    }

    fn check_symbol(&mut self, name: &str, span: Span) -> Option<hir::Expr> {
        if name == "None" && self.find_variant_union("None").is_none() {
            let ty = VexType::Option(Box::new(self.env.fresh_type_var()));
            return Some(hir::Expr::VariantConstructor {
                union_name: "Option".to_string(),
                variant_name: "None".to_string(),
                args: vec![],
                span,
                ty,
            });
        }

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
        if let ast::Expr::Symbol(name, _) = func {
            if let Some(result) = self.check_builtin_constructor(name, args, span) {
                return result;
            }

            if let Some(record_ty) = self.type_defs.get(name).cloned()
                && matches!(&record_ty, VexType::Record { .. })
            {
                return self.check_record_constructor(name, &record_ty, args, span);
            }

            if let Some((union_name, union_ty)) = self.find_variant_union(name) {
                return self.check_variant_constructor(&union_name, name, &union_ty, args, span);
            }
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

        let ty = match VexType::types_compatible(checked_then.ty(), checked_else.ty()) {
            Some(merged) => merged,
            None => {
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
        };

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
            ast::TypeExpr::Applied { name, args, span } => match name.as_str() {
                "Option" => {
                    if args.len() != 1 {
                        self.diagnostics.push(Diagnostic::error(
                            format!("Option requires 1 type argument, found {}", args.len()),
                            *span,
                        ));
                        return None;
                    }
                    let inner = self.resolve_type(&args[0])?;
                    Some(VexType::Option(Box::new(inner)))
                }
                "Result" => {
                    if args.len() != 2 {
                        self.diagnostics.push(Diagnostic::error(
                            format!("Result requires 2 type arguments, found {}", args.len()),
                            *span,
                        ));
                        return None;
                    }
                    let ok = self.resolve_type(&args[0])?;
                    let err = self.resolve_type(&args[1])?;
                    Some(VexType::Result {
                        ok: Box::new(ok),
                        err: Box::new(err),
                    })
                }
                _ => {
                    self.diagnostics.push(Diagnostic::error(
                        format!("unknown parametric type '{}'", name),
                        *span,
                    ));
                    None
                }
            },
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

            let ty = match VexType::types_compatible(checked_value.ty(), result.ty()) {
                Some(merged) => merged,
                None => {
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
            };

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
            match VexType::types_compatible(&declared, &body_ty) {
                Some(merged) => merged,
                None => {
                    self.diagnostics.push(Diagnostic::error(
                        format!(
                            "declared return type {} doesn't match body type {}",
                            declared, body_ty
                        ),
                        span,
                    ));
                    return None;
                }
            }
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

    fn check_match(
        &mut self,
        scrutinee: &ast::Expr,
        clauses: &[ast::MatchClause],
        span: Span,
    ) -> Option<hir::Expr> {
        let checked_scrutinee = self.check_expr(scrutinee)?;
        let scrutinee_ty = checked_scrutinee.ty().clone();

        let mut checked_clauses = Vec::new();
        let mut result_ty: Option<VexType> = None;

        for clause in clauses {
            self.env.push_scope();
            let checked_pattern = self.check_pattern(&clause.pattern, &scrutinee_ty)?;
            let checked_body = self.check_expr(&clause.body)?;
            let body_ty = checked_body.ty().clone();
            self.env.pop_scope();

            if let Some(ref expected) = result_ty {
                match VexType::types_compatible(&body_ty, expected) {
                    Some(merged) => result_ty = Some(merged),
                    None => {
                        self.diagnostics.push(Diagnostic::error(
                            format!(
                                "match clause body has type {}, expected {}",
                                body_ty, expected
                            ),
                            clause.body.span(),
                        ));
                        return None;
                    }
                }
            } else {
                result_ty = Some(body_ty);
            }

            checked_clauses.push(hir::MatchClause {
                pattern: checked_pattern,
                body: checked_body,
                span: clause.span,
            });
        }

        let ty = result_ty.unwrap_or(VexType::Unit);

        Some(hir::Expr::Match {
            scrutinee: Box::new(checked_scrutinee),
            clauses: checked_clauses,
            span,
            ty,
        })
    }

    fn check_pattern(
        &mut self,
        pattern: &ast::Pattern,
        expected_ty: &VexType,
    ) -> Option<hir::Pattern> {
        match pattern {
            ast::Pattern::Wildcard(span) => Some(hir::Pattern::Wildcard(*span)),
            ast::Pattern::Binding(name, span) => {
                if name == "None"
                    && matches!(expected_ty, VexType::Option(_))
                    && !matches!(expected_ty, VexType::Union { .. })
                {
                    return Some(hir::Pattern::Constructor {
                        union_name: "Option".to_string(),
                        variant_name: "None".to_string(),
                        bindings: vec![],
                        span: *span,
                    });
                }

                self.env.define(name.clone(), expected_ty.clone());
                Some(hir::Pattern::Binding {
                    name: name.clone(),
                    ty: expected_ty.clone(),
                    span: *span,
                })
            }
            ast::Pattern::Literal(expr) => {
                let checked = self.check_expr(expr)?;
                let lit_ty = checked.ty().clone();
                if &lit_ty != expected_ty {
                    self.diagnostics.push(Diagnostic::error(
                        format!(
                            "pattern literal has type {}, but scrutinee has type {}",
                            lit_ty, expected_ty
                        ),
                        expr.span(),
                    ));
                    return None;
                }
                Some(hir::Pattern::Literal(Box::new(checked)))
            }
            ast::Pattern::Constructor { name, args, span } => {
                if let Some(pat) = self.check_builtin_pattern(name, args, *span, expected_ty) {
                    return pat;
                }

                let union_ty = match expected_ty {
                    VexType::Union { .. } => expected_ty,
                    _ => {
                        self.diagnostics.push(Diagnostic::error(
                            format!("constructor pattern used on non-union type {}", expected_ty),
                            *span,
                        ));
                        return None;
                    }
                };

                let (union_name, variants) = match union_ty {
                    VexType::Union { name, variants } => (name, variants),
                    _ => unreachable!(),
                };

                let variant = variants.iter().find(|v| v.name == *name);
                let variant = match variant {
                    Some(v) => v,
                    None => {
                        self.diagnostics.push(Diagnostic::error(
                            format!("{} is not a variant of {}", name, union_name),
                            *span,
                        ));
                        return None;
                    }
                };

                if args.len() != variant.types.len() {
                    self.diagnostics.push(Diagnostic::error(
                        format!(
                            "variant {} has {} field(s), but pattern has {} binding(s)",
                            name,
                            variant.types.len(),
                            args.len()
                        ),
                        *span,
                    ));
                    return None;
                }

                let mut checked_bindings = Vec::new();
                for (arg_pat, field_ty) in args.iter().zip(variant.types.iter()) {
                    checked_bindings.push(self.check_pattern(arg_pat, field_ty)?);
                }

                Some(hir::Pattern::Constructor {
                    union_name: union_name.clone(),
                    variant_name: name.clone(),
                    bindings: checked_bindings,
                    span: *span,
                })
            }
        }
    }

    fn check_builtin_pattern(
        &mut self,
        name: &str,
        args: &[ast::Pattern],
        span: Span,
        expected_ty: &VexType,
    ) -> Option<Option<hir::Pattern>> {
        if matches!(expected_ty, VexType::Union { .. }) {
            return None;
        }

        match name {
            "Some" => {
                let inner = match expected_ty {
                    VexType::Option(inner) => inner.as_ref(),
                    _ => {
                        self.diagnostics.push(Diagnostic::error(
                            format!("Some pattern used on non-Option type {}", expected_ty),
                            span,
                        ));
                        return Some(None);
                    }
                };
                if args.len() != 1 {
                    self.diagnostics.push(Diagnostic::error(
                        format!("Some pattern requires 1 binding, found {}", args.len()),
                        span,
                    ));
                    return Some(None);
                }
                let checked = self.check_pattern(&args[0], inner)?;
                Some(Some(hir::Pattern::Constructor {
                    union_name: "Option".to_string(),
                    variant_name: "Some".to_string(),
                    bindings: vec![checked],
                    span,
                }))
            }
            "None" => {
                if !matches!(expected_ty, VexType::Option(_)) {
                    self.diagnostics.push(Diagnostic::error(
                        format!("None pattern used on non-Option type {}", expected_ty),
                        span,
                    ));
                    return Some(None);
                }
                if !args.is_empty() {
                    self.diagnostics.push(Diagnostic::error(
                        format!("None pattern takes no bindings, found {}", args.len()),
                        span,
                    ));
                    return Some(None);
                }
                Some(Some(hir::Pattern::Constructor {
                    union_name: "Option".to_string(),
                    variant_name: "None".to_string(),
                    bindings: vec![],
                    span,
                }))
            }
            "Ok" => {
                let ok_ty = match expected_ty {
                    VexType::Result { ok, .. } => ok.as_ref(),
                    _ => {
                        self.diagnostics.push(Diagnostic::error(
                            format!("Ok pattern used on non-Result type {}", expected_ty),
                            span,
                        ));
                        return Some(None);
                    }
                };
                if args.len() != 1 {
                    self.diagnostics.push(Diagnostic::error(
                        format!("Ok pattern requires 1 binding, found {}", args.len()),
                        span,
                    ));
                    return Some(None);
                }
                let checked = self.check_pattern(&args[0], ok_ty)?;
                Some(Some(hir::Pattern::Constructor {
                    union_name: "Result".to_string(),
                    variant_name: "Ok".to_string(),
                    bindings: vec![checked],
                    span,
                }))
            }
            "Err" => {
                let err_ty = match expected_ty {
                    VexType::Result { err, .. } => err.as_ref(),
                    _ => {
                        self.diagnostics.push(Diagnostic::error(
                            format!("Err pattern used on non-Result type {}", expected_ty),
                            span,
                        ));
                        return Some(None);
                    }
                };
                if args.len() != 1 {
                    self.diagnostics.push(Diagnostic::error(
                        format!("Err pattern requires 1 binding, found {}", args.len()),
                        span,
                    ));
                    return Some(None);
                }
                let checked = self.check_pattern(&args[0], err_ty)?;
                Some(Some(hir::Pattern::Constructor {
                    union_name: "Result".to_string(),
                    variant_name: "Err".to_string(),
                    bindings: vec![checked],
                    span,
                }))
            }
            _ => None,
        }
    }

    fn find_variant_union(&self, variant_name: &str) -> Option<(String, VexType)> {
        for ty in self.type_defs.values() {
            if let VexType::Union { name, variants } = ty
                && variants.iter().any(|v| v.name == variant_name)
            {
                return Some((name.clone(), ty.clone()));
            }
        }
        None
    }

    fn check_variant_constructor(
        &mut self,
        union_name: &str,
        variant_name: &str,
        union_ty: &VexType,
        args: &[ast::Expr],
        span: Span,
    ) -> Option<hir::Expr> {
        let variants = match union_ty {
            VexType::Union { variants, .. } => variants,
            _ => return None,
        };

        let variant = variants.iter().find(|v| v.name == variant_name)?;

        if variant.types.len() != args.len() {
            self.diagnostics.push(Diagnostic::error(
                format!(
                    "variant {} has {} field(s), but {} argument(s) were provided",
                    variant_name,
                    variant.types.len(),
                    args.len()
                ),
                span,
            ));
            return None;
        }

        let mut checked_args = Vec::new();
        for (arg, expected_ty) in args.iter().zip(variant.types.iter()) {
            let checked = self.check_expr(arg)?;
            let actual_ty = checked.ty().clone();
            if &actual_ty != expected_ty {
                self.diagnostics.push(Diagnostic::error(
                    format!(
                        "expected {} for variant field, found {}",
                        expected_ty, actual_ty
                    ),
                    arg.span(),
                ));
                return None;
            }
            checked_args.push(checked);
        }

        Some(hir::Expr::VariantConstructor {
            union_name: union_name.to_string(),
            variant_name: variant_name.to_string(),
            args: checked_args,
            span,
            ty: union_ty.clone(),
        })
    }

    fn check_builtin_constructor(
        &mut self,
        name: &str,
        args: &[ast::Expr],
        span: Span,
    ) -> Option<Option<hir::Expr>> {
        if self.find_variant_union(name).is_some() {
            return None;
        }

        match name {
            "Some" => {
                if args.len() != 1 {
                    self.diagnostics.push(Diagnostic::error(
                        format!("Some requires 1 argument, found {}", args.len()),
                        span,
                    ));
                    return Some(None);
                }
                let checked = match self.check_expr(&args[0]) {
                    Some(e) => e,
                    None => return Some(None),
                };
                let inner_ty = checked.ty().clone();
                let ty = VexType::Option(Box::new(inner_ty));
                Some(Some(hir::Expr::VariantConstructor {
                    union_name: "Option".to_string(),
                    variant_name: "Some".to_string(),
                    args: vec![checked],
                    span,
                    ty,
                }))
            }
            "None" => {
                if !args.is_empty() {
                    self.diagnostics.push(Diagnostic::error(
                        format!("None takes no arguments, found {}", args.len()),
                        span,
                    ));
                    return Some(None);
                }
                let ty = VexType::Option(Box::new(self.env.fresh_type_var()));
                Some(Some(hir::Expr::VariantConstructor {
                    union_name: "Option".to_string(),
                    variant_name: "None".to_string(),
                    args: vec![],
                    span,
                    ty,
                }))
            }
            "Ok" => {
                if args.len() != 1 {
                    self.diagnostics.push(Diagnostic::error(
                        format!("Ok requires 1 argument, found {}", args.len()),
                        span,
                    ));
                    return Some(None);
                }
                let checked = match self.check_expr(&args[0]) {
                    Some(e) => e,
                    None => return Some(None),
                };
                let ok_ty = checked.ty().clone();
                let ty = VexType::Result {
                    ok: Box::new(ok_ty),
                    err: Box::new(self.env.fresh_type_var()),
                };
                Some(Some(hir::Expr::VariantConstructor {
                    union_name: "Result".to_string(),
                    variant_name: "Ok".to_string(),
                    args: vec![checked],
                    span,
                    ty,
                }))
            }
            "Err" => {
                if args.len() != 1 {
                    self.diagnostics.push(Diagnostic::error(
                        format!("Err requires 1 argument, found {}", args.len()),
                        span,
                    ));
                    return Some(None);
                }
                let checked = match self.check_expr(&args[0]) {
                    Some(e) => e,
                    None => return Some(None),
                };
                let err_ty = checked.ty().clone();
                let ty = VexType::Result {
                    ok: Box::new(self.env.fresh_type_var()),
                    err: Box::new(err_ty),
                };
                Some(Some(hir::Expr::VariantConstructor {
                    union_name: "Result".to_string(),
                    variant_name: "Err".to_string(),
                    args: vec![checked],
                    span,
                    ty,
                }))
            }
            _ => None,
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
            match VexType::types_compatible(&declared, &body_ty) {
                Some(merged) => merged,
                None => {
                    self.diagnostics.push(Diagnostic::error(
                        format!(
                            "declared return type {} doesn't match body type {}",
                            declared, body_ty
                        ),
                        span,
                    ));
                    return None;
                }
            }
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

    fn check_defunion(
        &mut self,
        name: &str,
        variants: &[ast::Variant],
        span: Span,
    ) -> Option<hir::TopForm> {
        let mut checked_variants = Vec::new();
        for variant in variants {
            let mut resolved_types = Vec::new();
            for type_expr in &variant.types {
                resolved_types.push(self.resolve_type(type_expr)?);
            }
            checked_variants.push(UnionVariant {
                name: variant.name.clone(),
                types: resolved_types,
            });
        }

        let union_type = VexType::Union {
            name: name.to_string(),
            variants: checked_variants.clone(),
        };

        self.type_defs.insert(name.to_string(), union_type);

        Some(hir::TopForm::Defunion {
            name: name.to_string(),
            variants: checked_variants,
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
            ast::TopForm::Defunion {
                name,
                variants,
                span,
            } => self.check_defunion(name, variants, *span),
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

    #[test]
    fn defunion_simple() {
        let source = "(defunion Shape (Circle Float) (Rect Float Float))";
        let (module, diags) = check_source(source);
        assert!(diags.is_empty());
        assert_eq!(module.top_forms.len(), 1);
        if let hir::TopForm::Defunion { name, variants, .. } = &module.top_forms[0] {
            assert_eq!(name, "Shape");
            assert_eq!(variants.len(), 2);
            assert_eq!(variants[0].name, "Circle");
            assert_eq!(variants[0].types, vec![VexType::Float]);
            assert_eq!(variants[1].name, "Rect");
            assert_eq!(variants[1].types, vec![VexType::Float, VexType::Float]);
        } else {
            panic!("expected defunion");
        }
    }

    #[test]
    fn defunion_no_data_variant() {
        let source = "(defunion Option (Some Int) (None))";
        let (module, diags) = check_source(source);
        assert!(diags.is_empty());
        if let hir::TopForm::Defunion { variants, .. } = &module.top_forms[0] {
            assert_eq!(variants[0].types, vec![VexType::Int]);
            assert!(variants[1].types.is_empty());
        } else {
            panic!("expected defunion");
        }
    }

    #[test]
    fn defunion_registered_as_type() {
        let source = r#"
            (defunion Shape (Circle Float) (Rect Float Float))
            (defn area [s : Shape] : Float
              0.0)
        "#;
        let (module, diags) = check_source(source);
        assert!(diags.is_empty());
        assert_eq!(module.top_forms.len(), 2);
        if let hir::TopForm::Defn { params, .. } = &module.top_forms[1] {
            assert!(matches!(&params[0].ty, VexType::Union { name, .. } if name == "Shape"));
        } else {
            panic!("expected defn");
        }
    }

    #[test]
    fn defunion_unknown_type() {
        let source = "(defunion Bad (Foo Unknown))";
        let (_, diags) = check_source(source);
        assert_eq!(diags.len(), 1);
        assert!(diags[0].message.contains("unknown type"));
    }

    #[test]
    fn match_simple_wildcard() {
        let source = r#"
            (defn f [x : Int] : Int
              (match x
                _ 0))
        "#;
        let (module, diags) = check_source(source);
        assert!(diags.is_empty(), "{:?}", diags);
        if let hir::TopForm::Defn { body, .. } = &module.top_forms[0] {
            assert!(matches!(&body[0], hir::Expr::Match { ty, .. } if *ty == VexType::Int));
        }
    }

    #[test]
    fn match_constructor_bindings() {
        let source = r#"
            (defunion Option (Some Int) (None))
            (defn unwrap [o : Option] : Int
              (match o
                (Some x) x
                (None) 0))
        "#;
        let (module, diags) = check_source(source);
        assert!(diags.is_empty(), "{:?}", diags);
        assert_eq!(module.top_forms.len(), 2);
        if let hir::TopForm::Defn { body, .. } = &module.top_forms[1] {
            if let hir::Expr::Match { clauses, ty, .. } = &body[0] {
                assert_eq!(clauses.len(), 2);
                assert_eq!(*ty, VexType::Int);
            } else {
                panic!("expected match");
            }
        }
    }

    #[test]
    fn match_branch_type_mismatch() {
        let source = r#"
            (defunion Option (Some Int) (None))
            (defn f [o : Option] : Int
              (match o
                (Some x) x
                (None) "bad"))
        "#;
        let (_, diags) = check_source(source);
        assert_eq!(diags.len(), 1);
        assert!(diags[0].message.contains("match clause body has type"));
    }

    #[test]
    fn match_unknown_variant() {
        let source = r#"
            (defunion Option (Some Int) (None))
            (defn f [o : Option] : Int
              (match o
                (Bad x) 0))
        "#;
        let (_, diags) = check_source(source);
        assert_eq!(diags.len(), 1);
        assert!(diags[0].message.contains("not a variant"));
    }

    #[test]
    fn match_wrong_variant_arity() {
        let source = r#"
            (defunion Option (Some Int) (None))
            (defn f [o : Option] : Int
              (match o
                (Some x y) 0))
        "#;
        let (_, diags) = check_source(source);
        assert_eq!(diags.len(), 1);
        assert!(diags[0].message.contains("binding"));
    }

    #[test]
    fn match_literal_pattern() {
        let source = r#"
            (defn f [x : Int] : String
              (match x
                1 "one"
                2 "two"
                _ "other"))
        "#;
        let (module, diags) = check_source(source);
        assert!(diags.is_empty(), "{:?}", diags);
        if let hir::TopForm::Defn { body, .. } = &module.top_forms[0] {
            assert!(matches!(&body[0], hir::Expr::Match { ty, .. } if *ty == VexType::String));
        }
    }

    #[test]
    fn match_constructor_on_non_union() {
        let source = r#"
            (defn f [x : Int] : Int
              (match x
                (Foo y) 0))
        "#;
        let (_, diags) = check_source(source);
        assert_eq!(diags.len(), 1);
        assert!(diags[0].message.contains("non-union"));
    }

    #[test]
    fn variant_constructor_simple() {
        let source = r#"
            (defunion Shape (Circle Float) (Rect Float Float))
            (defn make-circle [] : Shape
              (Circle 5.0))
        "#;
        let (module, diags) = check_source(source);
        assert!(diags.is_empty(), "{:?}", diags);
        if let hir::TopForm::Defn { body, .. } = &module.top_forms[1] {
            assert!(matches!(&body[0], hir::Expr::VariantConstructor {
                union_name, variant_name, ..
            } if union_name == "Shape" && variant_name == "Circle"));
        }
    }

    #[test]
    fn variant_constructor_wrong_arity() {
        let source = r#"
            (defunion Shape (Circle Float) (Rect Float Float))
            (defn bad [] : Shape
              (Circle 5.0 6.0))
        "#;
        let (_, diags) = check_source(source);
        assert_eq!(diags.len(), 1);
        assert!(diags[0].message.contains("field"));
    }

    #[test]
    fn variant_constructor_wrong_type() {
        let source = r#"
            (defunion Shape (Circle Float) (Rect Float Float))
            (defn bad [] : Shape
              (Circle "bad"))
        "#;
        let (_, diags) = check_source(source);
        assert_eq!(diags.len(), 1);
        assert!(diags[0].message.contains("expected"));
    }

    #[test]
    fn variant_constructor_no_args() {
        let source = r#"
            (defunion Option (Some Int) (None))
            (defn make-none [] : Option
              (None))
        "#;
        let (module, diags) = check_source(source);
        assert!(diags.is_empty(), "{:?}", diags);
        if let hir::TopForm::Defn { body, .. } = &module.top_forms[1] {
            assert!(matches!(&body[0], hir::Expr::VariantConstructor {
                variant_name, args, ..
            } if variant_name == "None" && args.is_empty()));
        }
    }

    #[test]
    fn resolve_option_type() {
        let source = r#"
            (defn f [x : (Option Int)] : (Option Int)
              x)
        "#;
        let (module, diags) = check_source(source);
        assert!(diags.is_empty(), "{:?}", diags);
        if let hir::TopForm::Defn {
            params,
            return_type,
            ..
        } = &module.top_forms[0]
        {
            assert_eq!(params[0].ty, VexType::Option(Box::new(VexType::Int)));
            assert_eq!(*return_type, VexType::Option(Box::new(VexType::Int)));
        } else {
            panic!("expected defn");
        }
    }

    #[test]
    fn resolve_result_type() {
        let source = r#"
            (defn f [x : (Result Int String)] : (Result Int String)
              x)
        "#;
        let (module, diags) = check_source(source);
        assert!(diags.is_empty(), "{:?}", diags);
        if let hir::TopForm::Defn {
            params,
            return_type,
            ..
        } = &module.top_forms[0]
        {
            assert_eq!(
                params[0].ty,
                VexType::Result {
                    ok: Box::new(VexType::Int),
                    err: Box::new(VexType::String),
                }
            );
            assert_eq!(
                *return_type,
                VexType::Result {
                    ok: Box::new(VexType::Int),
                    err: Box::new(VexType::String),
                }
            );
        } else {
            panic!("expected defn");
        }
    }

    #[test]
    fn resolve_nested_option() {
        let source = r#"
            (defn f [x : (Option (Option Int))] : (Option (Option Int))
              x)
        "#;
        let (module, diags) = check_source(source);
        assert!(diags.is_empty(), "{:?}", diags);
        if let hir::TopForm::Defn { params, .. } = &module.top_forms[0] {
            assert_eq!(
                params[0].ty,
                VexType::Option(Box::new(VexType::Option(Box::new(VexType::Int))))
            );
        } else {
            panic!("expected defn");
        }
    }

    #[test]
    fn error_option_wrong_arity() {
        let (_, diags) = check_source("(defn f [x : (Option Int String)] x)");
        assert_eq!(diags.len(), 1);
        assert!(diags[0].message.contains("Option requires 1 type argument"));
    }

    #[test]
    fn error_result_wrong_arity() {
        let (_, diags) = check_source("(defn f [x : (Result Int)] x)");
        assert_eq!(diags.len(), 1);
        assert!(
            diags[0]
                .message
                .contains("Result requires 2 type arguments")
        );
    }

    #[test]
    fn error_unknown_parametric_type() {
        let (_, diags) = check_source("(defn f [x : (List Int)] x)");
        assert_eq!(diags.len(), 1);
        assert!(diags[0].message.contains("unknown parametric type"));
    }

    #[test]
    fn some_constructor() {
        let source = r#"
            (defn f [] : (Option Int)
              (Some 42))
        "#;
        let (module, diags) = check_source(source);
        assert!(diags.is_empty(), "{:?}", diags);
        if let hir::TopForm::Defn { body, .. } = &module.top_forms[0] {
            assert!(matches!(
                &body[0],
                hir::Expr::VariantConstructor {
                    union_name,
                    variant_name,
                    ..
                } if union_name == "Option" && variant_name == "Some"
            ));
            assert_eq!(body[0].ty(), &VexType::Option(Box::new(VexType::Int)));
        }
    }

    #[test]
    fn none_constructor_call() {
        let source = r#"
            (defn f [] : (Option Int)
              (None))
        "#;
        let (module, diags) = check_source(source);
        assert!(diags.is_empty(), "{:?}", diags);
        if let hir::TopForm::Defn { body, .. } = &module.top_forms[0] {
            assert!(matches!(
                &body[0],
                hir::Expr::VariantConstructor {
                    variant_name,
                    args,
                    ..
                } if variant_name == "None" && args.is_empty()
            ));
        }
    }

    #[test]
    fn none_constructor_bare() {
        let source = r#"
            (defn f [] : (Option Int)
              None)
        "#;
        let (module, diags) = check_source(source);
        assert!(diags.is_empty(), "{:?}", diags);
        if let hir::TopForm::Defn { body, .. } = &module.top_forms[0] {
            assert!(matches!(
                &body[0],
                hir::Expr::VariantConstructor {
                    variant_name,
                    ..
                } if variant_name == "None"
            ));
        }
    }

    #[test]
    fn ok_constructor() {
        let source = r#"
            (defn f [] : (Result Int String)
              (Ok 42))
        "#;
        let (module, diags) = check_source(source);
        assert!(diags.is_empty(), "{:?}", diags);
        if let hir::TopForm::Defn { body, .. } = &module.top_forms[0] {
            assert!(matches!(
                &body[0],
                hir::Expr::VariantConstructor {
                    union_name,
                    variant_name,
                    ..
                } if union_name == "Result" && variant_name == "Ok"
            ));
        }
    }

    #[test]
    fn err_constructor() {
        let source = r#"
            (defn f [] : (Result Int String)
              (Err "bad"))
        "#;
        let (module, diags) = check_source(source);
        assert!(diags.is_empty(), "{:?}", diags);
        if let hir::TopForm::Defn { body, .. } = &module.top_forms[0] {
            assert!(matches!(
                &body[0],
                hir::Expr::VariantConstructor {
                    union_name,
                    variant_name,
                    ..
                } if union_name == "Result" && variant_name == "Err"
            ));
        }
    }

    #[test]
    fn error_some_wrong_arity() {
        let (_, diags) = check_source("(Some 1 2)");
        assert_eq!(diags.len(), 1);
        assert!(diags[0].message.contains("Some requires 1 argument"));
    }

    #[test]
    fn error_none_with_args() {
        let (_, diags) = check_source("(None 42)");
        assert_eq!(diags.len(), 1);
        assert!(diags[0].message.contains("None takes no arguments"));
    }

    #[test]
    fn match_option_patterns() {
        let source = r#"
            (defn unwrap [o : (Option Int)] : Int
              (match o
                (Some x) x
                None 0))
        "#;
        let (module, diags) = check_source(source);
        assert!(diags.is_empty(), "{:?}", diags);
        if let hir::TopForm::Defn { body, .. } = &module.top_forms[0] {
            if let hir::Expr::Match { clauses, ty, .. } = &body[0] {
                assert_eq!(clauses.len(), 2);
                assert_eq!(*ty, VexType::Int);
            } else {
                panic!("expected match");
            }
        }
    }

    #[test]
    fn match_result_patterns() {
        let source = r#"
            (defn unwrap-result [r : (Result Int String)] : Int
              (match r
                (Ok x) x
                (Err e) -1))
        "#;
        let (module, diags) = check_source(source);
        assert!(diags.is_empty(), "{:?}", diags);
        if let hir::TopForm::Defn { body, .. } = &module.top_forms[0] {
            if let hir::Expr::Match { clauses, ty, .. } = &body[0] {
                assert_eq!(clauses.len(), 2);
                assert_eq!(*ty, VexType::Int);
            } else {
                panic!("expected match");
            }
        }
    }

    #[test]
    fn error_some_pattern_on_non_option() {
        let source = r#"
            (defn f [x : Int] : Int
              (match x
                (Some y) y))
        "#;
        let (_, diags) = check_source(source);
        assert_eq!(diags.len(), 1);
        assert!(diags[0].message.contains("non-Option"));
    }

    #[test]
    fn error_ok_pattern_on_non_result() {
        let source = r#"
            (defn f [x : Int] : Int
              (match x
                (Ok y) y))
        "#;
        let (_, diags) = check_source(source);
        assert_eq!(diags.len(), 1);
        assert!(diags[0].message.contains("non-Result"));
    }

    #[test]
    fn if_with_option_branches() {
        let source = r#"
            (defn f [x : Int] : (Option Int)
              (if (> x 0)
                (Some x)
                None))
        "#;
        let (module, diags) = check_source(source);
        assert!(diags.is_empty(), "{:?}", diags);
        if let hir::TopForm::Defn { return_type, .. } = &module.top_forms[0] {
            assert_eq!(*return_type, VexType::Option(Box::new(VexType::Int)));
        }
    }
}
