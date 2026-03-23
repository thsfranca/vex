# Dependency Management

## 1. Purpose

Vex compiles to Go, but dependency management operates at the Vex level first. A Vex project declares its dependencies — both Vex packages and Go modules — in a single manifest. The compiler resolves these dependencies, downloads them, and wires everything into the generated Go build.

### Why This Exists

- `import-go` only works with Go standard library packages today. Third-party Go modules (like the MCP SDK) require `require` directives in `go.mod`, which the compiler does not generate.
- Vex packages cannot depend on other Vex packages. There is no mechanism to share Vex code between projects.
- The MCP framework (`deftool`, `defresource`, `serve-mcp`) requires both Vex library code and a Go MCP SDK. Without dependency management, the language cannot fulfill its core purpose.

---

## 2. Module Identity

A Vex module is identified by its repository URL, following the Go module convention:

```
github.com/thsfranca/my-mcp-server
```

The module path serves three purposes:

- **Identity** — uniquely names the module across all Vex projects
- **Location** — tells `vex get` where to fetch the source code
- **Import path** — other projects reference this module by its path

---

## 3. `vex.mod` Format

Every Vex project has a `vex.mod` file at its root. This file declares the module identity, the Vex version, and all dependencies.

```
module github.com/thsfranca/my-mcp-server

vex 0.1.0

require (
  github.com/thsfranca/vex-mcp v0.1.0
  github.com/thsfranca/vex-utils v0.2.0
)

go (
  github.com/mark3labs/mcp-go v0.26.0
)

replace (
  github.com/thsfranca/vex-utils => ../vex-utils
)
```

### Directives

- **`module`** — the module path. Required. One per file.
- **`vex`** — the minimum Vex compiler version required. Required.
- **`require`** — Vex package dependencies. Each entry is a module path and a semver version. The version must correspond to a Git tag in the repository.
- **`go`** — Go module dependencies needed by `import-go` statements. Each entry is a Go module path and version. The compiler merges these into the generated `go.mod`.
- **`replace`** — overrides a module path with a local directory. Used for local development when working on multiple packages simultaneously. Replaces apply only to the current module — they do not propagate to dependents.

### Rules

- `vex.mod` must live at the project root
- `vex.mod` and `go.mod` cannot coexist in the same directory (see §7)
- Lines starting with `//` are comments
- Each directive uses the grouped syntax with parentheses, or a single-line form:

```
require github.com/thsfranca/vex-mcp v0.1.0
```

---

## 4. `vex.sum` Format

The `vex.sum` file records cryptographic hashes of every dependency for reproducible builds.

```
github.com/thsfranca/vex-mcp v0.1.0 h1:abc123def456...
github.com/thsfranca/vex-mcp v0.1.0/vex.mod h1:789ghi012jkl...
github.com/mark3labs/mcp-go v0.26.0 h1:mno345pqr678...
```

Each line contains:

- Module path
- Version
- Hash type and value (`h1:` prefix for SHA-256)

Two entries per Vex dependency: one for the module content, one for its `vex.mod` file. One entry per Go dependency (content hash only).

### Behavior

- `vex get` creates or updates `vex.sum` when adding dependencies
- `vex build` verifies downloaded content against `vex.sum` before compiling
- `vex.sum` should be committed to version control
- Hash mismatches produce a clear error and abort the build

---

## 5. `vex get` Command

`vex get` adds a dependency to the current project. It auto-detects whether the target is a Vex package or a Go module.

### Syntax

```bash
vex get github.com/thsfranca/vex-mcp@v0.1.0
```

### Auto-Detection Logic

1. Fetch the repository at the specified version (Git tag)
2. Check if the repository root contains a `vex.mod` file
3. If `vex.mod` exists → **Vex package** → add to `require` section
4. If `vex.mod` does not exist → **Go module** → add to `go` section
5. Download the source to the global cache
6. Update `vex.sum` with content hashes
7. Print what happened:

```
$ vex get github.com/thsfranca/vex-mcp@v0.1.0
added github.com/thsfranca/vex-mcp v0.1.0 (vex package)

$ vex get github.com/mark3labs/mcp-go@v0.26.0
added github.com/mark3labs/mcp-go v0.26.0 (go module)
```

### Transitive Dependencies

When adding a Vex package, `vex get` reads that package's `vex.mod` and recursively fetches its dependencies:

- Vex `require` entries become transitive Vex dependencies
- Go `go` entries get merged into the current project's `go` section

### Version Update

Running `vex get` with a package that already exists in `vex.mod` updates the version:

```bash
vex get github.com/thsfranca/vex-mcp@v0.2.0
```

```
updated github.com/thsfranca/vex-mcp v0.1.0 -> v0.2.0 (vex package)
```

### Removing Dependencies

```bash
vex get -remove github.com/thsfranca/vex-mcp
```

