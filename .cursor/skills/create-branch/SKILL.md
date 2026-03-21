---
name: create-branch
description: >-
  Creates a new git branch from an up-to-date main. Use when the user asks to
  create a new branch, start a new PR, or begin a new task on a fresh branch.
---

# Create Branch

When creating a new branch, always start from an up-to-date `main`:

```bash
git checkout main
git pull origin main
git checkout -b <branch-name>
```

Never create a branch from another feature branch unless explicitly asked.
