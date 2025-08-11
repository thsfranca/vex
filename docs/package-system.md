## Vex Package System (MVP)

### Overview
Vex supports directory-based packages, inspired by Go. Each directory containing `.vx` files represents a package. The CLI (`transpile`, `run`, `build`) automatically discovers local packages referenced by the entry file, concatenates sources in dependency order, and compiles them together. Circular dependencies are detected at compile-time and fail the build with a clear error.

Status in current branch:
- Directory-based local packages: implemented
- Import parsing: strings, arrays, and alias pairs: implemented
- `vex.pkg` module root detection: implemented
- Circular dependency detection: implemented (build fails with clear cycle chain)
- Exports: parsed; analyzer enforces exports for namespaced calls via package schemes; codegen double-checks in `transpile`/`run`

### Imports
- Basic form (string): `(import "fmt")`
- Array form (multiple): `(import ["a" "fmt"])`
- Alias pair (path + alias): `(import [["net/http" http] ["encoding/json" json]])`
- Combined: `(import ["a" "fmt" ["net/http" http]])`

Calls use alias or package name:
- `(http/Get "...")` → `http.Get(...)`
- `(json/Marshal x)` → `json.Marshal(x)`
- `(fmt/Println x)` → `fmt.Println(x)`

Local packages referenced by relative import paths (e.g., `"a"`, `"utils/math"`) are discovered and compiled into the program. Their import paths are not emitted as Go imports.

### Module Root (vex.pkg)
The resolver detects the module root by walking up from the entry file’s directory to find a `vex.pkg` file. When found, that directory is used as the module root for resolving local packages. If not found, the starting directory is used.

Minimal `vex.pkg`:
```
module myapp
```
(The schema can evolve. For MVP, only presence is required to mark the root.)

### External Package Management
- `run`: builds a temporary Go module and compiles the generated code; external Go imports must be available to the Go toolchain
- `build`: generates a temporary `go.mod`, detects external modules from imports, runs `go mod tidy`, and produces a binary

Keep the import path space unambiguous between local packages (directories under the module root) and external Go packages (module paths like `github.com/...`).

Vex package dependencies and versioning are planned for `vex.pkg` in a future step (`require` blocks and a lock file).

### Exports
Declare exports at the top of a package file:
```
(export [add sum-three])
```

Current behavior:
- Resolver parses exports for local packages
- Code generation enforces that only exported symbols of local packages are callable from other packages in `transpile` and `run` flows; `build` parity is being aligned
- Full analyzer-level enforcement is planned

### Circular Dependencies
The resolver builds a dependency graph from imports and fails compilation if a cycle is detected. The error message shows the cycle chain and edge file locations when available, e.g.:

```
error: circular dependency detected: a -(import at a/a.vx)-> b | b -(import at b/b.vx)-> c | c -> a
```

### Example
```
project/
  vex.pkg
  a/a.vx
  b/b.vx
  main.vx
```

`b/b.vx`:
```
(export [add])
(defn add [x y] (+ x y))
```

`a/a.vx`:
```
(import ["b"]) ; local package import
(export [sum-three])
(defn sum-three [x y z] (+ (b/add x y) z))
```

`main.vx`:
```
(import ["a" "fmt"]) 
(def result (a/sum-three 1 2 3))
(fmt/Println result)
```

Run:
```
./vex run -input main.vx
```

### Notes
- One package per directory
- Local packages are discovered recursively starting from the entry file
- Go imports and local packages can be mixed in the same `(import [...])` list

### Package Schemes
Resolver computes type schemes (`PkgSchemes`) for exported symbols of local packages by expanding macros (loading core macros) and running the analyzer per package. Analyzer uses these schemes across package boundaries to type namespaced calls and enforce exports.

### Standard Library Macros
The transpiler loads core macros; in CI/tests it prefers `core/core.vx` under the module root. Built-in stdlib packages under `stdlib/vex/` can provide additional macros (e.g., `conditions`, `collections`).

