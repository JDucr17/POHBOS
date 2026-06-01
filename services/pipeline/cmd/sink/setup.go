package main

import (
	"context"
	"fmt"

	"github.com/JDucr17/streamline/services/pipeline/internal/app/sink"
	"github.com/JDucr17/streamline/services/pipeline/internal/broker"
	"github.com/JDucr17/streamline/services/pipeline/internal/postgres"
)

type runner interface {
	Run(context.Context) error
}

func setupStorage(databaseURL string) (*postgres.DB, func(), error) {
	ctx, cancel := context.WithTimeout(context.Background(), startupTimeout)
	defer cancel()

	db, err := postgres.New(ctx, databaseURL)
	if err != nil {
		return nil, nil, fmt.Errorf("create postgres pool: %w", err)
	}

	return db, db.Close, nil
}

func setupBrokers(cfg appConfig) (*broker.Consumer, *broker.Producer, func(), error) {
	consumer, err := broker.NewConsumer(
		cfg.kafka.brokers,
		cfg.target.sourceTopic,
		cfg.target.consumerGroup,
	)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("create sink consumer: %w", err)
	}

	dlqProducer, err := broker.NewProducer(cfg.kafka.brokers, cfg.kafka.dlqTopic)
	if err != nil {
		consumer.Close()
		return nil, nil, nil, fmt.Errorf("create dlq producer: %w", err)
	}

	closeAll := func() {
		consumer.Close()
		dlqProducer.Close()
	}

	return consumer, dlqProducer, closeAll, nil
}

func buildSink(
	target sinkTarget,
	db *postgres.DB,
	consumer *broker.Consumer,
	dlqProducer *broker.Producer,
) (runner, error) {
	switch target.name {
	case "events":
		projection := sink.NewEventProjection(sink.NewEventWriter(db))
		return sink.NewSink[sink.EventInsert](consumer, dlqProducer, projection), nil

	case "decisions":
		projection := sink.NewDecisionProjection(sink.NewDecisionWriter(db))
		return sink.NewSink[sink.DecisionInsert](consumer, dlqProducer, projection), nil

	default:
		return nil, fmt.Errorf("unsupported sink target %q", target.name)
	}
}