# Vex Troubleshooting Guide

## Common Issues and Solutions

This guide covers the most frequent problems encountered when using Vex and their solutions.

## Installation and Setup Issues

### Build Errors

#### Problem: `make build-transpiler` fails
```
Error: go: cannot find main module
```

**Solution:**
```bash
# Ensure you're in the project root
cd /path/to/vex
ls go.mod  # Should exist

# Try direct build
go build -o vex cmd/vex-transpiler/main.go
```

#### Problem: ANTLR not found
```
Error: antlr4 command not found
```

**Solution:**
```bash
# Install ANTLR4 (choose your platform)

# macOS with Homebrew
brew install antlr

# Ubuntu/Debian
sudo apt-get install antlr4

# Manual installation
wget https://www.antlr.org/download/antlr-4.13.1-complete.jar
export CLASSPATH=".:antlr-4.13.1-complete.jar:$CLASSPATH"
alias antlr4='java -jar antlr-4.13.1-complete.jar'
```

#### Problem: Go version too old
```
Error: module requires go 1.21
```

**Solution:**
```bash
# Check current version
go version

# Install Go 1.21+ from https://golang.org/dl/
# Or update using your package manager
```

### Permission Issues

#### Problem: Cannot create executable
```
Error: Permission denied writing to 'vex'
```

**Solution:**
```bash
# Build in current directory (not system paths)
go build -o ./vex cmd/vex-transpiler/main.go

# Or fix permissions
sudo chown $USER:$USER .
chmod 755 .
```

## Runtime Errors

### Syntax Errors

#### Problem: Unmatched parentheses
```vex
(def x (+ 1 2
```
```
Error: Parse error at line 1, column 14
Expected ')' but found 'EOF'
```

**Solution:**
- **Check parentheses**: Every `(` needs a matching `)`
- **Use editor**: VSCode with Vex extension shows matching brackets
- **Test incremental**: Build up complex expressions step by step

#### Problem: Invalid symbols
```vex
(def 123invalid "value")
```
```
Error: Invalid symbol '123invalid'
```

**Solution:**
```vex
;; ✅ Valid symbols
(def valid-name "value")
(def another-name 42)
(def is-valid? true)

;; ❌ Invalid symbols  
(def 123invalid "no")     ; Can't start with number
(def "string-name" 42)    ; Can't use strings as names
```

#### Problem: Missing imports
```vex
(fmt/Println "Hello")
```
```
Error: Unknown namespace 'fmt'
```

**Solution:**
```vex
;; ✅ Always import before using
(import "fmt")
(fmt/Println "Hello")
```

### Transpilation Issues

#### Problem: Generated Go doesn't compile
```
Error: Go compilation failed
./temp.go:5:2: undefined: someFunction
```

**Solution:**
```bash
# Check generated Go code
./vex transpile -input problem.vx -output debug.go
cat debug.go

# Look for:
# - Missing imports
# - Undefined functions
# - Type mismatches
```

#### Problem: Macro expansion errors
```vex
(macro greet [name] (fmt/Printf "Hello %s" name))
(greet)  ; Missing argument
```
```
Error: macro greet expects 1 arguments, got 0
```

**Solution:**
```vex
;; ✅ Provide required arguments
(macro greet [name] (fmt/Printf "Hello %s" name))
(greet "World")

;; ✅ Use defn macro for functions instead of macro
(defn greet [name] (fmt/Printf "Hello %s" name))
(greet "World")
```

#### Problem: Defn macro issues
```vex
(defn add [x] (+ x y))  ; Undefined variable y
```
```
Error: undefined symbol: y
```

**Solution:**
```vex
;; ✅ Ensure all variables are defined or passed as parameters
(defn add [x y] (+ x y))

;; ✅ Or define the variable separately
(def y 10)
(defn add [x] (+ x y))
```

## Performance Issues

### Slow Compilation

#### Problem: `vex run` or `vex build` is slow

**Causes & Solutions:**

1. **Large programs**:
   ```bash
   # Break into smaller files
   # Use `vex transpile` to inspect generated Go
   ./vex transpile -input large.vx -output large.go
   wc -l large.go  # Check generated line count
   ```

2. **Complex macros**:
   ```vex
   ;; Avoid deeply nested macro expansions
   ;; Prefer simple, focused macros
   ```

3. **Go compilation overhead**:
   ```bash
   # Use go build cache
   export GOCACHE=/tmp/go-cache
   
   # For development, use `run` instead of `build`
   ./vex run -input app.vx
   ```

### Memory Issues

#### Problem: Out of memory during compilation

**Solutions:**
```bash
# Increase available memory
export GOGC=100  # Default garbage collection

# For large programs, build incrementally
./vex transpile -input app.vx -output app.go
go build app.go  # Use go directly with memory tuning
```

## Development Workflow Issues

### Debugging Generated Go

#### Problem: Runtime errors in generated code

**Debugging process:**
```bash
# 1. Transpile to inspect Go code
./vex transpile -input app.vx -output debug.go

# 2. Check generated Go syntax
go fmt debug.go  # Will fail if syntax errors

# 3. Build manually to see Go errors
go build debug.go

# 4. Run with Go tooling
go run debug.go

# 5. Use Go debugger if needed
dlv debug debug.go
```

