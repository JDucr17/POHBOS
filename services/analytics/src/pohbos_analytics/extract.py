from pathlib import Path

import duckdb

from pohbos_analytics.errors import ExtractionError
from pohbos_analytics.inspect import SnapshotInspection, inspect_raw
from pohbos_analytics.snapshots import EXPECTED_SNAPSHOTS, Snapshot


def extract_raw(
    postgres_url: str, raw_dir: Path
) -> tuple[SnapshotInspection, ...]:
    try:
        raw_dir.mkdir(parents=True, exist_ok=True)
    except OSError as error:
        raise ExtractionError(f"could not create raw directory {raw_dir}: {error}") from error

    try:
        connection = duckdb.connect()
    except duckdb.Error as error:
        raise ExtractionError(f"could not connect to DuckDB: {error}") from error

    try:
        _load_postgres_extension(connection)
        _attach_postgres(connection, postgres_url)
        _export_snapshots(connection, raw_dir)
    finally:
        connection.close()

    return inspect_raw(raw_dir)


def _load_postgres_extension(connection: duckdb.DuckDBPyConnection) -> None:
    try:
        connection.install_extension("postgres")
        connection.load_extension("postgres")
    except duckdb.Error as error:
        raise ExtractionError(
            f"could not install or load DuckDB Postgres extension: {error}"
        ) from error


def _attach_postgres(
    connection: duckdb.DuckDBPyConnection, postgres_url: str
) -> None:
    try:
        connection.execute(
            f"ATTACH {_sql_literal(postgres_url)} AS operational "
            "(TYPE postgres, READ_ONLY)"
        )
        connection.execute("USE operational")
    except duckdb.Error as error:
        raise ExtractionError(f"could not attach operational Postgres: {error}") from error


def _export_snapshots(connection: duckdb.DuckDBPyConnection, raw_dir: Path) -> None:
    paths = [
        (
            snapshot,
            raw_dir / snapshot.filename,
            raw_dir / f"{snapshot.filename}.tmp",
        )
        for snapshot in EXPECTED_SNAPSHOTS
    ]
    _remove_temp_files(paths)

    transaction_started = False
    current_filename = "raw snapshots"
    try:
        # Attached Postgres reads share DuckDB's repeatable-read transaction
        # final files are published only after every staged COPY commits.
        connection.execute("BEGIN TRANSACTION")
        transaction_started = True
        for snapshot, _, temp_path in paths:
            current_filename = snapshot.filename
            connection.execute(
                f"COPY ({snapshot.select_sql}) TO {_sql_literal(str(temp_path))} "
                "(FORMAT PARQUET, COMPRESSION ZSTD)"
            )
        connection.execute("COMMIT")
        transaction_started = False
    except duckdb.Error as error:
        if transaction_started:
            _rollback(connection)
        _remove_temp_files(paths)
        raise ExtractionError(f"could not export {current_filename}: {error}") from error

    for _, target_path, temp_path in paths:
        try:
            temp_path.replace(target_path)
        except OSError as error:
            _remove_temp_files(paths)
            raise ExtractionError(
                f"could not replace raw snapshot {target_path}: {error}"
            ) from error


def _remove_temp_files(paths: list[tuple[Snapshot, Path, Path]]) -> None:
    for _, _, temp_path in paths:
        try:
            temp_path.unlink(missing_ok=True)
        except OSError as error:
            raise ExtractionError(
                f"could not remove temporary snapshot {temp_path}: {error}"
            ) from error


def _rollback(connection: duckdb.DuckDBPyConnection) -> None:
    try:
        connection.execute("ROLLBACK")
    except duckdb.Error:
        pass


def _sql_literal(value: str) -> str:
    return "'" + value.replace("'", "''") + "'"
