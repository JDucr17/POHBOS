package main

import (
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
	var req config.RequiredVars

	cfg := kafkaConfig{
		brokers: req.CSV(envKafkaBrokers),
		topic:   req.Get(envKafkaRawEventsTopic),
	}

	if err := req.Err(); err != nil {
		return kafkaConfig{}, err
	}
	return cfg, nil
}