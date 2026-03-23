---
name: update-roadmap
description: >-
  Reviews and updates the project roadmap with current status of all items.
  Use when committing changes, completing a feature, or when the user asks
  to update the roadmap.
---

# Update Roadmap

The roadmap lives at `docs/roadmap.md`. It tracks what Vex needs next, the status of each item, and links to `docs/roadmap-rationale.md` for full analysis.

## When to Trigger

- Before every commit (called from the pre-commit-checks skill)
- A feature or milestone is completed
- A new roadmap item is identified or started
- User explicitly asks to update the roadmap

## Gather Current State

Before updating, read these sources to determine what has changed:

1. **`docs/roadmap.md`** — current roadmap with statuses
2. **`docs/roadmap-rationale.md`** — rationale and analysis for each item
3. **`src/lib.rs`** — which modules exist
4. **`src/*.rs`** — list files to see what phases and features are present
5. **Git diff / staged changes** — what changed in the current commit

## Update Process

1. Compare the staged changes against each roadmap item
2. If a roadmap item moved forward (new code, tests, or docs related to it), update its status:
   - **Not Started** → **In Progress** when work begins on a branch
   - **In Progress** → **Done** when merged to main with tests passing
3. If a completed item is not yet in the "Completed Milestones" section, move it there with the current date
4. If new work reveals a gap not on the roadmap, add it to the appropriate section
5. Cross-check `docs/roadmap-rationale.md` and `docs/language-design.md` for items, features, or gaps not yet tracked in the roadmap — add any missing items to the appropriate section with status **Not Started**
6. Update the **Last reviewed** date at the top of the document to today's date

## What to Check

For each roadmap section, verify:

### Design Constraint Enforcement
- Has the compiler core / binary boundary been enforced? Check if `lib.rs` compile functions do any IO.
- Has resilient parsing been added? Check `parser.rs` for error recovery logic.

### Type System
- Has parametric polymorphism work started? Check `types.rs` for `TypeParam`, check `typechecker.rs` for unification logic.

### Developer Experience
- Has `vex dev`, structured logging, connected REPL, error chain traces, or test framework work started? Check `src/` for new files or CLI commands.

### MCP Framework
- Has any MCP-specific macro or runtime work started? Check for `deftool`, `defresource`, `serve-mcp` in source or examples.

## Rules

- **Only report what is actually implemented.** Read the source — do not guess.
- Keep statuses consistent between the roadmap and the README phase table.
- Do not remove items from the roadmap — mark them Done and move to Completed Milestones.
- Do not add speculative items. Every item must trace back to a concrete gap identified in `docs/roadmap-rationale.md` or `docs/compiler-architecture.md`.
- **Follow `docs/documentation-guidelines.md`** — active voice, bullet points, no filler words.
