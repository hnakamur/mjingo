package mjingo

import "testing"

func BenchmarkValueTryToGoString(b *testing.B) {
	b.Run("methodStr", func(b *testing.B) {
		val := valueFromString("a string")
		for i := 0; i < b.N; i++ {
			valueTryToGoStringWithAsStr(val)
		}
	})
	b.Run("methodNotStr", func(b *testing.B) {
		val := valueFromI64(123)
		for i := 0; i < b.N; i++ {
			valueTryToGoStringWithAsStr(val)
		}
	})
	b.Run("typeSwitchStr", func(b *testing.B) {
		val := valueFromString("a string")
		for i := 0; i < b.N; i++ {
			valueTryToGoString(val)
		}
	})
	b.Run("typeSwitchNotStr", func(b *testing.B) {
		val := valueFromI64(123)
		for i := 0; i < b.N; i++ {
			valueTryToGoString(val)
		}
	})
}
