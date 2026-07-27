from datetime import datetime, timezone

from airflow.providers.standard.operators.bash import BashOperator
from airflow.sdk import DAG


ANALYTICS_DIR = "/opt/streamline/services/analytics"


with DAG(
    dag_id="pohbos_analytics_publish",
    schedule="@weekly",
    start_date=datetime(2026, 1, 1, tzinfo=timezone.utc),
    catchup=False,
    max_active_runs=1,
) as dag:
    extract = BashOperator(
        task_id="extract",
        bash_command=".venv/bin/pohbos-analytics extract --raw-dir data/raw",
        cwd=ANALYTICS_DIR,
        append_env=True,
    )

    build_and_test = BashOperator(
        task_id="build_and_test",
        bash_command="""
            .venv/bin/dbt build \
              --project-dir dbt \
              --profiles-dir dbt \
              --select +mart_source_daily_risk_summary
        """,
        cwd=ANALYTICS_DIR,
        append_env=True,
    )

    export = BashOperator(
        task_id="export",
        bash_command="""
            .venv/bin/dbt run \
              --project-dir dbt \
              --profiles-dir dbt \
              --select source_daily_risk_summary
        """,
        cwd=ANALYTICS_DIR,
        append_env=True,
    )

    extract >> build_and_test >> export
