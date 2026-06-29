package main

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/JDucr17/streamline/services/pipeline/internal/app/detector"
	"github.com/JDucr17/streamline/services/pipeline/internal/broker"
	"github.com/JDucr17/streamline/services/pipeline/internal/extractor"
	"github.com/JDucr17/streamline/services/pipeline/internal/hbos"
	"github.com/JDucr17/streamline/services/pipeline/internal/postgres"
)

// clients holds every dependency the detector and signal consumer need, plus a
// cleanup that closes them.
type clients struct {
	events    *broker.Consumer
	signals   *broker.Consumer
	decisions *broker.Producer
	dlq       *broker.Producer
	db        *postgres.DB
	cache     *detector.BaselineCache
	close     func()
}

// brokerClients bundles the detector's Kafka clients so construction and
// cleanup move together.
type brokerClients struct {
	events    *broker.Consumer
	signals   *broker.Consumer
	decisions *broker.Producer
	dlq       *broker.Producer
}

func (b brokerClients) close() {
	if b.events != nil {
		b.events.Close()
	}
	if b.signals != nil {
		b.signals.Close()
	}
	if b.decisions != nil {
		closeProducer(b.decisions, "decisions producer")
	}
	if b.dlq != nil {
		closeProducer(b.dlq, "dlq producer")
	}
}

// setupClients assembles the Kafka clients, the Postgres pool, and the baseline
// cache into the dependencies run() needs. On any failure it closes whatever
// was already built before returning.
func setupClients(ctx context.Context, kafkaCfg kafkaConfig, storeCfg storeConfig, windowCfg windowConfig) (clients, error) {
	brokers, err := setupBrokers(kafkaCfg)
	if err != nil {
		return clients{}, err
	}

	db, err := setupStore(ctx, storeCfg)
	if err != nil {
		brokers.close()
		return clients{}, err
	}

	cache, err := setupBaselineCache(ctx, db, windowCfg)
	if err != nil {
		db.Close()
		brokers.close()
		return clients{}, err
	}

	return clients{
		events:    brokers.events,
		signals:   brokers.signals,
		decisions: brokers.decisions,
		dlq:       brokers.dlq,
		db:        db,
		cache:     cache,
		close: func() {
			db.Close()
			brokers.close()
		},
	}, nil
}

// setupBrokers builds the event and baseline-signal consumers plus the
// decisions and DLQ producers. On any failure every client built so far is
// closed before returning.
func setupBrokers(cfg kafkaConfig) (brokerClients, error) {
	var brokers brokerClients
	ok := false
	defer func() {
		if !ok {
			brokers.close()
		}
	}()

	var err error

	brokers.events, err = broker.NewConsumer(cfg.brokers, cfg.sourceTopic, cfg.consumerGroup)
	if err != nil {
		return brokerClients{}, fmt.Errorf("create event consumer: %w", err)
	}

	brokers.signals, err = broker.NewConsumer(cfg.brokers, cfg.baselineSignalsTopic, cfg.baselineSignalsGroup)
	if err != nil {
		return brokerClients{}, fmt.Errorf("create baseline signal consumer: %w", err)
	}

	brokers.decisions, err = broker.NewProducer(cfg.brokers, cfg.decisionsTopic)
	if err != nil {
		return brokerClients{}, fmt.Errorf("create decisions producer: %w", err)
	}

	brokers.dlq, err = broker.NewProducer(cfg.brokers, cfg.dlqTopic)
	if err != nil {
		return brokerClients{}, fmt.Errorf("create dlq producer: %w", err)
	}

	ok = true
	return brokers, nil
}

func setupStore(ctx context.Context, cfg storeConfig) (*postgres.DB, error) {
	db, err := postgres.New(ctx, cfg.databaseURL)
	if err != nil {
		return nil, fmt.Errorf("connect postgres: %w", err)
	}
	return db, nil
}

// setupBaselineCache constructs the cache and loads every current baseline, so
// the returned cache is ready to score against without waiting for a signal.
func setupBaselineCache(ctx context.Context, db *postgres.DB, windowCfg windowConfig) (*detector.BaselineCache, error) {
	cache := detector.NewBaselineCache(db, servingWindowSpec(windowCfg))
	if err := cache.LoadAll(ctx); err != nil {
		return nil, fmt.Errorf("load baselines at startup: %w", err)
	}

	return cache, nil
}

// servingWindowSpec is the detector's serving window contract; the cache guard
// rejects any baseline whose fitted spec disagrees with it. HorizonDays is a
// provenance field, not part of the serving contract, so it is left zero.
func servingWindowSpec(cfg windowConfig) hbos.WindowSpec {
	return hbos.WindowSpec{
		LengthSeconds:  cfg.activityWindowSeconds,
		CadenceSeconds: cfg.windowCadenceSeconds,
		MinEvents:      cfg.minWindowEvents,
	}
}

// servingExtractorSpec is the same serving window in the form the scoring path
// and visitor store consume (EventsInWindow/Eligible/Compute and pruning).
func servingExtractorSpec(cfg windowConfig) extractor.WindowSpec {
	return extractor.WindowSpec{
		Length:    time.Duration(cfg.activityWindowSeconds) * time.Second,
		Cadence:   time.Duration(cfg.windowCadenceSeconds) * time.Second,
		MinEvents: cfg.minWindowEvents,
	}
}

func closeProducer(producer *broker.Producer, name string) {
	if err := producer.Close(); err != nil {
		slog.Warn("close producer failed",
			slog.String("producer", name),
			slog.Any("error", err),
		)
	}
}