select *
from {{ ref('mart_source_daily_risk_summary') }}
where
    request_count <= 0
    or evaluated_window_count not between 0 and request_count
    or scored_window_count < 0
    or no_baseline_window_count < 0
    or insufficient_history_window_count < 0
    or scored_window_count
        + no_baseline_window_count
        + insufficient_history_window_count
        != evaluated_window_count
    or low_risk_count < 0
    or medium_risk_count < 0
    or high_risk_count < 0
    or critical_risk_count < 0
    or low_risk_count
        + medium_risk_count
        + high_risk_count
        + critical_risk_count
        != scored_window_count
    or allow_action_count < 0
    or log_action_count < 0
    or challenge_action_count < 0
    or block_action_count < 0
    or allow_action_count
        + log_action_count
        + challenge_action_count
        + block_action_count
        != evaluated_window_count
