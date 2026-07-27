with events as (

    select * from {{ ref('stg_pipeline__events') }}

),

aggregated_to_source_day as (

    select
        source_id,
        event_date,
        count(*) as request_count,
        count(distinct visitor_id) as distinct_visitor_count,
        count(distinct uri) as distinct_uri_count,
        count(*) filter (where is_4xx) as http_4xx_count,
        count(*) filter (where referrer_present) as referrer_present_count
    from events
    group by
        source_id,
        event_date

),

calculated_rates as (

    select
        source_id,
        event_date,
        request_count,
        distinct_visitor_count,
        distinct_uri_count,
        http_4xx_count,
        cast(http_4xx_count as double) / request_count as http_4xx_rate,
        referrer_present_count,
        cast(referrer_present_count as double) / request_count
            as referrer_present_rate
    from aggregated_to_source_day

)

select * from calculated_rates
