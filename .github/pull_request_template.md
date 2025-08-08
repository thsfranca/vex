## Pull Request Checklist

**Brief description of changes:**
<!-- What does this PR do? Why is it needed? -->

### ✅ Requirements (Automatically Checked)

- [ ] **All tests pass** - CI workflow must be green
- [ ] **Coverage ≥ 85%** for `internal/transpiler` (parser package excluded). Build fails if below threshold
- [ ] **Code builds successfully** - All Go code compiles
- [ ] **Examples still work** - Existing Vex programs parse correctly

### 📝 Manual Checklist

- [ ] **Added tests** for new functionality (if applicable)
- [ ] **Updated documentation** (if public API changed)
- [ ] **Follows project patterns** - Code style matches existing codebase
- [ ] **Breaking changes noted** - Any changes that affect existing functionality

### 🎯 For Language Features

If this PR adds new language features:
- [ ] **Grammar updated** (`tools/grammar/Vex.g4`) if syntax changes
- [ ] **Examples added** to `examples/valid/` to demonstrate usage
- [ ] **Parser regenerated** if grammar changes (`make go`)
- [ ] **Issue reference** - Links to milestone issue (e.g., `Closes #1`)

### 🔧 For Infrastructure Changes

If this PR modifies CI/CD, workflows, or tools:
- [ ] **Tools built successfully** (`make build-tools`)
- [ ] **Workflow tested** - Changes tested in CI environment
- [ ] **Documentation updated** if automation changes

---

**Note**: PRs are automatically blocked if overall transpiler coverage drops below 85% (generated parser excluded).

<!-- 
Quality Philosophy: We maintain high standards because:
- This is a learning project - good practices matter
- Language bugs cascade through everything
- Future you will thank present you for good tests
-->
