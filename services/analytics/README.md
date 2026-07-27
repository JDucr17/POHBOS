# Analytics Service 

The analytics service turns stored pipeline data into insights about traffic and detector behavior. It summarizes total requests, scoring outcomes, risk levels, recommended actions, score trends, and baseline usage.

## Extraction 

The extractor service is responsible of creating Parquet representations of the events, decisions and baseline runs stored in the operational pipeline databse. 

### Extract snapshots

`pohbos-analytics extract [OPTIONS]`

### Options

* `--postgres-url TEXT` — Operational Postgres connection URL. When omitted, the command checks `POHBOS_ANALYTICS_POSTGRES_URL` and then `DATABASE_URL`.
* `--raw-dir PATH` — Directory where the Parquet snapshots are written. Defaults to `data/raw`.  

### Output

The command produces:

* `data/raw/events.parquet`
* `data/raw/decisions.parquet`
* `data/raw/baseline_runs.parquet`

### Example

```bash
pohbos-analytics extract --raw-dir data/raw
```

### Inspect snapshots

`pohbos-analytics inspect-raw [OPTIONS]`

#### Options

* `--raw-dir PATH` — Directory containing the Parquet snapshots. Defaults to `data/raw`.

The command displays each snapshot’s path, row count, and columns. 


## Publish the daily risk summary

Build and test the complete dependency graph, then publish the daily risk mart
to the Astro app as Parquet:

```bash
uv run dbt build \
  --project-dir dbt \
  --profiles-dir dbt \
  --select +source_daily_risk_summary
```

The generated artifact is written to:

```text
../../apps/showcase/public/data/source_daily_risk_summary.parquet
```
