package main

import (
	"log/slog"
	"os"
	"time"

	"github.com/JDucr17/streamline/services/pipeline/internal/app/baselineworker"
	"github.com/JDucr17/streamline/services/pipeline/internal/extractor"
	"github.com/JDucr17/streamline/services/pipeline/internal/logging"
)

func main() {
	logging.ConfigureFromEnv()

	if err := run(); err != nil {
		slog.Error("baseline-worker exited with error", slog.Any("error", err))
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

	producer, closeProducer, err := setupProducer(cfg.kafkaBrokers, cfg.baselineSignalsTopic)
	if err != nil {
		return err
	}
	defer closeProducer()

	spec := extractor.WindowSpec{
		Length:    time.Duration(cfg.activityWindowSeconds) * time.Second,
		Cadence:   time.Duration(cfg.windowCadenceSeconds) * time.Second,
		MinEvents: cfg.minWindowEvents,
	}
	worker := baselineworker.NewWorker(
		db, 
		producer, 
		cfg.baselineWindowDays, 
		spec, 
		cfg.minBaselineWindows, 
		cfg.minBaselineVisitors)
	return runOnce(worker, cfg)
}
