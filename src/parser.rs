use crate::ast::{Binding, CondClause, Expr, Field, Param, TopForm, TypeExpr};
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
            self.diagnostics.push(Diagnostic::error(
                "expected symbol but reached end of input",
                self.eof_span(),
            ));
            return None;
        }
        let span = self.tokens[self.pos].span;
        if let TokenKind::Symbol(ref name) = self.tokens[self.pos].kind {
            let name = name.clone();
            self.pos += 1;
            Some((name, span))
        } else {
            let desc = token_kind_name(&self.tokens[self.pos].kind);
            self.diagnostics.push(Diagnostic::error(
                format!("expected symbol, found {}", desc),
                span,
            ));
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
            && let Some(TokenKind::Symbol(name)) = self.tokens.get(self.pos + 1).map(|t| &t.kind)
        {
            match name.as_str() {
                "defn" => {
                    let open_span = self.tokens[self.pos].span;
                    self.pos += 1;
                    return self.parse_defn(open_span);
                }
                "def" => {
                    let open_span = self.tokens[self.pos].span;
                    self.pos += 1;
                    return self.parse_def(open_span);
                }
                "deftype" => {
                    let open_span = self.tokens[self.pos].span;
                    self.pos += 1;
                    return self.parse_deftype(open_span);
                }
                _ => {}
            }
        }

        let expr = self.parse_expr()?;
        Some(TopForm::Expr(expr))
    }

    fn parse_expr(&mut self) -> Option<Expr> {
        if self.at_end() {
            self.diagnostics.push(Diagnostic::error(
                "unexpected end of input",
                self.eof_span(),
            ));
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

        if let Some(TokenKind::Symbol(name)) = self.tokens.get(self.pos).map(|t| &t.kind) {
            match name.as_str() {
                "if" => return self.parse_if(open_span),
                "let" => return self.parse_let(open_span),
                "cond" => return self.parse_cond(open_span),
                "fn" => return self.parse_lambda(open_span),
                _ => {}
            }
        }

        self.parse_call(open_span)
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

    fn parse_if(&mut self, open_span: Span) -> Option<Expr> {
        self.pos += 1;

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
        self.pos += 1;

        let bindings = self.parse_binding_list()?;
        let body = self.parse_body()?;

        let close_span = self.expect_right_paren(open_span)?;

        Some(Expr::Let {
            bindings,
            body,
            span: Span::new(open_span.file, open_span.start, close_span.end),
        })
    }

    fn parse_binding_list(&mut self) -> Option<Vec<Binding>> {
        if self.at_end() {
            self.diagnostics.push(Diagnostic::error(
                "expected '[' for binding list but reached end of input",
                self.eof_span(),
            ));
            return None;
        }

        let open_span = self.tokens[self.pos].span;
        if !matches!(self.tokens[self.pos].kind, TokenKind::LeftBracket) {
            let desc = token_kind_name(&self.tokens[self.pos].kind);
            self.diagnostics.push(Diagnostic::error(
                format!("expected '[' for binding list, found {}", desc),
                open_span,
            ));
            return None;
        }
        self.pos += 1;

        let mut bindings = Vec::new();

        while !self.at_end() && !self.check(|k| matches!(k, TokenKind::RightBracket)) {
            let (name, name_span) = self.expect_symbol()?;
            let value = self.parse_expr()?;
            let span = Span::new(name_span.file, name_span.start, value.span().end);
            bindings.push(Binding { name, value, span });
        }

        self.expect_right_bracket(open_span)?;

        Some(bindings)
    }

    fn parse_body(&mut self) -> Option<Vec<Expr>> {
        let first = self.parse_expr()?;
        let mut body = vec![first];

        while !self.at_end() && !self.check(|k| matches!(k, TokenKind::RightParen)) {
            body.push(self.parse_expr()?);
        }

        Some(body)
    }

    fn parse_cond(&mut self, open_span: Span) -> Option<Expr> {
        self.pos += 1;

        let mut clauses = Vec::new();
        let mut else_body = None;

        while !self.at_end() && !self.check(|k| matches!(k, TokenKind::RightParen)) {
            if self.check(|k| matches!(k, TokenKind::Keyword(s) if s == "else")) {
                self.pos += 1;
                else_body = Some(Box::new(self.parse_expr()?));
                break;
            }

            let test = self.parse_expr()?;
            let value = self.parse_expr()?;
            let span = Span::new(test.span().file, test.span().start, value.span().end);
            clauses.push(CondClause { test, value, span });
        }

        let close_span = self.expect_right_paren(open_span)?;

        Some(Expr::Cond {
            clauses,
            else_body,
            span: Span::new(open_span.file, open_span.start, close_span.end),
        })
    }

    fn parse_lambda(&mut self, open_span: Span) -> Option<Expr> {
        self.pos += 1;

        let params = self.parse_param_list()?;

        let return_type = if self.check(|k| matches!(k, TokenKind::Colon)) {
            self.pos += 1;
            Some(self.parse_type()?)
        } else {
            None
        };

        let body = self.parse_body()?;
        let close_span = self.expect_right_paren(open_span)?;

        Some(Expr::Lambda {
            params,
            return_type,
            body,
            span: Span::new(open_span.file, open_span.start, close_span.end),
        })
    }

    fn parse_defn(&mut self, open_span: Span) -> Option<TopForm> {
        self.pos += 1;

        let (name, _) = self.expect_symbol()?;
        let params = self.parse_param_list()?;

        let return_type = if self.check(|k| matches!(k, TokenKind::Colon)) {
            self.pos += 1;
            Some(self.parse_type()?)
        } else {
            None
        };

        let body = self.parse_body()?;
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
        self.pos += 1;

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

    fn parse_deftype(&mut self, open_span: Span) -> Option<TopForm> {
        self.pos += 1;

        let (name, _) = self.expect_symbol()?;

        let mut fields = Vec::new();
        while !self.at_end() && !self.check(|k| matches!(k, TokenKind::RightParen)) {
            fields.push(self.parse_field()?);
        }

        let close_span = self.expect_right_paren(open_span)?;

        Some(TopForm::Deftype {
            name,
            fields,
            span: Span::new(open_span.file, open_span.start, close_span.end),
        })
    }

    fn parse_field(&mut self) -> Option<Field> {
        if self.at_end() {
            self.diagnostics.push(Diagnostic::error(
                "expected '(' for field declaration but reached end of input",
                self.eof_span(),
            ));
            return None;
        }

        let open_span = self.tokens[self.pos].span;
        if !matches!(self.tokens[self.pos].kind, TokenKind::LeftParen) {
            let desc = token_kind_name(&self.tokens[self.pos].kind);
            self.diagnostics.push(Diagnostic::error(
                format!("expected '(' for field declaration, found {}", desc),
                open_span,
            ));
            return None;
        }
        self.pos += 1;

        let (name, _) = self.expect_symbol()?;
        let type_expr = self.parse_type()?;
        let close_span = self.expect_right_paren(open_span)?;

        Some(Field {
            name,
            type_expr,
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

        let open_span = self.tokens[self.pos].span;
        if !matches!(self.tokens[self.pos].kind, TokenKind::LeftBracket) {
            let desc = token_kind_name(&self.tokens[self.pos].kind);
            self.diagnostics.push(Diagnostic::error(
                format!("expected '[' for parameter list, found {}", desc),
                open_span,
            ));
            return None;
        }
        self.pos += 1;

        let mut params = Vec::new();

        while !self.at_end() && !self.check(|k| matches!(k, TokenKind::RightBracket)) {
            let (name, name_span) = self.expect_symbol()?;

            let type_ann = if self.check(|k| matches!(k, TokenKind::Colon)) {
                self.pos += 1;
                let t = self.parse_type()?;
                Some(t)
            } else {
                None
            };

            let end = type_ann
                .as_ref()
                .map(|t| t.span().end)
                .unwrap_or(name_span.end);
            let span = Span::new(name_span.file, name_span.start, end);
            params.push(Param {
                name,
                type_ann,
                span,
            });
        }

        self.expect_right_bracket(open_span)?;

        Some(params)
    }

    fn parse_type(&mut self) -> Option<TypeExpr> {
        if self.at_end() {
            self.diagnostics.push(Diagnostic::error(
                "expected type but reached end of input",
                self.eof_span(),
            ));
            return None;
        }

        let span = self.tokens[self.pos].span;

        match &self.tokens[self.pos].kind {
            TokenKind::Symbol(name) => {
                let name = name.clone();
                self.pos += 1;
                Some(TypeExpr::Named { name, span })
            }
            TokenKind::LeftParen => {
                self.pos += 1;
                self.parse_compound_type(span)
            }
            _ => {
                let desc = token_kind_name(&self.tokens[self.pos].kind);
                self.diagnostics.push(Diagnostic::error(
                    format!("expected type, found {}", desc),
                    span,
                ));
                None
            }
        }
    }

    fn parse_compound_type(&mut self, open_span: Span) -> Option<TypeExpr> {
        if self.at_end() {
            self.diagnostics.push(Diagnostic::error(
                "expected type name but reached end of input",
                self.eof_span(),
            ));
            return None;
        }

        let (name, _) = self.expect_symbol()?;

        if name == "Fn" {
            return self.parse_fn_type(open_span);
        }

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

    fn parse_fn_type(&mut self, open_span: Span) -> Option<TypeExpr> {
        if self.at_end() {
            self.diagnostics.push(Diagnostic::error(
                "expected '[' for function type parameters but reached end of input",
                self.eof_span(),
            ));
            return None;
        }

        let bracket_span = self.tokens[self.pos].span;
        if !matches!(self.tokens[self.pos].kind, TokenKind::LeftBracket) {
            let desc = token_kind_name(&self.tokens[self.pos].kind);
            self.diagnostics.push(Diagnostic::error(
                format!("expected '[' for function type parameters, found {}", desc),
                bracket_span,
            ));
            return None;
        }
        self.pos += 1;

        let mut params = Vec::new();
        while !self.at_end() && !self.check(|k| matches!(k, TokenKind::RightBracket)) {
            params.push(self.parse_type()?);
        }

        self.expect_right_bracket(bracket_span)?;

        let ret = self.parse_type()?;
        let close_span = self.expect_right_paren(open_span)?;

        Some(TypeExpr::Function {
            params,
            ret: Box::new(ret),
            span: Span::new(open_span.file, open_span.start, close_span.end),
        })
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
        assert!(
            lex_diags.is_empty(),
            "unexpected lexer errors: {:?}",
            lex_diags
        );
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
        assert!(
            matches!(&forms[0], TopForm::Expr(Expr::Float(f, _)) if (*f - 3.14).abs() < f64::EPSILON)
        );
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
    fn multiple_atoms() {
        let (forms, diags) = parse_source("42 true nil");
        assert!(diags.is_empty());
        assert_eq!(forms.len(), 3);
        assert!(matches!(&forms[0], TopForm::Expr(Expr::Int(42, _))));
        assert!(matches!(&forms[1], TopForm::Expr(Expr::Bool(true, _))));
        assert!(matches!(&forms[2], TopForm::Expr(Expr::Nil(_))));
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
    fn call_no_args() {
        let (forms, diags) = parse_source("(foo)");
        assert!(diags.is_empty());
        if let TopForm::Expr(Expr::Call { func, args, .. }) = &forms[0] {
            assert!(matches!(func.as_ref(), Expr::Symbol(s, _) if s == "foo"));
            assert!(args.is_empty());
        } else {
            panic!("expected call expression");
        }
    }

    #[test]
    fn spans_cover_full_form() {
        let (forms, _) = parse_source("(+ 1 2)");
        let span = forms[0].span();
        assert_eq!(span.start, 0);
        assert_eq!(span.end, 7);
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
    fn error_unexpected_token() {
        let (_, diags) = parse_source("]");
        assert!(!diags.is_empty());
        assert!(diags[0].message.contains("unexpected"));
    }

    #[test]
    fn if_expression() {
        let (forms, diags) = parse_source("(if true 1 0)");
        assert!(diags.is_empty());
        assert_eq!(forms.len(), 1);
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
    fn if_with_nested_exprs() {
        let (forms, diags) = parse_source("(if (> x 0) (+ x 1) (- x 1))");
        assert!(diags.is_empty());
        if let TopForm::Expr(Expr::If {
            test,
            then_branch,
            else_branch,
            ..
        }) = &forms[0]
        {
            assert!(matches!(test.as_ref(), Expr::Call { .. }));
            assert!(matches!(then_branch.as_ref(), Expr::Call { .. }));
            assert!(matches!(else_branch.as_ref(), Expr::Call { .. }));
        } else {
            panic!("expected if expression");
        }
    }

    #[test]
    fn if_spans() {
        let (forms, _) = parse_source("(if true 1 0)");
        let span = forms[0].span();
        assert_eq!(span.start, 0);
        assert_eq!(span.end, 13);
    }

    #[test]
    fn let_simple() {
        let (forms, diags) = parse_source("(let [x 42] x)");
        assert!(diags.is_empty());
        if let TopForm::Expr(Expr::Let { bindings, body, .. }) = &forms[0] {
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
    fn let_multiple_bindings() {
        let (forms, diags) = parse_source("(let [x 1 y 2] (+ x y))");
        assert!(diags.is_empty());
        if let TopForm::Expr(Expr::Let { bindings, body, .. }) = &forms[0] {
            assert_eq!(bindings.len(), 2);
            assert_eq!(bindings[0].name, "x");
            assert_eq!(bindings[1].name, "y");
            assert_eq!(body.len(), 1);
        } else {
            panic!("expected let expression");
        }
    }

    #[test]
    fn let_multiple_body_exprs() {
        let (forms, diags) = parse_source("(let [x 1] (println x) x)");
        assert!(diags.is_empty());
        if let TopForm::Expr(Expr::Let { body, .. }) = &forms[0] {
            assert_eq!(body.len(), 2);
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
    fn let_spans() {
        let (forms, _) = parse_source("(let [x 42] x)");
        let span = forms[0].span();
        assert_eq!(span.start, 0);
        assert_eq!(span.end, 14);
    }

    #[test]
    fn cond_basic() {
        let (forms, diags) = parse_source("(cond (< x 0) \"neg\" (> x 0) \"pos\")");
        assert!(diags.is_empty());
        if let TopForm::Expr(Expr::Cond {
            clauses, else_body, ..
        }) = &forms[0]
        {
            assert_eq!(clauses.len(), 2);
            assert!(matches!(&clauses[0].test, Expr::Call { .. }));
            assert!(matches!(&clauses[0].value, Expr::String(s, _) if s == "neg"));
            assert!(matches!(&clauses[1].test, Expr::Call { .. }));
            assert!(matches!(&clauses[1].value, Expr::String(s, _) if s == "pos"));
            assert!(else_body.is_none());
        } else {
            panic!("expected cond expression");
        }
    }

    #[test]
    fn cond_with_else() {
        let (forms, diags) = parse_source("(cond (< x 0) \"neg\" :else \"zero\")");
        assert!(diags.is_empty());
        if let TopForm::Expr(Expr::Cond {
            clauses, else_body, ..
        }) = &forms[0]
        {
            assert_eq!(clauses.len(), 1);
            assert!(else_body.is_some());
            assert!(
                matches!(else_body.as_ref().unwrap().as_ref(), Expr::String(s, _) if s == "zero")
            );
        } else {
            panic!("expected cond expression");
        }
    }

    #[test]
    fn cond_empty() {
        let (forms, diags) = parse_source("(cond)");
        assert!(diags.is_empty());
        if let TopForm::Expr(Expr::Cond {
            clauses, else_body, ..
        }) = &forms[0]
        {
            assert!(clauses.is_empty());
            assert!(else_body.is_none());
        } else {
            panic!("expected cond expression");
        }
    }

    #[test]
    fn cond_spans() {
        let (forms, _) = parse_source("(cond :else 0)");
        let span = forms[0].span();
        assert_eq!(span.start, 0);
        assert_eq!(span.end, 14);
    }

    #[test]
    fn error_let_missing_bracket() {
        let (_, diags) = parse_source("(let x 42 x)");
        assert!(!diags.is_empty());
        assert!(diags[0].message.contains("'['"));
    }

    #[test]
    fn error_let_non_symbol_binding() {
        let (_, diags) = parse_source("(let [42 1] x)");
        assert!(!diags.is_empty());
        assert!(diags[0].message.contains("expected symbol"));
    }

    #[test]
    fn error_if_missing_else() {
        let (_, diags) = parse_source("(if true 1)");
        assert!(!diags.is_empty());
    }

    #[test]
    fn lambda_no_params() {
        let (forms, diags) = parse_source("(fn [] 42)");
        assert!(diags.is_empty());
        if let TopForm::Expr(Expr::Lambda {
            params,
            return_type,
            body,
            ..
        }) = &forms[0]
        {
            assert!(params.is_empty());
            assert!(return_type.is_none());
            assert_eq!(body.len(), 1);
            assert!(matches!(&body[0], Expr::Int(42, _)));
        } else {
            panic!("expected lambda expression");
        }
    }

    #[test]
    fn lambda_with_params() {
        let (forms, diags) = parse_source("(fn [x y] (+ x y))");
        assert!(diags.is_empty());
        if let TopForm::Expr(Expr::Lambda { params, body, .. }) = &forms[0] {
            assert_eq!(params.len(), 2);
            assert_eq!(params[0].name, "x");
            assert!(params[0].type_ann.is_none());
            assert_eq!(params[1].name, "y");
            assert!(params[1].type_ann.is_none());
            assert_eq!(body.len(), 1);
        } else {
            panic!("expected lambda expression");
        }
    }

    #[test]
    fn lambda_with_typed_params() {
        let (forms, diags) = parse_source("(fn [x : Int y : Int] (+ x y))");
        assert!(diags.is_empty());
        if let TopForm::Expr(Expr::Lambda { params, .. }) = &forms[0] {
            assert_eq!(params.len(), 2);
            assert_eq!(params[0].name, "x");
            assert!(
                matches!(&params[0].type_ann, Some(TypeExpr::Named { name, .. }) if name == "Int")
            );
            assert_eq!(params[1].name, "y");
            assert!(
                matches!(&params[1].type_ann, Some(TypeExpr::Named { name, .. }) if name == "Int")
            );
        } else {
            panic!("expected lambda expression");
        }
    }

    #[test]
    fn lambda_with_return_type() {
        let (forms, diags) = parse_source("(fn [n : Int] : Int n)");
        assert!(diags.is_empty());
        if let TopForm::Expr(Expr::Lambda {
            params,
            return_type,
            body,
            ..
        }) = &forms[0]
        {
            assert_eq!(params.len(), 1);
            assert_eq!(params[0].name, "n");
            assert!(matches!(return_type, Some(TypeExpr::Named { name, .. }) if name == "Int"));
            assert_eq!(body.len(), 1);
        } else {
            panic!("expected lambda expression");
        }
    }

    #[test]
    fn lambda_multi_body() {
        let (forms, diags) = parse_source("(fn [] (println 1) 42)");
        assert!(diags.is_empty());
        if let TopForm::Expr(Expr::Lambda { body, .. }) = &forms[0] {
            assert_eq!(body.len(), 2);
        } else {
            panic!("expected lambda expression");
        }
    }

    #[test]
    fn lambda_spans() {
        let (forms, _) = parse_source("(fn [] 42)");
        let span = forms[0].span();
        assert_eq!(span.start, 0);
        assert_eq!(span.end, 10);
    }

    #[test]
    fn type_named() {
        let (forms, diags) = parse_source("(fn [x : String] x)");
        assert!(diags.is_empty());
        if let TopForm::Expr(Expr::Lambda { params, .. }) = &forms[0] {
            assert!(
                matches!(&params[0].type_ann, Some(TypeExpr::Named { name, .. }) if name == "String")
            );
        } else {
            panic!("expected lambda expression");
        }
    }

    #[test]
    fn type_applied() {
        let (forms, diags) = parse_source("(fn [xs : (List Int)] xs)");
        assert!(diags.is_empty());
        if let TopForm::Expr(Expr::Lambda { params, .. }) = &forms[0] {
            if let Some(TypeExpr::Applied { name, args, .. }) = &params[0].type_ann {
                assert_eq!(name, "List");
                assert_eq!(args.len(), 1);
                assert!(matches!(&args[0], TypeExpr::Named { name, .. } if name == "Int"));
            } else {
                panic!("expected applied type");
            }
        } else {
            panic!("expected lambda expression");
        }
    }

    #[test]
    fn type_function() {
        let (forms, diags) = parse_source("(fn [f : (Fn [Int Int] Bool)] f)");
        assert!(diags.is_empty());
        if let TopForm::Expr(Expr::Lambda { params, .. }) = &forms[0] {
            if let Some(TypeExpr::Function {
                params: tp, ret, ..
            }) = &params[0].type_ann
            {
                assert_eq!(tp.len(), 2);
                assert!(matches!(&tp[0], TypeExpr::Named { name, .. } if name == "Int"));
                assert!(matches!(&tp[1], TypeExpr::Named { name, .. } if name == "Int"));
                assert!(matches!(ret.as_ref(), TypeExpr::Named { name, .. } if name == "Bool"));
            } else {
                panic!("expected function type");
            }
        } else {
            panic!("expected lambda expression");
        }
    }

    #[test]
    fn error_lambda_missing_bracket() {
        let (_, diags) = parse_source("(fn x 42)");
        assert!(!diags.is_empty());
        assert!(diags[0].message.contains("'['"));
    }

    #[test]
    fn error_type_unexpected() {
        let (_, diags) = parse_source("(fn [x : 42] x)");
        assert!(!diags.is_empty());
        assert!(diags[0].message.contains("expected type"));
    }

    #[test]
    fn defn_no_params() {
        let (forms, diags) = parse_source("(defn main [] (println \"hi\"))");
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
        } else {
            panic!("expected defn");
        }
    }

    #[test]
    fn defn_with_typed_params_and_return() {
        let (forms, diags) = parse_source("(defn add [a : Int b : Int] : Int (+ a b))");
        assert!(diags.is_empty());
        if let TopForm::Defn {
            name,
            params,
            return_type,
            body,
            ..
        } = &forms[0]
        {
            assert_eq!(name, "add");
            assert_eq!(params.len(), 2);
            assert_eq!(params[0].name, "a");
            assert!(params[0].type_ann.is_some());
            assert_eq!(params[1].name, "b");
            assert!(params[1].type_ann.is_some());
            assert!(matches!(return_type, Some(TypeExpr::Named { name, .. }) if name == "Int"));
            assert_eq!(body.len(), 1);
        } else {
            panic!("expected defn");
        }
    }

    #[test]
    fn defn_multi_body() {
        let (forms, diags) = parse_source("(defn f [] (println 1) (println 2) 42)");
        assert!(diags.is_empty());
        if let TopForm::Defn { body, .. } = &forms[0] {
            assert_eq!(body.len(), 3);
        } else {
            panic!("expected defn");
        }
    }

    #[test]
    fn defn_spans() {
        let src = "(defn main [] 42)";
        let (forms, _) = parse_source(src);
        let span = forms[0].span();
        assert_eq!(span.start, 0);
        assert_eq!(span.end, src.len() as u32);
    }

    #[test]
    fn def_simple() {
        let (forms, diags) = parse_source("(def x 42)");
        assert!(diags.is_empty());
        if let TopForm::Def {
            name,
            type_ann,
            value,
            ..
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
    fn def_with_type_annotation() {
        let (forms, diags) = parse_source("(def pi : Float 3.14)");
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
    fn def_spans() {
        let src = "(def x 42)";
        let (forms, _) = parse_source(src);
        let span = forms[0].span();
        assert_eq!(span.start, 0);
        assert_eq!(span.end, src.len() as u32);
    }

    #[test]
    fn multiple_top_forms() {
        let (forms, diags) = parse_source("(def x 1) (defn f [] x)");
        assert!(diags.is_empty());
        assert_eq!(forms.len(), 2);
        assert!(matches!(&forms[0], TopForm::Def { name, .. } if name == "x"));
        assert!(matches!(&forms[1], TopForm::Defn { name, .. } if name == "f"));
    }

    #[test]
    fn mixed_top_forms_and_exprs() {
        let (forms, diags) = parse_source("(def x 1) (+ x 2)");
        assert!(diags.is_empty());
        assert_eq!(forms.len(), 2);
        assert!(matches!(&forms[0], TopForm::Def { .. }));
        assert!(matches!(&forms[1], TopForm::Expr(Expr::Call { .. })));
    }

    #[test]
    fn error_defn_missing_name() {
        let (_, diags) = parse_source("(defn [] 42)");
        assert!(!diags.is_empty());
        assert!(diags[0].message.contains("expected symbol"));
    }

    #[test]
    fn error_def_missing_value() {
        let (_, diags) = parse_source("(def x)");
        assert!(!diags.is_empty());
    }

    #[test]
    fn hello_world_integration() {
        let src = "(defn main [] (println \"Hello, World!\"))";
        let (forms, diags) = parse_source(src);
        assert!(diags.is_empty());
        assert_eq!(forms.len(), 1);
        if let TopForm::Defn {
            name, params, body, ..
        } = &forms[0]
        {
            assert_eq!(name, "main");
            assert!(params.is_empty());
            assert_eq!(body.len(), 1);
            if let Expr::Call { func, args, .. } = &body[0] {
                assert!(matches!(func.as_ref(), Expr::Symbol(s, _) if s == "println"));
                assert_eq!(args.len(), 1);
                assert!(matches!(&args[0], Expr::String(s, _) if s == "Hello, World!"));
            } else {
                panic!("expected call in body");
            }
        } else {
            panic!("expected defn");
        }
    }

    #[test]
    fn fibonacci_integration() {
        let src = r#"(defn fib [n : Int] : Int
  (if (<= n 1)
    n
    (+ (fib (- n 1)) (fib (- n 2)))))"#;
        let (forms, diags) = parse_source(src);
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
            assert_eq!(name, "fib");
            assert_eq!(params.len(), 1);
            assert_eq!(params[0].name, "n");
            assert!(params[0].type_ann.is_some());
            assert!(return_type.is_some());
            assert_eq!(body.len(), 1);
            assert!(matches!(&body[0], Expr::If { .. }));
        } else {
            panic!("expected defn");
        }
    }

    #[test]
    fn multi_function_program() {
        let src = r#"(def pi : Float 3.14)

(defn circle-area [r : Float] : Float
  (* pi (* r r)))

(defn main []
  (println (circle-area 5.0)))"#;
        let (forms, diags) = parse_source(src);
        assert!(diags.is_empty());
        assert_eq!(forms.len(), 3);
        assert!(matches!(&forms[0], TopForm::Def { name, .. } if name == "pi"));
        assert!(matches!(&forms[1], TopForm::Defn { name, .. } if name == "circle-area"));
        assert!(matches!(&forms[2], TopForm::Defn { name, .. } if name == "main"));
    }

    #[test]
    fn deftype_simple() {
        let (forms, diags) = parse_source("(deftype Point (x Float) (y Float))");
        assert!(diags.is_empty());
        assert_eq!(forms.len(), 1);
        if let TopForm::Deftype { name, fields, .. } = &forms[0] {
            assert_eq!(name, "Point");
            assert_eq!(fields.len(), 2);
            assert_eq!(fields[0].name, "x");
            assert!(
                matches!(&fields[0].type_expr, TypeExpr::Named { name, .. } if name == "Float")
            );
            assert_eq!(fields[1].name, "y");
            assert!(
                matches!(&fields[1].type_expr, TypeExpr::Named { name, .. } if name == "Float")
            );
        } else {
            panic!("expected deftype");
        }
    }

    #[test]
    fn deftype_no_fields() {
        let (forms, diags) = parse_source("(deftype Empty)");
        assert!(diags.is_empty());
        if let TopForm::Deftype { name, fields, .. } = &forms[0] {
            assert_eq!(name, "Empty");
            assert!(fields.is_empty());
        } else {
            panic!("expected deftype");
        }
    }

    #[test]
    fn deftype_complex_field_types() {
        let (forms, diags) =
            parse_source("(deftype Config (name String) (handler (Fn [Int] Bool)))");
        assert!(diags.is_empty());
        if let TopForm::Deftype { name, fields, .. } = &forms[0] {
            assert_eq!(name, "Config");
            assert_eq!(fields.len(), 2);
            assert_eq!(fields[0].name, "name");
            assert!(
                matches!(&fields[0].type_expr, TypeExpr::Named { name, .. } if name == "String")
            );
            assert!(matches!(&fields[1].type_expr, TypeExpr::Function { .. }));
        } else {
            panic!("expected deftype");
        }
    }

    #[test]
    fn deftype_spans() {
        let src = "(deftype Point (x Int) (y Int))";
        let (forms, _) = parse_source(src);
        let span = forms[0].span();
        assert_eq!(span.start, 0);
        assert_eq!(span.end, src.len() as u32);
    }

    #[test]
    fn deftype_with_program() {
        let src = r#"(deftype Point (x Float) (y Float))

(defn main []
  (println "hello"))"#;
        let (forms, diags) = parse_source(src);
        assert!(diags.is_empty());
        assert_eq!(forms.len(), 2);
        assert!(matches!(&forms[0], TopForm::Deftype { name, .. } if name == "Point"));
        assert!(matches!(&forms[1], TopForm::Defn { name, .. } if name == "main"));
    }

    #[test]
    fn error_deftype_missing_name() {
        let (_, diags) = parse_source("(deftype)");
        assert!(!diags.is_empty());
        assert!(diags[0].message.contains("expected symbol"));
    }

    #[test]
    fn error_deftype_bad_field() {
        let (_, diags) = parse_source("(deftype Foo x)");
        assert!(!diags.is_empty());
        assert!(diags[0].message.contains("'('"));
    }
}
