package db

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Storage struct {
	Pool *pgxpool.Pool
}

func NewPostgres(ctx context.Context, dsn string) (*Storage, error) {
	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("invalid DSN: %w", err)
	}

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to db: %w", err)
	}

	// Проверка соединения
	/*if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("ping error: %w", err)
	}*/

	return &Storage{Pool: pool}, nil
}

func (s *Storage) Ping() error {
	return s.Pool.Ping(context.Background())
}

func (s *Storage) Close() {
	s.Pool.Close()
}
