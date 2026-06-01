package broker

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/twmb/franz-go/pkg/kgo"
)

var ErrConsume = errors.New("broker: consume failed")

// Consumer holds a kgo.Client configured for group consumption with manual
// offset marking.
type Consumer struct {
	Client *kgo.Client
	topic  string
	group  string
}

func NewConsumer(brokers []string, topic, group string) (*Consumer, error) {
	client, err := kgo.NewClient(
		kgo.SeedBrokers(brokers...),
		kgo.ConsumerGroup(group),
		kgo.ConsumeTopics(topic),
		kgo.Balancers(kgo.CooperativeStickyBalancer()),
		kgo.AutoCommitMarks(),
	)
	if err != nil {
		return nil, fmt.Errorf("create kafka client: %w", err)
	}

	return &Consumer{
		Client: client,
		topic:  topic,
		group:  group,
	}, nil
}

// Close leaves the consumer group and closes the client. The default
// OnPartitionsRevoked callback commits marked offsets during the leave-
// group rebalance. If a future change overrides OnPartitionsRevoked, the
// caller must commit marked offsets explicitly before calling Close
func (c *Consumer) Close() {
	c.Client.Close()
	slog.Info("kafka consumer closed",
		slog.String("topic", c.topic),
		slog.String("group", c.group),
	)
}

// LogFetchErrors logs unexpected fetch errors. Context cancellation and
// deadline expiration are filtered out, those are expected signals from
// PollFetches that the bounded poll context elapsed, not real errors
func LogFetchErrors(errs []kgo.FetchError) {
    for _, e := range errs {
        if errors.Is(e.Err, context.Canceled) || errors.Is(e.Err, context.DeadlineExceeded) {
            continue
        }
        slog.Error("kafka fetch error",
            slog.String("topic", e.Topic),
            slog.Int("partition", int(e.Partition)),
            slog.Any("error", e.Err),
        )
    }
}