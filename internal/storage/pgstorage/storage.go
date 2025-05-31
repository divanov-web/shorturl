package pgstorage

import (
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
	"math/rand"
	"time"
)

type Storage struct {
	pool *pgxpool.Pool
	rnd  *rand.Rand
}

const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

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
		rnd:  rand.New(rand.NewSource(time.Now().UnixNano())),
	}

	if err := storage.ensureTable(ctx); err != nil {
		return nil, fmt.Errorf("failed to ensure table: %w", err)
	}

	return storage, nil
}

func (s *Storage) ensureTable(ctx context.Context) error {
	_, err := s.pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS short_urls (
			id TEXT PRIMARY KEY,
			original_url TEXT NOT NULL
		)
	`)
	return err
}

func (s *Storage) generateID(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = charset[s.rnd.Intn(len(charset))]
	}
	return string(b)
}

func (s *Storage) MakeShort(original string) (string, error) {
	ctx := context.Background()
	id := s.generateID(8)
	_, err := s.pool.Exec(ctx, `INSERT INTO short_urls (id, original_url) VALUES ($1, $2)`, id, original)
	if err == nil {
		return id, nil
	}
	return "", errors.New("failed to generate unique short ID")
}

func (s *Storage) GetURL(id string) (string, bool) {
	ctx := context.Background()
	var url string
	err := s.pool.QueryRow(ctx, `SELECT original_url FROM short_urls WHERE id = $1`, id).Scan(&url)
	if err != nil {
		return "", false
	}
	return url, true
}

func (s *Storage) ForceSet(id, url string) {
	ctx := context.Background()
	_, _ = s.pool.Exec(ctx, `INSERT INTO short_urls (id, original_url) VALUES ($1, $2) ON CONFLICT (id) DO NOTHING`, id, url)
}

func (s *Storage) Ping() error {
	return s.pool.Ping(context.Background())
}

func (s *Storage) Close() {
	s.pool.Close()
}
