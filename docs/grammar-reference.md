# Flux Grammar Reference

## Overview

The Flux grammar defines a Lisp-like programming language with support for S-expressions, arrays, symbols, and strings.

## Grammar Rules

### Root Rule
```antlr
sp: list+ EOF ;
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
SYMBOL: (LETTER | INTEGER | '.')+ ;
```
Symbols can contain letters, numbers, and dots.

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

```lisp
; Simple list
(hello world)

; Function call with arguments
(print "Hello, World!")

; Nested expressions
(if (> x 0) (print "positive") (print "negative"))

; Arrays
[1 2 3 4]
["hello" "world"]

; Mixed structures
(function [arg1 arg2] (body goes here))
```

### Language Features Supported by Grammar

- **S-expressions**: `(operator operand1 operand2 ...)`
- **Array literals**: `[element1 element2 ...]`
- **String literals**: `"text with spaces"`
- **Symbols/Identifiers**: `variable-name`, `+`, `function123`
- **Comments**: `; comment text`
- **Nested structures**: Lists and arrays can contain other lists and arrays

## Usage with ANTLR

To generate a parser for your target language:

```bash
# For Java
antlr4 -Dlanguage=Java Flux.g4

# For Python
antlr4 -Dlanguage=Python3 Flux.g4

# For Go
antlr4 -Dlanguage=Go Flux.g4
```

Or use the provided Makefile:

```bash
make java     # Generate Java parser
make python   # Generate Python parser
make go       # Generate Go parser
```