package detector

import (
	"context"
	"log/slog"
	"time"

	"github.com/twmb/franz-go/pkg/kgo"

	"github.com/JDucr17/streamline/services/pipeline/internal/broker"
	"github.com/JDucr17/streamline/services/pipeline/internal/domain"
	"github.com/JDucr17/streamline/services/pipeline/internal/dlq"
	"github.com/JDucr17/streamline/services/pipeline/internal/extractor"
)

// Detector consumes raw_events, scores each event against its per-source
// baseline, and publishes a DecisionEnvelope to the decisions topic. Decode
// failures route to DLQ.
//
// On shutdown, records mid-flight may be left uncommitted and redeliver
// on next start. Downstream consumers must be idempotent by event_id.
type Detector struct {
	consumer  *broker.Consumer
	decisions *broker.Producer
	dlq       *broker.Producer

	visitors    *VisitorStore
	cache       *BaselineCache
	policy      Policy
	servingSpec extractor.WindowSpec

	// publish and commit are the record's two side effects, set by NewDetector to
	// the real Kafka operations and overridden in tests to drive the consume loop
	// without a live broker.
	publish func(ctx context.Context, decoded DecodedRecord, decision domain.Decision) error
	commit  func(r *kgo.Record)
}

func NewDetector(
	consumer *broker.Consumer,
	decisions, dlq *broker.Producer,
	visitors *VisitorStore,
	cache *BaselineCache,
	policy Policy,
	servingSpec extractor.WindowSpec,
) *Detector {
	d := &Detector{
		consumer:    consumer,
		decisions:   decisions,
		dlq:         dlq,
		visitors:    visitors,
		cache:       cache,
		policy:      policy,
		servingSpec: servingSpec,
	}
	d.publish = func(ctx context.Context, decoded DecodedRecord, decision domain.Decision) error {
		return publishDecision(ctx, d.decisions, decoded, decision)
	}
	d.commit = func(r *kgo.Record) {
		d.consumer.Client.MarkCommitRecords(r)
	}
	return d
}

// Run polls Kafka continuously and processes each fetched batch. It returns when
// ctx is cancelled, or with a fatal error when a batch hits a detector invariant
// violation. Uncommitted records redeliver on next start.
func (d *Detector) Run(ctx context.Context) error {
	for {
		fetches := d.consumer.Client.PollFetches(ctx)

		if err := ctx.Err(); err != nil {
			return err
		}

		if fetchErrs := fetches.Errors(); len(fetchErrs) > 0 {
			broker.LogFetchErrors(fetchErrs)
		}

		if err := d.processBatch(ctx, fetches.Records()); err != nil {
			return err
		}
	}
}

// processBatch processes the fetched records in order, then sweeps idle visitor
// state once. A publish failure stops the batch immediately: Apply has already
// advanced the visitor's max_event_time, so continuing to later records and then
// redelivering the failed one could see it falsely dropped as late. The failed
// record and everything after it stay uncommitted and redeliver. A returned
// error is a fatal invariant violation that halts the detector.
//
// Single-owner invariant: this is the only goroutine that touches VisitorStore,
// so eviction runs inline here rather than on a separate goroutine. A batch that
// stops early skips this cycle's eviction, which is harmless.
func (d *Detector) processBatch(ctx context.Context, records []*kgo.Record) error {
	for _, r := range records {
		stop, err := d.processRecord(ctx, r)
		if err != nil {
			return err
		}
		if stop {
			return nil
		}
	}

	d.visitors.Evict(time.Now())
	return nil
}

// processRecord runs one record through decode -> evaluate -> publish. It returns
// stop=true on a transient publish failure (the caller halts the batch, leaving
// the record uncommitted for redelivery) and a non-nil error on a fatal invariant
// violation (a score/classify failure on validated data). Decode failures route
// to the DLQ and commit.
func (d *Detector) processRecord(ctx context.Context, r *kgo.Record) (stop bool, err error) {
	decoded, err := decode(r)
	if err != nil {
		d.routeDecodeFailure(ctx, r, err)
		return false, nil
	}

	eval, err := d.evaluate(decoded)
	if err != nil {
		return false, err
	}

	if !eval.Emit {
		d.commit(r)
		return false, nil
	}

	if err := d.publish(ctx, decoded, eval.Decision); err != nil {
		slog.Error("decision publish failed, stopping batch uncommitted",
			slog.String("event_id", decoded.EventID),
			slog.Any("error", err),
		)
		return true, nil
	}

	// Decision is durably published, only now advance the gate, then commit.
	d.visitors.MarkEvaluated(eval.Key, eval.EventTime)
	d.commit(r)

	slog.Debug("decision published",
		slog.String("event_id", decoded.EventID),
		slog.String("status", eval.Decision.Status),
	)
	return false, nil
}

// routeDecodeFailure publishes a decode-failed record to the DLQ. If the
// DLQ publish itself fails, the original record stays uncommitted, the
// broker is the only durable handle on it
func (d *Detector) routeDecodeFailure(ctx context.Context, r *kgo.Record, cause error) {
	failed := []dlq.FailedRecord{{Record: r, Reason: cause.Error()}}

	if err := dlq.Route(ctx, d.dlq, failed); err != nil {
		slog.Error("dlq publish failed, leaving record uncommitted",
			slog.String("topic", r.Topic),
			slog.Int("partition", int(r.Partition)),
			slog.Int64("offset", r.Offset),
			slog.Any("error", err),
		)
		return
	}

	d.commit(r)
}
