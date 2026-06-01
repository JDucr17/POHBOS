package sink

import (
	"context"
	"time"

	"github.com/JDucr17/streamline/services/pipeline/internal/domain"
	"github.com/JDucr17/streamline/services/pipeline/internal/postgres"
)

// Query that composes decisions batch insertion
const insertDecisionsFromSinkSQL = `
INSERT INTO decisions (
    id,
    source_id,
    visitor_id,
    event_id,
    score_raw,
    risk_level,
    action,
    status,
    policy_version,
    baseline_run_id,
    decided_at
)
SELECT *
FROM unnest(
    $1::uuid[],
    $2::text[],
    $3::text[],
    $4::uuid[],
    $5::double precision[],
    $6::text[],
    $7::text[],
    $8::text[],
    $9::text[],
    $10::bigint[],
    $11::timestamptz[]
)
ON CONFLICT (event_id) DO NOTHING
`

type DecisionWriter struct {
	db *postgres.DB
}

func NewDecisionWriter(db *postgres.DB) *DecisionWriter {
	return &DecisionWriter{db: db}
}

// Pairs a decision envelope's identity with its scored payload.
type DecisionInsert struct {
	ID        string
	EventID   string
	SourceID  string
	VisitorID string
	Decision  domain.Decision
}

// Quantifies the result of a batch insertion.
type InsertDecisionsResult struct {
	Inserted int64
	Skipped  int64
}

func (w *DecisionWriter) InsertDecisions(ctx context.Context, decisions []DecisionInsert) (InsertDecisionsResult, error) {
	if len(decisions) == 0 {
		return InsertDecisionsResult{}, nil
	}

	cols := buildDecisionColumns(decisions)

	tag, err := w.db.Pool.Exec(ctx, insertDecisionsFromSinkSQL,
		cols.ids,
		cols.sources,
		cols.visitors,
		cols.eventIDs,
		cols.scoreRaws,
		cols.riskLevels,
		cols.actions,
		cols.statuses,
		cols.policyVersions,
		cols.baselineRunIDs,
		cols.decidedAt,
	)
	if err != nil {
		return InsertDecisionsResult{}, postgres.Classify(err)
	}

	inserted := tag.RowsAffected()

	return InsertDecisionsResult{
		Inserted: inserted,
		Skipped:  int64(len(decisions)) - inserted,
	}, nil
}

// Column arrays passed to UNNEST. All slices must have identical length.
type decisionColumns struct {
	ids            []string
	sources        []string
	visitors       []string
	eventIDs       []string
	scoreRaws      []*float64
	riskLevels     []*string
	actions        []string
	statuses       []string
	policyVersions []string
	baselineRunIDs []*int64
	decidedAt      []time.Time
}

func buildDecisionColumns(decisions []DecisionInsert) decisionColumns {
	cols := decisionColumns{
		ids:            make([]string, len(decisions)),
		sources:        make([]string, len(decisions)),
		visitors:       make([]string, len(decisions)),
		eventIDs:       make([]string, len(decisions)),
		scoreRaws:      make([]*float64, len(decisions)),
		riskLevels:     make([]*string, len(decisions)),
		actions:        make([]string, len(decisions)),
		statuses:       make([]string, len(decisions)),
		policyVersions: make([]string, len(decisions)),
		baselineRunIDs: make([]*int64, len(decisions)),
		decidedAt:      make([]time.Time, len(decisions)),
	}

	for i, item := range decisions {
		cols.ids[i] = item.ID
		cols.sources[i] = item.SourceID
		cols.visitors[i] = item.VisitorID
		cols.eventIDs[i] = item.EventID
		cols.scoreRaws[i] = item.Decision.ScoreRaw
		cols.riskLevels[i] = item.Decision.RiskLevel
		cols.actions[i] = item.Decision.Action
		cols.statuses[i] = item.Decision.Status
		cols.policyVersions[i] = item.Decision.PolicyVersion
		cols.baselineRunIDs[i] = item.Decision.BaselineRunID
		cols.decidedAt[i] = item.Decision.DecidedAt
	}

	return cols
}