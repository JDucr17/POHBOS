package sink

import (
	"context"
	"encoding/json"

	"github.com/twmb/franz-go/pkg/kgo"

	"github.com/JDucr17/streamline/services/pipeline/internal/envelope"
)

// DecisionProjection unwraps a DecisionEnvelope and writes a decision record.
type DecisionProjection struct {
	writer *DecisionWriter
}

func NewDecisionProjection(writer *DecisionWriter) *DecisionProjection {
	return &DecisionProjection{writer: writer}
}

func (p *DecisionProjection) Name() string {
	return "decisions"
}

func (p *DecisionProjection) Decode(record *kgo.Record) (DecisionInsert, error) {
	var env envelope.DecisionEnvelope
	if err := json.Unmarshal(record.Value, &env); err != nil {
		return DecisionInsert{}, err
	}

	return DecisionInsert{
		ID:        env.DecisionID,
		EventID:   env.EventID,
		SourceID:  env.SourceID,
		VisitorID: env.VisitorID,
		Decision:  env.Payload,
	}, nil
}

func (p *DecisionProjection) Write(ctx context.Context, items []DecisionInsert) (WriteResult, error) {
	result, err := p.writer.InsertDecisions(ctx, items)
	if err != nil {
		return WriteResult{}, err
	}

	return WriteResult{
		Inserted: result.Inserted,
		Skipped:  result.Skipped,
	}, nil
}