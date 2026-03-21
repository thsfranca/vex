use crate::ast::{Binding, CondClause, Expr, Param, TopForm, TypeExpr};
use crate::diagnostics::{Diagnostic, Label};
use crate::lexer::{Token, TokenKind};
use crate::source::{FileId, Span};

fn token_kind_name(kind: &TokenKind) -> &'static str {
    match kind {
        TokenKind::LeftParen => "'('",
        TokenKind::RightParen => "')'",
        TokenKind::LeftBracket => "'['",
        TokenKind::RightBracket => "']'",
        TokenKind::LeftBrace => "'{'",
        TokenKind::RightBrace => "'}'",
        TokenKind::Dot => "'.'",
        TokenKind::Colon => "':'",
        TokenKind::Symbol(_) => "symbol",
        TokenKind::Keyword(_) => "keyword",
        TokenKind::Integer(_) => "integer",
        TokenKind::Float(_) => "float",
        TokenKind::String(_) => "string",
        TokenKind::Boolean(_) => "boolean",
        TokenKind::Nil => "nil",
    }
}

struct Parser<'a> {
    tokens: &'a [Token],
    pos: usize,
    diagnostics: Vec<Diagnostic>,
}

impl<'a> Parser<'a> {
    fn new(tokens: &'a [Token]) -> Self {
        Self {
            tokens,
            pos: 0,
            diagnostics: Vec::new(),
        }
    }

    fn eof_span(&self) -> Span {
        self.tokens
            .last()
            .map(|t| Span::new(t.span.file, t.span.end, t.span.end))
            .unwrap_or(Span::new(FileId::new(0), 0, 0))
    }

    fn at_end(&self) -> bool {
        self.pos >= self.tokens.len()
    }

    fn check(&self, f: impl FnOnce(&TokenKind) -> bool) -> bool {
        self.tokens.get(self.pos).is_some_and(|t| f(&t.kind))
    }

    fn expect_right_paren(&mut self, open_span: Span) -> Option<Span> {
        if let Some(token) = self.tokens.get(self.pos) {
            if matches!(token.kind, TokenKind::RightParen) {
                let span = token.span;
                self.pos += 1;
                return Some(span);
            }
            let desc = token_kind_name(&token.kind);
            let span = token.span;
            self.diagnostics.push(
                Diagnostic::error(format!("expected ')', found {}", desc), span)
                    .with_label(Label::new(open_span, "to match this '('")),
            );
        } else {
            self.diagnostics.push(
                Diagnostic::error("expected ')' but reached end of input", self.eof_span())
                    .with_label(Label::new(open_span, "to match this '('")),
            );
        }
        None
    }

    fn expect_right_bracket(&mut self, open_span: Span) -> Option<Span> {
        if let Some(token) = self.tokens.get(self.pos) {
            if matches!(token.kind, TokenKind::RightBracket) {
                let span = token.span;
                self.pos += 1;
                return Some(span);
            }
            let desc = token_kind_name(&token.kind);
            let span = token.span;
            self.diagnostics.push(
                Diagnostic::error(format!("expected ']', found {}", desc), span)
                    .with_label(Label::new(open_span, "to match this '['")),
            );
        } else {
            self.diagnostics.push(
                Diagnostic::error("expected ']' but reached end of input", self.eof_span())
                    .with_label(Label::new(open_span, "to match this '['")),
            );
        }
        None
    }

    fn expect_symbol(&mut self) -> Option<(String, Span)> {
        if self.at_end() {
            self.diagnostics
                .push(Diagnostic::error("expected symbol but reached end of input", self.eof_span()));
            return None;
        }
        let span = self.tokens[self.pos].span;
        if let TokenKind::Symbol(ref name) = self.tokens[self.pos].kind {
            let name = name.clone();
            self.pos += 1;
            Some((name, span))
        } else {
            let desc = token_kind_name(&self.tokens[self.pos].kind);
            self.diagnostics
                .push(Diagnostic::error(format!("expected symbol, found {}", desc), span));
            None
        }
    }

