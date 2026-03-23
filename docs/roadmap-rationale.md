# Roadmap Rationale

This document records where Vex stands relative to the state of the art, what gaps exist, why each gap matters, and what trade-offs each solution involves. It guides the project's evolution without prescribing a fixed timeline.

Every recommendation below must pass through Vex's design constraints (see §0).

---

## 0. Design Constraints

These constraints filter which techniques and architectures apply to Vex. Each constraint is grounded in a concrete reason — not preference, not convention.

### Constraints carried forward

**Explicit over clever.** Prefer clarity over abstraction. Type annotations on function signatures, not global inference. This aligns with the direction of Rust, Go, and Zig — the field trends toward explicitness, not away from it.

**Separate AST and HIR.** The parser produces untyped `ast::*` types. The type checker produces typed `hir::*` types. Codegen only sees HIR. Standard practice in every serious modern compiler (Rust, Swift, Zig).

**Strict layering.** Modules form a DAG. No dependency cycles. Universal best practice.

**No indirection unless it solves a concrete problem.** No trait-based frameworks, no visitor patterns at this scale. Exhaustive matching on enums is the correct approach — Rust's compiler catches missing arms when a new variant is added. A visitor pattern adds indirection to solve a problem that `cargo build` already solves for free. This holds as long as Vex stays at its current scale (~15 type variants, a handful of consumers). Revisit if the project reaches rustc-level complexity (dozens of passes, hundreds of variants).

### Constraints revised

**Per-file independence** (replaces "pure transformations"). Each `.vx` file can be parsed, macro-expanded, and summarized (exported types and signatures extracted) without reading any other file. Cross-file analysis uses only summaries, not full ASTs.

Why this replaces "pure transformations": the old constraint described how to write the compiler (functions in, data out). The new constraint describes how to *design the language* so the compiler stays simple. Vex's module system (`module`, `export`, `import`) already provides explicit boundaries. There are no cross-file macros, no glob imports, no open trait impls. These properties enable the **map-reduce IDE architecture** (used by IntelliJ and Sorbet) — the simplest of the three architectures described in matklad's "Three Architectures for a Responsive IDE." Future language features must not break per-file independence, or Vex will be forced into a query-based architecture (Salsa, rust-analyzer) with far higher complexity.

The batch pipeline (`vex build`) stays pure — data in, data out, discard everything. Per-file independence is a stronger constraint that also governs a future LSP: index each file independently, merge indexes, resolve lazily, blow away caches on change.

**One file, one concept** (replaces "flat file structure"). A file owns one concept. Split when a file owns two independent concerns that don't share state. No line count threshold.

Why the 500-line threshold was wrong: `typechecker.rs` (3,560 lines), `codegen.rs` (3,060 lines), and `parser.rs` (2,437 lines) are each a single struct with methods that all operate on the same state. Splitting them into multiple files adds `pub` annotations, `mod` declarations, and cross-file navigation without changing the dependency structure. Zig's `Sema.zig` (the semantic analyzer) is over 20,000 lines in a single file by deliberate choice. Gleam splits by crate (`compiler-core`, `compiler-cli`, `language-server`), not by line count within a crate.

The meaningful split for Vex: separate the pure compiler core (data transformations, no IO) from the binary (CLI argument parsing, filesystem access, temp directory management, Go process invocation, REPL). `lib.rs` already provides `compile()` as a pure function. The split is partially done; the boundary needs to be enforced.

### Constraint added

**Resilient parsing.** The parser continues after errors and produces a partial tree. For a Lisp, recovery means skipping to the next top-level form when encountering unbalanced parentheses.

Why: in an editor, code is almost always in an invalid state mid-edit. A parser that stops at the first error produces no information for the rest of the file — no diagnostics, no type information, no hover. State-of-the-art parsers (Zig, rust-analyzer, tree-sitter) recover and keep going. This is a prerequisite for any future IDE support and costs little to implement for s-expression syntax.

---

## 1. Parametric Polymorphism

### What Vex has today

Vex's type system uses concrete types everywhere. `VexType` has `List(Box<VexType>)`, `Map { key, value }`, `Option(Box<VexType>)`, and so on — the *containers* are generic in representation, but the *functions that operate on them* are not.

The type checker handles `map`, `filter`, and `each` as special cases with dedicated methods (`check_map`, `check_filter`, `check_each`). Each method manually extracts the element type from the list, validates the callback signature, and constructs the result type. This pattern repeats for every collection operation that needs to work across element types.

The `builtins.rs` registry declares fixed signatures: `range` takes `(Int, Int) -> List(Int)`, arithmetic operators take `(Int, Int) -> Int`. Float overloads live in `resolve_call_type` as another special case.

### Why this matters

