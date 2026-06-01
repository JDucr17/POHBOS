package sink

import (
	"context"

	"github.com/twmb/franz-go/pkg/kgo"

	"github.com/JDucr17/streamline/services/pipeline/internal/dlq"
)

// Projection abstracts the topic specific concerns of a sink binary:
// Decoding records from a source topic into a typed value, and writing
// batches of those values to a downstream read model.
type Projection[T any] interface {
	// Name returns a short label used in logs and DLQ reasons.
	Name() string

	// Decode converts one Kafka record into the value this projection writes.
	Decode(record *kgo.Record) (T, error)

	// Write persists a decoded batch.
	Write(ctx context.Context, items []T) (WriteResult, error)
}

// Summarizes a batch write.
type WriteResult struct {
	Inserted int64
	Skipped  int64
}

// Groups the original Kafka record next to its decoded value.
type DecodedRecord[T any] struct {
	Record *kgo.Record
	Value  T
}

// Batch separates records that decoded cleanly from records that should go to DLQ.
type Batch[T any] struct {
	Good []DecodedRecord[T]
	Bad  []dlq.FailedRecord
}