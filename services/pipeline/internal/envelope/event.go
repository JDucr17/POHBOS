package envelope

import "github.com/JDucr17/streamline/services/pipeline/internal/domain"

const SchemaVersion = "v1"

// Envelope published to event topic
type EventEnvelope struct {
	EventID       string          `json:"event_id"`
	IngestedAt    int64           `json:"ingested_at"`
	SchemaVersion string          `json:"schema_version"`
	SourceID      string          `json:"source_id"`
	VisitorID     string          `json:"visitor_id"`
	Payload       domain.RawEvent `json:"payload"`
}