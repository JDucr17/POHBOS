package sink

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/JDucr17/streamline/services/pipeline/internal/domain"
	"github.com/JDucr17/streamline/services/pipeline/internal/postgres"
)

// Query that composes events batch insertion
const insertEventsFromSinkSQL = `
INSERT INTO events (id, source_id, visitor_id, event_time, payload, features)
SELECT *
FROM unnest(
	$1::uuid[],
	$2::text[],
	$3::text[],
	$4::timestamptz[],
	$5::jsonb[],
	$6::jsonb[]
)
ON CONFLICT (id) DO NOTHING
`

// Feature JSONB placeholder, detector populates column once scores are produced
const emptyFeaturesJSON = `{}`

type EventWriter struct {
	db *postgres.DB
}

func NewEventWriter(db *postgres.DB) *EventWriter {
	return &EventWriter{db: db}
}

// Pairs an event with its envelope-assigned ID
type EventInsert struct {
	ID    string
	Event domain.RawEvent
}

// Quantifies the result of a batch insertion
type InsertEventsResult struct {
	Inserted int64
	Skipped  int64
}

func (w *EventWriter) InsertEvents(ctx context.Context, events []EventInsert) (InsertEventsResult, error) {
	if len(events) == 0 {
		return InsertEventsResult{}, nil
	}

	cols, err := buildEventColumns(events)
	if err != nil {
		return InsertEventsResult{}, err
	}

	tag, err := w.db.Pool.Exec(ctx, insertEventsFromSinkSQL,
		cols.ids,
		cols.sources,
		cols.visitors,
		cols.times,
		cols.payloads,
		cols.features,
	)
	if err != nil {
		return InsertEventsResult{}, postgres.Classify(err)
	}

	inserted := tag.RowsAffected()

	return InsertEventsResult{
		Inserted: inserted,
		Skipped:  int64(len(events)) - inserted,
	}, nil
}

// Column arrays passed to UNNEST.
type eventColumns struct {
	ids      []string
	sources  []string
	visitors []string
	times    []time.Time
	payloads [][]byte
	features [][]byte
}

func buildEventColumns(events []EventInsert) (eventColumns, error) {
	cols := eventColumns{
		ids:      make([]string, len(events)),
		sources:  make([]string, len(events)),
		visitors: make([]string, len(events)),
		times:    make([]time.Time, len(events)),
		payloads: make([][]byte, len(events)),
		features: make([][]byte, len(events)),
	}

	for i, item := range events {
		payload, err := json.Marshal(item.Event)
		if err != nil {
			return eventColumns{}, fmt.Errorf("marshal event %d: %v", i, err)
		}

		cols.ids[i] = item.ID
		cols.sources[i] = item.Event.SourceID
		cols.visitors[i] = item.Event.VisitorID
		cols.times[i] = item.Event.EventTime
		cols.payloads[i] = payload
		cols.features[i] = []byte(emptyFeaturesJSON)
	}

	return cols, nil
}