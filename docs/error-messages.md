## Vex Compiler Error Message Conventions

Vex follows Go-style error messaging: concise, actionable, and easy to scan. The goals are fast diagnosis and clear fixes.

### Format
- Prefer: `path/to/file.vx:LINE:COL: error: message`
- If file/position is unknown (e.g., cross-file cycle detection), omit them and include best-available context.
- Paths are relative to the module root detected by `vex.pkg` when available.

### Position information
- Include line and column whenever the error originates from a specific token or node.
- If an error spans multiple nodes (e.g., import graphs), include at least the first offending location and any edges that clarify the issue.

### Suggesting fixes
Include a short fix suggestion when it is obvious and safe:
- Undefined symbol: suggest defining it or importing the right package.
- Non-exported symbol: suggest exporting it, e.g., `Export it with (export [name]) in that package.`
- Invalid import usage: show the correct syntax quickly.
- Cycle detection: show the full chain and, when available, the file where each import edge was declared.

Suggestions must be:
- Short (one line), concrete, and technically correct.
- Avoid guessing beyond common, low-risk fixes.

### Examples
- Syntax error
  - `app/main.vx:12:7: error: expected ')' but found 'EOF'`

- Undefined symbol
  - `services/auth/login.vx:34:3: error: undefined symbol 'verify-token'
    Suggestion: define 'verify-token' or import the package that provides it.`

- Non-exported symbol
  - `handlers/user.vx:18:5: error: symbol 'create-session' is not exported from package 'auth'. Export it with (export [create-session]) in that package.`

- Invalid import form
  - `lib/util.vx:3:1: error: import requires package path
    Suggestion: use (import "fmt") or (import ["a" ["net/http" http]]).`

- Circular dependency
  - `error: circular dependency detected: a -(import at a/a.vx)-> b | b -(import at b/b.vx)-> c | c -> a`

### Tone and style
- Lowercase `error:` prefix after position (Go-style).
- Keep messages short; avoid stack traces or verbose AST dumps.
- Prefer domain terms over internal implementation details.