- **Every new collection function requires a new special-case method in the type checker.** Adding `reduce`, `flat_map`, `zip`, `take`, `drop`, or `find` each requires 40-80 lines of hand-written type logic that follows the same structural pattern.
- **User-defined generic functions are impossible.** A Vex user cannot write a function that works on `(List Int)` and `(List String)` with the same definition. The language has generic *types* but not generic *functions*.
- **The type checker carries complexity that belongs in the type system.** The 3,500 lines in `typechecker.rs` are partly a consequence of doing by hand what type variables would do structurally.

### What state of the art looks like

**Coalton** (the leading typed Lisp in production) uses Hindley-Milner inference. Coalton 0.2 (March 2026) moved from curried functions to fixed-arity functions specifically because HM inference produced confusing error messages with missing or extra arguments. The field is converging on explicit arities with local inference.

**Rust** takes a different path — generic type parameters on function signatures are explicit, local variables inside bodies are inferred. No global inference across function boundaries.

**Go** added generics in 1.18 with explicit type constraints. Go 1.26 (2026) relaxed the recursive type parameter restriction, allowing self-referential constraints like `type Adder[A Adder[A]]`. This strengthens Vex's codegen path — more Vex type patterns can map directly to Go generics.

### Trade-off: inference vs. explicitness

Full Hindley-Milner conflicts with Vex's "explicit over clever" philosophy:

- **Dislocated errors.** When the compiler infers everything, type mismatches surface at the unification point, not at the mistake point. The user sees an error about a type variable they never wrote, at a location far from the actual bug.
- **Opaque code without tooling.** If `(defn foo [x] (+ x 1))` silently infers `(Fn [Int] Int)`, the reader cannot know the type without running the compiler or using an LSP. Vex has no LSP today.
- **Surprising principal types.** HM always finds the most general type. Sometimes that type is more polymorphic than the user intended, and the mismatch surfaces downstream in confusing ways.

The core need is **parametric polymorphism** (type variables in function signatures), not **global type inference** (omitting annotations). These are separate features that often get conflated.

### Recommended design

Require type annotations on `defn` signatures (like Rust). Allow type variables in those signatures. Infer types locally inside function bodies.

Before (current Vex — `map` as a special-case builtin):

```vex
;; map only works because the type checker has 80 lines of
;; special-case logic in check_map
(map my-list (fn [x] (+ x 1)))
```

After (with parametric polymorphism — `map` as a regular function):

```vex
(defn map [lst: (List a) f: (Fn [a] b)] -> (List b)
  ;; implementation
  )

(map my-list (fn [x: Int] -> Int (+ x 1)))
```

The type checker resolves `a = Int` and `b = Int` from the concrete arguments at the call site — standard **local type argument inference** without requiring the user to write `(map Int Int my-list f)`.

### What changes in the compiler

- **`types.rs`**: `VexType::TypeVar(u32)` already exists. Add a `TypeParam { name: String }` variant for named type variables in signatures (`a`, `b`), distinct from anonymous unification variables.
- **`typechecker.rs`**: Add a unification/substitution pass that resolves `TypeParam` to concrete types at call sites. Remove `check_map`, `check_filter`, `check_each` — they become regular generic function calls.
- **`builtins.rs`**: Declare `map`, `filter`, `each` with generic signatures: `(Fn [(List a) (Fn [a] b)] (List b))`.
- **`codegen.rs`**: Go generics (1.18+) map directly — `func Map[A any, B any](lst []A, f func(A) B) []B`. Go 1.26's recursive generics further reduce impedance mismatch for self-referential type patterns.
- **Parser**: No syntax changes needed. `a` and `b` in type position are already parseable as identifiers — the type checker distinguishes them from concrete type names by checking whether they're defined types.

### What this unlocks

- User-defined generic functions and data structures
- Collection operations as regular functions, not compiler special cases
- Reduction in type checker complexity (estimated 200-400 lines removed)

---

## 2. Error Propagation (`try` Macro)

### What Vex has today

Vex's design principle #3 says "Errors are values — Result types instead of exceptions; errors must be handled or explicitly propagated." The language has `Result` and `Option` types with `match` for destructuring. But there is no propagation mechanism — every `Result`-returning call requires a full `match` to unwrap.

A typical MCP handler chains multiple fallible operations:

```vex
(defn handle-search [params: ToolParams] -> (Result JsonValue Error)
  (match (validate params)
    (Ok validated) (match (db.query validated)
                     (Ok rows) (match (json.encode rows)
                                  (Ok json) (Ok json)
                                  (Err e) (Err e))
                     (Err e) (Err e))
    (Err e) (Err e)))
```

Three levels of nesting for three fallible calls. Every `(Err e) (Err e)` arm is pure boilerplate — it re-wraps the error and returns it unchanged. MCP servers are almost entirely I/O, so every handler looks like this.

### Why this matters

- **MCP handlers chain 3-5 fallible operations minimum** (parse request, validate input, query/fetch, transform, serialize response). Without propagation, nesting depth grows linearly with the number of fallible calls.
- **The boilerplate obscures the happy path.** The actual logic (`validate → query → encode`) is buried inside match arms. The reader has to mentally filter out identical error-forwarding branches to understand what the function does.
- **Vex's own design principle promises explicit propagation** but does not deliver it.

