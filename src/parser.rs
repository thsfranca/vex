use crate::ast::{
    Binding, CondClause, Expr, Field, MatchClause, Param, Pattern, TopForm, TypeExpr, Variant,
};
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
                "module" => {
                    let open_span = self.tokens[self.pos].span;
                    self.pos += 1;
                    return self.parse_module(open_span);
                }
                "export" => {
                    let open_span = self.tokens[self.pos].span;
                    self.pos += 1;
                    return self.parse_export(open_span);
                }
                "import" => {
                    let open_span = self.tokens[self.pos].span;
                    self.pos += 1;
                    return self.parse_import(open_span);
                }
                "import-go" => {
                    let open_span = self.tokens[self.pos].span;
                    self.pos += 1;
                    return self.parse_import_go(open_span);
                }
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
                "defunion" => {
                    let open_span = self.tokens[self.pos].span;
                    self.pos += 1;
                    return self.parse_defunion(open_span);
                }
                _ => {}
            }
        }

        let expr = self.parse_expr()?;
        Some(TopForm::Expr(expr))
    }

    fn parse_qualified_id(&mut self) -> Option<(String, Span)> {
        let (first, start_span) = self.expect_symbol()?;
        let mut name = first;
        let mut end_span = start_span;

        while self.check(|k| matches!(k, TokenKind::Dot)) {
            self.pos += 1;
            let (part, part_span) = self.expect_symbol()?;
            name.push('.');
            name.push_str(&part);
            end_span = part_span;
        }

        Some((
            name,
            Span::new(start_span.file, start_span.start, end_span.end),
        ))
    }

    fn parse_module(&mut self, open_span: Span) -> Option<TopForm> {
        self.pos += 1;

        let (name, _) = self.parse_qualified_id()?;
        let close_span = self.expect_right_paren(open_span)?;

        Some(TopForm::Module {
            name,
            span: Span::new(open_span.file, open_span.start, close_span.end),
        })
    }

    fn parse_export(&mut self, open_span: Span) -> Option<TopForm> {
        self.pos += 1;

        if self.at_end() || !self.check(|k| matches!(k, TokenKind::LeftBracket)) {
            let span = if self.at_end() {
                self.eof_span()
            } else {
                self.tokens[self.pos].span
            };
            self.diagnostics.push(Diagnostic::error(
                "expected '[' for export symbol list",
                span,
            ));
            return None;
        }
        let bracket_span = self.tokens[self.pos].span;
        self.pos += 1;

        let mut symbols = Vec::new();
        while !self.at_end() && !self.check(|k| matches!(k, TokenKind::RightBracket)) {
            let (sym, _) = self.expect_symbol()?;
            symbols.push(sym);
        }

        self.expect_right_bracket(bracket_span)?;
        let close_span = self.expect_right_paren(open_span)?;

        Some(TopForm::Export {
            symbols,
            span: Span::new(open_span.file, open_span.start, close_span.end),
        })
    }

    fn parse_import(&mut self, open_span: Span) -> Option<TopForm> {
        self.pos += 1;

        let (module_path, _) = self.parse_qualified_id()?;

        if self.at_end() || !self.check(|k| matches!(k, TokenKind::LeftBracket)) {
            let span = if self.at_end() {
                self.eof_span()
            } else {
                self.tokens[self.pos].span
            };
            self.diagnostics.push(Diagnostic::error(
                "expected '[' for import symbol list",
                span,
            ));
            return None;
        }
        let bracket_span = self.tokens[self.pos].span;
        self.pos += 1;

        let mut symbols = Vec::new();
        while !self.at_end() && !self.check(|k| matches!(k, TokenKind::RightBracket)) {
            let (sym, _) = self.expect_symbol()?;
            symbols.push(sym);
        }

        self.expect_right_bracket(bracket_span)?;
        let close_span = self.expect_right_paren(open_span)?;

        Some(TopForm::Import {
            module_path,
            symbols,
            span: Span::new(open_span.file, open_span.start, close_span.end),
        })
    }

    fn parse_import_go(&mut self, open_span: Span) -> Option<TopForm> {
        self.pos += 1;

        if self.at_end() {
            self.diagnostics.push(Diagnostic::error(
                "expected Go package path string",
                self.eof_span(),
            ));
            return None;
        }
        let go_package = if let TokenKind::String(ref s) = self.tokens[self.pos].kind {
            let s = s.clone();
            self.pos += 1;
            s
        } else {
            let desc = token_kind_name(&self.tokens[self.pos].kind);
            self.diagnostics.push(Diagnostic::error(
                format!("expected Go package path string, found {}", desc),
                self.tokens[self.pos].span,
            ));
            return None;
        };

        if self.at_end() || !self.check(|k| matches!(k, TokenKind::LeftBracket)) {
            let span = if self.at_end() {
                self.eof_span()
            } else {
                self.tokens[self.pos].span
            };
            self.diagnostics.push(Diagnostic::error(
                "expected '[' for import-go symbol list",
                span,
            ));
            return None;
        }
        let bracket_span = self.tokens[self.pos].span;
        self.pos += 1;

        let mut symbols = Vec::new();
        while !self.at_end() && !self.check(|k| matches!(k, TokenKind::RightBracket)) {
            let (sym, _) = self.expect_symbol()?;
            symbols.push(sym);
        }

        self.expect_right_bracket(bracket_span)?;
        let close_span = self.expect_right_paren(open_span)?;

        Some(TopForm::ImportGo {
            go_package,
            symbols,
            span: Span::new(open_span.file, open_span.start, close_span.end),
        })
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

        if self.check(|k| matches!(k, TokenKind::Dot)) {
            return self.parse_field_access(open_span);
        }

        if let Some(TokenKind::Symbol(name)) = self.tokens.get(self.pos).map(|t| &t.kind) {
            match name.as_str() {
                "if" => return self.parse_if(open_span),
                "let" => return self.parse_let(open_span),
                "cond" => return self.parse_cond(open_span),
                "match" => return self.parse_match(open_span),
                "fn" => return self.parse_lambda(open_span),
                "spawn" => return self.parse_spawn(open_span),
                "channel" => return self.parse_channel(open_span),
                "send" => return self.parse_send(open_span),
                "recv" => return self.parse_recv(open_span),
                _ => {}
            }
        }

        self.parse_call(open_span)
    }

    fn parse_spawn(&mut self, open_span: Span) -> Option<Expr> {
        self.pos += 1;
        let body = self.parse_expr()?;
        let close_span = self.expect_right_paren(open_span)?;
        Some(Expr::Spawn {
            body: Box::new(body),
            span: Span::new(open_span.file, open_span.start, close_span.end),
        })
    }

    fn parse_channel(&mut self, open_span: Span) -> Option<Expr> {
        self.pos += 1;
        let element_type = self.parse_type()?;
        let size = if !self.at_end() && !self.check(|k| matches!(k, TokenKind::RightParen)) {
            Some(Box::new(self.parse_expr()?))
        } else {
            None
        };
        let close_span = self.expect_right_paren(open_span)?;
        Some(Expr::Channel {
            element_type,
            size,
            span: Span::new(open_span.file, open_span.start, close_span.end),
        })
    }

    fn parse_send(&mut self, open_span: Span) -> Option<Expr> {
        self.pos += 1;
        let channel = self.parse_expr()?;
        let value = self.parse_expr()?;
        let close_span = self.expect_right_paren(open_span)?;
        Some(Expr::Send {
            channel: Box::new(channel),
            value: Box::new(value),
            span: Span::new(open_span.file, open_span.start, close_span.end),
        })
    }

    fn parse_recv(&mut self, open_span: Span) -> Option<Expr> {
        self.pos += 1;
        let channel = self.parse_expr()?;
        let close_span = self.expect_right_paren(open_span)?;
        Some(Expr::Recv {
            channel: Box::new(channel),
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

    fn parse_match(&mut self, open_span: Span) -> Option<Expr> {
        self.pos += 1;

        let scrutinee = Box::new(self.parse_expr()?);

        let mut clauses = Vec::new();
        while !self.at_end() && !self.check(|k| matches!(k, TokenKind::RightParen)) {
            clauses.push(self.parse_match_clause()?);
        }

        if clauses.is_empty() {
            self.diagnostics.push(Diagnostic::error(
                "match requires at least one clause",
                open_span,
            ));
            return None;
        }

        let close_span = self.expect_right_paren(open_span)?;

        Some(Expr::Match {
            scrutinee,
            clauses,
            span: Span::new(open_span.file, open_span.start, close_span.end),
        })
    }

    fn parse_match_clause(&mut self) -> Option<MatchClause> {
        let pattern = self.parse_pattern()?;
        let body = self.parse_expr()?;
        let span = Span::new(pattern.span().file, pattern.span().start, body.span().end);
        Some(MatchClause {
            pattern,
            body,
            span,
        })
    }

    fn parse_pattern(&mut self) -> Option<Pattern> {
        if self.at_end() {
            self.diagnostics.push(Diagnostic::error(
                "expected pattern but reached end of input",
                self.eof_span(),
            ));
            return None;
        }

        let span = self.tokens[self.pos].span;
        match &self.tokens[self.pos].kind {
            TokenKind::Integer(n) => {
                let n = *n;
                self.pos += 1;
                Some(Pattern::Literal(Box::new(Expr::Int(n, span))))
            }
            TokenKind::Float(f) => {
                let f = *f;
                self.pos += 1;
                Some(Pattern::Literal(Box::new(Expr::Float(f, span))))
            }
            TokenKind::String(s) => {
                let s = s.clone();
                self.pos += 1;
                Some(Pattern::Literal(Box::new(Expr::String(s, span))))
            }
            TokenKind::Boolean(b) => {
                let b = *b;
                self.pos += 1;
                Some(Pattern::Literal(Box::new(Expr::Bool(b, span))))
            }
            TokenKind::Nil => {
                self.pos += 1;
                Some(Pattern::Literal(Box::new(Expr::Nil(span))))
            }
            TokenKind::Symbol(s) if s == "_" => {
                self.pos += 1;
                Some(Pattern::Wildcard(span))
            }
            TokenKind::Symbol(s) => {
                let s = s.clone();
                self.pos += 1;
                Some(Pattern::Binding(s, span))
            }
            TokenKind::LeftParen => {
                self.pos += 1;
                let (name, _) = self.expect_symbol()?;
                let mut args = Vec::new();
                while !self.at_end() && !self.check(|k| matches!(k, TokenKind::RightParen)) {
                    args.push(self.parse_pattern()?);
                }
                let close_span = self.expect_right_paren(span)?;
                Some(Pattern::Constructor {
                    name,
                    args,
                    span: Span::new(span.file, span.start, close_span.end),
                })
            }
            _ => {
                let desc = token_kind_name(&self.tokens[self.pos].kind);
                self.diagnostics.push(Diagnostic::error(
                    format!("expected pattern, found {}", desc),
                    span,
                ));
                None
            }
        }
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

    fn parse_field_access(&mut self, open_span: Span) -> Option<Expr> {
        self.pos += 1;

        let object = self.parse_expr()?;
        let (field, _) = self.expect_symbol()?;
        let close_span = self.expect_right_paren(open_span)?;

        Some(Expr::FieldAccess {
            object: Box::new(object),
            field,
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

    fn parse_defunion(&mut self, open_span: Span) -> Option<TopForm> {
        self.pos += 1;

        let (name, _) = self.expect_symbol()?;

        let mut variants = Vec::new();
        while !self.at_end() && !self.check(|k| matches!(k, TokenKind::RightParen)) {
            variants.push(self.parse_variant()?);
        }

        let close_span = self.expect_right_paren(open_span)?;

        Some(TopForm::Defunion {
            name,
            variants,
            span: Span::new(open_span.file, open_span.start, close_span.end),
        })
    }

    fn parse_variant(&mut self) -> Option<Variant> {
        if self.at_end() {
            self.diagnostics.push(Diagnostic::error(
                "expected '(' for variant declaration but reached end of input",
                self.eof_span(),
            ));
            return None;
        }

        let open_span = self.tokens[self.pos].span;
        if !matches!(self.tokens[self.pos].kind, TokenKind::LeftParen) {
            let desc = token_kind_name(&self.tokens[self.pos].kind);
            self.diagnostics.push(Diagnostic::error(
                format!("expected '(' for variant declaration, found {}", desc),
                open_span,
            ));
            return None;
        }
        self.pos += 1;

        let (name, _) = self.expect_symbol()?;

        let mut types = Vec::new();
        while !self.at_end() && !self.check(|k| matches!(k, TokenKind::RightParen)) {
            types.push(self.parse_type()?);
        }

        let close_span = self.expect_right_paren(open_span)?;

        Some(Variant {
            name,
            types,
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

    #[test]
    fn field_access_simple() {
        let (forms, diags) = parse_source("(. point x)");
        assert!(diags.is_empty());
        assert_eq!(forms.len(), 1);
        if let TopForm::Expr(Expr::FieldAccess { object, field, .. }) = &forms[0] {
            assert!(matches!(object.as_ref(), Expr::Symbol(s, _) if s == "point"));
            assert_eq!(field, "x");
        } else {
            panic!("expected field access expression");
        }
    }

    #[test]
    fn field_access_nested() {
        let (forms, diags) = parse_source("(. (. line start) x)");
        assert!(diags.is_empty());
        if let TopForm::Expr(Expr::FieldAccess { object, field, .. }) = &forms[0] {
            assert!(matches!(object.as_ref(), Expr::FieldAccess { .. }));
            assert_eq!(field, "x");
        } else {
            panic!("expected field access expression");
        }
    }

    #[test]
    fn field_access_spans() {
        let src = "(. point x)";
        let (forms, _) = parse_source(src);
        let span = forms[0].span();
        assert_eq!(span.start, 0);
        assert_eq!(span.end, src.len() as u32);
    }

    #[test]
    fn defunion_simple() {
        let source = "(defunion Shape (Circle Float) (Rect Float Float))";
        let (forms, diags) = parse_source(source);
        assert!(diags.is_empty());
        assert_eq!(forms.len(), 1);
        if let TopForm::Defunion { name, variants, .. } = &forms[0] {
            assert_eq!(name, "Shape");
            assert_eq!(variants.len(), 2);
            assert_eq!(variants[0].name, "Circle");
            assert_eq!(variants[0].types.len(), 1);
            assert_eq!(variants[1].name, "Rect");
            assert_eq!(variants[1].types.len(), 2);
        } else {
            panic!("expected defunion");
        }
    }

    #[test]
    fn defunion_no_data_variant() {
        let source = "(defunion Option (Some Int) (None))";
        let (forms, diags) = parse_source(source);
        assert!(diags.is_empty());
        if let TopForm::Defunion { variants, .. } = &forms[0] {
            assert_eq!(variants[0].name, "Some");
            assert_eq!(variants[0].types.len(), 1);
            assert_eq!(variants[1].name, "None");
            assert!(variants[1].types.is_empty());
        } else {
            panic!("expected defunion");
        }
    }

    #[test]
    fn defunion_spans() {
        let source = "(defunion Msg (Req Int) (Resp String))";
        let (forms, diags) = parse_source(source);
        assert!(diags.is_empty());
        let span = forms[0].span();
        assert_eq!(span.start, 0);
        assert_eq!(span.end, source.len() as u32);
    }

    #[test]
    fn error_defunion_missing_name() {
        let source = "(defunion)";
        let (_, diags) = parse_source(source);
        assert!(!diags.is_empty());
    }

    #[test]
    fn error_defunion_bad_variant() {
        let source = "(defunion Foo bar)";
        let (_, diags) = parse_source(source);
        assert!(!diags.is_empty());
    }

    #[test]
    fn match_constructor_patterns() {
        let source = "(match x (Some v) v _ nil)";
        let (forms, diags) = parse_source(source);
        assert!(diags.is_empty(), "{:?}", diags);
        if let TopForm::Expr(Expr::Match {
            scrutinee, clauses, ..
        }) = &forms[0]
        {
            assert!(matches!(scrutinee.as_ref(), Expr::Symbol(s, _) if s == "x"));
            assert_eq!(clauses.len(), 2);
            assert!(
                matches!(&clauses[0].pattern, Pattern::Constructor { name, args, .. }
                if name == "Some" && args.len() == 1)
            );
            assert!(matches!(&clauses[1].pattern, Pattern::Wildcard(_)));
        } else {
            panic!("expected match");
        }
    }

    #[test]
    fn match_literal_patterns() {
        let source = r#"(match x 1 "one" 2 "two" _ "other")"#;
        let (forms, diags) = parse_source(source);
        assert!(diags.is_empty(), "{:?}", diags);
        if let TopForm::Expr(Expr::Match { clauses, .. }) = &forms[0] {
            assert_eq!(clauses.len(), 3);
            assert!(
                matches!(&clauses[0].pattern, Pattern::Literal(e) if matches!(e.as_ref(), Expr::Int(1, _)))
            );
            assert!(
                matches!(&clauses[1].pattern, Pattern::Literal(e) if matches!(e.as_ref(), Expr::Int(2, _)))
            );
            assert!(matches!(&clauses[2].pattern, Pattern::Wildcard(_)));
        } else {
            panic!("expected match");
        }
    }

    #[test]
    fn match_spans() {
        let source = "(match x _ nil)";
        let (forms, diags) = parse_source(source);
        assert!(diags.is_empty());
        let span = forms[0].span();
        assert_eq!(span.start, 0);
        assert_eq!(span.end, source.len() as u32);
    }

    #[test]
    fn error_match_no_clauses() {
        let source = "(match x)";
        let (_, diags) = parse_source(source);
        assert!(!diags.is_empty());
    }

    #[test]
    fn module_simple() {
        let (forms, diags) = parse_source("(module myapp)");
        assert!(diags.is_empty());
        assert_eq!(forms.len(), 1);
        assert!(matches!(&forms[0], TopForm::Module { name, .. } if name == "myapp"));
    }

    #[test]
    fn module_qualified() {
        let (forms, diags) = parse_source("(module vex.http.server)");
        assert!(diags.is_empty());
        assert_eq!(forms.len(), 1);
        assert!(matches!(&forms[0], TopForm::Module { name, .. } if name == "vex.http.server"));
    }

    #[test]
    fn export_symbols() {
        let (forms, diags) = parse_source("(export [foo bar baz])");
        assert!(diags.is_empty());
        assert_eq!(forms.len(), 1);
        if let TopForm::Export { symbols, .. } = &forms[0] {
            assert_eq!(symbols, &["foo", "bar", "baz"]);
        } else {
            panic!("expected Export");
        }
    }

    #[test]
    fn export_empty() {
        let (forms, diags) = parse_source("(export [])");
        assert!(diags.is_empty());
        assert_eq!(forms.len(), 1);
        if let TopForm::Export { symbols, .. } = &forms[0] {
            assert!(symbols.is_empty());
        } else {
            panic!("expected Export");
        }
    }

    #[test]
    fn import_simple() {
        let (forms, diags) = parse_source("(import math [add sub])");
        assert!(diags.is_empty());
        assert_eq!(forms.len(), 1);
        if let TopForm::Import {
            module_path,
            symbols,
            ..
        } = &forms[0]
        {
            assert_eq!(module_path, "math");
            assert_eq!(symbols, &["add", "sub"]);
        } else {
            panic!("expected Import");
        }
    }

    #[test]
    fn import_qualified() {
        let (forms, diags) = parse_source("(import vex.http [get post])");
        assert!(diags.is_empty());
        assert_eq!(forms.len(), 1);
        if let TopForm::Import {
            module_path,
            symbols,
            ..
        } = &forms[0]
        {
            assert_eq!(module_path, "vex.http");
            assert_eq!(symbols, &["get", "post"]);
        } else {
            panic!("expected Import");
        }
    }

    #[test]
    fn module_export_import_together() {
        let source = r#"
(module myapp)
(export [main])
(import math [add])
(defn main [] (add 1 2))
"#;
        let (forms, diags) = parse_source(source);
        assert!(diags.is_empty());
        assert_eq!(forms.len(), 4);
        assert!(matches!(&forms[0], TopForm::Module { name, .. } if name == "myapp"));
        assert!(matches!(&forms[1], TopForm::Export { symbols, .. } if symbols == &["main"]));
        assert!(
            matches!(&forms[2], TopForm::Import { module_path, symbols, .. } if module_path == "math" && symbols == &["add"])
        );
        assert!(matches!(&forms[3], TopForm::Defn { name, .. } if name == "main"));
    }

    #[test]
    fn error_export_no_bracket() {
        let (_, diags) = parse_source("(export foo)");
        assert!(!diags.is_empty());
    }

    #[test]
    fn error_import_no_bracket() {
        let (_, diags) = parse_source("(import math foo)");
        assert!(!diags.is_empty());
    }

    #[test]
    fn import_go_simple() {
        let (forms, diags) = parse_source(r#"(import-go "net/http" [Get Post])"#);
        assert!(diags.is_empty(), "{:?}", diags);
        assert_eq!(forms.len(), 1);
        if let TopForm::ImportGo {
            go_package,
            symbols,
            ..
        } = &forms[0]
        {
            assert_eq!(go_package, "net/http");
            assert_eq!(symbols, &["Get", "Post"]);
        } else {
            panic!("expected ImportGo");
        }
    }

    #[test]
    fn import_go_single_symbol() {
        let (forms, diags) = parse_source(r#"(import-go "os" [Exit])"#);
        assert!(diags.is_empty(), "{:?}", diags);
        if let TopForm::ImportGo {
            go_package,
            symbols,
            ..
        } = &forms[0]
        {
            assert_eq!(go_package, "os");
            assert_eq!(symbols, &["Exit"]);
        } else {
            panic!("expected ImportGo");
        }
    }

    #[test]
    fn error_import_go_no_string() {
        let (_, diags) = parse_source("(import-go foo [bar])");
        assert!(!diags.is_empty());
    }

    #[test]
    fn error_import_go_no_bracket() {
        let (_, diags) = parse_source(r#"(import-go "net/http" Get)"#);
        assert!(!diags.is_empty());
    }

    #[test]
    fn spawn_simple() {
        let (forms, diags) = parse_source("(spawn (println \"hi\"))");
        assert!(diags.is_empty(), "{:?}", diags);
        assert_eq!(forms.len(), 1);
        if let TopForm::Expr(Expr::Spawn { body, .. }) = &forms[0] {
            assert!(matches!(body.as_ref(), Expr::Call { .. }));
        } else {
            panic!("expected Spawn");
        }
    }

    #[test]
    fn spawn_spans() {
        let src = "(spawn (foo))";
        let (forms, diags) = parse_source(src);
        assert!(diags.is_empty(), "{:?}", diags);
        let span = forms[0].span();
        assert_eq!(span.start, 0);
        assert_eq!(span.end, src.len() as u32);
    }

    #[test]
    fn channel_buffered() {
        let (forms, diags) = parse_source("(channel Int 10)");
        assert!(diags.is_empty(), "{:?}", diags);
        assert_eq!(forms.len(), 1);
        if let TopForm::Expr(Expr::Channel {
            element_type, size, ..
        }) = &forms[0]
        {
            assert!(matches!(element_type, TypeExpr::Named { name, .. } if name == "Int"));
            assert!(matches!(size.as_deref(), Some(Expr::Int(10, _))));
        } else {
            panic!("expected Channel");
        }
    }

    #[test]
    fn channel_unbuffered() {
        let (forms, diags) = parse_source("(channel String)");
        assert!(diags.is_empty(), "{:?}", diags);
        assert_eq!(forms.len(), 1);
        if let TopForm::Expr(Expr::Channel {
            element_type, size, ..
        }) = &forms[0]
        {
            assert!(matches!(element_type, TypeExpr::Named { name, .. } if name == "String"));
            assert!(size.is_none());
        } else {
            panic!("expected Channel");
        }
    }

    #[test]
    fn channel_spans() {
        let src = "(channel Int 10)";
        let (forms, diags) = parse_source(src);
        assert!(diags.is_empty(), "{:?}", diags);
        let span = forms[0].span();
        assert_eq!(span.start, 0);
        assert_eq!(span.end, src.len() as u32);
    }

    #[test]
    fn send_simple() {
        let (forms, diags) = parse_source("(send ch 42)");
        assert!(diags.is_empty(), "{:?}", diags);
        assert_eq!(forms.len(), 1);
        if let TopForm::Expr(Expr::Send { channel, value, .. }) = &forms[0] {
            assert!(matches!(channel.as_ref(), Expr::Symbol(s, _) if s == "ch"));
            assert!(matches!(value.as_ref(), Expr::Int(42, _)));
        } else {
            panic!("expected Send");
        }
    }

    #[test]
    fn send_spans() {
        let src = "(send ch 42)";
        let (forms, diags) = parse_source(src);
        assert!(diags.is_empty(), "{:?}", diags);
        let span = forms[0].span();
        assert_eq!(span.start, 0);
        assert_eq!(span.end, src.len() as u32);
    }

    #[test]
    fn recv_simple() {
        let (forms, diags) = parse_source("(recv ch)");
        assert!(diags.is_empty(), "{:?}", diags);
        assert_eq!(forms.len(), 1);
        if let TopForm::Expr(Expr::Recv { channel, .. }) = &forms[0] {
            assert!(matches!(channel.as_ref(), Expr::Symbol(s, _) if s == "ch"));
        } else {
            panic!("expected Recv");
        }
    }

    #[test]
    fn recv_spans() {
        let src = "(recv ch)";
        let (forms, diags) = parse_source(src);
        assert!(diags.is_empty(), "{:?}", diags);
        let span = forms[0].span();
        assert_eq!(span.start, 0);
        assert_eq!(span.end, src.len() as u32);
    }

    #[test]
    fn concurrency_integration() {
        let source = r#"
(defn main []
  (let [ch (channel Int 10)]
    (spawn
      (send ch 42))
    (println (str (recv ch)))))
"#;
        let (forms, diags) = parse_source(source);
        assert!(diags.is_empty(), "{:?}", diags);
        assert_eq!(forms.len(), 1);
        assert!(matches!(&forms[0], TopForm::Defn { name, .. } if name == "main"));
    }
}
