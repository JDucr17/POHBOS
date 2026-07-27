select *
from {{ ref('int_baseline_usage_by_source_day') }}
where
    baseline_status is null
    or baseline_status != 'succeeded'
    or baseline_fit_at is null
    or baseline_event_count is null
    or baseline_window_count is null
    or baseline_distinct_visitor_count is null
    or baseline_registry_hash is null
    or baseline_features_fit is null
