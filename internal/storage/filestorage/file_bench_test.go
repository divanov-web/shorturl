package filestorage_test

import (
	"context"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/divanov-web/shorturl/internal/storage"
	"github.com/divanov-web/shorturl/internal/storage/filestorage"
)

func BenchmarkFileStorage_SaveURL(b *testing.B) {
	dir := b.TempDir()
	fp := filepath.Join(dir, "data.jsonl")

	s, err := filestorage.NewStorage(fp)
	if err != nil {
		b.Fatalf("NewStorage: %v", err)
	}

	const user = "bench-user"
	const base = "https://example.com/u/"

	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _ = s.SaveURL(context.Background(), user, base+strconv.Itoa(i))
	}
}

func BenchmarkFileStorage_GetURL(b *testing.B) {
	dir := b.TempDir()
	fp := filepath.Join(dir, "data.jsonl")

	s, err := filestorage.NewStorage(fp)
	if err != nil {
		b.Fatalf("NewStorage: %v", err)
	}

	const user = "bench-user"

	// Предзаполняем некоторый пул id, чтобы измерять чистое чтение из map под RLock.
	const N = 2000
	ids := make([]string, N)
	for i := 0; i < N; i++ {
		id, _ := s.SaveURL(context.Background(), user, "https://example.com/"+strconv.Itoa(i))
		ids[i] = id
	}

	b.ReportAllocs()
	b.ResetTimer()

	j := 0
	for i := 0; i < b.N; i++ {
		_, _ = s.GetURL(context.Background(), ids[j])
		j++
		if j == N {
			j = 0
		}
	}
}

func BenchmarkFileStorage_BatchSave_NewEntries(b *testing.B) {
	dir := b.TempDir()
	fp := filepath.Join(dir, "data.jsonl")

	s, err := filestorage.NewStorage(fp)
	if err != nil {
		b.Fatalf("NewStorage: %v", err)
	}

	const user = "bench-user"

	const batchSize = 100
	entries := make([]storage.BatchEntry, batchSize)

	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		// Обновляем содержимое батча, чтобы каждый вызов был с новыми ключами
		prefix := "iter" + strconv.Itoa(i) + "_"
		for j := 0; j < batchSize; j++ {
			id := prefix + strconv.Itoa(j)
			entries[j] = storage.BatchEntry{
				CorrelationID: id,
				ShortURL:      id,
				OriginalURL:   "https://example.com/" + id,
			}
		}
		_ = s.BatchSave(context.Background(), user, entries)
	}
}
