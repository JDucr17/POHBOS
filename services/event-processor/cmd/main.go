package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/JDucr17/streamline/services/event-processor/internal/storage"
)

const (
	envDatabaseURL = "DATABASE_URL"
	envRedisURL    = "REDIS_URL"
	envLogLevel    = "LOG_LEVEL"
	envLogFormat   = "LOG_FORMAT"

	startupTimeout = 10 * time.Second
)

func main() {
	configureLogging()

	if err := run(); err != nil {
		slog.Error("event-processor exited with error", slog.Any("error", err))
		os.Exit(1)
	}
}

func run() error {
	ctx, cancel := context.WithTimeout(context.Background(), startupTimeout)
	defer cancel()

	pg, err := storage.NewPostgres(ctx, os.Getenv(envDatabaseURL))
	if err != nil {
		return err
	}
	defer pg.Close()

	rdb, err := storage.NewRedis(ctx, os.Getenv(envRedisURL))
	if err != nil {
		return err
	}
	defer rdb.Close()

	slog.Info("event-processor ready")
	waitForShutdown()
	slog.Info("event-processor shutting down")
	return nil
}

func configureLogging() {
	level := slog.LevelInfo
	if os.Getenv(envLogLevel) == "debug" {
		level = slog.LevelDebug
	}

	opts := &slog.HandlerOptions{Level: level}

	var handler slog.Handler
	if os.Getenv(envLogFormat) == "json" {
		handler = slog.NewJSONHandler(os.Stdout, opts)
	} else {
		handler = slog.NewTextHandler(os.Stdout, opts)
	}

	slog.SetDefault(slog.New(handler))
}

func waitForShutdown() {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh
}