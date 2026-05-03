package storage

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/redis/go-redis/v9"
)

var (
	ErrRedisConfig      = errors.New("redis: invalid config")
	ErrRedisUnavailable = errors.New("redis: unavailable")
)

// Constants for redis connection pool
// Aggressive timeouts because Redis sits inside the synchronous decision path.
const (
	redisPoolSize     = 50
	redisMinIdleConns = 10
	redisDialTimeout  = 5 * time.Second
	redisReadTimeout  = 3 * time.Second
	redisWriteTimeout = 3 * time.Second
	redisPoolTimeout  = 4 * time.Second
)

type Redis struct {
	Client *redis.Client
}

func NewRedis(ctx context.Context, url string) (*Redis, error) {
	if url == "" {
		return nil, fmt.Errorf("%w: URL is required", ErrRedisConfig)
	}

	opts, err := redis.ParseURL(url)
	if err != nil {
		return nil, fmt.Errorf("%w: parse: %v", ErrRedisConfig, err)
	}

	opts.PoolSize = redisPoolSize
	opts.MinIdleConns = redisMinIdleConns
	opts.DialTimeout = redisDialTimeout
	opts.ReadTimeout = redisReadTimeout
	opts.WriteTimeout = redisWriteTimeout
	opts.PoolTimeout = redisPoolTimeout

	client := redis.NewClient(opts)

	if err := client.Ping(ctx).Err(); err != nil {
		client.Close()
		return nil, fmt.Errorf("%w: ping: %v", ErrRedisUnavailable, err)
	}

	slog.Info("redis connected")
	return &Redis{Client: client}, nil
}

func (r *Redis) Close() error {
	if r.Client == nil {
		return nil
	}
	err := r.Client.Close()
	slog.Info("redis closed")
	return err
}