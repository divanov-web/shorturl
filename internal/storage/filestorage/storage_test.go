package filestorage_test

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/divanov-web/shorturl/internal/storage"
	"github.com/divanov-web/shorturl/internal/storage/filestorage"
)

// TestNewStorage_LoadFromExistingFile Тестирование чтения файла
func TestNewStorage_LoadFromExistingFile(t *testing.T) {
	dir := t.TempDir()
	fp := filepath.Join(dir, "data.jsonl")

	// подготовим файл с двумя валидными строками
	f, err := os.Create(fp)
	if err != nil {
		t.Fatalf("create file: %v", err)
	}
	enc := json.NewEncoder(f)
	_ = enc.Encode(filestorage.Item{UUID: "1", ShortURL: "abc12345", OriginalURL: "https://a.com"})
	_ = enc.Encode(filestorage.Item{UUID: "2", ShortURL: "zzz00000", OriginalURL: "https://b.com"})
	_ = f.Close()

	s, err := filestorage.NewStorage(fp)
	if err != nil {
		t.Fatalf("NewStorage: %v", err)
	}

	if got, ok := s.GetURL("abc12345"); !ok || got != "https://a.com" {
		t.Fatalf("GetURL abc12345 = (%q,%v), want (https://a.com,true)", got, ok)
	}
	if got, ok := s.GetURL("zzz00000"); !ok || got != "https://b.com" {
		t.Fatalf("GetURL zzz00000 = (%q,%v), want (https://b.com,true)", got, ok)
	}
}

// TestSaveURL_WritesToMapAndFile тестирование записи в файл
func TestSaveURL_WritesToMapAndFile(t *testing.T) {
	dir := t.TempDir()
	fp := filepath.Join(dir, "data.jsonl")

	s, err := filestorage.NewStorage(fp)
	if err != nil {
		t.Fatalf("NewStorage: %v", err)
	}

	id, err := s.SaveURL("user1", "https://example.com/one")
	if err != nil {
		t.Fatalf("SaveURL: %v", err)
	}
	if len(id) != 8 {
		t.Fatalf("id length = %d, want 8", len(id))
	}

	if got, ok := s.GetURL(id); !ok || got != "https://example.com/one" {
		t.Fatalf("GetURL by id = (%q,%v), want (https://example.com/one,true)", got, ok)
	}

	// проверим, что дописалась строка в файл
	lines, err := countLines(fp)
	if err != nil {
		t.Fatalf("countLines: %v", err)
	}
	if lines != 1 {
		t.Fatalf("lines = %d, want 1", lines)
	}
}

// TestBatchSave_AddsOnlyNewAndAppendsToFile тест записи с дубликатами
func TestBatchSave_AddsOnlyNewAndAppendsToFile(t *testing.T) {
	dir := t.TempDir()
	fp := filepath.Join(dir, "data.jsonl")

	s, err := filestorage.NewStorage(fp)
	if err != nil {
		t.Fatalf("NewStorage: %v", err)
	}

	batch := []storage.BatchEntry{
		{ShortURL: "id1", OriginalURL: "https://a.com", CorrelationID: "1"},
		{ShortURL: "id2", OriginalURL: "https://b.com", CorrelationID: "2"},
	}
	if err := s.BatchSave("user1", batch); err != nil {
		t.Fatalf("BatchSave first: %v", err)
	}
	// повтор — записи уже существуют, в файл не должны добавиться дубликаты
	if err := s.BatchSave("user1", batch); err != nil {
		t.Fatalf("BatchSave second: %v", err)
	}

	// карта хранит обе записи
	if got, ok := s.GetURL("id1"); !ok || got != "https://a.com" {
		t.Fatalf("GetURL id1 = (%q,%v), want (https://a.com,true)", got, ok)
	}
	if got, ok := s.GetURL("id2"); !ok || got != "https://b.com" {
		t.Fatalf("GetURL id2 = (%q,%v), want (https://b.com,true)", got, ok)
	}

	// в файле должно быть ровно 2 строки (а не 4)
	lines, err := countLines(fp)
	if err != nil {
		t.Fatalf("countLines: %v", err)
	}
	if lines != 2 {
		t.Fatalf("lines = %d, want 2", lines)
	}
}

// --- helpers ---

func countLines(path string) (int, error) {
	f, err := os.Open(path)
	if err != nil {
		return 0, err
	}
	defer f.Close()

	sc := bufio.NewScanner(f)
	n := 0
	for sc.Scan() {
		n++
	}
	return n, sc.Err()
}

func errorsIs(err, target error) bool {
	if err == nil {
		return target == nil
	}
	type causer interface{ Unwrap() error }
	if err == target {
		return true
	}
	if e, ok := err.(causer); ok {
		return errorsIs(e.Unwrap(), target)
	}
	return false
}