    fn parse_program(&mut self) -> Vec<TopForm> {
        let mut forms = Vec::new();
        while !self.at_end() {
            if let Some(form) = self.parse_top_form() {
                forms.push(form);
            } else {
                break;
            }
        }
        forms
    }

    fn parse_top_form(&mut self) -> Option<TopForm> {
        if self.check(|k| matches!(k, TokenKind::LeftParen))
            && let Some(next) = self.tokens.get(self.pos + 1)
            && let TokenKind::Symbol(ref s) = next.kind
        {
            if s == "defn" {
                let open_span = self.tokens[self.pos].span;
                self.pos += 2;
                return self.parse_defn(open_span);
            }
            if s == "def" {
                let open_span = self.tokens[self.pos].span;
                self.pos += 2;
                return self.parse_def(open_span);
            }
        }
        let expr = self.parse_expr()?;
        Some(TopForm::Expr(expr))
    }

    fn parse_defn(&mut self, open_span: Span) -> Option<TopForm> {
        let (name, _) = self.expect_symbol()?;
        let params = self.parse_param_list()?;

        let return_type = if self.check(|k| matches!(k, TokenKind::Colon)) {
            self.pos += 1;
            Some(self.parse_type()?)
        } else {
            None
        };

        let body = self.parse_body()?;
        if body.is_empty() {
            let span = self.tokens.get(self.pos).map_or(self.eof_span(), |t| t.span);
            self.diagnostics
                .push(Diagnostic::error("expected at least one expression in function body", span));
            return None;
        }

        let close_span = self.expect_right_paren(open_span)?;

        Some(TopForm::Defn {
            name,
            params,
            return_type,
            body,
            span: Span::new(open_span.file, open_span.start, close_span.end),
        })
    }

    fn parse_def(&mut self, open_span: Span) -> Option<TopForm> {
        let (name, _) = self.expect_symbol()?;

        let type_ann = if self.check(|k| matches!(k, TokenKind::Colon)) {
            self.pos += 1;
            Some(self.parse_type()?)
        } else {
            None
        };

        let value = self.parse_expr()?;
        let close_span = self.expect_right_paren(open_span)?;

        Some(TopForm::Def {
            name,
            type_ann,
            value,
            span: Span::new(open_span.file, open_span.start, close_span.end),
        })
    }

    fn parse_expr(&mut self) -> Option<Expr> {
        if self.at_end() {
            self.diagnostics
                .push(Diagnostic::error("unexpected end of input", self.eof_span()));
            return None;
        }

        let span = self.tokens[self.pos].span;
        let kind = self.tokens[self.pos].kind.clone();

        match kind {
            TokenKind::Integer(n) => {
                self.pos += 1;
                Some(Expr::Int(n, span))
            }
            TokenKind::Float(f) => {
                self.pos += 1;
                Some(Expr::Float(f, span))
            }
            TokenKind::String(s) => {
                self.pos += 1;
                Some(Expr::String(s, span))
            }
            TokenKind::Boolean(b) => {
                self.pos += 1;
                Some(Expr::Bool(b, span))
            }
            TokenKind::Nil => {
                self.pos += 1;
                Some(Expr::Nil(span))
            }
            TokenKind::Symbol(name) => {
                self.pos += 1;
                Some(Expr::Symbol(name, span))
            }
            TokenKind::Keyword(name) => {
                self.pos += 1;
                Some(Expr::Keyword(name, span))
            }
            TokenKind::LeftParen => {
                self.pos += 1;
                self.parse_list_expr(span)
            }
            ref k => {
                self.diagnostics.push(Diagnostic::error(
                    format!("unexpected {} in expression", token_kind_name(k)),
                    span,
                ));
                None
            }
        }
    }

