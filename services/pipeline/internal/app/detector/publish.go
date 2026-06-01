package detector

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"

	"github.com/JDucr17/streamline/services/pipeline/internal/broker"
	"github.com/JDucr17/streamline/services/pipeline/internal/domain"
	"github.com/JDucr17/streamline/services/pipeline/internal/envelope"
)

// publishDecision wraps the decision in an envelope and publishes it to
// the decisions topic.
func publishDecision(
	ctx context.Context,
	producer *broker.Producer,
	decoded DecodedRecord,
	decision domain.Decision,
) error {
	decisionID, err := uuid.NewV7()
	if err != nil {
		return fmt.Errorf("generate decision id: %w", err)
	}

	env := envelope.DecisionEnvelope{
		DecisionID:    decisionID.String(),
		SchemaVersion: envelope.DecisionSchemaVersion,
		SourceID:      decoded.Event.SourceID,
		VisitorID:     decoded.Event.VisitorID,
		EventID:       decoded.EventID,
		Payload:       decision,
	}

	body, err := json.Marshal(env)
	if err != nil {
		return fmt.Errorf("marshal decision envelope: %w", err)
	}

	key := partitionKey(decoded.Event.SourceID, decoded.Event.VisitorID)
	return producer.Publish(ctx, key, body)
}

// partitionKey builds the source_id:visitor_id key used for partitioning.
// Same convention as raw_events so downstream consumers see the same
// per-visitor ordering
func partitionKey(sourceID, visitorID string) []byte {
	return []byte(sourceID + ":" + visitorID)
}