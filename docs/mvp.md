# Vex MVP Definition

**Status:** Complete
**Date:** 2026-03-21

---

## Goal

Prove the complete compiler pipeline works end-to-end: Vex source → lexer → parser → macro expansion → type checker → codegen → Go source → `go build` → native binary.

The MVP is the smallest useful subset of Vex that exercises every compiler phase and produces correct, runnable programs.

---

## Success Criteria

Three programs compile with `vex build` and execute correctly with `vex run`:

### Program 1 — Hello World

```
(defn main []
  (println "Hello, World!"))
```

Expected output: `Hello, World!`

### Program 2 — Fibonacci

```
(defn fib [n : Int] : Int
  (if (<= n 1)
    n
    (+ (fib (- n 1)) (fib (- n 2)))))

(defn main []
  (println (str (fib 10))))
```

Expected output: `55`

### Program 3 — FizzBuzz

```
(defn fizzbuzz [n : Int] : String
  (cond
    (= (mod n 15) 0) "FizzBuzz"
    (= (mod n 3) 0)  "Fizz"
    (= (mod n 5) 0)  "Buzz"
    :else             (str n)))

(defn fizzbuzz-loop [i : Int, max : Int]
  (if (<= i max)
    (let []
      (println (fizzbuzz i))
      (fizzbuzz-loop (+ i 1) max))
    nil))

(defn main []
  (fizzbuzz-loop 1 100))
```

Expected output: `1`, `2`, `Fizz`, `4`, `Buzz`, ..., `FizzBuzz`, ..., `Buzz` (100 lines).

---

## Language Features in MVP

### Core Forms

| Form | Syntax | Description |
|------|--------|-------------|
| `defn` | `(defn name [params] body)` | Named function definition |
| `def` | `(def name expr)` | Top-level constant binding |
| `let` | `(let [bindings] body)` | Local bindings with body |
| `if` | `(if test then else)` | Conditional (two branches) |
| `fn` | `(fn [params] body)` | Anonymous function (lambda) |
| Call | `(f arg1 arg2 ...)` | Function application |

### Compiler-Internal Macros

These forms are available to the user but expand to primitive forms before type checking:

| Macro | Expansion | Description |
|-------|-----------|-------------|
| `cond` | Nested `if` | `(cond t1 v1 t2 v2 :else d)` → `(if t1 v1 (if t2 v2 d))` |
| `and` | `if` | `(and a b)` → `(if a b false)` |
| `or` | `let` + `if` | `(or a b)` → `(let [tmp a] (if tmp tmp b))` |

### Types

| Type | Go Mapping | Notes |
|------|------------|-------|
| `Int` | `int64` | 64-bit signed integer |
| `Float` | `float64` | 64-bit IEEE 754 |
| `Bool` | `bool` | `true` / `false` |
| `String` | `string` | UTF-8 string |
| `Unit` | (no return) | Functions that return nothing |

Function types `(Fn [T1 T2] R)` are supported for passing functions as values.

### Type Annotations

- Parameter types are required: `[n : Int]`
- Return types are optional — inferred from the body when omitted
- `(defn main [] ...)` infers `Unit` return type from a body ending in `println`

### Built-in Functions

| Name | Type | Go Translation |
|------|------|----------------|
| `+` | `(Fn [Int Int] Int)` | `(a + b)` |
| `-` | `(Fn [Int Int] Int)` | `(a - b)` |
| `*` | `(Fn [Int Int] Int)` | `(a * b)` |
| `/` | `(Fn [Int Int] Int)` | `(a / b)` |
| `mod` | `(Fn [Int Int] Int)` | `(a % b)` |
| `<` | `(Fn [Int Int] Bool)` | `(a < b)` |
| `>` | `(Fn [Int Int] Bool)` | `(a > b)` |
| `<=` | `(Fn [Int Int] Bool)` | `(a <= b)` |
| `>=` | `(Fn [Int Int] Bool)` | `(a >= b)` |
| `=` | `(Fn [Int Int] Bool)` | `(a == b)` |
| `!=` | `(Fn [Int Int] Bool)` | `(a != b)` |
| `not` | `(Fn [Bool] Bool)` | `!a` |
| `println` | `(Fn [String] Unit)` | `fmt.Println(a)` |
| `str` | `(Fn [Any...] String)` | `fmt.Sprint(args...)` |

