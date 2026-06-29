package baselineworker

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/JDucr17/streamline/services/pipeline/internal/broker"
)

// baselineSignal is the payload emmitted when a baseline is fit, consumed by the detector to update its internal baselines used for scoring.
type baselineSignal struct {
	SourceID      string `json:"source_id"`
	BaselineRunID int64  `json:"baseline_run_id"`
}


func publishSignal(ctx context.Context, producer *broker.Producer, sourceID string, runID int64) error {
	payload, err := json.Marshal(baselineSignal{SourceID: sourceID, BaselineRunID: runID})
	if err != nil {
		return fmt.Errorf("marshal baseline signal: %w", err)
	}

	if err := producer.Publish(ctx, []byte(sourceID), payload); err != nil {
		return fmt.Errorf("publish baseline signal: %w", err)
	}
	return nil
}
