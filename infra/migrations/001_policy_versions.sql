-- +goose Up
CREATE TABLE policy_versions (
    version       text        PRIMARY KEY
                              CHECK (length(version) <= 64),
    policy        jsonb       NOT NULL
                              CHECK (jsonb_typeof(policy) = 'object'),
    active        boolean     NOT NULL DEFAULT false,
    created_at    timestamptz NOT NULL DEFAULT NOW(),
    activated_at  timestamptz
);

CREATE UNIQUE INDEX policy_versions_one_active
    ON policy_versions (active)
    WHERE active = true;

COMMENT ON TABLE policy_versions IS
    'Versioned HBOS score policy. At most one row active at a time.';

COMMENT ON COLUMN policy_versions.policy IS
    'Versioned score policy. Contains ordered bands over score_normalized in [0,1] and fallback actions for non-scored statuses. Bands use upper-exclusive semantics: a score belongs to the first band where score < max_score. The first band starts at 0; the final band must have max_score null as the open upper bound. Loading validates ascending max_score, non-empty bands, valid risk/action values, exactly one null final band, and that fallback_actions covers every non-scored status.';

COMMENT ON COLUMN policy_versions.activated_at IS
    'Timestamp of most recent activation.';

-- +goose Down
DROP TABLE IF EXISTS policy_versions;