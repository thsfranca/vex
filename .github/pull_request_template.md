## Pull Request Checklist

**Brief description of changes:**
<!-- What does this PR do? Why is it needed? -->

### ✅ Requirements (Automatically Checked)

- [ ] **All tests pass** - CI workflow must be green
- [ ] **Test coverage maintained** - No component drops below quality thresholds:
  - Parser: 95%+ (critical language component)
  - Transpiler: 90%+ (core functionality)  
  - Type System: 85%+ (type safety)
  - Standard Library: 80%+ (user-facing features)
  - Overall: 75%+ (project baseline)
- [ ] **Code builds successfully** - All Go code compiles
- [ ] **Examples still work** - Existing Fugo programs parse correctly

### 📝 Manual Checklist

- [ ] **Added tests** for new functionality (if applicable)
- [ ] **Updated documentation** (if public API changed)
- [ ] **Follows project patterns** - Code style matches existing codebase
- [ ] **Breaking changes noted** - Any changes that affect existing functionality

### 🎯 For Language Features

If this PR adds new language features:
- [ ] **Grammar updated** (`tools/grammar/Fugo.g4`) if syntax changes
- [ ] **Examples added** to demonstrate usage
- [ ] **Issue reference** - Links to milestone issue (e.g., `Closes #1`)

---

**Note**: PRs are automatically blocked if test coverage drops below thresholds. The coverage check will comment with detailed results and guidance if issues are found.

<!-- 
Quality Philosophy: We maintain high standards because:
- This is a learning project - good practices matter
- Language bugs cascade through everything
- Future you will thank present you for good tests
-->