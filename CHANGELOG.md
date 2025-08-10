# Changelog

All notable changes to Vex will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [v0.3.1] - 2025-08-10

## 🎉 Auto-Release from PR #51

**Release Type:** patch  
**Triggered by:** @thsfranca

### 📋 Changes from PR #51

**chore: switch auto-release to tag-based versioning; remove VERSION**

What: Move auto-release to derive current version from latest git tag; remove VERSION dependency.

Why: Prevent failures when VERSION is stale; make tags the single source of truth.

Changes:
- tools/release-manager: compute versions from latest tag; stop reading/writing VERSION.
- .github/workflows/auto-release.yml: fetch tags for version computation.
- .github/scripts/release/auto-bump-version.sh: compute next version only; no file writes.
- scripts/create-release-tag.sh: only tag/push; no VERSION commit.
- docs/release-process.md: remove VERSION references; document tag-based flow.
- Remove VERSION file.

CI: No functional changes to language build/tests; release flow now tag-driven.

Testing: All Go tests passing locally.

---

*This release was automatically created from the merged pull request.*

<!-- Release notes will be automatically generated here -->

## [v0.2.0] - 2025-08-04

## Release v0.2.0

**Release Type**: minor

**Changes:**
- patch: comprehensive auto-release and project improvements

**Full Changelog**: https://github.com/thsfranca/vex/compare/v0.1.1...v0.2.0

## [v0.1.2] - 2025-08-04

## Release v0.1.2

**Release Type**: patch

**Changes:**
- fix: repair auto-release workflow and add release badge

**Full Changelog**: https://github.com/thsfranca/vex/compare/v0.1.1...v0.1.2