### What state of the art looks like

Every modern language with Result-based error handling provides a propagation mechanism:

- **Rust** — `?` operator: `let rows = db.query(validated)?;`
- **Zig** — `try` keyword: `const rows = try db.query(validated);`
- **Gleam** — `use` expression: `use rows <- result.try(db.query(validated))`
- **Go** — `if err != nil { return err }` — verbose but explicit

### Why a simple `try` macro doesn't work

The obvious design — `(try expr)` expanding to `(match expr (Ok val) val (Err e) (Err e))` — breaks inside `let` bindings. If `expr` returns `Err`, the match evaluates to `(Err e)`, which gets **bound to the variable** instead of returning from the function. Execution continues with an error value where a normal value was expected.

```vex
;; BROKEN — (Err e) gets bound to `validated`, then db.query receives it
(let [validated (try (validate params))
      rows      (try (db.query validated))]
  (Ok rows))
```

Without early return semantics, an expression-level `try` cannot propagate errors through `let` bindings. Rust's `?` works because the compiler inserts a `return` — a macro cannot synthesize a `return` that exits the enclosing function.

### Recommended design: `try` / `catch`

A `try` macro that takes a binding list and a `catch` clause, expanding into nested `match`:

```vex
(try [validated (validate params)
      rows      (db.query validated)
      json      (json.encode rows)]
  (Ok json)
  (catch e (Err e)))
```

The macro expands this into nested `match`, one level per binding:

```vex
(match (validate params)
  (Err e) (Err e)
  (Ok validated)
    (match (db.query validated)
      (Err e) (Err e)
      (Ok rows)
        (match (json.encode rows)
          (Err e) (Err e)
          (Ok json) (Ok json))))
```

If any operation returns `Err`, the `catch` handler executes immediately — no variable binding, no continuation. The macro has access to all bindings, the body, and the catch clause, so it can restructure the entire form.

### Why `try` / `catch` over alternatives

Three approaches solve the propagation problem without early return:

1. **`try-let`** — a combined form that fuses `try` and `let`. Not idiomatic. Lisp favors orthogonal primitives that compose. Fusing two concerns into one compound form means you can't use error propagation outside of `let`, and you can't combine `let` with other effects.

2. **Gleam-style `use`** — a continuation-capturing form where `use x <- result.try(expr)` rewrites the rest of the block into a callback. More general than `try`/`catch`, but introduces a new syntactic concept (`use`) that Vex users would need to learn.

3. **Monadic `do` notation** — the fully general solution, works for any monad (`Result`, `Option`, `IO`). Requires higher-kinded types and type classes — heavy prerequisites that Vex doesn't need and isn't planning.

`try`/`catch` wins because:

- Every programmer knows the pattern. No new concepts to learn.
- The `catch` clause makes error handling explicit — the reader sees exactly what happens on failure, consistent with "explicit over clever."
- The `catch` clause gives flexibility that Rust's `?` does not:

```vex
;; Propagate
(catch e (Err e))

;; Propagate with context
(catch e (Err (wrap-error e "validation failed")))

;; Recover with a default
(catch e default-value)

;; Log and recover
(catch e (log-error e) empty-list)
```

### Single-expression form

For single fallible operations, `try`/`catch` also works as an expression:

```vex
(try (parse-int input)
  (catch e 0))
```

Expands to:

```vex
(match (parse-int input)
  (Ok val) val
  (Err e) 0)
```

This form works anywhere — in `let` bindings, function arguments, anywhere an expression fits — because the `catch` clause transforms the `Err` case into a non-`Result` value. The variable receives `0`, not `(Err e)`.

### What changes in the compiler

- **`stdlib/prelude.vx`**: Add the `try` macro definition. The macro inspects its arguments: if the first argument is a binding list, expand into nested `match`; if it's a single expression, expand into a flat `match`. The `catch` clause is destructured to extract the error binding name and handler body.
- **No AST, type checker, or codegen changes.** The macro expands to `match` with `Ok`/`Err` patterns — constructs the compiler already handles.

### Open question: macro complexity

The binding-list form requires the macro to iterate over pairs and build nested `match` expressions. This is more complex than `cond` (which iterates and nests `if`) but follows the same structural pattern. The single-expression form is trivial — a direct `match` expansion.

---

## 3. Pattern Match Exhaustiveness Checking

### What Vex has today

`check_match` in `typechecker.rs` validates that clause bodies have compatible types, but it does not check whether the clauses cover all variants of the scrutinee type. This compiles without any warning or error:

```vex
(defunion Shape
  (Circle Float)
  (Square Float)
  (Triangle Float Float))

(defn describe [s: Shape] -> String
  (match s
    (Circle r) (str "circle with radius " r)
    (Square s) (str "square with side " s)))
```

If `s` is a `Triangle` at runtime, the generated Go code hits an unmatched case — either a panic or silent wrong behavior, depending on the codegen.

