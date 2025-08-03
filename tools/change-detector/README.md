# Change Detector

A Go tool for analyzing repository changes to determine if Go-related tests should run in CI.

## Purpose

This tool helps optimize CI workflows by:
1. Analyzing Makefile changes for Go-related patterns
2. Making final decisions about whether to run Go tests
3. Avoiding unnecessary builds when only documentation or VSCode extension files change

## Commands

### `makefile-analysis`
Analyzes git diff of Makefile to detect Go-related changes.

```bash
./change-detector makefile-analysis
```

**Output:** Prints `go-related=true` or `go-related=false`

### `final-decision`
Makes final decision about running Go tests based on basic file changes and Makefile analysis.

```bash
./change-detector final-decision [basic-go-files] [makefile-go-related]
```

**Arguments:**
- `basic-go-files`: "true" if basic Go files changed (optional, also reads BASIC_GO env var)
- `makefile-go-related`: "true" if Makefile has Go-related changes (optional, also reads MAKEFILE_GO env var)

**Output:** Prints `go-files=true` or `go-files=false`

## Go-Related Patterns

The tool detects these patterns in Makefile changes:
- `go build`, `go test`, `go mod`, `go generate`
- `GOPATH`, `GOOS`, `GOARCH`
- `.go` file references
- `antlr.*-Dlanguage=Go`

## CI Integration

Used in CI workflows to optimize build times by skipping Go tests when only non-Go files are modified.