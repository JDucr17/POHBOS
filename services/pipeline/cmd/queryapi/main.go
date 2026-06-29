package main

import (
	"log/slog"
	"os"

	"github.com/JDucr17/streamline/services/pipeline/internal/app/queryapi"
	"github.com/JDucr17/streamline/services/pipeline/internal/logging"
)

func main() {
	logging.ConfigureFromEnv()

	if err := run(); err != nil {
		slog.Error("query api exited with error", slog.Any("error", err))
		os.Exit(1)
	}
}

func run() error {
	cfg, err := loadConfig()
	if err != nil {
		return err
	}

	db, closeDB, err := setupStorage(cfg.databaseURL)
	if err != nil {
		return err
	}
	defer closeDB()

	store, err := setupStore(cfg.store)
	if err != nil {
		return err
	}

	hub := queryapi.NewHub(cfg.sse.maxClients, cfg.sse.clientBuffer)

	consumer, closeConsumer, err := setupConsumer(cfg.kafka, store, hub)
	if err != nil {
		return err
	}
	defer closeConsumer()

	server := setupServer(cfg.http, cfg.sse, store, db, hub)

	return runWithShutdown(server, consumer, cfg)
}
