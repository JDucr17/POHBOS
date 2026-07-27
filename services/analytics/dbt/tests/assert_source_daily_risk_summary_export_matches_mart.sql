(
    select *
    from {{ ref('mart_source_daily_risk_summary') }}

    except all

    select *
    from {{ ref('source_daily_risk_summary') }}
)

union all

(
    select *
    from {{ ref('source_daily_risk_summary') }}

    except all

    select *
    from {{ ref('mart_source_daily_risk_summary') }}
)
