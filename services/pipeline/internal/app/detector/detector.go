package detector

import (
	"context"
	"log/slog"

	"github.com/twmb/franz-go/pkg/kgo"

	"github.com/JDucr17/streamline/services/pipeline/internal/broker"
	"github.com/JDucr17/streamline/services/pipeline/internal/dlq"
)

// Detector consumes raw_events, scores each event, and publishes a
// DecisionEnvelope to the decisions topic. Decode failures route to DLQ.
//
// On shutdown, records mid-flight may be left uncommitted and redeliver
// on next start. Downstream consumers must be idempotent by event_id.
type Detector struct {
	consumer  *broker.Consumer
	decisions *broker.Producer
	dlq       *broker.Producer
}

func NewDetector(consumer *broker.Consumer, decisions, dlq *broker.Producer) *Detector {
	return &Detector{
		consumer:  consumer,
		decisions: decisions,
		dlq:       dlq,
	}
}

// Run polls Kafka continuously and processes records per-event. Returns
// when ctx is cancelled. Uncommitted records redeliver on next start
func (d *Detector) Run(ctx context.Context) error {
	for {
		fetches := d.consumer.Client.PollFetches(ctx)

		if err := ctx.Err(); err != nil {
			return err
		}

		if fetchErrs := fetches.Errors(); len(fetchErrs) > 0 {
			broker.LogFetchErrors(fetchErrs)
		}

		fetches.EachRecord(func(r *kgo.Record) {
			d.processRecord(ctx, r)
		})
	}
}


// processRecord handles a single Kafka record through the full detector
// pipeline. Marks the offset only after the decision is published
func (d *Detector) processRecord(ctx context.Context, r *kgo.Record) {
	decoded, err := decode(r)
	if err != nil {
		d.routeDecodeFailure(ctx, r, err)
		return
	}

	decision := stubScore(decoded.Event)

	if err := publishDecision(ctx, d.decisions, decoded, decision); err != nil {
		slog.Error("decision publish failed, leaving record uncommitted",
			slog.String("event_id", decoded.EventID),
			slog.Any("error", err),
		)
		return
	}

	slog.Debug("decision published",
		slog.String("event_id", decoded.EventID),
		slog.String("status", decision.Status),
	)

	d.consumer.Client.MarkCommitRecords(r)
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

	d.consumer.Client.MarkCommitRecords(r)
}