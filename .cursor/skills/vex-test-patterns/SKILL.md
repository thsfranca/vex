---
name: vex-test-patterns
description: >-
  Provides testing patterns for the Vex compiler. Use when writing unit tests,
  integration tests, or test helpers for any compiler phase.
---

# Vex Test Patterns

Each compiler phase is tested independently by constructing its input and asserting its output.

## Test Location

| Test type | Location |
|-----------|----------|
| Unit tests | `#[cfg(test)] mod tests` at the bottom of each `src/*.rs` file |
| Integration tests | `tests/` directory at the crate root |

## Per-Phase Testing

### Lexer (`src/lexer.rs`)

Input: source string + `FileId`
Assert: token kinds, token values, spans

```rust
#[test]
fn lex_hello_world() {
    let source = r#"(println "Hello")"#;
    let file = FileId::new(0);
    let (tokens, diagnostics) = lex(source, file);
    assert!(diagnostics.is_empty());
    assert_eq!(tokens[0].kind, TokenKind::LeftParen);
    assert_eq!(tokens[1].kind, TokenKind::Symbol);
    assert_eq!(tokens[1].text, "println");
    // ...
}
```

Error case — assert diagnostics:

```rust
#[test]
fn lex_unterminated_string() {
    let source = r#"(println "oops)"#;
    let file = FileId::new(0);
    let (_, diagnostics) = lex(source, file);
    assert_eq!(diagnostics.len(), 1);
    assert_eq!(diagnostics[0].severity, Severity::Error);
}
```

### Parser (`src/parser.rs`)

Input: token vec (from lexer or hand-built)
Assert: AST structure via pattern matching

```rust
#[test]
fn parse_defn() {
    let source = r#"(defn main [] (println "hi"))"#;
    let tokens = lex(source, FileId::new(0)).0;
    let (forms, diagnostics) = parse(&tokens);
    assert!(diagnostics.is_empty());
    assert!(matches!(&forms[0], ast::TopForm::Defn { .. }));
}
```

### Type Checker (`src/typechecker.rs`)

Input: hand-built `ast::TopForm` or AST from parser
Assert: HIR types match expected `VexType`, or expected diagnostics for errors

```rust
#[test]
fn check_type_mismatch() {
    // Build AST that passes Int where String expected
    let ast = /* ... */;
    let (_, diagnostics) = check(&[ast]);
    assert_eq!(diagnostics.len(), 1);
    // Assert diagnostic message mentions the mismatch
}
```

### Codegen (`src/codegen.rs`)

Input: hand-built `hir::Module`
Assert: generated Go source string contains expected patterns

```rust
#[test]
fn generate_hello_world() {
    let module = /* hand-built HIR for hello world */;
    let go_source = generate(&module);
    assert!(go_source.contains("fmt.Println"));
    assert!(go_source.contains("func main()"));
}
```

### Integration (`tests/`)

Full pipeline: `.vx` source string → compile → assert Go output

```rust
#[test]
fn compile_hello_world() {
    let source = r#"(defn main [] (println "Hello, World!"))"#;
    let result = vex::compile(source);
    assert!(result.diagnostics.is_empty());
    assert!(result.go_source.contains("func main()"));
    assert!(result.go_source.contains(`fmt.Println("Hello, World!")`));
}
```

## Guidelines

- Always read the implementation before writing tests — don't guess at API shapes
- Test both happy path and error cases at every layer
- For diagnostics: assert count, severity, and that the message contains key terms
- Prefer pattern matching (`matches!()`) over field-by-field assertions for AST/HIR structure
- Integration tests should cover the full pipeline without intermediate assertions
- Every AST/HIR node carries a `Span` — use `Span::dummy()` or similar in hand-built test data if available, otherwise construct minimal valid spans
