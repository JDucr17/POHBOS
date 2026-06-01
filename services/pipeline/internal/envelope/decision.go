package envelope

import "github.com/JDucr17/streamline/services/pipeline/internal/domain"

const DecisionSchemaVersion = "v1"

type DecisionEnvelope struct {
    DecisionID    string          `json:"decision_id"`
    SchemaVersion string          `json:"schema_version"`
    SourceID      string          `json:"source_id"`
    VisitorID     string          `json:"visitor_id"`
    EventID       string          `json:"event_id"`
    Payload       domain.Decision `json:"payload"`
}