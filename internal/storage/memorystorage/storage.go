package memorystorage

import (
	"math/rand"
	"sync"
	"time"
)

type Storage struct {
	data map[string]string
	mu   sync.RWMutex
	rnd  *rand.Rand
}

const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func NewStorage() (*Storage, error) {
	return &Storage{
		data: make(map[string]string),
		rnd:  rand.New(rand.NewSource(time.Now().UnixNano())),
	}, nil
}

func (s *Storage) generateID(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = charset[s.rnd.Intn(len(charset))]
	}
	return string(b)
}

func (s *Storage) MakeShort(original string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	id := s.generateID(8)
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
		rnd:  rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}
