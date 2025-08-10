# Release Manager

A Go tool for managing Vex language releases triggered by PR labels.

## Purpose

Handles automatic version bumping and release creation when PRs are merged with specific labels.

## Commands

### Check Labels
```bash
./release-manager check-labels '["release:patch", "bug", "feature"]'
```
Checks for release labels and outputs the release type.

### Bump Version  
```bash
./release-manager bump-version patch
```
Computes the next version from the latest git tag and prints `old-version` and `new-version` for CI consumption.

### Create Release Notes
```bash
./release-manager create-notes '{"number":123,"title":"Fix parser","body":"Details...","author":"user","release_type":"patch"}'
```
Creates release notes from PR data.

### Create Tag
```bash
./release-manager create-tag 0.2.1 0.2.0 123 patch
```
Validates semver, creates and pushes `v<new-version>` tag with a descriptive message.

### Publish Release
```bash
./release-manager publish-release 0.2.1 patch
```
Creates a GitHub Release using `gh`, generates basic artifacts under `dist/`, and uploads them. Set `RELEASE_NOTES_PATH` to override notes path (defaults to `/tmp/release-notes.md`, falls back to `release-notes.md`).

## Release Types

- `release:major` - Breaking changes (1.0.0 → 2.0.0)
- `release:minor` - New features (1.0.0 → 1.1.0)  
- `release:patch` - Bug fixes (1.0.0 → 1.0.1)
- `release:alpha` - Alpha prerelease (1.0.0 → 1.0.1-alpha.1)
- `release:beta` - Beta prerelease (1.0.0 → 1.0.1-beta.1)
- `release:rc` - Release candidate (1.0.0 → 1.0.1-rc.1)

## Usage in CI

The auto-release workflow calls this tool to handle version management without embedded shell scripts.
