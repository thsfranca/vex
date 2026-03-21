---
name: vex-go-codegen
description: >-
  Reference for Vex-to-Go code generation rules, type mappings, and output model.
  Use when working on codegen.rs or debugging generated Go output.
---

# Vex Go Codegen Reference

Codegen takes typed HIR (`hir::Module`) and produces a Go source string. No diagnostics — HIR is valid by construction after type checking.

```
fn generate(module: &hir::Module) -> String
```

## Type Mapping

### Primitives

| Vex | Go |
|-----|-----|
| `Int` | `int64` |
| `Float` | `float64` |
| `Bool` | `bool` |
| `String` | `string` |
| `Char` | `rune` |
| `Unit` | (no return value); as type parameter: `struct{}` |

### Collections

| Vex | Go |
|-----|-----|
| `(List T)` | `[]T` |
| `(Map K V)` | `map[K]V` |
| `(Tuple T1 T2 ...)` | `vexrt.Tuple2[T1, T2]`, `vexrt.Tuple3[...]`, etc. |

### Functions

| Vex | Go |
|-----|-----|
| `(Fn [T1 T2] R)` | `func(T1, T2) R` |

### Option and Result

Defined in `vexrt/` as tagged generic structs:

```go
// vexrt/option.go
type Option[T any] struct { IsSome bool; Value T }
func Some[T any](v T) Option[T]  { return Option[T]{IsSome: true, Value: v} }
func None[T any]() Option[T]     { return Option[T]{} }

// vexrt/result.go
type Result[T any, E any] struct { IsOk bool; Value T; Error E }
func Ok[T any, E any](v T) Result[T, E]  { return Result[T, E]{IsOk: true, Value: v} }
func Err[T any, E any](e E) Result[T, E] { return Result[T, E]{Error: e} }
```

### Records (`deftype`)

Map to Go structs with exported (PascalCase) fields:

```
(deftype ToolInput (name String) (arguments (Map String JsonValue)))
→ type ToolInput struct { Name string; Arguments map[string]any }
```

### Unions (`defunion`)

Interface + struct-per-variant pattern:

```
(defunion Msg (Req Int) (Resp String))
→ type Msg interface { isMsg() }
  type Msg_Req struct { V0 int64 }
  func (Msg_Req) isMsg() {}
  type Msg_Resp struct { V0 string }
  func (Msg_Resp) isMsg() {}
```

### Concurrency

| Vex | Go |
|-----|-----|
| `(channel T)` | `chan T` |
| `(channel T size)` | `make(chan T, size)` |
| `(spawn expr)` | `go func() { expr }()` |
| `(send ch val)` | `ch <- val` |
| `(recv ch)` | `<-ch` |
| `(select ...)` | `select { ... }` |

### Pattern Matching

| Match target | Go codegen |
|-------------|------------|
| `Option[T]` | `if opt.IsSome { val := opt.Value; ... } else { ... }` |
| `Result[T, E]` | `if res.IsOk { val := res.Value; ... } else { err := res.Error; ... }` |
| `defunion` | `switch v := val.(type) { case Variant1: ...; case Variant2: ... }` |
| Literal | `switch val { case "foo": ...; case 42: ... }` |

## Naming Convention

Vex kebab-case → Go PascalCase:

| Vex | Go |
|-----|-----|
| `handle-tool-call` | `HandleToolCall` |
| field `name` | `Name` |
| variant `Request` | `McpMessage_Request` |

## Output Structure

```
/tmp/vex-build-XXXX/
  go.mod        — module "vex_out"
  main.go       — generated from .vx source
  vexrt/        — only if needed (Option, Result, Tuple, union helpers)
    result.go
    option.go
    union.go
```

The `vexrt/` directory is omitted entirely when the program doesn't use any runtime types.

## Build Flow

1. Create temp directory
2. Write `go.mod` + `main.go` + `vexrt/` (if needed)
3. Run `go build -o <output_path>`
4. Delete temp directory
5. Binary named after source file: `hello.vx` → `hello`

`--emit-go <dir>` writes the Go module to a directory instead of deleting it.
