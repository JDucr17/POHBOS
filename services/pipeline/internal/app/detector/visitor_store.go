package detector

import (
	"log/slog"
	"time"

	"github.com/JDucr17/streamline/services/pipeline/internal/domain"
)

// windowEvent wraps a raw event with its Kafka event_id so the ring can dedup
// redeliveries: a record can redeliver after a publish failure, and the ring
// must not double-count the triggering event.
type windowEvent struct {
	eventID string
	event   domain.RawEvent
}

// visitorState is one visitor's trailing window plus the three clocks. It is
// internal to VisitorStore and never escapes: callers receive only ApplyResult.
type visitorState struct {
	events          []windowEvent
	lastEvaluatedAt time.Time // event-time of the last publsihed eval
	maxEventTime    time.Time // greatest event-time seen, for late event detection
	lastObservedAt  time.Time // processing-time of the last event, for idle eviction
}

// VisitorStore holds per-visitor window state for one detector instance.
//
// It is single-owner and deliberately unsynchronized: the detector's one
// sequential consume loop is the sole caller of Apply, MarkEvaluated, and Evict
// (eviction runs inline in that loop after each batch, not on its own
// goroutine), so the visitor map is never touched concurrently. The store hands
// out only ApplyResult snapshots, never a *visitorState or internal slice.
//
// Scaling is by process, not in-process concurrency: run more detector instances
// in the same consumer group and Kafka assigns partitions across them.
// Partition-affine in-process workers are a future improvement option 
type VisitorStore struct {
	states  map[string]*visitorState
	length  time.Duration // trailing window span, for event-time pruning
	idleTTL time.Duration // processing-time idle timeout, for eviction
}

func NewVisitorStore(length, idleTTL time.Duration) *VisitorStore {
	return &VisitorStore{
		states:  make(map[string]*visitorState),
		length:  length,
		idleTTL: idleTTL,
	}
}

// ApplyResult is the immutable snapshot Apply returns. Events is unwrapped and
// valid only for synchronous use within the triggering evaluate call.
type ApplyResult struct {
	Key             string
	EventTime       time.Time
	Late            bool
	LastEvaluatedAt time.Time
	Events          []domain.RawEvent
}

// Apply folds one delivery into the visitor's window and returns an immutable
// snapshot. It is idempotent by eventID, meaning a redelivery does not double-append and
// never advances the cadence gate, only MarkEvaluated does.
func (s *VisitorStore) Apply(key, eventID string, event domain.RawEvent, now time.Time) ApplyResult {
	state := s.states[key]
	if state == nil {
		state = &visitorState{}
		s.states[key] = state
	}

	if event.EventTime.Before(state.maxEventTime) {
		logLateEvent(eventID, event, state.maxEventTime)
		return ApplyResult{Key: key, EventTime: event.EventTime, Late: true}
	}

	if !containsEventID(state.events, eventID) {
		state.appendAndPrune(windowEvent{eventID: eventID, event: event}, s.length)
		state.maxEventTime = event.EventTime
		state.lastObservedAt = now
	}

	return ApplyResult{
		Key:             key,
		EventTime:       event.EventTime,
		Late:            false,
		LastEvaluatedAt: state.lastEvaluatedAt,
		Events:          state.unwrapEvents(),
	}
}

// MarkEvaluated advances the cadence gate for a visitor. Called from
// processRecord only after the decision was durably published, so a publish
// failure leaves the gate unadvanced and the redelivered record re-evaluates.
func (s *VisitorStore) MarkEvaluated(key string, at time.Time) {
	if state := s.states[key]; state != nil {
		state.lastEvaluatedAt = at
	}
}

// Evict drops visitor state idle past idleTTL. The consume loop calls
// it once per poll cycle after the fetched batch is processed.
func (s *VisitorStore) Evict(now time.Time) {
	for key, state := range s.states {
		if now.Sub(state.lastObservedAt) > s.idleTTL {
			delete(s.states, key)
		}
	}
}

// appendAndPrune adds the event and trims everything older than the trailing
// window, both keyed on event-time. Events arrive in non-decreasing event-time
// so trimming the leading run suffices.
func (st *visitorState) appendAndPrune(we windowEvent, length time.Duration) {
	st.events = append(st.events, we)

	lower := we.event.EventTime.Add(-length)
	drop := 0
	for drop < len(st.events) && st.events[drop].event.EventTime.Before(lower) {
		drop++
	}
	st.events = st.events[drop:]
}

// unwrapEvents copies the ring into a fresh []domain.RawEvent. The new backing
// array means a caller mutating the result cannot alias a later Apply's window.
func (st *visitorState) unwrapEvents() []domain.RawEvent {
	events := make([]domain.RawEvent, len(st.events))
	for i, we := range st.events {
		events[i] = we.event
	}
	return events
}

// containsEventID scans the small ring for a redelivered event.
func containsEventID(events []windowEvent, eventID string) bool {
	for _, we := range events {
		if we.eventID == eventID {
			return true
		}
	}
	return false
}

// logLateEvent records an out of order event the scoring ring refuses.
func logLateEvent(eventID string, event domain.RawEvent, maxEventTime time.Time) {
	slog.Warn("dropping late event",
		slog.String("source_id", event.SourceID),
		slog.String("visitor_id", event.VisitorID),
		slog.String("event_id", eventID),
		slog.Time("event_time", event.EventTime),
		slog.Time("max_event_time", maxEventTime),
	)
}
