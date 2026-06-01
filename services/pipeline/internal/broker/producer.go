package broker

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/twmb/franz-go/pkg/kgo"
)

var ErrPublish = errors.New("broker: publish failed")

type Producer struct {
	client *kgo.Client
	topic  string
}

func NewProducer(brokers []string, topic string) (*Producer, error) {
	client, err := kgo.NewClient(
		kgo.SeedBrokers(brokers...),
		kgo.DefaultProduceTopic(topic),
		kgo.RequiredAcks(kgo.AllISRAcks()),
		kgo.RecordPartitioner(kgo.StickyKeyPartitioner(nil)),
	)
	if err != nil {
		return nil, fmt.Errorf("create kafka client: %w", err)
	}

	return &Producer{
		client: client,
		topic:  topic,
	}, nil
}

func (p *Producer) Publish(ctx context.Context, key, value []byte) error {
	record := &kgo.Record{
		Topic: p.topic,
		Key:   key,
		Value: value,
	}

	if err := p.client.ProduceSync(ctx, record).FirstErr(); err != nil {
		return fmt.Errorf("%w: %v", ErrPublish, err)
	}

	return nil
}

func (p *Producer) Ping(ctx context.Context) error {
    if err := p.client.Ping(ctx); err != nil {
        return fmt.Errorf("%w: %v", ErrPublish, err)
    }
    return nil
}

func (p *Producer) Close() error {
	p.client.Close()
	slog.Info("kafka producer closed")
	return nil
}