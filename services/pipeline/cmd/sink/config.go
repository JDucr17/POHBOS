package main

import (
	"fmt"
	"time"

	"github.com/JDucr17/streamline/services/pipeline/internal/config"
)

const (
	envKafkaBrokers        = "KAFKA_BROKERS"
	envKafkaRawEventsTopic = "KAFKA_RAW_EVENTS_TOPIC"
	envKafkaDecisionsTopic = "KAFKA_DECISIONS_TOPIC"
	envKafkaDLQTopic       = "KAFKA_DEAD_LETTER_TOPIC"

	envSinkTarget                 = "SINK_TARGET"
	envSinkEventsConsumerGroup    = "SINK_EVENTS_CONSUMER_GROUP"
	envSinkDecisionsConsumerGroup = "SINK_DECISIONS_CONSUMER_GROUP"
)

const (
	sinkTargetEvents    = "events"
	sinkTargetDecisions = "decisions"
)

const (
	startupTimeout  = 10 * time.Second
	shutdownTimeout = 30 * time.Second
)

type appConfig struct {
	databaseURL string
	kafka       kafkaConfig
	target      sinkTarget
}

type kafkaConfig struct {
	brokers  []string
	dlqTopic string
}

type sinkTarget struct {
	name          string
	sourceTopic   string
	consumerGroup string
}

func loadConfig() (appConfig, error) {
	databaseURL, err := config.RequiredEnv("DATABASE_URL")
	if err != nil {
		return appConfig{}, err
	}

	kafkaCfg, err := loadKafkaConfig()
	if err != nil {
		return appConfig{}, err
	}

	target, err := loadSinkTarget()
	if err != nil {
		return appConfig{}, err
	}

	return appConfig{
		databaseURL: databaseURL,
		kafka:       kafkaCfg,
		target:      target,
	}, nil
}

func loadKafkaConfig() (kafkaConfig, error) {
	var req config.RequiredVars

	cfg := kafkaConfig{
		brokers:  req.CSV(envKafkaBrokers),
		dlqTopic: req.Get(envKafkaDLQTopic),
	}

	if err := req.Err(); err != nil {
		return kafkaConfig{}, err
	}

	return cfg, nil
}

func loadSinkTarget() (sinkTarget, error) {
	target, err := config.RequiredEnv(envSinkTarget)
	if err != nil {
		return sinkTarget{}, err
	}

	switch target {
	case sinkTargetEvents:
		var req config.RequiredVars

		cfg := sinkTarget{
			name:          sinkTargetEvents,
			sourceTopic:   req.Get(envKafkaRawEventsTopic),
			consumerGroup: req.Get(envSinkEventsConsumerGroup),
		}

		if err := req.Err(); err != nil {
			return sinkTarget{}, fmt.Errorf("load events sink config: %w", err)
		}

		return cfg, nil

	case sinkTargetDecisions:
		var req config.RequiredVars

		cfg := sinkTarget{
			name:          sinkTargetDecisions,
			sourceTopic:   req.Get(envKafkaDecisionsTopic),
			consumerGroup: req.Get(envSinkDecisionsConsumerGroup),
		}

		if err := req.Err(); err != nil {
			return sinkTarget{}, fmt.Errorf("load decisions sink config: %w", err)
		}

		return cfg, nil

	default:
		return sinkTarget{}, fmt.Errorf(
			"invalid %s %q; expected %s or %s",
			envSinkTarget,
			target,
			sinkTargetEvents,
			sinkTargetDecisions,
		)
	}
}