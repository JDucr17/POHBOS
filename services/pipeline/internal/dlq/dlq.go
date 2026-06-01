package dlq

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/twmb/franz-go/pkg/kgo"

	"github.com/JDucr17/streamline/services/pipeline/internal/broker"
)

// DLQ for unprocessable entities by the pipeline. Decode failures and
// permanent downstream errors route here because retrying won't change
// the outcome, the record itself is wrong. Transient failures leave
// offsets uncommitted instead, so the records re-deliver on the next
// poll. 

// format published to the dead-letter topic. Original
// topic, partition, and offset preserve the source pointer so an operator
// can correlate the failure back to the upstream broker state
type Record struct {
	OriginalTopic     string `json:"original_topic"`
	OriginalPartition int32  `json:"original_partition"`
	OriginalOffset    int64  `json:"original_offset"`
	Reason            string `json:"reason"`
	At                int64  `json:"at"`
	Payload           []byte `json:"payload"`
}

// FailedRecord pairs a Kafka record with the reason it failed. Used as
// the in-memory shape passed through processing code before Route
// transforms it into the wire format
type FailedRecord struct {
	Record *kgo.Record
	Reason string
}

// Route publishes failed records to the dead-letter topic. Returns on the
// first publish error, the caller must not mark the original records as
// committed when this returns non-nil, since the broker is the only
// durable handle on the failed records at that point
func Route(ctx context.Context, producer *broker.Producer, failed []FailedRecord) error {
	now := time.Now().UnixMilli()

	for _, f := range failed {
		body, err := json.Marshal(newRecord(f, now))
		if err != nil {
			return fmt.Errorf("dlq marshal: %w", err)
		}

		if err := producer.Publish(ctx, f.Record.Key, body); err != nil {
			return fmt.Errorf("dlq publish: %w", err)
		}
	}

	return nil
}

// timestamp is passed in so a batch of failures shares one time value
func newRecord(f FailedRecord, at int64) Record {
	return Record{
		OriginalTopic:     f.Record.Topic,
		OriginalPartition: f.Record.Partition,
		OriginalOffset:    f.Record.Offset,
		Reason:            f.Reason,
		At:                at,
		Payload:           f.Record.Value,
	}
}