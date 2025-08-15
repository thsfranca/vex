# Vex CLI Reference
## Quickstart Tutorial

Follow these steps to clone, build, write code, and run Vex programs.

### Prerequisites

- Go 1.21+
- Git

### 1) Clone the repository

```bash
git clone https://github.com/thsfranca/vex.git
cd vex
```

### 2) Build the CLI

```bash
go build -o vex cmd/vex-transpiler/main.go
./vex --help
```

You should see the commands: `transpile`, `run`, and `build`.

### 3) Write your first program

Create `hello.vx`:

```vex
(import "fmt")
(def greeting "Hello, Vex!")
(fmt/Println greeting)
```

### 4) Run it

```bash
./vex run -input hello.vx
```

Expected output:

```
Hello, Vex!
```

### 5) Transpile to Go (optional)

```bash
./vex transpile -input hello.vx -output hello.go
cat hello.go
```

### 6) Build a binary (optional)

```bash
./vex build -input hello.vx -output hello
./hello
```

### 7) Playground example

Create `playground.vx`:

```vex
(import "fmt")
(def sum (+ 5 3))
(def product (* 4 7))
(def items [1 "two" 3])
(defn add [x y] (+ x y))
(def result (add sum product))
(fmt/Printf "sum=%d, product=%d, result=%d\n" sum product result)
(fmt/Println items)
```

Run:

```bash
./vex run -input playground.vx
```


## Overview

The Vex CLI provides complete commands for working with Vex programs: `transpile`, `run`, `build`, and `test`. Built on an advanced transpiler architecture with complete Hindley-Milner type system, comprehensive macro system, full package discovery, and optimized code generation, all commands provide excellent performance and detailed error reporting.

## Global Options

### Help
```bash
./vex --help              # Show all available commands
./vex [command] --help    # Show help for specific command
```

### Version Information
```bash
./vex --version           # Show Vex version
```

## Commands

### `transpile` - Convert Vex to Go

Transpiles Vex source code to Go source code for inspection or integration.

#### Syntax
```bash
./vex transpile -input <input-file> -output <output-file> [options]
```

#### Arguments
- `-input <file>` (required) - Input Vex source file (.vx)
- `-output <file>` (optional) - Output Go source file (.go). If omitted, prints to stdout
- `-verbose` (optional) - Prints progress information to stderr

#### Examples
```bash
# Basic transpilation
./vex transpile -input hello.vx -output hello.go

# Transpile with verbose output
./vex transpile -input complex.vx -output complex.go -verbose

# Check generated Go code
./vex transpile -input app.vx -output app.go && cat app.go
```

#### Generated Go Structure
```go
package main

import (
    // Auto-generated imports based on Vex (import) statements
)

func main() {
    // Transpiled Vex code
}
```

#### Use Cases
- **Code inspection**: Understand what Go code is generated
- **Integration**: Use transpiled Go in larger Go projects
- **Learning**: See how Vex constructs map to Go
- **Debugging**: Analyze transpilation issues

---

### `run` - Compile and Execute

Compiles and executes Vex programs directly without creating intermediate files.

#### Syntax
```bash
./vex run -input <input-file> [options]
```

#### Arguments
- `-input <file>` (required) - Input Vex source file (.vx)
- `-verbose` (optional) - Prints progress information to stderr

#### Advanced Features
- **Complete Macro System**: Full stdlib with 7 packages (core, test, collections, conditions, flow, threading, bindings)
- **Hindley-Milner Type System**: Complete Algorithm W inference with unification and error reporting
- **Complete Go Interoperability**: Full access to Go standard library with aliases and module detection
- **Structured Error Handling**: Detailed error reporting with stable codes (VEX-TYP-*) for all phases
- **Complete Package System**: Automatically discovers local Vex packages, resolves dependencies, enforces exports, detects cycles
- **Advanced Import System**: Support for aliases `(import ["fmt" ["net/http" http] ["encoding/json" json]])`
- **Complete Export System**: Private-by-default with `(export [name1 name2 ...])` enforcement
- **Go Compilation**: Optimized Go code generation with native compiler integration
- **Module Management**: Complete `vex.pkg` support with dependency resolution

