# Vex Language

<div align="center">
  <img src="assets/vex-logo.svg" alt="Vex Language Logo" width="128" height="128">
</div>

[![CI](https://github.com/thsfranca/vex/actions/workflows/ci.yml/badge.svg)](https://github.com/thsfranca/vex/actions/workflows/ci.yml)
[![Test Coverage](https://github.com/thsfranca/vex/actions/workflows/test-coverage.yml/badge.svg)](https://github.com/thsfranca/vex/actions/workflows/test-coverage.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/thsfranca/vex)](https://goreportcard.com/report/github.com/thsfranca/vex)

A statically-typed functional programming language that transpiles to Go, designed for learning language implementation concepts and exploring functional programming paradigms.

## Overview

Vex is an experimental programming language that combines functional programming principles with static typing, inspired by Lisp syntax and designed to transpile to Go for maximum performance. This is a **study project created for educational purposes and fun** - it's not intended for production use.

The goal is to explore language design concepts including parsing, type systems, code generation, and Go interoperability while building a complete language implementation from scratch.

## Project Structure

```
vex/
├── cmd/                            # Command-line tools
│   └── fugo-transpiler/            # Main transpiler CLI application
├── internal/                       # Core implementation packages
│   └── transpiler/                 # Vex to Go transpiler engine
│       ├── parser/                 # Generated ANTLR parser files
│       ├── transpiler.go           # Main transpiler logic
│       ├── ast_visitor.go          # AST traversal and code generation
│       └── codegen.go              # Go code generation utilities
├── docs/                           # Documentation
│   ├── grammar-reference.md        # Language grammar documentation
│   ├── vex-implementation-requirements.md # Development roadmap
│   └── release-process.md          # Release automation documentation
├── examples/                       # Example Vex programs
│   ├── valid/                      # Valid syntax examples for testing
│   ├── invalid/                    # Invalid syntax for parser validation
│   └── go-usage/                   # Go integration examples
├── tools/                          # Development and build tools
│   ├── grammar/                    # ANTLR4 grammar definition (Vex.g4)
│   ├── grammar-validator/          # Grammar validation with Go parser
│   └── (various other tools)/      # CI, debugging, and release utilities
├── .github/                        # CI/CD infrastructure
│   ├── workflows/                  # GitHub Actions workflows
│   └── scripts/                    # Extracted workflow scripts
├── assets/                         # Project assets (logo, etc.)
├── scripts/                        # Build and utility scripts
└── vscode-extension/               # VSCode language support (optional)
```

## Language Vision

Vex aims to be a functional programming language with:

- **Static typing** with type inference for performance and safety
- **Lisp-inspired syntax** using S-expressions and immutable data
- **Go transpilation** for native performance and ecosystem access
- **Concurrent programming** leveraging Go's goroutines and channels
- **Backend service focus** with built-in HTTP and concurrency primitives

### Current Status: Basic Transpiler Working ✅

The project currently includes:
- **ANTLR4 grammar** for S-expressions, arrays, symbols, and strings
- **Working transpiler** that converts Vex to executable Go code
- **CLI tool** (`fugo-transpiler`) for command-line transpilation
- **Basic language features**:
  - Variable definitions: `(def x 10)` → `x := 10`
  - Arithmetic expressions: `(+ 1 2)` → `1 + 2`
  - String and number literals with proper Go output
- **Comprehensive CI/CD pipeline** with automated quality checks
- **Automated release process** with PR label-based version management
- **Grammar validation system** testing both valid and invalid syntax
- **Example programs** demonstrating working features

### Working Today 🚀

You can transpile basic Vex programs to Go:

```bash
# Install and use the transpiler
go build -o fugo-transpiler cmd/fugo-transpiler/main.go
echo '(def x (+ 5 3))' > example.vex
./fugo-transpiler -input example.vex -output example.go
```

Outputs valid Go code:
```go
package main

func main() {
    x := 5 + 3
}
```

### Next Phase 🚧

- **Type system** (string, int, float, symbol, list, map)
- **Function definitions** and calls with proper Go function generation
- **Immutable data structures** with structural sharing
- **Standard library** for HTTP services and data processing
- **IDE support** with Language Server Protocol

## Usage

### Prerequisites

- **Go 1.21+** for parser validation and tooling
- **ANTLR4** for grammar compilation (automatically installed in CI)
- **Node.js** (optional, for VSCode extension development)

### Development Commands

```bash
# Validate grammar with comprehensive testing
make validate-grammar    # Test both valid and invalid syntax examples

# Build all development tools
make build-tools        # Compile change-detector, coverage-updater, etc.

# Generate Go parser locally
make go                 # Creates parser files for local development

# Clean generated files
make clean             # Remove all generated artifacts

# Show all available commands
make help              # Display detailed help
```

### Example Vex Code (Vision)

Here's what Vex programs might look like when fully implemented:

```vex
; HTTP service with static typing
(defn handle-user [req: HttpRequest] -> Response
  (let [user-id: int (parse-int req.params.id)
        user: User (db-get-user user-id)
        json: string (json-marshal user)]
    {:status 200 :body json}))

; Concurrent data processing
(defn process-items [items: [Item]] -> [Result]
  (parallel-map transform-item items))

; Immutable data structures
(def users: {string User} 
  (-> {}
      (assoc "alice" {:name "Alice" :age 30})
      (assoc "bob" {:name "Bob" :age 25})))
```

## Grammar Rules

The main grammar rules are:

- `program`: The root rule, matches one or more lists followed by EOF
- `list`: Matches `(` followed by elements followed by `)`
- `array`: Matches `[` followed by elements followed by `]`
- Elements can be: arrays, lists, symbols, or strings
- Supports arithmetic operators: `+`, `-`, `*`, `/`, and other symbols

## Learning Goals

This project explores key language implementation concepts:

- **Lexing and Parsing** with ANTLR4
- **Type Systems** and static analysis
- **Code Generation** and transpilation
- **Language Interoperability** 
- **Functional Programming** language design
- **Performance Optimization** through native compilation

## Implementation Roadmap

See [docs/vex-implementation-requirements.md](docs/vex-implementation-requirements.md) for the complete development plan, covering type systems, Go transpilation, immutable data structures, and production features.

## Project Status

**Current Phase**: Basic Go Transpilation (🚧 In Progress)  
**Next Phase**: Advanced Language Features and Type System  
**Timeline**: Personal learning project, developed for fun in spare time

### Infrastructure Achievements

✅ **Grammar Foundation**
- ANTLR4 grammar with comprehensive S-expression support
- Automated parser generation and validation
- Test-driven grammar development with valid/invalid examples

✅ **Development Tooling** 
- Modular Go tools for CI/CD operations
- Grammar validator with detailed error reporting
- VSCode extension with full language support

✅ **CI/CD Pipeline**
- Automated testing on all pull requests
- Grammar validation with positive/negative test cases
- Extension testing and quality gates
- Automated release management with semantic versioning

✅ **Quality Standards**
- Extracted workflow logic into maintainable Go tools
- Comprehensive test coverage tracking
- Automated code quality enforcement

### Test Coverage Standards

This project maintains high code quality through automated testing:

| Component | Target | Status | Purpose |
|-----------|--------|--------|---------|
| **Grammar Validation** | 100% | ✅ *Active* | Parser correctness with valid/invalid test cases |
| **Development Tools** | 90%+ | ✅ *Active* | CI/CD infrastructure reliability |
| **VSCode Extension** | 85%+ | ✅ *Active* | IDE integration quality |
| **Parser** | 95%+ | ⏳ *Next phase* | Critical language component |
| **Transpiler** | 90%+ | ⏳ *Planned* | Core functionality |
| **Type System** | 85%+ | ⏳ *Planned* | Type safety |
| **Standard Library** | 80%+ | ⏳ *Planned* | User-facing features |

> **Quality Philosophy**: Higher coverage requirements for more critical components. The current infrastructure ensures robust development practices for future language implementation.

## Release Process

Vex uses an automated release system triggered by PR labels:

- **`release:patch`** - Bug fixes and minor improvements
- **`release:minor`** - New features and enhancements  
- **`release:major`** - Breaking changes

When a PR with a release label is merged to `main`, the system automatically:
1. Bumps the version number
2. Creates a Git tag
3. Generates release notes
4. Updates the changelog

See [docs/release-process.md](docs/release-process.md) for detailed information.

## Contributing

This is a personal study project, but feel free to:
- Test the grammar validation system
- Suggest language design ideas
- Report issues with the grammar or tooling
- Fork for your own experiments
- Try the editor support tools

**Note**: This is an educational project for learning compiler/language implementation concepts. It's not intended for production use - just for the joy of building a programming language from scratch!

---

## Editor Support

For development convenience, a VSCode extension is available in the `vscode-extension/` directory with syntax highlighting and file icons for `.vx` files.

```bash
cd vscode-extension && ./quick-install.sh
```