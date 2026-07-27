select *
from {{ ref('mart_source_daily_risk_summary') }}
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
            or p99_normalized_score > maximum_normalized_score
        )
    )
