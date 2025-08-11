# Vex Language Implementation Requirements

## Overview

Vex is a statically-typed functional programming language designed specifically for **AI code generation** and high-performance backend services. The language transpiles to Go to achieve maximum performance while maintaining excellent Go ecosystem interoperability and providing uniform, predictable syntax patterns that AI models can reliably understand and generate. This document outlines the complete implementation roadmap.

## Core Design Principles

**🤖 AI-First Design**: Every language feature must be optimized for AI code generation. Uniform S-expression syntax, predictable patterns, and semantic clarity take priority over syntactic convenience.

**🎯 Scalable HTTP Specialization**: Language features specifically designed so that every program naturally handles multiple HTTP requests simultaneously, with JWT authentication and scalable web API development - the most market-valuable use cases.

**⚡ Scalable Performance First**: Every design decision ensures that all programs automatically achieve high performance and scalability, naturally handling thousands of simultaneous HTTP requests through Go's goroutines.

**🔒 Static Type Safety**: Compile-time type checking prevents runtime errors and enables aggressive optimization through Go transpilation.

**🔗 Go Ecosystem Integration**: Seamless interoperability with existing Go libraries, frameworks, and infrastructure.

**🧩 Functional Programming**: Immutable data structures, pure functions, and functional composition as primary paradigms that AI can reliably generate and that provide automatic thread-safety for all programs.

### AI-Friendly Design Goals

**Uniform Syntax**: S-expressions provide consistent `(operation args...)` structure with no precedence rules or special cases.

**Semantic Clarity**: Function names and operations clearly express intent, making AI generation more reliable.

**Composable Patterns**: Simple building blocks that AI can combine into complex systems.

**Predictable Structure**: Consistent patterns that AI models can learn once and apply everywhere.

**Minimal Cognitive Load**: Fewer syntax rules mean AI can focus on business logic rather than parsing complexity, while immutable-by-default design eliminates all concurrency complexity.

### Automatic Scalability Design Goals

**Thread-Safe by Default**: Immutable data structures make it impossible to have race conditions - all programs safely handle multiple simultaneous requests.

**Automatic Parallelization**: Every HTTP request runs in its own lightweight Go goroutine without any special syntax or configuration.

**Natural Statelessness**: Functional design inherently creates stateless request handling, perfect for horizontal scaling.

**Built-in Non-Blocking**: All I/O operations are naturally non-blocking through Go's runtime and standard libraries.

**Zero-Configuration Scaling**: No pools to configure, no limits to set - programs automatically scale to available resources.

## Phase 1: Core Language Foundation ✅ **COMPLETE**

### Parser and Grammar Foundation ✅ **COMPLETE**

**ANTLR Grammar** ✅ *Complete*
- `S-expressions`: Complete support for `(operation args...)` syntax
- `Arrays`: Basic support for `[element1 element2 ...]` syntax
- `Strings`: UTF-8 string literals with escape sequences
- `Symbols`: Identifiers including namespace syntax (`fmt/Println`)
- `Numbers`: Integer and float literal support
- `Comments`: Line comments starting with `;`

**Parser Integration** ✅ *Complete*
- ANTLR4 generated parser for Go
- AST construction from Vex source
- Error handling and reporting
- Multi-language parser support (Go, Java, Python, etc.)

### Basic Type Support ✅ **BASIC/IN-PROGRESS**

**Primitive Types** ✅ *Basic Support*
- `int`: Integers mapping to Go's `int`
- `string`: UTF-8 strings mapping to Go's `string`
- `symbol`: Identifiers and function names
- `bool`: Boolean values mapping to Go's `bool`

**Collection Types** ✅ *Basic Support*
- `[T]`: Arrays transpiled to `[]interface{}` type
- Basic array literal syntax support
- Collection operations through Go interop

