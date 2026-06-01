package postgres

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Constants for postgres connection pool
const (
	maxConns        = 25
	minConns        = 5
	maxConnLifetime = time.Hour
	maxConnIdleTime = 30 * time.Minute
)

type DB struct {
	Pool *pgxpool.Pool
}

func New(ctx context.Context, url string) (*DB, error) {
	if url == "" {
		return nil, fmt.Errorf("%w: URL is required", ErrConfig)
	}

	cfg, err := pgxpool.ParseConfig(url)
	if err != nil {
		return nil, fmt.Errorf("%w: parse: %v", ErrConfig, err)
	}

	cfg.MaxConns = maxConns
	cfg.MinConns = minConns
	cfg.MaxConnLifetime = maxConnLifetime
	cfg.MaxConnIdleTime = maxConnIdleTime

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("%w: create pool: %v", ErrUnavailable, err)
	}

	db := &DB{Pool: pool}

	if err := db.Ping(ctx); err != nil {
		db.Pool.Close()
		return nil, err
	}

	slog.Info("postgres connected")
	return db, nil
}

// Ping checks Postgres reachability. Returns ErrUnavailable wrapped with the
// underlying error on failure
func (db *DB) Ping(ctx context.Context) error {
	if err := db.Pool.Ping(ctx); err != nil {
		return fmt.Errorf("%w: %v", ErrUnavailable, err)
	}
	return nil
}

func (db *DB) Close() {
	if db.Pool == nil {
		return
	}
	db.Pool.Close()
	slog.Info("postgres closed")
}