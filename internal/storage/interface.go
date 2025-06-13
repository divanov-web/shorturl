package storage

import "errors"

type BatchEntry struct {
	ShortURL      string
	OriginalURL   string
	CorrelationID string
}

type UserURL struct {
	ShortURL    string
	OriginalURL string
}

type Storage interface {
	SaveURL(userID string, original string) (string, error)
	GetURL(id string) (string, bool)
	Ping() error
	BatchSave(userID string, entries []BatchEntry) error
	GetUserURLs(userID string) ([]UserURL, error)
}

var ErrConflict = errors.New("url already exists (storage)")