### Why this matters

- **Refactoring becomes unsafe.** Adding a variant to a union should make the compiler flag every match that doesn't handle it. Without exhaustiveness checking, the new variant silently falls through at runtime.
- **It defeats the purpose of static typing.** The type system knows the exact set of variants. Failing to use that information at match sites is leaving value on the table — the compiler has the information to catch the bug and doesn't.
- **`Option` and `Result` are the most common match targets.** Every `(match opt (Some x) ...)` that forgets `None` is a runtime crash that the compiler could prevent.

### What state of the art looks like

Every statically typed language with algebraic data types checks exhaustiveness:

- **Rust** — the `rustc_pattern_analysis` crate implements the full Maranget usefulness algorithm, handling nested patterns, or-patterns, guards, and GADTs
- **Gleam** — uses a decision-tree approach based on Jules Jacobs's pattern matching algorithm, with a dedicated `exhaustiveness.rs` module
- **Elm** — reports missing patterns with concrete examples of unhandled values
- **Zig** — checks exhaustiveness for switch statements on tagged unions and enums

### Why Vex's case is simpler

Vex's type system has a property that makes this much easier than in Rust or Gleam: the set of matchable types with known variants is small and fixed.

Three types have known constructor sets:

1. **User-defined unions** (`defunion`) — variants declared in the type definition
2. **`Option`** — exactly `Some` and `None`
3. **`Result`** — exactly `Ok` and `Err`

There are no GADTs, no nested constructor patterns, no or-patterns, no guard-dependent exhaustiveness. Patterns are flat — a constructor pattern binds variables, it doesn't nest constructors inside constructors.

### Recommended algorithm

A set-difference check at the end of `check_match`, after all clauses are type-checked:

