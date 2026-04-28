-- +goose Up
CREATE TABLE policy_versions (
    version       text        PRIMARY KEY
                              CHECK (length(version) <= 64),
    thresholds    jsonb       NOT NULL
                              CHECK (jsonb_typeof(thresholds) = 'object'),
    active        boolean     NOT NULL DEFAULT false,
    created_at    timestamptz NOT NULL DEFAULT NOW(),
    activated_at  timestamptz
);

CREATE UNIQUE INDEX policy_versions_one_active
    ON policy_versions (active)
    WHERE active = true;

COMMENT ON TABLE policy_versions IS
    'Versioned HBOS score thresholds. At most one row active at a time.';

COMMENT ON COLUMN policy_versions.thresholds IS
    'Maps a raw HBOS score to a risk_level and action via per-band max_score upper bounds. max_score null = open upper bound.';

COMMENT ON COLUMN policy_versions.activated_at IS
    'Timestamp of most recent activation.';

-- +goose Down
DROP TABLE IF EXISTS policy_versions;