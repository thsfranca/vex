use std::collections::BTreeSet;
use std::fmt::Write;

use crate::builtins;
use crate::builtins::GoTranslation;
use crate::hir;
use crate::types::{RecordField, UnionVariant, VexType};

pub fn generate(module: &hir::Module) -> String {
    let mut cg = Generator::new();
    cg.emit_module(module);
    cg.output
}

struct Generator {
    output: String,
    indent: usize,
}

impl Generator {
    fn new() -> Self {
        Self {
            output: String::new(),
            indent: 0,
        }
    }

    fn write(&mut self, s: &str) {
        self.output.push_str(s);
    }

    fn writeln(&mut self, s: &str) {
        self.write_indent();
        self.output.push_str(s);
        self.output.push('\n');
    }

    fn newline(&mut self) {
        self.output.push('\n');
    }

    fn write_indent(&mut self) {
        for _ in 0..self.indent {
            self.output.push('\t');
        }
    }

    fn emit_module(&mut self, module: &hir::Module) {
        self.writeln("package main");
        self.newline();

        let imports = collect_imports(module);
        if !imports.is_empty() {
            if imports.len() == 1 {
                writeln!(self.output, "import \"{}\"", imports.iter().next().unwrap()).unwrap();
            } else {
                self.writeln("import (");
                self.indent += 1;
                for imp in &imports {
                    self.write_indent();
                    writeln!(self.output, "\"{}\"", imp).unwrap();
                }
                self.indent -= 1;
                self.writeln(")");
            }
            self.newline();
        }

        for form in &module.top_forms {
            self.emit_top_form(form);
            self.newline();
        }
    }

    fn emit_top_form(&mut self, form: &hir::TopForm) {
        match form {
            hir::TopForm::Defn {
                name,
                params,
                return_type,
                body,
                ..
            } => self.emit_defn(name, params, return_type, body),
            hir::TopForm::Def {
                name, ty, value, ..
            } => self.emit_def(name, ty, value),
            hir::TopForm::Deftype { name, fields, .. } => self.emit_deftype(name, fields),
            hir::TopForm::Defunion { name, variants, .. } => self.emit_defunion(name, variants),
            hir::TopForm::Expr(expr) => {
                self.write_indent();
                self.emit_expr(expr);
                self.newline();
            }
        }
    }

    fn emit_defn(
        &mut self,
        name: &str,
        params: &[hir::Param],
        return_type: &VexType,
        body: &[hir::Expr],
    ) {
        self.write("func ");
        self.write(&vex_to_go_name(name));
        self.write("(");
        for (i, param) in params.iter().enumerate() {
            if i > 0 {
                self.write(", ");
            }
            self.write(&vex_to_go_name(&param.name));
            self.write(" ");
            self.write(&go_type(&param.ty));
        }
        self.write(")");
        let ret_str = go_type(return_type);
        if !ret_str.is_empty() {
            self.write(" ");
            self.write(&ret_str);
        }
        self.write(" {\n");
        self.indent += 1;

        for (i, expr) in body.iter().enumerate() {
            self.write_indent();
            if i == body.len() - 1 && return_type != &VexType::Unit {
                self.write("return ");
            }
            self.emit_expr(expr);
            self.newline();
        }

        self.indent -= 1;
        self.writeln("}");
    }

    fn emit_def(&mut self, name: &str, ty: &VexType, value: &hir::Expr) {
        self.write("var ");
        self.write(&vex_to_go_name(name));
        self.write(" ");
        self.write(&go_type(ty));
        self.write(" = ");
        self.emit_expr(value);
        self.newline();
    }

    fn emit_deftype(&mut self, name: &str, fields: &[RecordField]) {
        self.write("type ");
        self.write(&vex_to_go_public_name(name));
        self.write(" struct {");
        self.newline();
        self.indent += 1;
        for field in fields {
            self.write_indent();
            self.write(&vex_to_go_public_name(&field.name));
            self.write(" ");
            self.write(&go_type(&field.ty));
            self.newline();
        }
        self.indent -= 1;
        self.writeln("}");
    }

    fn emit_defunion(&mut self, name: &str, variants: &[UnionVariant]) {
        let go_name = vex_to_go_public_name(name);
        let marker = format!("is{}", go_name);

        self.write("type ");
        self.write(&go_name);
        self.write(" interface { ");
        self.write(&marker);
        self.write("() }");
        self.newline();
        self.newline();

        for variant in variants {
            let variant_go = format!("{}_{}", go_name, vex_to_go_public_name(&variant.name));

            self.write("type ");
            self.write(&variant_go);
            self.write(" struct {");
            if variant.types.is_empty() {
                self.write("}");
            } else {
                self.newline();
                self.indent += 1;
                for (i, ty) in variant.types.iter().enumerate() {
                    self.write_indent();
                    write!(self.output, "V{} {}", i, go_type(ty)).unwrap();
                    self.newline();
                }
                self.indent -= 1;
                self.write("}");
            }
            self.newline();

            self.write("func (");
            self.write(&variant_go);
            self.write(") ");
            self.write(&marker);
            self.write("() {}");
            self.newline();
            self.newline();
        }
    }

    fn emit_expr(&mut self, expr: &hir::Expr) {
        match expr {
            hir::Expr::Int(n, _) => write!(self.output, "{}", n).unwrap(),
            hir::Expr::Float(f, _) => {
                let s = format!("{}", f);
                if s.contains('.') {
                    self.write(&s);
                } else {
                    write!(self.output, "{}.0", f).unwrap();
                }
            }
            hir::Expr::String(s, _) => write!(self.output, "\"{}\"", s).unwrap(),
            hir::Expr::Bool(b, _) => write!(self.output, "{}", b).unwrap(),
            hir::Expr::Nil(_) => {}
            hir::Expr::Var { name, .. } => {
                if builtins::is_builtin(name) {
                    return;
                }
                self.write(&vex_to_go_name(name));
            }
            hir::Expr::If {
                test,
                then_branch,
                else_branch,
                ty,
                ..
            } => self.emit_if(test, then_branch, else_branch, ty),
            hir::Expr::Let { bindings, body, .. } => self.emit_let(bindings, body),
            hir::Expr::Call { func, args, ty, .. } => self.emit_call(func, args, ty),
            hir::Expr::Lambda {
                params,
                return_type,
                body,
                ..
            } => self.emit_lambda(params, return_type, body),
            hir::Expr::FieldAccess { object, field, .. } => {
                self.emit_expr(object);
                self.write(".");
                self.write(&vex_to_go_public_name(field));
            }
            hir::Expr::RecordConstructor { name, args, ty, .. } => {
                self.emit_record_constructor(name, args, ty);
            }
            hir::Expr::Match {
                scrutinee, clauses, ..
            } => self.emit_match(scrutinee, clauses),
            hir::Expr::VariantConstructor {
                union_name,
                variant_name,
                args,
                ty,
                ..
            } => {
                if union_name == "Option" || union_name == "Result" {
                    self.emit_builtin_variant_constructor(union_name, variant_name, args, ty);
                } else {
                    let go_variant = format!(
                        "{}_{}",
                        vex_to_go_public_name(union_name),
                        vex_to_go_public_name(variant_name)
                    );
                    self.write(&go_variant);
                    self.write("{");
                    for (i, arg) in args.iter().enumerate() {
                        if i > 0 {
                            self.write(", ");
                        }
                        write!(self.output, "V{}: ", i).unwrap();
                        self.emit_expr(arg);
                    }
                    self.write("}");
                }
            }
        }
    }

    fn emit_if(
        &mut self,
        test: &hir::Expr,
        then_branch: &hir::Expr,
        else_branch: &hir::Expr,
        ty: &VexType,
    ) {
        if ty == &VexType::Unit {
            self.write("if ");
            self.emit_expr(test);
            self.write(" {\n");
            self.indent += 1;
            self.write_indent();
            self.emit_expr(then_branch);
            self.newline();
            self.indent -= 1;
            self.write_indent();
            self.write("} else {\n");
            self.indent += 1;
            self.write_indent();
            self.emit_expr(else_branch);
            self.newline();
            self.indent -= 1;
            self.write_indent();
            self.write("}");
        } else {
            self.write("func() ");
            self.write(&go_type(ty));
            self.write(" { if ");
            self.emit_expr(test);
            self.write(" { return ");
            self.emit_expr(then_branch);
            self.write(" } else { return ");
            self.emit_expr(else_branch);
            self.write(" } }()");
        }
    }

    fn emit_let(&mut self, bindings: &[hir::Binding], body: &[hir::Expr]) {
        self.write("func() ");
        if let Some(last) = body.last() {
            let ty = last.ty();
            if ty != &VexType::Unit {
                self.write(&go_type(ty));
                self.write(" ");
            }
        }
        self.write("{\n");
        self.indent += 1;

        for binding in bindings {
            self.write_indent();
            self.write(&vex_to_go_name(&binding.name));
            self.write(" := ");
            self.emit_expr(&binding.value);
            self.newline();
            let _ = &binding.name;
        }

        for (i, expr) in body.iter().enumerate() {
            self.write_indent();
            if i == body.len() - 1 && expr.ty() != &VexType::Unit {
                self.write("return ");
            }
            self.emit_expr(expr);
            self.newline();
        }

        self.indent -= 1;
        self.write_indent();
        self.write("}()");
    }

