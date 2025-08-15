---
title: "Vex Language - Quick Reference"
version: "0.4.0"
compatibility: "Go 1.21+"
purpose: "Comprehensive language reference for developers"
---

# Vex Language Quick Reference

## Core Language Model

**Data Engineering Focus**: Specialized for ETL pipelines, stream processing, and real-time analytics with functional programming excellence  
**Execution Pipeline**: Parse S-expressions → AST → Macro expansion → HM type inference → Semantic analysis → Go code generation → Compile & execute  
**Syntax Pattern**: `(operation arg1 arg2 ...)` - optimized for data transformation readability  
**Output Target**: High-performance Go code for data processing workloads with proper imports and main function  
**Type System**: Complete Hindley-Milner Algorithm W ensures type-safe data transformations and prevents runtime pipeline failures  
**Package System**: Directory-based packages with automatic discovery, perfect for organizing data transformation libraries and stream processing modules  
**Performance**: Excellent performance via Go's runtime - essential for high-throughput data processing and real-time analytics  
**Macro System**: Advanced user-defined macros enable clean data pipeline definitions and transformation patterns  
**Diagnostics**: Structured error codes (VEX-TYP-*) prevent data pipeline errors before deployment  
**Data Engineering Philosophy**: Functional programming paradigm ensures immutable data transformations, making ETL pipelines reliable and debuggable  
**Project Status**: Complete foundation ready for data engineering specialization; stream processing and real-time analytics capabilities planned  

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
| Integer | ✅ | `42` | `42` | 64-bit signed with type inference |
| String | ✅ | `"hello"` | `"hello"` | UTF-8 with type checking |
| Boolean | ✅ | `true false` | `true false` | Go booleans with strict typing |
| Symbol | ✅ | `variable-name` | `variableName` | Identifiers with scope resolution |
| Array | ✅ | `[1 2 3]` | `[]interface{}{1, 2, 3}` | Element type unification |
| Map | ✅ | `{:key value}` | Go map literals | Key/value type unification |
| Record | ✅ | `(record Name [field: Type ...])` | Nominal type validation | Analyzer complete, codegen in progress |
| Function | ⏳ | `(fn [params] body)` | Go function types | Type inference from body |

## Core Operations Reference

### Variable Definition
```
Syntax: (def symbol value)
Example: (def x 42) → x := 42
Status: ✅ IMPLEMENTED
Purpose: Create variables with values
```

### Arithmetic & Comparison Operations
```
Syntax: (operator arg1 arg2 ...)
Examples: 
  (+ 1 2) → 1 + 2
  (* (+ 1 2) 3) → (1 + 2) * 3
Status: ✅ IMPLEMENTED
Operators: + - * / > < =
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
Status: ⏳ MACRO IMPLEMENTED (codegen improvements needed)
Purpose: Define reusable functions with type inference
```

### Record Declarations
```
Syntax: (record Name [field: Type ...])
Example: (record Person [name: string age: number])
Status: ✅ ANALYZER COMPLETE (nominal typing, validation)
Purpose: Define structured data with named fields and nominal typing
```

### Package System
```
Syntax: (import ["local-package" "fmt"])
        (export [function-name other-symbol])
Example: (import ["utils" ["net/http" http]])
         (export [add multiply])
Status: ✅ IMPLEMENTED (discovery, resolution, exports)
Purpose: Modular code organization with dependency management
```

## Data Engineering Patterns

### ETL Pipeline Patterns
```vex
;; Data transformation pipeline
(defn process-customer-events [events]
  (-> events
      (filter valid-event?)
      (map enrich-with-geo-data)
      (aggregate-by :customer-id)
      (emit-to data-warehouse)))

;; Stream processing with windowing
(defn user-session-analytics [click-stream]
  (-> click-stream
      (window-tumbling (minutes 5))
      (aggregate session-metrics)
      (filter anomaly-threshold?)
      (real-time-alert)))
```

### Real-time Analytics
```vex
;; Fraud detection pipeline
(defn detect-fraud [transaction-stream]
  (-> transaction-stream
      (window-by :account-id (minutes 5))
      (aggregate suspicious-patterns)
      (score-risk-level)
      (alert-if-high-risk)))

;; Real-time dashboard metrics
(defanalytics user-engagement
  :metrics [active-users conversion-rate page-views]
  :refresh-rate (seconds 5)
  :alerts [(threshold :conversion-rate < 0.03)])
```

## Common Programming Patterns

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

