package memorystorage_test

import (
	"strconv"
	"testing"

	"github.com/divanov-web/shorturl/internal/storage"
	"github.com/divanov-web/shorturl/internal/storage/memorystorage"
)

// Последовательный бенч вставки новых URL.
func BenchmarkMemoryStorage_SaveURL(b *testing.B) {
	s := memorystorage.NewTestStorage()
	const user = "bench-user"
	const base = "https://example.com/u/"

	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		// разные строки, чтобы каждый раз добавлять новый ключ в map
		_, _ = s.SaveURL(user, base+strconv.Itoa(i))
	}
}

// Последовательный бенч чтения существующих URL.
// Предзаполняем N ключей и циклически читаем их.
func BenchmarkMemoryStorage_GetURL(b *testing.B) {
	s := memorystorage.NewTestStorage()
	const user = "bench-user"

	// подготовим пул из 1000 записей
	const N = 1000
	ids := make([]string, N)
	for i := 0; i < N; i++ {
		id, _ := s.SaveURL(user, "https://example.com/"+strconv.Itoa(i))
		ids[i] = id
	}

	b.ReportAllocs()

	j := 0
	for i := 0; i < b.N; i++ {
		_, _ = s.GetURL(ids[j])
		j++
		if j == N {
			j = 0
		}
	}
}

// BatchSave: один и тот же батч записывается повторно.
func BenchmarkMemoryStorage_BatchSave(b *testing.B) {
	s := memorystorage.NewTestStorage()
	const user = "bench-user"

	const N = 1000
	entries := make([]storage.BatchEntry, N)
	for i := 0; i < N; i++ {
		entries[i] = storage.BatchEntry{
			CorrelationID: strconv.Itoa(i),
			ShortURL:      "id" + strconv.Itoa(i),
			OriginalURL:   "https://example.com/" + strconv.Itoa(i),
		}
	}

	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = s.BatchSave(user, entries)
	}
}