Arithmetic and comparison operators also work on `Float`. The type checker resolves overloads based on argument types.

### CLI

```
vex build hello.vx              # Compile to ./hello binary
vex build hello.vx -o server    # Compile to ./server binary
vex run hello.vx                # Compile and execute
vex build hello.vx --emit-go .  # Also write generated Go to current dir
```

Exit codes: `0` success, `1` compilation error, `2` Go build error, `3` CLI usage error.

---

## Explicitly Not in MVP

| Feature | Rationale |
|---------|-----------|
| `deftype` / `defunion` | Records and unions add type system complexity; not needed for the three demo programs |
| Traits | Requires `deftype` first |
| `Result` / `Option` | Requires `defunion` and pattern matching on ADTs |
| Pattern matching (`match`) | Only useful with ADTs; `if`/`cond` cover MVP needs |
| User-defined macros (`defmacro`) | Entire subsystem deferred to post-MVP; core macros (`cond`, `and`, `or`) are self-hosted via `defmacro` in the prelude |
| Modules / imports | Single-file compilation only |
| Go interop (`import-go`) | Not needed until stdlib work begins |
| Concurrency (`spawn`, `channel`) | Requires runtime support beyond basic codegen |
| REPL / interpreter | Separate execution path; not needed to prove the pipeline |
| Multi-file compilation | `SourceMap` supports it, but the pipeline processes one file |
| Tail call optimization | Nice-to-have optimization, not correctness |
| `List` / `Map` literals | Collection types need generics in codegen; defer |
| `Char` type | Rarely used in the demo programs |

---

## Implementation Plan

The MVP is built bottom-up. Each step is independently testable.

| Files | What It Delivers |
|-------|------------------|
| `source.rs`, `diagnostics.rs` | Source tracking and error reporting |
| `lexer.rs` | Tokenizer: source → token stream |
| `ast.rs` | Untyped AST types for all MVP forms |
| `parser.rs` | Recursive descent parser: tokens → AST |
| `macro_expand.rs` | Compiler-internal macro expansion: `cond`, `and`, `or` → primitive forms |
| `types.rs`, `hir.rs`, `builtins.rs` | Type representations, typed AST, built-in function registry |
| `typechecker.rs` | Type inference and checking: expanded AST → HIR |
| `codegen.rs` | Go code generation: HIR → `.go` source |
| `lib.rs`, `main.rs` | Full pipeline wiring, CLI, `go build` invocation |

---

## Generated Go Output

For the hello world program, the compiler generates:

```go
package main

import "fmt"

func main() {
	fmt.Println("Hello, World!")
}
```

For Fibonacci:

```go
package main

import "fmt"

func fib(n int64) int64 {
	if n <= 1 {
		return n
	}
	return fib(n-1) + fib(n-2)
}

func main() {
	fmt.Println(fmt.Sprint(fib(10)))
}
```

The `vexrt/` runtime package is **not** needed for the MVP — all types map directly to Go primitives, and no `Result`, `Option`, or `defunion` types are used.

---

## What Comes After MVP

Once the MVP milestone is reached, the next priorities (in order) are:

1. **Records** (`deftype`) — product types with field access
2. **Unions** (`defunion`) — sum types with `match`
3. **Result / Option** — error handling with `match`
4. **Collections** — `(List T)`, `(Map K V)` with `each`, `range`, `map`, `filter`
5. **Modules** — multi-file compilation with `import` / `export`
6. **Go interop** — `import-go` for Go package access
7. **Concurrency** — `spawn`, `channel`, `send`, `recv`
8. **REPL** — tree-walking interpreter
9. **MCP framework** — `deftool`, `defresource`, `serve-mcp` macros
