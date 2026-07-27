from pathlib import Path

import duckdb
import pytest
from typer.testing import CliRunner

import pohbos_analytics.extract as extract_module
from pohbos_analytics.cli import app
from pohbos_analytics.config import resolve_postgres_url
from pohbos_analytics.errors import ExtractionError
from pohbos_analytics.snapshots import EXPECTED_SNAPSHOTS, Snapshot

runner = CliRunner()


def test_extract_without_postgres_url_returns_clear_error(
    tmp_path: Path,
    monkeypatch: pytest.MonkeyPatch,
) -> None:
    monkeypatch.delenv("POHBOS_ANALYTICS_POSTGRES_URL", raising=False)
    monkeypatch.delenv("DATABASE_URL", raising=False)

    result = runner.invoke(app, ["extract", "--raw-dir", str(tmp_path)])

    assert result.exit_code == 1
    assert "no Postgres URL provided" in result.stderr


def test_export_snapshots_success_replaces_existing_target_and_stale_temp(
    tmp_path: Path,
    monkeypatch: pytest.MonkeyPatch,
) -> None:
    snapshot = Snapshot("events", "events.parquet", "SELECT 2::INTEGER AS value")
    monkeypatch.setattr(extract_module, "EXPECTED_SNAPSHOTS", (snapshot,))
    target_path = tmp_path / snapshot.filename
    temp_path = tmp_path / f"{snapshot.filename}.tmp"
    target_path.write_text("previous snapshot")
    temp_path.write_text("stale temporary snapshot")
    connection = duckdb.connect()

    try:
        extract_module._export_snapshots(connection, tmp_path)
        rows = duckdb.read_parquet(str(target_path)).fetchall()
    finally:
        connection.close()

    assert rows == [(2,)]
    assert not temp_path.exists()


def test_export_snapshots_copy_failure_preserves_existing_target(
    tmp_path: Path,
    monkeypatch: pytest.MonkeyPatch,
) -> None:
    snapshot = Snapshot("events", "events.parquet", "SELECT missing_column")
    monkeypatch.setattr(extract_module, "EXPECTED_SNAPSHOTS", (snapshot,))
    target_path = tmp_path / snapshot.filename
    temp_path = tmp_path / f"{snapshot.filename}.tmp"
    target_path.write_text("previous snapshot")
    connection = duckdb.connect()

    try:
        with pytest.raises(ExtractionError, match="could not export events.parquet"):
            extract_module._export_snapshots(connection, tmp_path)
    finally:
        connection.close()

    assert target_path.read_text() == "previous snapshot"
    assert not temp_path.exists()


def test_postgres_url_resolution_cli_value_overrides_environment() -> None:
    environment = {
        "POHBOS_ANALYTICS_POSTGRES_URL": "postgresql://analytics-env",
        "DATABASE_URL": "postgresql://database-env",
    }

    assert resolve_postgres_url("postgresql://cli", environment) == "postgresql://cli"


def test_postgres_url_resolution_analytics_environment_overrides_database_url() -> None:
    environment = {
        "POHBOS_ANALYTICS_POSTGRES_URL": "postgresql://analytics-env",
        "DATABASE_URL": "postgresql://database-env",
    }

    assert resolve_postgres_url(None, environment) == "postgresql://analytics-env"


def test_snapshot_extraction_definitions_use_expected_filenames() -> None:
    assert tuple(snapshot.filename for snapshot in EXPECTED_SNAPSHOTS) == (
        "events.parquet",
        "decisions.parquet",
        "baseline_runs.parquet",
    )


def test_snapshot_extraction_queries_use_explicit_select_lists() -> None:
    for snapshot in EXPECTED_SNAPSHOTS:
        assert "SELECT *" not in snapshot.select_sql.upper()


def test_snapshot_extraction_queries_preserve_uuid_identifiers() -> None:
    snapshots = {
        snapshot.logical_name: snapshot for snapshot in EXPECTED_SNAPSHOTS
    }

    assert "e.id::text" not in snapshots["events"].select_sql
    assert "d.id::text" not in snapshots["decisions"].select_sql
    assert "d.event_id::text" not in snapshots["decisions"].select_sql


def test_events_snapshot_query_excludes_raw_json_output_columns() -> None:
    events = next(
        snapshot for snapshot in EXPECTED_SNAPSHOTS if snapshot.logical_name == "events"
    )

    assert " AS payload" not in events.select_sql
    assert " AS features" not in events.select_sql
    assert "e.features" not in events.select_sql


def test_baseline_runs_snapshot_query_excludes_model_blob() -> None:
    baseline_runs = next(
        snapshot
        for snapshot in EXPECTED_SNAPSHOTS
        if snapshot.logical_name == "baseline_runs"
    )

    assert "b.baseline" not in baseline_runs.select_sql
