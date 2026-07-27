with staged_decisions as (

    select count(*) as decision_count
    from {{ ref('stg_pipeline__decisions') }}

),

enriched_decisions as (

    select count(*) as decision_count
    from {{ ref('int_decisions_with_event_date') }}

)

select
    staged_decisions.decision_count as staged_decision_count,
    enriched_decisions.decision_count as enriched_decision_count
from staged_decisions
cross join enriched_decisions
where staged_decisions.decision_count
    is distinct from enriched_decisions.decision_count