```
removed github.com/thsfranca/vex-mcp v0.1.0
```

---

## 6. Global Cache

All downloaded dependencies live in a shared global cache. No dependency files exist inside the project directory.

### Cache Location

- Default: `~/.vex/cache/`
- Override: `VEX_CACHE` environment variable

### Cache Structure

```
~/.vex/cache/
  github.com/
    thsfranca/
      vex-mcp/
        v0.1.0/
          vex.mod
          src/
            server.vx
            handler.vx
      vex-utils/
        v0.2.0/
          vex.mod
          src/
            strings.vx
    mark3labs/
      mcp-go/
        v0.26.0/
          go.mod
          server/
            server.go
```

### Cache Behavior

- Cache entries are keyed by `module_path/version` (e.g., `github.com/thsfranca/vex-mcp/v0.1.0/`)
- A cache entry is immutable — once downloaded, it never changes
- **Download is skipped** when the cache directory for that module path and version already exists. No network request, no Git fetch, no hash recomputation. The lookup is a local filesystem check
- `vex get` downloads to the cache only if the entry does not already exist. If the version in `vex.mod` has not changed, `vex get` is a no-op and prints nothing
- `vex deps` walks all entries in `vex.mod` and downloads only the missing ones. Already-cached dependencies are skipped silently
- `vex build` reads dependencies from the cache. If a required dependency is missing, it errors with a message suggesting `vex deps`
- `vex cache clean` removes all cached dependencies

### IDE Integration

The compiler resolves import paths to cache locations. IDEs can follow these paths for code navigation, autocomplete, and go-to-definition into dependency source code.

---

## 7. `vex.mod` / `go.mod` Mutual Exclusion

A directory cannot contain both `vex.mod` and `go.mod`. The Vex compiler enforces this rule and produces clear error messages.

### Why

- A Vex project generates its own `go.mod` as a build artifact. A pre-existing `go.mod` in the same directory creates ambiguity about which file governs the Go build.
- `vex get` auto-detects package type by checking for `vex.mod`. If both files could coexist, a repository would be ambiguously classified as both Vex and Go.
- Enforcing mutual exclusion makes the classification absolute: `vex.mod` present → Vex package. `go.mod` present (no `vex.mod`) → Go module. No edge cases.

### Error Messages

When `vex init` detects a `go.mod` in the current directory:

```
error: cannot create vex.mod — go.mod already exists in this directory

A directory cannot be both a Vex project and a Go module.
Move the Vex project to a separate directory, or remove go.mod first.
```

When `vex build` detects both files:

```
error: found both vex.mod and go.mod in /path/to/project

A directory cannot contain both files. The Vex compiler generates its
own go.mod during compilation.

To fix this:
  - If this is a Vex project, remove go.mod
  - If this is a Go project, move your .vx files to a separate directory
    with its own vex.mod
```

When `--emit-go` targets a directory containing `vex.mod`:

```
error: cannot emit Go source into /path/to/project — vex.mod exists

--emit-go would create a go.mod that conflicts with the Vex project.
Use a different output directory:

  vex build main.vx --emit-go ./go-output
```

### Subdirectories

The rule applies per directory. A `go.mod` in a subdirectory does not conflict with a `vex.mod` at the project root (they define separate module boundaries). A monorepo can have:

```
monorepo/
  go-service/
    go.mod          # fine — different directory
    main.go
  vex-server/
    vex.mod          # fine — different directory
    src/
      main.vx
```

---

## 8. Dependency Resolution Flow

When `vex build` runs, the compiler resolves dependencies in this order:

1. **Read `vex.mod`** — parse the manifest to find all `require`, `go`, and `replace` entries
2. **Apply replacements** — for each `replace` directive, use the local path instead of the cache
3. **Locate Vex dependencies** — find each required Vex package in the global cache. If missing, error with a message suggesting `vex deps` or `vex get`
4. **Extract summaries** — extract type signatures and exports from each dependency (future: summary extraction phase, see `roadmap-rationale.md` §9). Until summary extraction is implemented, the compiler compiles each dependency fully
5. **Compile Vex dependencies** — compile each dependency module before the main module (existing `compile_single` flow in `lib.rs`)
6. **Collect Go dependencies** — merge the `go` section from `vex.mod` with any Go dependencies declared by Vex package dependencies (transitive)
7. **Generate `go.mod`** — produce a `go.mod` with all Go `require` directives
8. **Run `go mod download`** — download Go dependencies before `go build`
9. **Build** — run `go build` as usual

### Pipeline Diagram

```
vex.mod ──→ resolve deps ──→ compile .vx deps ──→ compile main.vx
                │                                       │
                │                                       ▼
                └──→ collect go deps ──→ generate go.mod
                                              │
                                              ▼
                                        go mod download
                                              │
                                              ▼
                                          go build
                                              │
                                              ▼
                                           binary
```

