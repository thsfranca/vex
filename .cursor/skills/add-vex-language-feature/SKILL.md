---
name: add-vex-language-feature
description: >-
  Guides through adding a new syntax construct, keyword, type, or built-in to the Vex language.
  Use when adding tokens, AST nodes, parser rules, type system features, or built-in functions.
---

# Add Vex Language Feature

Adding a language feature touches multiple files in a strict order. Follow every step — skipping one causes downstream breakage.

## Checklist

Copy and track progress:

```
- [ ] 1. Update design doc (docs/language-design.md)
- [ ] 2. Lexer (src/lexer.rs) — new token kinds if needed
- [ ] 3. AST (src/ast.rs) — new untyped node variants
- [ ] 4. Parser (src/parser.rs) — parsing logic for new syntax
- [ ] 5. Types (src/types.rs) — new VexType variants if needed
- [ ] 6. HIR (src/hir.rs) — typed mirror of new AST nodes
- [ ] 7. Builtins (src/builtins.rs) — if adding a built-in function
- [ ] 8. Type checker (src/typechecker.rs) — AST→HIR lowering + validation
- [ ] 9. Codegen (src/codegen.rs) — HIR→Go emission
- [ ] 10. Interpreter (src/interpreter.rs) — HIR→Value evaluation (if REPL exists)
- [ ] 11. Tests at each layer
```

## Step Details

### 1. Design Doc

Add the feature to `docs/language-design.md`:
- Syntax example in §6 or §13
- Grammar rule in §7 (EBNF)
- Type system implications in §5 if applicable

### 2. Lexer — `src/lexer.rs`

Only needed if the feature introduces new token kinds (e.g., a new delimiter or literal type).

- Add variant to `TokenKind` enum
- Add lexing logic
- Add disambiguation rules if ambiguous (see §7.1 of design doc)
- Test: source string → assert token kinds and values

### 3. AST — `src/ast.rs`

Add untyped AST node variants. Every node must carry a `Span`.

- New `Expr` variants for expressions
- New `TopForm` variants for top-level declarations
- New `Pattern` variants for pattern matching extensions

### 4. Parser — `src/parser.rs`

Add parsing rules. The parser is hand-written recursive descent.

- Dispatch on `Symbol` token value for special forms (e.g., `"for"`, `"try"`)
- Produce the new AST node variants from step 3
- Test: token vec → assert AST structure via pattern matching

### 5. Types — `src/types.rs`

Only needed if the feature introduces new semantic types.

- Add variant to `VexType` enum
- Update `TypeEnv` if the type needs special handling

### 6. HIR — `src/hir.rs`

Mirror the new AST nodes but with resolved `VexType` on every expression.

- Structurally mirrors `ast.rs` additions
- Depends on `source` and `types`

### 7. Builtins — `src/builtins.rs`

Only needed when adding a built-in function (not syntax).

- Single record: name, Vex type signature (`VexType`), Go translation string
- Both type checker and codegen read from this file — one source of truth

### 8. Type Checker — `src/typechecker.rs`

Transform untyped AST → typed HIR for the new construct.

- Match on the new `ast::Expr` / `ast::TopForm` variant
- Produce the corresponding `hir::Expr` / `hir::TopForm`
- Emit `Diagnostic` for type errors
- Test: hand-built AST → assert HIR types or expected diagnostics

### 9. Codegen — `src/codegen.rs`

Emit Go source from the new HIR node.

- Match on the new `hir::Expr` / `hir::TopForm` variant
- Use the Vex-to-Go type mapping from `docs/compiler-architecture.md` §10
- Naming convention: kebab-case → PascalCase
- Test: hand-built HIR → assert generated Go contains expected patterns

### 10. Interpreter — `src/interpreter.rs`

Evaluate the new HIR node for the REPL (if interpreter is implemented).

- Match on the new `hir::Expr` variant
- Return `Value` or `RuntimeError`

### 11. Tests

Each layer gets its own tests in `#[cfg(test)] mod tests`:

| Layer | Input | Assert |
|-------|-------|--------|
| Lexer | source string | token kinds, values, spans |
| Parser | token vec | AST structure |
| Type checker | hand-built AST | HIR types or diagnostics |
| Codegen | hand-built HIR | Go source patterns |
| Integration | `.vx` source string | full pipeline Go output |

## Common Mistakes

- Adding an AST node without a `Span` — everything needs spans for error reporting
- Adding a built-in in `typechecker.rs` instead of `builtins.rs` — always use the single registry
- Forgetting the HIR mirror — codegen only sees HIR, never AST
- Not testing error cases — type checker diagnostics are as important as the happy path