    fn parse_list_expr(&mut self, open_span: Span) -> Option<Expr> {
        if self.check(|k| matches!(k, TokenKind::RightParen)) {
            let close_span = self.tokens[self.pos].span;
            self.pos += 1;
            self.diagnostics.push(Diagnostic::error(
                "unexpected empty expression '()'",
                Span::new(open_span.file, open_span.start, close_span.end),
            ));
            return None;
        }

        if let Some(token) = self.tokens.get(self.pos)
            && let TokenKind::Symbol(ref s) = token.kind
        {
            match s.as_str() {
                "if" => {
                    self.pos += 1;
                    return self.parse_if(open_span);
                }
                "let" => {
                    self.pos += 1;
                    return self.parse_let(open_span);
                }
                "cond" => {
                    self.pos += 1;
                    return self.parse_cond(open_span);
                }
                "fn" => {
                    self.pos += 1;
                    return self.parse_lambda(open_span);
                }
                _ => {}
            }
        }

        self.parse_call(open_span)
    }

    fn parse_if(&mut self, open_span: Span) -> Option<Expr> {
        let test = self.parse_expr()?;
        let then_branch = self.parse_expr()?;
        let else_branch = self.parse_expr()?;
        let close_span = self.expect_right_paren(open_span)?;

        Some(Expr::If {
            test: Box::new(test),
            then_branch: Box::new(then_branch),
            else_branch: Box::new(else_branch),
            span: Span::new(open_span.file, open_span.start, close_span.end),
        })
    }

    fn parse_let(&mut self, open_span: Span) -> Option<Expr> {
        let bindings = self.parse_binding_list()?;

        let body = self.parse_body()?;
        if body.is_empty() {
            let span = self.tokens.get(self.pos).map_or(self.eof_span(), |t| t.span);
            self.diagnostics
                .push(Diagnostic::error("expected at least one expression in let body", span));
            return None;
        }

        let close_span = self.expect_right_paren(open_span)?;

        Some(Expr::Let {
            bindings,
            body,
            span: Span::new(open_span.file, open_span.start, close_span.end),
        })
    }

    fn parse_cond(&mut self, open_span: Span) -> Option<Expr> {
        let mut clauses = Vec::new();
        let mut else_body = None;

        while !self.at_end() && !self.check(|k| matches!(k, TokenKind::RightParen)) {
            if self.check(|k| matches!(k, TokenKind::Keyword(s) if s == "else")) {
                let else_kw_span = self.tokens[self.pos].span;
                self.pos += 1;
                let body = self.parse_expr()?;
                else_body = Some(Box::new(body));

                if !self.check(|k| matches!(k, TokenKind::RightParen)) {
                    let span = self.tokens.get(self.pos).map_or(self.eof_span(), |t| t.span);
                    self.diagnostics.push(
                        Diagnostic::error("unexpected expression after :else clause", span)
                            .with_label(Label::new(else_kw_span, ":else must be the last clause")),
                    );
                    return None;
                }
                break;
            }

            let test_start = self.tokens.get(self.pos).map_or(self.eof_span(), |t| t.span);
            let test = self.parse_expr()?;
            let value = self.parse_expr()?;
            let clause_span = Span::new(test_start.file, test_start.start, value.span().end);
            clauses.push(CondClause {
                test,
                value,
                span: clause_span,
            });
        }

        let close_span = self.expect_right_paren(open_span)?;

        Some(Expr::Cond {
            clauses,
            else_body,
            span: Span::new(open_span.file, open_span.start, close_span.end),
        })
    }

    fn parse_lambda(&mut self, open_span: Span) -> Option<Expr> {
        let params = self.parse_param_list()?;

        let return_type = if self.check(|k| matches!(k, TokenKind::Colon)) {
            self.pos += 1;
            Some(self.parse_type()?)
        } else {
            None
        };

        let body = self.parse_body()?;
        if body.is_empty() {
            let span = self.tokens.get(self.pos).map_or(self.eof_span(), |t| t.span);
            self.diagnostics
                .push(Diagnostic::error("expected at least one expression in lambda body", span));
            return None;
        }

        let close_span = self.expect_right_paren(open_span)?;

        Some(Expr::Lambda {
            params,
            return_type,
            body,
            span: Span::new(open_span.file, open_span.start, close_span.end),
        })
    }

