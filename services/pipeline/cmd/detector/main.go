package main

import (
	"log/slog"
	"os"

	"github.com/JDucr17/streamline/services/pipeline/internal/app/detector"
	"github.com/JDucr17/streamline/services/pipeline/internal/logging"
)

func main() {
	logging.ConfigureFromEnv()

	if err := run(); err != nil {
		slog.Error("detector exited with error", slog.Any("error", err))
		os.Exit(1)
	}
}

func run() error {
	kafkaCfg, err := loadKafkaConfig()
	if err != nil {
		return err
	}

	consumer, decisionsProducer, dlqProducer, closeBrokers, err := setupBrokers(kafkaCfg)
	if err != nil {
		return err
	}
	defer closeBrokers()

	det := detector.NewDetector(consumer, decisionsProducer, dlqProducer)

	return runWithShutdown(det, kafkaCfg)
}