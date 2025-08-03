// Package tests provides performance benchmarks for the Vex transpiler
package tests

import (
	"strings"
	"testing"

	"github.com/thsfranca/vex/internal/transpiler"
)

// Sample Vex code for benchmarking
const (
	simpleExpression = `(+ 1 2)`

	variableDefinition = `(def x 42)`

	complexArithmetic = `(+ (* 2 3) (- 10 5) (* 4 2))`

	httpServerExample = `
(import "net/http")
(import "github.com/gorilla/mux")
(def router (mux/NewRouter))
(.HandleFunc router "/hello" (fn [w r] (.WriteString w "Hello World!")))
(.ListenAndServe http ":8080" router)
`

	macroExample = `
(macro simple-handler [name path response]
  (fn [w r] (.WriteString w ~response)))

(def hello-handler (simple-handler hello "/hello" "Hello World!"))
(def goodbye-handler (simple-handler goodbye "/bye" "Goodbye!"))
`

	nestedExpressions = `
(def result (+ (* (+ 1 2) (- 5 2)) (/ (* 8 4) (+ 2 2))))
(def complex (.Method (package/Function arg1 arg2) (+ x y)))
`
)

// BenchmarkSimpleExpression benchmarks simple arithmetic expressions
func BenchmarkSimpleExpression(b *testing.B) {
	tr := transpiler.New()
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		_, err := tr.TranspileFromInput(simpleExpression)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkVariableDefinition benchmarks variable definitions
func BenchmarkVariableDefinition(b *testing.B) {
	tr := transpiler.New()
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		_, err := tr.TranspileFromInput(variableDefinition)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkComplexArithmetic benchmarks complex arithmetic expressions
func BenchmarkComplexArithmetic(b *testing.B) {
	tr := transpiler.New()
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		_, err := tr.TranspileFromInput(complexArithmetic)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkHttpServerExample benchmarks HTTP server transpilation
func BenchmarkHttpServerExample(b *testing.B) {
	tr := transpiler.New()
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		_, err := tr.TranspileFromInput(httpServerExample)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkMacroExample benchmarks macro definition and usage
func BenchmarkMacroExample(b *testing.B) {
	tr := transpiler.New()
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		_, err := tr.TranspileFromInput(macroExample)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkNestedExpressions benchmarks deeply nested expressions
func BenchmarkNestedExpressions(b *testing.B) {
	tr := transpiler.New()
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		_, err := tr.TranspileFromInput(nestedExpressions)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkLargeFile benchmarks transpilation of a large file
func BenchmarkLargeFile(b *testing.B) {
	// Generate large file content
	var content strings.Builder
	for i := 0; i < 100; i++ {
		content.WriteString("(def var")
		content.WriteString(string(rune('0' + i%10)))
		content.WriteString(" ")
		content.WriteString(string(rune('0' + i%10)))
		content.WriteString(")\n")
		content.WriteString("(def result")
		content.WriteString(string(rune('0' + i%10)))
		content.WriteString(" (+ var")
		content.WriteString(string(rune('0' + i%10)))
		content.WriteString(" (* 2 ")
		content.WriteString(string(rune('0' + i%10)))
		content.WriteString(")))\n")
	}
	
	tr := transpiler.New()
	input := content.String()
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		_, err := tr.TranspileFromInput(input)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkTranspilerReuse benchmarks transpiler reuse vs recreation
func BenchmarkTranspilerReuse(b *testing.B) {
	tr := transpiler.New()
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		tr.Reset() // Reuse existing transpiler
		_, err := tr.TranspileFromInput(httpServerExample)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkTranspilerRecreation benchmarks creating new transpiler each time
func BenchmarkTranspilerRecreation(b *testing.B) {
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		tr := transpiler.New() // Create new transpiler each time
		_, err := tr.TranspileFromInput(httpServerExample)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkMemoryAllocations benchmarks memory allocations
func BenchmarkMemoryAllocations(b *testing.B) {
	tr := transpiler.New()
	b.ReportAllocs() // Report memory allocations
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		_, err := tr.TranspileFromInput(httpServerExample)
		if err != nil {
			b.Fatal(err)
		}
		tr.Reset()
	}
}

// BenchmarkStringOperations benchmarks string-heavy operations
func BenchmarkStringOperations(b *testing.B) {
	// Test with many string operations
	input := `
(def greeting "Hello")
(def name "World")
(def message (+ greeting " " name "!"))
(def path "/api/v1/users")
(def method "GET")
(def contentType "application/json")
`
	
	tr := transpiler.New()
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		_, err := tr.TranspileFromInput(input)
		if err != nil {
			b.Fatal(err)
		}
		tr.Reset()
	}
}

// BenchmarkSlashNotation benchmarks slash notation calls
func BenchmarkSlashNotation(b *testing.B) {
	input := `
(fmt/Println "Hello")
(json/Marshal data)
(http/Get "https://api.example.com")
(mux/NewRouter)
(strings/ToUpper name)
`
	
	tr := transpiler.New()
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		_, err := tr.TranspileFromInput(input)
		if err != nil {
			b.Fatal(err)
		}
		tr.Reset()
	}
}

// BenchmarkMethodCalls benchmarks method call transpilation
func BenchmarkMethodCalls(b *testing.B) {
	input := `
(.HandleFunc router "/api" handler)
(.WriteString writer "content")
(.SetHeader response "Content-Type" "application/json")
(.AddMiddleware app middleware)
(.Listen server ":8080")
`
	
	tr := transpiler.New()
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		_, err := tr.TranspileFromInput(input)
		if err != nil {
			b.Fatal(err)
		}
		tr.Reset()
	}
}