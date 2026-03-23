# Installation

This document describes how users install Vex on their system, how the release pipeline produces binaries, and the planned distribution channels.

For platform targets, see `language-design.md` §12. For the Go toolchain requirement, see `roadmap-rationale.md` §8.

---

## 1. Quick Install

A single command installs Vex on macOS and Linux:

```bash
curl -fsSL https://raw.githubusercontent.com/thsfranca/vex/main/install.sh | sh
```

The installer:

1. Detects the operating system and CPU architecture
2. Downloads the correct binary from GitHub Releases
3. Installs to `~/.vex/bin/vex`
4. Adds `~/.vex/bin` to the shell's `PATH` (auto-detects bash, zsh, and fish)
5. Checks for a Go toolchain (>= 1.21) and prints install instructions if missing

After installation, restart the shell or source the rc file:

```bash
source ~/.zshrc    # or ~/.bashrc, depending on your shell
vex version        # verify it works
```

### Installer Options

- **`--no-modify-path`** — skip automatic PATH modification. The installer prints the export line instead of writing it
- **`VEX_VERSION`** — install a specific version instead of the latest:

```bash
VEX_VERSION=v0.2.0 curl -fsSL https://raw.githubusercontent.com/thsfranca/vex/main/install.sh | sh
```

---

## 2. Manual Installation