    fn emit_call(&mut self, func: &hir::Expr, args: &[hir::Expr], _ty: &VexType) {
        let func_name = if let hir::Expr::Var { name, .. } = func {
            Some(name.as_str())
        } else {
            None
        };

        if func_name == Some("each") && args.len() == 2 {
            self.emit_each(&args[0], &args[1]);
            return;
        }

        if func_name == Some("map") && args.len() == 2 {
            self.emit_map(&args[0], &args[1], _ty);
            return;
        }

        if func_name == Some("filter") && args.len() == 2 {
            self.emit_filter(&args[0], &args[1], _ty);
            return;
        }

        if let Some(name) = func_name
            && let Some(builtin) = builtins::lookup(name)
        {
            match &builtin.go {
                GoTranslation::Infix(op) => {
                    self.write("(");
                    self.emit_expr(&args[0]);
                    write!(self.output, " {} ", op).unwrap();
                    self.emit_expr(&args[1]);
                    self.write(")");
                    return;
                }
                GoTranslation::Prefix(op) => {
                    self.write(op);
                    self.emit_expr(&args[0]);
                    return;
                }
                GoTranslation::FuncCall { go_name, .. } => {
                    self.write(go_name);
                    self.write("(");
                    for (i, arg) in args.iter().enumerate() {
                        if i > 0 {
                            self.write(", ");
                        }
                        self.emit_expr(arg);
                    }
                    self.write(")");
                    return;
                }
            }
        }

        self.emit_expr(func);
        self.write("(");
        for (i, arg) in args.iter().enumerate() {
            if i > 0 {
                self.write(", ");
            }
            self.emit_expr(arg);
        }
        self.write(")");
    }

    fn emit_each(&mut self, list: &hir::Expr, callback: &hir::Expr) {
        if let hir::Expr::Lambda { params, body, .. } = callback {
            let param_name = if let Some(p) = params.first() {
                vex_to_go_name(&p.name)
            } else {
                "_v".to_string()
            };
            self.write("for _, ");
            self.write(&param_name);
            self.write(" := range ");
            self.emit_expr(list);
            self.write(" {\n");
            self.indent += 1;
            for expr in body {
                self.write_indent();
                self.emit_expr(expr);
                self.newline();
            }
            self.indent -= 1;
            self.write_indent();
            self.write("}");
        } else {
            self.write("for _, _v := range ");
            self.emit_expr(list);
            self.write(" {\n");
            self.indent += 1;
            self.write_indent();
            self.emit_expr(callback);
            self.write("(_v)");
            self.newline();
            self.indent -= 1;
            self.write_indent();
            self.write("}");
        }
    }

    fn emit_map(&mut self, list: &hir::Expr, callback: &hir::Expr, result_ty: &VexType) {
        let result_go_ty = go_type(result_ty);
        self.write("func() ");
        self.write(&result_go_ty);
        self.write(" {\n");
        self.indent += 1;

        self.write_indent();
        self.write("_result := make(");
        self.write(&result_go_ty);
        self.write(", 0, len(");
        self.emit_expr(list);
        self.write("))");
        self.newline();

        if let hir::Expr::Lambda { params, body, .. } = callback {
            let param_name = if let Some(p) = params.first() {
                vex_to_go_name(&p.name)
            } else {
                "_v".to_string()
            };
            self.write_indent();
            self.write("for _, ");
            self.write(&param_name);
            self.write(" := range ");
            self.emit_expr(list);
            self.write(" {\n");
            self.indent += 1;
            self.write_indent();
            self.write("_result = append(_result, ");
            if let Some(last) = body.last() {
                self.emit_expr(last);
            }
            self.write(")");
            self.newline();
            self.indent -= 1;
            self.write_indent();
            self.write("}");
            self.newline();
        } else {
            self.write_indent();
            self.write("for _, _v := range ");
            self.emit_expr(list);
            self.write(" {\n");
            self.indent += 1;
            self.write_indent();
            self.write("_result = append(_result, ");
            self.emit_expr(callback);
            self.write("(_v))");
            self.newline();
            self.indent -= 1;
            self.write_indent();
            self.write("}");
            self.newline();
        }

        self.write_indent();
        self.write("return _result");
        self.newline();
        self.indent -= 1;
        self.write_indent();
        self.write("}()");
    }

    fn emit_filter(&mut self, list: &hir::Expr, predicate: &hir::Expr, result_ty: &VexType) {
        let result_go_ty = go_type(result_ty);
        self.write("func() ");
        self.write(&result_go_ty);
        self.write(" {\n");
        self.indent += 1;

        self.write_indent();
        self.write("var _result ");
        self.write(&result_go_ty);
        self.newline();

        if let hir::Expr::Lambda { params, body, .. } = predicate {
            let param_name = if let Some(p) = params.first() {
                vex_to_go_name(&p.name)
            } else {
                "_v".to_string()
            };
            self.write_indent();
            self.write("for _, ");
            self.write(&param_name);
            self.write(" := range ");
            self.emit_expr(list);
            self.write(" {\n");
            self.indent += 1;
            self.write_indent();
            self.write("if ");
            if let Some(last) = body.last() {
                self.emit_expr(last);
            }
            self.write(" {\n");
            self.indent += 1;
            self.write_indent();
            self.write("_result = append(_result, ");
            self.write(&param_name);
            self.write(")");
            self.newline();
            self.indent -= 1;
            self.write_indent();
            self.write("}");
            self.newline();
            self.indent -= 1;
            self.write_indent();
            self.write("}");
            self.newline();
        } else {
            self.write_indent();
            self.write("for _, _v := range ");
            self.emit_expr(list);
            self.write(" {\n");
            self.indent += 1;
            self.write_indent();
            self.write("if ");
            self.emit_expr(predicate);
            self.write("(_v) {\n");
            self.indent += 1;
            self.write_indent();
            self.write("_result = append(_result, _v)");
            self.newline();
            self.indent -= 1;
            self.write_indent();
            self.write("}");
            self.newline();
            self.indent -= 1;
            self.write_indent();
            self.write("}");
            self.newline();
        }

        self.write_indent();
        self.write("return _result");
        self.newline();
        self.indent -= 1;
        self.write_indent();
        self.write("}()");
    }

    fn emit_lambda(&mut self, params: &[hir::Param], return_type: &VexType, body: &[hir::Expr]) {
        self.write("func(");
        for (i, param) in params.iter().enumerate() {
            if i > 0 {
                self.write(", ");
            }
            self.write(&vex_to_go_name(&param.name));
            self.write(" ");
            self.write(&go_type(&param.ty));
        }
        self.write(")");
        let ret_str = go_type(return_type);
        if !ret_str.is_empty() {
            self.write(" ");
            self.write(&ret_str);
        }
        self.write(" {\n");
        self.indent += 1;

        for (i, expr) in body.iter().enumerate() {
            self.write_indent();
            if i == body.len() - 1 && return_type != &VexType::Unit {
                self.write("return ");
            }
            self.emit_expr(expr);
            self.newline();
        }

        self.indent -= 1;
        self.write_indent();
        self.write("}");
    }

    fn emit_record_constructor(&mut self, name: &str, args: &[hir::Expr], ty: &VexType) {
        let fields = match ty {
            VexType::Record { fields, .. } => fields,
            _ => return,
        };
        self.write(&vex_to_go_public_name(name));
        self.write("{");
        for (i, (field, arg)) in fields.iter().zip(args.iter()).enumerate() {
            if i > 0 {
                self.write(", ");
            }
            self.write(&vex_to_go_public_name(&field.name));
            self.write(": ");
            self.emit_expr(arg);
        }
        self.write("}");
    }

    fn emit_builtin_variant_constructor(
        &mut self,
        union_name: &str,
        variant_name: &str,
        args: &[hir::Expr],
        ty: &VexType,
    ) {
        match (union_name, variant_name) {
            ("Option", "Some") => {
                if let VexType::Option(inner) = ty {
                    write!(self.output, "vexrt.Some[{}]", go_type(inner)).unwrap();
                    self.write("(");
                    if let Some(arg) = args.first() {
                        self.emit_expr(arg);
                    }
                    self.write(")");
                }
            }
            ("Option", "None") => {
                if let VexType::Option(inner) = ty {
                    write!(self.output, "vexrt.None[{}]()", go_type(inner)).unwrap();
                }
            }
            ("Result", "Ok") => {
                if let VexType::Result { ok, err } = ty {
                    write!(self.output, "vexrt.Ok[{}, {}]", go_type(ok), go_type(err)).unwrap();
                    self.write("(");
                    if let Some(arg) = args.first() {
                        self.emit_expr(arg);
                    }
                    self.write(")");
                }
            }
            ("Result", "Err") => {
                if let VexType::Result { ok, err } = ty {
                    write!(self.output, "vexrt.Err[{}, {}]", go_type(ok), go_type(err)).unwrap();
                    self.write("(");
                    if let Some(arg) = args.first() {
                        self.emit_expr(arg);
                    }
                    self.write(")");
                }
            }
            _ => {}
        }
    }

    fn emit_match(&mut self, scrutinee: &hir::Expr, clauses: &[hir::MatchClause]) {
        let scrutinee_ty = scrutinee.ty();
        let is_option_or_result =
            matches!(scrutinee_ty, VexType::Option(_) | VexType::Result { .. });

        if is_option_or_result {
            self.emit_match_option_result(scrutinee, clauses);
            return;
        }

        let is_type_switch = clauses
            .iter()
            .any(|c| matches!(&c.pattern, hir::Pattern::Constructor { .. }));

        if is_type_switch {
            self.emit_match_type_switch(scrutinee, clauses);
        } else {
            self.emit_match_value_switch(scrutinee, clauses);
        }
    }

