---
name: pre-commit-tests
description: >-
  Runs project tests before creating git commits. Use when committing changes,
  creating commits, or running git commit.
---

# Pre-Commit Tests

Before creating any git commit, run the full test suite and ensure it passes:

```bash
cargo test
```

If tests fail, fix the issues before committing. Do not commit with failing tests.
