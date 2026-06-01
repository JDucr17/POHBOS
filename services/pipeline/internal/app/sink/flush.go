package sink

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/twmb/franz-go/pkg/kgo"

	"github.com/JDucr17/streamline/services/pipeline/internal/dlq"
)

// flush processes a batch of records and marks them committable only if
// every step succeeds. Failures leave records uncommitted for redelivery.
func (s *Sink[T]) flush(ctx context.Context, records []*kgo.Record) {
	if len(records) == 0 {
		return
	}

	batch := s.decodeRecords(records)

	if err := s.routeBadRecords(ctx, batch.Bad); err != nil {
		slog.Error("dlq publish failed, leaving batch uncommitted",
			slog.String("projection", s.projection.Name()),
			slog.Int("bad_records", len(batch.Bad)),
			slog.Any("error", err),
		)
		return
	}

	if err := s.writeGoodRecords(ctx, batch.Good); err != nil {
		slog.Error("batch write failed, leaving uncommitted",
			slog.String("projection", s.projection.Name()),
			slog.Int("records", len(batch.Good)),
			slog.Any("error", err),
		)
		return
	}

	s.consumer.Client.MarkCommitRecords(records...)
}

func (s *Sink[T]) decodeRecords(records []*kgo.Record) Batch[T] {
	batch := Batch[T]{
		Good: make([]DecodedRecord[T], 0, len(records)),
	}

	for _, record := range records {
		value, err := s.projection.Decode(record)
		if err != nil {
			batch.Bad = append(batch.Bad, dlq.FailedRecord{
				Record: record,
				Reason: fmt.Sprintf("decode %s: %v", s.projection.Name(), err),
			})
			continue
		}

		batch.Good = append(batch.Good, DecodedRecord[T]{
			Record: record,
			Value:  value,
		})
	}

	return batch
}

// routeBadRecords sends decode failures to DLQ.
func (s *Sink[T]) routeBadRecords(ctx context.Context, bad []dlq.FailedRecord) error {
	if len(bad) == 0 {
		return nil
	}

	if err := dlq.Route(ctx, s.dlq, bad); err != nil {
		return fmt.Errorf("route bad records to dlq: %w", err)
	}

	return nil
}

// writeGoodRecords delegates to the projection's Write with bounded retry.
// Permanent failures route the batch to DLQ; transient failures bubble up.
func (s *Sink[T]) writeGoodRecords(ctx context.Context, good []DecodedRecord[T]) error {
	if len(good) == 0 {
		return nil
	}

	items := values(good)

	var result WriteResult
	err := retry(ctx, func(ctx context.Context) error {
		var writeErr error
		result, writeErr = s.projection.Write(ctx, items)
		return writeErr
	})

	switch {
	case err == nil:
		s.logWriteResult(len(good), result)
		return nil

	case isPermanent(err):
		if routeErr := dlq.Route(ctx, s.dlq, failedRecords(good, err)); routeErr != nil {
			return fmt.Errorf("route permanently failed records to dlq: %w", routeErr)
		}
		return nil

	default:
		return err
	}
}

// values extracts decoded values from their Kafka records.
func values[T any](records []DecodedRecord[T]) []T {
	items := make([]T, len(records))

	for i, record := range records {
		items[i] = record.Value
	}

	return items
}

// logWriteResult records persistence counts and duplicate skips.
func (s *Sink[T]) logWriteResult(records int, result WriteResult) {
	slog.Debug("batch persisted",
		slog.String("projection", s.projection.Name()),
		slog.Int("records", records),
		slog.Int64("inserted", result.Inserted),
		slog.Int64("skipped", result.Skipped),
	)

	if result.Skipped == 0 {
		return
	}

	slog.Info("duplicate inserts skipped",
		slog.String("projection", s.projection.Name()),
		slog.Int64("inserted", result.Inserted),
		slog.Int64("skipped", result.Skipped),
	)
}

// failedRecords wraps decoded records with a DLQ reason.
func failedRecords[T any](records []DecodedRecord[T], cause error) []dlq.FailedRecord {
	reason := cause.Error()
	failed := make([]dlq.FailedRecord, len(records))

	for i, record := range records {
		failed[i] = dlq.FailedRecord{
			Record: record.Record,
			Reason: reason,
		}
	}

	return failed
}