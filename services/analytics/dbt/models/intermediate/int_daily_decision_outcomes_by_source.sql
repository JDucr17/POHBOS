with decisions as (

    select * from {{ ref('int_decisions_with_event_date') }}

),

aggregated_to_source_day as (

    select
        source_id,
        event_date,
        count(*) as evaluated_window_count,
        count(*) filter (
            where status = 'scored'
        ) as scored_window_count,
        count(*) filter (
            where status = 'no_baseline'
        ) as no_baseline_window_count,
        count(*) filter (
            where status = 'insufficient_history'
        ) as insufficient_history_window_count,
        count(*) filter (
            where risk_level = 'low'
        ) as low_risk_count,
        count(*) filter (
            where risk_level = 'medium'
        ) as medium_risk_count,
        count(*) filter (
            where risk_level = 'high'
        ) as high_risk_count,
        count(*) filter (
            where risk_level = 'critical'
        ) as critical_risk_count,
        count(*) filter (
            where action = 'allow'
        ) as allow_action_count,
        count(*) filter (
            where action = 'log'
        ) as log_action_count,
        count(*) filter (
            where action = 'challenge'
        ) as challenge_action_count,
        count(*) filter (
            where action = 'block'
        ) as block_action_count,
        avg(score_normalized) filter (
            where status = 'scored'
        ) as average_normalized_score,
        quantile_cont(score_normalized, 0.90) filter (
            where status = 'scored'
        ) as p90_normalized_score,
        quantile_cont(score_normalized, 0.99) filter (
            where status = 'scored'
        ) as p99_normalized_score,
        max(score_normalized) filter (
            where status = 'scored'
        ) as maximum_normalized_score
    from decisions
    group by
        source_id,
        event_date

)

select * from aggregated_to_source_day