Download a release archive from [GitHub Releases](https://github.com/thsfranca/vex/releases), extract the binary, and place it somewhere on your `PATH`:

```bash
tar -xzf vex-v0.1.0-darwin-arm64.tar.gz
mv vex ~/.local/bin/    # or /usr/local/bin/, or any directory on PATH
```

### Windows

Download the `.zip` archive from GitHub Releases, extract `vex.exe`, and add its directory to the system `PATH` via Settings > System > Environment Variables.

---

## 3. Prerequisites

Vex compiles to Go source code and requires the Go toolchain to produce binaries.

- **Minimum Go version:** 1.21
- The `vex build` and `vex run` commands check for Go automatically and produce clear error messages if Go is missing or outdated (see `roadmap-rationale.md` §8)

If Go is not installed:

```
error: Go toolchain not found

Vex compiles to Go source code and requires the Go toolchain to produce binaries.

Install Go: https://go.dev/dl/
  macOS:   brew install go
  Linux:   sudo apt install golang  (or download from go.dev)
  Windows: winget install GoLang.Go
```

---

## 4. Install Directory

The installer places the binary at `~/.vex/bin/vex`. This directory also serves as the root for other Vex-managed data:

```
~/.vex/
  bin/
    vex              # the compiler binary
  cache/             # global dependency cache (see dependency-management.md §6)
```

The `VEX_HOME` environment variable overrides the default `~/.vex/` location.

---

## 5. PATH Setup

The installer automatically adds `~/.vex/bin` to the user's shell configuration. The behavior follows the same pattern as rustup, Bun, and other modern toolchain installers:

- **bash** — appends to `~/.bash_profile` (falls back to `~/.bashrc` if `~/.bash_profile` does not exist)
- **zsh** — appends to `~/.zshrc`
- **fish** — appends to `~/.config/fish/config.fish`

The line the installer adds:

```bash
# bash / zsh
export PATH="$HOME/.vex/bin:$PATH"
```

```fish
# fish
set -gx PATH "$HOME/.vex/bin" $PATH
```

### Idempotency

The installer checks if the PATH line already exists before writing. Running the installer multiple times does not produce duplicate entries. The check uses a simple grep for `.vex/bin` in the target file.

### Opt-out

Pass `--no-modify-path` to skip automatic PATH modification:

```bash
curl -fsSL https://raw.githubusercontent.com/thsfranca/vex/main/install.sh | sh -s -- --no-modify-path
```

The installer prints the export line so the user can add it manually.

---

## 6. Release Artifacts

Each release produces pre-built binaries for all primary targets:

| Artifact name                      | OS      | Architecture | Format   |
| ---------------------------------- | ------- | ------------ | -------- |
| `vex-v{version}-darwin-amd64.tar.gz`  | macOS   | x86_64       | tar.gz   |
| `vex-v{version}-darwin-arm64.tar.gz`  | macOS   | ARM64        | tar.gz   |
| `vex-v{version}-linux-amd64.tar.gz`   | Linux   | x86_64       | tar.gz   |
| `vex-v{version}-linux-arm64.tar.gz`   | Linux   | ARM64        | tar.gz   |
| `vex-v{version}-windows-amd64.zip`    | Windows | x86_64       | zip      |

Naming follows the `{name}-v{version}-{os}-{arch}.{ext}` convention, using Go's `GOOS`/`GOARCH` naming for familiarity with the target audience.

Each archive contains a single `vex` binary (or `vex.exe` on Windows). No wrapper scripts, no configuration files, no runtime dependencies.

### Linux Static Linking

Linux binaries are compiled against `musl` (not glibc) to produce fully static executables. This eliminates glibc version mismatches across distributions — the binary runs on any Linux system regardless of the installed C library.

---

## 7. Release Pipeline

A GitHub Actions workflow builds and publishes release artifacts when a version tag is pushed:

```
git tag v0.1.0
git push origin v0.1.0
```

### Build Matrix

| Rust target                      | Runner         | Notes                              |
| -------------------------------- | -------------- | ---------------------------------- |
| `x86_64-apple-darwin`            | `macos-13`     | Intel Mac                          |
| `aarch64-apple-darwin`           | `macos-latest` | Apple Silicon (M-series)           |
| `x86_64-unknown-linux-musl`      | `ubuntu-latest`| Static binary, musl libc           |
| `aarch64-unknown-linux-musl`     | `ubuntu-latest`| Cross-compiled via `cross`         |
| `x86_64-pc-windows-msvc`         | `windows-latest`| MSVC toolchain                    |

### Workflow Steps

1. **Build** — each matrix entry compiles with `cargo build --release --target`, then packages the binary into the archive format (tar.gz for Unix, zip for Windows)
2. **Release** — after all builds succeed, creates a GitHub Release with all artifacts attached and auto-generated release notes

---

## 8. CLI Version and Help

The Vex binary supports standard version and help flags:

```bash
vex version          # prints: vex 0.1.0
vex --version        # same
vex -V               # same

vex --help           # prints usage with all subcommands
vex -h               # same
```

The version string uses the version from `Cargo.toml` via `env!("CARGO_PKG_VERSION")`, ensuring the binary version always matches the release tag.

---

## 9. Go Toolchain Detection

Before compiling any `.vx` file, `vex build` and `vex run` validate the Go toolchain:

1. Check if `go` exists on `PATH` by running `go version`
2. If `go` is not found, print the "Go toolchain not found" error (§3) and exit
3. If `go` is found, parse the version from the output (format: `go version go1.X.Y ...`)
4. If the version is below 1.21, print:

```
error: Go 1.21 or later required (found go1.18.3)

Update Go: https://go.dev/dl/
```

This check runs once per build invocation. It adds one subprocess call (~10ms) before compilation.

The check lives in the CLI (`main.rs`), not the compiler core (`lib.rs`). The compiler core remains a pure function with no IO.

---

## 10. Future Distribution Channels

These channels are planned but not implemented yet:

### Homebrew (macOS)

A Homebrew tap (`homebrew-vex`) with a formula that:

- Downloads the correct binary from GitHub Releases
- Links it into the Homebrew prefix
- Declares Go as a dependency
- Validates Go version on install

```bash
brew tap thsfranca/vex
brew install vex
```

### APT / RPM (Linux)

Debian and RPM packages for distribution via:

- A PPA (Ubuntu/Debian)
- A Copr repository (Fedora/RHEL)

Packages declare `golang` as a runtime dependency.

### Scoop / WinGet (Windows)

- **Scoop** — a manifest in a Scoop bucket repository
- **WinGet** — a manifest in the WinGet Community Repository

### Shell Completions

Generate shell completion scripts for bash, zsh, and fish via a CLI subcommand:

```bash
vex completions bash > /etc/bash_completion.d/vex
vex completions zsh > ~/.zfunc/_vex
vex completions fish > ~/.config/fish/completions/vex.fish
```

The installer can optionally install completions automatically alongside the binary.

---

## 11. Design Decisions

### 1. `~/.vex/bin/` as the install directory — not `/usr/local/bin/`

- No `sudo` required for installation
- Matches rustup (`~/.cargo/bin/`), Deno (`~/.deno/bin/`), and Bun (`~/.bun/bin/`)
- Co-locates with the global cache (`~/.vex/cache/`) under a single root
- Users who prefer a system-wide install can move the binary manually

### 2. Automatic PATH modification — opt-out, not opt-in

- rustup and Bun modify PATH automatically — the established pattern for developer toolchains
- Deno does not auto-modify PATH and faces persistent usability complaints (GitHub issue #286, PR #295)
- Auto-setup reduces first-run friction to zero for new users
- Experienced users who manage PATH manually use `--no-modify-path`
- Idempotent writes prevent accumulating duplicate entries

### 3. Go as an explicit dependency — not bundled

- Bundling Go adds ~150MB to the distribution
- Zig bundles its C compiler because no standard C compiler exists across platforms. Go has a single canonical installer at go.dev
- Gleam detects Erlang and prints install instructions — the same approach works for Vex
- Clear error messages (§9) make the dependency visible at the right moment

### 4. musl for Linux — not glibc

- glibc version mismatches cause "GLIBC_2.XX not found" errors on older distributions
- musl produces fully static binaries that run on any Linux kernel version
- Performance difference is negligible for a compiler binary (startup-dominated, not compute-bound)

### 5. Artifact naming uses Go-style `os-arch` — not Rust triple

- Vex targets Go developers. `darwin-arm64` is familiar; `aarch64-apple-darwin` is not
- Shorter names reduce visual clutter in release pages
- The installer maps `uname` output to these names, not Rust targets
