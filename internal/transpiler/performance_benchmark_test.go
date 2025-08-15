package transpiler

import (
	"testing"
)

// Benchmark simple transpilation performance
func BenchmarkTranspileSimple(b *testing.B) {
	input := `(+ 1 2 3)`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		transpiler, _ := NewBuilder().Build()
		_, err := transpiler.TranspileFromInput(input)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// Benchmark arithmetic expression building
func BenchmarkArithmeticExpression(b *testing.B) {
	input := `(+ 1 2 3 4 5 6 7 8 9 10)`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		transpiler, _ := NewBuilder().Build()
		_, err := transpiler.TranspileFromInput(input)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// Benchmark macro expansion performance
func BenchmarkMacroExpansion(b *testing.B) {
	input := `
(macro greet [name] (fmt/Println "Hello" name))
(greet "World")
`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		transpiler, _ := NewBuilder().Build()
		_, err := transpiler.TranspileFromInput(input)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// Benchmark debug vs non-debug mode
func BenchmarkDebugMode(b *testing.B) {
	input := `
(macro greet [name] (fmt/Println "Hello" name))
(greet "World")
(+ 1 2 3)
`

	b.Run("NoDebug", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			transpiler, _ := NewBuilder().Build()
			_, err := transpiler.TranspileFromInput(input)
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("WithDebug", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			transpiler, _ := NewBuilder().Build()
			_, err := transpiler.TranspileFromInput(input)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}

// Benchmark complex nested expressions
func BenchmarkComplexExpression(b *testing.B) {
	input := `
(def x (+ 1 2))
(def y (* 3 4))
(def z (/ 10 2))
(+ (* x y) (- z 1))
`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		transpiler, _ := NewBuilder().Build()
		_, err := transpiler.TranspileFromInput(input)
		if err != nil {
			b.Fatal(err)
		}
	}
}