#### Examples
```bash
# Simple execution
./vex run -input hello.vx

# Run with verbose output
./vex run -input app.vx -verbose

# Quick testing during development
echo '(import "fmt") (fmt/Println "Test")' > test.vx
./vex run -input test.vx

# Run program with function definitions
echo '(import "fmt") (defn greet [name] (fmt/Printf "Hello %s!\n" name)) (greet "World")' > func.vx
./vex run -input func.vx

# Run program with macros
echo '(import "fmt") (macro log [msg] (fmt/Printf "[LOG] %s\n" msg)) (log "Starting")' > macro.vx
./vex run -input macro.vx
```

Note about external Go modules:
- `run` does not create a `go.mod`. Programs that import third-party Go modules may fail to build under `run`. Use `build` for those cases.

#### Execution Process
1. Parse Vex source into AST using ANTLR parser
2. Discover and resolve package dependencies with cycle detection
3. Load and expand macros from stdlib packages with comprehensive validation
4. Perform Hindley-Milner type analysis with Algorithm W inference
5. Generate optimized Go source with complete type information
6. Compile Go source with `go build` and dependency management
7. Execute resulting binary
8. Clean up temporary files

#### Performance
- **Fast startup**: Direct execution without manual compilation steps
- **Go performance**: Native Go execution speed
- **Memory efficient**: Automatic cleanup of temporary files

---

### `build` - Create Binary Executable

Creates standalone binary executables from Vex programs.

#### Syntax
```bash
./vex build -input <input-file> -output <output-binary> [options]
```

#### Arguments
- `-input <file>` (required) - Input Vex source file (.vx)
- `-output <file>` (optional) - Output binary executable. Defaults to input filename without extension
- `-verbose` (optional) - Prints progress information to stderr

#### Examples
```bash
# Build standalone binary
./vex build -input server.vx -output server
./server  # Run the binary

# Build with specific name
./vex build -input calculator.vx -output calc
./calc

# Build for distribution
./vex build -input myapp.vx -output myapp
chmod +x myapp  # Ensure executable permissions
```

#### Binary Characteristics
- **Self-contained**: No dependencies on Vex runtime
- **Native performance**: Full Go compilation optimizations with HM type information
- **Cross-platform**: Can run on any platform Go supports
- **Optimized size**: Only includes necessary code with dead code elimination
- **Complete Package System**: Includes all discovered local packages with full dependency resolution
- **Advanced Import Handling**: Aliased imports fully supported in generated Go (`import http "net/http"`)
- **Complete Export System**: Comprehensive private-by-default enforcement with export validation

#### Distribution
```bash
# The generated binary is completely standalone
./vex build -input webserver.vx -output webserver

# Can be distributed without any Vex installation
scp webserver user@server:/usr/local/bin/
# Remote server can run ./webserver without Vex
```

---

### `test` - Discover and Run Tests

Comprehensive testing framework for Vex programs with automatic test discovery, macro-based assertions, and test coverage reporting.

#### Syntax
```bash
./vex test [-dir <path>] [-verbose] [-coverage] [-coverage-out <file>] [-enhanced-coverage] [-failfast] [-pattern <pattern>] [-timeout <duration>]
```

#### Arguments
- `-dir <path>` (optional) - Directory to search for test files (default: current directory)
- `-verbose` (optional) - Enable detailed output including test execution details
- `-coverage` (optional) - Generate test coverage report per package
- `-coverage-out <file>` (optional) - Write coverage data to JSON file for CI integration
- `-enhanced-coverage` (optional) - Generate advanced function-level coverage analysis with line, branch, and quality metrics
- `-failfast` (optional) - Stop execution on the first test failure
- `-pattern <pattern>` (optional) - Run only tests whose file paths contain the pattern
- `-timeout <duration>` (optional) - Maximum execution time per test (default: 30s)

