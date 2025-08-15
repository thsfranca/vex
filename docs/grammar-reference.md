# Vex Grammar Reference

## Overview

The Vex grammar defines a functional programming language specialized for data engineering, with S-expressions optimized for ETL pipelines, stream processing, and real-time analytics. The syntax supports data transformation patterns, windowing operations, and pipeline definitions.

## Grammar Rules

### Root Rule
```antlr
program: list+ EOF ;
```
A program consists of one or more lists followed by end-of-file.

### List Rule
```antlr
list: '(' (array | list | SYMBOL | STRING)+ ')' ;
```
A list is surrounded by parentheses and contains one or more elements.

### Array Rule
```antlr
array: '[' (array | list| SYMBOL | STRING)+ ']' ;
```
An array is surrounded by square brackets and contains one or more elements.

### Tokens

#### SYMBOL
```antlr
SYMBOL: (LETTER | INTEGER | '.' | '+' | '-' | '*' | '/' | '=' | '!' | '<' | '>' | '?' | '_' | ':' | '~')+ ;
```
Symbols can contain letters, numbers, dots, and various operators including arithmetic operators (+, -, *, /), comparison operators (=, !, <, >), namespace separators (:), and other special characters (?, _, ~).

**Naming Convention**: Vex enforces kebab-case (dash-separated) naming for **all symbols** including variables, functions, macros, and record names. This promotes consistency and readability. Examples:
- ✅ Valid: `my-variable`, `calculate-area`, `is-positive`, `user-data`
- ❌ Invalid: `my_variable`, `calculate_area`, `is_positive`, `user_data` (error code: SYMBOL-NAMING)

#### STRING
```antlr
STRING: '"' ( ~'"' | '\\' '"' )* '"' ;
```
Strings are surrounded by double quotes and support escaped quotes.

#### LETTER
```antlr
LETTER: [a-zA-Z] ;
```
Letters are standard ASCII characters.

#### INTEGER
```antlr
INTEGER: [0-9] ;
```
Single digits (can be combined in symbols).

### Comments and Whitespace

Comments start with `;` and continue to end of line.
Whitespace (spaces, tabs, newlines) and commas are ignored.

## Examples

### Data Engineering Examples

#### ETL Pipeline Syntax
```vex
;; Data transformation pipeline
(defn process-events [events]
  (-> events
      (filter valid-event?)
      (map enrich-with-data)
      (aggregate-by :user-id)
      (emit-to warehouse)))

;; Stream processing with windowing
(defstream user-clicks
  :source (kafka "clicks")
  :window (minutes 5)
  :processing-type :tumbling)
```

#### Real-time Analytics Patterns
```vex
;; Pipeline definition with sources and sinks
(defpipeline fraud-detection
  :sources [(database "transactions") (stream "user-activity")]
  :transforms [detect-anomalies calculate-risk-score]
  :sinks [(alerts "security") (database "risk-analysis")])

;; Complex event processing
(defpattern suspicious-login
  [:failed-login :location-change :password-reset]
  :within (minutes 10)
  :action (alert security-team))
```

### Valid Programs

#### Working Today (Transpiler Support)
```lisp
; Variable definitions
(def x 42)
(def message "Hello, World!")

; Arithmetic expressions  
(def sum (+ 10 20))
(def product (* 6 7))
(def result (+ x (- 100 5)))

; Import system
(import "fmt")
(import "strings")

; Go function calls with namespace syntax
(fmt/Println "Hello from Vex!")
(fmt/Printf "Value: %d\n" x)

; Conditional expressions
(if (> x 0) (fmt/Println "positive") (fmt/Println "negative"))

; Complex expressions
(def calculation (+ (* 10 5) (- 20 (/ 100 5))))

; Macro definitions
(macro greet [name] (fmt/Println "Hello" name))

; Function definitions using defn macro
(defn add [x y] (+ x y))
```

#### Planned Features (Grammar Ready)
```lisp
; Enhanced function definitions with types (in development)
(defn add [x: int y: int] -> int (+ x y))

; Enhanced typed function syntax (planned)
(defn add [x: int y: int] -> int (+ x y))

; Data processing patterns (planned)
(data-pipeline
  :input "data.csv"
  :transformations [(filter valid?) (map transform)])

; Data processing patterns (planned)
(data-pipeline
  :input "data.csv"
  :transformations [(filter valid?) (map transform)])
```

### Language Features Supported by Grammar

- **S-expressions**: `(operator operand1 operand2 ...)`
- **Array literals**: `[element1 element2 ...]` 
- **String literals**: `"text with spaces"`
- **Symbols/Identifiers**: `variable-name`, `+`, `function123`, `fmt/Println`
- **Comments**: `; comment text`
- **Nested structures**: Lists and arrays can contain other lists and arrays
- **Namespace syntax**: `namespace/function` for Go interop
- **Import statements**: `(import "package-name")`
- **Conditional expressions**: `(if condition then else)`
- **Complex arithmetic**: Nested mathematical expressions
- **Macro definitions**: `(macro name [params] body)` for metaprogramming
- **Function definitions**: `(defn name [params] body)` via built-in macros

## Usage with ANTLR

To generate a parser for your target language:

```bash
# For Java
antlr4 -Dlanguage=Java Vex.g4

# For Python
antlr4 -Dlanguage=Python3 Vex.g4

# For Go
antlr4 -Dlanguage=Go Vex.g4
```

Or use the provided Makefile:

```bash
make java     # Generate Java parser
make python   # Generate Python parser
make go       # Generate Go parser
```
