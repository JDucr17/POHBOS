-- +goose Up
CREATE TABLE baseline_runs (
    id            bigint      GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    source_id     text        NOT NULL
                              CHECK (length(source_id) <= 64),
    event_count   int         NOT NULL
                              CHECK (event_count >= 0),
    features_fit  text[]      NOT NULL,
    status        text        NOT NULL
                              CHECK (status IN ('succeeded', 'insufficient_history', 'failed')),
    fit_at        timestamptz NOT NULL DEFAULT NOW()
);

CREATE INDEX baseline_runs_source_fit_at_idx
    ON baseline_runs (source_id, fit_at DESC);

COMMENT ON TABLE baseline_runs IS
    'Audit log of baseline fit operations. One row per attempted fit per source.';

COMMENT ON COLUMN baseline_runs.event_count IS
    'Number of events that fed this fit operation.';

COMMENT ON COLUMN baseline_runs.features_fit IS
    'Names of features for which histograms were produced.';

COMMENT ON COLUMN baseline_runs.status IS
    'succeeded: baseline written to Redis. insufficient_history: too few events. failed: error during fitting.';

-- +goose Down
DROP TABLE IF EXISTS baseline_runs;