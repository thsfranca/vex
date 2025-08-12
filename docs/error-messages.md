## Vex Compiler Error Message Conventions

Vex follows Go-style error messaging: concise, actionable, and easy to scan — optimized for both humans and AI tools.

### Format
- Default text format (Go-style prefix): `path/to/file.vx:LINE:COL: error: [CODE]: short-message`
- If file/position is unknown (e.g., cross-file cycle detection), omit them and include best-available context.
- Paths are relative to the module root detected by `vex.pkg` when available.

Optional detail lines (keep order):
- `Expected: …` (when applicable)
- `Got: …` (when applicable)
- `Offender: …` or position detail such as `at index N` or `at pair N` (when helpful)
- `Suggestion: …` (only when low-risk and concrete)

### Position information
- Include line and column whenever the error originates from a specific token or node.
- If an error spans multiple nodes (e.g., import graphs), include at least the first offending location and any edges that clarify the issue.

### Error codes (stable)
- Prefix all messages with a stable code: `[CATEGORY-SPECIFIC]`
- Examples (non-exhaustive):
  - Type system: `TYPE-UNDEFINED`, `TYPE-CONDITION`, `TYPE-IF-MISMATCH`, `TYPE-ARRAY-ELEMENT`, `TYPE-MAP-KEY`, `TYPE-MAP-VALUE`, `TYPE-NUMBER`, `TYPE-EQUALITY`, `TYPE-NOT`, `TYPE-BOOLEAN-ARGS`
  - Arity/arguments: `ARITY-ARGUMENTS`
  - Macros: `MACRO-RESERVED`, `MACRO-UNDEFINED`
  - Records: `RECORD-FIELD`, `RECORD-FIELD-TYPE`, `RECORD-NOMINAL`
  - Imports/packages: `IMPORT-SYNTAX`, `PACKAGE-NOT-EXPORTED`, `PACKAGE-CYCLE`
  - Function naming: `FN-NAMING` (enforces kebab-case naming convention)
  - Symbol naming: `SYMBOL-NAMING` (enforces kebab-case for all symbols)

### Suggesting fixes
Include a one-line suggestion only when it is obvious and safe:
- Undefined symbol: suggest defining it or importing the right package.
- Non-exported symbol: suggest exporting it, e.g., `Export it with (export [name]) in that package.`
- Invalid import usage: show the correct syntax quickly.
- Cycle detection: show the full chain and, when available, the file where each import edge was declared.

Suggestions must be:
- Short, concrete, and technically correct.
- Avoid guessing beyond common, low-risk fixes.

### Macro expansion context
- For errors inside expanded code, append: `Macro: NAME (defined at FILE:LINE:COL), expanded at FILE:LINE:COL` when the information is available.

### Examples
- Syntax error
  - `app/main.vx:12:7: error: [SYNTAX-EXPECT]: expected ')' but found 'EOF'`

- Undefined symbol
  - `services/auth/login.vx:34:3: error: [TYPE-UNDEFINED]: undefined symbol 'verify-token'
    Suggestion: define 'verify-token' or import the package that provides it.`

- Non-exported symbol
  - `handlers/user.vx:18:5: error: [PACKAGE-NOT-EXPORTED]: symbol 'create-session' is not exported from package 'auth'
    Suggestion: Export it with (export [create-session]) in that package.`

- Invalid import form
  - `lib/util.vx:3:1: error: [IMPORT-SYNTAX]: import requires package path
    Suggestion: use (import "fmt") or (import ["a" ["net/http" http]]).`

- Circular dependency
  - `error: [PACKAGE-CYCLE]: a -(import at a/a.vx)-> b | b -(import at b/b.vx)-> c | c -> a`

- Type mismatch in if branches
  - `file.vx:10:5: error: [TYPE-IF-MISMATCH]: branch types differ
    Expected: type(then) == type(else)
    Got: then=int, else=string
    Suggestion: make both branches the same type or add explicit cast.`

- Array element mismatch
  - `file.vx:7:9: error: [TYPE-ARRAY-ELEMENT]: array elements must share a type
    First mismatch at index 2
    Expected: string
    Got: int`

- Map key mismatch
  - `file.vx:15:3: error: [TYPE-MAP-KEY]: map keys have incompatible types
    First mismatch at pair 3
    Expected: string
    Got: int`

- Function naming convention violation
  - `file.vx:8:1: error: [FN-NAMING]: function names must use kebab-case
    Name: my_function
    Suggestion: use kebab-case with dashes (e.g., 'my-function')`

- Symbol naming convention violation
  - `file.vx:5:1: error: [SYMBOL-NAMING]: symbol names must use kebab-case
    Name: user_data
    Suggestion: use kebab-case with dashes (e.g., 'user-data')`

### Tone and style
- Lowercase `error:` prefix after position (Go-style).
- Keep messages short; avoid stack traces or verbose AST dumps.
- Prefer domain terms over internal implementation details.

### Structured output
- A future `--machine` flag will emit JSON alongside text, e.g.:
  - `{ "code":"TYPE-IF-MISMATCH","file":"main.vx","line":12,"col":3,"expected":"same-type","got":{"then":"int","else":"string"},"suggestion":"make-branches-same-type" }`

### Implementation notes
- Diagnostics live in `internal/transpiler/diagnostics` with:
  - Stable codes (e.g., `TYPE-*`, `ARITY-*`, `MACRO-*`, `IMPORT-*`, `PACKAGE-*`)
  - A catalog of canonical short messages (templates)
  - Renderers for text and JSON
- Analyzer currently emits diagnostics for conditions, branch mismatches, arrays, maps, and macro/arity issues.