    fn parse_call(&mut self, open_span: Span) -> Option<Expr> {
        let func = self.parse_expr()?;
        let mut args = Vec::new();

        while !self.at_end() && !self.check(|k| matches!(k, TokenKind::RightParen)) {
            args.push(self.parse_expr()?);
        }

        let close_span = self.expect_right_paren(open_span)?;

        Some(Expr::Call {
            func: Box::new(func),
            args,
            span: Span::new(open_span.file, open_span.start, close_span.end),
        })
    }

    fn parse_param_list(&mut self) -> Option<Vec<Param>> {
        if self.at_end() {
            self.diagnostics.push(Diagnostic::error(
                "expected '[' for parameter list but reached end of input",
                self.eof_span(),
            ));
            return None;
        }

        if !self.check(|k| matches!(k, TokenKind::LeftBracket)) {
            let span = self.tokens[self.pos].span;
            let desc = token_kind_name(&self.tokens[self.pos].kind);
            self.diagnostics.push(Diagnostic::error(
                format!("expected '[' for parameter list, found {}", desc),
                span,
            ));
            return None;
        }

        let open_span = self.tokens[self.pos].span;
        self.pos += 1;

        let mut params = Vec::new();

        while !self.at_end() && !self.check(|k| matches!(k, TokenKind::RightBracket)) {
            let (name, name_span) = self.expect_symbol()?;

            let type_ann = if self.check(|k| matches!(k, TokenKind::Colon)) {
                self.pos += 1;
                Some(self.parse_type()?)
            } else {
                None
            };

            let param_end = type_ann
                .as_ref()
                .map(|t| t.span().end)
                .unwrap_or(name_span.end);

            params.push(Param {
                name,
                type_ann,
                span: Span::new(name_span.file, name_span.start, param_end),
            });
        }

        self.expect_right_bracket(open_span)?;
        Some(params)
    }

    fn parse_binding_list(&mut self) -> Option<Vec<Binding>> {
        if self.at_end() {
            self.diagnostics.push(Diagnostic::error(
                "expected '[' for binding list but reached end of input",
                self.eof_span(),
            ));
            return None;
        }

        if !self.check(|k| matches!(k, TokenKind::LeftBracket)) {
            let span = self.tokens[self.pos].span;
            let desc = token_kind_name(&self.tokens[self.pos].kind);
            self.diagnostics.push(Diagnostic::error(
                format!("expected '[' for binding list, found {}", desc),
                span,
            ));
            return None;
        }

        let open_span = self.tokens[self.pos].span;
        self.pos += 1;

        let mut bindings = Vec::new();

        while !self.at_end() && !self.check(|k| matches!(k, TokenKind::RightBracket)) {
            let (name, name_span) = self.expect_symbol()?;
            let value = self.parse_expr()?;
            let binding_span = Span::new(name_span.file, name_span.start, value.span().end);

            bindings.push(Binding {
                name,
                value,
                span: binding_span,
            });
        }

        self.expect_right_bracket(open_span)?;
        Some(bindings)
    }

    fn parse_type(&mut self) -> Option<TypeExpr> {
        if self.at_end() {
            self.diagnostics
                .push(Diagnostic::error("expected type but reached end of input", self.eof_span()));
            return None;
        }

        let span = self.tokens[self.pos].span;
        let kind = self.tokens[self.pos].kind.clone();

        match kind {
            TokenKind::Symbol(name) => {
                self.pos += 1;
                Some(TypeExpr::Named { name, span })
            }
            TokenKind::LeftParen => {
                self.pos += 1;
                self.parse_type_list(span)
            }
            ref k => {
                self.diagnostics.push(Diagnostic::error(
                    format!("expected type, found {}", token_kind_name(k)),
                    span,
                ));
                None
            }
        }
    }

