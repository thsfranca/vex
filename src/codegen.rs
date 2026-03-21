use std::collections::BTreeSet;
use std::fmt::Write;

use crate::builtins;
use crate::builtins::GoTranslation;
use crate::hir;
use crate::types::VexType;

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
}

fn collect_imports(module: &hir::Module) -> BTreeSet<&'static str> {
    let mut names = Vec::new();
    for form in &module.top_forms {
        collect_builtin_calls_top_form(form, &mut names);
    }
    let name_refs: Vec<&str> = names.iter().map(|s| s.as_str()).collect();
    builtins::go_imports(&name_refs).into_iter().collect()
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
        hir::TopForm::Expr(expr) => {
            collect_builtin_calls_expr(expr, names);
        }
    }
}

pub fn generate_go_mod() -> String {
    "module vex_out\n\ngo 1.21\n".to_string()
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
}
