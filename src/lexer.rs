use crate::diagnostics::Diagnostic;
use crate::source::{FileId, Span};

#[derive(Debug, Clone, PartialEq)]
pub enum TokenKind {
    LeftParen,
    RightParen,
    LeftBracket,
    RightBracket,
    LeftBrace,
    RightBrace,
    Dot,
    Colon,
    Symbol(String),
    Keyword(String),
    Integer(i64),
    Float(f64),
    String(String),
    Boolean(bool),
    Nil,
}

#[derive(Debug, Clone, PartialEq)]
pub struct Token {
    pub kind: TokenKind,
    pub span: Span,
}

impl Token {
    fn new(kind: TokenKind, span: Span) -> Self {
        Self { kind, span }
    }
}

struct Lexer<'a> {
    source: &'a str,
    bytes: &'a [u8],
    pos: usize,
    file: FileId,
    diagnostics: Vec<Diagnostic>,
}

impl<'a> Lexer<'a> {
    fn new(source: &'a str, file: FileId) -> Self {
        Self {
            source,
            bytes: source.as_bytes(),
            pos: 0,
            file,
            diagnostics: Vec::new(),
        }
    }

    fn peek(&self) -> Option<u8> {
        self.bytes.get(self.pos).copied()
    }

    fn advance(&mut self) -> Option<u8> {
        let b = self.bytes.get(self.pos).copied()?;
        self.pos += 1;
        Some(b)
    }

    fn span(&self, start: usize) -> Span {
        Span::new(self.file, start as u32, self.pos as u32)
    }

    fn skip_whitespace_and_comments(&mut self) {
        loop {
            match self.peek() {
                Some(b' ' | b'\t' | b'\n' | b'\r' | b',') => {
                    self.advance();
                }
                Some(b';') => {
                    while let Some(b) = self.peek() {
                        if b == b'\n' {
                            break;
                        }
                        self.advance();
                    }
                }
                _ => break,
            }
        }
    }

    fn lex_string(&mut self) -> Token {
        let start = self.pos;
        self.advance();
        let mut value = String::new();

        loop {
            match self.advance() {
                None => {
                    self.diagnostics.push(Diagnostic::error(
                        "unterminated string literal",
                        self.span(start),
                    ));
                    return Token::new(TokenKind::String(value), self.span(start));
                }
                Some(b'"') => {
                    return Token::new(TokenKind::String(value), self.span(start));
                }
                Some(b'\\') => match self.advance() {
                    None => {
                        self.diagnostics.push(Diagnostic::error(
                            "unterminated string literal",
                            self.span(start),
                        ));
                        return Token::new(TokenKind::String(value), self.span(start));
                    }
                    Some(b'\\') => value.push('\\'),
                    Some(b'"') => value.push('"'),
                    Some(b'n') => value.push('\n'),
                    Some(b't') => value.push('\t'),
                    Some(b'r') => value.push('\r'),
                    Some(b'0') => value.push('\0'),
                    Some(b'u') => {
                        let esc_start = self.pos - 2;
                        let mut hex = String::with_capacity(4);
                        for _ in 0..4 {
                            match self.advance() {
                                Some(b) if b.is_ascii_hexdigit() => hex.push(b as char),
                                _ => {
                                    self.diagnostics.push(Diagnostic::error(
                                        "invalid unicode escape: expected 4 hex digits",
                                        self.span(esc_start),
                                    ));
                                    break;
                                }
                            }
                        }
                        if hex.len() == 4 {
                            if let Some(ch) =
                                u32::from_str_radix(&hex, 16).ok().and_then(char::from_u32)
                            {
                                value.push(ch);
                            } else {
                                self.diagnostics.push(Diagnostic::error(
                                    "invalid unicode scalar value",
                                    self.span(esc_start),
                                ));
                            }
                        }
                    }
                    Some(b) => {
                        let esc_start = self.pos - 2;
                        self.diagnostics.push(Diagnostic::error(
                            format!("invalid escape sequence: \\{}", b as char),
                            self.span(esc_start),
                        ));
                    }
                },
                Some(b) => {
                    if b < 0x80 {
                        value.push(b as char);
                    } else {
                        let ch_start = self.pos - 1;
                        let rest = &self.source[ch_start..];
                        if let Some(ch) = rest.chars().next() {
                            value.push(ch);
                            self.pos = ch_start + ch.len_utf8();
                        }
                    }
                }
            }
        }
    }

