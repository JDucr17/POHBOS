package detector

import (
	"encoding/json"
	"fmt"

	"github.com/twmb/franz-go/pkg/kgo"

	"github.com/JDucr17/streamline/services/pipeline/internal/domain"
	"github.com/JDucr17/streamline/services/pipeline/internal/envelope"
)

// DecodedRecord pairs a Kafka record with its parsed envelope contents.
type DecodedRecord struct {
	Record  *kgo.Record
	EventID string
	Event   domain.RawEvent
}

// decode parses one record's EventEnvelope. Returns an error on malformed
// JSON so the caller can route the record to DLQ
func decode(r *kgo.Record) (DecodedRecord, error) {
	var env envelope.EventEnvelope
	if err := json.Unmarshal(r.Value, &env); err != nil {
		return DecodedRecord{}, fmt.Errorf("decode envelope: %w", err)
	}

	return DecodedRecord{
		Record:  r,
		EventID: env.EventID,
		Event:   env.Payload,
	}, nil
}