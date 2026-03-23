# Roadmap

**Last reviewed:** 2026-03-23

This document tracks what Vex needs next, why each item matters, and its current status. Each item links back to `docs/roadmap-rationale.md` for full analysis and trade-off discussion.

For design constraints that govern all future work, see `docs/roadmap-rationale.md` §0.

---

## Status Legend


| Symbol      | Meaning                        |
| ----------- | ------------------------------ |
| Not Started | Work has not begun             |
| In Progress | Active development on a branch |
| Done        | Merged to main, tested         |


---

## Design Constraint Enforcement

These are structural changes to align the codebase with the revised design constraints in `docs/roadmap-rationale.md` §0.


| Item                                    | Status      | Description                                                                                                                                                                                            |
| --------------------------------------- | ----------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| Enforce compiler core / binary boundary | Not Started | Separate pure compiler core (data transformations, no IO) from the binary (CLI, filesystem, Go process invocation). `lib.rs` already exposes `compile()` — enforce that no IO leaks into the core.     |
| Resilient parsing                       | Not Started | Parser continues after errors and produces a partial tree. For s-expression syntax, recovery means skipping to the next top-level form on unbalanced parentheses. Prerequisite for future IDE support. |
| Summary extraction phase                | Not Started | Extract exported types and function signatures from each file independently, before type-checking bodies. Formalizes the per-file independence constraint as a pipeline step. Prerequisite for multi-file compilation and LSP. See `roadmap-rationale.md` §9. |


---

## Type System


| Item                    | Status      | Rationale reference       |
| ----------------------- | ----------- | ------------------------- |
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
- `codegen.rs` — generate Go generics (1.18+; Go 1.26 recursive generics reduce impedance mismatch)
- Parser — no changes needed (lowercase identifiers in type position already parseable)

---

## Error Handling


| Item                                  | Status      | Rationale reference       |
| ------------------------------------- | ----------- | ------------------------- |
| Error propagation (`try` / `catch`)   | Not Started | `roadmap-rationale.md` §2 |
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

## Diagnostics


| Item                                    | Status      | Rationale reference        |
| --------------------------------------- | ----------- | -------------------------- |
| Unused bindings / imports warnings      | Not Started | `roadmap-rationale.md` §12 |


### Unused Bindings / Imports — Summary

The compiler accepts `let` bindings and `import` declarations that are never referenced. This makes dead code invisible and slows refactoring — the developer has no signal that an import or binding can be removed.

**Compiler changes required:**

- `typechecker.rs` — track a `used: bool` flag per binding in `Scope`. After checking a function body, emit a warning diagnostic for every unused entry
- `macro_expand.rs` or a post-expansion pass — track which imported names appear in the expanded AST. Emit warnings for unreferenced imports
- `diagnostics.rs` — add `Severity::Warning` (currently only `Error` exists)

---

## Concurrency


| Item                            | Status      | Rationale reference       |
| ------------------------------- | ----------- | ------------------------- |
| Structured concurrency (`task-group`) | Not Started | `roadmap-rationale.md` §5 |


### Structured Concurrency — Summary

Vex has `spawn` (fire-and-forget goroutine) and `channel` (Go channel) but no mechanism to scope concurrent tasks, wait for completion, or propagate errors from child tasks. MCP handlers that fan out concurrent work have no way to collect results or cancel on failure.

`task-group` adds scoped concurrency: all tasks spawned into a group must complete before the group exits, errors cancel remaining tasks, and spawned tasks return futures.

**Compiler changes required:**

- `ast.rs` / `hir.rs` — add `TaskGroup` expression node
- `parser.rs` — parse `(task-group [name] body)` as a special form
- `typechecker.rs` — type-check group body, track that `spawn` with a group argument returns a future type
- `codegen.rs` — generate `errgroup.Group` with `context.WithCancel`

---

## Developer Experience

Items from `docs/compiler-architecture.md` §13 and state-of-the-art gaps identified in `docs/roadmap-rationale.md`.


