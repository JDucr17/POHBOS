from dataclasses import dataclass


@dataclass(frozen=True, slots=True)
class Snapshot:
    logical_name: str
    filename: str
    select_sql: str


EXPECTED_SNAPSHOTS = (
    Snapshot(
        "events",
        "events.parquet",
        """
SELECT
    e.id,
    e.source_id,
    e.visitor_id,
    e.event_time,
    e.ingested_at,
    e.payload ->> 'uri' AS uri,
    e.payload ->> 'http_method' AS http_method,
    (e.payload ->> 'status_code')::integer AS status_code,
    e.payload ->> 'resource_type' AS resource_type,
    (e.payload ->> 'referrer_present')::boolean AS referrer_present,
    e.payload ->> 'user_agent' AS user_agent,
    (e.payload ->> 'bytes')::bigint AS bytes
FROM public.events AS e
""".strip(),
    ),
    Snapshot(
        "decisions",
        "decisions.parquet",
        """
SELECT
    d.id,
    d.event_id,
    d.source_id,
    d.visitor_id,
    d.decided_at,
    d.status,
    d.score_raw,
    d.score_normalized,
    d.risk_level,
    d.action,
    d.policy_version,
    d.baseline_run_id
FROM public.decisions AS d
""".strip(),
    ),
    Snapshot(
        "baseline_runs",
        "baseline_runs.parquet",
        """
SELECT
    b.id,
    b.source_id,
    b.status,
    b.fit_at,
    b.window_count,
    b.distinct_visitors,
    b.event_count,
    b.registry_hash,
    b.features_fit
FROM public.baseline_runs AS b
""".strip(),
    ),
)