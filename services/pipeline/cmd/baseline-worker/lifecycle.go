package main

import (
	"context"
	"errors"
	"log/slog"
	"os/signal"
	"syscall"

	"github.com/JDucr17/streamline/services/pipeline/internal/app/baselineworker"
)

// runOnce executes a single fit cycle and exits, the worker's whole lifecycle:
// a scheduler invokes it on a schedule rather than it looping. The
// context cancels on SIGINT/SIGTERM so an in-flight cycle aborts cleanly.
func runOnce(w *baselineworker.Worker, cfg appConfig) error {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	slog.Info("baseline-worker starting",
		slog.String("baseline_signals_topic", cfg.baselineSignalsTopic),
		slog.Int("window_days", cfg.baselineWindowDays),
	)

	if err := w.Run(ctx); err != nil {
		if errors.Is(err, context.Canceled) {
			return nil
		}
		return err
	}

	slog.Info("baseline-worker finished")
	return nil
}
