with source as (
    select * from {{ source('pipeline_raw', 'baseline_runs') }}
),

renamed as (
    select
        id as baseline_run_id,
        source_id,
        status,
        fit_at,
        event_count,
        window_count,
        distinct_visitors,
        registry_hash,
        features_fit
    from source
)

select * from renamed
