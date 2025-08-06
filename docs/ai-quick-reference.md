---
title: "Vex Language - AI Quick Reference"
version: "0.3.0"
compatibility: "Go 1.21+"
last-updated: "2025-01-09"
ai-model-compatibility: "GPT-4, Claude-3+, and similar models"
purpose: "Machine-readable reference for AI code generation"
---

# Vex Language AI Quick Reference

## Core Language Model

**Execution Pipeline**: Parse S-expressions → AST → Macro expansion → Semantic analysis → Go transpilation → Compile & execute  
**Syntax Pattern**: `(operation arg1 arg2 ...)`  
**Output Target**: Go source code  
**Type System**: Basic types with Go interop | Advanced typing planned  
**Concurrency**: Automatic via Go goroutines  
**Macro System**: Comprehensive user-defined macros with defn support  
**Project Status**: Phase 1-4 complete | Phase 5+ planned  

## Language Specification

### Syntax Rules
```ebnf
program     ::= list+ EOF
list        ::= '(' element+ ')'
array       ::= '[' element+ ']'
element     ::= list | array | SYMBOL | STRING | NUMBER
SYMBOL      ::= [a-zA-Z0-9.+\-*\/=!<>?_:~]+
STRING      ::= '"' (~'"' | '\"')* '"'
NUMBER      ::= [0-9]+
comment     ::= ';' .* '\n'
```

### Data Types
| Type | Status | Vex Syntax | Go Output | Notes |
|------|--------|------------|-----------|-------|
| Integer | ✅ | `42` | `42` | 64-bit signed |
| String | ✅ | `"hello"` | `"hello"` | UTF-8 |
| Symbol | ✅ | `variable-name` | `variableName` | Identifiers |
| Array | ✅ | `[1 2 3]` | `[]interface{}{1, 2, 3}` | Generic arrays |
| Boolean | ✅ | `true false` | `true false` | Go booleans |

## Core Operations Reference

### Variable Definition
```
Syntax: (def symbol value)
Example: (def x 42) → x := 42
Status: ✅ IMPLEMENTED
Purpose: Create variables with values
```

### Arithmetic Operations
```
Syntax: (operator arg1 arg2 ...)
Examples: 
  (+ 1 2) → 1 + 2
  (* (+ 1 2) 3) → (1 + 2) * 3
Status: ✅ IMPLEMENTED
Operators: + - * / > < = >= <=
```

### Import System
```
Syntax: (import "package-name")
Example: (import "fmt") → import "fmt"
Status: ✅ IMPLEMENTED
Purpose: Access Go standard library
```

### Go Function Calls
```
Syntax: (namespace/function args...)
Example: (fmt/Println "Hello") → fmt.Println("Hello")
Status: ✅ IMPLEMENTED
Purpose: Call Go functions with namespace syntax
```

### Conditional Expressions
```
Syntax: (if condition then-expr else-expr)
Example: (if (> x 0) "positive" "negative")
Status: ✅ IMPLEMENTED
Purpose: Conditional logic
```

### Sequential Execution
```
Syntax: (do expr1 expr2 ...)
Example: (do (def x 10) (fmt/Println x))
Status: ✅ IMPLEMENTED
Purpose: Execute multiple expressions in sequence
```

### Macro Definitions
```
Syntax: (macro name [params] body)
Example: (macro greet [name] (fmt/Println "Hello" name))
Status: ✅ IMPLEMENTED
Purpose: Code generation and metaprogramming
```

### Function Definitions
```
Syntax: (defn name [params] body)
Example: (defn add [x y] (+ x y))
Status: ✅ IMPLEMENTED (via defn macro)
Purpose: Define reusable functions
```

## AI Code Generation Patterns

### Basic Variable Assignment
```vex
;; Pattern: Simple value assignment
(def variable-name value)

;; Examples:
(def user-id 123)
(def message "Welcome")
(def is-active true)
```

### Function Definition Pattern
```vex
;; Pattern: Function with parameters
(defn function-name [param1 param2] body-expression)

;; Examples:
(defn calculate-total [price tax] (+ price (* price tax)))
(defn greet-user [name] (fmt/Printf "Hello, %s!" name))
(defn is-valid [input] (> (len input) 0))
```

### HTTP Service Pattern (Planned)
```vex
;; Pattern: AI-friendly HTTP endpoint
(defn api-endpoint
  ^{:http-endpoint "/api/path"
    :method "GET"
    :auth-required true}
  [request]
  (-> request validate-input process-data format-response))

;; Pattern: HTTP server
(http-server
  :port 8080
  :routes [
    (GET "/users" list-users)
    (POST "/users" create-user)
  ])
```

### Error Handling Pattern (Planned)
```vex
;; Pattern: Go-style error handling
(let [result err] (risky-operation data)
  (if err
    (handle-error err)
    (process-result result)))
```

## Decision Tree for AI Code Generation

