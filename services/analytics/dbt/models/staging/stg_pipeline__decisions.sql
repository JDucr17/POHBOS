with source as (
    select * from {{ source('pipeline_raw', 'decisions') }}
),

renamed as (
    select
        id as decision_id,
        event_id,
        source_id,
        visitor_id,
        decided_at,
        status,
        score_raw,
        score_normalized,
        risk_level,
        action,
        policy_version,
        baseline_run_id
    from source
)

select * from renamed
