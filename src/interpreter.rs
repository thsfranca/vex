use std::collections::HashMap;
use std::fmt;

use crate::builtins;
use crate::hir;
use crate::types::VexType;

#[derive(Debug, Clone)]
pub struct RuntimeError {
    pub message: String,
}

impl fmt::Display for RuntimeError {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        write!(f, "runtime error: {}", self.message)
    }
}

#[derive(Debug, Clone)]
pub enum Value {
    Int(i64),
    Float(f64),
    Bool(bool),
    String(String),
    Unit,
    List(Vec<Value>),
    Map(Vec<(Value, Value)>),
    Record {
        name: String,
        fields: Vec<(String, Value)>,
    },
    Variant {
        union_name: String,
        variant_name: String,
        values: Vec<Value>,
    },
    Fn {
        params: Vec<String>,
        body: Vec<hir::Expr>,
    },
    BuiltinFn(String),
}

impl fmt::Display for Value {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        match self {
            Value::Int(n) => write!(f, "{}", n),
            Value::Float(n) => write!(f, "{}", n),
            Value::Bool(b) => write!(f, "{}", b),
            Value::String(s) => write!(f, "\"{}\"", s),
            Value::Unit => write!(f, "nil"),
            Value::List(items) => {
                write!(f, "[")?;
                for (i, item) in items.iter().enumerate() {
                    if i > 0 {
                        write!(f, " ")?;
                    }
                    write!(f, "{}", item)?;
                }
                write!(f, "]")
            }
            Value::Map(entries) => {
                write!(f, "{{")?;
                for (i, (k, v)) in entries.iter().enumerate() {
                    if i > 0 {
                        write!(f, " ")?;
                    }
                    write!(f, "{} {}", k, v)?;
                }
                write!(f, "}}")
            }
            Value::Record { name, fields } => {
                write!(f, "({}", name)?;
                for (fname, fval) in fields {
                    write!(f, " :{} {}", fname, fval)?;
                }
                write!(f, ")")
            }
            Value::Variant {
                variant_name,
                values,
                ..
            } => {
                if values.is_empty() {
                    write!(f, "({})", variant_name)
                } else {
                    write!(f, "({}", variant_name)?;
                    for v in values {
                        write!(f, " {}", v)?;
                    }
                    write!(f, ")")
                }
            }
            Value::Fn { .. } => write!(f, "<fn>"),
            Value::BuiltinFn(name) => write!(f, "<builtin:{}>", name),
        }
    }
}

struct Env {
    scopes: Vec<HashMap<String, Value>>,
}

impl Env {
    fn new() -> Self {
        Self {
            scopes: vec![HashMap::new()],
        }
    }

    fn push_scope(&mut self) {
        self.scopes.push(HashMap::new());
    }

    fn pop_scope(&mut self) {
        self.scopes.pop();
    }

    fn define(&mut self, name: String, value: Value) {
        if let Some(scope) = self.scopes.last_mut() {
            scope.insert(name, value);
        }
    }

    fn lookup(&self, name: &str) -> Option<&Value> {
        for scope in self.scopes.iter().rev() {
            if let Some(value) = scope.get(name) {
                return Some(value);
            }
        }
        None
    }
}

pub struct Interpreter {
    env: Env,
}

impl Default for Interpreter {
    fn default() -> Self {
        Self::new()
    }
}

impl Interpreter {
    pub fn new() -> Self {
        let mut env = Env::new();
        for builtin in builtins::all_builtins() {
            env.define(
                builtin.name.to_string(),
                Value::BuiltinFn(builtin.name.to_string()),
            );
        }
        Self { env }
    }

    pub fn eval_module(&mut self, module: &hir::Module) -> Result<Value, RuntimeError> {
        let mut last = Value::Unit;
        for form in &module.top_forms {
            last = self.eval_top_form(form)?;
        }
        Ok(last)
    }

    pub fn eval_top_form(&mut self, form: &hir::TopForm) -> Result<Value, RuntimeError> {
        match form {
            hir::TopForm::Module { .. }
            | hir::TopForm::Export { .. }
            | hir::TopForm::Import { .. }
            | hir::TopForm::ImportGo { .. }
            | hir::TopForm::Deftype { .. }
            | hir::TopForm::Defunion { .. } => Ok(Value::Unit),

            hir::TopForm::Defn {
                name, params, body, ..
            } => {
                let func = Value::Fn {
                    params: params.iter().map(|p| p.name.clone()).collect(),
                    body: body.clone(),
                };
                self.env.define(name.clone(), func);
                Ok(Value::Unit)
            }

            hir::TopForm::Def { name, value, .. } => {
                let val = self.eval_expr(value)?;
                self.env.define(name.clone(), val);
                Ok(Value::Unit)
            }

            hir::TopForm::Expr(expr) => self.eval_expr(expr),
        }
    }

