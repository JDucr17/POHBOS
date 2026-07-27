select *
from {{ ref('int_baseline_usage_by_source_day') }}
where
    baseline_usage_window_count <= 0
    or scored_window_count not between 0 and baseline_usage_window_count
    or insufficient_history_window_count
        not between 0 and baseline_usage_window_count
    or scored_window_count + insufficient_history_window_count
        != baseline_usage_window_count