    fn lex_number(&mut self, start: usize) -> Token {
        while let Some(b) = self.peek() {
            if b.is_ascii_digit() {
                self.advance();
            } else {
                break;
            }
        }

        if self.peek() == Some(b'.') {
            let dot_pos = self.pos;
            if let Some(next) = self.bytes.get(dot_pos + 1)
                && next.is_ascii_digit()
            {
                self.advance();
                while let Some(b) = self.peek() {
                    if b.is_ascii_digit() {
                        self.advance();
                    } else {
                        break;
                    }
                }
                let text = &self.source[start..self.pos];
                let val: f64 = text.parse().unwrap_or(0.0);
                return Token::new(TokenKind::Float(val), self.span(start));
            }
        }

        let text = &self.source[start..self.pos];
        let val: i64 = text.parse().unwrap_or(0);
        Token::new(TokenKind::Integer(val), self.span(start))
    }

    fn is_ident_start(b: u8) -> bool {
        b.is_ascii_alphabetic() || b == b'_'
    }

    fn is_ident_continue(b: u8) -> bool {
        b.is_ascii_alphanumeric() || matches!(b, b'-' | b'_' | b'!' | b'?')
    }

    fn is_operator_char(b: u8) -> bool {
        matches!(
            b,
            b'+' | b'-'
                | b'*'
                | b'/'
                | b'<'
                | b'>'
                | b'='
                | b'!'
                | b'&'
                | b'|'
                | b'%'
                | b'^'
                | b'~'
        )
    }

    fn lex_alpha_symbol(&mut self) -> Token {
        let start = self.pos;
        self.advance();
        while let Some(b) = self.peek() {
            if Self::is_ident_continue(b) {
                self.advance();
            } else {
                break;
            }
        }
        let text = &self.source[start..self.pos];
        let kind = match text {
            "true" => TokenKind::Boolean(true),
            "false" => TokenKind::Boolean(false),
            "nil" => TokenKind::Nil,
            _ => TokenKind::Symbol(text.to_string()),
        };
        Token::new(kind, self.span(start))
    }

    fn lex_operator_symbol(&mut self) -> Token {
        let start = self.pos;
        while let Some(b) = self.peek() {
            if Self::is_operator_char(b) {
                self.advance();
            } else {
                break;
            }
        }
        let text = &self.source[start..self.pos];
        Token::new(TokenKind::Symbol(text.to_string()), self.span(start))
    }

    fn next_token(&mut self) -> Option<Token> {
        self.skip_whitespace_and_comments();

        let b = self.peek()?;
        match b {
            b'(' => {
                let start = self.pos;
                self.advance();
                Some(Token::new(TokenKind::LeftParen, self.span(start)))
            }
            b')' => {
                let start = self.pos;
                self.advance();
                Some(Token::new(TokenKind::RightParen, self.span(start)))
            }
            b'[' => {
                let start = self.pos;
                self.advance();
                Some(Token::new(TokenKind::LeftBracket, self.span(start)))
            }
            b']' => {
                let start = self.pos;
                self.advance();
                Some(Token::new(TokenKind::RightBracket, self.span(start)))
            }
            b'{' => {
                let start = self.pos;
                self.advance();
                Some(Token::new(TokenKind::LeftBrace, self.span(start)))
            }
            b'}' => {
                let start = self.pos;
                self.advance();
                Some(Token::new(TokenKind::RightBrace, self.span(start)))
            }
            b'.' => {
                let start = self.pos;
                self.advance();
                Some(Token::new(TokenKind::Dot, self.span(start)))
            }
            b':' => {
                let start = self.pos;
                self.advance();
                if let Some(next) = self.peek()
                    && Self::is_ident_start(next)
                {
                    self.advance();
                    while let Some(c) = self.peek() {
                        if Self::is_ident_continue(c) {
                            self.advance();
                        } else {
                            break;
                        }
                    }
                    let text = &self.source[start + 1..self.pos];
                    return Some(Token::new(
                        TokenKind::Keyword(text.to_string()),
                        self.span(start),
                    ));
                }
                Some(Token::new(TokenKind::Colon, self.span(start)))
            }
            b'"' => Some(self.lex_string()),
            b'-' => {
                if let Some(next) = self.bytes.get(self.pos + 1)
                    && next.is_ascii_digit()
                {
                    let start = self.pos;
                    self.advance();
                    return Some(self.lex_number(start));
                }
                Some(self.lex_operator_symbol())
            }
            b if b.is_ascii_digit() => {
                let start = self.pos;
                Some(self.lex_number(start))
            }
            b if Self::is_ident_start(b) => Some(self.lex_alpha_symbol()),
            b if Self::is_operator_char(b) => Some(self.lex_operator_symbol()),
            _ => {
                let start = self.pos;
                self.advance();
                self.diagnostics.push(Diagnostic::error(
                    format!("unrecognized character: '{}'", b as char),
                    self.span(start),
                ));
                self.next_token()
            }
        }
    }