    fn eval_expr(&mut self, expr: &hir::Expr) -> Result<Value, RuntimeError> {
        match expr {
            hir::Expr::Int(n, _) => Ok(Value::Int(*n)),
            hir::Expr::Float(n, _) => Ok(Value::Float(*n)),
            hir::Expr::String(s, _) => Ok(Value::String(s.clone())),
            hir::Expr::Bool(b, _) => Ok(Value::Bool(*b)),
            hir::Expr::Nil(_) => Ok(Value::Unit),

            hir::Expr::Var { name, .. } => {
                self.env.lookup(name).cloned().ok_or_else(|| RuntimeError {
                    message: format!("undefined variable: {}", name),
                })
            }

            hir::Expr::If {
                test,
                then_branch,
                else_branch,
                ..
            } => {
                let cond = self.eval_expr(test)?;
                match cond {
                    Value::Bool(true) => self.eval_expr(then_branch),
                    Value::Bool(false) => self.eval_expr(else_branch),
                    _ => Err(RuntimeError {
                        message: format!("if condition must be Bool, got {}", cond),
                    }),
                }
            }

            hir::Expr::Let { bindings, body, .. } => {
                self.env.push_scope();
                for binding in bindings {
                    let val = self.eval_expr(&binding.value)?;
                    self.env.define(binding.name.clone(), val);
                }
                let mut result = Value::Unit;
                for e in body {
                    result = self.eval_expr(e)?;
                }
                self.env.pop_scope();
                Ok(result)
            }

            hir::Expr::Lambda { params, body, .. } => Ok(Value::Fn {
                params: params.iter().map(|p| p.name.clone()).collect(),
                body: body.clone(),
            }),

            hir::Expr::Call { func, args, .. } => {
                let func_val = self.eval_expr(func)?;
                let mut arg_vals = Vec::new();
                for arg in args {
                    arg_vals.push(self.eval_expr(arg)?);
                }
                self.call_function(func_val, arg_vals)
            }

            hir::Expr::FieldAccess { object, field, .. } => {
                let obj = self.eval_expr(object)?;
                match obj {
                    Value::Record { fields, .. } => fields
                        .iter()
                        .find(|(name, _)| name == field)
                        .map(|(_, val)| val.clone())
                        .ok_or_else(|| RuntimeError {
                            message: format!("no field '{}' on record", field),
                        }),
                    _ => Err(RuntimeError {
                        message: format!("field access on non-record value: {}", obj),
                    }),
                }
            }

            hir::Expr::RecordConstructor { name, args, ty, .. } => {
                let field_names = match ty {
                    VexType::Record { fields, .. } => {
                        fields.iter().map(|f| f.name.clone()).collect::<Vec<_>>()
                    }
                    _ => {
                        return Err(RuntimeError {
                            message: "record constructor with non-record type".into(),
                        });
                    }
                };
                let mut field_values = Vec::new();
                for (i, arg) in args.iter().enumerate() {
                    let val = self.eval_expr(arg)?;
                    field_values.push((field_names[i].clone(), val));
                }
                Ok(Value::Record {
                    name: name.clone(),
                    fields: field_values,
                })
            }

            hir::Expr::Match {
                scrutinee, clauses, ..
            } => {
                let val = self.eval_expr(scrutinee)?;
                for clause in clauses {
                    if let Some(bindings) = match_pattern(&clause.pattern, &val) {
                        self.env.push_scope();
                        for (name, value) in bindings {
                            self.env.define(name, value);
                        }
                        let result = self.eval_expr(&clause.body)?;
                        self.env.pop_scope();
                        return Ok(result);
                    }
                }
                Err(RuntimeError {
                    message: "no matching clause in match expression".into(),
                })
            }

            hir::Expr::VariantConstructor {
                union_name,
                variant_name,
                args,
                ..
            } => {
                let mut values = Vec::new();
                for arg in args {
                    values.push(self.eval_expr(arg)?);
                }
                Ok(Value::Variant {
                    union_name: union_name.clone(),
                    variant_name: variant_name.clone(),
                    values,
                })
            }

            hir::Expr::Spawn { .. }
            | hir::Expr::Channel { .. }
            | hir::Expr::Send { .. }
            | hir::Expr::Recv { .. } => Err(RuntimeError {
                message: "concurrency primitives are not supported in the interpreter".into(),
            }),
        }
    }