**Type System Architecture** ✅ **HM IMPLEMENTED**
Hindley–Milner (HM) type inference runs post-macro expansion with the following in place:
- Algorithm W core with occur-check; substitutions threaded across arrays, calls, `if`, `do`, records
- Let-polymorphism: generalize at `def`/`defn` (value restriction applied), instantiate at use sites
- Function typing from AST body (no raw-text heuristics)
- Collections: arrays unify element types; maps unify key and value types with precise diagnostics
- Records: nominal typing; constructor attaches nominal type; dedicated `VEX-TYP-REC-NOMINAL` mismatch
- Equality via scheme: `∀a. a -> a -> bool`, strict mismatch diagnostics
- Package-boundary schemes: resolver computes `PkgSchemes`; analyzer enforces exports and types namespaced calls
- Diagnostics: stable codes and CLI propagation (UNDEF, COND, EQ, ARRAY-ELEM, MAP-KEY/VAL, REC-NOMINAL)

### Symbol System Design ✅ **BASIC IMPLEMENTATION**

**Symbol Resolution** ✅ *Basic Implementation*
- Namespace-qualified symbols for Go interop (`fmt/Println`)
- Basic symbol resolution for function calls
- Go function binding through namespace mapping
- Import statement processing and validation
- Symbol table management with scoping support

## Phase 2: Basic Transpilation Engine ✅ **COMPLETE**

### Transpiler Architecture ✅ **COMPLETE**

**Core Transpilation Pipeline** ✅ *Complete*
- Parse Vex source into AST using ANTLR parser
- Basic AST to Go code generation
- Variable definitions and expressions
- Go package structure with main function
- Import management for Go packages
- Arithmetic operations

**Language Constructs** ✅ *Basic Implementation*
- ✅ Variable declarations: `(def x 42)` → `x := 42`
- ✅ S-expression syntax for all language constructs
- ✅ Module import declarations: `(import "fmt")` → `import "fmt"`
- ✅ Go interop syntax: `(fmt/Println "hello")` → `fmt.Println("hello")`
- ✅ Arithmetic expressions: `(+ 1 2)` → `1 + 2`
- ✅ Conditional expressions: `(if condition then else)`
- ✅ Array literals: `[1 2 3]` → `[]interface{}{1, 2, 3}`
- ⏳ Pattern matching expressions for destructuring (planned)
- ⏳ Lambda expressions with capture semantics (planned)

**Current Working Syntax**
```vex
; Variable definitions
(def x 42)
(def message "Hello, World!")

; Import system
(import "fmt")

; Go function calls
(fmt/Println message)

; Arithmetic expressions
(def result (+ (* 10 5) (- 20 5)))

; Arrays
(def numbers [1 2 3 4])

; Conditional expressions
(if (> x 0) (fmt/Println "positive") (fmt/Println "negative"))

; Macro definitions
(macro greet [name] (fmt/Println "Hello" name))

; Function definitions using defn macro
(defn add [x y] (+ x y))
```

**Planned Function Definition Syntax**
```vex
(defn function-name [param1: Type1 param2: Type2] -> ReturnType
  body-expressions)
```

### Module System Architecture ✅ **BASIC IMPLEMENTATION**

**Current Module Support** ✅ *Basic Implementation*
- ✅ Go package imports for basic interoperability
- ✅ Namespace-qualified function calls (`fmt/Println`)
- ✅ Basic symbol resolution with namespace support
- ✅ Import statement parsing and validation

**Advanced Module Features** ⏳ *Planned*
- ⏳ Explicit exports for dependency management (see Package Discovery System below - HIGH PRIORITY)
- ⏳ Circular dependency detection and resolution (see Package Discovery System below - HIGH PRIORITY)

## Phase 3: Advanced Transpiler Architecture ✅ **ENHANCED IMPLEMENTATION**

### Modern Transpiler Architecture ✅ **ENHANCED IMPLEMENTATION**

**Multi-Stage Compilation Pipeline** ✅ *Enhanced*
1. ✅ Parse Vex source into AST using ANTLR parser
2. ✅ Macro registration and expansion phase with full defn macro support
3. ✅ Semantic analysis with symbol table management
4. 🚧 HM type inference baseline (post-macro) — expanding coverage incrementally
5. ✅ Advanced code generation with clean Go output
6. ✅ Sophisticated package structure with proper imports and main function

