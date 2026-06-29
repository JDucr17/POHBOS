package main

import (
	"context"
	"fmt"
	"net/http"

	"github.com/JDucr17/streamline/services/pipeline/internal/app/queryapi"
	"github.com/JDucr17/streamline/services/pipeline/internal/broker"
	"github.com/JDucr17/streamline/services/pipeline/internal/httpapi"
	"github.com/JDucr17/streamline/services/pipeline/internal/httpapi/middleware"
	"github.com/JDucr17/streamline/services/pipeline/internal/postgres"
)

func setupStorage(databaseURL string) (*postgres.DB, func(), error) {
	ctx, cancel := context.WithTimeout(context.Background(), startupTimeout)
	defer cancel()

	db, err := postgres.New(ctx, databaseURL)
	if err != nil {
		return nil, nil, fmt.Errorf("create postgres pool: %w", err)
	}

	return db, db.Close, nil
}

func setupStore(cfg storeConfig) (*queryapi.Store, error) {
	store, err := queryapi.NewStore(cfg.visitorTTL, cfg.ringCapacity)
	if err != nil {
		return nil, fmt.Errorf("create query store: %w", err)
	}

	return store, nil
}

func setupConsumer(cfg kafkaConfig, store *queryapi.Store, hub *queryapi.Hub) (*queryapi.Consumer, func(), error) {
	consumer, err := broker.NewConsumer(cfg.brokers, cfg.decisionsTopic, cfg.consumerGroup)
	if err != nil {
		return nil, nil, fmt.Errorf("create query api consumer: %w", err)
	}

	return queryapi.NewConsumer(consumer, store, hub), consumer.Close, nil
}

// setupServer wires HTTP routes to the Query API handlers. The API reads from
// memory first where appropriate and falls back to Postgres through Reader.
func setupServer(cfg httpConfig, sse sseConfig, store *queryapi.Store, db *postgres.DB, hub *queryapi.Hub) *http.Server {
	reader := queryapi.NewReader(db)
	api := queryapi.NewAPI(store, reader, hub, sse.backfillLimit, sse.allowedOrigin)

	mux := http.NewServeMux()
	mux.Handle("GET /health", httpapi.Handler(api.HandleHealth))
	mux.Handle("GET /decisions/latest", httpapi.Handler(api.HandleLatest))
	mux.Handle("GET /decisions", httpapi.Handler(api.HandleList))
	mux.Handle("GET /decisions/stream", httpapi.Handler(api.HandleStream))

	return &http.Server{
		Addr:         cfg.addr,
		Handler:      middleware.Logging(mux),
		ReadTimeout:  readTimeout,
		WriteTimeout: writeTimeout,
		IdleTimeout:  idleTimeout,
	}
}
