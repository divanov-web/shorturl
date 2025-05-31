package filestorage

import (
	"bufio"
	"encoding/json"
	"errors"
	"math/rand"
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
	rnd      *rand.Rand
}

const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func NewStorage(filePath string) (*Storage, error) {
	s := &Storage{
		data:     make(map[string]string),
		filePath: filePath,
		rnd:      rand.New(rand.NewSource(time.Now().UnixNano())),
	}

	// Загружаем данные из файла (если файл существует)
	if err := s.loadFromFile(); err != nil && !errors.Is(err, os.ErrNotExist) {
		return nil, err
	}

	return s, nil
}

func (s *Storage) generateID(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = charset[s.rnd.Intn(len(charset))]
	}
	return string(b)
}

func (s *Storage) MakeShort(original string) string {
	s.mu.Lock()
	defer s.mu.Unlock()

	id := s.generateID(8)
	s.data[id] = original

	_ = s.appendToFile(Item{
		UUID:        time.Now().Format("20060102150405.000000"), // временный uuid
		ShortURL:    id,
		OriginalURL: original,
	})

	return id
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

func NewTestStorage() *Storage {
	return &Storage{
		data: make(map[string]string),
		rnd:  rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}