---

## 9. Integration with the Compiler Pipeline

### Changes to `codegen.rs`

`generate_go_mod()` currently returns a hardcoded string:

```rust
pub fn generate_go_mod() -> String {
    "module vex_out\n\ngo 1.21\n".to_string()
}
```

This changes to accept a list of Go dependencies:

```rust
pub fn generate_go_mod(go_deps: &[(String, String)]) -> String
```

The function generates a `go.mod` with `require` directives for each Go dependency:

```
module vex_out

go 1.21

require (
    github.com/mark3labs/mcp-go v0.26.0
)
```

### Changes to `main.rs`

Before running `go build`, the compiler runs `go mod download` in the temp directory to fetch Go dependencies:

```
go mod download → go build → binary
```

### Changes to `lib.rs`

The `compile` function gains a new parameter or reads `vex.mod` from the source directory to determine dependencies. The `CompileResult` struct carries the Go dependencies so `generate_go_mod` can include them.

---

## 10. CLI Commands

### `vex init`

Creates a new `vex.mod` in the current directory.

```bash
$ vex init github.com/thsfranca/my-server
```

Creates:

```
module github.com/thsfranca/my-server

vex 0.1.0
```

Errors if `go.mod` exists in the current directory (see §7).

### `vex get <module>@<version>`

Adds or updates a dependency. Auto-detects Vex vs Go (see §5).

### `vex deps`

Downloads all dependencies listed in `vex.mod` to the global cache. Does not build.

```bash
$ vex deps
downloading github.com/thsfranca/vex-mcp v0.1.0... done
downloading github.com/mark3labs/mcp-go v0.26.0... done
all dependencies cached
```

### `vex cache clean`

Removes all entries from the global cache.

```bash
$ vex cache clean
removed 12 cached packages from ~/.vex/cache/
```

### Updated `vex build` and `vex run`

These commands now read `vex.mod` if present. If no `vex.mod` exists, they work exactly as before (single-file mode, backwards compatible).

---

## 11. Backwards Compatibility

Single-file compilation without `vex.mod` continues to work:

```bash
vex build hello.vx
vex run hello.vx
```

When no `vex.mod` exists in the source file's directory, the compiler behaves as it does today:

- No dependency resolution
- `import-go` limited to Go standard library
- Generated `go.mod` has no `require` directives
- No global cache interaction

This means every existing Vex program compiles without changes.

---

## 12. Design Decisions

### 1. Go-inspired module paths — repository URL as identity

- No package registry to build or maintain
- Any Git repository is a valid package
- The audience (Go developers) already understands this model
- Module path maps naturally to both Vex `import` paths and Go `require` entries

### 2. Auto-detection in `vex get` — presence of `vex.mod` determines type

- Single command for both Vex and Go dependencies
- Detection is deterministic — `vex.mod` present → Vex, absent → Go
- The mutual exclusion rule (§7) makes this classification absolute
- Feedback is immediate — `vex get` prints what it detected

### 3. Global cache only — no vendor directory

- One download serves all projects on the machine
- Projects stay clean — no dependency files in the repository
- IDEs resolve import paths to cache locations for code navigation
- CI uses `vex deps` to populate the cache before building
- Vendoring can be added later via `vex vendor` if demand arises
- Follows the Dart/Flutter model, proven at scale

### 4. `vex.mod` / `go.mod` mutual exclusion

- Eliminates ambiguity in auto-detection
- Prevents conflicts between user-authored and compiler-generated `go.mod`
- Simple rule: one directory, one toolchain
- Clear error messages guide the user toward the fix

### 5. `replace` for local development — not a first-class path dependency

- Follows Go's `replace` directive pattern
- Applies only to the current module (does not propagate)
- Sufficient for monorepo and multi-package development
- Avoids adding a second dependency resolution path

### 6. Explicit versions only — no ranges, no `^`, no `~`

- `vex.mod` lists exact semver versions
- No SAT solver needed — resolution is a simple lookup
- Reproducible without a lock file (though `vex.sum` adds integrity verification)
- Follows Go's Minimal Version Selection philosophy

---

## 13. Not in Scope

| Feature | Rationale |
|---------|-----------|
| Package registry | Repository URLs eliminate the need for a central registry |
| Version ranges / resolution algorithm | Exact versions keep resolution trivial |
| Vendor directory | Global cache is sufficient; vendor can be added later |
| Private repository authentication | Relies on Git credentials (SSH keys, credential helpers) already configured on the machine |
| Workspaces / multi-module projects | Can be addressed with `replace` directives for now |
| Dependency graph visualization | Nice-to-have, not needed for v0.1.0 |
| `vex tidy` (remove unused deps) | Can be added once dependency resolution and import tracking are stable |