    fn call_function(&mut self, func: Value, args: Vec<Value>) -> Result<Value, RuntimeError> {
        match func {
            Value::Fn { params, body } => {
                if params.len() != args.len() {
                    return Err(RuntimeError {
                        message: format!(
                            "function expects {} arguments, got {}",
                            params.len(),
                            args.len()
                        ),
                    });
                }
                self.env.push_scope();
                for (name, val) in params.iter().zip(args) {
                    self.env.define(name.clone(), val);
                }
                let mut result = Value::Unit;
                for expr in &body {
                    result = self.eval_expr(expr)?;
                }
                self.env.pop_scope();
                Ok(result)
            }
            Value::BuiltinFn(name) => self.call_builtin(&name, args),
            _ => Err(RuntimeError {
                message: format!("cannot call non-function value: {}", func),
            }),
        }
    }

    fn call_builtin(&mut self, name: &str, args: Vec<Value>) -> Result<Value, RuntimeError> {
        match name {
            "+" => numeric_binop(&args, |a, b| a + b, |a, b| a + b),
            "-" => numeric_binop(&args, |a, b| a - b, |a, b| a - b),
            "*" => numeric_binop(&args, |a, b| a * b, |a, b| a * b),
            "/" => {
                if let (Value::Int(_), Value::Int(0)) = (&args[0], &args[1]) {
                    return Err(RuntimeError {
                        message: "division by zero".into(),
                    });
                }
                numeric_binop(&args, |a, b| a / b, |a, b| a / b)
            }
            "mod" => match (&args[0], &args[1]) {
                (Value::Int(a), Value::Int(b)) => {
                    if *b == 0 {
                        Err(RuntimeError {
                            message: "modulo by zero".into(),
                        })
                    } else {
                        Ok(Value::Int(a % b))
                    }
                }
                _ => Err(RuntimeError {
                    message: "mod requires Int arguments".into(),
                }),
            },
            "<" => numeric_cmp(&args, |a, b| a < b, |a, b| a < b),
            ">" => numeric_cmp(&args, |a, b| a > b, |a, b| a > b),
            "<=" => numeric_cmp(&args, |a, b| a <= b, |a, b| a <= b),
            ">=" => numeric_cmp(&args, |a, b| a >= b, |a, b| a >= b),
            "=" => numeric_cmp(&args, |a, b| a == b, |a, b| a == b),
            "!=" => numeric_cmp(&args, |a, b| a != b, |a, b| a != b),
            "&&" => match (&args[0], &args[1]) {
                (Value::Bool(a), Value::Bool(b)) => Ok(Value::Bool(*a && *b)),
                _ => Err(RuntimeError {
                    message: "&& requires Bool arguments".into(),
                }),
            },
            "||" => match (&args[0], &args[1]) {
                (Value::Bool(a), Value::Bool(b)) => Ok(Value::Bool(*a || *b)),
                _ => Err(RuntimeError {
                    message: "|| requires Bool arguments".into(),
                }),
            },
            "not" => match &args[0] {
                Value::Bool(b) => Ok(Value::Bool(!b)),
                _ => Err(RuntimeError {
                    message: "not requires Bool argument".into(),
                }),
            },
            "println" => {
                println!("{}", value_to_string(&args[0]));
                Ok(Value::Unit)
            }
            "str" => {
                let mut result = String::new();
                for arg in &args {
                    result.push_str(&value_to_string(arg));
                }
                Ok(Value::String(result))
            }
            "range" => match (&args[0], &args[1]) {
                (Value::Int(start), Value::Int(end)) => {
                    let list: Vec<Value> = (*start..*end).map(Value::Int).collect();
                    Ok(Value::List(list))
                }
                _ => Err(RuntimeError {
                    message: "range requires Int arguments".into(),
                }),
            },
            _ => Err(RuntimeError {
                message: format!("unknown builtin: {}", name),
            }),
        }
    }
}

