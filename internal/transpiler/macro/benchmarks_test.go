package macro

import "testing"

func BenchmarkMacro_ExpandSimple(b *testing.B) {
	r := NewRegistry(Config{EnableValidation: false})
	_ = r.RegisterMacro("id", &Macro{Name: "id", Params: []string{"x"}, Body: "x"})
	e := NewExpander(r)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := e.ExpandMacro("id", []string{"42"}); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkMacro_ExpandChained(b *testing.B) {
	r := NewRegistry(Config{EnableValidation: false})
	_ = r.RegisterMacro("inc", &Macro{Name: "inc", Params: []string{"x"}, Body: "(+ x 1)"})
	e := NewExpander(r)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		v1, err := e.ExpandMacro("inc", []string{"10"})
		if err != nil {
			b.Fatal(err)
		}
		if _, err := e.ExpandMacro("inc", []string{v1}); err != nil {
			b.Fatal(err)
		}
	}
}
