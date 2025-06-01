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

type Storage struct {
	pool *pgxpool.Pool
}

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
			correlation_id TEXT
		);
	`)
	return err
}

func (s *Storage) SaveURL(original string) (string, error) {
	ctx := context.Background()
	shortURL := idgen.Generate(8)

	_, err := s.pool.Exec(ctx,
		`INSERT INTO short_urls (short_url, original_url) VALUES ($1, $2)`,
		shortURL, original,
	)
	if err == nil {
		return shortURL, nil
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
		// Запись уже есть — достаём существующий short_url
		var existing string
		queryErr := s.pool.QueryRow(ctx,
			`SELECT short_url FROM short_urls WHERE original_url = $1`,
			original,
		).Scan(&existing)
		if queryErr == nil {
			return existing, storage.ErrConflict
		}
	}

	return "", fmt.Errorf("save failed: %w", err)
}

func (s *Storage) GetURL(id string) (string, bool) {
	ctx := context.Background()
	var url string
	err := s.pool.QueryRow(ctx, `SELECT original_url FROM short_urls WHERE short_url = $1`, id).Scan(&url)
	if err != nil {
		return "", false
	}
	return url, true
}

func (s *Storage) ForceSet(shortURL, url string) {
	ctx := context.Background()
	_, _ = s.pool.Exec(ctx, `INSERT INTO short_urls (short_url, original_url) VALUES ($1, $2) ON CONFLICT (id) DO NOTHING`, shortURL, url)
}

func (s *Storage) Ping() error {
	return s.pool.Ping(context.Background())
}

func (s *Storage) Close() {
	s.pool.Close()
}

// BatchSave сохраняет парные значения id+url в рамках одной транзакции
func (s *Storage) BatchSave(entries []storage.BatchEntry) error {
	ctx := context.Background()
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	for _, e := range entries {
		_, err := tx.Exec(ctx, `
			INSERT INTO short_urls (short_url, original_url, correlation_id)
			VALUES ($1, $2, $3)
			ON CONFLICT (short_url) DO NOTHING
		`, e.ShortURL, e.OriginalURL, e.CorrelationID)
		if err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}
