# Diagnostics (structured compiler errors)

Purpose: Provide stable, structured diagnostics with codes, canonical messages, and renderers for human-readable text and (future) machine-readable JSON.

Key types:
- `Code`: stable IDs like `VEX-TYP-UNDEF`, `VEX-PKG-CYCLE` (see `codes.go`).
- `Diagnostic`: structured payload with `Code`, position, `Params`, and optional `Suggestion`.
- Renderers: `RenderText()`, `RenderJSON()` and `RenderBody()`.

Usage pattern:
1) Construct a diagnostic with code + params and position.
2) Render to text and send to the existing error reporter.

Example (pseudo-call site in analysis):
```
import (
    "github.com/thsfranca/vex/internal/transpiler/diagnostics"
)

diag := diagnostics.New(
    diagnostics.CodePkgNotExported,
    diagnostics.SeverityError,
    file, line, col,
    map[string]any{"Package": pkgPath, "Symbol": symbol},
).WithSuggestion("Export it with (export ["+symbol+"]) in that package.")

// If integrating with the current reporter without changing its API:
reporter.ReportDiagnosticBody(line, col, diag.RenderBody(), analysis.SemanticError)
```

Rendering format (text):
- `path/to/file.vx:LINE:COL: error: [CODE]: short-message` followed by optional lines for `Expected`, `Got`, `Offender`, `Suggestion`.

Notes:
- Short messages for common codes live in `catalog.go` and should stay concise.
- When a message requires custom text, set it via `WithMessage()` and still include a `Code`.
- A future `--machine` flag can surface `RenderJSON()` output.


