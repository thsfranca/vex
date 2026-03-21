# Vex Language Design Specification

**Version:** 0.1.0-draft
**Status:** Early Design Phase
**Date:** 2026-03-21

---

## Table of Contents

1. [Purpose and Vision](#1-purpose-and-vision)
2. [Design Principles](#2-design-principles)
3. [MCP Server Landscape Analysis](#3-mcp-server-landscape-analysis)
4. [Language Features Driven by MCP Needs](#4-language-features-driven-by-mcp-needs)
5. [Type System](#5-type-system)
6. [Syntax Overview](#6-syntax-overview)
7. [Formal Grammar (EBNF)](#7-formal-grammar-ebnf)
   - 7.1 [Lexical Grammar](#71-lexical-grammar)
   - 7.2 [Syntactic Grammar](#72-syntactic-grammar)
8. [Parser Strategy](#8-parser-strategy)
9. [Backend Strategy](#9-backend-strategy)
10. [Concurrency Model](#10-concurrency-model)
11. [Standard Library Priorities](#11-standard-library-priorities)
12. [Platform Support](#12-platform-support)
13. [Example Programs](#13-example-programs)
14. [Design Decisions](#14-design-decisions)

---

## 1. Purpose and Vision

Vex is a statically typed, Lisp-based language for building networked services — especially MCP servers.

### Goals

- **Typed Lisp semantics** — S-expression syntax with compile-time type checking; keeps Lisp's expressiveness and metaprogramming
- **Cross-platform** — native binaries on macOS, Linux, and Windows (x86_64 and ARM64)
- **Server-oriented** — first-class support for concurrency, JSON, HTTP, and stream handling
- **Low boilerplate** — the eventual MCP framework should let developers define tools, resources, and prompts with minimal ceremony

### Non-Goals

- GUI applications
- Embedded / bare-metal targets
- Replacing systems languages (Rust, C, Go)

---

## 2. Design Principles

1. **Explicitness over magic** — types are inferred where unambiguous but always expressible; no hidden coercions
2. **Composition over inheritance** — algebraic data types and traits, not class hierarchies
3. **Errors are values** — Result types instead of exceptions; errors must be handled or explicitly propagated
4. **Concurrency is structural** — lightweight tasks and channels built into the language, not a library bolt-on
5. **Interop is practical** — direct Go package imports for ecosystem access; JSON and HTTP in the standard library
6. **Macros are hygienic** — Lisp-style macros with hygienic scoping for DSL construction (critical for the MCP framework)

---

## 3. MCP Server Landscape Analysis

### What MCP Servers Do

MCP is a JSON-RPC 2.0 protocol connecting LLM applications to external data and tools. A server exposes three primitives:

| Primitive     | Purpose                            | Example                         |
|---------------|------------------------------------|---------------------------------|
| **Tools**     | Functions the AI model can invoke  | `search_database`, `send_email` |
| **Resources** | Data/context for users or models   | File contents, DB schemas       |
| **Prompts**   | Templated messages and workflows   | Code review template            |

Servers must handle:

- Lifecycle handshake (`initialize` → capability negotiation → `initialized`)
- JSON-RPC request/response and notification patterns
- Two transports: **stdio** (local) and **Streamable HTTP** (production)
- Concurrent sessions from multiple clients

### Current Pain Points (2025-2026)

- **Boilerplate** — manually wiring JSON-RPC handling, tool schemas, resource templates, error handling, auth, and rate limiting
- **Transport complexity** — stdio works locally, but production needs Streamable HTTP with SSE and auth; most frameworks handle one well, not both
- **Scaling** — stateful sessions conflict with horizontal scaling and load balancers
- **Error fragility** — tool failures crash entire agent sessions; no standard recovery pattern
- **Monitoring** — no standard watchdog/metrics integration for memory leaks or performance

### What This Means for Vex

Each pain point maps to a language-level feature:

| MCP Pain Point           | Vex Language Feature                                    |
|--------------------------|--------------------------------------------------------|
| Boilerplate              | Hygienic macros for declarative tool/resource/prompt definition |
| Transport complexity     | Built-in async I/O with stdio and HTTP support         |
| Scaling / state          | Lightweight tasks + channels; immutable-by-default data |
| Error fragility          | Result types (`Ok`/`Err`) enforced at compile time     |
| Monitoring               | Built-in structured logging and metrics hooks          |

---

## 4. Language Features Driven by MCP Needs

### 4.1 JSON as a First-Class Concern

MCP is entirely JSON-RPC, so JSON serialization must feel natural:

```
(deftype ToolInput
  (name String)
  (arguments (Map String JsonValue)))

(defn handle-tool-call [input : ToolInput] : (Result JsonValue Error)
  (match (. input name)
    "search" (search-handler (. input arguments))
    _        (err (unknown-tool (. input name)))))
```

### 4.2 Algebraic Data Types for Protocol Modeling

MCP messages are naturally modeled as tagged unions:

```
(defunion McpMessage
  (Request Int String JsonValue)
  (Response Int (Result JsonValue McpError))
  (Notification String JsonValue))
```

### 4.3 Async/Concurrent Primitives

MCP servers must handle multiple sessions and upstream API calls concurrently:

```
(defn serve [transport : Transport] : (Result Unit Error)
  (let [sessions (channel Session)]
    (spawn (accept-loop transport sessions))
    (each sessions
      (fn [session] (spawn (handle-session session))))))
```

### 4.4 Macro System for Framework DSLs

The eventual MCP framework should allow declarations like:

```
(deftool search
  :description "Search the knowledge base"
  :params ((query String) (limit (Option Int)))
  :handler (fn [params ctx]
    (kb.search (. params query) (or (. params limit) 10))))
```

`deftool` is a macro that expands into JSON schema registration, handler wiring, and error wrapping — eliminating the boilerplate that plagues current MCP SDKs.

---

## 5. Type System

### 5.1 Approach: Hindley-Milner with Extensions

Based on Hindley-Milner type inference (ML, Haskell, Typed Racket), extended with:

- **Algebraic Data Types** — sum and product types
- **Parametric polymorphism** — generics
- **Traits / Protocols** — ad-hoc polymorphism
- **Result types** — error handling via `(Result T E)`
- **Option types** — nullable values via `(Option T)`

### 5.2 Primitive Types

| Type      | Description                |
|-----------|----------------------------|
| `Int`     | 64-bit signed integer      |
| `Float`   | 64-bit IEEE 754            |
| `Bool`    | `true` / `false`           |
| `String`  | UTF-8 string               |
| `Char`    | Unicode scalar value       |
| `Unit`    | Zero-size type (like void) |

### 5.3 Composite Types

| Type                   | Syntax                              |
|------------------------|-------------------------------------|
| List                   | `(List T)`                          |
| Map                    | `(Map K V)`                         |
| Tuple                  | `(Tuple T1 T2 ...)`                |
| Option                 | `(Option T)`                        |
| Result                 | `(Result T E)`                      |
| Function               | `(Fn [T1 T2] R)`                   |
| Record (product type)  | `(deftype Name (field1 T1) ...)`    |
| Union (sum type)       | `(defunion Name (Variant1 T1) ...)` |

### 5.4 Traits / Protocols

```
(deftrait Serializable
  (serialize [self] : String)
  (deserialize [s : String] : (Result Self Error)))

(impl Serializable for ToolInput
  (defn serialize [self] (json.encode self))
  (defn deserialize [s] (json.decode s)))
```

### 5.5 Type Inference

- Types are inferred in most local contexts
- Return types are inferred from the function body when omitted (e.g., a body ending in `println` infers to `Unit`)
- Return type annotations are only needed when the compiler can't infer unambiguously
- Explicit annotations are **required** on:
  - Top-level function parameter types
  - Trait implementations
  - Ambiguous expressions

---

## 6. Syntax Overview

S-expression syntax — all code is data (homoiconicity), enabling the macro system.

### Bindings

```
(let [x 42
      y "hello"]
  (println y))
```

### Functions

```
(defn add [a : Int, b : Int] : Int
  (+ a b))

(defn greet [name : String] : String
  (str "Hello, " name "!"))
```

### Pattern Matching

```
(match value
  (Some x) (println (str "Got: " x))
  None     (println "Nothing"))
```

### Control Flow

```
(if (> x 0)
  "positive"
  "non-positive")

(cond
  (< x 0) "negative"
  (= x 0) "zero"
  :else    "positive")
```

### Modules

```
(module vex.http)
(export [serve request response])

(defn serve [port : Int, handler : (Fn [Request] Response)] : (Result Unit Error)
  ...)
```

---

## 7. Formal Grammar (EBNF)

Initial draft — will evolve as the language develops.

### 7.1 Lexical Grammar

The lexer transforms source text into a flat stream of tokens:

- Whitespace and comments are consumed during lexing but do not produce tokens

#### Token Types

| Token          | Examples                        | Description                                 |
|----------------|---------------------------------|---------------------------------------------|
| `LeftParen`    | `(`                             | Opens a list / form                         |
| `RightParen`   | `)`                             | Closes a list / form                        |
| `LeftBracket`  | `[`                             | Opens a param list / binding list           |
| `RightBracket` | `]`                             | Closes a param list / binding list          |
| `LeftBrace`    | `{`                             | Opens a map literal                         |
| `RightBrace`   | `}`                             | Closes a map literal                        |
| `Dot`          | `.`                             | Field access, qualified identifiers         |
| `Colon`        | `:`                             | Type annotations (standalone)               |
| `Symbol`       | `foo`, `defn`, `+`, `>=`        | Identifiers and operators                   |
| `Keyword`      | `:name`, `:else`                | Colon-prefixed identifiers                  |
| `Integer`      | `42`, `-7`, `0`                 | 64-bit signed integer literal               |
| `Float`        | `3.14`, `-0.5`                  | 64-bit float literal                        |
| `String`       | `"hello"`, `"a\nb"`            | UTF-8 string literal                        |
| `Boolean`      | `true`, `false`                 | Boolean literal                             |
| `Nil`          | `nil`                           | Nil literal                                 |

#### Whitespace

Spaces, tabs, newlines, carriage returns, and commas are all whitespace. Commas are optional visual separators with no syntactic meaning (Clojure convention).

```
[a : Int, b : Int]   ; equivalent
[a : Int  b : Int]   ; equivalent
```

#### Comments

Line comments start with `;` and extend to end of line.

```
; inline comment
;; full-line comment (convention, not distinct syntax)
```

No block comment syntax in v0.1.

#### Strings

Strings are delimited by double quotes with the following escape sequences:

| Escape     | Meaning                          |
|------------|----------------------------------|
| `\\`       | Backslash                        |
| `\"`       | Double quote                     |
| `\n`       | Newline (LF)                     |
| `\t`       | Tab                              |
| `\r`       | Carriage return                  |
| `\0`       | Null byte                        |
| `\uXXXX`   | Unicode scalar (4 hex digits)    |

Strings may span multiple lines — newlines inside a string are literal newline characters.

#### Numbers

Integers are sequences of digits with an optional leading `-`.

```
42   -7   0
```

Floats require digits on both sides of the decimal point.

```
3.14   -0.5   0.0
```

No underscores in numeric literals, no hex/octal/binary syntax in v0.1.

#### Symbols

Symbols come in two flavors, both producing the same `Symbol` token.

**Alphabetic symbols** start with a letter or underscore, followed by letters, digits, `-`, `_`, `!`, or `?`:

```
foo   bar_baz   handle-tool-call   empty?   set!   _unused
```

**Operator symbols** are sequences of one or more operator characters:

```
+   -   *   /   <   >   =   !=   >=   <=   &&   ||
```

Operator characters: `+` `-` `*` `/` `<` `>` `=` `!` `&` `|` `%` `^` `~`

- Special forms (`def`, `defn`, `let`, `if`, `match`, etc.) are **not** reserved words — they are lexed as ordinary `Symbol` tokens; the parser dispatches on their values
- `true`, `false`, and `nil` are the only words the lexer handles specially — when alphabetic symbol lexing produces one of these values, the lexer emits a `Boolean` or `Nil` token instead of `Symbol`

#### Keywords

A colon immediately followed by an alphabetic symbol (no space) produces a `Keyword` token.

```
:name   :description   :else   :transport
```

A standalone colon (followed by whitespace or a non-alphabetic character) produces a `Colon` token for type annotations.

```
[x : Int]    ; Colon token between Symbol("x") and Symbol("Int")
:else        ; Keyword token
```

#### Dot

- `.` is its own token type, distinct from symbols
- The parser uses it for qualified identifiers and field access
- `vex.http` is lexed as three tokens: `Symbol("vex")`, `Dot`, `Symbol("http")`

#### Disambiguation Rules

These rules resolve ambiguities that pure EBNF cannot express:

1. **Negative numbers** — `-` immediately followed by a digit (no whitespace) is lexed as a negative number. Otherwise `-` is a `Symbol`.
   - `-42` → `Integer(-42)`
   - `(- 42)` → `LeftParen`, `Symbol("-")`, `Integer(42)`, `RightParen`
2. **Keywords vs Colon** — `:` immediately followed by an `ident-start` character (no whitespace) is a `Keyword`. Otherwise it is a `Colon`.
3. **Boolean/Nil promotion** — after lexing an alphabetic symbol, if the value is `true` or `false`, emit `Boolean`. If `nil`, emit `Nil`.
4. **Longest match** — the lexer always consumes the longest valid token. `>=` is one `Symbol`, not `>` followed by `=`.

#### Lexical Grammar (EBNF)

```ebnf
token          = lparen | rparen | lbracket | rbracket
               | lbrace | rbrace | dot | colon
               | string | integer | float
               | boolean | nil | keyword | symbol ;

lparen         = "(" ;
rparen         = ")" ;
lbracket       = "[" ;
rbracket       = "]" ;
lbrace         = "{" ;
rbrace         = "}" ;
dot            = "." ;
colon          = ":" ;  (* when NOT immediately followed by ident-start *)

(* --- Whitespace (consumed, not emitted) --- *)
whitespace     = " " | "\t" | "\n" | "\r" | "," ;

(* --- Comments (consumed, not emitted) --- *)
comment        = ";" { ? any character except newline ? } ( "\n" | ? end of input ? ) ;

(* --- Strings --- *)
string         = '"' { string-char } '"' ;
string-char    = escape-seq | ? any character except '"' or '\' ? ;
escape-seq     = "\" ( '"' | "\" | "n" | "t" | "r" | "0" )
               | "\u" hex hex hex hex ;
hex            = digit | "a" | ... | "f" | "A" | ... | "F" ;

(* --- Numbers --- *)
integer        = [ "-" ] digit { digit } ;
float          = [ "-" ] digit { digit } "." digit { digit } ;

(* --- Booleans and Nil --- *)
boolean        = "true" | "false" ;
nil            = "nil" ;

(* --- Keywords --- *)
keyword        = ":" alpha-ident ;

(* --- Symbols --- *)
symbol         = alpha-ident | operator-ident ;
alpha-ident    = ident-start { ident-continue } ;
ident-start    = letter | "_" ;
ident-continue = letter | digit | "-" | "_" | "!" | "?" ;
operator-ident = op-char { op-char } ;
op-char        = "+" | "-" | "*" | "/" | "<" | ">"
               | "=" | "!" | "&" | "|" | "%" | "^" | "~" ;

(* --- Character classes --- *)
letter         = "a" | ... | "z" | "A" | ... | "Z" ;
digit          = "0" | ... | "9" ;
```

### 7.2 Syntactic Grammar

The parser consumes the token stream and produces a typed AST:

- Quoted strings like `"defn"` and `"let"` below match against `Symbol` token values, not raw characters
- `SYMBOL`, `STRING`, `INTEGER`, `FLOAT`, `BOOLEAN`, `NIL`, `KEYWORD` in uppercase refer to token types from the lexical grammar

```ebnf
program        = { top-form } ;

top-form       = module-decl
               | export-decl
               | import-decl
               | import-go-decl
               | def-form
               | defn-form
               | deftype-form
               | defunion-form
               | deftrait-form
               | impl-form
               | defmacro-form
               | expression ;

module-decl    = "(" "module" qualified-id ")" ;

export-decl    = "(" "export" "[" { SYMBOL } "]" ")" ;

import-decl    = "(" "import" qualified-id "[" { SYMBOL } "]" ")" ;

import-go-decl = "(" "import-go" STRING "[" { SYMBOL } "]" ")" ;

def-form       = "(" "def" SYMBOL [ type-ann ] expression ")" ;

defn-form      = "(" "defn" SYMBOL param-list [ ":" type ] body ")" ;

deftype-form   = "(" "deftype" SYMBOL { field-decl } ")" ;

defunion-form  = "(" "defunion" SYMBOL { variant-decl } ")" ;

deftrait-form  = "(" "deftrait" SYMBOL { trait-method } ")" ;

impl-form      = "(" "impl" SYMBOL "for" SYMBOL { defn-form } ")" ;

defmacro-form  = "(" "defmacro" SYMBOL param-list body ")" ;

param-list     = "[" { param } "]" ;

param          = SYMBOL [ ":" type ] ;

field-decl     = "(" SYMBOL type ")" ;

variant-decl   = "(" SYMBOL { type } ")" ;

trait-method   = "(" SYMBOL param-list ":" type ")" ;

type-ann       = ":" type ;

type           = SYMBOL
               | "(" "Fn" "[" { type } "]" type ")"
               | "(" SYMBOL { type } ")" ;

body           = expression { expression } ;

expression     = literal
               | qualified-id
               | KEYWORD
               | "(" "let" binding-list body ")"
               | "(" "if" expression expression expression ")"
               | "(" "match" expression match-clause { match-clause } ")"
               | "(" "cond" { cond-clause } ")"
               | "(" "fn" param-list [ ":" type ] body ")"
               | "(" "spawn" expression ")"
               | "(" "channel" type [ expression ] ")"
               | "(" "send" expression expression ")"
               | "(" "recv" expression ")"
               | "(" "select" { select-clause } ")"
               | "(" "." expression SYMBOL ")"
               | "(" "quote" expression ")"
               | "(" "unquote" expression ")"
               | "(" "splice" expression ")"
               | map-literal
               | "(" expression { expression } ")" ;

binding-list   = "[" { SYMBOL expression } "]" ;

match-clause   = pattern [ ":when" expression ] body ;

cond-clause    = expression body ;

select-clause  = "[" "(" "recv" expression ")" SYMBOL body "]"
               | "[" "(" "send" expression expression ")" body "]"
               | "[" ":default" body "]" ;

pattern        = literal
               | "_"
               | SYMBOL
               | "(" SYMBOL { pattern } ")" ;

map-literal    = "{" { map-entry } "}" ;

map-entry      = expression expression ;

literal        = INTEGER | FLOAT | STRING | BOOLEAN | NIL | KEYWORD ;

qualified-id   = SYMBOL { "." SYMBOL } ;
```

---

## 8. Parser Strategy

### Approach: Hand-Written Recursive Descent Parser

S-expressions are syntactically simple (atoms and lists) — a parser generator would add complexity without benefit.

#### Why Not Parser Generators?

| Factor                  | Parser Generator                   | Hand-Written Recursive Descent     |
|-------------------------|------------------------------------|------------------------------------|
| S-expression complexity | Overkill — syntax is trivial       | Trivially handles S-expressions    |
| Error messages          | Generic, hard to customize         | Full control over diagnostics      |
| Dependencies            | Adds build-time dependency         | Zero dependencies                  |
| Maintenance             | Grammar file + generated code      | Single codebase                    |
| Incremental parsing     | Difficult to add later             | Natural extension                  |

#### Implementation Plan

Two phases:

**Phase 1 — Lexer** (see §7.1 for full specification): scans source text into tokens.

- Delimiters: `LeftParen`, `RightParen`, `LeftBracket`, `RightBracket`, `LeftBrace`, `RightBrace`
- Atoms: `Symbol`, `Keyword`, `String`, `Integer`, `Float`, `Boolean`, `Nil`
- Punctuation: `Colon` (type annotations), `Dot` (field access, qualified identifiers)
- Whitespace (including commas) and comments are consumed, not emitted

**Phase 2 — Parser**: reads tokens, produces a typed AST.

- Atom token (symbol, number, string, bool) → AST leaf
- `(` token → parse list of expressions until `)` → AST node
- Top-level forms (`defn`, `deftype`, etc.) identified by the first symbol in a list

#### Why Rust for the Compiler

- Strong type system catches compiler bugs at compile time
- Excellent performance for parsing and compilation
- Cross-compiles to all target platforms (macOS, Linux, Windows)
- Rich ecosystem (`logos` for lexing if desired, though hand-written is fine)

---

## 9. Backend Strategy

### Analysis of Options

| Approach | Cross-Platform | Performance | Complexity | Ecosystem | MCP Suitability |
|----------|---------------|-------------|------------|-----------|-----------------|
| **Transpile to C** | Excellent | High | Medium | Broad C FFI | Needs manual async runtime |
| **Transpile to Go** | Excellent | Good | Low | Built-in HTTP/JSON/concurrency | Very High |
| **LLVM** | Excellent | Highest | Very High | Full native | Needs full runtime |
| **QBE** | No Windows | Good (70% of LLVM) | Low | C ABI only | Limited |
| **Cranelift** | Good | High | High | Rust ecosystem | Needs runtime |

### Decision: Rust Compiler → Go Transpilation Target

The Vex compiler is written in Rust. It outputs Go source code, which the Go toolchain compiles to a native binary.

#### Why Go as the target

- **Built-in concurrency** — goroutines and channels map directly to Vex's `spawn` and `channel`; no custom scheduler needed
- **Built-in HTTP and JSON** — `net/http` and `encoding/json` are production-grade and needed on day one
- **Cross-compilation** — `GOOS`/`GOARCH` flags produce static binaries for any platform from any platform
- **Static binaries** — single-file executables with no runtime dependencies
- **Low-latency GC** — tuned for server workloads
- **Fast compilation** — keeps the Vex → Go → binary pipeline fast
- **Mature networking** — TLS, HTTP/2, SSE, and WebSocket out of the box

#### Alternatives considered

| Approach | Pros | Cons |
|----------|------|------|
| **Rust → Go** (chosen) | Each language at its strength; Go runtime is free | Two toolchains; debugging spans two worlds |
| **Go → Go** | Single language; simpler onboarding | Go lacks sum types / pattern matching, making AST code verbose |
| **Rust → LLVM** | Max performance; no transpilation layer | Must build full runtime (GC, scheduler, HTTP, JSON) |
| **Rust → C** | Universal target | Weak server ecosystem; no built-in concurrency/HTTP/JSON |
| **Rust → Rust** | Strong safety in output; access to tokio/serde | Generating borrow-checker-valid Rust from codegen is extremely hard |

#### Why this split works

- Compilers are parsing + type checking + tree transforms — Rust excels here
- Output programs are networked servers — Go's runtime provides everything with zero custom infrastructure
- The cost (two toolchains) is worth the massive reduction in runtime work

#### Known limitations

- Go's generics may not fully express Vex's type system in generated code
- GC pauses are small but not zero
- Transpiled code is harder to debug than natively compiled code

#### Why Go is the only backend

LLVM was considered as a future backend and dropped. MCP servers are bottlenecked by I/O and protocol architecture, not raw CPU:

- **CPU performance** — tool handlers wait on databases, APIs, and LLM calls; JSON-RPC routing overhead is tiny
- **GC pauses** — Go's are <500us; MCP requests have 1-100ms network latency, so GC pauses are noise
- **Binary size** — Go binaries are ~10MB; irrelevant for Docker/VM/Lambda deployments
- The 2026 MCP roadmap's top scaling challenges are transport evolution and stateless session design — architectural problems, not compute problems
- An LLVM backend would require building a GC, a scheduler, and the entire stdlib natively — enormous effort for a problem that doesn't exist in this workload

#### Transpilation Architecture

```
Source (.vx)
    │
    ▼
┌─────────┐
│  Lexer  │    Rust
└────┬────┘
     │ tokens
     ▼
┌─────────┐
│ Parser  │    Rust
└────┬────┘
     │ AST
     ▼
┌──────────┐
│Type Check│   Rust
└────┬─────┘
     │ Typed AST
     ▼
┌──────────┐
│ Codegen  │   Rust → Go source
└────┬─────┘
     │ .go files
     ▼
┌──────────┐
│ Go Build │   go build
└────┬─────┘
     │
     ▼
  Binary
```

---

## 10. Concurrency Model

Lightweight tasks and channels, inspired by Go's goroutines and CSP.

```
(let [ch (channel Int 10)]
  (spawn
    (each (range 0 10)
      (fn [i] (send ch i))))
  (each (range 0 10)
    (fn [_] (println (recv ch)))))
```

- `spawn` → goroutine
- `channel` → Go channel

This gives Vex a battle-tested concurrent runtime for free.

---

## 11. Standard Library Priorities

Ordered by importance for the MCP server use case:

| Priority | Module        | Contents                                    |
|----------|--------------|---------------------------------------------|
| P0       | `vex.json`   | JSON encode/decode, JsonValue type          |
| P0       | `vex.io`     | Stdin/stdout, file I/O, buffered readers    |
| P0       | `vex.string` | String manipulation, formatting             |
| P0       | `vex.result` | Result/Option types, combinators            |
| P1       | `vex.http`   | HTTP client and server                      |
| P1       | `vex.net`    | TCP/UDP sockets, TLS                        |
| P1       | `vex.async`  | Channels, spawn, select, timeouts           |
| P1       | `vex.log`    | Structured logging                          |
| P2       | `vex.os`     | Environment variables, process management   |
| P2       | `vex.map`    | Persistent/immutable map                    |
| P2       | `vex.list`   | List operations, functional transforms      |
| P3       | `vex.test`   | Testing framework                           |
| P3       | `vex.macro`  | Macro utilities                             |

---

## 12. Platform Support

### Target Matrix

| Platform       | Architecture | Status  |
|---------------|-------------|---------|
| macOS         | x86_64      | Primary |
| macOS         | ARM64 (M-series) | Primary |
| Linux         | x86_64      | Primary |
| Linux         | ARM64       | Primary |
| Windows       | x86_64      | Primary |
| Windows       | ARM64       | Secondary |

All targets are supported via Go's `GOOS`/`GOARCH` cross-compilation from a single build machine.

### Compiler Distribution

Distributed as a native Rust binary per platform:

- GitHub Releases
- Homebrew (macOS)
- APT/RPM (Linux)
- Scoop or WinGet (Windows)

---

## 13. Example Programs

### Hello World

```
(defn main []
  (println "Hello, World!"))
```

### FizzBuzz

```
(defn fizzbuzz [n : Int] : String
  (cond
    (= (mod n 15) 0) "FizzBuzz"
    (= (mod n 3) 0)  "Fizz"
    (= (mod n 5) 0)  "Buzz"
    :else             (str n)))

(defn main []
  (each (range 1 101)
    (fn [i] (println (fizzbuzz i)))))
```

### Simple HTTP Server

```
(import vex.http [serve, response])

(defn handler [req : http.Request] : http.Response
  (response 200 "OK" (json.encode {:status "healthy"})))

(defn main []
  (match (serve 8080 handler)
    (Ok _)  (println "Server started on :8080")
    (Err e) (println (str "Failed: " e))))
```

### MCP Tool Definition (Future Framework Preview)

```
(import vex.mcp [deftool, defresource, serve-mcp])

(deftool search-docs
  :description "Search the documentation"
  :params ((query : String) (limit : (Option Int)))
  (let [results (db.search query (or limit 10))]
    (ok (json.encode results))))

(defresource schema
  :uri "schema://main"
  :description "Database schema"
  (ok (db.get-schema)))

(defn main []
  (serve-mcp {:port 3000
              :transport :streamable-http}))
```

---

## 14. Design Decisions

### 1. Memory management — GC-only

- Vex is garbage-collected; no exposed memory management — Go's GC handles everything
- Ownership (Rust/Carp style) and arenas were rejected:
  - MCP servers are I/O-bound — no benefit from manual memory control
  - It would fundamentally change the Lisp programming model
- Value types (stack-allocated records) can still be used as a codegen optimization, but are not user-facing

### 2. Macro expansion — before type checking

- Pipeline: `Parse → Macro Expand → Type Check → Codegen`
- Macros operate on raw S-expressions as pure syntax transformations
- The type checker validates the fully expanded code
- Source spans must be tracked through expansion so errors point at the macro call site
- Type-aware macros were rejected:
  - Creates a chicken-and-egg problem in the compiler
  - Target use cases (`deftool`, `defresource`, `defprompt`) are purely syntactic and don't need type information

### 3. Module system — one file = one module, no cyclic imports

- File path determines module name (`src/http/server.vx` → `http.server`)
- Each file declares its exports
- Modules split into submodules when they grow (not multiple files per module)
- Compiler errors on any dependency cycle
- `(module ...)` is a top-of-file declaration, not a wrapper around all forms

### 4. Go interop — direct, first-class feature

- Vex exposes Go package imports (e.g., `(import-go "net/http" [...])`) as a standard language feature
- The stdlib already needs this mechanism internally
- Forcing users through C FFI via cgo would break cross-compilation
- With Go as the sole backend, there is no portability concern — Go interop is native, not an escape hatch

### 5. String interpolation — `(str ...)` only, no interpolation syntax

- String construction uses the `str` function, which macros can generate and transform
- MCP servers rarely need manual string building — JSON uses `json.encode`, errors are Result types, logging uses structured key-value pairs
- Interpolation syntax (`"Hello, {name}"`) was rejected:
  - Adds parser complexity (recursive expression parsing inside strings)
  - Compromises homoiconicity
  - Ergonomic gain over `(str "Hello, " name "!")` is minimal

### 6. REPL — tree-walking interpreter

- Uses a recursive AST interpreter for instant feedback
- Shares the same type checker as the compiled path
- Strong early milestone — `Parse → Type Check → Interpret` can work before codegen exists
- JIT adds unjustifiable complexity for REPL-sized expressions
- Transpile-and-run (full Go pipeline per expression) adds seconds of latency

### 7. Tail call optimization — automatic direct tail recursion → loop

- Compiler detects self-recursive calls in tail position and generates Go `for` loops with variable reassignment
- Semantically identical, constant stack, faster (no call overhead, better cache, no goroutine stack growth)
- No user annotation needed — tail position detection is automatic
- Non-tail recursive functions get a compiler warning about stack usage proportional to input size
- Mutual tail recursion is **not** optimized — trampolining adds allocation overhead for a rare case in MCP code

### 8. Effect system — no effect tracking

- Side effects are not tracked in the type system
- MCP servers are inherently effectful — nearly every function does I/O, so effect annotations would be noise
- Go has no effect concept, making encoding effects in generated code costly
- Compiler complexity is better spent on the MCP framework
- Testing is covered by Result types and dependency injection
- If purity annotations are ever wanted, an opt-in `:pure` marker can be added later without breaking existing code

---

## Appendix A: Prior Art

| Language | Relevance to Vex |
|----------|------------------|
| **Clojure** | Lisp on JVM, practical focus, data-oriented |
| **Typed Racket** | Gradual typing on Lisp, occurrence typing |
| **Carp** | Statically typed Lisp, no GC, Rust-like ownership |
| **Hy** | Lisp transpiled to Python |
| **Fennel** | Lisp transpiled to Lua |
| **Elm** | ML-family, Result types, no runtime exceptions |
| **Go** | Target language, concurrency model inspiration |

## Appendix B: Glossary

| Term | Definition |
|------|-----------|
| **S-expression** | Symbolic expression — nested list notation `(op arg1 arg2)` |
| **Homoiconicity** | Code-as-data; the program's syntax tree is a data structure in the language |
| **ADT** | Algebraic Data Type — sum types (tagged unions) + product types (records) |
| **HM** | Hindley-Milner — a type inference algorithm that can infer types without annotations |
| **MCP** | Model Context Protocol — open protocol for LLM-tool integration |
| **JSON-RPC** | Remote procedure call protocol encoded in JSON |
| **TCO** | Tail Call Optimization — converting recursive tail calls into loops |
| **CSP** | Communicating Sequential Processes — concurrency model using channels |
