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
	envHTTPAddr = "QUERY_API_HTTP_ADDR"

	envKafkaBrokers          = "KAFKA_BROKERS"
	envKafkaDecisionsTopic   = "KAFKA_DECISIONS_TOPIC"
	envQueryAPIConsumerGroup = "QUERY_API_CONSUMER_GROUP"
	envVisitorTTL            = "QUERY_VISITOR_TTL"
	envRecentRingCapacity    = "QUERY_RECENT_RING_CAPACITY"
)

const (
	defaultHTTPAddr           = ":8083"
	defaultVisitorTTL         = 30 * time.Minute
	defaultRecentRingCapacity = 1 << 17 // 131,072 retained decisions

	startupTimeout  = 10 * time.Second
	shutdownTimeout = 30 * time.Second

	readTimeout  = 10 * time.Second
	writeTimeout = 10 * time.Second
	idleTimeout  = 60 * time.Second
)

type appConfig struct {
	databaseURL string
	http        httpConfig
	kafka       kafkaConfig
	store       storeConfig
}

type httpConfig struct {
	addr string
}

type kafkaConfig struct {
	brokers        []string
	decisionsTopic string
	consumerGroup  string
}

type storeConfig struct {
	visitorTTL   time.Duration
	ringCapacity int
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

	storeCfg, err := loadStoreConfig()
	if err != nil {
		return appConfig{}, err
	}

	return appConfig{
		databaseURL: databaseURL,
		http: httpConfig{
			addr: config.EnvOrDefault(envHTTPAddr, defaultHTTPAddr),
		},
		kafka: kafkaCfg,
		store: storeCfg,
	}, nil
}

func loadKafkaConfig() (kafkaConfig, error) {
	brokers := config.SplitCSV(os.Getenv(envKafkaBrokers))
	topic := strings.TrimSpace(os.Getenv(envKafkaDecisionsTopic))
	group := strings.TrimSpace(os.Getenv(envQueryAPIConsumerGroup))

	if len(brokers) == 0 || topic == "" || group == "" {
		return kafkaConfig{}, errors.New(
			"KAFKA_BROKERS, KAFKA_DECISIONS_TOPIC, and QUERY_API_CONSUMER_GROUP are required",
		)
	}

	return kafkaConfig{
		brokers:        brokers,
		decisionsTopic: topic,
		consumerGroup:  group,
	}, nil
}

func loadStoreConfig() (storeConfig, error) {
	ttl, err := config.ParseDurationEnv(envVisitorTTL, defaultVisitorTTL)
	if err != nil {
		return storeConfig{}, err
	}

	capacity, err := config.ParseIntEnv(envRecentRingCapacity, defaultRecentRingCapacity)
	if err != nil {
		return storeConfig{}, err
	}

	// RecentRing uses bitwise wrap-around, so deployment overrides must keep
	// the capacity as a power of two.
	if capacity <= 0 || capacity&(capacity-1) != 0 {
		return storeConfig{}, fmt.Errorf("%s must be a positive power of two, got %d", envRecentRingCapacity, capacity)
	}

	return storeConfig{
		visitorTTL:   ttl,
		ringCapacity: capacity,
	}, nil
}