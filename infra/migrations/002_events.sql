-- +goose Up
CREATE TABLE events (
    id           uuid        PRIMARY KEY,
    source_id    text        NOT NULL
                             CHECK (length(source_id) <= 64),
    visitor_id   text        NOT NULL
                             CHECK (length(visitor_id) <= 128),
    event_time   timestamptz NOT NULL,
    ingested_at  timestamptz NOT NULL DEFAULT NOW(),
    payload      jsonb       NOT NULL
                             CHECK (jsonb_typeof(payload) = 'object'),
    features     jsonb       NOT NULL
                             CHECK (jsonb_typeof(features) = 'object')
);

CREATE INDEX events_source_event_time_idx
    ON events (source_id, event_time);

CREATE INDEX events_visitor_event_time_idx
    ON events (visitor_id, event_time);

COMMENT ON TABLE events IS
    'Append-only log of HTTP events ingested into the pipeline.';

COMMENT ON COLUMN events.event_time IS
    'Timestamp of the event at the source.';

COMMENT ON COLUMN events.ingested_at IS
    'Timestamp of when Postgres received the row.';

COMMENT ON COLUMN events.payload IS
    'Raw HTTP record from the source.';

COMMENT ON COLUMN events.features IS
    'Feature values computed for this event, keyed by feature name.';

-- +goose Down
DROP TABLE IF EXISTS events;