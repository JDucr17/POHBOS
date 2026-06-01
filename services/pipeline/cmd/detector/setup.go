package main

import (
	"fmt"
	"log/slog"

	"github.com/JDucr17/streamline/services/pipeline/internal/broker"
)

// setupBrokers builds the consumer for the source topic plus producers
// for decisions and DLQ. The returned cleanup function closes all clients.
func setupBrokers(cfg kafkaConfig) (*broker.Consumer, *broker.Producer, *broker.Producer, func(), error) {
	consumer, err := broker.NewConsumer(cfg.brokers, cfg.sourceTopic, cfg.consumerGroup)
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf("create detector consumer: %w", err)
	}

	decisionsProducer, err := broker.NewProducer(cfg.brokers, cfg.decisionsTopic)
	if err != nil {
		consumer.Close()
		return nil, nil, nil, nil, fmt.Errorf("create decisions producer: %w", err)
	}

	dlqProducer, err := broker.NewProducer(cfg.brokers, cfg.dlqTopic)
	if err != nil {
		consumer.Close()
		closeProducer(decisionsProducer, "decisions producer")
		return nil, nil, nil, nil, fmt.Errorf("create dlq producer: %w", err)
	}

	closeAll := func() {
		consumer.Close()
		closeProducer(decisionsProducer, "decisions producer")
		closeProducer(dlqProducer, "dlq producer")
	}

	return consumer, decisionsProducer, dlqProducer, closeAll, nil
}

func closeProducer(producer *broker.Producer, name string) {
	if err := producer.Close(); err != nil {
		slog.Warn("close producer failed",
			slog.String("producer", name),
			slog.Any("error", err),
		)
	}
}