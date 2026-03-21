use std::fmt::Write;

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

#[allow(dead_code)]
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

        for form in &module.top_forms {
            self.emit_top_form(form);
            self.newline();
        }
    }

    fn emit_top_form(&mut self, _form: &hir::TopForm) {}

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
                if crate::builtins::is_builtin(name) {
                    return;
                }
                self.write(&vex_to_go_name(name));
            }
            _ => {}
        }
    }
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
}
