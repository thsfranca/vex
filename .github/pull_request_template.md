## Pull Request Checklist

**Brief description of changes:**
<!-- What does this PR do? Why is it needed? -->

### ✅ Requirements (Automatically Checked)

- [ ] **All tests pass** - CI workflow must be green
- [ ] **Test coverage maintained** - No component drops below quality thresholds:
  - Parser: 95%+ (critical language component)
  - Transpiler: 90%+ (core functionality)  
  - Types: 85%+ (type system implementation)
  - Standard Library: 80%+ (user-facing features)
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

**Note**: PRs are automatically blocked if test coverage drops below thresholds. The coverage check will comment with detailed results and guidance if issues are found.

<!-- 
Quality Philosophy: We maintain high standards because:
- This is a learning project - good practices matter
- Language bugs cascade through everything
- Future you will thank present you for good tests
-->