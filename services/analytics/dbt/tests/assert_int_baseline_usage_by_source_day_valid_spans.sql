select *
from {{ ref('int_baseline_usage_by_source_day') }}
where
    first_event_time > last_event_time
    or first_decided_at > last_decided_at
    or baseline_fit_at > first_decided_at
    or cast(first_event_time at time zone 'UTC' as date)
        is distinct from event_date
    or cast(last_event_time at time zone 'UTC' as date)
        is distinct from event_date