**Advanced Code Generation Strategy** ✅ **ENHANCED IMPLEMENTATION**
The transpiler generates Go code that:
- ✅ Uses idiomatic variable declarations and expressions
- ✅ Maps core Vex expressions to clean Go syntax
- ✅ Generates proper package structure with main function
- ✅ Implements sophisticated import management for Go packages
- ✅ Handles arithmetic and array operations
- ✅ Supports conditional expressions and control flow
- ✅ Generates function definitions from defn macro
- ⏳ Implements immutable collections (planned)
- ⏳ Generates efficient iteration patterns (planned)

**Memory Management** ✅ **LEVERAGES GO GC**
The transpiler generates code that:
- ✅ Works efficiently with Go's garbage collector
- ✅ Uses Go's built-in memory management
- ✅ Generates memory-efficient code patterns
- ⏳ Will implement object pooling for high-frequency allocations (planned)
- ⏳ Will implement structural sharing for immutable collections (planned)

### Go Interoperability Layer ✅ **COMPREHENSIVE IMPLEMENTATION**

**Function Binding System** ✅ **COMPREHENSIVE SUPPORT**
Mechanism to expose Go functions to Vex code through:
- ✅ Namespace-qualified function calls (`fmt/Println`)
- ✅ Import management with basic module detection
- ✅ Clean function call generation with proper argument handling
- ✅ Access to Go standard library via imports
- ⏳ Error handling integration (planned)
- ⏳ Goroutine management for concurrent operations (planned)

**Standard Library Integration** ✅ **SUPPORTED VIA IMPORTS**
Integration with Go standard library:
- ✅ Import system for Go packages
- ✅ Function call generation with clean syntax
- ✅ Basic module detection for third-party packages
- ✅ Proper package structure generation
- ⏳ HTTP handling through net/http (planned)
- ⏳ JSON processing with encoding/json (planned)
- ⏳ Database operations via database/sql (planned)

### Package Discovery and Module System ✅ **MVP IMPLEMENTED**

**Vex Package Discovery System (MVP)**
Implemented foundational package discovery following Go's directory model:
- **Directory-based package structure**: Each directory represents a package (one package per directory)
- **Import path resolution**: Resolve local Vex packages by directory path first; if unresolved, treat as Go import (arrays and alias pairs supported)
- **Automatic package scanning**: Build dependency graph from entry package; perform topological sort for build order
- **Module root detection**: `vex.pkg` is used to detect the module root by walking up from the entry file
- **CLI integration**: `vex transpile`, `vex run`, and `vex build` automatically include discovered packages

**Circular Dependency Prevention (Enforced)**
- **Static dependency analysis**: Build-time cycle detection is mandatory; cycles cause compilation to fail
- **Dependency graph validation**: Analyze the full graph before code generation
- **Clear error messages**: Report the precise cycle chain with edge file locations when available

**Directory Hierarchy and Namespace Management**
Strict directory-based namespace system - **FOUNDATION FOR LARGE PROJECTS**:
- **Package name inference**: Package names automatically derived from directory names
- **Nested package support**: Multi-level package hierarchies (`utils/http/client`, `services/auth/jwt`)
- **Explicit exports**: `(export [sym1 sym2 ...])` parsed; codegen enforces cross-package access to exported symbols; analyzer enforcement planned
- **Private symbol enforcement**: Compile-time enforcement currently in codegen; analyzer-level checks planned
- **Cross-package visibility**: Controlled access between packages in the same module

**Module Boundary Management (Planned follow-up)**
- **Module root detection**: Introduce `vex.pkg` for module boundaries and import path roots
- **Initialization order**: Deterministic init respecting dependencies
- **Cross-package symbol resolution**: Enhanced rules with explicit exports
- **Namespace collision prevention**: Compile-time detection of naming conflicts

**Integration with Go Module System (Planned)**
- **Go module interoperability**: Maintain seamless Go imports; advanced interop after MVP
- **Mixed-language builds**: Planned unified builds for `.vex` and `.go`
- **Dependency version management**: Align with Go modules after `vex.pkg`
- **Third-party library access**: Continue using Go imports for external libs

