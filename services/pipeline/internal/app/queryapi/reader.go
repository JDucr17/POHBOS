package queryapi

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"

	"github.com/JDucr17/streamline/services/pipeline/internal/domain"
	"github.com/JDucr17/streamline/services/pipeline/internal/envelope"
	"github.com/JDucr17/streamline/services/pipeline/internal/postgres"
)

const selectLatestByVisitorSQL = `
SELECT id, event_id, source_id, visitor_id,
       score_raw, risk_level, action, status,
       policy_version, baseline_run_id, decided_at
FROM decisions
WHERE source_id = $1 AND visitor_id = $2
ORDER BY decided_at DESC
LIMIT 1
`

const selectBySourceSQL = `
SELECT id, event_id, source_id, visitor_id,
       score_raw, risk_level, action, status,
       policy_version, baseline_run_id, decided_at
FROM decisions
WHERE source_id = $1
ORDER BY decided_at DESC
LIMIT $2
`

const (
	defaultListLimit = 100
	maxListLimit     = 1000
)

// Reader serves decisions from Postgres for cache misses and queries outside
// the in-memory window.
type Reader struct {
	db *postgres.DB
}

func NewReader(db *postgres.DB) *Reader {
	return &Reader{db: db}
}

// LatestByVisitor returns the most recent decision for a visitor.
func (r *Reader) LatestByVisitor(ctx context.Context, sourceID, visitorID string) (envelope.DecisionEnvelope, bool, error) {
	row := r.db.Pool.QueryRow(ctx, selectLatestByVisitorSQL, sourceID, visitorID)

	env, err := scanDecision(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return envelope.DecisionEnvelope{}, false, nil
	}
	if err != nil {
		return envelope.DecisionEnvelope{}, false, postgres.Classify(err)
	}

	return env, true, nil
}

// ListBySource returns recent decisions for a source, newest first.
func (r *Reader) ListBySource(ctx context.Context, sourceID string, limit int) ([]envelope.DecisionEnvelope, error) {
	limit = normalizeLimit(limit)

	rows, err := r.db.Pool.Query(ctx, selectBySourceSQL, sourceID, limit)
	if err != nil {
		return nil, postgres.Classify(err)
	}
	defer rows.Close()

	out := make([]envelope.DecisionEnvelope, 0, limit)
	for rows.Next() {
		env, err := scanDecision(rows)
		if err != nil {
			return nil, postgres.Classify(err)
		}

		out = append(out, env)
	}

	if err := rows.Err(); err != nil {
		return nil, postgres.Classify(err)
	}

	return out, nil
}

func normalizeLimit(limit int) int {
	if limit <= 0 {
		return defaultListLimit
	}
	if limit > maxListLimit {
		return maxListLimit
	}
	return limit
}

// scanDecision reads one decisions row into a DecisionEnvelope. The envelope
// schema version is stamped, not stored, because it is wire-format metadata.
func scanDecision(row pgx.Row) (envelope.DecisionEnvelope, error) {
	var (
		decision domain.Decision
		env      envelope.DecisionEnvelope
	)

	err := row.Scan(
		&env.DecisionID,
		&env.EventID,
		&env.SourceID,
		&env.VisitorID,
		&decision.ScoreRaw,
		&decision.RiskLevel,
		&decision.Action,
		&decision.Status,
		&decision.PolicyVersion,
		&decision.BaselineRunID,
		&decision.DecidedAt,
	)
	if err != nil {
		return envelope.DecisionEnvelope{}, err
	}

	env.SchemaVersion = envelope.DecisionSchemaVersion
	env.Payload = decision

	return env, nil
}