#### Test Discovery and Validation
- Recursively finds files matching `*_test.vx` or `*_test.vex` patterns
- Skips hidden directories (`.`), `node_modules`, `bin`, `gen`, `vendor`, `coverage`
- Applies pattern filtering if specified with `-pattern`
- **Complete test validation**: Only allows code inside `(deftest ...)` declarations
- **Strict validation**: Files with code outside `deftest` blocks fail validation with clear error messages
- **Complete test pipeline**: Each valid test file is transpiled with full HM type checking and executed independently

#### Test Execution Pipeline
1. **Test File Validation** - Ensures only `(deftest ...)` declarations exist with comprehensive validation
2. **Package Resolution** - Complete dependency resolution with cycle detection
3. **Macro Expansion** - Applies complete stdlib test macros and user-defined macros
4. **Hindley-Milner Type Analysis** - Full type checking with Algorithm W inference
5. **Code Generation** - Optimized Go code generation with type information
6. **Compilation** - Builds temporary Go binary with timeout protection and dependency management
7. **Execution** - Runs only the `deftest` blocks with configurable timeout
8. **Coverage Analysis** - Per-package coverage analysis with visual indicators
9. **Cleanup** - Removes temporary files automatically

#### Built-in Test Macros
Automatically available in all test files:

- **`assert-eq actual expected "message"`** - Equality assertion
  - Prints "✅ message" for assertions
  - Simple verification that actual equals expected
- **`deftest "name" (body...)`** - Named test definition
  - Prints "🧪 RUN: name" before execution
  - Executes test body
  - Prints "✅ PASS: name" after successful completion
  - Supports single statements or multiple assertions

#### Test Coverage Analysis

**Basic Coverage** (with `-coverage` flag):
- Analyzes test coverage per package (directory-based)
- Reports percentage of source files that have corresponding test files
- Provides visual indicators: ✅ (80%+), 📈 (50-79%), ⚠️ (<50%), ❌ (0%)
- Shows overall project coverage statistics

**Enhanced Coverage** (with `-enhanced-coverage` flag):
- **Function-Level Analysis**: Tracks which specific functions (`defn`, `defmacro`, `def+fn`, `macro`) are tested
- **Line-Level Precision**: Analyzes coverage of individual code lines, excluding comments and imports
- **Branch Coverage**: Detects conditional branches (`if`, `when`, `unless`, `cond`) and tracks true/false path coverage
- **Test Quality Scoring**: Evaluates test quality with metrics including:
  - Assertion density (assertions per test)
  - Edge case coverage (boundary values, nil/empty inputs)
  - Test method diversity (variety of assertion types)
  - Naming quality (descriptive test names)
- **Smart Suggestions**: Provides actionable recommendations for improving test coverage and quality
- **Red Flag Detection**: Identifies problematic patterns like low assertion density or missing edge cases
- **Multi-Dimensional Reports**: Shows function, line, branch, and quality coverage simultaneously
- **CI/CD Integration**: Generates detailed JSON reports with specific untested functions and improvement suggestions

#### Examples

**Valid Test File** (`math_test.vx`):
```vex
;; Only deftest declarations are allowed in test files
(import ["fmt" "test"])

;; Simple test with assertion
(deftest "addition"
  (do
    (fmt/Println "Testing addition")
    (assert-eq (+ 1 2) 3 "basic addition")))

;; Test with multiple assertions
(deftest "multiplication"
  (do
    (fmt/Println "Testing multiplication")
    (assert-eq (* 4 5) 20 "basic multiplication")
    (assert-true (> (* 4 5) 0) "result is positive")))

;; Test with string operations
(deftest "string-operations"
  (do
    (fmt/Println "Testing string operations")
    (assert-eq "hello" "hello" "string equality")
    (assert-false (= "hello" "world") "string inequality")))

;; Test with custom output
(deftest "custom-test"
  (do
    (fmt/Println "This test demonstrates custom logic")
    (def result (+ 10 5))
    (assert-eq result 15 "calculation result")))
```

**Invalid Test File** (will fail validation):
```vex
(import "fmt")

;; This code outside deftest will cause validation to fail
(defn helper-function [x] (* x 2))
(fmt/Println "This print statement is not allowed!")

(deftest "my-test"
  (assert-eq (+ 1 1) 2 "addition"))
```

