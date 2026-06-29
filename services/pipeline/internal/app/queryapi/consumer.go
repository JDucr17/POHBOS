package queryapi

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"time"

	"github.com/twmb/franz-go/pkg/kgo"

	"github.com/JDucr17/streamline/services/pipeline/internal/broker"
	"github.com/JDucr17/streamline/services/pipeline/internal/envelope"
)

const (
	pollTimeout      = 1 * time.Second
	evictionInterval = 5 * time.Minute
)

type Consumer struct {
	consumer *broker.Consumer
	store    *Store
	hub      *Hub
}

func NewConsumer(consumer *broker.Consumer, store *Store, hub *Hub) *Consumer {
	return &Consumer{
		consumer: consumer,
		store:    store,
		hub:      hub,
	}
}

// Run consumes decisions until ctx is canceled. A background sweep removes
// expired visitor entries.
func (c *Consumer) Run(ctx context.Context) error {
	go c.runEviction(ctx)

	for {
		if err := c.pollAndProcess(ctx); err != nil {
			if errors.Is(err, context.Canceled) {
				return nil
			}
			return err
		}
	}
}

func (c *Consumer) runEviction(ctx context.Context) {
	ticker := time.NewTicker(evictionInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return

		case <-ticker.C:
			c.store.EvictExpired()
			slog.Debug("visitor eviction sweep complete",
				slog.Int("visitor_count", c.store.VisitorCount()),
			)
		}
	}
}

// pollAndProcess fetches one batch, records decoded decisions in memory,
// and marks fetched records for mark-based autocommit.
func (c *Consumer) pollAndProcess(ctx context.Context) error {
	pollCtx, cancel := context.WithTimeout(ctx, pollTimeout)
	defer cancel()

	fetches := c.consumer.Client.PollFetches(pollCtx)

	if err := ctx.Err(); err != nil {
		return err
	}

	if fetchErrs := fetches.Errors(); len(fetchErrs) > 0 {
		broker.LogFetchErrors(fetchErrs)
	}

	records := fetches.Records()
	if len(records) == 0 {
		return nil
	}

	decisions := decodeDecisions(records)
	c.store.RecordBatch(decisions)
	c.hub.PublishBatch(decisions)

	c.consumer.Client.MarkCommitRecords(records...)
	return nil
}

func decodeDecisions(records []*kgo.Record) []envelope.DecisionEnvelope {
	envelopes := make([]envelope.DecisionEnvelope, 0, len(records))

	for _, record := range records {
		var env envelope.DecisionEnvelope
		if err := json.Unmarshal(record.Value, &env); err != nil {
			slog.Warn("skipped malformed decision in query api",
				slog.String("topic", record.Topic),
				slog.Int("partition", int(record.Partition)),
				slog.Int64("offset", record.Offset),
				slog.Any("error", err),
			)
			continue
		}

		envelopes = append(envelopes, env)
	}

	return envelopes
}
