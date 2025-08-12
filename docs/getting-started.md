# Getting Started with Vex

## Quick Start Guide

Vex is a functional programming language that transpiles to Go, designed specifically for AI code generation and concurrent HTTP services. With a complete macro system, comprehensive Go interoperability, and Hindley-Milner type inference, Vex provides a clean, predictable syntax that both humans and AI can understand. This guide will get you up and running in minutes.

## Prerequisites

- **Go 1.21+** - For building and running the transpiler
- **Git** - For cloning the repository
- **Basic terminal knowledge** - For running commands

## Installation

### 1. Clone the Repository
```bash
git clone https://github.com/thsfranca/vex.git
cd vex
```

### 2. Build the Transpiler
```bash
# Build the Vex transpiler
go build -o vex cmd/vex-transpiler/main.go
```

### 3. Verify Installation
```bash
./vex --help
```

You should see the available commands: `transpile`, `run`, `build`, and `test`.

The Vex transpiler includes:
- **Complete macro system** with comprehensive stdlib packages (core, test, collections, conditions, flow, threading, bindings)
- **Full Go interoperability** for calling standard library functions via `package/Function` syntax with aliases
- **Advanced code generation** that produces optimized Go code with complete type checking
- **Complete package system**: directory-based packages, `vex.pkg` module root detection, circular dependency detection, export enforcement
- **Complete export system**: private-by-default with `(export [name1 name2 ...])` enforcing cross-package access
- **Hindley-Milner type system** with Algorithm W inference, unification, and comprehensive error reporting
- **Complete testing framework**: test discovery, coverage analysis, CI/CD integration

## Your First Vex Program

### 1. Create a Hello World Program
Create a file called `hello.vx`:

```vex
;; Import Go's fmt package
(import "fmt")

;; Define a greeting message
(def greeting "Hello, Vex!")

;; Print the greeting
(fmt/Println greeting)
```

### 2. Run Your Program
```bash
# Compile and execute directly
./vex run -input hello.vx
```

Output:
```
Hello, Vex!
```

### 3. See the Generated Go Code
```bash
# Transpile to Go for inspection
./vex transpile -input hello.vx -output hello.go
cat hello.go
```

Output:
```go
package main

import "fmt"

func main() {
    greeting := "Hello, Vex!"
    fmt.Println(greeting)
}
```

## Core Language Concepts

### S-Expression Syntax
Everything in Vex follows the pattern: `(operation arguments...)`

```vex
;; Arithmetic
(+ 1 2 3)           ; Addition: 1 + 2 + 3
(* (+ 2 3) 4)       ; Nested: (2 + 3) * 4

;; Variable definition
(def name value)    ; Create variable

;; Function calls
(fmt/Println "Hi")  ; Call Go function
```

### Variables and Values
```vex
;; Numbers
(def age 25)
(def price 19.99)

;; Strings
(def name "Alice")
(def message "Welcome to Vex!")

;; Booleans
(def is-active true)
(def is-complete false)

;; Arrays
(def numbers [1 2 3 4 5])
(def names ["Alice" "Bob" "Charlie"])
```

### Function Definitions
```vex
;; Define a function using the defn macro with explicit types (current requirement)
(defn greet [name: string] -> string
  (fmt/Sprintf "Hello, %s!" name))

;; Define a function with multiple parameters
(defn add [x: number y: number] -> number
  (+ x y))

;; Define a function with complex logic
(defn absolute [x: number] -> number
  (if (< x 0) (- x) x))

;; Call the functions
(def greeting (greet "World"))
(def result (add 5 3))
(def positive (absolute -42))
(fmt/Println greeting result positive)
```

**Naming Convention**: Vex enforces kebab-case (dash-separated) naming for **all symbols** including variables, functions, macros, and records. Always use dashes instead of underscores:
- ✅ Variables: `user-name`, `api-key`, `is-active`
- ✅ Functions: `calculate-total`, `process-data`, `is-valid`
- ✅ Records: `user-profile`, `order-item`
- ❌ Invalid: `user_name`, `calculate_total`, `user_profile`

### Conditional Logic
```vex
(def score 85)

(if (> score 90)
    (fmt/Println "Excellent!")
    (fmt/Println "Good job!"))
```

### Working with Go Libraries
```vex
;; Import any Go package
(import "strings")
(import "time")

;; Use Go functions with namespace syntax
(def upper-name (strings/ToUpper "alice"))
(fmt/Println upper-name)  ; Prints: ALICE
```

## Example Programs

### 1. Simple Calculator
```vex
;; calculator.vx
(import "fmt")

(defn add [x y] (+ x y))
(defn multiply [x y] (* x y))

(def a 10)
(def b 5)

(fmt/Printf "%d + %d = %d\n" a b (add a b))
(fmt/Printf "%d * %d = %d\n" a b (multiply a b))
```

Run with: `./vex run -input calculator.vx`

