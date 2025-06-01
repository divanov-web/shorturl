package storage

import "errors"

type BatchEntry struct {
	ShortURL      string
	OriginalURL   string
	CorrelationID string
}

type Storage interface {
	SaveURL(original string) (string, error)
	GetURL(id string) (string, bool)
	Ping() error
	BatchSave(entries []BatchEntry) error
}

var ErrConflict = errors.New("url already exists (storage)")