    fn emit_match_option_result(&mut self, scrutinee: &hir::Expr, clauses: &[hir::MatchClause]) {
        self.write("func() ");
        if let Some(first) = clauses.first() {
            let ret_type = go_type(first.body.ty());
            if !ret_type.is_empty() {
                self.write(&ret_type);
                self.write(" ");
            }
        }
        self.write("{");
        self.newline();
        self.indent += 1;

        let mut some_ok_clause: Option<&hir::MatchClause> = None;
        let mut none_err_clause: Option<&hir::MatchClause> = None;
        let mut wildcard_clause: Option<&hir::MatchClause> = None;

        for clause in clauses {
            match &clause.pattern {
                hir::Pattern::Constructor { variant_name, .. }
                    if variant_name == "Some" || variant_name == "Ok" =>
                {
                    some_ok_clause = Some(clause);
                }
                hir::Pattern::Constructor { variant_name, .. }
                    if variant_name == "None" || variant_name == "Err" =>
                {
                    none_err_clause = Some(clause);
                }
                hir::Pattern::Wildcard(_) | hir::Pattern::Binding { .. } => {
                    wildcard_clause = Some(clause);
                }
                _ => {}
            }
        }

        let is_option = clauses.iter().any(|c| {
            matches!(&c.pattern, hir::Pattern::Constructor { union_name, .. } if union_name == "Option")
        });

        let condition_field = if is_option { "IsSome" } else { "IsOk" };

        self.write_indent();
        self.write("if ");
        self.emit_expr(scrutinee);
        write!(self.output, ".{} ", condition_field).unwrap();
        self.write("{");
        self.newline();
        self.indent += 1;

        if let Some(clause) = some_ok_clause {
            if let hir::Pattern::Constructor { bindings, .. } = &clause.pattern {
                for binding in bindings {
                    if let hir::Pattern::Binding { name, .. } = binding {
                        self.write_indent();
                        self.write(&vex_to_go_name(name));
                        self.write(" := ");
                        self.emit_expr(scrutinee);
                        self.write(".Value");
                        self.newline();
                    }
                }
            }
            self.write_indent();
            self.write("return ");
            self.emit_expr(&clause.body);
            self.newline();
        }

        self.indent -= 1;
        self.write_indent();
        self.write("} else {");
        self.newline();
        self.indent += 1;

        if let Some(clause) = none_err_clause {
            if let hir::Pattern::Constructor { bindings, .. } = &clause.pattern {
                for binding in bindings {
                    if let hir::Pattern::Binding { name, .. } = binding {
                        self.write_indent();
                        self.write(&vex_to_go_name(name));
                        self.write(" := ");
                        self.emit_expr(scrutinee);
                        self.write(".Error");
                        self.newline();
                    }
                }
            }
            self.write_indent();
            self.write("return ");
            self.emit_expr(&clause.body);
            self.newline();
        } else if let Some(clause) = wildcard_clause {
            self.write_indent();
            self.write("return ");
            self.emit_expr(&clause.body);
            self.newline();
        } else {
            self.write_indent();
            self.write("return *new(");
            if let Some(first) = clauses.first() {
                self.write(&go_type(first.body.ty()));
            }
            self.write(")");
            self.newline();
        }

        self.indent -= 1;
        self.write_indent();
        self.write("}");
        self.newline();

        self.indent -= 1;
        self.write("}()");
    }

    fn emit_match_type_switch(&mut self, scrutinee: &hir::Expr, clauses: &[hir::MatchClause]) {
        self.write("func() ");
        if let Some(first) = clauses.first() {
            let ret_type = go_type(first.body.ty());
            if !ret_type.is_empty() {
                self.write(&ret_type);
                self.write(" ");
            }
        }
        self.write("{ switch _v := ");
        self.emit_expr(scrutinee);
        self.write(".(type) {");
        self.newline();

        for clause in clauses {
            match &clause.pattern {
                hir::Pattern::Constructor {
                    union_name,
                    variant_name,
                    bindings,
                    ..
                } => {
                    let variant_go = format!(
                        "{}_{}",
                        vex_to_go_public_name(union_name),
                        vex_to_go_public_name(variant_name)
                    );
                    self.write_indent();
                    self.write("case ");
                    self.write(&variant_go);
                    self.write(":");
                    self.newline();
                    self.indent += 1;
                    for (i, binding) in bindings.iter().enumerate() {
                        if let hir::Pattern::Binding { name, .. } = binding {
                            self.write_indent();
                            self.write(&vex_to_go_name(name));
                            write!(self.output, " := _v.V{}", i).unwrap();
                            self.newline();
                        }
                    }
                    self.write_indent();
                    self.write("return ");
                    self.emit_expr(&clause.body);
                    self.newline();
                    self.indent -= 1;
                }
                hir::Pattern::Wildcard(_) | hir::Pattern::Binding { .. } => {
                    self.write_indent();
                    self.write("default:");
                    self.newline();
                    self.indent += 1;
                    self.write_indent();
                    self.write("return ");
                    self.emit_expr(&clause.body);
                    self.newline();
                    self.indent -= 1;
                }
                _ => {}
            }
        }

        self.write_indent();
        self.write("}");
        self.newline();
        self.write_indent();
        self.write("return *new(");
        if let Some(first) = clauses.first() {
            self.write(&go_type(first.body.ty()));
        }
        self.write(")");
        self.newline();
        self.write("}()");
    }

    fn emit_match_value_switch(&mut self, scrutinee: &hir::Expr, clauses: &[hir::MatchClause]) {
        self.write("func() ");
        if let Some(first) = clauses.first() {
            let ret_type = go_type(first.body.ty());
            if !ret_type.is_empty() {
                self.write(&ret_type);
                self.write(" ");
            }
        }
        self.write("{ switch ");
        self.emit_expr(scrutinee);
        self.write(" {");
        self.newline();

        for clause in clauses {
            match &clause.pattern {
                hir::Pattern::Literal(expr) => {
                    self.write_indent();
                    self.write("case ");
                    self.emit_expr(expr);
                    self.write(":");
                    self.newline();
                    self.indent += 1;
                    self.write_indent();
                    self.write("return ");
                    self.emit_expr(&clause.body);
                    self.newline();
                    self.indent -= 1;
                }
                hir::Pattern::Wildcard(_) | hir::Pattern::Binding { .. } => {
                    self.write_indent();
                    self.write("default:");
                    self.newline();
                    self.indent += 1;
                    self.write_indent();
                    self.write("return ");
                    self.emit_expr(&clause.body);
                    self.newline();
                    self.indent -= 1;
                }
                _ => {}
            }
        }

        self.write_indent();
        self.write("}");
        self.newline();
        self.write_indent();
        self.write("return *new(");
        if let Some(first) = clauses.first() {
            self.write(&go_type(first.body.ty()));
        }
        self.write(")");
        self.newline();
        self.write("}()");
    }
}

fn collect_imports(module: &hir::Module) -> BTreeSet<String> {
    let mut names = Vec::new();
    for form in &module.top_forms {
        collect_builtin_calls_top_form(form, &mut names);
    }
    let name_refs: Vec<&str> = names.iter().map(|s| s.as_str()).collect();
    let mut imports: BTreeSet<String> = builtins::go_imports(&name_refs)
        .into_iter()
        .map(|s| s.to_string())
        .collect();
    if needs_vexrt(module) {
        imports.insert("vex_out/vexrt".to_string());
    }
    imports
}

fn collect_builtin_calls_expr(expr: &hir::Expr, names: &mut Vec<String>) {
    match expr {
        hir::Expr::Int(..)
        | hir::Expr::Float(..)
        | hir::Expr::String(..)
        | hir::Expr::Bool(..)
        | hir::Expr::Nil(..) => {}
        hir::Expr::Var { .. } => {}
        hir::Expr::If {
            test,
            then_branch,
            else_branch,
            ..
        } => {
            collect_builtin_calls_expr(test, names);
            collect_builtin_calls_expr(then_branch, names);
            collect_builtin_calls_expr(else_branch, names);
        }
        hir::Expr::Let { bindings, body, .. } => {
            for binding in bindings {
                collect_builtin_calls_expr(&binding.value, names);
            }
            for expr in body {
                collect_builtin_calls_expr(expr, names);
            }
        }
        hir::Expr::Call { func, args, .. } => {
            if let hir::Expr::Var { name, .. } = func.as_ref()
                && builtins::is_builtin(name)
            {
                names.push(name.clone());
            }
            collect_builtin_calls_expr(func, names);
            for arg in args {
                collect_builtin_calls_expr(arg, names);
            }
        }
        hir::Expr::Lambda { body, .. } => {
            for expr in body {
                collect_builtin_calls_expr(expr, names);
            }
        }
        hir::Expr::FieldAccess { object, .. } => {
            collect_builtin_calls_expr(object, names);
        }
        hir::Expr::RecordConstructor { args, .. } => {
            for arg in args {
                collect_builtin_calls_expr(arg, names);
            }
        }
        hir::Expr::Match {
            scrutinee, clauses, ..
        } => {
            collect_builtin_calls_expr(scrutinee, names);
            for clause in clauses {
                collect_builtin_calls_expr(&clause.body, names);
            }
        }
        hir::Expr::VariantConstructor { args, .. } => {
            for arg in args {
                collect_builtin_calls_expr(arg, names);
            }
        }
    }
}

fn collect_builtin_calls_top_form(form: &hir::TopForm, names: &mut Vec<String>) {
    match form {
        hir::TopForm::Defn { body, .. } => {
            for expr in body {
                collect_builtin_calls_expr(expr, names);
            }
        }
        hir::TopForm::Def { value, .. } => {
            collect_builtin_calls_expr(value, names);
        }
        hir::TopForm::Deftype { .. } => {}
        hir::TopForm::Defunion { .. } => {}
        hir::TopForm::Expr(expr) => {
            collect_builtin_calls_expr(expr, names);
        }
    }
}

