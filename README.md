# Vex

[![CI](https://github.com/thsfranca/vex/actions/workflows/ci.yml/badge.svg)](https://github.com/thsfranca/vex/actions/workflows/ci.yml)

A statically typed Lisp for building MCP servers. The compiler is written in Rust and transpiles to Go source code.

**This is a study project.**

## Overview

Vex uses S-expression syntax with compile-time type checking, targeting networked services — especially [MCP](https://modelcontextprotocol.io/) servers. The compiler pipeline (lexer → parser → type checker → codegen) produces Go source that `go build` compiles to a native binary.

## Example

```lisp
(defn main []
  (println "Hello, World!"))
```

## Usage

```bash
cargo build
vex build hello.vx              # Compile to ./hello binary
vex build hello.vx -o server    # Custom output name
vex build hello.vx --emit-go .  # Also write generated Go source
vex run hello.vx                # Build and run immediately
```

## Current Status

Tree-walking interpreter is implemented — evaluates typed HIR directly for instant feedback without the Go compile cycle.

### Compiler Phases

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
| `interpreter.rs` — HIR → Value (tree-walking eval) | Done |
| `lib.rs` / `main.rs` — Full pipeline, CLI | Done |

### Implemented Features

- **Primitives:** integers, floats, strings, booleans, nil
- **Functions:** `defn`, `def`, `fn` (lambdas), higher-order functions
- **Control flow:** `if`, `cond`, `let`, pattern matching (`match`)
- **Data types:** records (`deftype`), field access (`.`), record constructors, unions (`defunion`)
- **Built-in types:** `Option`, `Result`, `List`, `Map`
- **Collections:** `each`, `range`, `map`, `filter`
- **Modules:** `module`, `export`, `import`, Go interop (`import-go`)
- **Concurrency:** `spawn`, `channel`, `send`, `recv`
- **Interpreter:** tree-walking HIR evaluator with all expression forms
- **Operators:** arithmetic, comparison, logical
- **Built-in functions:** `println`, `str`, `mod`

## Architecture

- Pipeline: Source → Lexer → Parser → Type Checker → Codegen → `go build` → Binary
- Alternative: Source → Lexer → Parser → Type Checker → Interpreter (direct eval)
- Design docs: [`docs/language-design.md`](docs/language-design.md), [`docs/compiler-architecture.md`](docs/compiler-architecture.md)

## Development

```bash
cargo build
cargo test
```

## License

MIT