```
Need to store a value?
├─ Simple value → (def name value)
└─ Computed value → (def name (computation))

Need reusable logic?
├─ Simple function → (defn name [args] body)
└─ Code generation → (macro name [args] template)

Need Go library access?
├─ Import first → (import "package")
└─ Call function → (package/function args)

Need conditional behavior?
├─ Simple choice → (if condition then else)
└─ Multiple cases → (cond [(test1 result1) (test2 result2)])

Need to process data?
├─ Single operation → (operation data)
└─ Pipeline → (-> data op1 op2 op3)
```

## Common Patterns & Idioms

### Naming Conventions
```
Variables: kebab-case (user-name, api-key)
Functions: verb-noun (process-data, validate-input)
Predicates: question-suffix (is-valid?, has-data?)
Constants: ALL-CAPS (MAX-RETRIES, DEFAULT-PORT)
```

### Composition Patterns
```vex
;; Sequential execution
(do
  (def data (fetch-data))
  (def processed (process data))
  (save processed))

;; Function composition (planned)
(-> input
    validate
    transform
    save)

;; Conditional processing
(if (valid? data)
    (process data)
    (log-error "Invalid data"))
```

## Implementation Status Matrix

| Feature | Status | Syntax | Example | AI Usage |
|---------|--------|---------|---------|----------|
| **Variables** | ✅ | `(def x v)` | `(def count 0)` | Store values |
| **Arithmetic** | ✅ | `(+ a b)` | `(+ 1 2)` | Math operations |
| **Imports** | ✅ | `(import "pkg")` | `(import "fmt")` | Access Go libs |
| **Go calls** | ✅ | `(ns/fn args)` | `(fmt/Println x)` | Call Go functions |
| **Conditionals** | ✅ | `(if c t e)` | `(if (> x 0) x 0)` | Logic branching |
| **Sequences** | ✅ | `(do e1 e2)` | `(do (def x 1) x)` | Multiple exprs |
| **Macros** | ✅ | `(macro n [p] b)` | `(macro log [m] ...)` | Code generation |
| **Functions** | ✅ | `(defn n [p] b)` | `(defn add [x y] (+ x y))` | Reusable logic |
| **Arrays** | ✅ | `[a b c]` | `[1 2 3]` | Collections |
| **Symbol tables** | ✅ | Variable scoping | Automatic | Proper scoping |
| **Error handling** | ✅ | Parse/compile errors | Comprehensive | Error reporting |
| **Types** | ⏳ | `[x: int]` | Future | Type annotations |
| **HTTP** | ⏳ | `(http-server)` | Future | Web services |
| **Loops** | ⏳ | `(for x in xs)` | Future | Iteration |

## Error Prevention for AI

### Invalid Syntax (Avoid)
```vex
;; ❌ Missing arguments
(def)
(defn add)
(macro)

;; ❌ Mismatched parentheses  
(def x (+ 1 2
(def x (+ 1 2)))

;; ❌ Invalid symbols
(def 123invalid "value")
(def "string-name" 42)
```

### Valid Syntax (Use)
```vex
;; ✅ Complete definitions
(def x 42)
(defn add [x y] (+ x y))
(macro greet [name] (fmt/Println "Hello" name))

;; ✅ Proper nesting
(def result (+ (* 2 3) (- 10 5)))

;; ✅ Valid identifiers
(def user-count 0)
(def is-valid? true)
(def process-data-fn (fn [x] x))
```

## CLI Usage for AI Testing

```bash
# Build transpiler
go build -o vex cmd/vex-transpiler/main.go

# Test AI-generated code
echo '(def x (+ 5 3))' > test.vx
./vex transpile -input test.vx -output test.go
./vex run -input test.vx

# Expected Go output:
# package main
# func main() { x := 5 + 3 }
```

## AI Model Integration Notes

**Best Practices for AI Code Generation**:
1. Always wrap complete expressions in parentheses
2. Use descriptive variable names with kebab-case
3. Prefer function composition over complex nesting
4. Test generated code with `vex run` command
5. Use macros for repetitive code patterns

**Current Capabilities**:
- ✅ Comprehensive function definitions with defn macro
- ✅ Advanced macro system for code generation
- ✅ Complete Go interoperability for library access
- ✅ Symbol table management for proper scoping
- ✅ Sophisticated error reporting and validation
- ✅ Clean code generation producing idiomatic Go

**Current Limitations**:
- No advanced type checking (basic types supported, use Go types via interop)
- No structured error handling (use Go patterns)  
- No loops (use recursion or Go interop)
- No immutable data structures (use Go types)

**Future AI Enhancements** (Planned):
- Advanced type inference and checking
- Package discovery and dependency management
- Semantic annotations for intent-based generation
- Type-aware code completion
- Automatic HTTP endpoint generation
- Structured error handling patterns
- Performance optimization hints

---

**For Human Documentation**: See [README.md](../README.md) and [vex-implementation-requirements.md](vex-implementation-requirements.md)  
**For Examples**: See [examples/valid/](../examples/valid/) directory  
**For Grammar**: See [grammar-reference.md](grammar-reference.md)  
**Repository**: https://github.com/thsfranca/vex