### 2. Working with Arrays and Collections
```vex
;; arrays.vx
(import "fmt")

(def fruits ["apple" "banana" "orange"])
(def numbers [1 2 3 4 5])

;; Use stdlib collection functions
(def first-fruit (first fruits))          ; requires stdlib/vex/collections
(def rest-numbers (rest numbers))         ; requires stdlib/vex/collections
(def fruit-count (count fruits))          ; requires stdlib/vex/collections

(fmt/Println "Fruits:" fruits)
(fmt/Println "First fruit:" first-fruit)
(fmt/Println "Numbers:" numbers)
(fmt/Println "Count:" fruit-count)
```

### 3. Advanced Function Definitions with Types
```vex
;; advanced.vx
(import "fmt")

;; Define a function that uses other functions with explicit types
(defn square [x: number] -> number (* x x))
(defn sum-of-squares [a: number b: number] -> number 
  (+ (square a) (square b)))

;; Define a function with conditional logic
(defn max [a: number b: number] -> number
  (if (> a b) a b))

;; Use the functions
(def result (sum-of-squares 3 4))  ; 25
(def larger (max 10 5))            ; 10
(fmt/Printf "Result: %d, Larger: %d\n" result larger)
```

### 4. Custom Macros
```vex
;; macros.vx
(import "fmt")

;; Define a logging macro
(macro log [message]
  (fmt/Printf "[LOG] %s\n" message))

;; Define a debugging macro
(macro debug [var]
  (fmt/Printf "[DEBUG] %s = %v\n" var var))

;; Use the macros
(log "Application started")
(def x 42)
(debug x)
```

### 5. Records and Package System
```vex
;; records.vx
(import "fmt")

;; Declare a record schema (analyzer complete, validates fields and types)
(record Person [name: string age: number])
(record Point [x: number y: number])

;; Records pass complete HM type analysis
;; Construction/access patterns defined (codegen in progress)

;; Package system example:
;; In mymath/operations.vx:
;; (export [add multiply])
;; (defn add [x: number y: number] -> number (+ x y))

;; In main.vx:
;; (import ["mymath"])
;; (def result (mymath/add 10 20))
```

## Available Commands

### `vex run`
Compile and execute Vex programs directly:
```bash
./vex run -input program.vx
```

Notes:
- Automatically discovers local packages starting from the entry file
- Uses `vex.pkg` (if present) to detect the module root
- Enforces exports when calling across local packages
- Complete HM type checking with structured diagnostics
- Supports package aliases: `(import [["mypackage" mp]])`

### `vex transpile`
Convert Vex to Go source code:
```bash
./vex transpile -input program.vx -output program.go
```

### `vex build`
Create standalone binary executables:
```bash
./vex build -input program.vx -output my-program
./my-program  # Run the binary
```

Notes:
- Discovers and includes local packages with dependency resolution
- Generates a temporary `go.mod` and runs `go mod tidy` when external Go modules are detected
- Enforces exports when calling across local packages
- Complete HM type checking and validation
- Optimized Go code generation with proper module management

### `vex test`
Advanced testing framework with real execution-based coverage analysis:
```bash
# Basic testing with file-level coverage
./vex test -dir . -coverage -verbose

# Enhanced coverage with real execution data analysis  
./vex test -enhanced-coverage -coverage-out coverage.json
```

- **Discovery**: Recursively finds `*_test.vx` files
- **Macros**: Uses stdlib test macros `assert-eq`, `assert-true`, `assert-false`, and `deftest`
- **Basic Coverage**: Per-package file-level coverage reporting with `-coverage`
- **Enhanced Coverage**: Real execution-based analysis with `-enhanced-coverage`:
  - **Execution-Based Coverage**: Uses Go runtime instrumentation for 100% accurate data
  - **Real Coverage Data**: Parses actual Go coverage profiles from test execution
  - **Precision Reporting**: Shows "REAL execution data ✅" with accurate percentages
  - **Failure Handling**: Reports "No coverage data available" when tests fail
- **Output**: Detailed test results with real execution coverage metrics
- **CI/CD Ready**: JSON coverage exports based on actual test execution
- **Exit Codes**: Non-zero if any test fails

#### Test File Example (`calculator_test.vx`):
```vex
;; Import required modules for testing
(import ["fmt" "test"])

;; Only deftest blocks are allowed in test files
(deftest "basic-arithmetic"
  (do
    (fmt/Println "Testing basic arithmetic")
    (def result (+ 2 3))
    (assert-eq result 5 "addition works")
    (def product (* 4 5))
    (assert-eq product 20 "multiplication works")))

(deftest "edge-cases"
  (do
    (fmt/Println "Testing edge cases")
    (assert-eq (+ 0 5) 5 "adding zero")
    (assert-eq (* 0 5) 0 "multiplying by zero")
    (assert-true (> 5 0) "positive number check")))

;; Invalid: code outside deftest blocks will fail validation
;; (fmt/Println "This would cause test validation to fail")
```

## Development Workflow

