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

**Coalton** (the leading typed Lisp in production) uses full Hindley-Milner inference — users rarely write type annotations, and the system infers polymorphic types automatically.

**Rust** takes a different path — generic type parameters on function signatures are explicit, local variables inside bodies are inferred. No global inference across function boundaries.

**Go** added generics in 1.18 with explicit type constraints and no inference of type parameters at call sites (the compiler infers from arguments in practice, but the mechanism is structural, not HM).

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
- **`codegen.rs`**: Go generics (1.18+) map directly — `func Map[A any, B any](lst []A, f func(A) B) []B`. Alternatively, generate monomorphized (type-erased) Go code with `interface{}` and type assertions if targeting Go < 1.18.
- **Parser**: No syntax changes needed. `a` and `b` in type position are already parseable as identifiers — the type checker distinguishes them from concrete type names by checking whether they're defined types.

### What this unlocks

- User-defined generic functions and data structures
- Collection operations as regular functions, not compiler special cases
- Reduction in type checker complexity (estimated 200-400 lines removed)
- Foundation for type classes / traits if Vex ever needs constrained polymorphism (e.g., `Numeric` constraint for `+`)

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

### Recommended design

A `try` macro in the prelude, alongside `cond`, `and`, and `or`:

```vex
(defmacro try [expr]
  (list (quote match) expr
    (list (quote Ok) (quote __try_val)) (quote __try_val)
    (list (quote Err) (quote __try_err)) (list (quote Err) (quote __try_err))))
```

The handler becomes:

```vex
(defn handle-search [params: ToolParams] -> (Result JsonValue Error)
  (let [validated (try (validate params))
        rows      (try (db.query validated))
        json      (try (json.encode rows))]
    (Ok json)))
```

Flat, readable, and the happy path reads top to bottom.

### Why a macro and not a special form

- **The macro system already supports this.** The prelude defines `cond`, `and`, and `or` as macros that expand to `if` and `let`. `try` expands to `match` — the macro expander already handles `match` in expanded output (lines 288-297 of `macro_expand.rs`). The type checker already handles `Ok`/`Err` patterns in `check_pattern`.
- **No compiler changes needed.** `try` is a purely syntactic transformation: `(try expr)` → `(match expr (Ok val) val (Err e) (Err e))`. No new AST nodes, no new type checker logic, no new codegen.
- **Consistent with the design philosophy.** `cond`, `and`, `or`, and `try` are all control-flow macros that expand to primitive forms. The compiler core stays small.

### Trade-off: early return semantics

The `try` macro as shown above works only in `let` bindings where the surrounding function returns `Result`. It does **not** provide Rust-style `?` that can appear anywhere in an expression (e.g., `(+ (try x) (try y))`), because the macro expands to a `match` that needs to be the tail expression of its scope to propagate the `Err`.

This is acceptable for Vex's use case:
- MCP handlers are sequences of steps (validate, query, transform, respond) — a `let`-binding chain is the natural structure
- Forcing `try` into `let` bindings keeps the early-return points visible, consistent with "explicit over clever"
- Rust's `?` works in arbitrary positions because it's a language-level operator with special desugaring, not a macro — that level of integration adds compiler complexity

### What changes in the compiler

- **`stdlib/prelude.vx`**: Add the `try` macro definition (5 lines)
- **Nothing else.** The macro expander, type checker, and codegen already handle every construct that `try` expands to.

### Open question: hygiene of `__try_val` and `__try_err`

The bindings `__try_val` and `__try_err` in the macro expansion are introduced by the macro. Vex's hygienic macro system automatically renames macro-introduced bindings to unique names, so these won't conflict with user code. The names shown above are for readability — the actual expanded code uses compiler-generated unique identifiers.

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
