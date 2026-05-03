package storage

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrPostgresConfig      = errors.New("postgres: invalid config")
	ErrPostgresUnavailable = errors.New("postgres: unavailable")
)

// Constants for postgres connection pool
const (
	postgresMaxConns        = 25
	postgresMinConns        = 5
	postgresMaxConnLifetime = time.Hour
	postgresMaxConnIdleTime = 30 * time.Minute
)

type Postgres struct {
	Pool *pgxpool.Pool
}

func NewPostgres(ctx context.Context, url string) (*Postgres, error) {
	if url == "" {
		return nil, fmt.Errorf("%w: URL is required", ErrPostgresConfig)
	}

	cfg, err := pgxpool.ParseConfig(url)
	if err != nil {
		return nil, fmt.Errorf("%w: parse: %v", ErrPostgresConfig, err)
	}

	cfg.MaxConns = postgresMaxConns
	cfg.MinConns = postgresMinConns
	cfg.MaxConnLifetime = postgresMaxConnLifetime
	cfg.MaxConnIdleTime = postgresMaxConnIdleTime

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("%w: create pool: %v", ErrPostgresUnavailable, err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("%w: ping: %v", ErrPostgresUnavailable, err)
	}

	slog.Info("postgres connected")
	return &Postgres{Pool: pool}, nil
}

func (p *Postgres) Close() {
	if p.Pool == nil {
		return
	}
	p.Pool.Close()
	slog.Info("postgres closed")
}