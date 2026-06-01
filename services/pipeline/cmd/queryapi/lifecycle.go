package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"golang.org/x/sync/errgroup"

	"github.com/JDucr17/streamline/services/pipeline/internal/app/queryapi"
)

// runWithShutdown runs the HTTP server and Kafka consumer as one unit.
// A signal or component failure cancels the shared context and waits for
// both goroutines to exit before returning.
func runWithShutdown(server *http.Server, consumer *queryapi.Consumer, cfg appConfig) error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	group, groupCtx := errgroup.WithContext(ctx)

	slog.Info("query api starting",
		slog.String("addr", cfg.http.addr),
		slog.String("decisions_topic", cfg.kafka.decisionsTopic),
		slog.String("group", cfg.kafka.consumerGroup),
		slog.Duration("visitor_ttl", cfg.store.visitorTTL),
		slog.Int("ring_capacity", cfg.store.ringCapacity),
	)

	group.Go(func() error {
		return consumer.Run(groupCtx)
	})

	group.Go(func() error {
		err := server.ListenAndServe()
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return err
	})

	group.Go(func() error {
		<-groupCtx.Done()

		slog.Info("query api shutting down")

		shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
		defer cancel()

		return server.Shutdown(shutdownCtx)
	})

	if err := group.Wait(); err != nil && !errors.Is(err, context.Canceled) {
		return err
	}

	return nil
}