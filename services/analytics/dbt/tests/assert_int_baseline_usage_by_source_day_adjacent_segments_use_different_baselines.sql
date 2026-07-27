with segments_with_previous_baseline as (

    select
        source_id,
        event_date,
        segment_number,
        baseline_run_id,
        lag(baseline_run_id) over (
            partition by source_id, event_date
            order by segment_number
        ) as previous_baseline_run_id
    from {{ ref('int_baseline_usage_by_source_day') }}

)

select *
from segments_with_previous_baseline
where baseline_run_id = previous_baseline_run_id
