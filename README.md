# Vex

[![CI](https://github.com/thsfranca/vex/actions/workflows/ci.yml/badge.svg)](https://github.com/thsfranca/vex/actions/workflows/ci.yml)

A statically typed Lisp for building MCP servers. The compiler is written in Rust and transpiles to Go source code.

**This is a study project.**

## Overview

Vex uses S-expression syntax with compile-time type checking, targeting networked services — especially [MCP](https://modelcontextprotocol.io/) servers. The compiler pipeline (lexer → parser → type checker → codegen) produces Go source that `go build` compiles to a native binary.

## Example

```
(defn main []
  (println "Hello, World!"))
```

## Current Status

Collections are implemented — `(List T)` and `(Map K V)` types with `each`, `range`, `map`, `filter` built-ins.

| Phase | Status |
|-------|--------|
| `source.rs` — FileId, Span, SourceMap | Done |
| `diagnostics.rs` — Diagnostic, Severity, formatting | Done |
| `lexer.rs` — Tokenizer | Done |
| `ast.rs` — Untyped AST types | Done |
| `parser.rs` — Recursive descent parser | Done |
| `types.rs` / `hir.rs` / `builtins.rs` — Type system | Done |
| `typechecker.rs` — AST → HIR | Done |
| `codegen.rs` — HIR → Go source | Done |
| `lib.rs` / `main.rs` — Full pipeline, CLI | Done |

**Implemented features:** integers, floats, strings, booleans, nil, `defn`, `def`, `if`, `cond`, `let`, `fn` (lambdas), higher-order functions, records (`deftype`), field access (`.`), record constructors, unions (`defunion`), pattern matching (`match`), `Option`/`Result` types, collections (`List`, `Map`), `each`, `range`, `map`, `filter`, arithmetic/comparison/logical operators, `println`, `str`, `mod`.

## Architecture

- Pipeline: Source → Lexer → Parser → Type Checker → Codegen → `go build` → Binary
- Design docs: [`docs/language-design.md`](docs/language-design.md), [`docs/compiler-architecture.md`](docs/compiler-architecture.md)

## Usage

```
cargo build
vex build hello.vx              # Compile to ./hello binary
vex build hello.vx -o server    # Custom output name
vex build hello.vx --emit-go .  # Also write generated Go source
vex run hello.vx                # Build and run immediately
```

## Building

```
cargo build
cargo test
```

## License

MIT
