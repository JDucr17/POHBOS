package detector

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"time"

	"github.com/twmb/franz-go/pkg/kgo"

	"github.com/JDucr17/streamline/services/pipeline/internal/broker"
)

const signalPollTimeout = 1 * time.Second

// baselineSignal is the payload the baseline-worker publishes after it commits
// a new baseline. It names which source changed, the detector reloads that
// source's latest baseline from Postgres in response.
type baselineSignal struct {
	SourceID      string `json:"source_id"`
	BaselineRunID int64  `json:"baseline_run_id"`
}

// BaselineSignalConsumer consumes baseline-update signals and refreshes the
// cache. Signals are hints, not authoritative state: every signal is marked
// consumed even when the refresh fails, so a bad signal never blocks the
// partition. The baseline itself is durable in Postgres and is recovered by
// the next signal or a restart's LoadAll.
type BaselineSignalConsumer struct {
	consumer *broker.Consumer
	cache    *BaselineCache
}

func NewBaselineSignalConsumer(consumer *broker.Consumer, cache *BaselineCache) *BaselineSignalConsumer {
	return &BaselineSignalConsumer{
		consumer: consumer,
		cache:    cache,
	}
}

// Run consumes signals until ctx is canceled.
func (c *BaselineSignalConsumer) Run(ctx context.Context) error {
	for {
		if err := c.pollAndProcess(ctx); err != nil {
			if errors.Is(err, context.Canceled) {
				return nil
			}
			return err
		}
	}
}

func (c *BaselineSignalConsumer) pollAndProcess(ctx context.Context) error {
	pollCtx, cancel := context.WithTimeout(ctx, signalPollTimeout)
	defer cancel()

	fetches := c.consumer.Client.PollFetches(pollCtx)

	if err := ctx.Err(); err != nil {
		return err
	}

	if fetchErrs := fetches.Errors(); len(fetchErrs) > 0 {
		broker.LogFetchErrors(fetchErrs)
	}

	records := fetches.Records()
	for _, record := range records {
		if err := ctx.Err(); err != nil {
			return err
		}
		c.process(ctx, record)
	}

	if len(records) > 0 {
		c.consumer.Client.MarkCommitRecords(records...)
	}
	return nil
}

// process refreshes the cache for one signal. The caller marks every signal
// consumed regardless of outcome, so decode failures, invalid fields, and
// refresh failures are logged and skipped rather than retried.
func (c *BaselineSignalConsumer) process(ctx context.Context, record *kgo.Record) {
	var signal baselineSignal
	if err := json.Unmarshal(record.Value, &signal); err != nil {
		slog.Warn("skipped malformed baseline signal",
			slog.Int64("offset", record.Offset),
			slog.Any("error", err),
		)
		return
	}

	if signal.SourceID == "" || signal.BaselineRunID <= 0 {
		slog.Warn("skipped invalid baseline signal",
			slog.Int64("offset", record.Offset),
			slog.String("source_id", signal.SourceID),
			slog.Int64("signaled_run_id", signal.BaselineRunID),
		)
		return
	}

	slog.Info("baseline signal received",
		slog.String("source_id", signal.SourceID),
		slog.Int64("signaled_run_id", signal.BaselineRunID),
	)

	if err := c.cache.Refresh(ctx, signal.SourceID); err != nil {
		slog.Error("baseline refresh failed",
			slog.String("source_id", signal.SourceID),
			slog.Int64("signaled_run_id", signal.BaselineRunID),
			slog.Any("error", err),
		)
	}
}