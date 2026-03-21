# Vex

[![CI](https://github.com/thsfranca/vex/actions/workflows/ci.yml/badge.svg)](https://github.com/thsfranca/vex/actions/workflows/ci.yml)

A statically typed Lisp for building MCP servers. The compiler is written in Rust and transpiles to Go source code.

**This is a study project.**

## Overview

Vex uses S-expression syntax with compile-time type checking, targeting networked services — especially [MCP](https://modelcontextprotocol.io/) servers. The compiler pipeline (lexer → parser → type checker → codegen) produces Go source that `go build` compiles to a native binary.

## Building

```
cargo build
```

## License

MIT