## Phase 4: Macro System and Metaprogramming ✅ **COMPREHENSIVE IMPLEMENTATION**

**Advanced Macro Definition and Expansion** ✅ **COMPREHENSIVE IMPLEMENTATION**
A sophisticated macro system has been implemented with:
- ✅ User-defined macro registration using `(macro name [params] body)` syntax
- ✅ Dynamic macro expansion during compilation with full error handling
- ✅ Advanced macro template system with parameter substitution
- ✅ Built-in defn macro for comprehensive function definitions
- ✅ Integration with semantic analysis and symbol table management
- ✅ Full compilation pipeline with macro preprocessing
- ✅ Macro validation and error reporting
- ✅ Support for complex macro bodies and nested macro calls

**Comprehensive Macro Architecture** ✅ **COMPREHENSIVE IMPLEMENTATION**
- **Macro Registry**: ✅ Sophisticated registration and lookup system with validation
- **Macro Collector**: ✅ Advanced registration phase with error handling
- **Macro Expander**: ✅ Robust template expansion with parameter validation
- **Error Handling**: ✅ Comprehensive error reporting for macro issues
- **Integration**: ✅ Seamless integration with transpiler pipeline
- **Testing**: ✅ Extensive test coverage for macro functionality

**Defn Macro Implementation** ✅ **COMPLETE**
The defn macro provides comprehensive function definition capabilities:
- ✅ Function definitions: `(defn add [x y] (+ x y))`
- ✅ Parameter list validation and processing
- ✅ Function body expansion and validation
- ✅ Integration with Go function generation
- ✅ Support for complex function bodies with conditionals
- ✅ Error handling for malformed function definitions

This metaprogramming capability enables sophisticated AI code generation patterns and demonstrates advanced language design concepts through uniform macro syntax.

## Phase 4.5: Transpiler Performance Baseline — HIGH PRIORITY — IMMEDIATE IMPLEMENTATION

### Goals

- Establish a low-overhead baseline for the transpilation pipeline used by CLI (`transpile`, `run`, `build`) and APIs.
- Reduce per-call allocations and avoid repeated I/O and object construction work.
- Improve macro expansion efficiency without sacrificing correctness.

### Tasks

- Transpiler instance reuse
  - Keep a long-lived `VexTranspiler` instance in the CLI and public API when processing multiple files; avoid rebuilding parser/analyzer/codegen and re-loading core macros on every call.
  - Ensure `GetDetectedModules()` and other per-call state are explicitly reset between invocations.

- Core macro caching
  - Make core macro loading a process-wide singleton cache to eliminate repeated `core/core.vx` reads and parsing.
  - Provide an override for tests to force reload when needed.

- Parser/analyzer pooling
  - Introduce `sync.Pool` for ANTLR lexer/parser/token streams and analyzer symbol tables to amortize allocations across calls.

- Macro expansion without re-parsing
  - Replace reconstruct+reparse loops with AST-level parameter substitution during macro expansion.
  - Where parsing is still necessary, use a lightweight S-expression tokenizer instead of a full ANTLR roundtrip.

- String building efficiency
  - Replace string concatenations in macro reconstruction paths with `strings.Builder` to reduce temporary strings and GC pressure.

- Reduce adapter churn
  - Align `Value` and `SymbolTable` interfaces across analysis and codegen to avoid creating wrapper objects per visit.
  - If full alignment is not yet feasible, introduce zero-allocation thin adapters.

- Resolver graph cache
  - Cache discovered package graphs keyed by module root and input file mtimes; skip re-walking stable trees.

- CLI build-cache reuse
  - Use a stable temporary module directory and prefer `go` build cache pathways, especially when only stdlib imports are present, to maximize compile-time reuse for `run`/`build`.

### Expected Impact

- Significant reduction in per-transpile allocations and time, especially in macro-heavy code.
- Fewer filesystem accesses and object constructions per CLI invocation.

### Acceptance Criteria

