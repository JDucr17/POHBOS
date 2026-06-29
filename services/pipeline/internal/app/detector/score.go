package detector

import (
	"fmt"
	"time"

	"github.com/JDucr17/streamline/services/pipeline/internal/domain"
	"github.com/JDucr17/streamline/services/pipeline/internal/extractor"
)

// Evaluation is the outcome of one cadence-gated evaluation. Emit is false when
// the event was late or the gate was closed, in which case Decision is unset and
// no decision is published. Key and EventTime carry what processRecord needs to
// advance the gate via MarkEvaluated AFTER a successful publish.
type Evaluation struct {
	Key       string
	EventTime time.Time
	Decision  domain.Decision
	Emit      bool
}

// evaluate folds the event into the visitor's window and, if the cadence gate is
// open, builds exactly one decision. It doesnt advance the gate, processRecord
// calls MarkEvaluated only after the decision is durably published, so a publish
// failure re-evaluates cleanly on redelivery. A returned error is a detector
// invariant violation (scoring/classification of validated data should not fail)
// and is fatal to the consume loop.
func (d *Detector) evaluate(decoded DecodedRecord) (Evaluation, error) {
	key := decoded.Event.SourceID + ":" + decoded.Event.VisitorID
	applied := d.visitors.Apply(key, decoded.EventID, decoded.Event, time.Now())

	if applied.Late {
		return Evaluation{Emit: false}, nil
	}
	if gateClosed(applied.LastEvaluatedAt, applied.EventTime, d.servingSpec.Cadence) {
		return Evaluation{Emit: false}, nil
	}

	decision, err := d.decide(decoded, applied)
	if err != nil {
		return Evaluation{}, err
	}
	return Evaluation{Key: key, EventTime: applied.EventTime, Decision: decision, Emit: true}, nil
}

// gateClosed reports whether the cadence gate is still shut: a prior evaluation
// exists and this event falls within Cadence of it. Event-time, mirroring the
// fitter's gateOpen inverted.
func gateClosed(lastEvaluatedAt, eventTime time.Time, cadence time.Duration) bool {
	return !lastEvaluatedAt.IsZero() && eventTime.Sub(lastEvaluatedAt) < cadence
}

// decide resolves the gate-open evaluation to one of the three decision shapes.
func (d *Detector) decide(decoded DecodedRecord, applied ApplyResult) (domain.Decision, error) {
	model := d.cache.Get(decoded.Event.SourceID)
	if model == nil {
		return d.fallbackDecision("no_baseline", nil)
	}

	window := extractor.EventsInWindow(applied.Events, applied.EventTime, d.servingSpec)
	if !extractor.Eligible(window, d.servingSpec) {
		runID := model.RunID()
		return d.fallbackDecision("insufficient_history", &runID)
	}

	raw, normalized, err := model.Score(extractor.Compute(window))
	if err != nil {
		return domain.Decision{}, fmt.Errorf("score source %q: %w", decoded.Event.SourceID, err)
	}
	risk, action, err := d.policy.Classify(normalized)
	if err != nil {
		return domain.Decision{}, fmt.Errorf("classify source %q: %w", decoded.Event.SourceID, err)
	}

	runID := model.RunID()
	return domain.Decision{
		DecidedAt:       time.Now(),
		ScoreRaw:        &raw,
		ScoreNormalized: &normalized,
		RiskLevel:       &risk,
		Action:          action,
		Status:          "scored",
		PolicyVersion:   d.policy.Version,
		BaselineRunID:   &runID,
	}, nil
}

// fallbackDecision builds a non-scored decision: score fields and risk stay null,
// the action comes from the policy fallback, and baselineRunID is set only for
// insufficient_history (the compatible baseline active at evaluation time).
func (d *Detector) fallbackDecision(status string, baselineRunID *int64) (domain.Decision, error) {
	action, err := d.policy.FallbackAction(status)
	if err != nil {
		return domain.Decision{}, err
	}
	return domain.Decision{
		DecidedAt:     time.Now(),
		Action:        action,
		Status:        status,
		PolicyVersion: d.policy.Version,
		BaselineRunID: baselineRunID,
	}, nil
}
