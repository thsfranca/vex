# Vex

[![CI](https://github.com/thsfranca/vex/actions/workflows/ci.yml/badge.svg)](https://github.com/thsfranca/vex/actions/workflows/ci.yml)

A statically typed Lisp for building [MCP](https://modelcontextprotocol.io/) servers. The compiler is written in Rust and transpiles to Go source code, which `go build` compiles to a native binary.

**This is a study project.**

## Quick Look

```lisp
(defn main []
  (println "Hello, World!"))
```

```lisp
(defn fib [n : Int] : Int
  (if (<= n 1)
    n
    (+ (fib (- n 1)) (fib (- n 2)))))

(defn main []
  (println (str "fib(10) = " (fib 10))))
```

```lisp
(defmacro unless [test body]
  (list (quote if) test (quote nil) body))

(defn main []
  (unless (> 1 10)
    (println "1 is not greater than 10")))
```

More examples in [`examples/`](examples/).

## Features

- **Type system** — `Int`, `Float`, `String`, `Bool`, `Option`, `Result`, records (`deftype`), unions (`defunion`)
- **Functions** — `defn`, `def`, `fn` (lambdas), higher-order functions, closures
- **Control flow** — `if`, `cond`, `match` (pattern matching), `and`, `or`, `let`
- **Macros** — `defmacro` with `quote`/`unquote`/`splice`, automatic hygiene, self-hosted prelude
- **Collections** — `List`, `Map`, `each`, `range`, `map`, `filter`
- **Modules** — `module`, `export`, `import`, Go interop (`import-go`)
- **Concurrency** — `spawn`, `channel`, `send`, `recv`, `select`
- **REPL** — interactive `vex repl` with multi-line input and persistent state

## Getting Started

### Prerequisites

- [Rust](https://www.rust-lang.org/tools/install) (stable)
- [Go](https://go.dev/dl/) (1.21+)

### Build the compiler

```bash
cargo build --release
```

### Compile and run a program

```bash
vex build hello.vx              # Compile to ./hello binary
vex build hello.vx -o server    # Custom output name
vex run hello.vx                # Build and run immediately
vex repl                        # Interactive REPL
vex build hello.vx --emit-go .  # Write generated Go source for inspection
```

## Architecture

```
Source → Lexer → Parser → Macro Expand → Type Checker → Codegen → go build → Binary
                                                      → Interpreter (REPL)
```

Each compiler phase is a pure function: data in, data out, plus diagnostics. No global state, no singletons.

| Phase | File | Description |
|-------|------|-------------|
| Source tracking | `source.rs` | `FileId`, `Span`, `SourceMap` |
| Diagnostics | `diagnostics.rs` | Error/warning accumulation and formatting |
| Lexer | `lexer.rs` | Source text → token stream |
| AST | `ast.rs` | Untyped syntax tree types |
| Parser | `parser.rs` | Recursive descent, tokens → AST |
| Macro expansion | `macro_expand.rs` | AST → AST, prelude + `defmacro` with hygiene |
| Type system | `types.rs`, `hir.rs`, `builtins.rs` | Semantic types, typed HIR, built-in registry |
| Type checker | `typechecker.rs` | AST → HIR with type inference |
| Code generation | `codegen.rs` | HIR → Go source |
| Interpreter | `interpreter.rs` | HIR → Value (tree-walking, for REPL) |
| Pipeline | `lib.rs`, `main.rs` | Full compiler pipeline and CLI |

## Current Status

All compiler phases are implemented and pass 551 tests. The pipeline compiles Vex source to working Go binaries end-to-end. The self-hosted macro system (`defmacro`) supports user-defined compile-time macros with automatic hygiene.

See [`docs/roadmap.md`](docs/roadmap.md) for planned features: parametric polymorphism, error propagation macros, exhaustiveness checking, structured concurrency, formatter, LSP, and the MCP framework.

## Documentation

| Document | Contents |
|----------|----------|
| [`language-design.md`](docs/language-design.md) | Syntax, type system, grammar, backend strategy, design decisions |
| [`compiler-architecture.md`](docs/compiler-architecture.md) | Pipeline, file structure, phase contracts, testing strategy |
| [`dependency-management.md`](docs/dependency-management.md) | `vex.mod` manifest, `vex get`, global cache, Go module integration |
| [`roadmap.md`](docs/roadmap.md) | Feature roadmap with priorities and status |
| [`roadmap-rationale.md`](docs/roadmap-rationale.md) | Design analysis and trade-offs behind roadmap items |
| [`installation.md`](docs/installation.md) | Installation process, release pipeline, distribution channels |
| [`mvp.md`](docs/mvp.md) | MVP definition and success criteria |

## Development

```bash
cargo build           # Build the compiler
cargo test            # Run all 551 tests
cargo clippy          # Lint
cargo fmt             # Format
```

## License

MIT