**Running Tests**:
```bash
# Run all tests
./vex test

# Run tests with detailed output
./vex test -verbose

# Run tests with coverage report
./vex test -coverage

# Test specific directory
./vex test -dir ./src -verbose -coverage

# Run only calculator tests
./vex test -pattern "calculator"

# Stop on first failure with custom timeout
./vex test -failfast -timeout 10s

# Combined options with coverage file output for CI
./vex test -dir ./src -pattern "api" -coverage -coverage-out coverage.json -verbose -timeout 5s

# Enhanced coverage analysis (real execution-based coverage)
./vex test -enhanced-coverage

# Enhanced coverage with JSON output for CI/CD pipelines
./vex test -enhanced-coverage -coverage-out advanced-coverage.json

# Comprehensive testing with all coverage features
./vex test -dir ./stdlib -enhanced-coverage -coverage-out stdlib-coverage.json -verbose
```

**Sample Output**:
```
🧪 Running Vex tests in examples/test-demo
⏱️  Timeout: 30s

▶ examples/test-demo/math_test.vx
   Found 4 deftest declaration(s)
✅ PASS: examples/test-demo/math_test.vx (1.162s)
🧪 RUN: addition
✅ basic addition
✅ PASS: addition
🧪 RUN: multiplication  
✅ basic multiplication
✅ PASS: multiplication
🧪 RUN: string-operations
✅ string equality
✅ PASS: string-operations
🧪 RUN: custom-test
This test demonstrates custom logic
✅ PASS: custom-test

📊 Generating test coverage report...

📋 Test Coverage Report
─────────────────────────
✅ examples/test-demo: 100.0% (1/1 files)
─────────────────────────
📊 Overall: 100.0% (1/1 packages have tests)

🚀 Generating Enhanced Coverage Analysis...

📊 Enhanced Coverage Report
═══════════════════════════
📄 Coverage detected: /tmp/math_test_test_bin.go
📈 Overall Coverage:
   Execution-Based: 100.0% (1/1 files executed)
   Profile Sources: 1 coverage profile(s)
   Data Quality: REAL execution data ✅

💡 Coverage Insights:
   Coverage precision: REAL execution data (100% accurate)
   Data source: Go runtime instrumentation

🏁 Test Summary
─────────────────────────
📁 Total: 1 tests
✅ Passed: 1
⏱️  Duration: 1.162s
🎉 Success rate: 100.0%
```

#### Error Types and Status Codes
- **PASS**: Test executed successfully (exit code 0)
- **FAIL**: Test execution failed (runtime error)
- **TRANSPILE_ERROR**: Test validation, package resolution, or transpilation failed
  - Includes validation errors for code outside `deftest` declarations
- **BUILD_ERROR**: Go compilation failed on generated code
- **TIMEOUT**: Test exceeded the specified timeout duration
- **SKIP**: Test was skipped (future feature)

#### Test Validation Errors
- **Invalid test structure**: Code outside `(deftest ...)` declarations
- **Missing deftest**: Test files must contain at least one `deftest` block
- **Unclosed deftest**: Parentheses mismatch in `deftest` declarations

#### Exit Codes
- **0**: All tests passed
- **1**: One or more tests failed or errors occurred

#### Integration with CI/CD
```bash
# CI/CD example
./vex test -coverage
if [ $? -eq 0 ]; then
    echo "✅ All tests passed"
    # Deploy or continue pipeline
else
    echo "❌ Tests failed"
    exit 1
fi
```

---

## Command Comparison

| Feature | `transpile` | `run` | `build` | `test` |
|---------|-------------|-------|---------|--------|
| **Output** | Go source | Direct execution | Binary executable | Test report + Coverage |
| **Use Case** | Inspection/Integration | Development | Distribution | Testing/QA |
| **Speed** | Fast | Medium | Slow (compilation) | Medium |
| **Dependencies** | None | Go compiler | Go compiler | Go compiler |
| **Result** | .go file | Temporary execution | Standalone binary | Exit code/coverage % |
| **Coverage Analysis** | ❌ | ❌ | ❌ | ✅ |

