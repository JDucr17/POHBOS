with input_decisions as (

    select count(*) as decision_count
    from {{ ref('int_decisions_with_event_date') }}

),

aggregated_decisions as (

    select sum(evaluated_window_count) as decision_count
    from {{ ref('int_daily_decision_outcomes_by_source') }}

)

select
    input_decisions.decision_count as input_decision_count,
    aggregated_decisions.decision_count as aggregated_decision_count
from input_decisions
cross join aggregated_decisions
where input_decisions.decision_count
    is distinct from aggregated_decisions.decision_count
