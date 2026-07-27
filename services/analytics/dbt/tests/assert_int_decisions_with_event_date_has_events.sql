-- depends_on: {{ ref('int_decisions_with_event_date') }}

select
    decisions.decision_id,
    decisions.event_id
from {{ ref('stg_pipeline__decisions') }} as decisions
left join {{ ref('stg_pipeline__events') }} as events
    on decisions.event_id = events.event_id
where events.event_id is null
