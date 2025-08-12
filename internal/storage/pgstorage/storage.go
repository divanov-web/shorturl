package pgstorage

import (
	"context"
	"errors"
	"fmt"

	"github.com/divanov-web/shorturl/internal/storage"
	"github.com/divanov-web/shorturl/internal/utils/idgen"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Storage описывает сам Storage хранения в БД.
type Storage struct {
	pool *pgxpool.Pool
}

// NewPool создаёт новое подключение к пулу PostgreSQL по переданному DSN.
func NewPool(ctx context.Context, dsn string) (*pgxpool.Pool, error) {
	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("invalid DSN: %w", err)
	}
	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to db: %w", err)
	}
	return pool, nil
}

// NewStorage создаёт хранилище в PostgreSQL и гарантирует наличие таблицы.
func NewStorage(ctx context.Context, pool *pgxpool.Pool) (*Storage, error) {
	storage := &Storage{
		pool: pool,
	}

	if err := storage.ensureTable(ctx); err != nil {
		return nil, fmt.Errorf("failed to ensure table: %w", err)
	}

	return storage, nil
}

func (s *Storage) ensureTable(ctx context.Context) error {
	_, err := s.pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS short_urls (
			id SERIAL PRIMARY KEY,
			short_url TEXT UNIQUE NOT NULL,
			original_url TEXT UNIQUE NOT NULL,
			user_guid TEXT NOT NULL,
			correlation_id TEXT,
			is_deleted BOOLEAN NOT NULL DEFAULT FALSE
		);
	`)
	return err
}

// SaveURL сохраняет оригинальный URL и возвращает его короткий идентификатор.
// При повторной вставке того же URL возвращает существующий short_url и ErrConflict.
// Используется ON CONFLICT только для инкремента с оптимизацией производительности.
func (s *Storage) SaveURL(ctx context.Context, userID string, original string) (string, error) {
	const maxRetries = 3
	for i := 0; i < maxRetries; i++ {
		candidate := idgen.Generate(8)

		// новый short_url, если вставка прошла
		// существующий short_url, если сработал конфликт по original_url
		var out string
		err := s.pool.QueryRow(ctx, `
			INSERT INTO short_urls (short_url, original_url, user_guid)
			VALUES ($1, $2, $3)
			ON CONFLICT (original_url) DO UPDATE
				SET original_url = EXCLUDED.original_url
			RETURNING short_url
		`, candidate, original, userID).Scan(&out)

		if err == nil {
			// если short совпал с существующим для другого original_url
			if out != candidate {
				return out, storage.ErrConflict
			}
			return out, nil
		}

		// Если это именно коллизия по short_url — перегенерируем и пробуем снова
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
			// Пробуем ещё раз с новым id
			continue
		}

		// Иная ошибка — отдаём наверх
		return "", fmt.Errorf("save failed: %w", err)
	}

	return "", fmt.Errorf("save failed: short id collision after %d retries", maxRetries)
}

// GetURL возвращает оригинальный URL по его короткому идентификатору.
func (s *Storage) GetURL(ctx context.Context, id string) (string, bool) {

	var url string
	err := s.pool.QueryRow(ctx, `
		SELECT original_url 
		FROM short_urls 
		WHERE short_url = $1 AND is_deleted = FALSE
	`, id).Scan(&url)
	if err != nil {
		return "", false
	}
	return url, true
}

// ForceSet добавляет или обновляет запись с указанным идентификатором и URL.
// Используется в тестах.
func (s *Storage) ForceSet(shortURL, url string) {
	ctx := context.Background()
	_, _ = s.pool.Exec(ctx, `INSERT INTO short_urls (short_url, original_url) VALUES ($1, $2) ON CONFLICT (id) DO NOTHING`, shortURL, url)
}

// Ping проверяет доступность хранилища (заглушка).
func (s *Storage) Ping() error {
	return s.pool.Ping(context.Background())
}

// Close закрывает пул соединений с базой данных.
func (s *Storage) Close() {
	s.pool.Close()
}

// BatchSave сохраняет парные значения id+url в рамках одной транзакции.
func (s *Storage) BatchSave(ctx context.Context, userID string, entries []storage.BatchEntry) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	for _, e := range entries {
		_, err := tx.Exec(ctx, `
			INSERT INTO short_urls (short_url, original_url, correlation_id, user_guid)
			VALUES ($1, $2, $3, $4)
			ON CONFLICT (short_url) DO NOTHING
		`, e.ShortURL, e.OriginalURL, e.CorrelationID, userID)
		if err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}

// GetUserURLs возвращает все ссылки пользователя, которые не помечены удалёнными.
func (s *Storage) GetUserURLs(ctx context.Context, userID string) ([]storage.UserURL, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT short_url, original_url
		FROM short_urls
		WHERE user_guid = $1 AND is_deleted = FALSE
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []storage.UserURL
	for rows.Next() {
		var item storage.UserURL
		if err := rows.Scan(&item.ShortURL, &item.OriginalURL); err != nil {
			return nil, err
		}
		result = append(result, item)
	}

	return result, nil
}

// MarkAsDeleted помечает указанные короткие ссылки пользователя как удалённые.
func (s *Storage) MarkAsDeleted(ctx context.Context, userID string, ids []string) error {
	if len(ids) == 0 {
		return nil
	}

	// Преобразуем []string в интерфейсный срез для передачи как $2
	rows := make([]interface{}, len(ids))
	for i, v := range ids {
		rows[i] = v
	}

	_, err := s.pool.Exec(ctx, `
		UPDATE short_urls
		SET is_deleted = TRUE
		WHERE user_guid = $1 AND short_url = ANY($2)
	`, userID, ids)

	return err
}
