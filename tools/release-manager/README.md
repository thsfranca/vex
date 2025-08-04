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
Bumps the version in the VERSION file based on release type.

### Create Release Notes
```bash
./release-manager create-notes '{"number":123,"title":"Fix parser","body":"Details...","author":"user","release_type":"patch"}'
```
Creates release notes from PR data.

## Release Types

- `release:major` - Breaking changes (1.0.0 → 2.0.0)
- `release:minor` - New features (1.0.0 → 1.1.0)  
- `release:patch` - Bug fixes (1.0.0 → 1.0.1)
- `release:alpha` - Alpha prerelease (1.0.0 → 1.0.1-alpha.1)
- `release:beta` - Beta prerelease (1.0.0 → 1.0.1-beta.1)
- `release:rc` - Release candidate (1.0.0 → 1.0.1-rc.1)

## Usage in CI

The auto-release workflow calls this tool to handle version management without embedded shell scripts.
