package detector

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"

	"github.com/JDucr17/streamline/services/pipeline/internal/postgres"
)

const selectActivePolicySQL = `
SELECT version, policy
FROM policy_versions
WHERE active = true
`

// LoadActivePolicy reads and validates the currently active policy. A missing
// active row is a fatal configuration error, not a runtime status: with no
// policy there is no valid decision to write , so the detector refuses to start.
func LoadActivePolicy(ctx context.Context, db *postgres.DB) (Policy, error) {
	var version string
	var raw []byte
	err := db.Pool.QueryRow(ctx, selectActivePolicySQL).Scan(&version, &raw)
	if errors.Is(err, pgx.ErrNoRows) {
		return Policy{}, fmt.Errorf("no active policy in policy_versions: detector cannot start")
	}
	if err != nil {
		return Policy{}, postgres.Classify(err)
	}

	policy, err := decodePolicy(version, raw)
	if err != nil {
		return Policy{}, err
	}

	if err := ValidatePolicy(policy); err != nil {
		return Policy{}, fmt.Errorf("validate active policy %q: %w", version, err)
	}
	return policy, nil
}

func decodePolicy(version string, raw []byte) (Policy, error) {
	var policy Policy
	if err := json.Unmarshal(raw, &policy); err != nil {
		return Policy{}, fmt.Errorf("unmarshal active policy %q: %w", version, err)
	}
	policy.Version = version
	return policy, nil
}
