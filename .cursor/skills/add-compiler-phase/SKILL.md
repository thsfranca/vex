---
name: add-compiler-phase
description: >-
  Guides through creating a new compiler phase in the Vex pipeline.
  Use when adding a transformation pass, optimization, or analysis step to the compiler.
---

# Add Compiler Phase

Every Vex compiler phase is a pure function: data in, data out, plus diagnostics. No global state, no singletons.

## Phase Contract Template

```rust
fn phase_name(input: &InputType) -> (OutputType, Vec<Diagnostic>)
```

Existing phase signatures for reference:

| Phase | Signature |
|-------|-----------|
| Lexer | `fn lex(source: &str, file: FileId) -> (Vec<Token>, Vec<Diagnostic>)` |
| Parser | `fn parse(tokens: &[Token]) -> (Vec<ast::TopForm>, Vec<Diagnostic>)` |
| Type Checker | `fn check(program: &[ast::TopForm]) -> (hir::Module, Vec<Diagnostic>)` |
| Codegen | `fn generate(module: &hir::Module) -> String` |
| Interpreter | `fn eval(module: &hir::Module) -> Result<Value, RuntimeError>` |

## Steps

### 1. Determine Where It Fits

The pipeline is linear. Identify the input/output boundary:

```
Source → Lexer → Parser → [Macro Expansion] → Type Checker → Codegen/Interpreter
```

- Between Parser and Type Checker: operates on `ast::TopForm`, produces `ast::TopForm`
- After Type Checker: operates on `hir::Module`, produces `hir::Module`

### 2. Create the File

One file per concept in `src/`. Name it after the phase (e.g., `src/optimizer.rs`).

Check the dependency table in `docs/compiler-architecture.md` §3 to determine allowed imports:
- The new file may only depend on modules below it in the DAG
- It must not create dependency cycles

### 3. Define the Function Signature

Follow the contract pattern:
- Takes a reference to input data
- Returns owned output data + `Vec<Diagnostic>`
- If the phase cannot fail (like codegen), omit diagnostics

### 4. Register in `lib.rs`

Add the module declaration and wire it into the `compile()` pipeline function.

The pipeline in `lib.rs` chains phases sequentially:
1. Lex
2. Parse
3. (New phase here if AST→AST)
4. Type check
5. (New phase here if HIR→HIR)
6. Generate / Eval

### 5. Update the Dependency Table

Add the new file to the dependency table in `docs/compiler-architecture.md` §3.

### 6. Tests

Unit tests in `#[cfg(test)] mod tests` within the file:
- Construct input data by hand or using earlier phases
- Assert output structure and diagnostics
- Test error cases — assert expected `Diagnostic` messages

## Rules

- No mutable statics or global state
- Receive `&mut Vec<Diagnostic>` to push errors, or return `Vec<Diagnostic>` alongside output
- The caller (`lib.rs`) decides whether to continue after errors
- Each file owns one concept; split when a file owns two independent concerns, not by line count
