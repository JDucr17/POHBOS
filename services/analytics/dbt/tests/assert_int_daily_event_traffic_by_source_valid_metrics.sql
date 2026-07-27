select *
from {{ ref('int_daily_event_traffic_by_source') }}
where
    request_count <= 0
    or distinct_visitor_count not between 0 and request_count
    or distinct_uri_count not between 0 and request_count
    or http_4xx_count not between 0 and request_count
    or referrer_present_count not between 0 and request_count
    or http_4xx_rate not between 0 and 1
    or referrer_present_rate not between 0 and 1
