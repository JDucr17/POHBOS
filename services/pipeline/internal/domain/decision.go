package domain

import "time"

type Decision struct {
	DecidedAt       time.Time `json:"decided_at"`
	ScoreRaw        *float64  `json:"score_raw,omitempty"`
	ScoreNormalized *float64  `json:"score_normalized,omitempty"`
	RiskLevel       *string   `json:"risk_level,omitempty"`
	Action          string    `json:"action"`
	Status          string    `json:"status"`
	PolicyVersion   string    `json:"policy_version"`
	BaselineRunID   *int64    `json:"baseline_run_id,omitempty"`
}