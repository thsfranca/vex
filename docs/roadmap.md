# Roadmap

**Last reviewed:** 2026-03-23

This document tracks what Vex needs next, why each item matters, and its current status. Each item links back to `docs/roadmap-rationale.md` for full analysis and trade-off discussion.

For design constraints that govern all future work, see `docs/roadmap-rationale.md` §0.

---

## Status Legend

| Symbol | Meaning |
|--------|---------|
| Not Started | Work has not begun |
| In Progress | Active development on a branch |
| Done | Merged to main, tested |

---

## Design Constraint Enforcement

These are structural changes to align the codebase with the revised design constraints in `docs/roadmap-rationale.md` §0.

| Item | Status | Description |
|------|--------|-------------|
| Enforce compiler core / binary boundary | Not Started | Separate pure compiler core (data transformations, no IO) from the binary (CLI, filesystem, Go process invocation). `lib.rs` already exposes `compile()` — enforce that no IO leaks into the core. |
| Resilient parsing | Not Started | Parser continues after errors and produces a partial tree. For s-expression syntax, recovery means skipping to the next top-level form on unbalanced parentheses. Prerequisite for future IDE support. |

---

## Type System

| Item | Status | Rationale reference |
|------|--------|---------------------|
| Parametric polymorphism | Not Started | `roadmap-rationale.md` §1 |

### Parametric Polymorphism — Summary

The type system has generic containers (`List(Box<VexType>)`, `Map { key, value }`, `Option`) but no generic functions. Collection operations (`map`, `filter`, `each`) are special-cased in the type checker. Adding type variables to function signatures enables:

- User-defined generic functions
- Collection operations as regular functions (removes ~200-400 lines of special-case type checker code)
- Foundation for future type classes / traits

**Compiler changes required:**

- `types.rs` — add `TypeParam { name: String }` variant for named type variables in signatures
- `typechecker.rs` — add unification/substitution pass; remove `check_map`, `check_filter`, `check_each`
- `builtins.rs` — declare `map`, `filter`, `each` with generic signatures
- `codegen.rs` — generate Go generics (1.18+) or monomorphized code
- Parser — no changes needed (lowercase identifiers in type position already parseable)

---

## Error Handling

| Item | Status | Rationale reference |
|------|--------|---------------------|
| Error propagation (`try` / `catch`) | Not Started | `roadmap-rationale.md` §2 |
| Pattern match exhaustiveness checking | Not Started | `roadmap-rationale.md` §3 |

### Error Propagation — Summary

Vex has `Result` and `Option` types but no propagation mechanism. Every fallible call requires a full `match` with boilerplate `(Err e) (Err e)` arms. MCP handlers chain 3-5 fallible operations, creating deep nesting.

A `try` / `catch` macro solves this with two forms:

- **Block form** — takes a binding list and a `catch` clause, expands into nested `match`:
  `(try [x (op1) y (op2 x)] (Ok y) (catch e (Err e)))`
- **Expression form** — single operation with recovery:
  `(try (parse-int s) (catch e 0))`

**Compiler changes required:**

- `stdlib/prelude.vx` — add the `try` macro definition
- No AST, type checker, or codegen changes needed

### Exhaustiveness Checking — Summary

`match` does not verify that all variants of a union, `Option`, or `Result` are covered. Missing variants cause runtime failures that the compiler has enough information to catch at compile time.

**Compiler changes required:**

- `typechecker.rs` — add ~30-50 lines at the end of `check_match` to compare covered constructors against the scrutinee type's variant set

---

## Developer Experience

Items from `docs/compiler-architecture.md` §13 that are not yet implemented.

| Item | Priority | Status | Description |
|------|----------|--------|-------------|
| `vex dev` (hot reload) | P1 | Not Started | File watcher that recompiles and restarts on source changes |
| Structured logging | P1 | Not Started | Key-value structured log output via `vex.log` stdlib |
| Connected REPL | P2 | Not Started | REPL that connects to a running `vex dev` process (nREPL model) |
| Error chain traces | P2 | Not Started | Full chain display when `Result` errors propagate through multiple functions |
| Test framework | P3 | Not Started | Assertions, test discovery, test runner via `vex.test` stdlib |

---

## MCP Framework

The end goal of Vex. These items build on top of the language-level features above.

| Item | Status | Description |
|------|--------|-------------|
| `deftool` / `defresource` / `serve-mcp` macros | Not Started | Core MCP server authoring macros |
| Request/response tracing | Not Started | Logging all incoming JSON-RPC requests and outgoing responses |
| MCP-aware test utilities | Not Started | Mock MCP client, JSON schema validation, session lifecycle simulation |
| Transport dev mode | Not Started | Auto-configuration of stdio and Streamable HTTP transports |
| Protocol error diagnostics | Not Started | Mapping Vex `Result` errors to JSON-RPC error codes |

---

## Completed Milestones

| Milestone | Date |
|-----------|------|
| MVP — hello world, fibonacci, fizzbuzz compile and run | Done |
| Records (`deftype`) and field access | Done |
| Unions (`defunion`) and pattern matching (`match`) | Done |
| `Result` / `Option` types | Done |
| Collections — `List`, `Map`, `each`, `range`, `map`, `filter` | Done |
| Modules — `module`, `export`, `import` | Done |
| Go interop — `import-go` | Done |
| Concurrency — `spawn`, `channel`, `send`, `recv` | Done |
| REPL — tree-walking interpreter | Done |
| Self-hosted macros — `defmacro` with automatic hygiene | Done |