- Benchmarks (guidelines; adjust with hardware):
  - `internal/transpiler`: `BenchmarkTranspileSimple` < 40µs/op, < 40KB/op, < 600 allocs/op; macro-heavy cases show proportional improvements.
  - `internal/transpiler/macro`: `BenchmarkMacro_ExpandChained` < 3µs/op, < 5KB/op, < 100 allocs/op.
  - `internal/transpiler/packages`: Resolver benchmark shows cached steady-state in sub-10µs/op on repeated runs within the same process.

Note: This phase executes before Phase 5. Subsequent phase numbers reflect execution order.

## Phase 5: Immutable Data Structures ⏳ **PLANNED**

### Persistent Collection Implementation

**List Implementation**
Implement persistent vectors using bit-partitioned vector tries for:
- O(log32 N) access, update, and append operations
- Structural sharing to minimize memory overhead
- Efficient iteration and transformation operations
- Thread-safe access patterns for concurrent use

**Map Implementation**
Implement hash array mapped tries (HAMT) for persistent maps with:
- O(log32 N) lookup, insertion, and deletion
- Structural sharing for memory efficiency
- Ordered iteration support for deterministic behavior
- Integration with Go's map syntax where possible

**Optimization Strategies**
- Use transients for bulk operations to improve performance
- Implement lazy evaluation for collection transformations
- Cache hash codes for map keys to avoid recomputation
- Use bit manipulation tricks for efficient tree traversal

## Phase 6: Concurrency and Backend Service Features ⏳ **PLANNED**

### Concurrency Model

**Goroutine Integration**
Design Vex concurrency primitives that map to Go goroutines:
- Lightweight process spawning with isolated state
- Channel-based communication for message passing
- Actor-like patterns for stateful services
- Supervisor hierarchies for fault tolerance

**State Management**
Implement immutable state management patterns:
- Software transactional memory for coordinated updates
- Atomic reference types for lock-free programming
- Event sourcing support for audit trails
- CQRS pattern implementation for read/write separation

### HTTP Service Framework

**Request Handling Pipeline**
Built-in support for HTTP service development:
- Middleware composition for cross-cutting concerns
- Route definition with pattern matching
- Request/response transformation functions
- Automatic JSON serialization/deserialization

**Performance Optimizations**
- Request pooling to reduce allocation overhead
- Response caching with TTL support
- Connection pooling for database operations
- Metrics collection for observability

## Phase 7: Development Tooling and Developer Experience ⏳ **PLANNED**

### Compiler Implementation

**Error Reporting System** ✅ **BASELINE IMPLEMENTED**
Error messages now follow a Go-style, AI-friendly standard with stable error codes and structured details:
- Format: `file:line:col: error: [CODE]: short-message`
- Optional lines: `Expected: …`, `Got: …`, `Suggestion: …`, and location details (e.g., first mismatch index)
- Examples include `VEX-TYP-IF-MISMATCH`, `VEX-ARI-ARGS`, `VEX-TYP-ARRAY-ELEM`, `VEX-TYP-MAP-KEY`, `VEX-TYP-MAP-VAL`
Further work:
- Add machine-readable output flag for structured diagnostics
- Extend coverage beyond analyzer to package resolver and codegen diagnostics

**Incremental Compilation**
Fast development cycle through:
- Module-level compilation caching
- Dependency graph analysis for minimal recompilation
- Hot code reloading for development servers
- Integration with Go's build tools

### IDE and Editor Support

**Language Server Protocol Implementation**
Full LSP server providing:
- Syntax highlighting with semantic tokens
- Auto-completion based on type information
- Go-to-definition across module boundaries
- Real-time error reporting and type hints

**Debugging Support**
- Source map generation for debugging transpiled Go code
- REPL implementation for interactive development

### Testing Framework and Infrastructure ⏳ **PLANNED**

**Native Vex Testing Framework**
Built-in testing capabilities that integrate seamlessly with the language:
- `(deftest test-name "description" body)` macro for test definitions
- `(assert-eq expected actual)` and assertion macros for validations
- `(test-group "group-name" tests...)` for organizing related tests
- Automatic test discovery and execution through `vex test` command

