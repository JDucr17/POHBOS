with decisions as (

    select * from {{ ref('stg_pipeline__decisions') }}

),

events as (

    select * from {{ ref('stg_pipeline__events') }}

),

decisions_with_event_date as (

    select
        decisions.decision_id,
        decisions.event_id,
        decisions.source_id,
        decisions.visitor_id,
        events.event_time,
        events.event_date,
        decisions.decided_at,
        decisions.status,
        decisions.score_raw,
        decisions.score_normalized,
        decisions.risk_level,
        decisions.action,
        decisions.policy_version,
        decisions.baseline_run_id
    from decisions
    left join events
        on decisions.event_id = events.event_id

)

select * from decisions_with_event_date
