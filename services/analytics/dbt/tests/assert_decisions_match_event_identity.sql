-- depends_on: {{ ref('int_decisions_with_event_date') }}

select
    decisions.decision_id,
    decisions.event_id,
    decisions.source_id as decision_source_id,
    events.source_id as event_source_id,
    decisions.visitor_id as decision_visitor_id,
    events.visitor_id as event_visitor_id
from {{ ref('stg_pipeline__decisions') }} as decisions
inner join {{ ref('stg_pipeline__events') }} as events
    on decisions.event_id = events.event_id
where
    decisions.source_id is distinct from events.source_id
    or decisions.visitor_id is distinct from events.visitor_id
