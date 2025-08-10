// Package storage слой storage(repository) для коротких ссылок
package storage

import "errors"

// BatchEntry структура полученного url для групповых батч записей
type BatchEntry struct {
	ShortURL      string
	OriginalURL   string
	CorrelationID string
}

// UserURL структура полученного url от пользователя для одиночных записей
type UserURL struct {
	ShortURL    string
	OriginalURL string
	DeletedFlag bool
}

type Storage interface {
	SaveURL(userID string, original string) (string, error)
	GetURL(id string) (string, bool)
	Ping() error
	BatchSave(userID string, entries []BatchEntry) error
	GetUserURLs(userID string) ([]UserURL, error)
	MarkAsDeleted(userID string, ids []string) error
}

var ErrConflict = errors.New("url already exists (storage)")
var ErrNotImplemented = errors.New("MarkAsDeleted not implemented in memory storage")
