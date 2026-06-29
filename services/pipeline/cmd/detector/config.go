package main

import (
	"fmt"
	"time"

	"github.com/JDucr17/streamline/services/pipeline/internal/config"
)

const (
	startupTimeout  = 10 * time.Second
	shutdownTimeout = 30 * time.Second
)

const (
	envKafkaBrokers          = "KAFKA_BROKERS"
	envKafkaRawEventsTopic   = "KAFKA_RAW_EVENTS_TOPIC"
	envKafkaDecisionsTopic   = "KAFKA_DECISIONS_TOPIC"
	envKafkaDLQTopic         = "KAFKA_DEAD_LETTER_TOPIC"
	envDetectorConsumerGroup = "DETECTOR_CONSUMER_GROUP"
	envBaselineSignalsTopic  = "KAFKA_BASELINE_SIGNALS_TOPIC"
	envBaselineSignalsGroup  = "DETECTOR_BASELINE_SIGNALS_GROUP"
	envDatabaseURL           = "DATABASE_URL"
)

// Window env vars are the pipeline-level training/serving contract: the baseline
// worker reads the same names to fit under, and the detector resolves them into
// its serving spec. Both agree by construction; the cache guard still checks
// each loaded baseline's stored spec against the detector's resolved spec.
const (
	envActivityWindowSeconds = "ACTIVITY_WINDOW_SECONDS"
	envWindowCadenceSeconds  = "WINDOW_CADENCE_SECONDS"
	envMinWindowEvents       = "MIN_WINDOW_EVENTS"
)

const (
	defaultActivityWindowSeconds = 300
	defaultWindowCadenceSeconds  = 60
	defaultMinWindowEvents       = 3
)

type kafkaConfig struct {
	brokers              []string
	sourceTopic          string
	decisionsTopic       string
	dlqTopic             string
	consumerGroup        string
	baselineSignalsTopic string
	baselineSignalsGroup string
}

// storeConfig holds the connection string for the detector's Postgres store,
// which serves baselines and the active policy.
type storeConfig struct {
	databaseURL string
}

func loadKafkaConfig() (kafkaConfig, error) {
	var req config.RequiredVars

	cfg := kafkaConfig{
		brokers:              req.CSV(envKafkaBrokers),
		sourceTopic:          req.Get(envKafkaRawEventsTopic),
		decisionsTopic:       req.Get(envKafkaDecisionsTopic),
		dlqTopic:             req.Get(envKafkaDLQTopic),
		consumerGroup:        req.Get(envDetectorConsumerGroup),
		baselineSignalsTopic: req.Get(envBaselineSignalsTopic),
		baselineSignalsGroup: req.Get(envBaselineSignalsGroup),
	}

	if err := req.Err(); err != nil {
		return kafkaConfig{}, err
	}
	return cfg, nil
}

func loadStoreConfig() (storeConfig, error) {
	databaseURL, err := config.RequiredEnv(envDatabaseURL)
	if err != nil {
		return storeConfig{}, err
	}
	return storeConfig{databaseURL: databaseURL}, nil
}

// windowConfig is the detector's resolved serving window contract, sourced from
// the same env vars the baseline worker fits under.
type windowConfig struct {
	activityWindowSeconds int
	windowCadenceSeconds  int
	minWindowEvents       int
}

func loadWindowConfig() (windowConfig, error) {
	var opt config.OptionalVars
	activityWindow := opt.IntAtLeast(envActivityWindowSeconds, defaultActivityWindowSeconds, 1)
	cadence := opt.IntAtLeast(envWindowCadenceSeconds, defaultWindowCadenceSeconds, 1)
	minEvents := opt.IntAtLeast(envMinWindowEvents, defaultMinWindowEvents, 1)
	if err := opt.Err(); err != nil {
		return windowConfig{}, err
	}

	// Cadence must not exceed the window it samples.
	if cadence > activityWindow {
		return windowConfig{}, fmt.Errorf("%s (%d) must be <= %s (%d)",
			envWindowCadenceSeconds, cadence, envActivityWindowSeconds, activityWindow)
	}

	return windowConfig{
		activityWindowSeconds: activityWindow,
		windowCadenceSeconds:  cadence,
		minWindowEvents:       minEvents,
	}, nil
}