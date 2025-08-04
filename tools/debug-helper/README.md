# Debug Helper

A Go tool for debugging CI workflows and running manual tests.

## Purpose

This tool provides utilities for:
1. Debugging CI environment and repository state
2. Simulating skip scenarios for performance testing
3. Running isolated build, test, and lint operations
4. Manual testing workflow components

## Commands

### `debug-info`
Shows comprehensive debug information about the repository and environment.

```bash
./debug-helper debug-info
```

**What it shows:**
- Environment variables (test type, debug mode, GitHub context)
- Repository structure (first 10 Go files)
- Go module information
- Current working directory context

### `skip-simulation`
Simulates the skip scenario for performance testing.

```bash
./debug-helper skip-simulation
```

**Purpose:** Shows what happens when only non-Go files change, demonstrating time savings.

### `build-only`
Builds Go packages without running tests.

```bash
./debug-helper build-only
```

**What it does:**
- Checks for main.go files
- Runs `go build ./...`
- Reports build status

### `test-only`
Runs Go tests without building or linting.

```bash
./debug-helper test-only
```

**What it does:**
- Checks for *_test.go files
- Runs `go test -v ./...`
- Reports test results

### `lint-only`
Runs linting checks without building or testing.

```bash
./debug-helper lint-only
```

**What it does:**
- Runs `go vet ./...`
- Reports linting status

## Environment Variables

The tool reads these environment variables when available:
- `TEST_TYPE` - Type of test being run
- `DEBUG_MODE` - Whether debug mode is enabled
- `GITHUB_EVENT_NAME` - GitHub event triggering the workflow
- `GITHUB_REF` - Git reference
- `GITHUB_SHA` - Commit SHA

## CI Integration

Used in manual test workflows to provide isolated testing capabilities and debug information for troubleshooting CI issues.
