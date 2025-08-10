package idgen_test

import (
	"github.com/divanov-web/shorturl/internal/utils/idgen"
	"testing"
)

// Базовый бенч на фиксированную длину.
func BenchmarkGenerate8(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = idgen.Generate(8)
	}
}

func BenchmarkGenerate16(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = idgen.Generate(16)
	}
}

// Имитация конкуренции: ближе к реальному трафику. Пока что не проходит проверку
/*func BenchmarkGenerateParallel(b *testing.B) {
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = idgen.Generate(10)
		}
	})
}*/