    fn parse_type_list(&mut self, open_span: Span) -> Option<TypeExpr> {
        let (name, _) = self.expect_symbol()?;

        if name == "Fn" {
            if self.at_end() || !self.check(|k| matches!(k, TokenKind::LeftBracket)) {
                let span = self.tokens.get(self.pos).map_or(self.eof_span(), |t| t.span);
                self.diagnostics.push(Diagnostic::error(
                    "expected '[' for function parameter types",
                    span,
                ));
                return None;
            }

            let bracket_span = self.tokens[self.pos].span;
            self.pos += 1;

            let mut param_types = Vec::new();
            while !self.at_end() && !self.check(|k| matches!(k, TokenKind::RightBracket)) {
                param_types.push(self.parse_type()?);
            }
            self.expect_right_bracket(bracket_span)?;

            let ret = self.parse_type()?;
            let close_span = self.expect_right_paren(open_span)?;

            Some(TypeExpr::Function {
                params: param_types,
                ret: Box::new(ret),
                span: Span::new(open_span.file, open_span.start, close_span.end),
            })
        } else {
            let mut args = Vec::new();
            while !self.at_end() && !self.check(|k| matches!(k, TokenKind::RightParen)) {
                args.push(self.parse_type()?);
            }
            let close_span = self.expect_right_paren(open_span)?;

            Some(TypeExpr::Applied {
                name,
                args,
                span: Span::new(open_span.file, open_span.start, close_span.end),
            })
        }
    }

    fn parse_body(&mut self) -> Option<Vec<Expr>> {
        let mut exprs = Vec::new();
        while !self.at_end() && !self.check(|k| matches!(k, TokenKind::RightParen)) {
            exprs.push(self.parse_expr()?);
        }
        Some(exprs)
    }
}

