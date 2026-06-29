package main

import (
	"context"
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

	storeCfg, err := loadStoreConfig()
	if err != nil {
		return err
	}

	windowCfg, err := loadWindowConfig()
	if err != nil {
		return err
	}

	startupCtx, cancel := context.WithTimeout(context.Background(), startupTimeout)
	defer cancel()

	c, err := setupClients(startupCtx, kafkaCfg, storeCfg, windowCfg)
	if err != nil {
		return err
	}
	defer c.close()

	// No active policy is a configuration error, not a runtime status: refuse to
	// start rather than emit decisions with no policy_version.
	policy, err := detector.LoadActivePolicy(startupCtx, c.db)
	if err != nil {
		return err
	}

	servingSpec := servingExtractorSpec(windowCfg)
	visitors := detector.NewVisitorStore(servingSpec.Length, servingSpec.Length)

	det := detector.NewDetector(c.events, c.decisions, c.dlq, visitors, c.cache, policy, servingSpec)
	signalConsumer := detector.NewBaselineSignalConsumer(c.signals, c.cache)

	return runWithShutdown(det, signalConsumer, kafkaCfg)
}