select
    source_id,
    event_date,
    count(*) as row_count
from {{ ref('mart_source_daily_risk_summary') }}
group by
    source_id,
    event_date
having count(*) > 1
