-- +goose Up
CREATE TABLE decisions (
    id               uuid             PRIMARY KEY,
    source_id        text             NOT NULL
                                      CHECK (length(source_id) <= 64),
    visitor_id       text             NOT NULL
                                      CHECK (length(visitor_id) <= 128),
    event_id         uuid             NOT NULL
                                      REFERENCES events(id)
                                      ON DELETE RESTRICT,
    score_raw        double precision
                                      CHECK (score_raw IS NULL OR score_raw >= 0),
    risk_level       text
                                      CHECK (risk_level IS NULL OR risk_level IN ('low', 'medium', 'high', 'critical')),
    action           text             NOT NULL
                                      CHECK (action IN ('allow', 'log', 'challenge', 'block')),
    status           text             NOT NULL
                                      CHECK (status IN ('scored', 'insufficient_history')),
    policy_version   text             NOT NULL
                                      REFERENCES policy_versions(version)
                                      ON DELETE RESTRICT,
    baseline_run_id  bigint           REFERENCES baseline_runs(id)
                                      ON DELETE RESTRICT,
    decided_at       timestamptz      NOT NULL DEFAULT NOW(),

    CHECK (
        (status = 'scored'               AND score_raw IS NOT NULL AND risk_level IS NOT NULL AND baseline_run_id IS NOT NULL) OR
        (status = 'insufficient_history' AND score_raw IS NULL     AND risk_level IS NULL     AND baseline_run_id IS NULL)
    )
);

CREATE INDEX decisions_source_decided_at_idx
    ON decisions (source_id, decided_at);

CREATE INDEX decisions_visitor_decided_at_idx
    ON decisions (visitor_id, decided_at);

COMMENT ON TABLE decisions IS
    'Append-only log of HBOS scoring outcomes. One row per scoring event.';

COMMENT ON COLUMN decisions.score_raw IS
    'Raw HBOS score. NULL when status is insufficient_history.';

COMMENT ON COLUMN decisions.risk_level IS
    'Risk band derived from score_raw via the active policy thresholds.';

COMMENT ON COLUMN decisions.action IS
    'Action taken: allow, log, challenge, or block.';

COMMENT ON COLUMN decisions.policy_version IS
    'Policy version active at scoring time.';

COMMENT ON COLUMN decisions.baseline_run_id IS
    'Baseline run whose fit produced this score. NULL when status is insufficient_history.';

-- +goose Down
DROP TABLE IF EXISTS decisions;