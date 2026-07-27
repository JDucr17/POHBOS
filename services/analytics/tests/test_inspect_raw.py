from pathlib import Path

import duckdb
from typer.testing import CliRunner

from pohbos_analytics.cli import app
from pohbos_analytics.config import DEFAULT_RAW_DIR
from pohbos_analytics.snapshots import EXPECTED_SNAPSHOTS

runner = CliRunner()


def write_snapshots(raw_dir: Path) -> None:
    queries = {
        "events.parquet": """
            SELECT UUID '01900000-0000-7000-8000-000000000001' AS id,
                   'source-a'::VARCHAR AS source_id,
                   TIMESTAMPTZ '2026-07-20 12:00:00+00' AS event_time,
                   200::INTEGER AS status_code
        """,
        "decisions.parquet": """
            SELECT UUID '01900000-0000-7000-8000-000000000002' AS id,
                   UUID '01900000-0000-7000-8000-000000000001' AS event_id,
                   'scored'::VARCHAR AS status, 'allow'::VARCHAR AS action
        """,
        "baseline_runs.parquet": """
            SELECT 1::BIGINT AS id, 'source-a'::VARCHAR AS source_id,
                   'succeeded'::VARCHAR AS status, 10::INTEGER AS window_count
        """,
    }
    for filename, query in queries.items():
        duckdb.sql(query).write_parquet(str(raw_dir / filename))


def test_inspect_raw_reports_schema_for_readable_snapshots(tmp_path: Path) -> None:
    write_snapshots(tmp_path)

    result = runner.invoke(app, ["inspect-raw", "--raw-dir", str(tmp_path)])

    assert result.exit_code == 0
    for snapshot in EXPECTED_SNAPSHOTS:
        assert snapshot.filename in result.stdout
    assert result.stdout.count("rows: 1") == 3
    assert "id UUID" in result.stdout
    assert "event_id UUID" in result.stdout
    assert "event_time TIMESTAMP WITH TIME ZONE" in result.stdout
    assert "action VARCHAR" in result.stdout
    assert "window_count INTEGER" in result.stdout


def test_inspect_raw_missing_snapshots_returns_clear_error(tmp_path: Path) -> None:
    result = runner.invoke(app, ["inspect-raw", "--raw-dir", str(tmp_path)])

    assert result.exit_code == 1
    assert "missing expected snapshot file(s)" in result.stderr
    for snapshot in EXPECTED_SNAPSHOTS:
        assert snapshot.filename in result.stderr


def test_inspect_raw_unreadable_snapshot_returns_clear_error(tmp_path: Path) -> None:
    write_snapshots(tmp_path)
    (tmp_path / "events.parquet").write_text("not parquet")

    result = runner.invoke(app, ["inspect-raw", "--raw-dir", str(tmp_path)])

    assert result.exit_code == 1
    assert "could not read snapshot events.parquet" in result.stderr


def test_cli_help_lists_extract_and_inspect_raw_commands() -> None:
    result = runner.invoke(app, ["--help"])

    assert result.exit_code == 0
    assert "extract" in result.stdout
    assert "inspect-raw" in result.stdout


def test_cli_version_prints_package_version() -> None:
    result = runner.invoke(app, ["version"])

    assert result.exit_code == 0
    assert result.stdout == "pohbos-analytics 0.1.0\n"


def test_raw_snapshot_contract_uses_expected_default_dir_and_filenames() -> None:
    assert DEFAULT_RAW_DIR == Path("data/raw")
    assert tuple(snapshot.filename for snapshot in EXPECTED_SNAPSHOTS) == (
        "events.parquet",
        "decisions.parquet",
        "baseline_runs.parquet",
    )
