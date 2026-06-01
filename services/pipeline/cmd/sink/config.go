package main

import (
	"errors"
	"fmt"
	"os"
	"strings"
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
	brokers := config.SplitCSV(os.Getenv(envKafkaBrokers))
	dlqTopic := strings.TrimSpace(os.Getenv(envKafkaDLQTopic))

	if len(brokers) == 0 || dlqTopic == "" {
		return kafkaConfig{}, errors.New("KAFKA_BROKERS and KAFKA_DEAD_LETTER_TOPIC are required")
	}

	return kafkaConfig{
		brokers:  brokers,
		dlqTopic: dlqTopic,
	}, nil
}

func loadSinkTarget() (sinkTarget, error) {
	target := strings.TrimSpace(os.Getenv(envSinkTarget))

	switch target {
	case "events":
		topic := strings.TrimSpace(os.Getenv(envKafkaRawEventsTopic))
		group := strings.TrimSpace(os.Getenv(envSinkEventsConsumerGroup))

		if topic == "" || group == "" {
			return sinkTarget{}, errors.New(
				"KAFKA_RAW_EVENTS_TOPIC and SINK_EVENTS_CONSUMER_GROUP are required for SINK_TARGET=events",
			)
		}

		return sinkTarget{
			name:          "events",
			sourceTopic:   topic,
			consumerGroup: group,
		}, nil

	case "decisions":
		topic := strings.TrimSpace(os.Getenv(envKafkaDecisionsTopic))
		group := strings.TrimSpace(os.Getenv(envSinkDecisionsConsumerGroup))

		if topic == "" || group == "" {
			return sinkTarget{}, errors.New(
				"KAFKA_DECISIONS_TOPIC and SINK_DECISIONS_CONSUMER_GROUP are required for SINK_TARGET=decisions",
			)
		}

		return sinkTarget{
			name:          "decisions",
			sourceTopic:   topic,
			consumerGroup: group,
		}, nil

	default:
		return sinkTarget{}, fmt.Errorf("invalid SINK_TARGET %q; expected events or decisions", target)
	}
}