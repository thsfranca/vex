# Vex

[![CI](https://github.com/thsfranca/vex/actions/workflows/ci.yml/badge.svg)](https://github.com/thsfranca/vex/actions/workflows/ci.yml)

A statically typed Lisp for building MCP servers. The compiler is written in Rust and transpiles to Go source code.

**This is a study project.**

## Overview

Vex uses S-expression syntax with compile-time type checking, targeting networked services — especially [MCP](https://modelcontextprotocol.io/) servers. The compiler pipeline (lexer → parser → macro expand → type checker → codegen) produces Go source that `go build` compiles to a native binary.

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
vex repl                        # Interactive REPL
```

## Current Status

All compiler phases are implemented. The self-hosted macro system (`defmacro`) supports user-defined compile-time macros with automatic hygiene.

### Compiler Phases

| Phase | Status |
|-------|--------|
| `source.rs` — FileId, Span, SourceMap | Done |
| `diagnostics.rs` — Diagnostic, Severity, formatting | Done |
| `lexer.rs` — Tokenizer | Done |
| `ast.rs` — Untyped AST types | Done |
| `parser.rs` — Recursive descent parser | Done |
| `types.rs` / `hir.rs` / `builtins.rs` — Type system | Done |
| `macro_expand.rs` — Compiler macros + self-hosted `defmacro` with hygiene | Done |
| `typechecker.rs` — AST → HIR | Done |
| `codegen.rs` — HIR → Go source | Done |
| `interpreter.rs` — HIR → Value (tree-walking eval) | Done |
| `lib.rs` / `main.rs` — Full pipeline, CLI | Done |

### Implemented Features

- **Primitives:** integers, floats, strings, booleans, nil
- **Functions:** `defn`, `def`, `fn` (lambdas), higher-order functions
- **Control flow:** `if`, `cond`, `and`, `or`, `let`, pattern matching (`match`)
- **Macros:** `defmacro`, `quote`, `unquote`, `splice`, macro helpers (`syntax-list`, `syntax-cons`, etc.), automatic hygiene via `gensym`
- **Data types:** records (`deftype`), field access (`.`), record constructors, unions (`defunion`)
- **Built-in types:** `Option`, `Result`, `List`, `Map`
- **Collections:** `each`, `range`, `map`, `filter`
- **Modules:** `module`, `export`, `import`, Go interop (`import-go`)
- **Concurrency:** `spawn`, `channel`, `send`, `recv`
- **REPL:** interactive `vex repl` with multi-line input and persistent state
- **Interpreter:** tree-walking HIR evaluator with all expression forms
- **Built-in functions:** `println`, `str`, `mod`, arithmetic/comparison operators

## Architecture

- Pipeline: Source → Lexer → Parser → Macro Expand → Type Checker → Codegen → `go build` → Binary
- Alternative: Source → Lexer → Parser → Macro Expand → Type Checker → Interpreter (direct eval)
- Design docs: [`docs/language-design.md`](docs/language-design.md), [`docs/compiler-architecture.md`](docs/compiler-architecture.md)

## Development

```bash
cargo build
cargo test
```

## License

MIT