pub fn parse(tokens: &[Token]) -> (Vec<TopForm>, Vec<Diagnostic>) {
    let mut parser = Parser::new(tokens);
    let forms = parser.parse_program();
    (forms, parser.diagnostics)
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::lexer::lex;
    use crate::source::FileId;

    fn parse_source(source: &str) -> (Vec<TopForm>, Vec<Diagnostic>) {
        let (tokens, lex_diags) = lex(source, FileId::new(0));
        assert!(lex_diags.is_empty(), "unexpected lexer errors: {:?}", lex_diags);
        parse(&tokens)
    }

    #[test]
    fn empty_input() {
        let (forms, diags) = parse_source("");
        assert!(forms.is_empty());
        assert!(diags.is_empty());
    }

    #[test]
    fn literal_int() {
        let (forms, diags) = parse_source("42");
        assert!(diags.is_empty());
        assert_eq!(forms.len(), 1);
        assert!(matches!(&forms[0], TopForm::Expr(Expr::Int(42, _))));
    }

    #[test]
    fn literal_float() {
        let (forms, diags) = parse_source("3.14");
        assert!(diags.is_empty());
        assert!(matches!(&forms[0], TopForm::Expr(Expr::Float(f, _)) if (*f - 3.14).abs() < f64::EPSILON));
    }

    #[test]
    fn literal_string() {
        let (forms, diags) = parse_source(r#""hello""#);
        assert!(diags.is_empty());
        assert!(matches!(&forms[0], TopForm::Expr(Expr::String(s, _)) if s == "hello"));
    }

    #[test]
    fn literal_bool() {
        let (forms, diags) = parse_source("true");
        assert!(diags.is_empty());
        assert!(matches!(&forms[0], TopForm::Expr(Expr::Bool(true, _))));
    }

    #[test]
    fn literal_nil() {
        let (forms, diags) = parse_source("nil");
        assert!(diags.is_empty());
        assert!(matches!(&forms[0], TopForm::Expr(Expr::Nil(_))));
    }

    #[test]
    fn symbol_expr() {
        let (forms, diags) = parse_source("foo");
        assert!(diags.is_empty());
        assert!(matches!(&forms[0], TopForm::Expr(Expr::Symbol(s, _)) if s == "foo"));
    }

    #[test]
    fn keyword_expr() {
        let (forms, diags) = parse_source(":name");
        assert!(diags.is_empty());
        assert!(matches!(&forms[0], TopForm::Expr(Expr::Keyword(s, _)) if s == "name"));
    }

    #[test]
    fn simple_call() {
        let (forms, diags) = parse_source("(+ 1 2)");
        assert!(diags.is_empty());
        assert_eq!(forms.len(), 1);
        if let TopForm::Expr(Expr::Call { func, args, .. }) = &forms[0] {
            assert!(matches!(func.as_ref(), Expr::Symbol(s, _) if s == "+"));
            assert_eq!(args.len(), 2);
            assert!(matches!(&args[0], Expr::Int(1, _)));
            assert!(matches!(&args[1], Expr::Int(2, _)));
        } else {
            panic!("expected call expression");
        }
    }

    #[test]
    fn nested_calls() {
        let (forms, diags) = parse_source("(+ (- 3 1) (* 2 4))");
        assert!(diags.is_empty());
        if let TopForm::Expr(Expr::Call { func, args, .. }) = &forms[0] {
            assert!(matches!(func.as_ref(), Expr::Symbol(s, _) if s == "+"));
            assert_eq!(args.len(), 2);
            assert!(matches!(&args[0], Expr::Call { .. }));
            assert!(matches!(&args[1], Expr::Call { .. }));
        } else {
            panic!("expected call expression");
        }
    }

    #[test]
    fn hello_world() {
        let source = r#"(defn main [] (println "Hello, World!"))"#;
        let (forms, diags) = parse_source(source);
        assert!(diags.is_empty());
        assert_eq!(forms.len(), 1);
        if let TopForm::Defn {
            name,
            params,
            return_type,
            body,
            ..
        } = &forms[0]
        {
            assert_eq!(name, "main");
            assert!(params.is_empty());
            assert!(return_type.is_none());
            assert_eq!(body.len(), 1);
            assert!(matches!(&body[0], Expr::Call { .. }));
        } else {
            panic!("expected defn");
        }
    }

    #[test]
    fn defn_with_types() {
        let source = "(defn add [a : Int, b : Int] : Int (+ a b))";
        let (forms, diags) = parse_source(source);
        assert!(diags.is_empty());
        if let TopForm::Defn {
            name,
            params,
            return_type,
            ..
        } = &forms[0]
        {
            assert_eq!(name, "add");
            assert_eq!(params.len(), 2);
            assert_eq!(params[0].name, "a");
            assert!(matches!(&params[0].type_ann, Some(TypeExpr::Named { name, .. }) if name == "Int"));
            assert_eq!(params[1].name, "b");
            assert!(matches!(&params[1].type_ann, Some(TypeExpr::Named { name, .. }) if name == "Int"));
            assert!(matches!(return_type, Some(TypeExpr::Named { name, .. }) if name == "Int"));
        } else {
            panic!("expected defn");
        }
    }

    #[test]
    fn fibonacci() {
        let source = "(defn fib [n : Int] : Int
          (if (<= n 1)
            n
            (+ (fib (- n 1)) (fib (- n 2)))))";
        let (forms, diags) = parse_source(source);
        assert!(diags.is_empty());
        if let TopForm::Defn {
            name,
            params,
            return_type,
            body,
            ..
        } = &forms[0]
        {
            assert_eq!(name, "fib");
            assert_eq!(params.len(), 1);
            assert_eq!(params[0].name, "n");
            assert!(return_type.is_some());
            assert_eq!(body.len(), 1);
            assert!(matches!(&body[0], Expr::If { .. }));
        } else {
            panic!("expected defn");
        }
    }

    #[test]
    fn def_constant() {
        let source = "(def pi : Float 3.14)";
        let (forms, diags) = parse_source(source);
        assert!(diags.is_empty());
        if let TopForm::Def {
            name,
            type_ann,
            value,
            ..
        } = &forms[0]
        {
            assert_eq!(name, "pi");
            assert!(matches!(type_ann, Some(TypeExpr::Named { name, .. }) if name == "Float"));
            assert!(matches!(value, Expr::Float(f, _) if (*f - 3.14).abs() < f64::EPSILON));
        } else {
            panic!("expected def");
        }
    }

    #[test]
    fn def_without_type() {
        let source = "(def x 42)";
        let (forms, diags) = parse_source(source);
        assert!(diags.is_empty());
        if let TopForm::Def {
            name, type_ann, value, ..
        } = &forms[0]
        {
            assert_eq!(name, "x");
            assert!(type_ann.is_none());
            assert!(matches!(value, Expr::Int(42, _)));
        } else {
            panic!("expected def");
        }
    }

    #[test]
    fn if_expression() {
        let (forms, diags) = parse_source("(if true 1 0)");
        assert!(diags.is_empty());
        if let TopForm::Expr(Expr::If {
            test,
            then_branch,
            else_branch,
            ..
        }) = &forms[0]
        {
            assert!(matches!(test.as_ref(), Expr::Bool(true, _)));
            assert!(matches!(then_branch.as_ref(), Expr::Int(1, _)));
            assert!(matches!(else_branch.as_ref(), Expr::Int(0, _)));
        } else {
            panic!("expected if expression");
        }
    }

    #[test]
    fn let_expression() {
        let (forms, diags) = parse_source("(let [x 42] x)");
        assert!(diags.is_empty());
        if let TopForm::Expr(Expr::Let {
            bindings, body, ..
        }) = &forms[0]
        {
            assert_eq!(bindings.len(), 1);
            assert_eq!(bindings[0].name, "x");
            assert!(matches!(&bindings[0].value, Expr::Int(42, _)));
            assert_eq!(body.len(), 1);
            assert!(matches!(&body[0], Expr::Symbol(s, _) if s == "x"));
        } else {
            panic!("expected let expression");
        }
    }

    #[test]
    fn let_empty_bindings() {
        let (forms, diags) = parse_source("(let [] 42)");
        assert!(diags.is_empty());
        if let TopForm::Expr(Expr::Let { bindings, body, .. }) = &forms[0] {
            assert!(bindings.is_empty());
            assert_eq!(body.len(), 1);
        } else {
            panic!("expected let expression");
        }
    }

    #[test]
    fn let_multiple_bindings_and_body() {
        let source = "(let [x 1 y 2] (+ x y) x)";
        let (forms, diags) = parse_source(source);
        assert!(diags.is_empty());
        if let TopForm::Expr(Expr::Let {
            bindings, body, ..
        }) = &forms[0]
        {
            assert_eq!(bindings.len(), 2);
            assert_eq!(bindings[0].name, "x");
            assert_eq!(bindings[1].name, "y");
            assert_eq!(body.len(), 2);
        } else {
            panic!("expected let expression");
        }
    }

    #[test]
    fn cond_expression() {
        let source = r#"(cond
          (= x 0) "zero"
          (> x 0) "positive"
          :else "negative")"#;
        let (forms, diags) = parse_source(source);
        assert!(diags.is_empty());
        if let TopForm::Expr(Expr::Cond {
            clauses,
            else_body,
            ..
        }) = &forms[0]
        {
            assert_eq!(clauses.len(), 2);
            assert!(else_body.is_some());
            assert!(matches!(else_body.as_deref(), Some(Expr::String(s, _)) if s == "negative"));
        } else {
            panic!("expected cond expression");
        }
    }

    #[test]
    fn cond_without_else() {
        let (forms, diags) = parse_source("(cond (= x 0) 0)");
        assert!(diags.is_empty());
        if let TopForm::Expr(Expr::Cond {
            clauses,
            else_body,
            ..
        }) = &forms[0]
        {
            assert_eq!(clauses.len(), 1);
            assert!(else_body.is_none());
        } else {
            panic!("expected cond expression");
        }
    }

    #[test]
    fn lambda_expression() {
        let (forms, diags) = parse_source("(fn [x : Int] : Int (+ x 1))");
        assert!(diags.is_empty());
        if let TopForm::Expr(Expr::Lambda {
            params,
            return_type,
            body,
            ..
        }) = &forms[0]
        {
            assert_eq!(params.len(), 1);
            assert_eq!(params[0].name, "x");
            assert!(return_type.is_some());
            assert_eq!(body.len(), 1);
        } else {
            panic!("expected lambda expression");
        }
    }

    #[test]
    fn lambda_no_type_annotations() {
        let (forms, diags) = parse_source("(fn [x] x)");
        assert!(diags.is_empty());
        if let TopForm::Expr(Expr::Lambda {
            params,
            return_type,
            ..
        }) = &forms[0]
        {
            assert_eq!(params.len(), 1);
            assert!(params[0].type_ann.is_none());
            assert!(return_type.is_none());
        } else {
            panic!("expected lambda expression");
        }
    }

    #[test]
    fn multiple_top_forms() {
        let source = "(def x 42) (defn main [] (println x))";
        let (forms, diags) = parse_source(source);
        assert!(diags.is_empty());
        assert_eq!(forms.len(), 2);
        assert!(matches!(&forms[0], TopForm::Def { .. }));
        assert!(matches!(&forms[1], TopForm::Defn { .. }));
    }

    #[test]
    fn type_function() {
        let (forms, diags) = parse_source("(def f : (Fn [Int Int] Bool) nil)");
        assert!(diags.is_empty());
        if let TopForm::Def {
            type_ann: Some(TypeExpr::Function { params, ret, .. }),
            ..
        } = &forms[0]
        {
            assert_eq!(params.len(), 2);
            assert!(matches!(ret.as_ref(), TypeExpr::Named { name, .. } if name == "Bool"));
        } else {
            panic!("expected def with function type");
        }
    }

    #[test]
    fn type_applied() {
        let (forms, diags) = parse_source("(def xs : (List Int) nil)");
        assert!(diags.is_empty());
        if let TopForm::Def {
            type_ann: Some(TypeExpr::Applied { name, args, .. }),
            ..
        } = &forms[0]
        {
            assert_eq!(name, "List");
            assert_eq!(args.len(), 1);
        } else {
            panic!("expected def with applied type");
        }
    }

    #[test]
    fn spans_cover_full_form() {
        let source = "(+ 1 2)";
        let (forms, _) = parse_source(source);
        let span = forms[0].span();
        assert_eq!(span.start, 0);
        assert_eq!(span.end, 7);
    }

    #[test]
    fn fizzbuzz_program() {
        let source = r#"(defn fizzbuzz [n : Int] : String
  (cond
    (= (mod n 15) 0) "FizzBuzz"
    (= (mod n 3) 0)  "Fizz"
    (= (mod n 5) 0)  "Buzz"
    :else             (str n)))

