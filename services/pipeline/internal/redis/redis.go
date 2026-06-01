package redis

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	goredis "github.com/redis/go-redis/v9"
)

var (
	ErrConfig      = errors.New("redis: invalid config")
	ErrUnavailable = errors.New("redis: unavailable")
)

// Constants for redis connection pool. Aggressive timeouts because Redis
// sits inside the synchronous decision path
const (
	poolSize     = 50
	minIdleConns = 10
	dialTimeout  = 5 * time.Second
	readTimeout  = 3 * time.Second
	writeTimeout = 3 * time.Second
	poolTimeout  = 4 * time.Second
)

type Store struct {
	Client *goredis.Client
}

func New(ctx context.Context, url string) (*Store, error) {
	if url == "" {
		return nil, fmt.Errorf("%w: URL is required", ErrConfig)
	}

	opts, err := goredis.ParseURL(url)
	if err != nil {
		return nil, fmt.Errorf("%w: parse: %v", ErrConfig, err)
	}

	opts.PoolSize = poolSize
	opts.MinIdleConns = minIdleConns
	opts.DialTimeout = dialTimeout
	opts.ReadTimeout = readTimeout
	opts.WriteTimeout = writeTimeout
	opts.PoolTimeout = poolTimeout

	s := &Store{Client: goredis.NewClient(opts)}

	if err := s.Ping(ctx); err != nil {
		s.Client.Close()
		return nil, err
	}

	slog.Info("redis connected")
	return s, nil
}

// Ping checks Redis reachability. Returns ErrUnavailable wrapped with the
// underlying error on failure
func (s *Store) Ping(ctx context.Context) error {
	if err := s.Client.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("%w: %v", ErrUnavailable, err)
	}
	return nil
}

func (s *Store) Close() error {
	if s.Client == nil {
		return nil
	}
	err := s.Client.Close()
	slog.Info("redis closed")
	return err
}