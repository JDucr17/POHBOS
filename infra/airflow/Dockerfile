# syntax=docker/dockerfile:1.7

FROM golang:1.25-bookworm AS baseline-worker-builder

WORKDIR /src

COPY services/pipeline/go.mod services/pipeline/go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

COPY services/pipeline/ ./
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 go build \
        -trimpath \
        -o /out/baseline-worker \
        ./cmd/baseline-worker


FROM ghcr.io/astral-sh/uv:0.11.16 AS uv


FROM apache/airflow:3.3.0-python3.14

USER root

COPY --from=uv /uv /usr/local/bin/uv
COPY --from=baseline-worker-builder \
    --chown=airflow:0 \
    /out/baseline-worker \
    /opt/streamline/bin/baseline-worker

RUN install -d -o airflow -g 0 -m 0775 \
    /opt/streamline/airflow/dags \
    /opt/streamline/airflow/src \
    /opt/streamline/services/analytics \
    /opt/streamline/services/analytics/data \
    /opt/streamline/apps/showcase/public/data

COPY --chown=airflow:0 \
    services/analytics/pyproject.toml \
    services/analytics/uv.lock \
    services/analytics/README.md \
    /opt/streamline/services/analytics/

COPY --chown=airflow:0 \
    services/analytics/src/ \
    /opt/streamline/services/analytics/src/

COPY --chown=airflow:0 \
    services/analytics/dbt/ \
    /opt/streamline/services/analytics/dbt/

COPY --chown=airflow:0 \
    infra/airflow/dags/ \
    /opt/streamline/airflow/dags/

USER airflow
WORKDIR /opt/streamline/services/analytics

RUN umask 0002 && \
    /usr/local/bin/uv sync --frozen --no-dev --no-editable

RUN .venv/bin/python -c \
    "import duckdb; connection = duckdb.connect(); connection.install_extension('postgres'); connection.load_extension('postgres'); connection.close()"

WORKDIR /opt/airflow
