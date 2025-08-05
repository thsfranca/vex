# Vex Grammar Reference

## Overview

The Vex grammar defines a Lisp-like programming language with support for S-expressions, arrays, symbols, and strings.

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

### Valid Programs

#### Working Today (Full Transpiler Support)
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

; Conditional expressions (planned)
(if (> x 0) (print "positive") (print "negative"))

; Arrays (grammar ready)
[1 2 3 4]
["hello" "world"]

; HTTP server patterns (planned)
(http-server
  :port 8080
  :routes [(GET "/api/users" get-users)])
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
