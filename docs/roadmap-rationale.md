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
