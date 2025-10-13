package memorystorage

import (
	"context"
	"sync"

	"github.com/divanov-web/shorturl/internal/storage"
	"github.com/divanov-web/shorturl/internal/utils/idgen"
)

// Storage описывает хранение в оперативной памяти.
type Storage struct {
	data map[string]string
	mu   sync.RWMutex
}

// NewStorage создаёт новое хранилище в оперативной памяти.
func NewStorage() (*Storage, error) {
	return &Storage{
		data: make(map[string]string),
	}, nil
}

// SaveURL сохраняет оригинальный URL и возвращает его короткий идентификатор.
func (s *Storage) SaveURL(ctx context.Context, userID string, original string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	id := idgen.Generate(8)
	s.data[id] = original

	return id, nil
}

// GetURL возвращает оригинальный URL по его короткому идентификатору.
func (s *Storage) GetURL(ctx context.Context, id string) (string, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	url, ok := s.data[id]
	return url, ok
}

// ForceSet добавляет или обновляет запись с указанным идентификатором и URL.
// Используется в тестах.
func (s *Storage) ForceSet(id, url string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data[id] = url
}

// Ping проверяет доступность хранилища (заглушка).
func (s *Storage) Ping() error {
	return nil
}

// NewTestStorage создаёт тестовое хранилище в памяти.
// Используется в тестах.
func NewTestStorage() *Storage {
	return &Storage{
		data: make(map[string]string),
	}
}

// BatchSave сохраняет несколько записей за один вызов.
func (s *Storage) BatchSave(ctx context.Context, userID string, entries []storage.BatchEntry) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, entry := range entries {
		// не перезаписываем, если уже существует
		if _, exists := s.data[entry.ShortURL]; !exists {
			s.data[entry.ShortURL] = entry.OriginalURL
		}
	}

	return nil
}

// GetUserURLs - заглушка, отправляем пустой список
func (s *Storage) GetUserURLs(ctx context.Context, userID string) ([]storage.UserURL, error) {
	var result []storage.UserURL
	return result, nil
}

// MarkAsDeleted помечает ссылки пользователя как удалённые (заглушка).
func (s *Storage) MarkAsDeleted(ctx context.Context, userID string, ids []string) error {
	return storage.ErrNotImplemented
}

// Shutdown корректно завершает memorystorage, заглушка
func (s *Storage) Shutdown(ctx context.Context) error {
	return nil
}

// CountURLs возвращает количество сокращённых URL (по размеру карты).
func (s *Storage) CountURLs(ctx context.Context) (int, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.data), nil
}

// CountUsers возвращает количество пользователей — заглушка (нет трекинга пользователей).
func (s *Storage) CountUsers(ctx context.Context) (int, error) {
	return 0, nil
}
