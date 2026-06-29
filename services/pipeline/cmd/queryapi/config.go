package main

import (
	"fmt"
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

	envSSEMaxClients    = "QUERY_SSE_MAX_CLIENTS"
	envSSEClientBuffer  = "QUERY_SSE_CLIENT_BUFFER"
	envSSEBackfillLimit = "QUERY_SSE_BACKFILL_LIMIT"
	envSSEAllowOrigin   = "QUERY_SSE_ALLOW_ORIGIN"
)

const (
	defaultHTTPAddr           = ":8083"
	defaultVisitorTTL         = 30 * time.Minute
	defaultRecentRingCapacity = 1 << 17 // 131,072 retained decisions

	defaultSSEMaxClients    = 100
	defaultSSEClientBuffer  = 256
	defaultSSEBackfillLimit = 50

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
	sse         sseConfig
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

type sseConfig struct {
	maxClients    int
	clientBuffer  int
	backfillLimit int
	allowedOrigin string
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

	sseCfg, err := loadSSEConfig(storeCfg.ringCapacity)
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
		sse:   sseCfg,
	}, nil
}

func loadKafkaConfig() (kafkaConfig, error) {
	var req config.RequiredVars

	cfg := kafkaConfig{
		brokers:        req.CSV(envKafkaBrokers),
		decisionsTopic: req.Get(envKafkaDecisionsTopic),
		consumerGroup:  req.Get(envQueryAPIConsumerGroup),
	}

	if err := req.Err(); err != nil {
		return kafkaConfig{}, err
	}
	return cfg, nil
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

func loadSSEConfig(ringCapacity int) (sseConfig, error) {
	var opt config.OptionalVars

	cfg := sseConfig{
		maxClients:    opt.IntAtLeast(envSSEMaxClients, defaultSSEMaxClients, 1),
		clientBuffer:  opt.IntAtLeast(envSSEClientBuffer, defaultSSEClientBuffer, 1),
		backfillLimit: opt.IntAtLeast(envSSEBackfillLimit, defaultSSEBackfillLimit, 1),
		allowedOrigin: config.EnvOrDefault(envSSEAllowOrigin, ""),
	}

	if err := opt.Err(); err != nil {
		return sseConfig{}, err
	}

	// Backfill replays from the recent ring, so it can never ask for more than
	// the ring retains.
	if cfg.backfillLimit > ringCapacity {
		return sseConfig{}, fmt.Errorf("%s (%d) must not exceed %s (%d)",
			envSSEBackfillLimit, cfg.backfillLimit, envRecentRingCapacity, ringCapacity)
	}

	return cfg, nil
}
