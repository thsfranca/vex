---
name: update-readme
description: >-
  Updates the project README.md with current project information, development state,
  and next steps. Use proactively when the project reaches a new implementation
  milestone or a compiler phase changes status. Do not wait for PRs to be merged.
---

# Update README

The README lives at the project root (`README.md`). It is the public face of the project and must accurately reflect the current state.

## When to Trigger

Use this skill proactively — do not wait for the user to ask or for a PR to merge.

- A commit reaches an implementation milestone — see the "Done when" column in `docs/compiler-architecture.md` §12 for the concrete criteria
- A new compiler phase is added to the pipeline
- User explicitly asks to update the README

## Gather Current State

Before writing, read these sources to determine what is actually implemented:

1. **`src/lib.rs`** — which modules are declared tells you what exists
2. **`src/*.rs`** — list the files to see what phases are present
3. **`docs/compiler-architecture.md` §12** — the implementation order table defines milestones and the planned PR sequence
4. **`Cargo.toml`** — version, dependencies
5. **`docs/language-design.md` §1** — purpose and vision for the project summary

Cross-reference `lib.rs` module declarations against the full implementation order to determine:
- Which phases are **done** (module exists and has tests)
- Which phase is **in progress** (module exists but may be incomplete, or is on the current branch)
- Which phases are **next** (not yet started)

## README Structure

Use this template. Keep each section concise — the README is a quick overview, not a design doc.

```markdown
# Vex

[![CI](https://github.com/thsfranca/vex/actions/workflows/ci.yml/badge.svg)](https://github.com/thsfranca/vex/actions/workflows/ci.yml)

A statically typed Lisp for building MCP servers. The compiler is written in Rust and transpiles to Go source code.

**This is a study project.**

## Overview

[2-3 sentences: what Vex is, S-expression syntax, compile-time type checking, pipeline produces Go → native binary. Reference MCP.]

## Example

[Show the hello world program — this never changes until the language syntax changes]

\```
(defn main []
  (println "Hello, World!"))
\```

## Current Status

[State the current milestone clearly. Use the implementation order from the architecture doc.]

| Phase | Status |
|-------|--------|
| `source.rs` — FileId, Span, SourceMap | Done |
| `diagnostics.rs` — Diagnostic, Severity, formatting | Done |
| `lexer.rs` — Tokenizer | [Done / In Progress / Not Started] |
| `ast.rs` — Untyped AST types | [Done / In Progress / Not Started] |
| `parser.rs` — Recursive descent parser | [Done / In Progress / Not Started] |
| `types.rs` / `hir.rs` / `builtins.rs` — Type system | [Done / In Progress / Not Started] |
| `typechecker.rs` — AST → HIR | [Done / In Progress / Not Started] |
| `codegen.rs` — HIR → Go source | [Done / In Progress / Not Started] |
| `lib.rs` / `main.rs` — Full pipeline, CLI | [Done / In Progress / Not Started] |

**First milestone:** `(defn main [] (println "Hello, World!"))` compiles to a working Go binary.

## Architecture

[1-2 sentences pointing to the design docs for details.]

- Pipeline: Source → Lexer → Parser → Type Checker → Codegen → `go build` → Binary
- Design docs: [`docs/language-design.md`](docs/language-design.md), [`docs/compiler-architecture.md`](docs/compiler-architecture.md)

## Building

\```
cargo build
cargo test
\```

## License

MIT
```

## Review Existing Documentation

After updating the README, review all files in `docs/` against `docs/documentation-guidelines.md`. Fix any violations you find:

- Rewrite passive voice sentences into active voice
- Break dense paragraphs into bullet points or numbered lists
- Remove filler words ("basically," "simply," "just," "actually," "really")
- Ensure each section starts with a clear one-line summary

Only fix style and structure — do not change technical content or meaning.

## Rules

- **Only report what is actually implemented.** Read the source files — do not guess.
- Keep the status table aligned with the implementation order in `docs/compiler-architecture.md` §12.
- Don't add sections for features that don't exist yet (e.g., don't add "Usage" until the CLI works).
- Don't add verbose descriptions of planned features — the design docs handle that.
- Keep the README under 80 lines. It's a quick reference, not documentation.
- Preserve the CI badge and "study project" note — these are intentional.
- **Follow `docs/documentation-guidelines.md`** — use active voice, prefer bullet points over paragraphs, remove filler words, write for humans first.
