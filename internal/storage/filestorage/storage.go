package filestorage

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"os"
	"sync"
	"time"

	"github.com/divanov-web/shorturl/internal/storage"
	"github.com/divanov-web/shorturl/internal/utils/idgen"
)

// Item описывает ссылку для сохранения в файле.
type Item struct {
	UUID        string `json:"uuid"`
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

// Storage описывает сам Storage файлового хранилища.
type Storage struct {
	data     map[string]string
	mu       sync.RWMutex
	filePath string
}

// NewStorage создаёт файловое хранилище и загружает данные из указанного файла.
func NewStorage(filePath string) (*Storage, error) {
	s := &Storage{
		data:     make(map[string]string),
		filePath: filePath,
	}

	// Загружаем данные из файла (если файл существует)
	if err := s.loadFromFile(); err != nil && !errors.Is(err, os.ErrNotExist) {
		return nil, err
	}

	return s, nil
}

// SaveURL сохраняет оригинальный URL и возвращает его короткий идентификатор.
func (s *Storage) SaveURL(ctx context.Context, userID string, original string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	id := idgen.Generate(8)
	s.data[id] = original

	_ = s.appendToFile(Item{
		UUID:        time.Now().Format("20060102150405.000000"), // временный uuid
		ShortURL:    id,
		OriginalURL: original,
	})

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

func (s *Storage) appendToFile(item Item) error {
	if s.filePath == "" {
		return nil
	}
	file, err := os.OpenFile(s.filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	enc := json.NewEncoder(file)
	return enc.Encode(item)
}

func (s *Storage) loadFromFile() error {
	if s.filePath == "" {
		return nil
	}
	file, err := os.Open(s.filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		var item Item
		if err := json.Unmarshal(scanner.Bytes(), &item); err != nil {
			continue
		}
		s.data[item.ShortURL] = item.OriginalURL
	}
	return scanner.Err()
}

// Ping проверяет доступность хранилища (заглушка).
func (s *Storage) Ping() error {
	return nil
}

// NewTestStorage создаёт тестовую версию файлового хранилища в памяти.
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
		if _, exists := s.data[entry.ShortURL]; exists {
			continue
		}
		s.data[entry.ShortURL] = entry.OriginalURL

		_ = s.appendToFile(Item{
			UUID:        time.Now().Format("20060102150405.000000"),
			ShortURL:    entry.ShortURL,
			OriginalURL: entry.OriginalURL,
		})
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

// Shutdown корректно завершает файловое хранилище
// Вроде как в текущей реализации все файлы закрываются сами
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
