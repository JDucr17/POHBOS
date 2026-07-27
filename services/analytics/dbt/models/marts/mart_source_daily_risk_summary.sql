with traffic as (

    select * from {{ ref('int_daily_event_traffic_by_source') }}

),

decision_outcomes as (

    select * from {{ ref('int_daily_decision_outcomes_by_source') }}

),

baseline_usage as (

    select * from {{ ref('int_baseline_usage_by_source_day') }}

),

daily_baseline_usage as (

    select
        source_id,
        event_date,
        list(
            struct_pack(
                segment_number := segment_number,
                baseline_run_id := baseline_run_id,
                first_event_time := first_event_time,
                last_event_time := last_event_time,
                baseline_usage_window_count := baseline_usage_window_count,
                scored_window_count := scored_window_count,
                insufficient_history_window_count := insufficient_history_window_count,
                baseline_fit_at := baseline_fit_at
            )
            order by segment_number
        ) as baseline_usage_segments
    from baseline_usage
    group by
        source_id,
        event_date

),

source_daily_risk_summary as (

    select
        traffic.source_id,
        traffic.event_date,
        traffic.request_count,
        traffic.distinct_visitor_count,
        traffic.distinct_uri_count,
        traffic.http_4xx_count,
        traffic.http_4xx_rate,
        traffic.referrer_present_count,
        traffic.referrer_present_rate,
        coalesce(decision_outcomes.evaluated_window_count, 0) as evaluated_window_count,
        coalesce(decision_outcomes.scored_window_count, 0) as scored_window_count,
        coalesce(decision_outcomes.no_baseline_window_count, 0) as no_baseline_window_count,
        coalesce(decision_outcomes.insufficient_history_window_count, 0) as insufficient_history_window_count,
        coalesce(decision_outcomes.low_risk_count, 0) as low_risk_count,
        coalesce(decision_outcomes.medium_risk_count, 0) as medium_risk_count,
        coalesce(decision_outcomes.high_risk_count, 0) as high_risk_count,
        coalesce(decision_outcomes.critical_risk_count, 0) as critical_risk_count,
        coalesce(decision_outcomes.allow_action_count, 0) as allow_action_count,
        coalesce(decision_outcomes.log_action_count, 0) as log_action_count,
        coalesce(decision_outcomes.challenge_action_count, 0) as challenge_action_count,
        coalesce(decision_outcomes.block_action_count, 0) as block_action_count,
        decision_outcomes.average_normalized_score,
        decision_outcomes.p90_normalized_score,
        decision_outcomes.p99_normalized_score,
        decision_outcomes.maximum_normalized_score,
        daily_baseline_usage.baseline_usage_segments
    from traffic
    left join decision_outcomes
        on traffic.source_id = decision_outcomes.source_id
        and traffic.event_date = decision_outcomes.event_date
    left join daily_baseline_usage
        on traffic.source_id = daily_baseline_usage.source_id
        and traffic.event_date = daily_baseline_usage.event_date

)

select * from source_daily_risk_summary
