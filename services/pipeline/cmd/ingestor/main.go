package main

import (
	"log/slog"
	"os"

	"github.com/JDucr17/streamline/services/pipeline/internal/logging"
)

func main() {
	logging.ConfigureFromEnv()

	if err := run(); err != nil {
		slog.Error("ingestor exited with error", slog.Any("error", err))
		os.Exit(1)
	}
}

func run() error {
	cfg, err := loadConfig()
	if err != nil {
		return err
	}

	producer, closeProducer, err := setupProducer(cfg.kafka)
	if err != nil {
		return err
	}
	defer closeProducer()

	server := setupServer(cfg.http.addr, producer)

	return runWithShutdown(server, cfg)
}