## File Handling

### Input Files
- **Extension**: `.vx` (recommended) or any text file
- **Encoding**: UTF-8
- **Size**: No practical limits (memory dependent)

### Output Files
- **`transpile`**: `.go` extension recommended
- **`build`**: No extension needed (platform-specific executable)

### Temporary Files
- **`run`**: Creates and cleans up temporary Go files automatically
- **`build`**: Creates and cleans up temporary Go files automatically
- **Location**: System temporary directory

## Error Handling

### Common Errors

#### Syntax Errors
```bash
$ ./vex run -input bad.vx
Error: Parse error at line 3, column 5
Expected ')' but found 'EOF'
```

#### Missing Files
```bash
$ ./vex transpile -input missing.vx -output out.go
Error: Input file 'missing.vx' not found
```

#### Permission Errors
```bash
$ ./vex build -input app.vx -output /usr/bin/app
Error: Permission denied writing to '/usr/bin/app'
```

### Error Codes
- **0**: Success
- **1**: General error (syntax, file not found, etc.)
- **2**: Compilation error (Go compilation failed)
- **3**: Permission error (file system access)

### HM Typing Diagnostics
The compiler uses Hindley–Milner (HM) type inference with strict diagnostics. Common diagnostic codes:

- `VEX-TYP-UNDEF`: Unknown identifier. Define the symbol or import the correct package.
- `VEX-TYP-COND`: If-condition must be boolean. Adjust the condition to return `bool`.
- `VEX-TYP-EQ`: Equality arguments differ in type. Make both sides the same type.
- `VEX-TYP-ARRAY-ELEM`: Array elements have inconsistent types. Make all elements the same type.
- `VEX-TYP-MAP-KEY` / `VEX-TYP-MAP-VAL`: Map keys/values have inconsistent types across pairs.
- `VEX-TYP-REC-NOMINAL`: Nominal record mismatch (e.g., `A` used where `B` is required).

Example (CLI stderr snippet):

```
path/to/file.vx:12:3: error: [VEX-TYP-COND]: if condition must be bool
Expected: bool
Got: number
```

## Advanced Usage

### Integration with Make
```makefile
# Makefile
build: main.vx
	./vex build -input main.vx -output myapp

run: main.vx
	./vex run -input main.vx

clean:
	rm -f myapp *.go

.PHONY: build run clean
```

### Development Workflow
```bash
# Edit-test cycle
vim myprogram.vx
./vex run -input myprogram.vx

# Build for testing
./vex build -input myprogram.vx -output myprogram-test
./myprogram-test

# Inspect generated Go
./vex transpile -input myprogram.vx -output debug.go
```

### CI/CD Integration
```bash
# In CI scripts
./vex build -input src/main.vx -output dist/application
test -x dist/application  # Verify binary was created
```

## Performance Tips

### Development
- Use `run` for quick iteration and testing
- Use `transpile` to understand generated Go code
- Use `build` only when you need a distributable binary

### Deployment
- Always use `build` for final deployments
- The generated binaries have full Go performance characteristics
- No Vex runtime overhead in final executables

### Optimization
- Vex programs benefit from Go's built-in optimizations
- Consider Go build flags for additional optimization if needed
- Profile using Go's standard profiling tools on generated code

## Troubleshooting

### Build Issues
```bash
# Verify Vex installation
./vex --version

# Check Go installation
go version

# Test with minimal program
echo '(import "fmt") (fmt/Println "OK")' > test.vx
./vex run -input test.vx
```

### Runtime Issues
```bash
# Check file permissions
ls -la *.vx

# Verify syntax with transpile first
./vex transpile -input problem.vx -output debug.go
cat debug.go  # Inspect generated Go
```

### Performance Issues
```bash
# Profile generated binaries
./vex build -input app.vx -output app
go tool pprof ./app  # Use Go's profiling tools
```

For more help, see [Getting Started](getting-started.md) or [Troubleshooting Guide](troubleshooting.md).
Also see [Package System](package-system.md) for multi-package projects, import arrays/aliases, `vex.pkg`, and exports.