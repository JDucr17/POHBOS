select
    traffic.source_id as traffic_source_id,
    mart.source_id as mart_source_id,
    traffic.event_date as traffic_event_date,
    mart.event_date as mart_event_date
from {{ ref('int_daily_event_traffic_by_source') }} as traffic
full outer join {{ ref('mart_source_daily_risk_summary') }} as mart
    on traffic.source_id = mart.source_id
    and traffic.event_date = mart.event_date
where
    traffic.source_id is null
    or mart.source_id is null
    or traffic.request_count is distinct from mart.request_count
    or traffic.distinct_visitor_count is distinct from mart.distinct_visitor_count
    or traffic.distinct_uri_count is distinct from mart.distinct_uri_count
    or traffic.http_4xx_count is distinct from mart.http_4xx_count
    or traffic.http_4xx_rate is distinct from mart.http_4xx_rate
    or traffic.referrer_present_count is distinct from mart.referrer_present_count
    or traffic.referrer_present_rate is distinct from mart.referrer_present_rate
