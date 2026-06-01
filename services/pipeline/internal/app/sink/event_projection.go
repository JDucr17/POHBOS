package sink

import (
	"context"
	"encoding/json"

	"github.com/twmb/franz-go/pkg/kgo"

	"github.com/JDucr17/streamline/services/pipeline/internal/envelope"
)

// EventProjection writes raw_events envelopes into the events table.
type EventProjection struct {
	writer *EventWriter
}

func NewEventProjection(writer *EventWriter) *EventProjection {
	return &EventProjection{writer: writer}
}

func (p *EventProjection) Name() string {
	return "events"
}

func (p *EventProjection) Decode(record *kgo.Record) (EventInsert, error) {
	var env envelope.EventEnvelope
	if err := json.Unmarshal(record.Value, &env); err != nil {
		return EventInsert{}, err
	}

	return EventInsert{
		ID:    env.EventID,
		Event: env.Payload,
	}, nil
}

func (p *EventProjection) Write(ctx context.Context, items []EventInsert) (WriteResult, error) {
	result, err := p.writer.InsertEvents(ctx, items)
	if err != nil {
		return WriteResult{}, err
	}

	return WriteResult{
		Inserted: result.Inserted,
		Skipped:  result.Skipped,
	}, nil
}