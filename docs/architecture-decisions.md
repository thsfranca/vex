## ADR: Core macros organized as stdlib packages

- Context: Core Vex macros were in a monolithic `core/core.vx`. We want single-responsibility packages and a stable stdlib layout.
- Decision: Introduce stdlib packages under `stdlib/vex/` and load macros from them:
  - `stdlib/vex/core` for `defn`
  - `stdlib/vex/conditions` for `when`, `unless`
  - `stdlib/vex/collections` for `first`, `rest`, `count`, `cons`, `empty?`
- Consequences:
  - Macro registry loads all `.vx` files in these directories, in addition to an explicit path when provided.
  - Legacy fallback `core/core.vx` remains supported.
  - Future stdlib features can add new packages without changing the transpiler.

# Architecture Decision Records (ADR)

This document records significant architectural decisions for the Vex project. Each entry summarizes the context, decision, consequences, and references to the commits where the decision was introduced.

## ADR-0001: Primary compilation target is Go

- Status: Accepted
- Decision: Vex transpiles to Go as its primary target.
- Context: The language, tooling, and examples rely on Go constructs (goroutines, `net/http`, GC), and the CLI builds and runs via the Go toolchain.
- Consequences: Tight integration with Go stdlib and modules; concurrency model built on goroutines; generated code uses Go idioms (`interface{}`, `[]interface{}`, `len`, `append`).
 - References: [6713572](https://github.com/thsfranca/vex/commit/6713572), [PR #38](https://github.com/thsfranca/vex/pull/38)

## ADR-0002: CLI commands — transpile, run, build (no exec)

- Status: Accepted
- Decision: The CLI exposes `transpile`, `run`, and `build`. The `exec` command is removed.
- Context: Streamline developer workflow and align with common compiler ergonomics.
- Consequences: Users transpile to Go, run directly using Go build pipeline, or build binaries; simpler surface area.
 - References: [dbf3cde](https://github.com/thsfranca/vex/commit/dbf3cde), [PR #43](https://github.com/thsfranca/vex/pull/43)

## ADR-0003: Package discovery MVP with directory-based packages and exports enforcement

- Status: Accepted
- Decision: Adopt Go-inspired directory-based package model with automatic discovery from the entry file, topological ordering, explicit `(export [...])`, private-by-default symbols, and circular dependency detection that fails compilation.
- Context: Scale to multi-package projects and enable controlled cross-package visibility.
- Consequences: Resolver walks package graph; local packages are compiled into the program and suppressed from Go import emission; cycles and unexported symbol calls are compile-time errors.
 - References: [341407b](https://github.com/thsfranca/vex/commit/341407b), [PR #49](https://github.com/thsfranca/vex/pull/49)

## ADR-0004: Coverage policy — threshold 85%, exclude generated parser

- Status: Accepted
- Decision: Enforce overall coverage of at least 85% for `internal/transpiler`, excluding generated parser code from coverage calculations.
- Context: Focus coverage on hand-written compiler components; generated code would skew metrics.
- Consequences: CI builds fail below threshold; parser package excluded from coverpkg list; badges and reports reflect policy.
 - References: [3bdf321](https://github.com/thsfranca/vex/commit/3bdf321), [PR #47](https://github.com/thsfranca/vex/pull/47); [36bd4f3](https://github.com/thsfranca/vex/commit/36bd4f3), [PR #48](https://github.com/thsfranca/vex/pull/48)

## ADR-0005: Release and changelog automation via tags and Go tooling

- Status: Accepted
- Decision: Switch to tag-based versioning, generate releases with a Go-based `tools/release-manager`, and update changelog via automated PRs using a bot token.
- Context: Reliability and portability of release flows; avoid embedding complex logic in workflow YAML.
- Consequences: Version bumps, tags, and release notes are automated; changelog updates flow through PRs.
 - References: [57a5eba](https://github.com/thsfranca/vex/commit/57a5eba), [PR #51](https://github.com/thsfranca/vex/pull/51); [37edf09](https://github.com/thsfranca/vex/commit/37edf09), [PR #51](https://github.com/thsfranca/vex/pull/51); [6276c88](https://github.com/thsfranca/vex/commit/6276c88), [PR #51](https://github.com/thsfranca/vex/pull/51); [ca33216](https://github.com/thsfranca/vex/commit/ca33216), [PR #52](https://github.com/thsfranca/vex/pull/52); [eb18524](https://github.com/thsfranca/vex/commit/eb18524), [PR #53](https://github.com/thsfranca/vex/pull/53)

## ADR-0006: Macro system — no fallback; compile-time expansion; defn built-in

- Status: Accepted
- Decision: Remove fallback macro code paths, require explicit registration/expansion at compile time, and provide a built-in `defn` macro for function definitions.
- Context: Predictability and correctness over permissive fallbacks; aligns with preference that wrong macro usage should fail compilation.
- Consequences: Stricter compilation errors for macro misuse; clearer expansion pipeline; simpler runtime.
 - References: [e814a07](https://github.com/thsfranca/vex/commit/e814a07), [PR #47](https://github.com/thsfranca/vex/pull/47); [77b13bc](https://github.com/thsfranca/vex/commit/77b13bc), [PR #45](https://github.com/thsfranca/vex/pull/45); [a754330](https://github.com/thsfranca/vex/commit/a754330), [PR #46](https://github.com/thsfranca/vex/pull/46)

## ADR-0007: Error message standards for compiler and resolver

- Status: Accepted
- Decision: Standardize error messages to include precise context (e.g., cycle chains with file locations) and clear, actionable wording.
- Context: Improve developer UX and debuggability in multi-package projects.
- Consequences: Resolver and transpiler emit structured, informative errors; tests assert on standardized formats.
 - References: [341407b](https://github.com/thsfranca/vex/commit/341407b), [PR #49](https://github.com/thsfranca/vex/pull/49)

## ADR-0008: Implicit returns for top-level defs and dependency handling for self-hosting

- Status: Accepted
- Decision: Implement implicit return handling for top-level definitions to satisfy the Go compiler and enable self-hosting workflows that depend on definition ordering.
- Context: Generated Go requires statements that compile even when values are not directly used; facilitates bootstrapping.
- Consequences: Top-level `def` generates an assignment and a subsequent usage to avoid unused variable errors.
 - References: [cf509a0](https://github.com/thsfranca/vex/commit/cf509a0), [PR #38](https://github.com/thsfranca/vex/pull/38)

## ADR-0009: Performance baselining with micro-benchmarks

- Status: Accepted
- Decision: Add micro-benchmarks for codegen, macro expander, analyzer, and package resolver to track and enforce performance baselines.
- Context: Guard against regressions as features grow; inform future optimization phases.
- Consequences: Benchmarks become part of routine validation; informs Phase 4.5 performance goals.
 - References: [b3ed0cc](https://github.com/thsfranca/vex/commit/b3ed0cc), [PR #50](https://github.com/thsfranca/vex/pull/50)

## ADR-0010: Type system — adopt Hindley–Milner (HM) inference post-macro

- Status: Accepted
- Decision: Implement a Hindley–Milner type system in the analyzer after macro expansion. Use Algorithm W with unification (occur-check), principal types, let-polymorphism (generalize at `def`/`defn`), and instantiation at use sites. Annotate AST nodes with inferred types and report type errors through the existing reporter. Keep grammar unchanged.
- Context: Alternatives considered included explicit annotations only, bidirectional local inference, constraint-based inference with subtyping/qualified types, Go-like nominal generics, type classes, row polymorphism, and gradual typing. Given AI-generation goals, performance, and Go interop, HM provides concise programs with strong static guarantees, preserves the functional paradigm, and fits the current pipeline (post-macro, pre-codegen). Subtyping/interfaces and advanced features can be layered later where they map cleanly to Go.
- Consequences:
  - Pros: principal types with minimal annotations; strong compile-time safety; better AI ergonomics; enables more specialized codegen where types are known.
  - Cons: no native subtyping; FFI/Go interop may require boundary annotations; inference engine adds complexity; error localization can be harder than fully annotated systems; macros require inference after expansion.
  - Implementation notes: introduce `Type`, `TypeVar`, `TypeScheme`, `TypeEnv`; build a unifier with occur-check; implement inference for core forms (`def`, `fn`, calls, `if`, arrays, records); generalize at let-bindings; instantiate at use; consider value restriction for future effects; integrate with codegen to reduce `interface{}`; add tests and micro-benchmarks.
    - Naming: Prefer descriptive names (e.g., `Type`, `TypeVariable`, `TypeScheme`, `TypeEnvironment`, `Unifier`, `Infer`) while ADRs/docs may reference Hindley–Milner.
- References: To be added when the implementation PR is created.

## ADR-0011: Structured diagnostics with renderer (text and JSON-ready)

- Status: Accepted
- Decision: Introduce a structured diagnostics system that assigns a stable error code to every compiler/resolver diagnostic, renders human-friendly text following our error message conventions, and supports optional machine-readable JSON output behind a future `--machine` flag.
- Context: The codebase had many ad-hoc hardcoded strings for errors, making consistency, refactoring, and AI/tooling integration difficult. Our documented standards require stable codes (e.g., `VEX-TYP-…`), Go-style location prefixes, and optional lines for `Expected`, `Got`, `Offender`, and concise `Suggestion` messages.
- Consequences:
  - A new package `internal/transpiler/diagnostics` provides:
    - `Diagnostic` struct with code, position, params, and optional suggestion
    - A message catalog with short, canonical templates per code
    - Renderers for text (default) and JSON (opt-in future)
  - The existing `ErrorReporter` gains adapter methods to accept diagnostics without breaking current call sites.
  - Error tests can assert on codes and parameters rather than brittle full strings.
  - Migration will progressively replace `fmt.Errorf` strings with diagnostics builders in high-traffic areas (resolver, analyzer, macros, codegen validation).
- References: To be updated with the PR implementing the package and adapters.

---

Conventions:
- New architectural decisions should be added as new ADR entries following this format.
- Keep entries concise and cite the commits that implemented the decision.


