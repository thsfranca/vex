# Vex Language

<div align="center">
  <img src="vscode-extension/vex-logo.svg" alt="Vex Language Logo" width="128" height="128">
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
fugo/
├── docs/                           # Documentation
├── examples/                       # Example Vex programs
│   └── go-usage/                   # Go integration examples
├── tools/                          # Development and build tools
│   ├── grammar/                    # ANTLR4 grammar definition
│   ├── gen/                        # Generated parser files (created by make)
│   ├── change-detector/            # CI tool for detecting file changes
│   ├── coverage-updater/           # Tool for updating test coverage
│   ├── debug-helper/               # Debugging utilities
│   ├── extension-tester/           # VSCode extension testing tool
│   └── grammar-validator/          # Grammar validation tool
└── scripts/                        # Build and utility scripts
```

## Language Vision

Vex aims to be a functional programming language with:

- **Static typing** with type inference for performance and safety
- **Lisp-inspired syntax** using S-expressions and immutable data
- **Go transpilation** for native performance and ecosystem access
- **Concurrent programming** leveraging Go's goroutines and channels
- **Backend service focus** with built-in HTTP and concurrency primitives

### Current Status: Parser Foundation ✅

The project currently includes:
- **ANTLR4 grammar** for S-expressions, arrays, symbols, and strings
- **Multi-language parser generation** (Go, Java, Python, C++, JavaScript)
- **Example programs** demonstrating the syntax
- **Test coverage enforcement** with quality gates
- **CI/CD pipeline** with automated quality checks

### Planned Features 🚧

- **Type system** (string, int, float, symbol, list, map)
- **Go transpilation engine** for native performance
- **Immutable data structures** with structural sharing
- **Standard library** for HTTP services and data processing
- **IDE support** with Language Server Protocol

## Usage

### Prerequisites

- ANTLR4 installed and available in your PATH
- Target language runtime (Java, Go, Python, etc.) if you plan to use the generated parser

### Generating Parsers

```bash
# Generate parsers for all supported languages
make generate

# Generate parser for specific language
make java        # Java parser
make go          # Go parser
make python      # Python parser
make cpp         # C++ parser
make javascript  # JavaScript parser

# Clean generated files
make clean

# Show help
make help
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

- `sp`: The root rule, matches one or more lists followed by EOF
- `list`: Matches `(` followed by elements followed by `)`
- `array`: Matches `[` followed by elements followed by `]`
- Elements can be: arrays, lists, symbols, or strings

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

**Current Phase**: Parser and Grammar (✅ Complete)  
**Next Phase**: Type System and Transpiler  
**Timeline**: Personal learning project, developed for fun in spare time

### Test Coverage Standards

This project maintains high code quality through automated testing:

| Component | Target | Status | Purpose |
|-----------|--------|--------|---------|
| **Parser** | 95%+ | ⏳ *Not implemented yet* | Critical language component |
| **Transpiler** | 90%+ | ⏳ *Not implemented yet* | Core functionality |
| **Type System** | 85%+ | ⏳ *Not implemented yet* | Type safety |
| **Standard Library** | 80%+ | ⏳ *Not implemented yet* | User-facing features |
| **Overall Project** | 75%+ | ⏳ *Not implemented yet* | Quality baseline |

> **Quality Philosophy**: Higher coverage requirements for more critical components. PRs that reduce coverage below these thresholds are automatically blocked.

## Contributing

This is a personal study project, but feel free to:
- Try the parser generators
- Suggest language design ideas
- Report issues with the grammar
- Fork for your own experiments

**Note**: This is an educational project for learning compiler/language implementation concepts. It's not intended for production use - just for the joy of building a programming language from scratch!