### Data Processing Pattern (Planned)
```vex
;; Pattern: AI-friendly data pipeline
(defn process-dataset
  ^{:data-pipeline "transform"
    :input-type "csv"
    :output-type "json"}
  [data]
  (-> data validate-input transform-data format-output))

;; Pattern: Data processing workflow
(data-pipeline
  :input "data.csv"
  :transformations [
    (filter valid-record?)
    (map extract-features)
    (reduce aggregate-results)
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

## Decision Tree for Code Development

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

Need to handle collections?
├─ Transform → (map function collection)
└─ Filter → (filter predicate collection)
```

## Common Patterns & Idioms

### Naming Conventions
```
All symbols: kebab-case (user-name, api-key, process-data)
Functions: verb-noun kebab-case (process-data, validate-input)
Predicates: question-suffix kebab-case (is-valid?, has-data?)
Records: noun kebab-case (user-profile, order-item)
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
| **Arrays** | ✅ | `[a b c]` | `[]interface{}{1, 2, 3}` | Collections |
| **Symbol tables** | ✅ | Variable scoping | Automatic | Proper scoping |
| **Error handling** | ✅ | Parse/compile errors | Comprehensive | Error reporting |
| **Types** | ⏳ | `[x: int]` | Future | Type annotations |
| **Data Processing** | ⏳ | `(data-pipeline)` | Future | Data transformation |
| **Loops** | ⏳ | `(for x in xs)` | Future | Iteration |

## HM Typing Diagnostics

Common diagnostic codes to detect and fix:

- `VEX-TYP-UNDEF`: Unknown identifier
- `VEX-TYP-COND`: If-condition must be `bool`
- `VEX-TYP-EQ`: Equality argument types differ
- `VEX-TYP-ARRAY-ELEM`: Array elements mismatch
- `VEX-TYP-MAP-KEY` / `VEX-TYP-MAP-VAL`: Map key/value types mismatch across pairs
- `VEX-TYP-REC-NOMINAL`: Nominal record mismatch (A vs B)

## Test Message Standards

### Test Definition Patterns
```vex
;; Pattern: "subject-action-expectation" (kebab-case)
(deftest "user-authentication-validates-credentials" ...)
(deftest "payment-processing-handles-invalid-cards" ...)
(deftest "api-endpoint-returns-proper-json-structure" ...)

;; Assertion pattern: "action-context-expectation" (kebab-case)
(assert-eq (validate-email "user@test.com") true "valid-email-format-passes-validation")
(assert-eq (calculate-tax 100.00) 8.50 "standard-tax-rate-applies-correctly")
(assert-eq (status-code response) 200 "get-users-endpoint-returns-success-status")
```

### Framework Output (Clean Text)
```
Running test: user-authentication-validates-credentials
PASS: valid-email-format-passes-validation
FAIL: invalid-credentials-fail-authentication
  Expected: true
  Actual: false
```

> **📖 Reference**: See [test-message-standards.md](test-message-standards.md) for complete standards

## Error Prevention

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
;; ❌ Invalid record fields
(record Person [name string])    ; missing ':'
(record Person [name:])          ; missing type
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

## CLI Usage

```bash
# Build transpiler
go build -o vex cmd/vex-transpiler/main.go

# Test Vex code
echo '(def x (+ 5 3))' > test.vx
./vex transpile -input test.vx -output test.go
./vex run -input test.vx

# Expected Go output:
# package main
# func main() { x := 5 + 3 }

# Experimental: records (analyzer validates; generation pending)
echo '(record Person [name: string age: number])' > rec.vx
./vex run -input rec.vx
```

## Development Best Practices

**Best Practices for Vex Development**:
1. Always wrap complete expressions in parentheses
2. Use descriptive symbol names with kebab-case (all variables, functions, records)
3. Prefer function composition over complex nesting
4. Test code with `vex run` command
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

**Future Enhancements** (Planned):
- Advanced type inference and checking
- Package discovery and dependency management
- Enhanced IDE support with language server
- Type-aware code completion and IntelliSense
- Advanced data processing patterns
- Structured error handling patterns
- Performance optimization hints

**Codebase Developer Experience** (Phase 4.6 - Planned):
- Domain-driven package organization for clearer navigation
- Rich domain types (`SourceCode`, `GoCode`, `SymbolName`) replacing generic strings
- Simplified interfaces with unified compilation results
- Enhanced error handling with structured diagnostics and builder patterns
- Functional programming patterns aligned with Vex philosophy
- Comprehensive documentation with examples for all APIs

---

**For Human Documentation**: See [README.md](../README.md) and [vex-implementation-requirements.md](vex-implementation-requirements.md)  
**For Examples**: See [examples/valid/](../examples/valid/) directory  
**For Grammar**: See [grammar-reference.md](grammar-reference.md)  
**Repository**: https://github.com/thsfranca/vex