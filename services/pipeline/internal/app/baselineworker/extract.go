package baselineworker

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/JDucr17/streamline/services/pipeline/internal/domain"
	"github.com/JDucr17/streamline/services/pipeline/internal/extractor"
	"github.com/JDucr17/streamline/services/pipeline/internal/postgres"
)

// SourceExtract is one source's sampled-window matrix plus fit-gate counts(event count and distinct visitors).
// Matrix columns follow registry feature order.
type SourceExtract struct {
	Matrix           [][]float64
	EventCount       int
	DistinctVisitors int
}

// Querier is the DB read surface Extract needs.
// *pgxpool.Pool and pgx.Tx both satisfy it.
type Querier interface {
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
}

// Source histories are anchored to each source's max(event_time). The query
// reads one window-length before the fit horizon so edge windows have full
// context, pre-horizon events are buffered but never emitted as samples.
const selectEventsSQL = `
	WITH source_anchors AS (
		SELECT
			source_id,
			max(event_time) - ($1::int * interval '1 day') AS horizon_start
		FROM events
		GROUP BY source_id
	)
	SELECT e.source_id, e.visitor_id, e.event_time, e.payload, a.horizon_start
	FROM events e
	JOIN source_anchors a USING (source_id)
	WHERE e.event_time >= a.horizon_start - ($2::int * interval '1 second')
	ORDER BY e.source_id, e.visitor_id, e.event_time, e.id
`

// returns per-source matrices of eligible sampled trailing windows.
func Extract(ctx context.Context, db Querier, windowDays int, spec extractor.WindowSpec) (map[string]*SourceExtract, error) {
	rows, err := db.Query(ctx, selectEventsSQL, windowDays, int(spec.Length.Seconds()))
	if err != nil {
		return nil, postgres.Classify(err)
	}
	defer rows.Close()

	replay := newWindowReplay(spec)
	for rows.Next() {
		var sourceID, visitorID string
		var eventTime, horizonStart time.Time
		var payload []byte
		if err := rows.Scan(&sourceID, &visitorID, &eventTime, &payload, &horizonStart); err != nil {
			return nil, postgres.Classify(err)
		}

		event, err := decodeEvent(sourceID, visitorID, eventTime, payload)
		if err != nil {
			return nil, fmt.Errorf("decode event: %w", err)
		}
		replay.add(sourceID, visitorID, event, horizonStart)
	}
	if err := rows.Err(); err != nil {
		return nil, postgres.Classify(err)
	}

	return replay.extracts, nil
}

// rebuilds RawEvent from payload, but trusts normalized columns for
// source_id, visitor_id, and event_time.
func decodeEvent(sourceID, visitorID string, eventTime time.Time, payload []byte) (domain.RawEvent, error) {
	var event domain.RawEvent
	if err := json.Unmarshal(payload, &event); err != nil {
		return domain.RawEvent{}, err
	}
	event.SourceID = sourceID
	event.VisitorID = visitorID
	event.EventTime = eventTime
	return event, nil
}

// converts ordered events into sampled trailing-window matrices.
type windowReplay struct {
	spec     extractor.WindowSpec
	extracts map[string]*SourceExtract

	started        bool
	keySource      string
	keyVisitor     string
	buffer         []domain.RawEvent
	lastEmitted    time.Time
	visitorEmitted bool
}

func newWindowReplay(spec extractor.WindowSpec) *windowReplay {
	return &windowReplay{
		spec:     spec,
		extracts: make(map[string]*SourceExtract),
	}
}

// add feeds one event in stream order. Key changes reset the per-visitor buffer
// and recency gate. Pre-horizon events are context only.
func (r *windowReplay) add(sourceID, visitorID string, e domain.RawEvent, horizonStart time.Time) {
	ex := r.source(sourceID)

	if !r.started || sourceID != r.keySource || visitorID != r.keyVisitor {
		r.startVisitor(sourceID, visitorID)
	}

	// Bound the buffer, EventsInWindow still applies exact [end-Length, end]
	// membership before feature computation.
	lower := e.EventTime.Add(-r.spec.Length)
	drop := 0
	for drop < len(r.buffer) && r.buffer[drop].EventTime.Before(lower) {
		drop++
	}
	r.buffer = append(r.buffer[drop:], e)

	if e.EventTime.Before(horizonStart) {
		return
	}
	ex.EventCount++

	if !r.gateOpen(e) {
		return
	}

	window := extractor.EventsInWindow(r.buffer, e.EventTime, r.spec)
	if !extractor.Eligible(window, r.spec) {
		return
	}

	ex.Matrix = append(ex.Matrix, extractor.Compute(window))
	r.lastEmitted = e.EventTime
	if !r.visitorEmitted {
		ex.DistinctVisitors++
		r.visitorEmitted = true
	}
}

// gateOpen allows the first sample, then one sample per Cadence.
func (r *windowReplay) gateOpen(e domain.RawEvent) bool {
	return r.lastEmitted.IsZero() || e.EventTime.Sub(r.lastEmitted) >= r.spec.Cadence
}

func (r *windowReplay) startVisitor(sourceID, visitorID string) {
	r.started = true
	r.keySource, r.keyVisitor = sourceID, visitorID
	r.buffer = nil
	r.lastEmitted = time.Time{}
	r.visitorEmitted = false
}

// source returns the running extract for sourceID, creating it on first sight.
func (r *windowReplay) source(sourceID string) *SourceExtract {
	ex := r.extracts[sourceID]
	if ex == nil {
		ex = &SourceExtract{}
		r.extracts[sourceID] = ex
	}
	return ex
}