fn match_pattern(pattern: &hir::Pattern, value: &Value) -> Option<Vec<(String, Value)>> {
    match pattern {
        hir::Pattern::Wildcard(_) => Some(vec![]),
        hir::Pattern::Binding { name, .. } => Some(vec![(name.clone(), value.clone())]),
        hir::Pattern::Literal(lit) => {
            let lit_val = match lit.as_ref() {
                hir::Expr::Int(n, _) => Value::Int(*n),
                hir::Expr::Float(n, _) => Value::Float(*n),
                hir::Expr::String(s, _) => Value::String(s.clone()),
                hir::Expr::Bool(b, _) => Value::Bool(*b),
                hir::Expr::Nil(_) => Value::Unit,
                _ => return None,
            };
            if values_equal(&lit_val, value) {
                Some(vec![])
            } else {
                None
            }
        }
        hir::Pattern::Constructor {
            variant_name,
            bindings,
            ..
        } => match value {
            Value::Variant {
                variant_name: vn,
                values,
                ..
            } if vn == variant_name => {
                if bindings.len() != values.len() {
                    return None;
                }
                let mut all_bindings = Vec::new();
                for (pat, val) in bindings.iter().zip(values.iter()) {
                    let sub_bindings = match_pattern(pat, val)?;
                    all_bindings.extend(sub_bindings);
                }
                Some(all_bindings)
            }
            _ => None,
        },
    }
}

fn values_equal(a: &Value, b: &Value) -> bool {
    match (a, b) {
        (Value::Int(a), Value::Int(b)) => a == b,
        (Value::Float(a), Value::Float(b)) => a == b,
        (Value::Bool(a), Value::Bool(b)) => a == b,
        (Value::String(a), Value::String(b)) => a == b,
        (Value::Unit, Value::Unit) => true,
        _ => false,
    }
}

fn value_to_string(val: &Value) -> String {
    match val {
        Value::Int(n) => n.to_string(),
        Value::Float(n) => n.to_string(),
        Value::Bool(b) => b.to_string(),
        Value::String(s) => s.clone(),
        Value::Unit => "nil".into(),
        other => format!("{}", other),
    }
}

fn numeric_binop(
    args: &[Value],
    int_op: impl Fn(i64, i64) -> i64,
    float_op: impl Fn(f64, f64) -> f64,
) -> Result<Value, RuntimeError> {
    match (&args[0], &args[1]) {
        (Value::Int(a), Value::Int(b)) => Ok(Value::Int(int_op(*a, *b))),
        (Value::Float(a), Value::Float(b)) => Ok(Value::Float(float_op(*a, *b))),
        _ => Err(RuntimeError {
            message: "arithmetic requires matching numeric types".into(),
        }),
    }
}

fn numeric_cmp(
    args: &[Value],
    int_op: impl Fn(i64, i64) -> bool,
    float_op: impl Fn(f64, f64) -> bool,
) -> Result<Value, RuntimeError> {
    match (&args[0], &args[1]) {
        (Value::Int(a), Value::Int(b)) => Ok(Value::Bool(int_op(*a, *b))),
        (Value::Float(a), Value::Float(b)) => Ok(Value::Bool(float_op(*a, *b))),
        _ => Err(RuntimeError {
            message: "comparison requires matching numeric types".into(),
        }),
    }
}

