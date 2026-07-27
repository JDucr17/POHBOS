with staged_events as (

    select count(*) as event_count
    from {{ ref('stg_pipeline__events') }}

),

aggregated_events as (

    select sum(request_count) as event_count
    from {{ ref('int_daily_event_traffic_by_source') }}

)

select
    staged_events.event_count as staged_event_count,
    aggregated_events.event_count as aggregated_event_count
from staged_events
cross join aggregated_events
where staged_events.event_count
    is distinct from aggregated_events.event_count
