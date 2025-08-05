# Vex Transpiler Performance Optimizations

## Critical Fixes (Immediate Impact)

### 1. Remove Debug Output
```go
// Replace this:
t.output.WriteString(fmt.Sprintf("// DEBUG: Program has %d children\n", len(children)))

// With conditional debug (use build tags):
// +build debug
func (t *Transpiler) debugf(format string, args ...interface{}) {
    t.output.WriteString(fmt.Sprintf("// DEBUG: "+format+"\n", args...))
}

// +build !debug  
func (t *Transpiler) debugf(format string, args ...interface{}) {}
```

### 2. Optimize String Building
```go
// In expressions.go, replace string concatenation:
func buildArithmeticExpression(op string, args []string) string {
    var builder strings.Builder
    builder.WriteString("(")
    builder.WriteString(args[0])
    builder.WriteString(" " + op + " ")
    builder.WriteString(args[1])
    builder.WriteString(")")
    
    for i := 2; i < len(args); i++ {
        result := builder.String()
        builder.Reset()
        builder.WriteString("(")
        builder.WriteString(result)
        builder.WriteString(" " + op + " ")
        builder.WriteString(args[i])
        builder.WriteString(")")
    }
    return builder.String()
}
```

### 3. Pre-allocate Slices
```go
// Replace:
args := make([]string, 0)

// With:
args := make([]string, 0, childCount-3) // Estimate capacity
```

### 4. Cache Standard Library Map
```go
// At package level:
var standardLibs = map[string]bool{
    "fmt": true, "os": true, "strings": true, // ...
}

// In isThirdPartyModule:
func (t *Transpiler) isThirdPartyModule(importPath string) bool {
    return !standardLibs[importPath] && !strings.HasPrefix(importPath, "github.com/thsfranca/vex")
}
```

### 5. Optimize Macro Expansion
```go
// Instead of creating new transpiler, use expression evaluation:
func (t *Transpiler) expandMacroOptimized(macro *Macro, args []string) {
    // Simple parameter substitution without full parsing
    expandedBody := macro.Body
    for i, param := range macro.Params {
        expandedBody = strings.ReplaceAll(expandedBody, param, args[i])
    }
    
    // Only parse if truly necessary
    if strings.HasPrefix(expandedBody, "(") {
        // Use existing evaluateExpression instead of new transpiler
        result := t.evaluateSimpleExpression(expandedBody)
        t.output.WriteString(fmt.Sprintf("_ = %s\n", result))
    } else {
        t.output.WriteString(fmt.Sprintf("_ = %s\n", expandedBody))
    }
}
```

## Medium Priority Optimizations

### 6. Reduce Function Call Overhead
- Inline simple operations like `visitNode` for terminal nodes
- Use switch statements instead of multiple if conditions
- Cache frequently accessed fields

### 7. Memory Pool for Temporary Objects
```go
// Use sync.Pool for frequently allocated temporary objects
var builderPool = sync.Pool{
    New: func() interface{} {
        return &strings.Builder{}
    },
}

func getBuilder() *strings.Builder {
    return builderPool.Get().(*strings.Builder)
}

func putBuilder(b *strings.Builder) {
    b.Reset()
    builderPool.Put(b)
}
```

## Performance Benchmarks to Add

```go
func BenchmarkTranspileSimple(b *testing.B) {
    transpiler := New()
    input := `(+ 1 2 3)`
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        _, err := transpiler.TranspileFromInput(input)
        if err != nil {
            b.Fatal(err)
        }
    }
}

func BenchmarkMacroExpansion(b *testing.B) {
    // Test macro expansion performance
}
```

## Estimated Performance Gains

1. **Remove debug output**: 20-30% faster transpilation
2. **Optimize string building**: 15-25% reduction in allocations  
3. **Fix macro expansion**: 50-70% faster macro-heavy code
4. **Pre-allocate slices**: 10-15% fewer allocations
5. **Cache standard libs**: 5-10% faster import processing

Total estimated improvement: **2-3x faster transpilation** for typical Vex programs.