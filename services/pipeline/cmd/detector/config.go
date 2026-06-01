package main

import (
	"errors"
	"os"
	"strings"
	"time"

	"github.com/JDucr17/streamline/services/pipeline/internal/config"
)

const shutdownTimeout = 30 * time.Second

const (
	envKafkaBrokers          = "KAFKA_BROKERS"
	envKafkaRawEventsTopic   = "KAFKA_RAW_EVENTS_TOPIC"
	envKafkaDecisionsTopic   = "KAFKA_DECISIONS_TOPIC"
	envKafkaDLQTopic         = "KAFKA_DEAD_LETTER_TOPIC"
	envDetectorConsumerGroup = "DETECTOR_CONSUMER_GROUP"
)

type kafkaConfig struct {
	brokers        []string
	sourceTopic    string
	decisionsTopic string
	dlqTopic       string
	consumerGroup  string
}

func loadKafkaConfig() (kafkaConfig, error) {
	source := strings.TrimSpace(os.Getenv(envKafkaRawEventsTopic))
	decisions := strings.TrimSpace(os.Getenv(envKafkaDecisionsTopic))
	dlq := strings.TrimSpace(os.Getenv(envKafkaDLQTopic))
	group := strings.TrimSpace(os.Getenv(envDetectorConsumerGroup))

	brokers := config.SplitCSV(os.Getenv(envKafkaBrokers))

	if len(brokers) == 0 || source == "" || decisions == "" || dlq == "" || group == "" {
		return kafkaConfig{}, errors.New(
			"KAFKA_BROKERS, KAFKA_RAW_EVENTS_TOPIC, KAFKA_DECISIONS_TOPIC, KAFKA_DEAD_LETTER_TOPIC, and DETECTOR_CONSUMER_GROUP are required",
		)
	}

	return kafkaConfig{
		brokers:        brokers,
		sourceTopic:    source,
		decisionsTopic: decisions,
		dlqTopic:       dlq,
		consumerGroup:  group,
	}, nil
}