pub fn eval(module: &hir::Module) -> Result<Value, RuntimeError> {
    let mut interpreter = Interpreter::new();
    interpreter.eval_module(module)
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::lexer;
    use crate::parser;
    use crate::source::SourceMap;
    use crate::typechecker;

    fn eval_source(source: &str) -> Result<Value, RuntimeError> {
        let mut source_map = SourceMap::new();
        let file_id = source_map.add_file("test.vx".into(), source.into());
        let (tokens, lex_diags) = lexer::lex(source, file_id);
        assert!(lex_diags.is_empty(), "lex errors: {:?}", lex_diags);
        let (ast, parse_diags) = parser::parse(&tokens);
        assert!(parse_diags.is_empty(), "parse errors: {:?}", parse_diags);
        let (hir_module, check_diags) = typechecker::check(&ast);
        assert!(
            check_diags.is_empty(),
            "type check errors: {:?}",
            check_diags
        );
        eval(&hir_module)
    }

    #[test]
    fn int_literal() {
        let result = eval_source("42").unwrap();
        assert!(matches!(result, Value::Int(42)));
    }

    #[test]
    fn float_literal() {
        let result = eval_source("3.14").unwrap();
        assert!(matches!(result, Value::Float(f) if (f - 3.14).abs() < f64::EPSILON));
    }

    #[test]
    fn string_literal() {
        let result = eval_source(r#""hello""#).unwrap();
        assert!(matches!(result, Value::String(ref s) if s == "hello"));
    }

    #[test]
    fn bool_literal() {
        assert!(matches!(eval_source("true").unwrap(), Value::Bool(true)));
        assert!(matches!(eval_source("false").unwrap(), Value::Bool(false)));
    }

    #[test]
    fn nil_literal() {
        assert!(matches!(eval_source("nil").unwrap(), Value::Unit));
    }

    #[test]
    fn addition() {
        let result = eval_source("(+ 1 2)").unwrap();
        assert!(matches!(result, Value::Int(3)));
    }

    #[test]
    fn nested_arithmetic() {
        let result = eval_source("(* (+ 1 2) (- 5 3))").unwrap();
        assert!(matches!(result, Value::Int(6)));
    }

    #[test]
    fn comparison() {
        assert!(matches!(eval_source("(< 1 2)").unwrap(), Value::Bool(true)));
        assert!(matches!(
            eval_source("(> 1 2)").unwrap(),
            Value::Bool(false)
        ));
        assert!(matches!(
            eval_source("(<= 2 2)").unwrap(),
            Value::Bool(true)
        ));
        assert!(matches!(eval_source("(= 3 3)").unwrap(), Value::Bool(true)));
        assert!(matches!(
            eval_source("(!= 1 2)").unwrap(),
            Value::Bool(true)
        ));
    }

    #[test]
    fn logical_ops() {
        assert!(matches!(
            eval_source("(&& true false)").unwrap(),
            Value::Bool(false)
        ));
        assert!(matches!(
            eval_source("(|| true false)").unwrap(),
            Value::Bool(true)
        ));
        assert!(matches!(
            eval_source("(not true)").unwrap(),
            Value::Bool(false)
        ));
    }

    #[test]
    fn modulo() {
        let result = eval_source("(mod 10 3)").unwrap();
        assert!(matches!(result, Value::Int(1)));
    }

    #[test]
    fn if_true_branch() {
        let result = eval_source("(if true 1 2)").unwrap();
        assert!(matches!(result, Value::Int(1)));
    }

    #[test]
    fn if_false_branch() {
        let result = eval_source("(if false 1 2)").unwrap();
        assert!(matches!(result, Value::Int(2)));
    }

    #[test]
    fn let_binding() {
        let result = eval_source("(let [x 42] x)").unwrap();
        assert!(matches!(result, Value::Int(42)));
    }

    #[test]
    fn let_multiple_bindings() {
        let result = eval_source("(let [x 1 y 2] (+ x y))").unwrap();
        assert!(matches!(result, Value::Int(3)));
    }

    #[test]
    fn let_multiple_body_exprs() {
        let result = eval_source("(let [x 10] (+ x 1) (+ x 2))").unwrap();
        assert!(matches!(result, Value::Int(12)));
    }

    #[test]
    fn defn_and_call() {
        let source = "(defn double [x : Int] : Int (* x 2))\n(double 21)";
        let result = eval_source(source).unwrap();
        assert!(matches!(result, Value::Int(42)));
    }

    #[test]
    fn recursion() {
        let source = r#"
            (defn fib [n : Int] : Int
              (if (<= n 1)
                n
                (+ (fib (- n 1)) (fib (- n 2)))))
            (fib 10)
        "#;
        let result = eval_source(source).unwrap();
        assert!(matches!(result, Value::Int(55)));
    }

    #[test]
    fn cond_expression() {
        let source = r#"
            (defn classify [n : Int] : String
              (cond
                (< n 0) "negative"
                (= n 0) "zero"
                :else    "positive"))
            (classify 5)
        "#;
        let result = eval_source(source).unwrap();
        assert!(matches!(result, Value::String(ref s) if s == "positive"));
    }

    #[test]
    fn str_builtin() {
        let result = eval_source(r#"(str "x = " 42)"#).unwrap();
        assert!(matches!(result, Value::String(ref s) if s == "x = 42"));
    }

    #[test]
    fn def_constant() {
        let source = "(def pi 3.14)\npi";
        let result = eval_source(source).unwrap();
        assert!(matches!(result, Value::Float(f) if (f - 3.14).abs() < f64::EPSILON));
    }

    #[test]
    fn lambda() {
        let source = "(let [sq (fn [x : Int] : Int (* x x))] (sq 5))";
        let result = eval_source(source).unwrap();
        assert!(matches!(result, Value::Int(25)));
    }

    #[test]
    fn division() {
        let result = eval_source("(/ 10 3)").unwrap();
        assert!(matches!(result, Value::Int(3)));
    }

    #[test]
    fn division_by_zero() {
        let result = eval_source("(/ 42 0)");
        assert!(result.is_err());
    }

    #[test]
    fn modulo_by_zero() {
        let result = eval_source("(mod 42 0)");
        assert!(result.is_err());
    }

    #[test]
    fn range_builtin() {
        let result = eval_source("(range 0 5)").unwrap();
        if let Value::List(items) = result {
            assert_eq!(items.len(), 5);
            assert!(matches!(items[0], Value::Int(0)));
            assert!(matches!(items[4], Value::Int(4)));
        } else {
            panic!("expected List value");
        }
    }

    #[test]
    fn record_construction_and_field_access() {
        let source = r#"
            (deftype Point (x Float) (y Float))
            (let [p (Point 1.0 2.0)]
              (. p x))
        "#;
        let result = eval_source(source).unwrap();
        assert!(matches!(result, Value::Float(f) if (f - 1.0).abs() < f64::EPSILON));
    }

    #[test]
    fn variant_and_match() {
        let source = r#"
            (defunion Shape
              (Circle Float)
              (Rect Float Float))
            (let [s (Circle 5.0)]
              (match s
                (Circle r) r
                (Rect w h) (+ w h)))
        "#;
        let result = eval_source(source).unwrap();
        assert!(matches!(result, Value::Float(f) if (f - 5.0).abs() < f64::EPSILON));
    }

    #[test]
    fn match_wildcard() {
        let source = r#"
            (defunion Color (Red) (Green) (Blue))
            (let [c (Green)]
              (match c
                (Red) "red"
                _ "other"))
        "#;
        let result = eval_source(source).unwrap();
        assert!(matches!(result, Value::String(ref s) if s == "other"));
    }

    #[test]
    fn fizzbuzz() {
        let source = r#"
            (defn fizzbuzz [n : Int] : String
              (cond
                (= (mod n 15) 0) "FizzBuzz"
                (= (mod n 3) 0)  "Fizz"
                (= (mod n 5) 0)  "Buzz"
                :else             (str n)))
            (fizzbuzz 15)
        "#;
        let result = eval_source(source).unwrap();
        assert!(matches!(result, Value::String(ref s) if s == "FizzBuzz"));
    }

    #[test]
    fn value_display() {
        assert_eq!(format!("{}", Value::Int(42)), "42");
        assert_eq!(format!("{}", Value::Float(3.14)), "3.14");
        assert_eq!(format!("{}", Value::Bool(true)), "true");
        assert_eq!(format!("{}", Value::String("hi".into())), "\"hi\"");
        assert_eq!(format!("{}", Value::Unit), "nil");
        assert_eq!(
            format!(
                "{}",
                Value::List(vec![Value::Int(1), Value::Int(2), Value::Int(3)])
            ),
            "[1 2 3]"
        );
        assert_eq!(
            format!(
                "{}",
                Value::Variant {
                    union_name: "Option".into(),
                    variant_name: "Some".into(),
                    values: vec![Value::Int(42)],
                }
            ),
            "(Some 42)"
        );
        assert_eq!(
            format!(
                "{}",
                Value::Variant {
                    union_name: "Option".into(),
                    variant_name: "None".into(),
                    values: vec![],
                }
            ),
            "(None)"
        );
    }

    #[test]
    fn multiple_defns() {
        let source = r#"
            (defn add [a : Int b : Int] : Int (+ a b))
            (defn mul [a : Int b : Int] : Int (* a b))
            (add (mul 3 4) 5)
        "#;
        let result = eval_source(source).unwrap();
        assert!(matches!(result, Value::Int(17)));
    }

    #[test]
    fn println_returns_unit() {
        let source = r#"(println "hello")"#;
        let result = eval_source(source).unwrap();
        assert!(matches!(result, Value::Unit));
    }
}
