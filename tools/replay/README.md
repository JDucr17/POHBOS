# Request replay tool

Generates realistic HTTP traffic for the HBOS scoring pipeline by replaying
real e-commerce access logs (the EClog dataset) into the Streamline ingestor.

Dataset: EClog, Harvard Dataverse, DOI 10.7910/DVN/Z834IK
Source: https://dataverse.harvard.edu/dataset.xhtml?persistentId=doi:10.7910/DVN/Z834IK

## Installation

Requires Python 3.14+ and [uv](https://docs.astral.sh/uv/).

```bash
uv sync
```

## Usage

Replay events into a running ingestor:

```bash
uv run eclog-replay eclog \
  --input path/to/eclog.csv \
  --target-url http://localhost:8080/events \
  --source-id eclog-bookstore
```

### Commands

| Command   | Description                          |
|-----------|--------------------------------------|
| `eclog`   | Replay EClog rows into the ingestor. |
| `version` | Print the replay driver version.     |

### Options (`eclog`)

| Option              | Default    | Description                                                        |
|---------------------|------------|--------------------------------------------------------------------|
| `--input`, `-i`     | (required) | Path to the EClog CSV/tabular file.                                |
| `--target-url`      | (required) | Full ingestor endpoint, e.g. `http://localhost:8080/events`.       |
| `--source-id`       | (required) | Source identifier attached to replayed events.                     |
| `--limit`           | whole file | Maximum events to replay. Cannot be combined with `--loop`.        |
| `--concurrency`     | `1`        | Maximum number of in-flight HTTP requests.                         |
| `--rate`            | unbounded  | Target dispatch rate in events/sec. Omit to send as fast as possible. |
| `--loop`            | off        | Replay the dataset repeatedly until interrupted. Excludes `--limit`. |
| `--timeout-seconds` | `5.0`      | HTTP timeout per request, in seconds.                              |

### Examples

```bash
# Bounded smoke run
uv run eclog-replay eclog -i eclog.csv \
  --target-url http://localhost:8080/events --source-id eclog-bookstore --limit 1000

# Build the baseline from one honest pass of the full dataset
uv run eclog-replay eclog -i eclog.csv \
  --target-url http://localhost:8080/events --source-id eclog-bookstore

# Continuous demo traffic, paced at ~2000 events/sec, until Ctrl+C
uv run eclog-replay eclog -i eclog.csv \
  --target-url http://localhost:8080/events --source-id eclog-bookstore-live \
  --rate 2000 --concurrency 25 --loop
```

## Exit codes

| Code | Meaning                                              |
|------|------------------------------------------------------|
| `0`  | All attempted events accepted.                       |
| `1`  | Runtime failure (some sends failed, or no events).   |
| `2`  | Usage error (e.g. `--limit` combined with `--loop`). |
| `130`| Interrupted (Ctrl+C).                                |