### 1. Write-Test-Run Cycle
```bash
# Edit your .vx file
vim my-program.vx

# Write corresponding tests
vim my-program_test.vx

# Run tests with real execution coverage
./vex test -enhanced-coverage -verbose

# Test quickly during development
./vex run -input my-program.vx

# Check generated Go (optional)
./vex transpile -input my-program.vx -output debug.go
```

### 2. Building for Distribution
```bash
# Create optimized binary
./vex build -input my-app.vx -output my-app

# Distribute the binary (no dependencies needed)
./my-app
```

## Common Patterns

### Error Handling (using Go patterns)
```vex
(import "fmt")
(import "strconv")

;; Use Go's error handling patterns
(def value "123")
(def result (strconv/Atoi value))
;; Handle errors in Go style when needed
```

### Multiple Statements
```vex
;; Use 'do' for sequential execution
(do
  (def name "Alice")
  (def age 30)
  (fmt/Printf "Name: %s, Age: %d\n" name age))
```

### Complex Expressions
```vex
;; Nested operations are natural
(def total (+ 
  (* 10 5)      ; 50
  (- 20 5)      ; 15
  (/ 100 4)))   ; 25
;; total = 90
```

## Next Steps

1. **Explore Examples**: Check out `examples/valid/` for more programs
2. **Read the Grammar**: See `docs/grammar-reference.md` for complete syntax
3. **AI Integration**: Check `docs/ai-quick-reference.md` for AI code generation
4. **Go Integration**: See `examples/go-usage/` for parser usage
5. **Development**: Read `docs/development-guide.md` (if contributing)

## HM Typing Diagnostics (Quick Reference)
Vex uses Hindley–Milner (HM) type inference with clear, code-based diagnostics:

- `VEX-TYP-UNDEF`: Unknown identifier (define or import it)
- `VEX-TYP-COND`: If-condition must be `bool`
- `VEX-TYP-EQ`: Equality arguments have different types
- `VEX-TYP-ARRAY-ELEM`: Array elements have inconsistent types
- `VEX-TYP-MAP-KEY` / `VEX-TYP-MAP-VAL`: Map key/value types inconsistent across pairs
- `VEX-TYP-REC-NOMINAL`: Nominal record mismatch (e.g., `A` used where `B` is required)

Example stderr:

```
path/to/file.vx:10:5: error: [VEX-TYP-UNDEF]: unknown identifier
Name: y
```

## Getting Help

- **Examples**: `examples/valid/` directory
- **Language Reference**: `docs/grammar-reference.md`
- **AI Reference**: `docs/ai-quick-reference.md`
- **Implementation Details**: `docs/vex-implementation-requirements.md`
- **Issues**: [GitHub Issues](https://github.com/thsfranca/vex/issues)

## What's Working Now

✅ **Complete type system** - Full Hindley-Milner inference with Algorithm W, unification, generalization/instantiation  
✅ **Advanced transpiler** - Multi-stage compilation with parse → macro expansion → type analysis → code generation  
✅ **Complete package system** - Directory-based packages, exports, circular dependency detection, `vex.pkg` support  
✅ **Full macro system** - Complete stdlib with 7 packages: core, test, collections, conditions, flow, threading, bindings  
✅ **Advanced CLI** - `transpile`, `run`, `build`, `test` commands with comprehensive options and coverage analysis  
✅ **Testing framework** - Complete test discovery, validation, coverage analysis, CI/CD integration  
✅ **Go interoperability** - Full import system with aliases, comprehensive standard library access  
✅ **Function definitions** - Complete `defn` macro with explicit type annotations (current requirement)  
✅ **Records** - Complete nominal type analysis with record declarations (analyzer complete)  
✅ **Structured diagnostics** - Stable error codes (VEX-TYP-*) with AI-friendly formatting  
✅ **Quality infrastructure** - 85%+ test coverage, automated CI/CD, comprehensive documentation  

## Current Architecture

The Vex transpiler is feature-complete with:

- **Complete compilation pipeline** with parsing, macro expansion, HM type analysis, and optimized code generation
- **Advanced macro system** with comprehensive stdlib and template expansion
- **Complete package system** with dependency resolution, export enforcement, and circular dependency detection
- **Hindley-Milner type system** with Algorithm W inference, unification, and comprehensive error reporting
- **Full testing framework** with discovery, validation, coverage analysis, and CI/CD integration
- **Advanced Go interoperability** with complete import system and optimized code generation

## Coming Soon

⏳ **Performance optimizations** - Explicit stdlib imports, transpiler instance reuse, macro caching  
⏳ **Enhanced coverage analysis** - Function-level tracking, branch coverage, statement-level analysis  
⏳ **Record construction/access** - Complete implementation of record operations  
⏳ **Advanced control flow** - `when`, `unless`, `cond`, pattern matching constructs  
⏳ **HTTP server framework** - Built-in web service capabilities with concurrent request handling  
⏳ **Concurrency primitives** - Goroutine and channel operations with type safety  

Happy coding with Vex! 🚀