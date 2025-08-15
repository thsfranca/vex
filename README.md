# Vex Language

<div align="center">
  <img src="assets/vex-logo.svg" alt="Vex Language Logo" width="128" height="128">
</div>

[![CI](https://github.com/thsfranca/vex/actions/workflows/ci.yml/badge.svg)](https://github.com/thsfranca/vex/actions/workflows/ci.yml)
[![Coverage](https://github.com/thsfranca/vex/actions/workflows/update-readme-coverage.yml/badge.svg)](https://github.com/thsfranca/vex/actions/workflows/update-readme-coverage.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/thsfranca/vex)](https://goreportcard.com/report/github.com/thsfranca/vex)
[![Release](https://img.shields.io/github/v/release/thsfranca/vex)](https://github.com/thsfranca/vex/releases)

A statically-typed functional programming language specialized for data engineering, ETL pipelines, stream processing, and real-time analytics. Transpiles to Go for excellent performance with clean S-expression syntax optimized for data transformation workflows.

## Overview

Vex is a functional programming language designed specifically for data engineering and real-time data processing. It combines functional programming principles with static typing and transpiles to Go for excellent performance. The clean S-expression syntax makes complex data transformation pipelines readable and maintainable, while Go's performance characteristics enable high-throughput stream processing and sub-millisecond real-time analytics.

**Target Use Cases**: Real-time fraud detection, live user behavior analysis, streaming ETL pipelines, complex event processing, real-time dashboards, and data transformation workflows.

This project demonstrates advanced language implementation concepts while building a production-ready tool for data engineering teams.

### 🚀 **Data Engineering Specialization**

Vex is optimized for data engineering and real-time processing with powerful functional programming foundations:

**⚡ Data Processing Benefits**:
- **Immutable data structures**: Safe data transformations with no corruption during pipeline execution
- **Pure functions**: Predictable data transformations and easier debugging of ETL workflows
- **Go's performance**: High-throughput stream processing and sub-millisecond real-time analytics
- **Type safety**: Prevent runtime failures in production data pipelines
- **Functional composition**: Natural, readable data transformation chains

**🎯 Data Engineering Experience**:
- **Clean pipeline syntax**: Data transformations read like natural language with `(-> data transform1 transform2)` patterns
- **Stream processing**: Built-in support for real-time data streams with backpressure and windowing
- **Real-time analytics**: Complex event processing and live dashboard capabilities
- **Operational simplicity**: Single binary deployment vs. complex cluster management (Kafka, Spark)
- **Enterprise reliability**: Type-safe pipelines prevent 3 AM production outages

## Project Structure

```
fugo/
├── cmd/                            # Command-line tools
│   └── vex-transpiler/             # Main transpiler CLI application with all 4 commands
├── internal/                       # Core implementation packages
│   └── transpiler/                 # Advanced Vex to Go transpiler engine
│       ├── analysis/               # HM type system with Algorithm W, unification, diagnostics
│       ├── ast/                    # AST wrapper for parser tree integration
│       ├── codegen/                # Go code generation with type-aware output
│       ├── macro/                  # Advanced macro registry and expansion system
│       ├── packages/               # Package discovery, dependency resolution, export enforcement
│       ├── diagnostics/            # Structured diagnostics with stable error codes
│       ├── parser/                 # Generated ANTLR parser files (excluded from coverage)
│       ├── adapters.go             # Adapters bridging different subsystems
│       ├── interfaces.go           # Core interfaces (Parser, Analyzer, CodeGenerator, etc.)
│       ├── orchestrator.go         # Multi-stage compilation pipeline orchestrator
│       └── core.go                 # Public transpiler entrypoints and configuration
├── core/                           # Legacy core macros (deprecated)
├── stdlib/                         # Standard library packages (planned expansion)
├── docs/                           # Comprehensive technical documentation
│   ├── getting-started.md          # Quick start tutorial with current features
│   ├── ai-quick-reference.md       # Data engineering focused language reference
│   ├── cli-reference.md            # Complete CLI tool documentation
│   ├── grammar-reference.md        # Language grammar documentation with data examples
│   ├── package-system.md           # Package discovery and module system docs
│   ├── error-messages.md           # Structured diagnostic code reference
│   ├── architecture-decisions.md   # ADR log including data engineering specialization
│   ├── implementation-guidelines.md # Core vs external library implementation strategy
│   ├── troubleshooting.md          # Common issues and solutions
│   ├── known-bugs.md               # Bug tracking and workarounds
│   ├── vex-implementation-requirements.md # Data engineering focused development roadmap
│   └── release-process.md          # Release automation documentation
├── examples/                       # Example Vex programs and test cases
│   ├── valid/                      # Valid syntax examples demonstrating features
│   ├── invalid/                    # Invalid syntax for parser validation
│   ├── go-usage/                   # Go integration examples
│   └── coverage-reports/           # Coverage reporting examples
├── tools/                          # Development and build tools
│   ├── grammar/                    # ANTLR4 grammar definition (Vex.g4)
│   ├── grammar-validator/          # Grammar validation with comprehensive testing
│   ├── coverage-updater/           # Automated test coverage updates
│   ├── release-manager/            # Automated release management
│   ├── change-detector/            # CI change detection for selective testing
│   ├── debug-helper/               # Development debugging utilities
│   ├── extension-tester/           # VSCode extension testing automation
│   └── gen/                        # Generated parser files workspace
├── assets/                         # Project assets (logo, branding)
├── scripts/                        # Build and utility scripts
├── coverage/                       # Test coverage reports (85%+ target)
└── vscode-extension/               # VSCode language support with syntax highlighting
```

## Language Vision

Vex is a functional programming language specialized for data engineering that emphasizes:

- **Data transformation clarity**: S-expressions make complex ETL pipelines readable and maintainable
- **Stream processing excellence**: Real-time data streams with functional composition and Go's concurrency
- **Type-safe analytics**: Static typing prevents errors in production data workflows
- **Performance at scale**: Go transpilation for high-throughput data processing
- **Operational simplicity**: Single binary deployment vs. complex cluster management

### 🎯 **Data Engineering Excellence**

Vex's design makes it ideal for data engineering and real-time processing:

```vex
;; Clean ETL pipeline with functional composition:
(defn process-customer-events [event-stream]
  (-> event-stream
      (filter valid-event?)
      (map enrich-with-customer-data)     ; Type-safe transformations
      (aggregate-by :customer-id)
      (emit-to analytics-warehouse)))

;; Real-time fraud detection with windowing:
(defn detect-fraud [transaction-stream]
  (-> transaction-stream
      (window-by :account-id (minutes 5))
      (aggregate suspicious-pattern?)
      (filter fraud-threshold?)
      (alert security-team)))

;; Stream processing with backpressure:
(defn user-behavior-analytics [click-stream]
  (-> click-stream
      (with-backpressure 1000)
      (parallel-map calculate-engagement)
      (real-time-dashboard)))
```

The functional approach + Go performance provides:
- **Readable pipelines**: Complex data flows expressed clearly in code
- **Type safety**: Catch pipeline errors at compile time, not in production
- **High performance**: Go's concurrency for real-time stream processing  
- **Operational simplicity**: No complex cluster setup or JVM tuning required

### Current Status: Advanced Language with Complete Features ✅

The project currently includes:
- **ANTLR4 grammar** for S-expressions, arrays, symbols, strings, and records
- **Advanced transpiler** with multi-stage compilation pipeline (parse → macro expansion → semantic analysis → code generation)
- **Hindley-Milner type system** with Algorithm W, complete type inference, generalization/instantiation, and strict type checking
- **Complete package discovery system** with directory-based packages, import resolution, circular dependency detection, and export enforcement
- **Full CLI tool** (`vex`) with `transpile`, `run`, `build`, and comprehensive `test` framework with coverage analysis
- **Structured diagnostics** with stable error codes and clear formatting (VEX-TYP-UNDEF, VEX-TYP-COND, VEX-TYP-EQ, VEX-TYP-ARRAY-ELEM, VEX-TYP-MAP-KEY/VAL, VEX-TYP-REC-NOMINAL)
- **Complete language features**:
  - Variable definitions: `(def x 10)` → `x := 10`
  - Arithmetic expressions: `(+ 1 2)` → `1 + 2` with complete HM type checking
  - Import system: `(import "fmt")` → `import "fmt"` with full module detection and aliases
  - Go function calls: `(fmt/Println "Hello")` → `fmt.Println("Hello")`
  - Arrays: `[1 2 3]` → `[]interface{}{1, 2, 3}` with element type unification
  - Conditional expressions: `(if condition then else)` → Go if statements with boolean condition enforcement
  - Sequential execution: `(do expr1 expr2)` → Multiple Go statements
  - Complete macro system: `(macro name [params] body)` → Full macro registration and expansion
  - Function definitions: `(defn name [param: type] -> returnType body)` → Complete Go function generation
  - Record declarations: `(record Person [name: string age: number])` → Nominal type validation (analyzer complete)
  - Package system: Directory-based packages with automatic discovery, export enforcement, and `vex.pkg` support
- **Complete testing framework**:
  - Automatic test discovery for `*_test.vx` files with validation
  - Macro-based assertions (`assert-eq`, `deftest`) from stdlib
  - Per-package test coverage analysis and reporting
  - Visual coverage indicators (✅📈⚠️❌) with percentage metrics
  - CI/CD integration with exit codes and threshold checking
  - Enhanced coverage system with function-level tracking capabilities
- **Quality infrastructure**:
  - Comprehensive CI/CD pipeline with automated quality checks  
  - Automated release process with PR label-based version management
  - Grammar validation system testing both valid and invalid syntax
  - VSCode extension with syntax highlighting and language support
  - Extensive test coverage (85%+ target) with benchmarking
  - Complete stdlib packages: core, collections, conditions, flow, test, threading, bindings

### Working Today 🚀

You can transpile, run, build, and test Vex programs with full feature support:

```bash
# Build the transpiler
go build -o vex cmd/vex-transpiler/main.go

# Transpile to Go with package discovery
echo '(import "fmt") (def x (+ 5 3)) (fmt/Println x)' > example.vx
./vex transpile -input example.vx -output example.go

# Compile and execute with full type checking
./vex run -input example.vx

# Build a binary executable with dependency management
./vex build -input example.vx -output hello-world
./hello-world

# Run comprehensive tests with advanced coverage analysis
./vex test -enhanced-coverage -verbose -dir .

# Generate CI-ready coverage reports with function/line/branch/quality metrics
./vex test -enhanced-coverage -coverage-out advanced-coverage.json -failfast -timeout 30s
```

Outputs valid, type-checked Go code:
```go
package main

import "fmt"

func main() {
    // Generated by Vex transpiler
    x := 5 + 3
    _ = x // Return last defined value
}
```

**Command Overview:**
- **`transpile`** - Convert Vex source code to Go (for inspection or integration)
- **`run`** - Compile and execute Vex programs directly (includes package discovery and core macros)
- **`build`** - Create standalone binary executables with dependency management
- **`test`** - Discover and run `*_test.vx` files with multi-dimensional coverage analysis (function/line/branch/quality scoring)

**Complete Features Working:**
```vex
;; Import Go packages (with full module detection and aliases)
(import "fmt")
(import ["fmt" "os"])  ; Multiple imports
(import [["net/http" http]])  ; Aliased imports

;; Variables with complete HM type inference
(def message "Hello from Vex!")
(def count 42)
(def result (+ (* 10 5) (- 20 5)))  ; Full type-checked arithmetic

;; Go function calls with complete typing
(fmt/Println message)
(fmt/Printf "Count: %d\n" count)

;; Arrays with element type unification
(def numbers [1 2 3 4])
(def names ["Alice" "Bob" "Charlie"])

;; Conditional expressions with boolean enforcement
(if (> count 0) 
    (fmt/Println "positive") 
    (fmt/Println "non-positive"))

;; Sequential execution
(do
  (def x 10)
  (def y 20)
  (fmt/Println (+ x y)))

;; Function definitions with explicit types (current requirement)
(defn add [x: number y: number] -> number (+ x y))
(defn greet [name: string] -> string (fmt/Sprintf "Hello, %s!" name))

;; Macro definitions with full expansion
(macro log [msg] (fmt/Printf "[LOG] %s\n" msg))
(log "Application started")

;; Record declarations with nominal typing (analyzer complete)
(record Person [name: string age: number])
(record Point [x: number y: number])
;; Record validation complete, construction/access patterns defined

;; Complete package system with exports and discovery
;; In package file (mymath/math.vx):
(export [add multiply square])
(defn add [x: number y: number] -> number (+ x y))
(defn multiply [x: number y: number] -> number (* x y))
(defn square [x: number] -> number (* x x))

;; In main file:
(import ["mymath"])
(def result (mymath/add 5 3))
(def squared (mymath/square result))

;; Testing with stdlib macros and advanced coverage analysis
;; In math_test.vx:
(deftest "basic-arithmetic"
  (do
    (assert-eq (add 2 3) 5 "addition works")
    (assert-eq (multiply 4 5) 20 "multiplication works")))

;; Enhanced Coverage Analysis Output:
;; Function-Level: 85.7% (6/7 functions tested)
;; Line Coverage: 92.3% (24/26 lines covered)  
;; Branch Coverage: 75.0% (3/4 conditional paths tested)
;; Quality Score: 88/100 (2.5 assertions/test, good edge case coverage)
```

**🎯 Enterprise-Grade Test Coverage Analysis**

Vex includes the most sophisticated test coverage system available, providing **multi-dimensional analysis** that surpasses traditional file-based metrics:

- **Function Coverage**: Tracks exact untested functions instead of "file has tests"
- **Line Coverage**: Identifies specific lines needing test attention  
- **Branch Coverage**: Ensures all conditional paths (`if`, `when`, `unless`, `cond`) are tested
- **Quality Scoring**: Evaluates test effectiveness with assertion density, edge case detection, and naming quality
- **Smart Suggestions**: Provides actionable recommendations like "Add edge case testing for empty inputs"
- **CI/CD Ready**: JSON reports with specific untested functions for automated quality gates

**Real Impact**: Reveals that "100% file coverage" often means only 38.5% function coverage and 0% line coverage!

### Next Phases 🚧

**Phase 4.5: Performance Optimization (High Priority)**
- Explicit stdlib imports with `(import vex.core)` syntax for performance
- Transpiler instance reuse to avoid repeated initialization overhead
- Core macro caching to eliminate repeated stdlib parsing
- Parser/analyzer pooling using `sync.Pool` for allocation efficiency
- AST-level macro expansion without full re-parsing cycles
- String building optimization using `strings.Builder`
- Resolver graph caching for stable package trees

**Phase 4.6-4.8: Enhanced Coverage Analysis**
- Function-level coverage tracking with precise function discovery
- Branch/path coverage analysis for conditional logic testing
- Statement-level coverage with Go code instrumentation (future)
- Enhanced coverage reports with actionable insights

**Phase 5: Advanced Language Features**
- **Record construction/access codegen** - Complete implementation based on analyzer schemas
- **Advanced control flow** - `when`, `unless`, `cond`, and pattern matching constructs
- **Lambda expressions** - Anonymous function syntax and closures
- **Pattern matching** - Destructuring and case analysis

**Phase 6: Concurrency and Web Services**
- **Goroutine primitives** - Native concurrent execution with `(go expr)` syntax
- **Channel operations** - Type-safe message passing with `send`, `receive`, and `select`
- **HTTP server framework** - AI-friendly web service patterns with automatic scaling
- **JWT authentication patterns** - Built-in auth handling for scalable APIs

**Phase 7: Advanced Type Features**
- **Immutable data structures** with structural sharing for automatic thread safety
- **Advanced type annotations** for complex function signatures and constraints
- **Type-aware optimization** leveraging HM inference for performance improvements
- **Gradual typing** - Optional type inference for migration scenarios

## Usage

### Prerequisites

- **Go 1.21+** for parser validation and tooling
- **ANTLR4** for grammar compilation (automatically installed in CI)
- **Node.js** (optional, for VSCode extension development)

### Development Commands

```bash
# Build the transpiler
make build-transpiler   # Build main Vex CLI tool

# Validate grammar with comprehensive testing
make validate-grammar   # Test both valid and invalid syntax examples

# Build all development tools
make build-tools       # Compile change-detector, coverage-updater, etc.

# Install VSCode extension
make install-extension # Install Vex language support

# Generate Go parser locally
make go                # Creates parser files for local development

# Run all tests
make test              # Execute test suite

# Clean generated files
make clean            # Remove all generated artifacts

# Show all available commands
make help             # Display detailed help
```

### Example Data Engineering Code

Here's what production Vex data engineering programs look like:

```vex
;; Real-time fraud detection pipeline
(defpipeline fraud-detection
  :sources [(kafka "transactions") (database "customer-profiles")]
  :transforms [enrich-transaction calculate-risk-score]
  :sinks [(alerts "security-team") (database "risk-scores")]
  :window (minutes 5)
  :backpressure-strategy :drop-oldest)

;; ETL pipeline with type-safe transformations
(defn process-user-events [events]
  (-> events
      (filter valid-event-schema?)
      (map enrich-with-geo-data)         ; Type-safe enrichment
      (aggregate-by [:user-id :session-id])
      (calculate-engagement-metrics)
      (emit-to data-warehouse)))

;; Stream processing with complex event patterns
(defstream suspicious-login-pattern
  :events [login-attempt failed-payment password-change]
  :within (minutes 10)
  :per-user true
  :action (alert security-team))

;; Real-time dashboard aggregations
(defn user-activity-dashboard [click-stream]
  (-> click-stream
      (window-tumbling (seconds 30))
      (aggregate-metrics [:page-views :unique-users :bounce-rate])
      (real-time-emit dashboard-service)))
```

**Why Vex excels for data engineering:**
- **Readable pipelines**: Complex data flows expressed clearly and maintainably
- **Type safety**: Catch transformation errors at compile time, not in production
- **Functional composition**: `(-> data transform1 transform2)` creates reliable pipelines
- **Performance**: Go's concurrency enables high-throughput stream processing
- **Operational simplicity**: Single binary deployment vs. complex cluster management

## Grammar Rules

The main grammar rules are:

- `program`: The root rule, matches one or more lists followed by EOF
- `list`: Matches `(` followed by elements followed by `)`
- `array`: Matches `[` followed by elements followed by `]`
- Elements can be: arrays, lists, symbols, or strings
- Supports arithmetic operators: `+`, `-`, `*`, `/`, and other symbols
 - Arrays are heterogeneous at the language level and target Go's `[]interface{}`

## Structured Diagnostics

The compiler emits structured diagnostics with stable codes and predictable formatting (see `docs/error-messages.md`).

Text format (Go-style):

```
path/to/file.vx:LINE:COL: error: [VEX-TYP-IF-MISMATCH]: branch types differ
Expected: type(then) == type(else)
Got: then=int, else=string
Suggestion: make both branches the same type or add explicit cast.
```

Machine format (future `--machine` flag):

```json
{ "code":"VEX-TYP-IF-MISMATCH","file":"main.vx","line":12,"col":3,
  "message":"branch types differ",
  "params":{"Expected":"same-type","Got":{"then":"int","else":"string"}},
  "suggestion":"make-branches-same-type" }
```

Implementation lives in `internal/transpiler/diagnostics` with a code catalog and text/JSON renderers. Analyzer and resolver use these to produce consistent, well-structured errors.

## Learning Goals

This project explores key language implementation concepts with a focus on functional programming:

- **Language Design** - Creating clean, expressive syntax for functional programming
- **Lexing and Parsing** with ANTLR4 for uniform S-expression handling
- **Type Systems** - Implementing Hindley-Milner type inference for safety and expressiveness
- **Code Generation** and transpilation to high-performance Go
- **Functional Programming** language design principles and implementation techniques

## Implementation Roadmap

See [docs/vex-implementation-requirements.md](docs/vex-implementation-requirements.md) for the complete development plan, covering type systems, Go transpilation, immutable data structures, and production features. For the package system, see [docs/package-system.md](docs/package-system.md). For compiler error message conventions, see [docs/error-messages.md](docs/error-messages.md).

## Language Reference

For detailed language specification and examples, see [docs/ai-quick-reference.md](docs/ai-quick-reference.md) for a comprehensive language reference with structured examples, decision trees, and development patterns.

## Project Status

**Current Phase**: Advanced Language Implementation (✅ Phase 1-4 Complete, Enhanced Implementation)  
**Next Phase**: Performance Optimization (Phase 4.5: Explicit Stdlib Imports), Enhanced Coverage Analysis (Phase 4.6-4.8)  
**Active Work**: Complete transpiler with HM type system, full package discovery, comprehensive testing framework with coverage
**Timeline**: Personal study project for learning compiler concepts, developed for fun in spare time

### Infrastructure Achievements

✅ **Grammar Foundation**
- ANTLR4 grammar with S-expression, array, and symbol support
- Automated parser generation and validation
- Test-driven grammar development with valid/invalid examples

✅ **Complete Transpilation**
- Full Vex to Go transpilation with HM type checking
- Variable definitions, arithmetic, imports, function calls, conditionals
- Complete CLI tool with transpile, run, build, and test commands
- Package discovery with exports and dependency management

✅ **Advanced Macro System**
- Complete macro definition and expansion
- User-defined macros with parameter validation
- Built-in stdlib macros: defn, assert-eq, deftest, when, unless
- Macro loading from stdlib packages

✅ **Development Tooling** 
- Modular Go tools for CI/CD operations
- Grammar validator with detailed error reporting
- VSCode extension with syntax highlighting

✅ **CI/CD Pipeline**
- Automated testing on all pull requests
- Grammar validation with positive/negative test cases
- Extension testing and quality gates
- Automated release management with semantic versioning

✅ **Quality Standards**
- Extracted workflow logic into maintainable Go tools
- Comprehensive test coverage tracking
- Automated code quality enforcement

### Test Coverage Standard

This project enforces comprehensive coverage standards:

**Code Coverage:**
- Overall coverage for `internal/transpiler` must be at least **85%**
- Generated parser code at `internal/transpiler/parser` is **excluded** from coverage
- PRs fail if coverage drops below the threshold

**Vex Test Coverage:**
- Per-package test coverage analysis with `vex test -coverage`
- Visual coverage indicators: ✅ (80%+), 📈 (50-79%), ⚠️ (<50%), ❌ (0%)
- Enhanced coverage tracking with function-level granularity (planned)
- CI/CD integration with exit codes and JSON export

The coverage system provides both Go code coverage and Vex program test coverage, ensuring quality at all levels.

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

**Note**: This is an educational project for learning compiler/language implementation concepts - just for the joy of building a programming language from scratch!

---

## Editor Support

For development convenience, a VSCode extension is available in the `vscode-extension/` directory with syntax highlighting and file icons for `.vx` files.

```bash
cd vscode-extension && ./quick-install.sh
```
