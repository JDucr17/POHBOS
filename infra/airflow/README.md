# Airflow

Custom Airflow image for Streamline’s baseline-refresh and analytics workflows.

It contains:

- Airflow 3.3 with Python 3.14
- the compiled baseline worker
- the analytics package and dbt project
- the DuckDB Postgres extension

## Orchestration

Apache Airflow is used to orchestrate the analytics workflow. It is divided into three tasks:

- **Extract:** Copies events, decisions, and baseline runs from Postgres into Parquet snapshots.
- **Build and test:** Uses dbt and DuckDB to build the daily source-risk mart and validate its results.
- **Export:** Writes the completed mart as the Parquet file consumed by the showcase.

The workflow runs once a week and allows only one active run at a time.

Build from the repository root:

```bash
docker build -f infra/airflow/Dockerfile -t streamline-airflow:3.3.0 .
