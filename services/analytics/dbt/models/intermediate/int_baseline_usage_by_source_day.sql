with baseline_backed_decisions as (

    select
        decision_id,
        source_id,
        event_date,
        baseline_run_id,
        event_time,
        decided_at,
        status
    from {{ ref('int_decisions_with_event_date') }}
    where baseline_run_id is not null

),

baseline_usage_segments as (

    select
        *,
        lag(baseline_run_id) over (
            partition by source_id, event_date
            order by event_time, decided_at, decision_id
        ) as previous_baseline_run_id
    from baseline_backed_decisions

),

marked_baseline_changes as (

    select
        *,
        case
            when previous_baseline_run_id is null
                or baseline_run_id is distinct from previous_baseline_run_id
                then 1
            else 0
        end as starts_new_segment
    from baseline_usage_segments

),

decisions_with_segment_number as (

    select
        *,
        cast(
            sum(starts_new_segment) over (
                partition by source_id, event_date
                order by event_time, decided_at, decision_id
                rows between unbounded preceding and current row
            ) as bigint
        ) as segment_number
    from marked_baseline_changes

),

aggregated_baseline_usage_segments as (

    select
        source_id,
        event_date,
        segment_number,
        baseline_run_id,
        count(*) as baseline_usage_window_count,
        count(*) filter (
            where status = 'scored'
        ) as scored_window_count,
        count(*) filter (
            where status = 'insufficient_history'
        ) as insufficient_history_window_count,
        min(event_time) as first_event_time,
        max(event_time) as last_event_time,
        min(decided_at) as first_decided_at,
        max(decided_at) as last_decided_at
    from decisions_with_segment_number
    group by
        source_id,
        event_date,
        segment_number,
        baseline_run_id

),

baseline_runs as (

    select * from {{ ref('stg_pipeline__baseline_runs') }}

),

baseline_run_usage_with_context as (

    select
        baseline_usage.source_id,
        baseline_usage.event_date,
        baseline_usage.segment_number,
        baseline_usage.baseline_run_id,
        baseline_usage.baseline_usage_window_count,
        baseline_usage.scored_window_count,
        baseline_usage.insufficient_history_window_count,
        baseline_usage.first_event_time,
        baseline_usage.last_event_time,
        baseline_usage.first_decided_at,
        baseline_usage.last_decided_at,
        baseline_runs.status as baseline_status,
        baseline_runs.fit_at as baseline_fit_at,
        baseline_runs.event_count as baseline_event_count,
        baseline_runs.window_count as baseline_window_count,
        baseline_runs.distinct_visitors as baseline_distinct_visitor_count,
        baseline_runs.registry_hash as baseline_registry_hash,
        baseline_runs.features_fit as baseline_features_fit
    from aggregated_baseline_usage_segments as baseline_usage
    left join baseline_runs
        on baseline_usage.baseline_run_id = baseline_runs.baseline_run_id
        and baseline_usage.source_id = baseline_runs.source_id

)

select * from baseline_run_usage_with_context
