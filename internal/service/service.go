package service

import (
	"errors"
	"fmt"
	"strings"

	"github.com/divanov-web/shorturl/internal/storage"
	"github.com/divanov-web/shorturl/internal/utils/idgen"
)

type BatchRequestItem struct {
	CorrelationID string `json:"correlation_id"`
	OriginalURL   string `json:"original_url"`
}

type ShortenBatchResult struct {
	CorrelationID string `json:"correlation_id"`
	ShortURL      string `json:"short_url"`
}

type URLService struct {
	BaseURL string
	Repo    storage.Storage
}

var ErrAlreadyExists = errors.New("url already exists (service)")

func NewURLService(baseURL string, repo storage.Storage) *URLService {
	return &URLService{
		BaseURL: baseURL,
		Repo:    repo,
	}
}

func (s *URLService) CreateShort(userID string, original string) (string, error) {
	original = strings.TrimSpace(original)
	if original == "" {
		return "", fmt.Errorf("empty original URL")
	}

	id, err := s.Repo.SaveURL(userID, original)
	if errors.Is(err, storage.ErrConflict) {
		return fmt.Sprintf("%s/%s", s.BaseURL, id), ErrAlreadyExists
	}

	return fmt.Sprintf("%s/%s", s.BaseURL, id), err
}

func (s *URLService) CreateShortBatch(userID string, input []BatchRequestItem) ([]ShortenBatchResult, error) {
	entries := make([]storage.BatchEntry, 0, len(input))
	results := make([]ShortenBatchResult, 0, len(input))

	for _, item := range input {
		short := idgen.Generate(8)
		entries = append(entries, storage.BatchEntry{
			ShortURL:      short,
			OriginalURL:   item.OriginalURL,
			CorrelationID: item.CorrelationID,
		})
		results = append(results, ShortenBatchResult{
			CorrelationID: item.CorrelationID,
			ShortURL:      fmt.Sprintf("%s/%s", s.BaseURL, short),
		})
	}

	if err := s.Repo.BatchSave(userID, entries); err != nil {
		return nil, err
	}

	return results, nil
}

func (s *URLService) ResolveShort(id string) (string, bool) {
	return s.Repo.GetURL(id)
}

func (s *URLService) Ping() error {
	if pinger, ok := s.Repo.(interface{ Ping() error }); ok {
		return pinger.Ping()
	}
	return nil
}