#### Problem: Symbol resolution errors
```vex
(defn helper [x] (* x 2))
(defn main [] (helper 5))  ; May fail if helper not properly resolved
```

**Solution:**
```bash
# Check symbol table in transpiled Go
./vex transpile -input problem.vx -output debug.go
grep -n "helper" debug.go  # Check function generation

# Ensure proper function ordering in Vex
(defn helper [x] (* x 2))
(defn main [] (helper 5))
```

#### Problem: Macro expansion debugging
```vex
(macro debug-print [var] 
  (fmt/Printf "DEBUG: %s = %v\n" var var))
(debug-print x)  ; May expand incorrectly
```

**Solution:**
```bash
# Check macro expansion in generated Go
./vex transpile -input macro-test.vx -output debug.go
# Look for macro expansion comments in output
```

### Testing Strategies

#### Problem: Hard to test Vex programs

**Solutions:**
```bash
# 1. Test small parts incrementally
echo '(+ 1 2)' > test.vx
./vex run -input test.vx

# 2. Test function definitions
echo '(import "fmt") (defn add [x y] (+ x y)) (fmt/Println (add 3 4))' > func-test.vx
./vex run -input func-test.vx

# 3. Use transpilation for unit testing
./vex transpile -input logic.vx -output logic.go
# Then test logic.go with Go testing tools

# 4. Test macros separately
echo '(import "fmt") (macro log [msg] (fmt/Printf "[LOG] %s\n" msg)) (log "test")' > macro-test.vx
./vex run -input macro-test.vx

# 5. Test complex programs step by step
./vex transpile -input complex.vx -output debug.go
go fmt debug.go  # Check syntax
go build debug.go  # Check compilation
```

## Integration Issues

### Go Library Integration

#### Problem: Go function not accessible
```vex
(import "strings")
(strings/InvalidFunction "test")
```
```
Error: undefined: strings.InvalidFunction
```

**Solution:**
```bash
# Check Go documentation
go doc strings | grep -i function

# Use correct function names
(strings/ToUpper "hello")  # Not InvalidFunction
```

#### Problem: Type mismatches with Go
```vex
(import "strconv")
(strconv/Atoi 123)  ; Should be string, not number
```

**Solution:**
```vex
;; ✅ Use correct types
(import "strconv")
(strconv/Atoi "123")  ; String argument

;; Check Go function signatures
;; func Atoi(s string) (int, error)
```

### VSCode Extension Issues

#### Problem: No syntax highlighting

**Solutions:**
```bash
# Install extension
cd vscode-extension
./quick-install.sh

# Verify installation
code --list-extensions | grep vex

# Check file association
# Ensure files have .vx extension
```

#### Problem: Extension errors

**Solutions:**
```bash
# Rebuild extension
cd vscode-extension
npm install
./quick-install.sh

# Check VSCode logs
# View > Output > Select "Vex Language" from dropdown
```

## Error Message Reference

### Parser Errors
| Error | Meaning | Solution |
|-------|---------|----------|
| `Expected ')' but found 'EOF'` | Unclosed parenthesis | Add missing `)` |
| `Invalid symbol` | Illegal identifier | Use valid symbol syntax |
| `Unexpected token` | Syntax error | Check Vex grammar rules |

### Transpilation Errors
| Error | Meaning | Solution |
|-------|---------|----------|
| `Unknown namespace` | Missing import | Add `(import "package")` |
| `Macro expects N arguments` | Wrong macro usage | Check macro definition |
| `Go compilation failed` | Generated Go invalid | Check transpiled output |

### Runtime Errors
| Error | Meaning | Solution |
|-------|---------|----------|
| `File not found` | Input file missing | Verify file path |
| `Permission denied` | Access error | Check file permissions |
| `Command not found` | Missing binary | Build vex transpiler |

## Getting Help

### Debug Information to Collect

When reporting issues, include:

```bash
# 1. Version information
./vex --version
go version

# 2. Minimal reproducing example
cat > minimal.vx << EOF
(your minimal example here)
EOF

# 3. Error output
./vex run -input minimal.vx 2>&1

# 4. Generated Go (if applicable)
./vex transpile -input minimal.vx -output debug.go
cat debug.go
```

### Resources

- **Documentation**: Check [docs/](.) for complete guides
- **Examples**: See [examples/valid/](../examples/valid/) for working code
- **Grammar**: Reference [grammar-reference.md](grammar-reference.md)
- **Known Bugs**: Check [known-bugs.md](known-bugs.md) for tracked issues and workarounds
- **GitHub Issues**: [Report bugs](https://github.com/thsfranca/vex/issues)
- **AI Reference**: [ai-quick-reference.md](ai-quick-reference.md) for structured help

### Community Support

1. **Check existing issues** on GitHub
2. **Search documentation** for similar problems  
3. **Create minimal reproduction** case
4. **Report with full context** (versions, environment, code)

### Self-Help Checklist

Before asking for help:

- [ ] Verified Vex and Go versions are compatible
- [ ] Tested with minimal example
- [ ] Checked generated Go code with `transpile`
- [ ] Reviewed error messages carefully
- [ ] Searched existing documentation
- [ ] Tried suggested solutions from this guide

Most issues can be resolved by carefully reading error messages and checking the generated Go code! 🛠️