package main

import (
	"log/slog"
	"os"

	"github.com/JDucr17/streamline/services/pipeline/internal/logging"
)

func main() {
	logging.ConfigureFromEnv()

	if err := run(); err != nil {
		slog.Error("sink exited with error", slog.Any("error", err))
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

	consumer, dlqProducer, closeBrokers, err := setupBrokers(cfg)
	if err != nil {
		return err
	}
	defer closeBrokers()

	runner, err := buildSink(cfg.target, db, consumer, dlqProducer)
	if err != nil {
		return err
	}

	return runWithShutdown(runner, cfg)
}