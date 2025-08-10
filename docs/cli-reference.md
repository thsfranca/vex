# Vex CLI Reference

## Overview

The Vex CLI provides three main commands for working with Vex programs: `transpile`, `run`, and `build`. Built on a sophisticated transpiler architecture with advanced macro system, comprehensive Go interoperability, and clean code generation, all commands follow Go's performance characteristics and provide detailed error reporting.

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
- **Macro System**: Comprehensive macro expansion including defn macro
- **Symbol Table**: Advanced variable scoping and resolution
- **Go Interoperability**: Complete access to Go standard library
- **Error Handling**: Detailed error reporting for all phases
 - **Core Library**: Automatically includes `core.vx` from the working directory if present
- **Go Compilation**: Uses Go's native compiler for execution
- **Memory Management**: Leverages Go's garbage collector
  - **Package Discovery (MVP)**: Automatically discovers local Vex packages from the entry file, resolves imports, orders packages, and prevents circular dependencies (compile-time error on cycles). `vex.pkg` is supported for module root detection.
  - **Import arrays and aliases**: `(import ["fmt" ["net/http" http] ["encoding/json" json]])`. Calls use alias or package name: `(http/Get ...)`, `(json/Marshal ...)`, `(fmt/Println ...)`.
  - **Exports (MVP)**: Private-by-default. Declare public API with `(export [name1 name2 ...])` at package top-level. Cross-package access to non-exported symbols fails.

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
2. Register and expand macros (including defn) with comprehensive validation
3. Perform semantic analysis with symbol table management
4. Transpile to Go source with clean code generation
5. Compile Go source with `go build`
6. Execute resulting binary
7. Clean up temporary files

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
- **Native performance**: Full Go compilation optimizations
- **Cross-platform**: Can run on any platform Go supports
- **Small size**: Only includes necessary code
  - **Package Discovery (MVP)**: Includes all discovered local packages starting from the entry file; builds fail on circular dependencies. `vex.pkg` is supported for module root detection.
  - **Import arrays and aliases**: Aliased imports are respected in generated Go (`import http "net/http"`).
  - **Exports (MVP)**: Private-by-default. Public symbols must be listed in `(export [...])`; otherwise builds can fail when accessed from other packages.

#### Distribution
```bash
# The generated binary is completely standalone
./vex build -input webserver.vx -output webserver

# Can be distributed without any Vex installation
scp webserver user@server:/usr/local/bin/
# Remote server can run ./webserver without Vex
```

---

## Command Comparison

| Feature | `transpile` | `run` | `build` |
|---------|-------------|-------|---------|
| **Output** | Go source | Direct execution | Binary executable |
| **Use Case** | Inspection/Integration | Development/Testing | Distribution |
| **Speed** | Fast | Medium | Slow (compilation) |
| **Dependencies** | None | Go compiler | Go compiler |
| **Result** | .go file | Temporary execution | Standalone binary |

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

### Production
- Always use `build` for production deployments
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