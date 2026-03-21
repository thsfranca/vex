# Vex Compiler Architecture

**Version:** 0.1.0-draft
**Date:** 2026-03-21

---

## Table of Contents

1. [Pipeline Overview](#1-pipeline-overview)
2. [File Structure](#2-file-structure)
3. [Dependency Graph](#3-dependency-graph)
4. [Shared Foundation](#4-shared-foundation)
5. [Phase Contracts](#5-phase-contracts)
6. [Two AST Representations](#6-two-ast-representations)
7. [Built-in Functions](#7-built-in-functions)
8. [Go Runtime Library](#8-go-runtime-library)
9. [Go Output Model](#9-go-output-model)
10. [Vex-to-Go Type Mapping](#10-vex-to-go-type-mapping)
11. [Testing Strategy](#11-testing-strategy)
12. [Implementation Order](#12-implementation-order)
13. [Developer Experience](#13-developer-experience)
14. [Intentionally Excluded from v0.1](#14-intentionally-excluded-from-v01)

---

## 1. Pipeline Overview

The compiler is a linear sequence of pure transformations:

- Each phase is a function that takes data in and produces data out, plus diagnostics
- No global state, no trait-based abstractions, no framework indirection

```
Source (.vx)
    │
    ▼
┌─────────┐
│  Lexer  │    &str, FileId → Vec<Token>
└────┬────┘
     │
     ▼
┌─────────┐
│ Parser  │    &[Token] → Vec<ast::TopForm>
└────┬────┘
     │
     ▼
┌─────────────────┐
│ Macro Expansion │    Vec<ast::TopForm> → Vec<ast::TopForm>
└────────┬────────┘
         │
         ▼
┌──────────────┐
│ Type Checker │    &[ast::TopForm] → hir::Module
└──────┬───────┘
       │
       ├──────────────────┐
       ▼                  ▼
┌──────────┐       ┌─────────────┐
│ Codegen  │       │ Interpreter │
│          │       │   (REPL)    │
└────┬─────┘       └──────┬──────┘
     │ .go source         │ Value
     ▼                    ▼
┌──────────┐         Interactive
│ go build │         session
└────┬─────┘
     │
     ▼
   Binary
```

Macro expansion sits between Parser and Type Checker:

- Core macros (`cond`, `and`, `or`) are self-hosted — defined with `defmacro` in a prelude module embedded in the compiler binary (see `language-design.md` §14, Decisions 8–9)
- All macros (prelude and user-defined) execute at compile time via a dedicated AST evaluator (see `language-design.md` §4.5):
  - The prelude source is expanded first, populating the macro registry with `cond`, `and`, `or`
  - Pass 1: collect `defmacro` forms from user code, store body AST + parameter names in the macro registry
  - Pass 2: walk the AST — on macro call, convert arguments to `Syntax` values, evaluate the macro body via the AST evaluator, apply hygiene, convert the result back to AST, re-expand
- Macro helper functions (`list`, `cons`, `first`, `rest`, `symbol?`, `list?`, `concat`) exist only in the AST evaluator — they are not global builtins
- Macros operate on `ast::TopForm` (untyped S-expression-derived trees) and produce `ast::TopForm`
- The type checker validates the fully expanded result
- `defmacro` forms are erased — they do not appear in the HIR or generated Go output

---

## 2. File Structure

Flat files. No directories for modules until a file exceeds ~500 lines. Each file owns one concept.

```
src/
  main.rs           CLI: argument parsing, orchestration, go build invocation
  lib.rs            Library root: re-exports all modules, exposes compile() pipeline function

  source.rs         FileId, SourceMap, Span — the foundation everything else depends on
  diagnostics.rs    Diagnostic, Severity, Label — error/warning accumulation and formatting

  lexer.rs          Lexer struct, TokenKind enum, Token struct, lex() function
  ast.rs            Untyped AST: Expr, TopForm, Pattern, TypeExpr, Param, Field, etc.
  parser.rs         Parser struct, parse() function (tokens → AST)
  macro_expand.rs   expand() function (AST → AST), prelude loading, user-defined macro execution via AST evaluator

  hir.rs            Typed AST: mirrors ast.rs but every node has a resolved type
  types.rs          VexType enum (semantic types: Int, Float, Function, etc.), TypeEnv
  builtins.rs       Built-in function registry: names, type signatures, Go translations
  typechecker.rs    check() function (AST → HIR), type inference, unification

  codegen.rs        generate() function (HIR → Go source string)
  interpreter.rs    eval() function (HIR → Value), for REPL

lib/
  prelude.vx        Self-hosted macros (cond, and, or) — embedded into the binary via include_str!
```

14 Rust files in `src/`, each with a single responsibility. Vex library source lives in `lib/`, separate from compiler implementation.

The key split is `ast.rs` vs `hir.rs`, mirroring the Lexer→Parser boundary:

- **Tokens** — what characters are there
- **AST** (`ast.rs`) — what structure is there (syntax — what the programmer wrote)
- **HIR** (`hir.rs`) — what types are there (semantics — what it means after type resolution)

---

## 3. Dependency Graph

Strict layering, no cycles. `source.rs` at the bottom, `main.rs` at the top.

```
                    main.rs
                      │
                    lib.rs
                   ╱  │  ╲
                 ╱    │    ╲
               ╱      │      ╲
         codegen.rs   │   interpreter.rs
              │       │       │
              └───┬───┘───────┘
                  │
            typechecker.rs
            ╱    │    ╲
          ╱      │      ╲
        ╱        │        ╲
    ast.rs    hir.rs    types.rs
       │        │  ╲      │
       │        │   ╲     │
       │        │  types.rs
       │        │         │
       └────────┼─────────┘
            ╱   │
macro_expand.rs parser.rs
            ╱       ╲
       lexer.rs    ast.rs
            │        │
            ┴────────┘
            │
     diagnostics.rs
            │
       source.rs
```

Precise per-file dependencies:

| File | Depends on |
|------|-----------|
| `source.rs` | (nothing) |
| `diagnostics.rs` | `source` |
| `lexer.rs` | `source`, `diagnostics` |
| `ast.rs` | `source` |
| `parser.rs` | `source`, `diagnostics`, `lexer`, `ast` |
| `macro_expand.rs` | `source`, `diagnostics`, `ast`, `types` |
| `types.rs` | `source` |
| `hir.rs` | `source`, `types` |
| `builtins.rs` | `types` |
| `typechecker.rs` | `source`, `diagnostics`, `ast`, `hir`, `types`, `builtins` |
| `codegen.rs` | `hir`, `types`, `builtins` |
| `interpreter.rs` | `hir`, `types`, `builtins` |
| `lib.rs` | all modules |
| `main.rs` | `lib` |

---

## 4. Shared Foundation

- Every token, AST node, and HIR node carries a `Span`
- Every error references a `Span`
- These two files are the foundation all other modules depend on, and they depend on nothing else

### `source.rs`

Owns:
- `FileId` — opaque identifier for a source file (a newtype around `u32`)
- `Span` — byte offset range within a file (`file: FileId`, `start: u32`, `end: u32`)
- `SourceMap` — maps `FileId` → file name + source text; provides methods to convert byte offsets into line/column for error display

`FileId` exists from day one even though v0.1 compiles one file at a time — adding it later would require touching every type that holds a `Span`.

### `diagnostics.rs`

Owns:
- `Severity` — `Error` or `Warning`
- `Label` — a `Span` plus a message string, for secondary annotations
- `Diagnostic` — severity, primary message, primary `Span`, and `Vec<Label>` for secondary spans
- Formatting logic to render diagnostics with source snippets, line numbers, and underlines

- All phases that can produce errors receive a `&mut Vec<Diagnostic>` to push into
- The caller (`lib.rs`) decides whether to continue after errors (for error recovery) or stop

---

## 5. Phase Contracts

Each phase is a standalone function with a clear signature.

### Lexer

```
fn lex(source: &str, file: FileId) -> (Vec<Token>, Vec<Diagnostic>)
```

- Always produces a token stream (possibly with error tokens or partial results)
- Diagnostics: unterminated strings, invalid escape sequences, unrecognized characters

### Parser

```
fn parse(tokens: &[Token]) -> (Vec<ast::TopForm>, Vec<Diagnostic>)
```

- Consumes the token slice, produces an untyped AST
- Diagnostics: unexpected tokens, missing delimiters, malformed forms

### Macro Expansion

```
fn expand(program: Vec<ast::TopForm>) -> (Vec<ast::TopForm>, Vec<Diagnostic>)
```

- Loads the prelude (embedded `prelude.vx`) first, expanding its `defmacro` definitions to populate the macro registry with `cond`, `and`, `or`
- Then processes user code:
  - Pass 1: collect `defmacro` forms, store body AST and parameter names in the macro registry (which already contains prelude macros)
  - Pass 2: on macro call (prelude or user-defined), convert arguments to `Syntax` values, evaluate body via AST evaluator, apply hygiene (rename macro-introduced bindings), convert result back to AST, re-expand
- Macro helper functions (`list`, `cons`, `first`, `rest`, `symbol?`, `list?`, `concat`) exist only in the AST evaluator — not as global builtins
- Produces an AST with only primitive forms — no `defmacro` forms or macro calls survive this phase
- Diagnostics: malformed macro invocations, evaluation errors in macro bodies, expansion depth limit exceeded

### Type Checker

```
fn check(program: &[ast::TopForm]) -> (hir::Module, Vec<Diagnostic>)
```

- Transforms untyped AST into typed HIR
- Diagnostics: type mismatches, undefined symbols, arity errors

### Codegen

```
fn generate(module: &hir::Module) -> String
```

- Produces Go source code
- No diagnostics — the HIR is valid by construction after type checking succeeds

### Interpreter

```
fn eval(module: &hir::Module) -> Result<Value, RuntimeError>
```

- Tree-walking evaluation of the typed HIR
- Runtime errors returned as `Result` (division by zero, channel deadlock, etc.)

---

## 6. Two AST Representations

This is the most important architectural decision.

### The Problem

The parser produces syntax trees. The type checker needs to annotate every node with its resolved type.

**Option A (rejected): Single AST + side table**

Parser produces nodes with `NodeId`s. Type checker fills a `HashMap<NodeId, VexType>`. Codegen reads both.

- Codegen must always cross-reference two data structures
- Easy to forget a lookup — no compile-time guarantee that types exist
- "Has this been type-checked?" is a runtime question, not a type-level one

**Option B (chosen): Separate AST and HIR types**

Parser produces `ast::Expr`. Type checker produces `hir::Expr`. Codegen only sees HIR.

- Each phase has an unambiguous input and output type
- Codegen cannot accidentally operate on untyped data
- The "duplication" is small because Lisp ASTs are structurally simple
- Macro expansion operates on AST and produces AST — the boundary stays clean

### `ast.rs` — Untyped AST

Defines the syntax tree exactly as the parser produces it:

- Type annotations from the source (`[x : Int]`) are stored as `Option<TypeExpr>`
- `TypeExpr` is a syntactic representation (just names and nesting), not a resolved semantic type

Key types: `Expr`, `TopForm`, `Pattern`, `TypeExpr`, `Param`, `Field`, `Variant`, `MatchClause`.

Every node carries a `Span`.

### `hir.rs` — Typed HIR

Structurally mirrors `ast.rs`, but:

- Every expression node carries a resolved `VexType` (from `types.rs`)
- All `Option<TypeExpr>` are resolved to concrete `VexType` values
- Macro forms are already expanded before type checking (e.g., `cond` → nested `if` happens during macro expansion)
- No unresolved names — all symbols are resolved to their definitions

Key types: `Expr`, `TopForm`, `Module` (plus whatever mirrors are needed from AST).

### `types.rs` — Semantic Types

Defines `VexType`, the compiler's internal representation of types:

- Primitive types: `Int`, `Float`, `Bool`, `String`, `Char`, `Unit`
- Composite types: `List(Box<VexType>)`, `Map(Box<VexType>, Box<VexType>)`, `Tuple(Vec<VexType>)`, `Option(Box<VexType>)`, `Result(Box<VexType>, Box<VexType>)`
- Function types: `Fn(Vec<VexType>, Box<VexType>)`
- Named types: user-defined records and unions
- Type variables: for inference (unification variables)

Also defines `TypeEnv` for tracking bindings during type checking.

---

## 7. Built-in Functions

Vex provides built-in functions (`println`, `str`, `+`, `-`, `mod`, `range`, etc.) that are available without imports:

- The type checker needs their type signatures to validate calls
- The codegen needs their Go translations to emit correct code
- Both concerns live in one place

### `builtins.rs`

Each built-in is defined as a single record containing its name, Vex type signature, and Go translation. Both `typechecker.rs` and `codegen.rs` depend on this file.

- Adding a built-in is a single change in a single file
- No second place to update, no risk of type checker and codegen disagreeing about what exists
- The type checker calls into `builtins` to populate the initial `TypeEnv` with all built-in names and their types
- The codegen calls into `builtins` to look up the Go code that a built-in name maps to

### Why not split across typechecker and codegen?

- Adding a built-in would require editing two files, with no compile-time guarantee they stay in sync
- A forgotten update in either direction produces a confusing bug:
  - Type checker accepts a call that codegen doesn't know how to emit
  - Codegen encounters a name the type checker should have rejected

### Why not put it in `types.rs`?

- `types.rs` defines Vex's type representations — it's about the type system, not about what functions exist
- Coupling Go translation strings into `types.rs` would mix semantics with backend concerns
- `builtins.rs` is a small, focused file that bridges the two

---

## 8. Go Runtime Library

Vex has types that don't exist in Go — `Result`, `Option`, and `defunion` (sum types). Generated Go code needs helper type definitions to represent them.

### Approach: Embedded Runtime

The Vex compiler binary embeds Go source files for the runtime via Rust's `include_str!`:

- At build time, the compiler writes these files alongside the user's generated code into a temporary Go module
- The user's generated `main.go` imports `vex_out/vexrt` like any Go package

```
/tmp/vex-build-XXXX/
  go.mod              generated, module name "vex_out"
  main.go             generated from user's .vx source
  vexrt/
    result.go         Result[T, E] type definition
    option.go         Option[T] type definition
    union.go          union dispatch helpers
```

### Why not an external Go package?

- Requires network access at build time or pre-installation
- Creates version coupling between the compiler and the runtime — compiler v0.2 might need runtime v0.2, but the user could have v0.1 cached
- For a young language, this is a distribution headache with no upside

### Why not inline everything into the generated file?

- Works for small programs but scales poorly
- As more Vex features are used, the generated file fills with boilerplate that obscures user logic
- Multi-file compilation would duplicate the definitions in every output file

### Growth path

- For the hello world milestone, the runtime is not needed — `println` maps to `fmt.Println` with no helpers
- The `vexrt/` directory is omitted entirely when empty
- As Result, Option, and union support are added, each gets a `.go` file embedded in the compiler binary
- The runtime grows with the language

---

## 9. Go Output Model

When the user runs `vex hello.vx`, the compiler produces a native binary. Intermediate Go artifacts are invisible by default.

### Default behavior

1. Create a temporary directory
2. Write `go.mod` + generated `main.go` + `vexrt/` (if needed) into it
3. Run `go build -o <output_path>` against the temporary module
4. Delete the temporary directory
5. Binary is placed in the current working directory, named after the source file without extension: `hello.vx` → `hello` (or `hello.exe` on Windows)

### CLI interface

```
vex build hello.vx              Build hello.vx → ./hello binary
vex build hello.vx -o server    Build hello.vx → ./server binary
vex build hello.vx --emit-go .  Build and also write generated Go to current directory
vex run hello.vx                Build and immediately execute
```

`--emit-go <dir>` writes the generated Go module to the specified directory instead of deleting it — the escape hatch for inspecting and debugging generated code.

### Exit codes

| Code | Meaning |
|------|---------|
| 0 | Success |
| 1 | Compilation error (lex, parse, type check) |
| 2 | Go build failed (generated code error — compiler bug) |
| 3 | CLI usage error (bad arguments, missing file) |

---

## 10. Vex-to-Go Type Mapping

Every Vex type must have a concrete Go representation. This mapping drives codegen and determines what the `vexrt/` runtime package contains.

### Primitive Types

| Vex | Go | Notes |
|-----|-----|-------|
| `Int` | `int64` | |
| `Float` | `float64` | |
| `Bool` | `bool` | |
| `String` | `string` | |
| `Char` | `rune` | |
| `Unit` | (no return value) | Functions returning `Unit` emit no Go return type. When `Unit` appears as a type parameter (e.g., `Result Unit Error`), it maps to `struct{}`. |

### Collection Types

| Vex | Go | Notes |
|-----|-----|-------|
| `(List T)` | `[]T` | |
| `(Map K V)` | `map[K]V` | |
| `(Tuple T1 T2 ...)` | `vexrt.Tuple2[T1, T2]`, `vexrt.Tuple3[T1, T2, T3]`, etc. | One generic struct per arity in the runtime: `type Tuple2[T1 any, T2 any] struct { V0 T1; V1 T2 }` |

### Function Types

| Vex | Go |
|-----|-----|
| `(Fn [T1 T2] R)` | `func(T1, T2) R` |

### Option and Result

Option and Result are tagged generic structs in `vexrt/`, using a boolean discriminant rather than the interface pattern:

- Both have exactly two variants, so the struct approach is a natural fit
- Value semantics — no heap allocation from interface boxing
- Preserves Go's generic type safety

```go
// vexrt/option.go
type Option[T any] struct {
    IsSome bool
    Value  T
}

// vexrt/result.go
type Result[T any, E any] struct {
    IsOk  bool
    Value T
    Error E
}
```

Constructors:

```go
func Some[T any](v T) Option[T]            { return Option[T]{IsSome: true, Value: v} }
func None[T any]() Option[T]               { return Option[T]{} }
func Ok[T any, E any](v T) Result[T, E]    { return Result[T, E]{IsOk: true, Value: v} }
func Err[T any, E any](e E) Result[T, E]   { return Result[T, E]{Error: e} }
```

### User-Defined Records (`deftype`)

Records map directly to Go structs. Fields are exported (PascalCase) for access and JSON marshaling.

Vex:
```
(deftype ToolInput
  (name String)
  (arguments (Map String JsonValue)))
```

Go:
```go
type ToolInput struct {
    Name      string
    Arguments map[string]any
}
```

Field access `(. input name)` generates `input.Name`.

### User-Defined Unions (`defunion`)

Unions use the Go interface + struct-per-variant pattern. Each union becomes an interface with a private marker method. Each variant becomes a struct that implements it. Pattern matching generates a type switch.

Vex:
```
(defunion McpMessage
  (Request Int String JsonValue)
  (Response Int (Result JsonValue McpError))
  (Notification String JsonValue))
```

Go:
```go
type McpMessage interface { isMcpMessage() }

type McpMessage_Request struct {
    V0 int64
    V1 string
    V2 any
}
func (McpMessage_Request) isMcpMessage() {}

type McpMessage_Response struct {
    V0 int64
    V1 vexrt.Result[any, McpError]
}
func (McpMessage_Response) isMcpMessage() {}

type McpMessage_Notification struct {
    V0 string
    V1 any
}
func (McpMessage_Notification) isMcpMessage() {}
```

- The Vex type checker validates all types at the Vex level
- The Go interface is not generic — it uses `any` where Vex type parameters appear — because Go's generics cannot express the full Vex type system
- This is acceptable since the generated code is correct by construction after type checking

### Concurrency Primitives

| Vex | Go |
|-----|-----|
| `(channel T)` | `chan T` |
| `(channel T size)` | `make(chan T, size)` |
| `(spawn expr)` | `go func() { expr }()` |
| `(send ch val)` | `ch <- val` |
| `(recv ch)` | `<-ch` |
| `(select ...)` | `select { ... }` |

### Pattern Matching Codegen

The `match` expression generates different Go code depending on what is being matched:

| Match target | Go codegen |
|-------------|------------|
| `Option[T]` | `if opt.IsSome { val := opt.Value; ... } else { ... }` |
| `Result[T, E]` | `if res.IsOk { val := res.Value; ... } else { err := res.Error; ... }` |
| `defunion` value | `switch v := val.(type) { case Variant1: ...; case Variant2: ... }` |
| Literal (String, Int) | `switch val { case "foo": ...; case 42: ... }` |

### Other Values

| Vex | Go | Notes |
|-----|-----|-------|
| Keyword (`:name`) | `string` | String value in generated code |
| `nil` | `nil` | Go's nil for interface/slice/map zero values |

### Naming Convention

Codegen converts Vex's kebab-case to Go's PascalCase:

| Vex | Go |
|-----|-----|
| `handle-tool-call` | `HandleToolCall` |
| `deftype` field `name` | struct field `Name` |
| `defunion` variant `Request` | type `McpMessage_Request` |

---

## 11. Testing Strategy

Each phase is tested independently by constructing its input and asserting its output.

| Phase | Test approach |
|-------|--------------|
| **Lexer** | Source string in → assert token kinds, values, and spans |
| **Parser** | Token vec in → assert AST structure (pattern match on nodes) |
| **Type checker** | Hand-built AST in → assert HIR types, or assert expected diagnostics |
| **Codegen** | Hand-built HIR in → assert generated Go source contains expected patterns |
| **Integration** | `.vx` source string → full pipeline → assert Go output |

- **Unit tests** live alongside the code in each module via `#[cfg(test)] mod tests`
- **Integration tests** go in the `tests/` directory at the crate root — full pipeline from source string to generated Go, verifying the output

---

## 12. Implementation Order

Build bottom-up, one phase at a time, each immediately testable.

| Step | Files | Milestone | Done when |
|------|-------|-----------|-----------|
| 1 | `source.rs`, `diagnostics.rs` | Foundation | `SourceMap` can store a file and convert byte offsets to line/column; `Diagnostic` can render an error with source snippet and underline. Unit tests pass. |
| 2 | `lexer.rs` | Lexer | `lex()` tokenizes `(defn main [] (println "Hello, World!"))` into the correct token sequence. Tests assert token kinds, values, and spans. |
| 3 | `ast.rs` | AST | All untyped AST node types (`Expr`, `TopForm`, `Param`, `TypeExpr`, etc.) are defined and can represent the hello world program. |
| 4 | `parser.rs` | Parser | `parse()` converts hello world tokens into the expected AST. Tests assert the resulting tree structure by pattern-matching on nodes. |
| 5 | `macro_expand.rs` | Macro expansion | `expand()` loads the prelude (self-hosted `cond`, `and`, `or` defined with `defmacro`), then expands all macro calls via a dedicated AST evaluator with automatic hygiene. |
| 6 | `types.rs`, `hir.rs`, `builtins.rs` | Type system | `VexType` enum covers all primitive and compound types. `hir::Module` mirrors the AST with resolved types. `BuiltinRegistry` contains `println` with its type signature. Unit tests pass. |
| 7 | `typechecker.rs` | Type checker | `check()` transforms the expanded AST into a valid `hir::Module` where every node carries a resolved type. Tests assert HIR types and diagnostic output for invalid programs. |
| 8 | `codegen.rs` | Codegen | `generate()` produces valid Go source from the hello world HIR. Tests assert the output contains `package main`, `func main()`, and the `fmt.Println` call. |
| 9 | `lib.rs`, `main.rs` | Pipeline | `vex build hello.vx` runs the full pipeline and `go build` produces a working binary. Integration tests pass end-to-end from `.vx` source to Go output. |

**End-to-end milestone:** `(defn main [] (println "Hello, World!"))` compiles to a working Go binary via `go build`.

### Development Process

The compiler is built incrementally through small PRs — each one adds a testable piece of the pipeline. This keeps progress visible and makes it easy to revisit decisions later.

Planned PR sequence:

| PR | Branch | Contents |
|----|--------|----------|
| 1 | `source-diagnostics` | `source.rs` + `diagnostics.rs` — FileId, Span, SourceMap, Diagnostic types |
| 2 | `lexer` | `lexer.rs` — tokenizer for hello world |
| 3 | `ast` | `ast.rs` — untyped AST types |
| 4 | `parser` | `parser.rs` — recursive descent parser |
| 5 | `macro-expand` | `macro_expand.rs` — prelude loading (self-hosted cond, and, or) + user-defined macros (defmacro) with AST evaluator and hygiene |
| 6 | `types-hir-builtins` | `types.rs` + `hir.rs` + `builtins.rs` — type representations and built-in registry |
| 7 | `typechecker` | `typechecker.rs` — AST → HIR |
| 8 | `codegen` | `codegen.rs` — HIR → Go source |
| 9 | `pipeline` | `lib.rs` + `main.rs` — wire full pipeline, CLI |

---

## 13. Developer Experience

Vex is a Lisp that targets networked servers — both of these shape its DX tooling in ways that diverge from conventional language tooling.

### Why not breakpoints

Traditional breakpoints are a poor fit for Vex for three reasons:

- **Server workload** — MCP servers handle concurrent JSON-RPC sessions; a breakpoint pauses the entire process, breaking every connected client. Backend developers debug with logging and request inspection, not by freezing execution.
- **Transpilation gap** — Vex compiles to Go. Even with `//line` directives, stepping through a debugger exposes Go's execution model, Go-mangled variable names, and multi-statement expansions of single Vex forms. The abstraction leaks more than it helps.
- **Lisp tradition** — the REPL is the debugger. Evaluate expressions against live state, test functions in isolation, inspect intermediate values. Clojure, Common Lisp, and Racket developers rarely reach for traditional debuggers because the REPL provides a tighter feedback loop.

### Language-level DX tools

These are tools the Vex compiler and toolchain provide, independent of any framework built on top.

| Priority | Tool | Owner | Rationale |
|----------|------|-------|-----------|
| P0 | Error diagnostics | Compiler (`diagnostics.rs`) | Span-based errors with source snippets, line numbers, and underlines. The single most important DX feature for a new language. |
| P0 | REPL | Interpreter (`interpreter.rs`) | Tree-walking evaluation of typed HIR. Instant feedback, no compile cycle. The primary development workflow for a Lisp. |
| P0 | `--emit-go` | CLI (`main.rs`) | Write the generated Go module to a directory instead of deleting it. Essential for understanding and debugging codegen output. |
| P1 | `vex dev` (hot reload) | CLI | File watcher that recompiles and restarts on source changes. Sub-second reload is more productive than any debugger for server development. |
| P1 | Structured logging | Stdlib (`vex.log`) | Key-value structured log output. The server developer's primary diagnostic tool in production and development. |
| P2 | Connected REPL | Toolchain | REPL that connects to a running `vex dev` process (nREPL model). Evaluate expressions in the server's context, redefine functions without restarting, inspect live state. |
| P2 | Error chain traces | Runtime / stdlib | When a `Result` error propagates through multiple functions, display the full chain with source locations. Makes error flows debuggable without a traditional stack trace debugger. |
| P3 | Test framework | Stdlib (`vex.test`) | Assertions, test discovery, test runner. General-purpose, not framework-specific. |

### What belongs in the MCP framework, not the language

The following serve MCP server authors specifically and will be part of the future MCP framework (`vex.mcp`), built as a library on top of the language-level tools above:

- **Request/response tracing** — logging all incoming JSON-RPC requests and outgoing responses with pretty-printed JSON, timing, and session context
- **MCP-aware test utilities** — mock MCP client, JSON schema validation for tool outputs, session lifecycle simulation, transport-agnostic test harness
- **Transport dev mode** — auto-configuration of stdio and Streamable HTTP transports for local testing
- **MCP Inspector compatibility** — integration with the standard MCP debugging client
- **Protocol error diagnostics** — mapping Vex `Result` errors to JSON-RPC error codes, per-tool error isolation, error response formatting

---

## 14. Intentionally Excluded from v0.1

| Feature | Rationale |
|---------|-----------|
| **String interning** | Use `String` everywhere. Optimize later if profiling shows it matters. |
| **Arena allocation** | Use `Box` and `Vec`. Swap for arenas later if needed. |
| **User-defined macros (`defmacro`)** | Implemented. All macros — including core control flow (`cond`, `and`, `or`) defined in the self-hosted prelude — execute via a dedicated AST evaluator with automatic hygiene. See `language-design.md` §4.5 and §14.11. |
| **Multi-file compilation** | `SourceMap` supports `FileId` from day one, but the pipeline processes one file at a time. |
| **Error recovery in parser** | Stop at first error initially. Accumulate multiple errors later. |
| **LSP / incremental compilation** | Not a concern at this stage. |