pub fn generate_go_mod() -> String {
    "module vex_out\n\ngo 1.21\n".to_string()
}

pub fn needs_vexrt(module: &hir::Module) -> bool {
    fn check_type(ty: &VexType) -> bool {
        matches!(
            ty,
            VexType::Option(_) | VexType::Result { .. } | VexType::List(_) | VexType::Map { .. }
        )
    }

    fn check_expr(expr: &hir::Expr) -> bool {
        match expr {
            hir::Expr::VariantConstructor { union_name, .. } => {
                union_name == "Option" || union_name == "Result"
            }
            hir::Expr::Match { clauses, .. } => clauses.iter().any(|c| {
                matches!(
                    &c.pattern,
                    hir::Pattern::Constructor { union_name, .. }
                    if union_name == "Option" || union_name == "Result"
                )
            }),
            hir::Expr::Call { func, args, .. } => {
                if let hir::Expr::Var { name, .. } = func.as_ref()
                    && name == "range"
                {
                    return true;
                }
                args.iter().any(check_expr)
            }
            hir::Expr::If {
                test,
                then_branch,
                else_branch,
                ..
            } => check_expr(test) || check_expr(then_branch) || check_expr(else_branch),
            hir::Expr::Let { bindings, body, .. } => {
                bindings.iter().any(|b| check_expr(&b.value)) || body.iter().any(check_expr)
            }
            hir::Expr::Lambda { body, .. } => body.iter().any(check_expr),
            _ => false,
        }
    }

    for form in &module.top_forms {
        match form {
            hir::TopForm::Defn {
                params,
                return_type,
                body,
                ..
            } => {
                if check_type(return_type) || params.iter().any(|p| check_type(&p.ty)) {
                    return true;
                }
                if body.iter().any(check_expr) {
                    return true;
                }
            }
            hir::TopForm::Def { ty, value, .. } => {
                if check_type(ty) || check_expr(value) {
                    return true;
                }
            }
            hir::TopForm::Expr(expr) => {
                if check_expr(expr) {
                    return true;
                }
            }
            _ => {}
        }
    }
    false
}

pub fn generate_vexrt_option() -> String {
    r#"package vexrt

type Option[T any] struct {
	IsSome bool
	Value  T
}

func Some[T any](v T) Option[T] {
	return Option[T]{IsSome: true, Value: v}
}

func None[T any]() Option[T] {
	return Option[T]{}
}
"#
    .to_string()
}

pub fn generate_vexrt_result() -> String {
    r#"package vexrt

type Result[T any, E any] struct {
	IsOk  bool
	Value T
	Error E
}

func Ok[T any, E any](v T) Result[T, E] {
	return Result[T, E]{IsOk: true, Value: v}
}

func Err[T any, E any](e E) Result[T, E] {
	return Result[T, E]{Error: e}
}
"#
    .to_string()
}

pub fn generate_vexrt_collections() -> String {
    r#"package vexrt

func Range(start, end int64) []int64 {
	if end <= start {
		return nil
	}
	result := make([]int64, 0, end-start)
	for i := start; i < end; i++ {
		result = append(result, i)
	}
	return result
}
"#
    .to_string()
}

pub fn go_type(ty: &VexType) -> String {
    match ty {
        VexType::Int => "int64".to_string(),
        VexType::Float => "float64".to_string(),
        VexType::Bool => "bool".to_string(),
        VexType::String => "string".to_string(),
        VexType::Unit => "".to_string(),
        VexType::Fn { params, ret } => {
            let param_strs: Vec<String> = params.iter().map(go_type).collect();
            let ret_str = go_type(ret);
            if ret_str.is_empty() {
                format!("func({})", param_strs.join(", "))
            } else {
                format!("func({}) {}", param_strs.join(", "), ret_str)
            }
        }
        VexType::Record { name, .. } => vex_to_go_public_name(name),
        VexType::Union { name, .. } => vex_to_go_public_name(name),
        VexType::List(inner) => format!("[]{}", go_type(inner)),
        VexType::Map { key, value } => format!("map[{}]{}", go_type(key), go_type(value)),
        VexType::Option(inner) => format!("vexrt.Option[{}]", go_type(inner)),
        VexType::Result { ok, err } => {
            format!("vexrt.Result[{}, {}]", go_type(ok), go_type(err))
        }
        VexType::TypeVar(id) => format!("T{}", id),
    }
}

pub fn vex_to_go_name(name: &str) -> String {
    let mut result = String::new();
    let mut capitalize_next = false;

    for (i, ch) in name.chars().enumerate() {
        if ch == '-' || ch == '_' {
            capitalize_next = true;
        } else if i == 0 {
            result.push(ch.to_ascii_lowercase());
        } else if capitalize_next {
            result.push(ch.to_ascii_uppercase());
            capitalize_next = false;
        } else {
            result.push(ch);
        }
    }

    result
}

