package sink

import (
	"context"
	"log/slog"
	"time"

	"github.com/twmb/franz-go/pkg/kgo"

	"github.com/JDucr17/streamline/services/pipeline/internal/broker"
)

// bounds how long the final shutdown flush is allowed
// to run. Uses a fresh context independent of the canceled run context so
// the flush can actually reach Postgres and DLQ
const shutdownFlushTimeout = 10 * time.Second

type Sink[T any] struct {
	consumer   *broker.Consumer
	dlq        *broker.Producer
	projection Projection[T]
}

func NewSink[T any](
	consumer *broker.Consumer,
	dlq *broker.Producer,
	projection Projection[T],
) *Sink[T] {
	return &Sink[T]{
		consumer:   consumer,
		dlq:        dlq,
		projection: projection,
	}
}

// Run polls Kafka continuously and flushes records in bounded batches.
// On ctx cancellation, performs a final flush with a fresh timeout so
// pending records can land downstream before exit.
func (s *Sink[T]) Run(ctx context.Context) error {
	batch := NewRecordBatcher()

	for {
		fetches := s.pollUntilDeadline(ctx, batch)

		if err := ctx.Err(); err != nil {
			s.shutdownFlush(batch.Flush())
			return err
		}

		broker.LogFetchErrors(fetches.Errors())

		fetches.EachRecord(func(r *kgo.Record) {
			batch.Add(r)
		})

		if batch.Ready() {
			s.flush(ctx, batch.Flush())
		}
	}
}

// pollUntilDeadline polls Kafka but bounds the wait by the batch's age
// deadline so the age threshold gets honored even at zero traffic.
func (s *Sink[T]) pollUntilDeadline(ctx context.Context, batch *RecordBatcher) kgo.Fetches {
	deadline, hasDeadline := batch.Deadline()
	if !hasDeadline {
		return s.consumer.Client.PollFetches(ctx)
	}

	pollCtx, cancel := context.WithDeadline(ctx, deadline)
	defer cancel()
	return s.consumer.Client.PollFetches(pollCtx)
}

// shutdownFlush performs the final batch flush during shutdown with a
// fresh context.
func (s *Sink[T]) shutdownFlush(records []*kgo.Record) {
	if len(records) == 0 {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), shutdownFlushTimeout)
	defer cancel()

	slog.Info("sink final flush started",
		slog.String("projection", s.projection.Name()),
		slog.Int("records", len(records)),
	)
	s.flush(ctx, records)
}