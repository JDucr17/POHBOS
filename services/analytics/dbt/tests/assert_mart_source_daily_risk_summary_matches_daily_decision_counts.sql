with decision_outcomes as (

    select * from {{ ref('int_daily_decision_outcomes_by_source') }}

),

mart as (

    select * from {{ ref('mart_source_daily_risk_summary') }}

),

missing_from_mart as (

    select
        decision_outcomes.source_id,
        decision_outcomes.event_date,
        'missing_from_mart' as failure
    from decision_outcomes
    left join mart
        on decision_outcomes.source_id = mart.source_id
        and decision_outcomes.event_date = mart.event_date
    where mart.source_id is null

),

unexpected_mart_counts as (

    select
        mart.source_id,
        mart.event_date,
        'unexpected_mart_counts' as failure
    from mart
    left join decision_outcomes
        on mart.source_id = decision_outcomes.source_id
        and mart.event_date = decision_outcomes.event_date
    where
        decision_outcomes.source_id is null
        and (
            mart.evaluated_window_count is distinct from 0
            or mart.scored_window_count is distinct from 0
            or mart.no_baseline_window_count is distinct from 0
            or mart.insufficient_history_window_count is distinct from 0
            or mart.low_risk_count is distinct from 0
            or mart.medium_risk_count is distinct from 0
            or mart.high_risk_count is distinct from 0
            or mart.critical_risk_count is distinct from 0
            or mart.allow_action_count is distinct from 0
            or mart.log_action_count is distinct from 0
            or mart.challenge_action_count is distinct from 0
            or mart.block_action_count is distinct from 0
        )

),

changed_mart_counts as (

    select
        decision_outcomes.source_id,
        decision_outcomes.event_date,
        'changed_mart_counts' as failure
    from decision_outcomes
    inner join mart
        on decision_outcomes.source_id = mart.source_id
        and decision_outcomes.event_date = mart.event_date
    where
        decision_outcomes.evaluated_window_count
            is distinct from mart.evaluated_window_count
        or decision_outcomes.scored_window_count
            is distinct from mart.scored_window_count
        or decision_outcomes.no_baseline_window_count
            is distinct from mart.no_baseline_window_count
        or decision_outcomes.insufficient_history_window_count
            is distinct from mart.insufficient_history_window_count
        or decision_outcomes.low_risk_count is distinct from mart.low_risk_count
        or decision_outcomes.medium_risk_count is distinct from mart.medium_risk_count
        or decision_outcomes.high_risk_count is distinct from mart.high_risk_count
        or decision_outcomes.critical_risk_count is distinct from mart.critical_risk_count
        or decision_outcomes.allow_action_count is distinct from mart.allow_action_count
        or decision_outcomes.log_action_count is distinct from mart.log_action_count
        or decision_outcomes.challenge_action_count is distinct from mart.challenge_action_count
        or decision_outcomes.block_action_count is distinct from mart.block_action_count

)

select * from missing_from_mart
union all
select * from unexpected_mart_counts
union all
select * from changed_mart_counts
