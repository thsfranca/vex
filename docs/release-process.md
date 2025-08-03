# Vex Language Release Process

## 🎯 PR Label-Based Auto Release

The Vex language uses an automated release system triggered by PR labels. When you merge a PR with specific labels, it automatically creates a new release.

## 🏷️ Release Labels

Add one of these labels to your PR **before merging**:

| Label | Release Type | Example | Use Case |
|-------|--------------|---------|----------|
| `release:patch` | Patch release | `0.1.0` → `0.1.1` | Bug fixes, small improvements |
| `release:minor` | Minor release | `0.1.0` → `0.2.0` | New features, non-breaking changes |
| `release:major` | Major release | `0.1.0` → `1.0.0` | Breaking changes, major milestones |
| `release:alpha` | Alpha prerelease | `0.1.0` → `0.1.1-alpha.1` | Experimental features |
| `release:beta` | Beta prerelease | `0.1.0` → `0.1.1-beta.1` | Feature complete, testing |
| `release:rc` | Release candidate | `0.1.0` → `0.1.1-rc.1` | Final testing before stable |

## 🔄 How It Works

### 1. Add Release Label to PR
```
# In your PR, add one of the release labels
# Example: Add "release:patch" label for a bug fix
```

### 2. Merge PR to Main
When you merge the PR:
- ✅ Auto-release workflow triggers
- ✅ Version gets bumped automatically  
- ✅ Git tag is created (`v0.1.1`)
- ✅ Release notes generated from PR
- ✅ Language release workflow builds artifacts

### 3. Release Created
- 🎉 New release appears on GitHub
- 📦 Artifacts built automatically
- 📝 Release notes include PR details

## 📋 Examples

### Bug Fix Release
```bash
# PR: "Fix parser memory leak"
# Label: release:patch
# Result: 0.1.0 → 0.1.1
```

### New Feature Release
```bash
# PR: "Add array literal syntax" 
# Label: release:minor
# Result: 0.1.0 → 0.2.0
```

### Experimental Release
```bash
# PR: "Experimental: async syntax"
# Label: release:alpha
# Result: 0.1.0 → 0.1.1-alpha.1
```

### Breaking Change
```bash
# PR: "BREAKING: New syntax for functions"
# Label: release:major  
# Result: 0.1.0 → 1.0.0
```

## 🛠️ Manual Release (Alternative)

If you prefer manual control, you can still:

1. **Edit VERSION file**: Change `0.1.0` to `0.1.1`
2. **Create tag manually**: `git tag v0.1.1 && git push origin v0.1.1`
3. **Use existing workflow**: `language-release.yml` builds the release

## 📁 Files Involved

```
.github/workflows/auto-release.yml     # Main auto-release workflow
tools/release-manager/                 # Go tool for version management
scripts/create-release-tag.sh          # Git tagging script
VERSION                                # Current version file
```

## 🎯 Best Practices

### PR Titles
- Use clear, descriptive titles
- They become release notes
- Examples:
  - ✅ "Fix grammar validation for nested arrays"
  - ✅ "Add support for string interpolation" 
  - ❌ "Update stuff"

### When to Release
- **Patch**: Bug fixes, documentation, small tweaks
- **Minor**: New language features, grammar additions
- **Major**: Breaking syntax changes, major rewrites
- **Alpha/Beta**: Experimental features, testing

### Release Frequency
- **For study projects**: Release whenever you have something working
- **Document progress**: Each release shows learning milestones
- **Keep experimenting**: Alpha releases are perfect for trying new ideas

## 🚨 Troubleshooting

### No Release Created
- ✅ Check if PR was merged (not just closed)
- ✅ Verify release label was added before merge
- ✅ Check workflow logs in GitHub Actions

### Wrong Version
- ✅ Check current VERSION file content
- ✅ Verify label type (patch vs minor vs major)
- ✅ Look at recent releases for version history

### Manual Fix
If something goes wrong:
```bash
# Fix VERSION file manually
echo "0.1.2" > VERSION
git add VERSION
git commit -m "fix: correct version"
git tag v0.1.2
git push origin main v0.1.2
```

---

**This system makes releasing easy while keeping you in control! 🚀**