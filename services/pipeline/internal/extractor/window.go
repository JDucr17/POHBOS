package extractor

import (
	"time"

	"github.com/JDucr17/streamline/services/pipeline/internal/domain"
)

// WindowSpec is the shared rule defining a per-visitor trailing activity window.
// The fitter samples historical windows under this spec and the detector scores
// live windows under the same spec, so both must construct windows identically.
type WindowSpec struct {
	Length    time.Duration // trailing window span
	Cadence   time.Duration // recency gate: at most one window per Cadence per visitor
	MinEvents int           // eligibility floor, applied on both fit and score
}

// EventsInWindow returns the events in [end-Length, end], inclusive on both
// bounds. It assumes events is sorted ascending by EventTime.
func EventsInWindow(events []domain.RawEvent, end time.Time, spec WindowSpec) []domain.RawEvent {
	lower := end.Add(-spec.Length)

	start := 0
	for start < len(events) && events[start].EventTime.Before(lower) {
		start++
	}

	stop := start
	for stop < len(events) && !events[stop].EventTime.After(end) {
		stop++
	}

	return events[start:stop]
}

// Eligible reports whether a window has enough events to fit or score. Fewer
// than MinEvents leaves features degenerate (gaps undefined, counts trivial).
func Eligible(windowEvents []domain.RawEvent, spec WindowSpec) bool {
	return len(windowEvents) >= spec.MinEvents
}
