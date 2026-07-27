with expected_baseline_usage as (

    select
        source_id,
        event_date,
        list(
            struct_pack(
                segment_number := segment_number,
                baseline_run_id := baseline_run_id,
                first_event_time := first_event_time,
                last_event_time := last_event_time,
                baseline_usage_window_count := baseline_usage_window_count,
                scored_window_count := scored_window_count,
                insufficient_history_window_count := insufficient_history_window_count,
                baseline_fit_at := baseline_fit_at
            )
            order by segment_number
        ) as baseline_usage_segments
    from {{ ref('int_baseline_usage_by_source_day') }}
    group by
        source_id,
        event_date

)

select
    expected.source_id as expected_source_id,
    mart.source_id as mart_source_id,
    expected.event_date as expected_event_date,
    mart.event_date as mart_event_date
from expected_baseline_usage as expected
full outer join {{ ref('mart_source_daily_risk_summary') }} as mart
    on expected.source_id = mart.source_id
    and expected.event_date = mart.event_date
where
    mart.source_id is null
    or (
        expected.source_id is null
        and mart.baseline_usage_segments is not null
    )
    or (
        expected.source_id is not null
        and expected.baseline_usage_segments
            is distinct from mart.baseline_usage_segments
    )
