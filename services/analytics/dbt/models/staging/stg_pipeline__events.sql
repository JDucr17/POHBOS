with source as (
    select * from {{ source('pipeline_raw', 'events') }}
),

renamed as (
    select
        id as event_id,
        source_id,
        visitor_id,
        event_time,
        cast(event_time at time zone 'UTC' as date) as event_date,
        ingested_at,
        uri,
        http_method,
        status_code,
        resource_type,
        referrer_present,
        user_agent,
        bytes,
        status_code between 400 and 499 as is_4xx
    from source
)

select * from renamed
