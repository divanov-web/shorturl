package filestorage

import (
	"bufio"
	"encoding/json"
	"errors"
	"github.com/divanov-web/shorturl/internal/storage"
	"github.com/divanov-web/shorturl/internal/utils/idgen"
	"os"
	"sync"
	"time"
)

type Item struct {
	UUID        string `json:"uuid"`
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

type Storage struct {
	data     map[string]string
	mu       sync.RWMutex
	filePath string
}

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

func (s *Storage) SaveURL(userID string, original string) (string, error) {
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

func (s *Storage) GetURL(id string) (string, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	url, ok := s.data[id]
	return url, ok
}

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

func (s *Storage) Ping() error {
	return nil
}

func NewTestStorage() *Storage {
	return &Storage{
		data: make(map[string]string),
	}
}

func (s *Storage) BatchSave(userID string, entries []storage.BatchEntry) error {
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
func (s *Storage) GetUserURLs(userID string) ([]storage.UserURL, error) {
	var result []storage.UserURL
	return result, nil
}
