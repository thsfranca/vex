---
name: pre-commit-checks
description: >-
  Runs cargo fmt and cargo clippy before every git commit to prevent CI failures.
  Use when committing changes, creating commits, or running git commit.
---

# Pre-Commit Checks

Before every `git commit`, run these commands in order and fix any issues before proceeding:

```bash
cargo fmt
cargo clippy -- -D warnings
```

1. **`cargo fmt`** — auto-formats code. Stage any reformatted files with `git add` before committing.
2. **`cargo clippy -- -D warnings`** — if it reports warnings, fix them before committing.

Only create the commit after both commands succeed with no errors.

3. **README check** — if the commit reaches a new implementation milestone or changes the status of a compiler phase, also run the **update-readme** skill before committing.
