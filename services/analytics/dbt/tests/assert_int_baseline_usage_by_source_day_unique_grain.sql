select
    source_id,
    event_date,
    segment_number,
    count(*) as row_count
from {{ ref('int_baseline_usage_by_source_day') }}
group by
    source_id,
    event_date,
    segment_number
having count(*) > 1
