package sink

import (
	"time"

	"github.com/twmb/franz-go/pkg/kgo"
)


const (
	SinkBatchSize = 200 // maximum amount of records held in one batch before flushing
	SinkBatchAge  = 500 * time.Millisecond // how long to wait before flushing
)

// RecordBatcher buffers Kafka records until the size or age threshold is met.
// A fresh deadline starts when the first record of a new batch arrives
type RecordBatcher struct {
	records  []*kgo.Record
	deadline time.Time
}

func NewRecordBatcher() *RecordBatcher {
	return &RecordBatcher{
		records: make([]*kgo.Record, 0, SinkBatchSize),
	}
}

// Add appends a record, starting the age deadline on the first record
func (b *RecordBatcher) Add(r *kgo.Record) {
	if len(b.records) == 0 {
		b.deadline = time.Now().Add(SinkBatchAge)
	}
	b.records = append(b.records, r)
}

// Deadline returns the current batch's age deadline. return value
// is false if the batch is empty
func (b *RecordBatcher) Deadline() (time.Time, bool) {
    if len(b.records) == 0 {
        return time.Time{}, false
    }
    return b.deadline, true
}

// Ready reports whether the batch should be flushed: full or past its deadline
func (b *RecordBatcher) Ready() bool {
	return len(b.records) >= SinkBatchSize ||
		(len(b.records) > 0 && time.Now().After(b.deadline))
}

// Flush returns the buffered records and resets the batcher
func (b *RecordBatcher) Flush() []*kgo.Record {
	out := b.records
	b.records = make([]*kgo.Record, 0, SinkBatchSize)
	b.deadline = time.Time{}
	return out
}