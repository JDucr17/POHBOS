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

// runWithShutdown starts the detector and the baseline-signal consumer, then
// blocks until an OS signal arrives or either loop returns. On shutdown it
// cancels both and waits up to shutdownTimeout for them to return. Any
// uncommitted records may redeliver on next start.
func runWithShutdown(det *detector.Detector, signals *detector.BaselineSignalConsumer, cfg kafkaConfig) error {
	runCtx, cancel := context.WithCancel(context.Background())
	defer cancel()

	runErr := make(chan error, 2)

	slog.Info("detector starting",
		slog.String("source_topic", cfg.sourceTopic),
		slog.String("decisions_topic", cfg.decisionsTopic),
		slog.String("dlq_topic", cfg.dlqTopic),
		slog.String("baseline_signals_topic", cfg.baselineSignalsTopic),
		slog.String("events_group", cfg.consumerGroup),
		slog.String("baseline_signals_group", cfg.baselineSignalsGroup),
	)

	go func() { runErr <- det.Run(runCtx) }()
	go func() { runErr <- signals.Run(runCtx) }()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(sigCh)

	var firstErr error
	oneReturned := false

	select {
	case <-sigCh:
		slog.Info("detector shutting down")
	case firstErr = <-runErr:
		oneReturned = true
	}

	cancel()
	drainErr := drainAll(runErr, oneReturned)

	if firstErr != nil && !errors.Is(firstErr, context.Canceled) {
		return firstErr
	}
	return drainErr
}

// drainAll waits for the still-running loops to return, bounded by
// shutdownTimeout. If oneAlreadyReturned, one loop remains; otherwise two.
func drainAll(runErr <-chan error, oneAlreadyReturned bool) error {
	remaining := 2
	if oneAlreadyReturned {
		remaining = 1
	}

	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	for remaining > 0 {
		select {
		case err := <-runErr:
			remaining--
			if err != nil && !errors.Is(err, context.Canceled) {
				return err
			}
		case <-ctx.Done():
			return errors.New("shutdown timed out waiting for run loops to return")
		}
	}
	return nil
}