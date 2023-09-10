package mjingo

import "testing"

func BenchmarkValueTryToGoString(b *testing.B) {
	blackbox := func(string, error) {}

	b.Run("Str", func(b *testing.B) {
		val := valueFromString("a string")
		for i := 0; i < b.N; i++ {
			blackbox(valueTryToGoString(val))
		}
	})
	b.Run("NotStr", func(b *testing.B) {
		val := valueFromI64(123)
		for i := 0; i < b.N; i++ {
			blackbox(valueTryToGoString(val))
		}
	})
}

func BenchmarkValueAsGoString(b *testing.B) {
	blackbox := func(string, bool) {}

	b.Run("Str", func(b *testing.B) {
		val := valueFromString("a string")
		for i := 0; i < b.N; i++ {
			blackbox(valueAsGoString(val))
		}
	})
	b.Run("NotStr", func(b *testing.B) {
		val := valueFromI64(123)
		for i := 0; i < b.N; i++ {
			blackbox(valueAsGoString(val))
		}
	})
}