(defn fizzbuzz-loop [i : Int, max : Int]
  (if (<= i max)
    (let []
      (println (fizzbuzz i))
      (fizzbuzz-loop (+ i 1) max))
    nil))

(defn main []
  (fizzbuzz-loop 1 100))
"#;
        let (forms, diags) = parse_source(source);
        assert!(diags.is_empty(), "unexpected errors: {:?}", diags);
        assert_eq!(forms.len(), 3);
        assert!(matches!(&forms[0], TopForm::Defn { name, .. } if name == "fizzbuzz"));
        assert!(matches!(&forms[1], TopForm::Defn { name, .. } if name == "fizzbuzz-loop"));
        assert!(matches!(&forms[2], TopForm::Defn { name, .. } if name == "main"));
    }

    #[test]
    fn error_unclosed_paren() {
        let (_, diags) = parse_source("(+ 1 2");
        assert_eq!(diags.len(), 1);
        assert!(diags[0].message.contains("expected ')'"));
    }

    #[test]
    fn error_empty_parens() {
        let (_, diags) = parse_source("()");
        assert_eq!(diags.len(), 1);
        assert!(diags[0].message.contains("empty expression"));
    }

    #[test]
    fn error_defn_missing_name() {
        let (_, diags) = parse_source("(defn)");
        assert!(!diags.is_empty());
    }

    #[test]
    fn error_defn_missing_params() {
        let (_, diags) = parse_source("(defn main (println))");
        assert!(!diags.is_empty());
        assert!(diags[0].message.contains("'['"));
    }

    #[test]
    fn error_let_empty_body() {
        let (_, diags) = parse_source("(let [x 1])");
        assert!(!diags.is_empty());
        assert!(diags[0].message.contains("at least one expression"));
    }

    #[test]
    fn error_unexpected_token() {
        let (_, diags) = parse_source("]");
        assert!(!diags.is_empty());
        assert!(diags[0].message.contains("unexpected"));
    }
}
