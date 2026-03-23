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
