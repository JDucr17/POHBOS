select *
from {{ ref('int_daily_decision_outcomes_by_source') }}
where
    evaluated_window_count <= 0
    or scored_window_count not between 0 and evaluated_window_count
    or no_baseline_window_count not between 0 and evaluated_window_count
    or insufficient_history_window_count not between 0 and evaluated_window_count
    or scored_window_count
        + no_baseline_window_count
        + insufficient_history_window_count
        != evaluated_window_count
    or low_risk_count not between 0 and scored_window_count
    or medium_risk_count not between 0 and scored_window_count
    or high_risk_count not between 0 and scored_window_count
    or critical_risk_count not between 0 and scored_window_count
    or low_risk_count
        + medium_risk_count
        + high_risk_count
        + critical_risk_count
        != scored_window_count
    or allow_action_count not between 0 and evaluated_window_count
    or log_action_count not between 0 and evaluated_window_count
    or challenge_action_count not between 0 and evaluated_window_count
    or block_action_count not between 0 and evaluated_window_count
    or allow_action_count
        + log_action_count
        + challenge_action_count
        + block_action_count
        != evaluated_window_count
