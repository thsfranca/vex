# Vex

[![CI](https://github.com/thsfranca/vex/actions/workflows/ci.yml/badge.svg)](https://github.com/thsfranca/vex/actions/workflows/ci.yml)

A statically typed Lisp for building MCP servers. The compiler is written in Rust and transpiles to Go source code.

**This is a study project.**

## Overview

Vex combines S-expression syntax with compile-time type checking, targeting networked services — especially [MCP](https://modelcontextprotocol.io/) servers. The compiler pipeline produces Go source that `go build` compiles to a native binary. An alternative path runs code directly through a tree-walking interpreter for the REPL.

## Examples

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
(deftype Point (x Float) (y Float))

(defn distance [p : Point] : String
  (str "(" (. p x) ", " (. p y) ")"))

(defn main []
  (let [origin (Point 0.0 0.0)
        target (Point 3.0 4.0)]
    (println (str "origin = " (distance origin)))
    (println (str "target = " (distance target)))))
```

## Usage

```bash
vex build hello.vx              # Compile to ./hello binary
vex build hello.vx -o server    # Custom output name
vex build hello.vx --emit-go .  # Also write generated Go source
vex run hello.vx                # Build and run immediately
vex repl                        # Interactive REPL
```

## Current Status

All compiler phases are implemented and pass 551 tests. The self-hosted macro system (`defmacro`) supports user-defined compile-time macros with automatic hygiene.

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

### Language Features

- **Primitives:** integers, floats, strings, booleans, nil
- **Functions:** `defn`, `def`, `fn` (lambdas), higher-order functions
- **Control flow:** `if`, `cond`, `and`, `or`, `let`, pattern matching (`match`)
- **Macros:** `defmacro`, `quote`, `unquote`, `splice`, macro helpers, automatic hygiene via `gensym`
- **Data types:** records (`deftype`), field access (`.`), unions (`defunion`)
- **Built-in types:** `Option`, `Result`, `List`, `Map`
- **Collections:** `each`, `range`, `map`, `filter`
- **Modules:** `module`, `export`, `import`, Go interop (`import-go`)
- **Concurrency:** `spawn`, `channel`, `send`, `recv`
- **REPL:** interactive `vex repl` with multi-line input and persistent state
- **Built-in functions:** `println`, `str`, `mod`, arithmetic and comparison operators

## Architecture

```
Source → Lexer → Parser → Macro Expand → Type Checker → Codegen → go build → Binary
                                                      → Interpreter (REPL)
```

## Documentation

- [`docs/language-design.md`](docs/language-design.md) — syntax, type system, grammar, backend strategy, design decisions
- [`docs/compiler-architecture.md`](docs/compiler-architecture.md) — pipeline, file structure, phase contracts, testing strategy
- [`docs/dependency-management.md`](docs/dependency-management.md) — `vex.mod` manifest, `vex get`, global cache, Go module integration
- [`docs/mvp.md`](docs/mvp.md) — MVP definition and success criteria

## Development

```bash
cargo build
cargo test
```

## License

MIT
