package main

import (
	"context"
	"fmt"

	"github.com/JDucr17/streamline/services/pipeline/internal/broker"
	"github.com/JDucr17/streamline/services/pipeline/internal/postgres"
)

func setupStorage(databaseURL string) (*postgres.DB, func(), error) {
	ctx, cancel := context.WithTimeout(context.Background(), startupTimeout)
	defer cancel()

	db, err := postgres.New(ctx, databaseURL)
	if err != nil {
		return nil, nil, fmt.Errorf("create postgres pool: %w", err)
	}

	return db, db.Close, nil
}

func setupProducer(brokers []string, topic string) (*broker.Producer, func(), error) {
	producer, err := broker.NewProducer(brokers, topic)
	if err != nil {
		return nil, nil, fmt.Errorf("create baseline signal producer: %w", err)
	}

	return producer, func() { producer.Close() }, nil
}
