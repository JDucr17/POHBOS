package main

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
)

func runWithShutdown(r runner, cfg appConfig) error {
	runCtx, cancel := context.WithCancel(context.Background())
	defer cancel()

	runErr := make(chan error, 1)
	go func() {
		slog.Info("sink ready",
			slog.String("target", cfg.target.name),
			slog.String("source_topic", cfg.target.sourceTopic),
			slog.String("dlq_topic", cfg.kafka.dlqTopic),
			slog.String("group", cfg.target.consumerGroup),
		)
		runErr <- r.Run(runCtx)
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(sigCh)

	select {
	case <-sigCh:
		slog.Info("sink shutting down",
			slog.String("target", cfg.target.name),
		)
		cancel()
		return waitForRun(runErr)

	case err := <-runErr:
		if errors.Is(err, context.Canceled) {
			return nil
		}
		return err
	}
}

func waitForRun(runErr <-chan error) error {
	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	select {
	case err := <-runErr:
		if errors.Is(err, context.Canceled) {
			return nil
		}
		return err

	case <-ctx.Done():
		return errors.New("shutdown timed out waiting for Run to return")
	}
}