pub fn vex_to_go_public_name(name: &str) -> String {
    let mut result = String::new();
    let mut capitalize_next = true;

    for ch in name.chars() {
        if ch == '-' || ch == '_' {
            capitalize_next = true;
        } else if capitalize_next {
            result.push(ch.to_ascii_uppercase());
            capitalize_next = false;
        } else {
            result.push(ch);
        }
    }

    result
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::source::{FileId, Span};

    fn span(start: u32, end: u32) -> Span {
        Span::new(FileId::new(0), start, end)
    }

    #[test]
    fn go_type_primitives() {
        assert_eq!(go_type(&VexType::Int), "int64");
        assert_eq!(go_type(&VexType::Float), "float64");
        assert_eq!(go_type(&VexType::Bool), "bool");
        assert_eq!(go_type(&VexType::String), "string");
        assert_eq!(go_type(&VexType::Unit), "");
    }

    #[test]
    fn go_type_list() {
        assert_eq!(go_type(&VexType::List(Box::new(VexType::Int))), "[]int64");
        assert_eq!(
            go_type(&VexType::List(Box::new(VexType::String))),
            "[]string"
        );
    }

    #[test]
    fn go_type_list_nested() {
        assert_eq!(
            go_type(&VexType::List(Box::new(VexType::List(Box::new(
                VexType::Int
            ))))),
            "[][]int64"
        );
    }

    #[test]
    fn go_type_map() {
        assert_eq!(
            go_type(&VexType::Map {
                key: Box::new(VexType::String),
                value: Box::new(VexType::Int),
            }),
            "map[string]int64"
        );
    }

    #[test]
    fn go_type_function() {
        let fn_ty = VexType::Fn {
            params: vec![VexType::Int, VexType::String],
            ret: Box::new(VexType::Bool),
        };
        assert_eq!(go_type(&fn_ty), "func(int64, string) bool");
    }

    #[test]
    fn go_type_function_unit_return() {
        let fn_ty = VexType::Fn {
            params: vec![VexType::String],
            ret: Box::new(VexType::Unit),
        };
        assert_eq!(go_type(&fn_ty), "func(string)");
    }

    #[test]
    fn go_type_function_no_params() {
        let fn_ty = VexType::Fn {
            params: vec![],
            ret: Box::new(VexType::Int),
        };
        assert_eq!(go_type(&fn_ty), "func() int64");
    }

    #[test]
    fn naming_simple() {
        assert_eq!(vex_to_go_name("x"), "x");
        assert_eq!(vex_to_go_name("count"), "count");
    }

    #[test]
    fn naming_kebab_case() {
        assert_eq!(vex_to_go_name("my-var"), "myVar");
        assert_eq!(vex_to_go_name("handle-tool-call"), "handleToolCall");
    }

    #[test]
    fn naming_public() {
        assert_eq!(vex_to_go_public_name("main"), "Main");
        assert_eq!(vex_to_go_public_name("handle-tool-call"), "HandleToolCall");
    }

    #[test]
    fn emit_int() {
        let module = hir::Module {
            top_forms: vec![hir::TopForm::Expr(hir::Expr::Int(42, span(0, 2)))],
        };
        let output = generate(&module);
        assert!(output.contains("package main"));
    }

    #[test]
    fn expr_int() {
        let mut cg = Generator::new();
        cg.emit_expr(&hir::Expr::Int(42, span(0, 2)));
        assert_eq!(cg.output, "42");
    }

    #[test]
    fn expr_float() {
        let mut cg = Generator::new();
        cg.emit_expr(&hir::Expr::Float(3.14, span(0, 4)));
        assert_eq!(cg.output, "3.14");
    }

    #[test]
    fn expr_float_whole() {
        let mut cg = Generator::new();
        cg.emit_expr(&hir::Expr::Float(1.0, span(0, 3)));
        assert_eq!(cg.output, "1.0");
    }

    #[test]
    fn expr_string() {
        let mut cg = Generator::new();
        cg.emit_expr(&hir::Expr::String("hello".into(), span(0, 7)));
        assert_eq!(cg.output, "\"hello\"");
    }

    #[test]
    fn expr_bool() {
        let mut cg = Generator::new();
        cg.emit_expr(&hir::Expr::Bool(true, span(0, 4)));
        assert_eq!(cg.output, "true");

        let mut cg2 = Generator::new();
        cg2.emit_expr(&hir::Expr::Bool(false, span(0, 5)));
        assert_eq!(cg2.output, "false");
    }

    #[test]
    fn expr_var() {
        let mut cg = Generator::new();
        cg.emit_expr(&hir::Expr::Var {
            name: "my-count".into(),
            span: span(0, 8),
            ty: VexType::Int,
        });
        assert_eq!(cg.output, "myCount");
    }

    #[test]
    fn expr_if_value() {
        let mut cg = Generator::new();
        cg.emit_expr(&hir::Expr::If {
            test: Box::new(hir::Expr::Bool(true, span(0, 4))),
            then_branch: Box::new(hir::Expr::Int(1, span(5, 6))),
            else_branch: Box::new(hir::Expr::Int(2, span(7, 8))),
            span: span(0, 9),
            ty: VexType::Int,
        });
        assert_eq!(
            cg.output,
            "func() int64 { if true { return 1 } else { return 2 } }()"
        );
    }

    #[test]
    fn expr_if_unit() {
        let mut cg = Generator::new();
        cg.emit_expr(&hir::Expr::If {
            test: Box::new(hir::Expr::Bool(true, span(0, 4))),
            then_branch: Box::new(hir::Expr::Nil(span(5, 8))),
            else_branch: Box::new(hir::Expr::Nil(span(9, 12))),
            span: span(0, 13),
            ty: VexType::Unit,
        });
        assert!(cg.output.starts_with("if true {\n"));
        assert!(cg.output.contains("} else {\n"));
        assert!(cg.output.ends_with("}"));
    }

    #[test]
    fn expr_call_infix() {
        let mut cg = Generator::new();
        cg.emit_expr(&hir::Expr::Call {
            func: Box::new(hir::Expr::Var {
                name: "+".into(),
                span: span(1, 2),
                ty: VexType::Fn {
                    params: vec![VexType::Int, VexType::Int],
                    ret: Box::new(VexType::Int),
                },
            }),
            args: vec![hir::Expr::Int(1, span(3, 4)), hir::Expr::Int(2, span(5, 6))],
            span: span(0, 7),
            ty: VexType::Int,
        });
        assert_eq!(cg.output, "(1 + 2)");
    }

    #[test]
    fn expr_call_prefix() {
        let mut cg = Generator::new();
        cg.emit_expr(&hir::Expr::Call {
            func: Box::new(hir::Expr::Var {
                name: "not".into(),
                span: span(1, 4),
                ty: VexType::Fn {
                    params: vec![VexType::Bool],
                    ret: Box::new(VexType::Bool),
                },
            }),
            args: vec![hir::Expr::Bool(true, span(5, 9))],
            span: span(0, 10),
            ty: VexType::Bool,
        });
        assert_eq!(cg.output, "!true");
    }

    #[test]
    fn expr_call_func() {
        let mut cg = Generator::new();
        cg.emit_expr(&hir::Expr::Call {
            func: Box::new(hir::Expr::Var {
                name: "println".into(),
                span: span(1, 8),
                ty: VexType::Fn {
                    params: vec![VexType::String],
                    ret: Box::new(VexType::Unit),
                },
            }),
            args: vec![hir::Expr::String("hello".into(), span(9, 16))],
            span: span(0, 17),
            ty: VexType::Unit,
        });
        assert_eq!(cg.output, "fmt.Println(\"hello\")");
    }

    #[test]
    fn expr_call_user_func() {
        let mut cg = Generator::new();
        cg.emit_expr(&hir::Expr::Call {
            func: Box::new(hir::Expr::Var {
                name: "my-func".into(),
                span: span(1, 8),
                ty: VexType::Fn {
                    params: vec![VexType::Int],
                    ret: Box::new(VexType::Int),
                },
            }),
            args: vec![hir::Expr::Int(42, span(9, 11))],
            span: span(0, 12),
            ty: VexType::Int,
        });
        assert_eq!(cg.output, "myFunc(42)");
    }

    #[test]
    fn expr_call_nested() {
        let mut cg = Generator::new();
        cg.emit_expr(&hir::Expr::Call {
            func: Box::new(hir::Expr::Var {
                name: "+".into(),
                span: span(1, 2),
                ty: VexType::Fn {
                    params: vec![VexType::Int, VexType::Int],
                    ret: Box::new(VexType::Int),
                },
            }),
            args: vec![
                hir::Expr::Call {
                    func: Box::new(hir::Expr::Var {
                        name: "*".into(),
                        span: span(4, 5),
                        ty: VexType::Fn {
                            params: vec![VexType::Int, VexType::Int],
                            ret: Box::new(VexType::Int),
                        },
                    }),
                    args: vec![hir::Expr::Int(2, span(6, 7)), hir::Expr::Int(3, span(8, 9))],
                    span: span(3, 10),
                    ty: VexType::Int,
                },
                hir::Expr::Int(1, span(11, 12)),
            ],
            span: span(0, 13),
            ty: VexType::Int,
        });
        assert_eq!(cg.output, "((2 * 3) + 1)");
    }

    #[test]
    fn expr_let_simple() {
        let mut cg = Generator::new();
        cg.emit_expr(&hir::Expr::Let {
            bindings: vec![hir::Binding {
                name: "x".into(),
                ty: VexType::Int,
                value: hir::Expr::Int(42, span(6, 8)),
                span: span(5, 8),
            }],
            body: vec![hir::Expr::Var {
                name: "x".into(),
                span: span(10, 11),
                ty: VexType::Int,
            }],
            span: span(0, 12),
            ty: VexType::Int,
        });
        assert!(cg.output.contains("x := 42"));
        assert!(cg.output.contains("return x"));
        assert!(cg.output.starts_with("func() int64 {"));
    }

    #[test]
    fn expr_lambda() {
        let mut cg = Generator::new();
        cg.emit_expr(&hir::Expr::Lambda {
            params: vec![hir::Param {
                name: "x".into(),
                ty: VexType::Int,
                span: span(5, 6),
            }],
            return_type: VexType::Int,
            body: vec![hir::Expr::Call {
                func: Box::new(hir::Expr::Var {
                    name: "+".into(),
                    span: span(8, 9),
                    ty: VexType::Fn {
                        params: vec![VexType::Int, VexType::Int],
                        ret: Box::new(VexType::Int),
                    },
                }),
                args: vec![
                    hir::Expr::Var {
                        name: "x".into(),
                        span: span(10, 11),
                        ty: VexType::Int,
                    },
                    hir::Expr::Int(1, span(12, 13)),
                ],
                span: span(7, 14),
                ty: VexType::Int,
            }],
            span: span(0, 15),
            ty: VexType::Fn {
                params: vec![VexType::Int],
                ret: Box::new(VexType::Int),
            },
        });
        assert!(cg.output.contains("func(x int64) int64 {"));
        assert!(cg.output.contains("return (x + 1)"));
    }

    #[test]
    fn expr_var_builtin_skipped() {
        let mut cg = Generator::new();
        cg.emit_expr(&hir::Expr::Var {
            name: "+".into(),
            span: span(0, 1),
            ty: VexType::Fn {
                params: vec![VexType::Int, VexType::Int],
                ret: Box::new(VexType::Int),
            },
        });
        assert_eq!(cg.output, "");
    }

    #[test]
    fn top_form_defn() {
        let module = hir::Module {
            top_forms: vec![hir::TopForm::Defn {
                name: "add".into(),
                params: vec![
                    hir::Param {
                        name: "x".into(),
                        ty: VexType::Int,
                        span: span(10, 11),
                    },
                    hir::Param {
                        name: "y".into(),
                        ty: VexType::Int,
                        span: span(18, 19),
                    },
                ],
                return_type: VexType::Int,
                body: vec![hir::Expr::Call {
                    func: Box::new(hir::Expr::Var {
                        name: "+".into(),
                        span: span(28, 29),
                        ty: VexType::Fn {
                            params: vec![VexType::Int, VexType::Int],
                            ret: Box::new(VexType::Int),
                        },
                    }),
                    args: vec![
                        hir::Expr::Var {
                            name: "x".into(),
                            span: span(30, 31),
                            ty: VexType::Int,
                        },
                        hir::Expr::Var {
                            name: "y".into(),
                            span: span(32, 33),
                            ty: VexType::Int,
                        },
                    ],
                    span: span(27, 34),
                    ty: VexType::Int,
                }],
                span: span(0, 35),
            }],
        };
        let output = generate(&module);
        assert!(output.contains("func add(x int64, y int64) int64 {"));
        assert!(output.contains("return (x + y)"));
    }

    #[test]
    fn top_form_defn_unit_return() {
        let module = hir::Module {
            top_forms: vec![hir::TopForm::Defn {
                name: "main".into(),
                params: vec![],
                return_type: VexType::Unit,
                body: vec![hir::Expr::Call {
                    func: Box::new(hir::Expr::Var {
                        name: "println".into(),
                        span: span(16, 23),
                        ty: VexType::Fn {
                            params: vec![VexType::String],
                            ret: Box::new(VexType::Unit),
                        },
                    }),
                    args: vec![hir::Expr::String("Hello, World!".into(), span(24, 39))],
                    span: span(15, 40),
                    ty: VexType::Unit,
                }],
                span: span(0, 41),
            }],
        };
        let output = generate(&module);
        assert!(output.contains("func main() {\n"));
        assert!(output.contains("fmt.Println(\"Hello, World!\")"));
    }

    #[test]
    fn top_form_def() {
        let module = hir::Module {
            top_forms: vec![hir::TopForm::Def {
                name: "pi".into(),
                ty: VexType::Float,
                value: hir::Expr::Float(3.14, span(15, 19)),
                span: span(0, 20),
            }],
        };
        let output = generate(&module);
        assert!(output.contains("var pi float64 = 3.14"));
    }

    #[test]
    fn imports_single() {
        let module = hir::Module {
            top_forms: vec![hir::TopForm::Expr(hir::Expr::Call {
                func: Box::new(hir::Expr::Var {
                    name: "println".into(),
                    span: span(1, 8),
                    ty: VexType::Fn {
                        params: vec![VexType::String],
                        ret: Box::new(VexType::Unit),
                    },
                }),
                args: vec![hir::Expr::String("hi".into(), span(9, 13))],
                span: span(0, 14),
                ty: VexType::Unit,
            })],
        };
        let output = generate(&module);
        assert!(output.contains("import \"fmt\""));
    }

    #[test]
    fn imports_none() {
        let module = hir::Module {
            top_forms: vec![hir::TopForm::Expr(hir::Expr::Call {
                func: Box::new(hir::Expr::Var {
                    name: "+".into(),
                    span: span(1, 2),
                    ty: VexType::Fn {
                        params: vec![VexType::Int, VexType::Int],
                        ret: Box::new(VexType::Int),
                    },
                }),
                args: vec![hir::Expr::Int(1, span(3, 4)), hir::Expr::Int(2, span(5, 6))],
                span: span(0, 7),
                ty: VexType::Int,
            })],
        };
        let output = generate(&module);
        assert!(!output.contains("import"));
    }

    #[test]
    fn go_mod() {
        let go_mod = generate_go_mod();
        assert!(go_mod.contains("module vex_out"));
        assert!(go_mod.contains("go 1.21"));
    }

    #[test]
    fn top_form_deftype() {
        let module = hir::Module {
            top_forms: vec![hir::TopForm::Deftype {
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
            }],
        };
        let output = generate(&module);
        assert!(output.contains("type Point struct {"));
        assert!(output.contains("\tX float64"));
        assert!(output.contains("\tY float64"));
        assert!(output.contains("}"));
    }

    #[test]
    fn top_form_deftype_empty() {
        let module = hir::Module {
            top_forms: vec![hir::TopForm::Deftype {
                name: "Empty".into(),
                fields: vec![],
                span: span(0, 15),
            }],
        };
        let output = generate(&module);
        assert!(output.contains("type Empty struct {"));
        assert!(output.contains("}"));
    }

    #[test]
    fn top_form_deftype_field_naming() {
        let module = hir::Module {
            top_forms: vec![hir::TopForm::Deftype {
                name: "tool-input".into(),
                fields: vec![RecordField {
                    name: "full-name".into(),
                    ty: VexType::String,
                }],
                span: span(0, 40),
            }],
        };
        let output = generate(&module);
        assert!(output.contains("type ToolInput struct {"));
        assert!(output.contains("\tFullName string"));
    }

    #[test]
    fn top_form_defunion() {
        let module = hir::Module {
            top_forms: vec![hir::TopForm::Defunion {
                name: "Msg".into(),
                variants: vec![
                    UnionVariant {
                        name: "Req".into(),
                        types: vec![VexType::Int],
                    },
                    UnionVariant {
                        name: "Resp".into(),
                        types: vec![VexType::String],
                    },
                ],
                span: span(0, 40),
            }],
        };
        let output = generate(&module);
        assert!(output.contains("type Msg interface { isMsg() }"));
        assert!(output.contains("type Msg_Req struct {\n\tV0 int64\n}"));
        assert!(output.contains("func (Msg_Req) isMsg() {}"));
        assert!(output.contains("type Msg_Resp struct {\n\tV0 string\n}"));
        assert!(output.contains("func (Msg_Resp) isMsg() {}"));
    }

    #[test]
    fn top_form_defunion_no_data() {
        let module = hir::Module {
            top_forms: vec![hir::TopForm::Defunion {
                name: "Option".into(),
                variants: vec![
                    UnionVariant {
                        name: "Some".into(),
                        types: vec![VexType::Int],
                    },
                    UnionVariant {
                        name: "None".into(),
                        types: vec![],
                    },
                ],
                span: span(0, 30),
            }],
        };
        let output = generate(&module);
        assert!(output.contains("type Option interface { isOption() }"));
        assert!(output.contains("type Option_Some struct {\n\tV0 int64\n}"));
        assert!(output.contains("type Option_None struct {}"));
        assert!(output.contains("func (Option_None) isOption() {}"));
    }

    #[test]
    fn top_form_defunion_multi_field() {
        let module = hir::Module {
            top_forms: vec![hir::TopForm::Defunion {
                name: "Shape".into(),
                variants: vec![UnionVariant {
                    name: "Rect".into(),
                    types: vec![VexType::Float, VexType::Float],
                }],
                span: span(0, 30),
            }],
        };
        let output = generate(&module);
        assert!(output.contains("type Shape_Rect struct {\n\tV0 float64\n\tV1 float64\n}"));
    }

    #[test]
    fn hello_world_full() {
        let module = hir::Module {
            top_forms: vec![hir::TopForm::Defn {
                name: "main".into(),
                params: vec![],
                return_type: VexType::Unit,
                body: vec![hir::Expr::Call {
                    func: Box::new(hir::Expr::Var {
                        name: "println".into(),
                        span: span(16, 23),
                        ty: VexType::Fn {
                            params: vec![VexType::String],
                            ret: Box::new(VexType::Unit),
                        },
                    }),
                    args: vec![hir::Expr::String("Hello, World!".into(), span(24, 39))],
                    span: span(15, 40),
                    ty: VexType::Unit,
                }],
                span: span(0, 41),
            }],
        };
        let output = generate(&module);
        assert!(output.contains("package main"));
        assert!(output.contains("import \"fmt\""));
        assert!(output.contains("func main() {\n"));
        assert!(output.contains("fmt.Println(\"Hello, World!\")"));
    }

    #[test]
    fn expr_field_access() {
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
        let module = hir::Module {
            top_forms: vec![hir::TopForm::Defn {
                name: "get-x".into(),
                params: vec![hir::Param {
                    name: "p".into(),
                    ty: point_ty.clone(),
                    span: span(14, 24),
                }],
                return_type: VexType::Float,
                body: vec![hir::Expr::FieldAccess {
                    object: Box::new(hir::Expr::Var {
                        name: "p".into(),
                        span: span(40, 41),
                        ty: point_ty,
                    }),
                    field: "x".into(),
                    span: span(38, 44),
                    ty: VexType::Float,
                }],
                span: span(0, 45),
            }],
        };
        let output = generate(&module);
        assert!(output.contains("p.X"));
    }

    #[test]
    fn expr_field_access_kebab_case() {
        let rec_ty = VexType::Record {
            name: "Config".into(),
            fields: vec![RecordField {
                name: "max-retries".into(),
                ty: VexType::Int,
            }],
        };
        let module = hir::Module {
            top_forms: vec![hir::TopForm::Expr(hir::Expr::FieldAccess {
                object: Box::new(hir::Expr::Var {
                    name: "cfg".into(),
                    span: span(3, 6),
                    ty: rec_ty,
                }),
                field: "max-retries".into(),
                span: span(0, 20),
                ty: VexType::Int,
            })],
        };
        let output = generate(&module);
        assert!(output.contains("cfg.MaxRetries"));
    }

    #[test]
    fn expr_record_constructor() {
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
        let module = hir::Module {
            top_forms: vec![hir::TopForm::Expr(hir::Expr::RecordConstructor {
                name: "Point".into(),
                args: vec![
                    hir::Expr::Float(1.0, span(7, 10)),
                    hir::Expr::Float(2.0, span(11, 14)),
                ],
                span: span(0, 15),
                ty: point_ty,
            })],
        };
        let output = generate(&module);
        assert!(output.contains("Point{X: 1.0, Y: 2.0}"));
    }

    #[test]
    fn expr_record_constructor_kebab_names() {
        let rec_ty = VexType::Record {
            name: "tool-input".into(),
            fields: vec![RecordField {
                name: "full-name".into(),
                ty: VexType::String,
            }],
        };
        let module = hir::Module {
            top_forms: vec![hir::TopForm::Expr(hir::Expr::RecordConstructor {
                name: "tool-input".into(),
                args: vec![hir::Expr::String("test".into(), span(12, 18))],
                span: span(0, 19),
                ty: rec_ty,
            })],
        };
        let output = generate(&module);
        assert!(output.contains("ToolInput{FullName: \"test\"}"));
    }

    fn wrap_in_defn(body: hir::Expr) -> hir::Module {
        hir::Module {
            top_forms: vec![hir::TopForm::Defn {
                name: "test".into(),
                params: vec![],
                return_type: body.ty().clone(),
                body: vec![body],
                span: span(0, 100),
            }],
        }
    }

    #[test]
    fn expr_match_type_switch() {
        let union_ty = VexType::Union {
            name: "Option".into(),
            variants: vec![
                UnionVariant {
                    name: "Some".into(),
                    types: vec![VexType::Int],
                },
                UnionVariant {
                    name: "None".into(),
                    types: vec![],
                },
            ],
        };
        let expr = hir::Expr::Match {
            scrutinee: Box::new(hir::Expr::Var {
                name: "o".into(),
                span: span(0, 1),
                ty: union_ty,
            }),
            clauses: vec![
                hir::MatchClause {
                    pattern: hir::Pattern::Constructor {
                        union_name: "Option".into(),
                        variant_name: "Some".into(),
                        bindings: vec![hir::Pattern::Binding {
                            name: "x".into(),
                            ty: VexType::Int,
                            span: span(10, 11),
                        }],
                        span: span(5, 12),
                    },
                    body: hir::Expr::Var {
                        name: "x".into(),
                        span: span(13, 14),
                        ty: VexType::Int,
                    },
                    span: span(5, 14),
                },
                hir::MatchClause {
                    pattern: hir::Pattern::Wildcard(span(15, 16)),
                    body: hir::Expr::Int(0, span(17, 18)),
                    span: span(15, 18),
                },
            ],
            span: span(0, 19),
            ty: VexType::Int,
        };
        let output = generate(&wrap_in_defn(expr));
        assert!(output.contains("switch _v := o.(type)"), "{}", output);
        assert!(output.contains("case Option_Some:"), "{}", output);
        assert!(output.contains("x := _v.V0"), "{}", output);
        assert!(output.contains("return x"), "{}", output);
        assert!(output.contains("default:"), "{}", output);
        assert!(output.contains("return 0"), "{}", output);
    }

    #[test]
    fn expr_match_value_switch() {
        let expr = hir::Expr::Match {
            scrutinee: Box::new(hir::Expr::Var {
                name: "x".into(),
                span: span(0, 1),
                ty: VexType::Int,
            }),
            clauses: vec![
                hir::MatchClause {
                    pattern: hir::Pattern::Literal(Box::new(hir::Expr::Int(1, span(5, 6)))),
                    body: hir::Expr::String("one".into(), span(7, 12)),
                    span: span(5, 12),
                },
                hir::MatchClause {
                    pattern: hir::Pattern::Wildcard(span(13, 14)),
                    body: hir::Expr::String("other".into(), span(15, 22)),
                    span: span(13, 22),
                },
            ],
            span: span(0, 23),
            ty: VexType::String,
        };
        let output = generate(&wrap_in_defn(expr));
        assert!(output.contains("switch x {"), "{}", output);
        assert!(output.contains("case 1:"), "{}", output);
        assert!(output.contains("return \"one\""), "{}", output);
        assert!(output.contains("default:"), "{}", output);
        assert!(output.contains("return \"other\""), "{}", output);
    }

    #[test]
    fn expr_option_some_constructor() {
        let expr = hir::Expr::VariantConstructor {
            union_name: "Option".into(),
            variant_name: "Some".into(),
            args: vec![hir::Expr::Int(42, span(5, 7))],
            span: span(0, 8),
            ty: VexType::Option(Box::new(VexType::Int)),
        };
        let output = generate(&wrap_in_defn(expr));
        assert!(output.contains("vexrt.Some[int64](42)"), "{}", output);
    }

    #[test]
    fn expr_option_none_constructor() {
        let expr = hir::Expr::VariantConstructor {
            union_name: "Option".into(),
            variant_name: "None".into(),
            args: vec![],
            span: span(0, 4),
            ty: VexType::Option(Box::new(VexType::Int)),
        };
        let output = generate(&wrap_in_defn(expr));
        assert!(output.contains("vexrt.None[int64]()"), "{}", output);
    }

    #[test]
    fn expr_result_ok_constructor() {
        let expr = hir::Expr::VariantConstructor {
            union_name: "Result".into(),
            variant_name: "Ok".into(),
            args: vec![hir::Expr::Int(1, span(3, 4))],
            span: span(0, 5),
            ty: VexType::Result {
                ok: Box::new(VexType::Int),
                err: Box::new(VexType::String),
            },
        };
        let output = generate(&wrap_in_defn(expr));
        assert!(output.contains("vexrt.Ok[int64, string](1)"), "{}", output);
    }

    #[test]
    fn expr_result_err_constructor() {
        let expr = hir::Expr::VariantConstructor {
            union_name: "Result".into(),
            variant_name: "Err".into(),
            args: vec![hir::Expr::String("bad".into(), span(4, 9))],
            span: span(0, 10),
            ty: VexType::Result {
                ok: Box::new(VexType::Int),
                err: Box::new(VexType::String),
            },
        };
        let output = generate(&wrap_in_defn(expr));
        assert!(
            output.contains("vexrt.Err[int64, string](\"bad\")"),
            "{}",
            output
        );
    }

    #[test]
    fn expr_match_option() {
        let option_ty = VexType::Option(Box::new(VexType::Int));
        let expr = hir::Expr::Match {
            scrutinee: Box::new(hir::Expr::Var {
                name: "o".into(),
                span: span(0, 1),
                ty: option_ty,
            }),
            clauses: vec![
                hir::MatchClause {
                    pattern: hir::Pattern::Constructor {
                        union_name: "Option".into(),
                        variant_name: "Some".into(),
                        bindings: vec![hir::Pattern::Binding {
                            name: "x".into(),
                            ty: VexType::Int,
                            span: span(10, 11),
                        }],
                        span: span(5, 12),
                    },
                    body: hir::Expr::Var {
                        name: "x".into(),
                        span: span(13, 14),
                        ty: VexType::Int,
                    },
                    span: span(5, 14),
                },
                hir::MatchClause {
                    pattern: hir::Pattern::Constructor {
                        union_name: "Option".into(),
                        variant_name: "None".into(),
                        bindings: vec![],
                        span: span(15, 19),
                    },
                    body: hir::Expr::Int(0, span(20, 21)),
                    span: span(15, 21),
                },
            ],
            span: span(0, 22),
            ty: VexType::Int,
        };
        let output = generate(&wrap_in_defn(expr));
        assert!(output.contains("if o.IsSome {"), "{}", output);
        assert!(output.contains("x := o.Value"), "{}", output);
        assert!(output.contains("return x"), "{}", output);
        assert!(output.contains("} else {"), "{}", output);
        assert!(output.contains("return 0"), "{}", output);
    }

    #[test]
    fn expr_match_result() {
        let result_ty = VexType::Result {
            ok: Box::new(VexType::Int),
            err: Box::new(VexType::String),
        };
        let expr = hir::Expr::Match {
            scrutinee: Box::new(hir::Expr::Var {
                name: "r".into(),
                span: span(0, 1),
                ty: result_ty,
            }),
            clauses: vec![
                hir::MatchClause {
                    pattern: hir::Pattern::Constructor {
                        union_name: "Result".into(),
                        variant_name: "Ok".into(),
                        bindings: vec![hir::Pattern::Binding {
                            name: "v".into(),
                            ty: VexType::Int,
                            span: span(10, 11),
                        }],
                        span: span(5, 12),
                    },
                    body: hir::Expr::Var {
                        name: "v".into(),
                        span: span(13, 14),
                        ty: VexType::Int,
                    },
                    span: span(5, 14),
                },
                hir::MatchClause {
                    pattern: hir::Pattern::Constructor {
                        union_name: "Result".into(),
                        variant_name: "Err".into(),
                        bindings: vec![hir::Pattern::Binding {
                            name: "e".into(),
                            ty: VexType::String,
                            span: span(20, 21),
                        }],
                        span: span(15, 22),
                    },
                    body: hir::Expr::Int(-1, span(23, 25)),
                    span: span(15, 25),
                },
            ],
            span: span(0, 26),
            ty: VexType::Int,
        };
        let output = generate(&wrap_in_defn(expr));
        assert!(output.contains("if r.IsOk {"), "{}", output);
        assert!(output.contains("v := r.Value"), "{}", output);
        assert!(output.contains("return v"), "{}", output);
        assert!(output.contains("} else {"), "{}", output);
        assert!(output.contains("e := r.Error"), "{}", output);
        assert!(output.contains("return -1"), "{}", output);
    }

    #[test]
    fn expr_range_call() {
        let module = hir::Module {
            top_forms: vec![hir::TopForm::Expr(hir::Expr::Call {
                func: Box::new(hir::Expr::Var {
                    name: "range".into(),
                    span: span(1, 6),
                    ty: VexType::Fn {
                        params: vec![VexType::Int, VexType::Int],
                        ret: Box::new(VexType::List(Box::new(VexType::Int))),
                    },
                }),
                args: vec![
                    hir::Expr::Int(0, span(7, 8)),
                    hir::Expr::Int(10, span(9, 11)),
                ],
                span: span(0, 12),
                ty: VexType::List(Box::new(VexType::Int)),
            })],
        };
        let output = generate(&module);
        assert!(output.contains("vexrt.Range(0, 10)"), "{}", output);
        assert!(output.contains("\"vex_out/vexrt\""), "{}", output);
    }

    #[test]
    fn needs_vexrt_true_for_range() {
        let module = hir::Module {
            top_forms: vec![hir::TopForm::Expr(hir::Expr::Call {
                func: Box::new(hir::Expr::Var {
                    name: "range".into(),
                    span: span(1, 6),
                    ty: VexType::Fn {
                        params: vec![VexType::Int, VexType::Int],
                        ret: Box::new(VexType::List(Box::new(VexType::Int))),
                    },
                }),
                args: vec![
                    hir::Expr::Int(0, span(7, 8)),
                    hir::Expr::Int(10, span(9, 11)),
                ],
                span: span(0, 12),
                ty: VexType::List(Box::new(VexType::Int)),
            })],
        };
        assert!(needs_vexrt(&module));
    }

    #[test]
    fn needs_vexrt_true_for_list_param() {
        let module = hir::Module {
            top_forms: vec![hir::TopForm::Defn {
                name: "f".into(),
                params: vec![hir::Param {
                    name: "xs".into(),
                    ty: VexType::List(Box::new(VexType::Int)),
                    span: span(0, 5),
                }],
                return_type: VexType::Unit,
                body: vec![hir::Expr::Nil(span(10, 13))],
                span: span(0, 14),
            }],
        };
        assert!(needs_vexrt(&module));
    }

    #[test]
    fn expr_each_with_lambda() {
        let mut cg = Generator::new();
        let list_expr = hir::Expr::Call {
            func: Box::new(hir::Expr::Var {
                name: "range".into(),
                span: span(6, 11),
                ty: VexType::Fn {
                    params: vec![VexType::Int, VexType::Int],
                    ret: Box::new(VexType::List(Box::new(VexType::Int))),
                },
            }),
            args: vec![
                hir::Expr::Int(0, span(12, 13)),
                hir::Expr::Int(10, span(14, 16)),
            ],
            span: span(5, 17),
            ty: VexType::List(Box::new(VexType::Int)),
        };
        let callback = hir::Expr::Lambda {
            params: vec![hir::Param {
                name: "x".into(),
                ty: VexType::Int,
                span: span(23, 24),
            }],
            return_type: VexType::Unit,
            body: vec![hir::Expr::Call {
                func: Box::new(hir::Expr::Var {
                    name: "println".into(),
                    span: span(30, 37),
                    ty: VexType::Fn {
                        params: vec![VexType::String],
                        ret: Box::new(VexType::Unit),
                    },
                }),
                args: vec![hir::Expr::Call {
                    func: Box::new(hir::Expr::Var {
                        name: "str".into(),
                        span: span(39, 42),
                        ty: VexType::Fn {
                            params: vec![],
                            ret: Box::new(VexType::String),
                        },
                    }),
                    args: vec![hir::Expr::Var {
                        name: "x".into(),
                        span: span(43, 44),
                        ty: VexType::Int,
                    }],
                    span: span(38, 45),
                    ty: VexType::String,
                }],
                span: span(29, 46),
                ty: VexType::Unit,
            }],
            span: span(18, 47),
            ty: VexType::Fn {
                params: vec![VexType::Int],
                ret: Box::new(VexType::Unit),
            },
        };
        cg.emit_each(&list_expr, &callback);
        assert!(
            cg.output.contains("for _, x := range vexrt.Range(0, 10) {"),
            "{}",
            cg.output
        );
        assert!(
            cg.output.contains("fmt.Println(fmt.Sprint(x))"),
            "{}",
            cg.output
        );
    }

    #[test]
    fn expr_each_with_var() {
        let mut cg = Generator::new();
        let list_expr = hir::Expr::Var {
            name: "items".into(),
            span: span(6, 11),
            ty: VexType::List(Box::new(VexType::Int)),
        };
        let callback = hir::Expr::Var {
            name: "process-item".into(),
            span: span(12, 24),
            ty: VexType::Fn {
                params: vec![VexType::Int],
                ret: Box::new(VexType::Unit),
            },
        };
        cg.emit_each(&list_expr, &callback);
        assert!(
            cg.output.contains("for _, _v := range items {"),
            "{}",
            cg.output
        );
        assert!(cg.output.contains("processItem(_v)"), "{}", cg.output);
    }

    #[test]
    fn expr_map_with_lambda() {
        let mut cg = Generator::new();
        let list_expr = hir::Expr::Var {
            name: "xs".into(),
            span: span(5, 7),
            ty: VexType::List(Box::new(VexType::Int)),
        };
        let callback = hir::Expr::Lambda {
            params: vec![hir::Param {
                name: "x".into(),
                ty: VexType::Int,
                span: span(13, 14),
            }],
            return_type: VexType::Int,
            body: vec![hir::Expr::Call {
                func: Box::new(hir::Expr::Var {
                    name: "*".into(),
                    span: span(20, 21),
                    ty: VexType::Fn {
                        params: vec![VexType::Int, VexType::Int],
                        ret: Box::new(VexType::Int),
                    },
                }),
                args: vec![
                    hir::Expr::Var {
                        name: "x".into(),
                        span: span(22, 23),
                        ty: VexType::Int,
                    },
                    hir::Expr::Int(2, span(24, 25)),
                ],
                span: span(19, 26),
                ty: VexType::Int,
            }],
            span: span(8, 27),
            ty: VexType::Fn {
                params: vec![VexType::Int],
                ret: Box::new(VexType::Int),
            },
        };
        let result_ty = VexType::List(Box::new(VexType::Int));
        cg.emit_map(&list_expr, &callback, &result_ty);
        assert!(cg.output.contains("func() []int64 {"), "{}", cg.output);
        assert!(
            cg.output.contains("_result := make([]int64, 0, len(xs))"),
            "{}",
            cg.output
        );
        assert!(
            cg.output.contains("for _, x := range xs {"),
            "{}",
            cg.output
        );
        assert!(
            cg.output.contains("_result = append(_result, (x * 2))"),
            "{}",
            cg.output
        );
        assert!(cg.output.contains("return _result"), "{}", cg.output);
    }

    #[test]
    fn expr_map_with_var() {
        let mut cg = Generator::new();
        let list_expr = hir::Expr::Var {
            name: "xs".into(),
            span: span(5, 7),
            ty: VexType::List(Box::new(VexType::Int)),
        };
        let callback = hir::Expr::Var {
            name: "double".into(),
            span: span(8, 14),
            ty: VexType::Fn {
                params: vec![VexType::Int],
                ret: Box::new(VexType::Int),
            },
        };
        let result_ty = VexType::List(Box::new(VexType::Int));
        cg.emit_map(&list_expr, &callback, &result_ty);
        assert!(
            cg.output.contains("for _, _v := range xs {"),
            "{}",
            cg.output
        );
        assert!(
            cg.output.contains("_result = append(_result, double(_v))"),
            "{}",
            cg.output
        );
    }

    #[test]
    fn expr_filter_with_lambda() {
        let mut cg = Generator::new();
        let list_expr = hir::Expr::Var {
            name: "xs".into(),
            span: span(8, 10),
            ty: VexType::List(Box::new(VexType::Int)),
        };
        let callback = hir::Expr::Lambda {
            params: vec![hir::Param {
                name: "x".into(),
                ty: VexType::Int,
                span: span(16, 17),
            }],
            return_type: VexType::Bool,
            body: vec![hir::Expr::Call {
                func: Box::new(hir::Expr::Var {
                    name: ">".into(),
                    span: span(25, 26),
                    ty: VexType::Fn {
                        params: vec![VexType::Int, VexType::Int],
                        ret: Box::new(VexType::Bool),
                    },
                }),
                args: vec![
                    hir::Expr::Var {
                        name: "x".into(),
                        span: span(27, 28),
                        ty: VexType::Int,
                    },
                    hir::Expr::Int(5, span(29, 30)),
                ],
                span: span(24, 31),
                ty: VexType::Bool,
            }],
            span: span(11, 32),
            ty: VexType::Fn {
                params: vec![VexType::Int],
                ret: Box::new(VexType::Bool),
            },
        };
        let result_ty = VexType::List(Box::new(VexType::Int));
        cg.emit_filter(&list_expr, &callback, &result_ty);
        assert!(cg.output.contains("func() []int64 {"), "{}", cg.output);
        assert!(cg.output.contains("var _result []int64"), "{}", cg.output);
        assert!(
            cg.output.contains("for _, x := range xs {"),
            "{}",
            cg.output
        );
        assert!(cg.output.contains("if (x > 5) {"), "{}", cg.output);
        assert!(
            cg.output.contains("_result = append(_result, x)"),
            "{}",
            cg.output
        );
        assert!(cg.output.contains("return _result"), "{}", cg.output);
    }

    #[test]
    fn expr_filter_with_var() {
        let mut cg = Generator::new();
        let list_expr = hir::Expr::Var {
            name: "xs".into(),
            span: span(8, 10),
            ty: VexType::List(Box::new(VexType::Int)),
        };
        let callback = hir::Expr::Var {
            name: "is-even".into(),
            span: span(11, 18),
            ty: VexType::Fn {
                params: vec![VexType::Int],
                ret: Box::new(VexType::Bool),
            },
        };
        let result_ty = VexType::List(Box::new(VexType::Int));
        cg.emit_filter(&list_expr, &callback, &result_ty);
        assert!(
            cg.output.contains("for _, _v := range xs {"),
            "{}",
            cg.output
        );
        assert!(cg.output.contains("if isEven(_v) {"), "{}", cg.output);
        assert!(
            cg.output.contains("_result = append(_result, _v)"),
            "{}",
            cg.output
        );
    }

    #[test]
    fn needs_vexrt_false_for_basic() {
        let module = hir::Module {
            top_forms: vec![hir::TopForm::Expr(hir::Expr::Int(42, span(0, 2)))],
        };
        assert!(!needs_vexrt(&module));
    }

    #[test]
    fn needs_vexrt_true_for_option_return() {
        let module = hir::Module {
            top_forms: vec![hir::TopForm::Defn {
                name: "f".into(),
                params: vec![],
                return_type: VexType::Option(Box::new(VexType::Int)),
                body: vec![hir::Expr::VariantConstructor {
                    union_name: "Option".into(),
                    variant_name: "None".into(),
                    args: vec![],
                    span: span(0, 4),
                    ty: VexType::Option(Box::new(VexType::Int)),
                }],
                span: span(0, 10),
            }],
        };
        assert!(needs_vexrt(&module));
    }

    #[test]
    fn import_vexrt_when_needed() {
        let module = hir::Module {
            top_forms: vec![hir::TopForm::Defn {
                name: "f".into(),
                params: vec![],
                return_type: VexType::Option(Box::new(VexType::Int)),
                body: vec![hir::Expr::VariantConstructor {
                    union_name: "Option".into(),
                    variant_name: "None".into(),
                    args: vec![],
                    span: span(0, 4),
                    ty: VexType::Option(Box::new(VexType::Int)),
                }],
                span: span(0, 10),
            }],
        };
        let output = generate(&module);
        assert!(output.contains("\"vex_out/vexrt\""), "{}", output);
    }
}