1. Collect the set of constructor names covered by the clause patterns
2. If any clause has a `Wildcard` or `Binding` pattern (catches everything), the match is exhaustive — done
3. Get the full variant set from the scrutinee type:
   - `VexType::Union { variants, .. }` → all variant names
   - `VexType::Option(_)` → `{"Some", "None"}`
   - `VexType::Result { .. }` → `{"Ok", "Err"}`
   - Primitive types (`Int`, `String`, `Float`) → require a wildcard/binding (can't enumerate all values)
   - `Bool` → `{"true", "false"}` (could be special-cased, but requiring a wildcard is sufficient)
4. Compute `missing = full_set - covered_set`
5. If `missing` is non-empty, emit a diagnostic:
   - For unions: `"non-exhaustive match: missing variant Triangle of Shape"`
   - For `Option`: `"non-exhaustive match: missing None"`
   - For `Result`: `"non-exhaustive match: missing Err"`

### Example diagnostic

```
error: non-exhaustive match: missing variant Triangle of Shape
  --> src/main.vx:7:3
   |
 7 | (match s
   |  ^^^^^
   |
   = help: add a (Triangle _ _) clause or a wildcard (_) pattern
```

### What changes in the compiler

- **`typechecker.rs`**: Add ~30-50 lines at the end of `check_match`, after the existing clause loop. Walk the checked clauses, collect covered constructor names, compare against the scrutinee type's variant set.
- **No AST, HIR, parser, or codegen changes.** The check runs entirely within the type checker on already-validated pattern information.

### Limitations of this approach

This handles Vex's current pattern matching but does not cover:

- **Nested patterns** — `(Some (Ok x))` matching on `(Option (Result Int String))` would require tracking coverage at each nesting level
- **Or-patterns** — multiple patterns per clause (not in Vex today)
- **Literal coverage** — `(match x 1 "one" 2 "two")` on `Int` requires a wildcard; no attempt to enumerate integers

If Vex later adds nested constructor patterns, the set-difference approach extends naturally: at each nesting level, collect the covered constructors and check against the type's variant set. The full Maranget usefulness algorithm becomes necessary only with or-patterns or guard-dependent exhaustiveness.

---

## 4. Removal of Traits / Protocols

### What the design doc had

§5.4 defined `deftrait` and `impl` syntax. The grammar included `deftrait-form` and `impl-form` as top-level forms. Design principle #2 referenced "algebraic data types and traits."

### Why traits were removed

Traits solve the problem of attaching type-specific behavior to types without modifying the type definition — the Expression Problem. This is an object-oriented concern. Functional programming solves the same cases with tools Vex already has:

- **Higher-order functions** — pass behavior as a function argument instead of requiring the type to implement an interface. `(defn save [item: T to-string: (Fn [T] String)] ...)` lets the caller decide how to serialize without any trait declaration.
- **Union types + pattern matching** — when you control the set of types, model them as variants and match. The exhaustiveness checker (§3) ensures every variant is handled.
- **Parametric polymorphism** (§1) — generic functions that work on any type without constraints. Collection operations like `map` and `filter` don't need trait bounds — the type parameter is unconstrained.

The concrete use cases from the design doc don't need traits:

- **Serialization** (`Serializable` trait) — Vex targets Go, where `encoding/json` handles serialization structurally via reflection. A built-in `json.encode` that works on records and unions covers the MCP use case without a trait system.
- **Operator overloading** — Vex has two numeric types (`Int`, `Float`). The compiler handles overloading with a finite dispatch table in `resolve_call_type`. MCP servers don't write generic numeric algorithms.
- **Interface abstraction** — Go uses structural interfaces. Any Go type with the right methods automatically satisfies the interface. Vex can lean on this through `import-go` rather than building a separate nominal trait system.

The complexity cost of traits is high:

- Trait resolution algorithm (which impl applies at each call site?)
- Coherence / orphan rules (can module A implement a trait for module B's type?)
- Interaction with type inference (ambiguous type variables when multiple impls exist)
- Go codegen impedance mismatch (Go interfaces are structural, traits are nominal — the mapping is awkward)

Gleam chose not to have type classes and is a successful production language. Vex follows the same path: if a concrete use case that cannot be solved with higher-order functions or pattern matching arises, traits can be reconsidered from that specific need.

### What changed

- **`language-design.md`**: Removed §5.4 (Traits / Protocols), `deftrait-form` and `impl-form` from the grammar, trait references from §2 and §5.1
- **Compiler**: No changes — `deftrait` and `impl` were never implemented in the parser, AST, type checker, or codegen

---

## 5. Structured Concurrency (`task-group`)

### What Vex has today

Vex has two concurrency primitives: `spawn` (fire-and-forget goroutine) and `channel` (Go channel). `spawn` maps directly to `go func() { ... }()` in the generated Go code. There is no mechanism to wait for spawned tasks to complete, propagate errors from child tasks, or scope a set of concurrent tasks to a lifetime.

The design doc's concurrency section (§10) is four lines of example code and "spawn → goroutine" as the entire model.

### Why this matters

The `mcp-go` project (a Go MCP SDK) had a goroutine leak bug in its SSE implementation — when clients disconnected, goroutines waiting on channel sends accumulated until the server exhausted memory. The fix required `context.WithCancel`, `sync.WaitGroup`, and explicit cleanup. This is the class of bug that structured concurrency prevents.

MCP request handlers commonly fan out concurrent work:

```vex
(defn handle-query [params: QueryParams] -> (Result Response Error)
  (spawn (fetch-user (. params user-id)))
  (spawn (fetch-orders (. params user-id)))
  ;; how do we wait for both? how do we get their results?
  ;; how do we cancel both if one fails?
  )
```

With raw `spawn`, there is no answer to any of those questions. The spawned goroutines are detached — the function returns before they complete, and their results are inaccessible.

### What state of the art looks like

Every major language has added structured concurrency alongside fire-and-forget:

- **Kotlin** — `coroutineScope { launch { ... } }` as the default, `GlobalScope.launch` for detached (discouraged)
- **Swift** — `TaskGroup` for scoped work, `Task.detached` for background work
- **Java (JDK 26)** — `StructuredTaskScope` with `fork()` and `join()`
- **Python 3.11** — `asyncio.TaskGroup` in the standard library
- **Go** — `errgroup.Group` as a library (not a language feature)

None of these removed fire-and-forget. They added structured concurrency as the recommended path for request-scoped work, while keeping detached tasks available for legitimate background work (file watchers, periodic cleanup, long-running listeners).

### Recommended design

Keep `spawn` as fire-and-forget. Add `task-group` as a scoped concurrency construct.

```vex
(defn handle-query [params: QueryParams] -> (Result Response Error)
  (task-group [g]
    (let [user   (spawn g (fetch-user (. params user-id)))
          orders (spawn g (fetch-orders (. params user-id)))]
      (Ok (build-response (try user) (try orders))))))
```

`task-group` provides:

- **Scoped lifetime** — all tasks spawned into `g` must complete before the `task-group` body exits
- **Error propagation** — if any task returns `Err`, the group cancels remaining tasks and returns the error
- **Result collection** — spawned tasks return futures that `try` can unwrap

`spawn` without a group argument remains fire-and-forget for background work:

```vex
(spawn (watch-file-changes config))
```

### What changes in the compiler

- **`ast.rs`**: Add `TaskGroup` expression node with a group binding name and body
- **`parser.rs`**: Parse `(task-group [name] body)` as a new special form
- **`typechecker.rs`**: Type-check the group body, track that `spawn` with a group argument returns a future type
- **`codegen.rs`**: Generate `errgroup.Group` with `context.WithCancel`. `spawn g expr` generates `g.Go(func() error { ... })`. The group body ends with an implicit `g.Wait()`.
- **`hir.rs`**: Add corresponding HIR node

### Trade-off: complexity vs. safety

This is more compiler work than the `try` macro (which required zero compiler changes). It adds a new AST node, a new HIR node, a new special form in the parser, type checker logic for futures, and Go codegen for `errgroup`. It's a real feature, not syntactic sugar.

The question is whether the MCP use case justifies it now. MCP servers in practice handle one request at a time with a small number of concurrent operations. The goroutine leak risk is real but manageable with careful use of channels. This feature becomes more valuable as Vex servers scale to handle concurrent sessions with fan-out patterns.

### Why not a macro

Unlike `try` (which expands to `match`) and `cond` (which expands to `if`), `task-group` cannot be implemented as a macro. It requires:

- A new expression type that the type checker understands (futures with typed results)
- Codegen that produces `errgroup.Group` initialization and `g.Wait()` at scope exit
- Context cancellation wiring that has no Vex-level equivalent today

These are compiler-level concerns, not syntax transformations.

---

## 6. Formatter (`vex fmt`)

### What Vex has today

No formatting tool. The user decides indentation, alignment, and line breaks manually.

### Why this matters

Every modern language ships a formatter as a first-class tool:

- **Go** — `gofmt` (2012, day one)
- **Rust** — `rustfmt` (official tool, CI-enforced across the ecosystem)
- **Gleam** — `gleam format` (ships with the compiler)
- **Zig** — `zig fmt` (ships with the compiler)

A formatter eliminates style discussions, makes code reviews focus on logic, and produces consistent output across projects. For a new language, a formatter signals maturity and reduces friction for new contributors.

### Why this is simpler for Vex than for most languages

S-expression formatting has fewer decisions than algol-style formatting:

- No operator precedence ambiguity — everything is parenthesized
- No semicolons, braces, or optional syntax — indentation follows nesting depth
- No complex expression wrapping rules — a form either fits on one line or each subform gets its own line

The core algorithm: indent each nesting level by two spaces, keep short forms on one line (under a configurable width), and break long forms with one subform per line. Special-case `defn`, `let`, `match`, `cond`, and `if` with conventional Lisp indentation rules (first argument on the same line as the head).

### Recommended design

Ship `vex fmt` as a CLI subcommand. Read `.vx` files, reformat in place (or to stdout with `--check` for CI). Use the existing lexer and parser to produce an AST, then pretty-print from the AST while preserving comments.

### What changes in the compiler

- **New file**: `formatter.rs` — AST pretty-printer with comment preservation
- **`main.rs`**: Add `vex fmt` subcommand
- **Lexer/Parser**: No changes, but comments must round-trip (the lexer already tracks comment positions via spans)

### Constraints

- Comments must survive formatting — the formatter attaches comments to the nearest AST node and re-emits them
- `vex fmt` must be idempotent — running it twice produces the same output
- The formatter reads from the parser's AST, not from raw text — this guarantees syntactically valid output

---

## 7. Source Location Mapping

### What Vex has today

The `--emit-go` flag writes generated Go source for manual inspection. No source location information connects generated Go back to `.vx` source.

### Why this matters

When a generated Go program panics, the stack trace shows Go file names, Go line numbers, and Go function names — none of which correspond to what the user wrote. Profiling tools (`pprof`, `go tool trace`) report Go-level locations. Without source mapping, debugging and performance analysis require mentally reverse-engineering the codegen.

### What state of the art looks like

- **Go** supports `//line` directives: `//line filename:line` changes the reported source location for subsequent lines in the Go file. The Go compiler, `go vet`, runtime panic traces, and `pprof` all respect these directives.
- **TypeScript** emits `.map` files for JavaScript source mapping
- **Gleam** generates Erlang with source location attributes

### Recommended design

Emit `//line` directives in generated Go code. Every HIR node carries a `Span` that maps to the original `.vx` source. The codegen phase resolves each span to a `filename:line` pair and emits a `//line` directive before the corresponding Go code.

```go
//line main.vx:5
func HandleSearch(params ToolParams) vexrt.Result[any, error] {
//line main.vx:6
    validated := Validate(params)
```

### What changes in the compiler

- **`codegen.rs`**: Before emitting each Go statement or declaration, emit a `//line` directive using the HIR node's span resolved through the `SourceMap`
- **`lib.rs`**: Pass the `SourceMap` to the codegen phase (currently only the diagnostics formatter uses it)

### Trade-offs

- Generated Go becomes harder to read with `//line` directives scattered throughout. The `--emit-go` output for debugging purposes should have an option to suppress directives (`--emit-go --no-line-directives`).
- Go's `//line` directive syntax changed slightly between versions. Target the format supported by Go 1.21+ (Vex's minimum Go version).

---

## 8. Go Toolchain Detection

### What Vex has today

`vex build` invokes `go build` and assumes Go is installed and on `PATH`. If Go is missing, the user sees a raw OS error ("command not found" or similar).

### Why this matters

Vex's build pipeline requires a Go toolchain. Every Vex user must have Go installed. This is a hard dependency that the installer and CLI should handle gracefully:

- New users who install Vex via Homebrew or a binary release may not have Go
- The error when Go is missing should explain what is needed and how to get it
- The required Go version (1.21+) should be validated, not assumed

### What state of the art looks like

- **Zig** — bundles a C compiler, eliminating the external dependency entirely
- **Gleam** — detects Erlang/Elixir installation and prints clear instructions when missing
- **Dart** — ships a full toolchain in a single SDK download

Bundling Go is possible (the Go toolchain is a single directory with no global state) but adds ~150MB to the distribution. The lighter approach: detect, validate, and guide.

### Recommended design

On `vex build` (before any compilation):

1. Check if `go` is on `PATH`
2. If missing, print:
   ```
   error: Go toolchain not found

   Vex compiles to Go source code and requires the Go toolchain to produce binaries.

   Install Go: https://go.dev/dl/
     macOS:   brew install go
     Linux:   sudo apt install golang  (or download from go.dev)
     Windows: winget install GoLang.Go
   ```
3. If found, run `go version` and parse the output
4. If the version is below 1.21, print:
   ```
   error: Go 1.21 or later required (found go1.18.3)

   Update Go: https://go.dev/dl/
   ```

### What changes in the compiler

- **`main.rs`**: Add a `check_go_toolchain()` function that runs before compilation. Call it at the start of `build` and `run` subcommands.
- No changes to the compiler core — this is a CLI concern.

---

## 9. Summary Extraction Phase

### What Vex has today

The compiler pipeline processes one file at a time: lex → parse → expand → type-check → codegen. Multi-file compilation (planned in `docs/dependency-management.md`) will compile dependency modules before the main module, but the pipeline has no explicit step to extract a module's public interface (types and function signatures) separately from type-checking function bodies.

### Why this matters

The per-file independence constraint (§0) says each file can be "summarized (exported types and signatures extracted) without reading any other file." But the architecture does not formalize summary extraction as a pipeline phase.

matklad's "Against Query Based Compilers" (February 2026) describes the map-reduce architecture that Vex's constraints are designed to enable:

> In parallel, a "summary" is extracted from each file, which is essentially just a list of types and signatures, with function bodies empty.
>
> Sequentially, a "signature evaluation" phase is run on this set of summaries, which turns type references in signatures into actual types, dealing with mutual dependencies between files. This phase is re-run whenever a summary of a file changes. Conversely, changes to the body of any function do not invalidate resolved signatures.
>
> In parallel, every function's body is type-checked.

This architecture gives two properties:

- **Parallelism** — function bodies type-check independently once signatures are resolved
- **Incremental invalidation** — changing a function body does not invalidate other files (only signature changes propagate)

Without a formalized summary phase, multi-file compilation and the future LSP will need to reinvent this boundary ad hoc.

### Recommended design

Add a **summary extraction** step between macro expansion and type checking:

```
expanded AST → extract_summary() → ModuleSummary
```

A `ModuleSummary` contains:

- Module name
- Exported type definitions (`deftype`, `defunion`) with field/variant types as unresolved `TypeExpr`
- Exported function signatures (name, parameter types, return type) as unresolved `TypeExpr`
- No function bodies

For single-file compilation, this phase is a no-op pass-through. For multi-file compilation, the pipeline becomes:

1. Parse + expand all files (parallel, per-file)
2. Extract summaries from all files (parallel, per-file)
3. Resolve signatures across summaries (sequential, cross-file)
4. Type-check function bodies (parallel, per-file, using resolved signatures)
5. Codegen (parallel, per-file)

### What changes in the compiler

- **New type**: `ModuleSummary` in `types.rs` or a new `summary.rs` — holds exported names, type definitions, and function signatures
- **`lib.rs`**: Insert `extract_summary()` between `expand()` and `check()` in the pipeline
- **`typechecker.rs`**: Accept a set of `ModuleSummary` values for imported modules when type-checking a file

### When to implement

Not needed for single-file compilation. Implement when multi-file compilation begins (`docs/dependency-management.md` §8) — this is the natural point where the summary boundary becomes load-bearing.

---

## 10. LSP Architecture

### What Vex has today

No IDE support beyond syntax highlighting (if the user configures a generic Lisp mode). The compiler runs as a batch process.

### Why this matters

Gleam v1.14-1.15 (December 2025 — March 2026) shows what drives adoption for a young language: type-directed autocompletion, context-aware compilation, code actions ("add missing type parameter," "merge case branches"), and hover documentation. These features require a language server.

The per-file independence constraint (§0) and summary extraction phase (§9) exist specifically to enable an LSP without forcing Vex into a query-based architecture.

### What state of the art looks like

matklad's "Against Query Based Compilers" (February 2026) recommends pushing queries as late as possible and using direct approaches first. Vex's language properties — explicit imports, no cross-file macros, no glob imports — align with the map-reduce model:

1. **Per-file work** (parallel): parse, expand macros, extract summary, lower to IR
2. **Cross-file merge** (sequential): resolve signatures from summaries
3. **Per-file work** (parallel): type-check bodies using resolved signatures

On file change: re-run step 1 for the changed file, diff the new summary against the old one. If the summary changed, re-run steps 2-3. If only the body changed, re-run step 3 for that file only.

The Ori language (2025-2026) distributes its LSP as a CLI subcommand (`ori lsp`), ensuring version consistency between compiler and language server. Vex should follow this pattern: `vex lsp` launches the language server, using the same compiler binary.

### Recommended capabilities (ordered by value)

1. **Diagnostics** — stream type errors and warnings on file save. Requires resilient parsing (§0) and the batch type checker.
2. **Hover** — show resolved types for expressions and function signatures. Requires the HIR with type annotations.
3. **Go to definition** — resolve symbols to their definition site. Requires the type checker's symbol table.
4. **Completions** — suggest names in scope. Requires the type environment at the cursor position.
5. **Formatter** — format on save via `vex fmt` (§6).

### What changes in the compiler

- **New file**: `lsp.rs` — LSP server using `tower-lsp` and `lsp-types` crates
- **`main.rs`**: Add `vex lsp` subcommand
- **`lib.rs`**: The `compile()` pipeline already returns diagnostics. The LSP calls the same pipeline functions and streams diagnostics to the editor.
- **Resilient parsing** (§0 constraint): required before the LSP can provide useful results on incomplete code

### When to implement

After resilient parsing and summary extraction. The LSP does not need every capability on day one — diagnostics-only is already valuable and validates the architecture.

---

## 11. Tree-sitter Grammar

### What Vex has today

No editor syntax highlighting support. Users who want highlighting must configure a generic Lisp mode.

### Why this matters

Syntax highlighting is the minimum bar for editor integration. Without it, Vex code looks like plain text. Tree-sitter grammars provide highlighting in Neovim, Helix, Zed, Emacs (tree-sitter mode), and VS Code (via extensions). A tree-sitter grammar gives Vex instant editor support across all major editors with a single implementation.

### Why this is trivial for Vex

S-expression syntax maps directly to tree-sitter's grammar DSL. The entire grammar has roughly three rules:

- **Program** → list of forms
- **Form** → atom | `(` form* `)` | `[` form* `]` | `{` form* `}`
- **Atom** → symbol | keyword | string | number | boolean | nil

Special forms (`defn`, `let`, `match`, `deftype`, `defunion`, `defmacro`) need field annotations for accurate highlighting (function names, type names, parameter names), adding ~10-15 rules.

### Recommended design

Create a `tree-sitter-vex` repository with the grammar definition (`grammar.js`), highlight queries (`queries/highlights.scm`), and installation instructions for each editor.

### What changes in the compiler

Nothing — the tree-sitter grammar is an external artifact. It references the same token and syntax rules from `docs/language-design.md` §7 but does not depend on compiler code.

### When to implement

Anytime. No dependencies on other roadmap items. High value-to-effort ratio — a few hours of work gives highlighting across all tree-sitter-enabled editors.

---

## 12. Unused Bindings and Import Warnings

### Why it matters

Every modern statically typed language warns on unused variables and imports: Rust, Go (errors on unused imports), Gleam, Elm. Without these warnings, dead code accumulates silently:

- Unused imports inflate the dependency graph and confuse readers
- Unused bindings hide logic errors — a typo in a variable name creates a new binding while the intended one goes unused
- Refactoring becomes uncertain — removing a function or type requires manual search to confirm nothing references it

For a language that targets MCP server development — where handlers evolve rapidly and prototype code is common — catching dead references early prevents production surprises.

### State of the art

- **Go** — unused imports and declared-but-not-used variables are compile **errors**, not warnings
- **Rust** — `#[warn(unused)]` is on by default; unused variables, imports, and functions produce warnings
- **Gleam** — warns on unused variables, unused function arguments, unused imports
- **Elm** — warns on unused imports and unused top-level definitions

Vex should warn (not error) on unused bindings and imports. Warnings keep the developer informed without blocking compilation during exploratory coding.

### What to implement

1. **Unused `let` bindings** — after type-checking a function body, report bindings that were never read. Convention: a leading `_` suppresses the warning (matches Rust/Go)
2. **Unused `import` names** — after macro expansion, report imported names that appear nowhere in the expanded AST
3. **Unused function parameters** — report parameters that are never referenced in the body. Convention: `_` name suppresses the warning

### Compiler changes required

- `diagnostics.rs` — add `Severity::Warning` if not already present
- `typechecker.rs` — add a `used: bool` flag to each binding in `Scope`. After checking a body, sweep unused entries and emit warnings. Skip `_`-prefixed names
- A post-macro-expansion pass (or an addition to `typechecker.rs`) — cross-reference imported names against names used in the checked module

### Trade-offs

- **Warning vs error**: warnings keep the development loop fast during prototyping. A future `--deny-warnings` flag can promote them to errors in CI
- **False positives in macros**: macro-generated code may reference bindings the user cannot see. Suppressing warnings for compiler-generated names (gensym bindings) avoids noise
- **Performance**: the `used` flag adds one boolean per binding — negligible cost
