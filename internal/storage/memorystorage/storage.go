package memorystorage

import (
	"github.com/divanov-web/shorturl/internal/storage"
	"github.com/divanov-web/shorturl/internal/utils/idgen"
	"sync"
)

type Storage struct {
	data map[string]string
	mu   sync.RWMutex
}

func NewStorage() (*Storage, error) {
	return &Storage{
		data: make(map[string]string),
	}, nil
}

func (s *Storage) SaveURL(original string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	id := idgen.Generate(8)
	s.data[id] = original

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

func (s *Storage) Ping() error {
	return nil
}

func NewTestStorage() *Storage {
	return &Storage{
		data: make(map[string]string),
	}
}

func (s *Storage) BatchSave(entries []storage.BatchEntry) error {
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
