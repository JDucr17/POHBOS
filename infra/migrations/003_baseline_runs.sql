-- +goose Up
CREATE TABLE baseline_runs (
    id            bigint      GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    source_id     text        NOT NULL
                              CHECK (length(source_id) <= 64),
    event_count   int         NOT NULL
                              CHECK (event_count >= 0),
    window_count  int         NOT NULL
                              CHECK (window_count >= 0),
    distinct_visitors int     NOT NULL
                              CHECK (distinct_visitors >= 0),
    registry_hash text        NOT NULL
                              CHECK (length(registry_hash) = 64),
    features_fit  text[]      NOT NULL,
    status        text        NOT NULL
                              CHECK (status IN ('succeeded', 'insufficient_history', 'failed')),
    baseline      bytea,
    fit_at        timestamptz NOT NULL DEFAULT NOW(),

    CONSTRAINT baseline_present_iff_succeeded CHECK (
        (status = 'succeeded'     AND baseline IS NOT NULL) OR
        (status <> 'succeeded'    AND baseline IS NULL)
    )
);

CREATE INDEX baseline_runs_source_fit_at_idx
    ON baseline_runs (source_id, fit_at DESC);

COMMENT ON TABLE baseline_runs IS
    'Baseline fit operations and the fitted model artifact. One row per attempted fit per source; the serialized baseline lives in this row, making it the single source of truth for both audit and model.';

COMMENT ON COLUMN baseline_runs.event_count IS
    'Number of events that fed this fit operation.';

COMMENT ON COLUMN baseline_runs.window_count IS
    'Number of sampled trailing windows derived from those events. A baseline is fitted for the source only if this and the distinct-visitor count meet their minimums, otherwise the run is categorized as insufficient_history.';

COMMENT ON COLUMN baseline_runs.distinct_visitors IS
    'Number of distinct visitors that contributed at least one sampled window. Gated alongside window_count so a few heavy visitors cannot define the baseline.';

COMMENT ON COLUMN baseline_runs.registry_hash IS
    'Feature-registry fingerprint the run was evaluated under, for drift analytics. Set on every run from the worker''s compiled-in registry. The detector verifies compatibility against the hash inside the baseline blob, not this column.';

COMMENT ON COLUMN baseline_runs.features_fit IS
    'Names of features for which histograms were produced.';

COMMENT ON COLUMN baseline_runs.baseline IS
    'Serialized Baseline JSON. Present only for succeeded runs; the detector reads this as the authoritative model.';

COMMENT ON COLUMN baseline_runs.status IS
    'succeeded: baseline fitted and stored in this row. insufficient_history: too few sampled windows or too few distinct visitors. failed: error during fitting.';

-- +goose Down
DROP TABLE IF EXISTS baseline_runs;