**Property-Based Testing Support**
AI-friendly generative testing:
- `(defproperty prop-name [generators] body)` for property definitions
- Built-in generators for primitive types, collections, and custom types
- Shrinking capabilities to find minimal failing cases
- Integration with Go's testing.T for seamless CI/CD workflows

**HTTP Service Testing**
Specialized testing for backend services:
- `(defservice-test service-name endpoint-tests...)` for API testing
- Built-in HTTP client with authentication support
- Mock server capabilities for dependency isolation
- Load testing primitives for performance validation

**Macro and Transpilation Testing**
Developer tooling for language extension:
- `(defmacro-test macro-name input expected-expansion)` for macro validation
- Transpilation output testing to ensure correct Go code generation
- Performance regression testing for transpiler optimizations
- Integration tests for complete Vex-to-Go-to-binary pipeline

**Test Execution and Reporting**
Comprehensive test runner with detailed feedback:
- Parallel test execution leveraging Go's goroutines
- Detailed failure reporting with source location mapping
- Code coverage analysis for both Vex source and generated Go
- Integration with CI/CD pipelines and standard test formats
- Watch mode for automatic re-testing during development

**AI-Assisted Test Generation**
Leveraging AI for comprehensive test coverage:
- Automatic test case generation from function signatures
- Edge case discovery through static analysis
- Test data generation based on type constraints
- Regression test creation from bug reports and fixes
- Stack trace mapping from Go back to Vex
- Variable inspection with type information

## Phase 8: Standard Library and Ecosystem ⏳ **PLANNED**

### Core Standard Library

**Essential Functions**
Comprehensive standard library covering:
- Collection manipulation functions (map, filter, reduce, etc.)
- String processing and regular expressions
- Mathematical operations and number formatting
- Date/time handling with timezone support
- File system operations and path manipulation

**HTTP and Web Services**
Specialized libraries for backend development:
- HTTP client with connection pooling
- WebSocket support for real-time communication
- Template engines for response generation
- Authentication and authorization helpers

### Package Management

**Dependency Management System**
- Module versioning with semantic versioning support
- Dependency resolution with conflict detection
- Integration with Go modules for Go library dependencies
- Package registry for sharing Vex libraries

## Phase 9: Performance Optimization and Production Readiness ⏳ **PLANNED**

### Compilation Optimizations

**Code Generation Improvements**
- Inline small functions to reduce call overhead
- Eliminate unnecessary allocations through escape analysis
- Generate specialized functions for common type combinations
- Optimize tail-recursive functions into loops

**Benchmarking and Profiling**
- Built-in benchmarking framework for performance testing
- Integration with Go's pprof for profiling transpiled code
- Memory usage analysis and optimization suggestions
- Concurrent load testing utilities

Note: Baseline transpiler performance work (instance reuse, macro caching, pooling, macro expansion without re-parsing, builder usage, adapter reductions, resolver/CLI cache improvements) has been elevated to Phase 4.5. Phase 9 focuses on advanced codegen optimizations and production readiness.

### Production Deployment

**Observability Integration**
- Structured logging with configurable levels
- Metrics collection compatible with Prometheus
- Distributed tracing support for microservices
- Health check endpoints for load balancers

**Security Considerations**
- Input validation and sanitization helpers
- SQL injection prevention in database queries
- Cross-site scripting protection for web responses
- Rate limiting and request throttling utilities

## Success Criteria

**Performance Targets**
- HTTP request handling latency within 10% of equivalent Go code
- Memory usage comparable to idiomatic Go implementations
- Successful deployment in production handling 10K+ requests per second
- Compilation time under 500ms for medium-sized projects

**Developer Experience Metrics**
- Complete IDE support with LSP implementation
- Comprehensive error messages with suggested fixes
- Documentation coverage above 90% for standard library
- Tutorial and example coverage for common use cases

**Ecosystem Integration**
- Seamless integration with existing Go libraries
- Database driver compatibility with major databases
- Cloud platform deployment support (AWS, GCP, Azure)
- Container orchestration with Docker and Kubernetes

This implementation roadmap provides a clear path to creating a production-ready functional programming language optimized for backend services while maintaining the performance characteristics and ecosystem benefits of Go.
