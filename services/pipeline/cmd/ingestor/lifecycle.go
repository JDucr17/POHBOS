package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

// runWithShutdown starts the HTTP server and blocks until either a shutdown
// signal arrives or the server exits unexpectedly.
func runWithShutdown(server *http.Server, cfg appConfig) error {
	serverErr := make(chan error, 1)

	slog.Info("ingestor starting",
		slog.String("addr", cfg.http.addr),
		slog.String("topic", cfg.kafka.topic),
	)

	go func() {
		err := server.ListenAndServe()
		if errors.Is(err, http.ErrServerClosed) {
			err = nil
		}
		serverErr <- err
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(sigCh)

	select {
	case <-sigCh:
		slog.Info("ingestor shutting down")
		return shutdownServer(server)

	case err := <-serverErr:
		return err
	}
}

// shutdownServer gives in-flight HTTP requests a bounded window to finish.
func shutdownServer(server *http.Server) error {
	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	return server.Shutdown(ctx)
}