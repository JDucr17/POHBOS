-- +goose Up
CREATE TABLE decisions (
    id               uuid             PRIMARY KEY,
    source_id        text             NOT NULL
                                      CHECK (length(source_id) <= 64),
    visitor_id       text             NOT NULL
                                      CHECK (length(visitor_id) <= 128),
    event_id         uuid             NOT NULL,
    score_raw        double precision
                                      CHECK (score_raw IS NULL OR score_raw >= 0),
    score_normalized double precision
                                      CHECK (score_normalized IS NULL OR (score_normalized >= 0 AND score_normalized <= 1)),
    risk_level       text
                                      CHECK (risk_level IS NULL OR risk_level IN ('low', 'medium', 'high', 'critical')),
    action           text             NOT NULL
                                      CHECK (action IN ('allow', 'log', 'challenge', 'block')),
    status           text             NOT NULL
                                      CHECK (status IN ('scored', 'insufficient_history', 'no_baseline')),
    policy_version   text             NOT NULL
                                      REFERENCES policy_versions(version)
                                      ON DELETE RESTRICT,
    baseline_run_id  bigint           REFERENCES baseline_runs(id)
                                      ON DELETE RESTRICT,
    decided_at       timestamptz      NOT NULL DEFAULT NOW(),

    CONSTRAINT decisions_event_id_unique UNIQUE (event_id),

    CHECK (
        (status = 'scored'
            AND score_raw IS NOT NULL AND score_normalized IS NOT NULL
            AND risk_level IS NOT NULL AND baseline_run_id IS NOT NULL) OR
        (status = 'insufficient_history'
            AND score_raw IS NULL AND score_normalized IS NULL
            AND risk_level IS NULL AND baseline_run_id IS NOT NULL) OR
        (status = 'no_baseline'
            AND score_raw IS NULL AND score_normalized IS NULL
            AND risk_level IS NULL AND baseline_run_id IS NULL)
    )
);

CREATE INDEX decisions_source_decided_at_idx
    ON decisions (source_id, decided_at);

CREATE INDEX decisions_visitor_decided_at_idx
    ON decisions (visitor_id, decided_at);

COMMENT ON TABLE decisions IS
    'Append-only log of HBOS scoring outcomes. One row per cadence-gated evaluation.';

COMMENT ON COLUMN decisions.score_raw IS
    'Raw HBOS surprise score, the unbounded sum of per-feature anomaly contributions. Used for debugging and recalibration. NULL when no score was produced.';

COMMENT ON COLUMN decisions.score_normalized IS
    'Calibrated anomaly percentile in [0,1], derived by comparing score_raw to the baseline score distribution. Policy bands use this value. NULL when no score was produced.';

COMMENT ON COLUMN decisions.risk_level IS
    'Risk band derived from score_normalized via the active policy bands. NULL when no score was produced.';

COMMENT ON COLUMN decisions.action IS
    'Recommended action for downstream enforcement: allow, log, challenge, or block. Streamline flags and recommends; enforcement is performed by other layers.';

COMMENT ON COLUMN decisions.status IS
    'Scoring outcome: scored (a baseline produced a score), insufficient_history (a compatible baseline existed but the visitor had too few events to score), or no_baseline (no baseline existed for the source).';

COMMENT ON COLUMN decisions.policy_version IS
    'Policy version active at scoring time.';

COMMENT ON COLUMN decisions.baseline_run_id IS
    'Compatible active baseline available at evaluation time. Set for scored and insufficient_history; NULL only for no_baseline.';

-- +goose Down
DROP TABLE IF EXISTS decisions;