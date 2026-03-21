---
name: update-skills
description: >-
  Reviews and updates existing Cursor skills when project conventions, compiler phases,
  type mappings, or testing patterns change. Use proactively after significant changes
  to the compiler pipeline, design docs, or project structure.
---

# Update Skills

When the Vex project evolves, skills can become stale. This skill triggers a review after significant changes.

## When to Trigger

Review skills after any of these:

- New compiler phase added or phase signature changed
- New type added to `VexType` or Vex-to-Go type mapping changed
- New file added to `src/` or file responsibilities shifted
- Design docs (`docs/language-design.md`, `docs/compiler-architecture.md`) updated with new conventions
- Testing strategy or patterns changed
- New built-in function patterns introduced
- Build or output model changed

## Review Process

1. Read the current design docs to understand what changed:
   - `docs/language-design.md` — syntax, types, grammar
   - `docs/compiler-architecture.md` — pipeline, file structure, dependencies, type mapping, testing

2. Read each skill and compare against the current state:

   | Skill | What to check |
   |-------|--------------|
   | `add-vex-language-feature` | Does the checklist match the current file list in `src/`? Are all pipeline steps still accurate? Any new steps needed? |
   | `add-compiler-phase` | Are the phase signatures still correct? Does the dependency graph match? Is the `lib.rs` wiring description accurate? |
   | `vex-test-patterns` | Do the code examples match the current API shapes? Are there new test patterns being used that should be documented? |
   | `vex-go-codegen` | Does the type mapping table match `docs/compiler-architecture.md` §10? Any new Go patterns for new Vex features? |

3. For each skill that needs updating:
   - Read the full skill file
   - Read the relevant source files to verify current behavior
   - Update only what changed — don't rewrite stable sections
   - Keep skills under 500 lines

4. Check if a new skill is warranted — if a recurring multi-step workflow has emerged that no existing skill covers, create one.

## What NOT to Update

- Don't update skills for trivial changes (renaming a variable, fixing a typo in source)
- Don't add speculative content about unimplemented features
- Don't duplicate information that belongs in the design docs — skills reference docs, not replace them
