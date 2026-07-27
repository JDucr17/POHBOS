select *
from {{ ref('int_daily_decision_outcomes_by_source') }}
where
    (
        scored_window_count = 0
        and (
            average_normalized_score is not null
            or p90_normalized_score is not null
            or p99_normalized_score is not null
            or maximum_normalized_score is not null
        )
    )
    or (
        scored_window_count > 0
        and (
            average_normalized_score is null
            or average_normalized_score not between 0 and 1
            or p90_normalized_score is null
            or p90_normalized_score not between 0 and 1
            or p99_normalized_score is null
            or p99_normalized_score not between 0 and 1
            or maximum_normalized_score is null
            or maximum_normalized_score not between 0 and 1
            or p90_normalized_score > p99_normalized_score
            or maximum_normalized_score < p99_normalized_score
        )
    )
