// Package service слой сервиса для коротких ссылок
package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/divanov-web/shorturl/internal/storage"
	"github.com/divanov-web/shorturl/internal/utils/idgen"
)

// BatchRequestItem описывает входные данные для пакетного создания коротких ссылок.
type BatchRequestItem struct {
	CorrelationID string `json:"correlation_id"`
	OriginalURL   string `json:"original_url"`
}

// ShortenBatchResult описывает результат пакетного сокращения ссылок.
type ShortenBatchResult struct {
	CorrelationID string `json:"correlation_id"`
	ShortURL      string `json:"short_url"`
}

type deleteTask struct {
	UserID string
	IDs    []string
}

// URLService описывает бизнес-логику сервиса коротких ссылок.
type URLService struct {
	BaseURL    string
	Repo       storage.Storage
	deleteChan chan deleteTask
}

// ErrAlreadyExists Ошибка url уже существует (от уровня сервиса)
var ErrAlreadyExists = errors.New("url already exists (service)")

// NewURLService создаёт новый сервис для работы с короткими ссылками и запускает воркер удаления.
func NewURLService(ctx context.Context, baseURL string, repo storage.Storage) *URLService {
	svc := &URLService{
		BaseURL:    baseURL,
		Repo:       repo,
		deleteChan: make(chan deleteTask, 5),
	}

	go svc.startDeleteWorker(ctx)
	return svc
}

// CreateShort создаёт короткую ссылку для переданного оригинального URL.
func (s *URLService) CreateShort(ctx context.Context, userID string, original string) (string, error) {
	original = strings.TrimSpace(original)
	if original == "" {
		return "", fmt.Errorf("empty original URL")
	}

	id, err := s.Repo.SaveURL(ctx, userID, original)
	if errors.Is(err, storage.ErrConflict) {
		return fmt.Sprintf("%s/%s", s.BaseURL, id), ErrAlreadyExists
	}

	return fmt.Sprintf("%s/%s", s.BaseURL, id), err
}

// CreateShortBatch создаёт несколько коротких ссылок за один запрос.
func (s *URLService) CreateShortBatch(ctx context.Context, userID string, input []BatchRequestItem) ([]ShortenBatchResult, error) {
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

	if err := s.Repo.BatchSave(ctx, userID, entries); err != nil {
		return nil, err
	}

	return results, nil
}

// ResolveShort возвращает оригинальный URL по идентификатору короткой ссылки.
func (s *URLService) ResolveShort(ctx context.Context, id string) (string, bool) {
	return s.Repo.GetURL(ctx, id)
}

// Ping проверяет доступность хранилища, если оно поддерживает метод Ping.
func (s *URLService) Ping() error {
	if pinger, ok := s.Repo.(interface{ Ping() error }); ok {
		return pinger.Ping()
	}
	return nil
}

// GetUserURLs возвращает список коротких ссылок пользователя.
func (s *URLService) GetUserURLs(ctx context.Context, userID string) ([]storage.UserURL, error) {
	return s.Repo.GetUserURLs(ctx, userID)
}

// DeleteUserURLs помечает ссылки пользователя как удалённые.
func (s *URLService) DeleteUserURLs(ctx context.Context, userID string, ids []string) error {
	return s.Repo.MarkAsDeleted(ctx, userID, ids)
}

// startDeleteWorker запускает фоновую обработку задач на удаление ссылок.
func (s *URLService) startDeleteWorker(ctx context.Context) {
	const maxBatchSize = 100

	buffer := make(map[string][]string)
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case task, ok := <-s.deleteChan:
			if !ok {
				// Канал закрыт — flush и выход
				s.flushBuffer(ctx, buffer)
				return
			}
			buffer[task.UserID] = append(buffer[task.UserID], task.IDs...)
			if len(buffer[task.UserID]) >= maxBatchSize {
				_ = s.Repo.MarkAsDeleted(ctx, task.UserID, buffer[task.UserID])
				buffer[task.UserID] = buffer[task.UserID][:0]
			}
		case <-ticker.C:
			// периодический сброс буфера
			s.flushBuffer(ctx, buffer)
		case <-ctx.Done():
			// Завершение через context — тоже flush
			s.flushBuffer(ctx, buffer)
			return
		}
	}
}

// flushBuffer - сейчас перед удалением накапливается буфер из задач на удаление.
// Если буфер не накопился, его нужно сбрасывать вручную
func (s *URLService) flushBuffer(ctx context.Context, buffer map[string][]string) {
	for userID, ids := range buffer {
		if len(ids) > 0 {
			_ = s.Repo.MarkAsDeleted(ctx, userID, ids)
		}
	}
}

// DeleteShortURLsAsync добавляет задачу на асинхронное удаление ссылок.
func (s *URLService) DeleteShortURLsAsync(userID string, ids []string) {
	s.deleteChan <- deleteTask{
		UserID: userID,
		IDs:    ids,
	}
}

// Stats возвращает агрегированную статистику по сервису.
type Stats struct {
	URLs  int `json:"urls"`
	Users int `json:"users"`
}

func (s *URLService) Stats(ctx context.Context) (Stats, error) {
	urls, err := s.Repo.CountURLs(ctx)
	if err != nil {
		return Stats{}, err
	}
	users, err := s.Repo.CountUsers(ctx)
	if err != nil {
		return Stats{}, err
	}
	return Stats{URLs: urls, Users: users}, nil
}
