## Pull Request Checklist

**Brief description of changes:**
<!-- What does this PR do? Why is it needed? -->

### ✅ Requirements (Automatically Checked)

- [ ] **All tests pass** - CI workflow must be green
- [ ] **Test coverage maintained** - Coverage tracking for learning purposes:
  - **Dynamic targets**: 40%+ for growing project, 70%+ for mature codebase
  - **Component goals**: Parser (95%), Transpiler (90%), Types (85%), Standard Library (80%)
  - **Note**: Coverage is tracked and reported but doesn't block PRs in this study project
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

**Note**: This is a study project with educational focus. Coverage is tracked and reported for learning purposes but doesn't block PRs. The CI uses adaptive coverage suggestions based on project maturity.

<!-- 
Quality Philosophy: We maintain high standards because:
- This is a learning project - good practices matter
- Language bugs cascade through everything
- Future you will thank present you for good tests
-->