    fn lex_all(mut self) -> (Vec<Token>, Vec<Diagnostic>) {
        let mut tokens = Vec::new();
        while let Some(token) = self.next_token() {
            tokens.push(token);
        }
        (tokens, self.diagnostics)
    }
}

pub fn lex(source: &str, file: FileId) -> (Vec<Token>, Vec<Diagnostic>) {
    Lexer::new(source, file).lex_all()
}

#[cfg(test)]
mod tests {
    use super::*;

    fn lex_test(source: &str) -> (Vec<Token>, Vec<Diagnostic>) {
        lex(source, FileId::new(0))
    }

    fn kinds(source: &str) -> Vec<TokenKind> {
        lex_test(source).0.into_iter().map(|t| t.kind).collect()
    }

    #[test]
    fn delimiters() {
        assert_eq!(
            kinds("()[]{}"),
            vec![
                TokenKind::LeftParen,
                TokenKind::RightParen,
                TokenKind::LeftBracket,
                TokenKind::RightBracket,
                TokenKind::LeftBrace,
                TokenKind::RightBrace,
            ]
        );
    }

    #[test]
    fn dot_and_colon() {
        assert_eq!(kinds("."), vec![TokenKind::Dot]);
        assert_eq!(kinds(": "), vec![TokenKind::Colon]);
    }

    #[test]
    fn integers() {
        assert_eq!(kinds("42"), vec![TokenKind::Integer(42)]);
        assert_eq!(kinds("-7"), vec![TokenKind::Integer(-7)]);
        assert_eq!(kinds("0"), vec![TokenKind::Integer(0)]);
    }

    #[test]
    fn floats() {
        assert_eq!(kinds("3.14"), vec![TokenKind::Float(3.14)]);
        assert_eq!(kinds("-0.5"), vec![TokenKind::Float(-0.5)]);
        assert_eq!(kinds("0.0"), vec![TokenKind::Float(0.0)]);
    }

