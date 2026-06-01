package main

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/JDucr17/streamline/services/pipeline/internal/app/ingestor"
	"github.com/JDucr17/streamline/services/pipeline/internal/broker"
	"github.com/JDucr17/streamline/services/pipeline/internal/httpapi"
	"github.com/JDucr17/streamline/services/pipeline/internal/httpapi/middleware"
)

func setupProducer(cfg kafkaConfig) (*broker.Producer, func(), error) {
	producer, err := broker.NewProducer(cfg.brokers, cfg.topic)
	if err != nil {
		return nil, nil, fmt.Errorf("create ingestor producer: %w", err)
	}

	closeProducer := func() {
		if err := producer.Close(); err != nil {
			slog.Warn("close ingestor producer failed", slog.Any("error", err))
		}
	}

	return producer, closeProducer, nil
}

// setupServer assembles the HTTP server with routes and middleware.
func setupServer(addr string, producer *broker.Producer) *http.Server {
	api := &ingestor.API{Producer: producer}

	mux := http.NewServeMux()
	mux.Handle("POST /events", httpapi.Handler(api.HandleEvent))
	mux.Handle("GET /health", httpapi.Handler(api.HandleHealth))

	return &http.Server{
		Addr:         addr,
		Handler:      middleware.Logging(mux),
		ReadTimeout:  readTimeout,
		WriteTimeout: writeTimeout,
		IdleTimeout:  idleTimeout,
	}
}