package main

import (
	"errors"
	"os"
	"strings"
	"time"

	"github.com/JDucr17/streamline/services/pipeline/internal/config"
)

const (
	envHTTPAddr            = "HTTP_ADDR"
	envKafkaBrokers        = "KAFKA_BROKERS"
	envKafkaRawEventsTopic = "KAFKA_RAW_EVENTS_TOPIC"
)

const (
	defaultHTTPAddr = ":8080"

	shutdownTimeout = 30 * time.Second

	readTimeout  = 10 * time.Second
	writeTimeout = 10 * time.Second
	idleTimeout  = 60 * time.Second
)

type appConfig struct {
	http  httpConfig
	kafka kafkaConfig
}

type httpConfig struct {
	addr string
}

type kafkaConfig struct {
	brokers []string
	topic   string
}

func loadConfig() (appConfig, error) {
	kafkaCfg, err := loadKafkaConfig()
	if err != nil {
		return appConfig{}, err
	}

	return appConfig{
		http: httpConfig{
			addr: config.EnvOrDefault(envHTTPAddr, defaultHTTPAddr),
		},
		kafka: kafkaCfg,
	}, nil
}

func loadKafkaConfig() (kafkaConfig, error) {
	topic := strings.TrimSpace(os.Getenv(envKafkaRawEventsTopic))
	brokers := config.SplitCSV(os.Getenv(envKafkaBrokers))

	if len(brokers) == 0 || topic == "" {
		return kafkaConfig{}, errors.New("KAFKA_BROKERS and KAFKA_RAW_EVENTS_TOPIC are required")
	}

	return kafkaConfig{
		brokers: brokers,
		topic:   topic,
	}, nil
}