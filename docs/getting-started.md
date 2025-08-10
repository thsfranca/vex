# Getting Started with Vex

## Quick Start Guide

Vex is a functional programming language that transpiles to Go, designed specifically for AI code generation and concurrent HTTP services. With a sophisticated macro system and comprehensive Go interoperability, Vex provides a clean, predictable syntax that both humans and AI can understand. This guide will get you up and running in minutes.

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

You should see the available commands: `transpile`, `run`, and `build`.

The Vex transpiler includes:
- **Macro system** with core macros (including `defn`) loaded from `core/core.vx` (auto-detected)
- **Go interoperability** for calling standard library functions via `package/Function` syntax
- **Code generation** that produces executable Go code with a `main` function
- **Package discovery (MVP)**: local directory packages, `vex.pkg` module root detection, cycle detection
- **Exports (MVP)**: private-by-default; `(export [name1 name2 ...])` required for cross-package access
- **Semantic analysis** with symbol table management and validations

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
;; Define a function using the defn macro
(defn greet [name]
  (fmt/Printf "Hello, %s!\n" name))

;; Define a function with multiple parameters
(defn add [x y]
  (+ x y))

;; Define a function with complex logic
(defn absolute [x]
  (if (< x 0) (- x) x))

;; Call the functions
(greet "World")
(def result (add 5 3))
(def positive (absolute -42))
```

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

### 2. Working with Arrays
```vex
;; arrays.vx
(import "fmt")

(def fruits ["apple" "banana" "orange"])
(def numbers [1 2 3 4 5])

(fmt/Println "Fruits:" fruits)
(fmt/Println "Numbers:" numbers)
```

### 3. Advanced Function Definitions
```vex
;; advanced.vx
(import "fmt")

;; Define a function that uses other functions
(defn square [x] (* x x))
(defn sum-of-squares [a b] (+ (square a) (square b)))

;; Define a function with conditional logic
(defn max [a b]
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
- Discovers and includes local packages
- Generates a temporary `go.mod` and runs `go mod tidy` when external Go modules are detected
- Enforces exports when calling across local packages

## Development Workflow

### 1. Write-Test-Run Cycle
```bash
# Edit your .vx file
vim my-program.vx

# Test quickly
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

## Getting Help

- **Examples**: `examples/valid/` directory
- **Language Reference**: `docs/grammar-reference.md`
- **AI Reference**: `docs/ai-quick-reference.md`
- **Implementation Details**: `docs/vex-implementation-requirements.md`
- **Issues**: [GitHub Issues](https://github.com/thsfranca/vex/issues)

## What's Working Now

✅ **Variables and basic types** - Complete support for integers, strings, booleans, symbols  
✅ **Arithmetic operations** - Full arithmetic with nested expressions  
✅ **Go function calls** - Comprehensive Go interoperability with namespace syntax  
✅ **Import system** - Advanced import management with module detection  
✅ **Sophisticated macro system** - User-defined macros with parameter validation  
✅ **Function definitions** - Complete defn macro with parameter lists and complex bodies  
✅ **Conditional expressions** - if/then/else with proper code generation  
✅ **Arrays** - Array literals and basic operations  
✅ **Symbol table management** - Proper scoping and variable resolution  
✅ **Error handling** - Comprehensive error reporting for parsing and transpilation  
✅ **Advanced code generation** - Clean, idiomatic Go output  

## Current Architecture

The Vex transpiler now includes:

- **Multi-stage compilation pipeline** with parsing, macro expansion, semantic analysis, and code generation
- **Advanced macro system** with template expansion and validation
- **Symbol table management** for proper variable scoping
- **Comprehensive Go interoperability** for seamless library access
- **Clean code generation** producing idiomatic Go output

## Coming Soon

⏳ **Enhanced type system** - Type inference and checking  
⏳ **Package discovery system** - Advanced module management  
⏳ **HTTP server framework** - Built-in web service capabilities  
⏳ **Standard library** - Core functions and utilities  
⏳ **Performance optimizations** - Advanced compiler optimizations  

Happy coding with Vex! 🚀