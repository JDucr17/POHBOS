package main

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/JDucr17/streamline/services/pipeline/internal/app/detector"
)

// runWithShutdown starts the detector and blocks until signal or run error.
// On signal, cancels the run context and waits up to shutdownTimeout for
// Run to return. Any uncommitted records may redeliver on next start.
func runWithShutdown(det *detector.Detector, cfg kafkaConfig) error {
	runCtx, cancel := context.WithCancel(context.Background())
	defer cancel()

	runErr := make(chan error, 1)

	slog.Info("detector starting",
		slog.String("source_topic", cfg.sourceTopic),
		slog.String("decisions_topic", cfg.decisionsTopic),
		slog.String("dlq_topic", cfg.dlqTopic),
		slog.String("group", cfg.consumerGroup),
	)

	go func() {
		runErr <- det.Run(runCtx)
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(sigCh)

	select {
	case <-sigCh:
		slog.Info("detector shutting down")
		cancel()
		return waitForRun(runErr)

	case err := <-runErr:
		if errors.Is(err, context.Canceled) {
			return nil
		}
		return err
	}
}

// waitForRun blocks until Run returns or shutdownTimeout elapses.
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