| Item                          | Priority | Status      | Rationale reference              | Description                                                                    |
| ----------------------------- | -------- | ----------- | -------------------------------- | ------------------------------------------------------------------------------ |
| `vex fmt` (formatter)         | P1       | Not Started | `roadmap-rationale.md` §6        | Opinionated code formatter for `.vx` files, shipped as a CLI subcommand        |
| `vex dev` (hot reload)        | P1       | Not Started | `compiler-architecture.md` §13   | File watcher that recompiles and restarts on source changes                    |
| Structured logging            | P1       | Not Started | `compiler-architecture.md` §13   | Key-value structured log output via `vex.log` stdlib                           |
| Source location mapping       | P1       | Not Started | `roadmap-rationale.md` §7        | Emit `//line` directives in generated Go so stack traces point to `.vx` source |
| Go toolchain detection        | P1       | Not Started | `roadmap-rationale.md` §8        | Validate Go installation and version before compilation with clear error messages |
| Connected REPL                | P2       | Not Started | `compiler-architecture.md` §13   | REPL that connects to a running `vex dev` process (nREPL model)                |
| Error chain traces            | P2       | Not Started | `compiler-architecture.md` §13   | Full chain display when `Result` errors propagate through multiple functions   |
| Test framework                | P3       | Not Started | `compiler-architecture.md` §13   | Assertions, test discovery, test runner via `vex.test` stdlib                  |


---

## IDE Support

Editor tooling, from lightweight (tree-sitter) to full (LSP). Each item builds on the one above.


| Item                  | Priority | Status      | Rationale reference        | Description                                                                                  |
| --------------------- | -------- | ----------- | -------------------------- | -------------------------------------------------------------------------------------------- |
| Tree-sitter grammar   | P1       | Not Started | `roadmap-rationale.md` §11 | `tree-sitter-vex` grammar for syntax highlighting in Neovim, Helix, Zed, Emacs, and VS Code  |
| `vex lsp` (LSP)       | P2       | Not Started | `roadmap-rationale.md` §10 | Language server shipped as a CLI subcommand, starting with diagnostics and hover              |


### Tree-sitter Grammar — Summary

S-expression syntax maps directly to tree-sitter's grammar DSL (~15-20 rules). Provides instant syntax highlighting across all tree-sitter-enabled editors. No compiler dependencies — can be built anytime.

### LSP — Summary

The per-file independence constraint (§0) and summary extraction phase enable a map-reduce LSP architecture without query-based complexity (Salsa, rust-analyzer). The architecture follows matklad's recommendation: parse/expand/summarize per-file in parallel, merge summaries sequentially, type-check bodies per-file in parallel. On file change, re-run only the changed file's pipeline; propagate only if the summary changed.

**Prerequisites:** resilient parsing, summary extraction phase

**Capabilities (ordered by implementation priority):**

1. Diagnostics — stream errors/warnings on file save
2. Hover — show resolved types for expressions
3. Go to definition — resolve symbols to definition sites
4. Completions — suggest names in scope
5. Format on save — via `vex fmt`

---

## MCP Framework

The end goal of Vex. These items build on top of the language-level features above.


| Item                                           | Status      | Description                                                                      |
| ---------------------------------------------- | ----------- | -------------------------------------------------------------------------------- |
| `deftool` / `defresource` / `serve-mcp` macros | Not Started | Core MCP server authoring macros                                                 |
| Request/response tracing                       | Not Started | Logging all incoming JSON-RPC requests and outgoing responses                    |
| MCP-aware test utilities                       | Not Started | Mock MCP client, JSON schema validation, session lifecycle simulation            |
| Transport dev mode                             | Not Started | Auto-configuration of stdio and Streamable HTTP transports                       |
| Session management                             | Not Started | Session ID handling, Origin header validation per the 2025-11-25 MCP spec update |
| Protocol error diagnostics                     | Not Started | Mapping Vex `Result` errors to JSON-RPC error codes                              |


---

## Completed Milestones


| Milestone                                                     | Date |
| ------------------------------------------------------------- | ---- |
| MVP — hello world, fibonacci, fizzbuzz compile and run        | Done |
| Records (`deftype`) and field access                          | Done |
| Unions (`defunion`) and pattern matching (`match`)            | Done |
| `Result` / `Option` types                                     | Done |
| Collections — `List`, `Map`, `each`, `range`, `map`, `filter` | Done |
| Modules — `module`, `export`, `import`                        | Done |
| Go interop — `import-go`                                      | Done |
| Concurrency — `spawn`, `channel`, `send`, `recv`              | Done |
| REPL — tree-walking interpreter                               | Done |
| Self-hosted macros — `defmacro` with automatic hygiene        | Done |


