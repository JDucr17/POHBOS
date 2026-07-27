with baseline_backed_decisions as (

    select count(*) as decision_count
    from {{ ref('int_decisions_with_event_date') }}
    where baseline_run_id is not null

),

aggregated_baseline_usage as (

    select sum(baseline_usage_window_count) as decision_count
    from {{ ref('int_baseline_usage_by_source_day') }}

)

select
    baseline_backed_decisions.decision_count as input_decision_count,
    aggregated_baseline_usage.decision_count as aggregated_decision_count
from baseline_backed_decisions
cross join aggregated_baseline_usage
where baseline_backed_decisions.decision_count
    is distinct from aggregated_baseline_usage.decision_count