    #[test]
    fn strings() {
        assert_eq!(kinds(r#""hello""#), vec![TokenKind::String("hello".into())]);
        assert_eq!(kinds(r#""a\nb""#), vec![TokenKind::String("a\nb".into())]);
        assert_eq!(
            kinds(r#""tab\there""#),
            vec![TokenKind::String("tab\there".into())]
        );
    }

    #[test]
    fn string_escape_sequences() {
        assert_eq!(kinds(r#""\\""#), vec![TokenKind::String("\\".into())]);
        assert_eq!(kinds(r#""\"""#), vec![TokenKind::String("\"".into())]);
        assert_eq!(kinds(r#""\r\0""#), vec![TokenKind::String("\r\0".into())]);
    }

    #[test]
    fn string_unicode_escape() {
        assert_eq!(kinds(r#""\u0041""#), vec![TokenKind::String("A".into())]);
    }

    #[test]
    fn unterminated_string() {
        let (tokens, diags) = lex_test(r#""oops"#);
        assert_eq!(tokens.len(), 1);
        assert_eq!(diags.len(), 1);
        assert!(diags[0].message.contains("unterminated"));
    }

    #[test]
    fn invalid_escape() {
        let (_, diags) = lex_test(r#""\q""#);
        assert_eq!(diags.len(), 1);
        assert!(diags[0].message.contains("invalid escape"));
    }

    #[test]
    fn booleans_and_nil() {
        assert_eq!(kinds("true"), vec![TokenKind::Boolean(true)]);
        assert_eq!(kinds("false"), vec![TokenKind::Boolean(false)]);
        assert_eq!(kinds("nil"), vec![TokenKind::Nil]);
    }

    #[test]
    fn symbols_alphabetic() {
        assert_eq!(kinds("foo"), vec![TokenKind::Symbol("foo".into())]);
        assert_eq!(kinds("defn"), vec![TokenKind::Symbol("defn".into())]);
        assert_eq!(
            kinds("handle-tool-call"),
            vec![TokenKind::Symbol("handle-tool-call".into())]
        );
        assert_eq!(kinds("empty?"), vec![TokenKind::Symbol("empty?".into())]);
        assert_eq!(kinds("set!"), vec![TokenKind::Symbol("set!".into())]);
        assert_eq!(kinds("_unused"), vec![TokenKind::Symbol("_unused".into())]);
    }

    #[test]
    fn symbols_operator() {
        assert_eq!(kinds("+"), vec![TokenKind::Symbol("+".into())]);
        assert_eq!(kinds(">="), vec![TokenKind::Symbol(">=".into())]);
        assert_eq!(kinds("!="), vec![TokenKind::Symbol("!=".into())]);
        assert_eq!(kinds("&&"), vec![TokenKind::Symbol("&&".into())]);
    }

    #[test]
    fn keywords() {
        assert_eq!(kinds(":name"), vec![TokenKind::Keyword("name".into())]);
        assert_eq!(kinds(":else"), vec![TokenKind::Keyword("else".into())]);
    }

    #[test]
    fn colon_vs_keyword() {
        assert_eq!(
            kinds("[x : Int]"),
            vec![
                TokenKind::LeftBracket,
                TokenKind::Symbol("x".into()),
                TokenKind::Colon,
                TokenKind::Symbol("Int".into()),
                TokenKind::RightBracket,
            ]
        );
    }

    #[test]
    fn negative_number_vs_minus_operator() {
        assert_eq!(kinds("-42"), vec![TokenKind::Integer(-42)]);
        assert_eq!(
            kinds("(- 42)"),
            vec![
                TokenKind::LeftParen,
                TokenKind::Symbol("-".into()),
                TokenKind::Integer(42),
                TokenKind::RightParen,
            ]
        );
    }

    #[test]
    fn whitespace_and_commas() {
        assert_eq!(kinds("[a : Int, b : Int]"), kinds("[a : Int  b : Int]"),);
    }

    #[test]
    fn comments_skipped() {
        assert_eq!(kinds("; comment\n42"), vec![TokenKind::Integer(42)]);
        assert_eq!(kinds("42 ; trailing"), vec![TokenKind::Integer(42)]);
    }

    #[test]
    fn qualified_identifier() {
        assert_eq!(
            kinds("vex.http"),
            vec![
                TokenKind::Symbol("vex".into()),
                TokenKind::Dot,
                TokenKind::Symbol("http".into()),
            ]
        );
    }

    #[test]
    fn hello_world_program() {
        let source = r#"(defn main [] (println "Hello, World!"))"#;
        assert_eq!(
            kinds(source),
            vec![
                TokenKind::LeftParen,
                TokenKind::Symbol("defn".into()),
                TokenKind::Symbol("main".into()),
                TokenKind::LeftBracket,
                TokenKind::RightBracket,
                TokenKind::LeftParen,
                TokenKind::Symbol("println".into()),
                TokenKind::String("Hello, World!".into()),
                TokenKind::RightParen,
                TokenKind::RightParen,
            ]
        );
    }

    #[test]
    fn spans_are_correct() {
        let source = "(+ 1 2)";
        let (tokens, _) = lex_test(source);
        assert_eq!(tokens[0].span, Span::new(FileId::new(0), 0, 1));
        assert_eq!(tokens[1].span, Span::new(FileId::new(0), 1, 2));
        assert_eq!(tokens[2].span, Span::new(FileId::new(0), 3, 4));
        assert_eq!(tokens[3].span, Span::new(FileId::new(0), 5, 6));
        assert_eq!(tokens[4].span, Span::new(FileId::new(0), 6, 7));
    }

    #[test]
    fn unrecognized_character() {
        let (_, diags) = lex_test("@");
        assert_eq!(diags.len(), 1);
        assert!(diags[0].message.contains("unrecognized character"));
    }

    #[test]
    fn empty_input() {
        let (tokens, diags) = lex_test("");
        assert!(tokens.is_empty());
        assert!(diags.is_empty());
    }

    #[test]
    fn float_without_decimal_digits_is_int_then_dot() {
        assert_eq!(
            kinds("42.foo"),
            vec![
                TokenKind::Integer(42),
                TokenKind::Dot,
                TokenKind::Symbol("foo".into()),
            ]
        );
    }

    #[test]
    fn map_literal() {
        assert_eq!(
            kinds("{:a 1}"),
            vec![
                TokenKind::LeftBrace,
                TokenKind::Keyword("a".into()),
                TokenKind::Integer(1),
                TokenKind::RightBrace,
            ]
        );
    }
}
