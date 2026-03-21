use crate::ast::{Expr, TopForm};
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
}
