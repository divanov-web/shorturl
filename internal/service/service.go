package service

import (
	"fmt"
	"strings"

	"github.com/divanov-web/shorturl/internal/storage"
)

type URLService struct {
	BaseURL string
	Repo    storage.Storage
}

func NewURLService(baseURL string, repo storage.Storage) *URLService {
	return &URLService{
		BaseURL: baseURL,
		Repo:    repo,
	}
}

func (s *URLService) CreateShort(original string) (string, error) {
	original = strings.TrimSpace(original)
	if original == "" {
		return "", fmt.Errorf("empty original URL")
	}
	id, err := s.Repo.MakeShort(original)
	return fmt.Sprintf("%s/%s", s.BaseURL, id), err
}

func (s *URLService) ResolveShort(id string) (string, bool) {
	return s.Repo.GetURL(id)
}
