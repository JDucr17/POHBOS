select
    source_id,
    event_date,
    count(*) as row_count
from {{ ref('int_daily_decision_outcomes_by_source') }}
group by
    source_id,
    event_date
having count(*) > 1
