from dataclasses import dataclass
from pathlib import Path

import duckdb

from pohbos_analytics.errors import SnapshotInspectionError
from pohbos_analytics.snapshots import EXPECTED_SNAPSHOTS, Snapshot


@dataclass(frozen=True, slots=True)
class SnapshotColumn:
    name: str
    data_type: str


@dataclass(frozen=True, slots=True)
class SnapshotInspection:
    filename: str
    path: Path
    row_count: int
    columns: tuple[SnapshotColumn, ...]


def inspect_raw(raw_dir: Path) -> tuple[SnapshotInspection, ...]:
    snapshots = [
        (snapshot, raw_dir / snapshot.filename) for snapshot in EXPECTED_SNAPSHOTS
    ]
    missing = [path.name for _, path in snapshots if not path.is_file()]
    if missing:
        raise SnapshotInspectionError(
            f"missing expected snapshot file(s) in {raw_dir}: {', '.join(missing)}"
        )

    try:
        connection = duckdb.connect()
    except duckdb.Error as error:
        raise SnapshotInspectionError(
            f"could not connect to DuckDB for raw snapshot inspection: {error}"
        ) from error

    try:
        return tuple(
            _inspect_snapshot(connection, snapshot, path)
            for snapshot, path in snapshots
        )
    finally:
        connection.close()


def _inspect_snapshot(
    connection: duckdb.DuckDBPyConnection,
    snapshot: Snapshot,
    path: Path,
) -> SnapshotInspection:
    try:
        relation = connection.read_parquet(str(path))
        count_row = relation.aggregate("count(*)").fetchone()
        if count_row is None:
            raise SnapshotInspectionError(f"could not count rows in snapshot: {path}")
        columns = tuple(
            SnapshotColumn(name=name, data_type=str(data_type))
            for name, data_type in zip(
                relation.columns,
                relation.types,
                strict=True,
            )
        )
    except duckdb.Error as error:
        raise SnapshotInspectionError(
            f"could not read snapshot {snapshot.filename} at {path}: {error}"
        ) from error

    return SnapshotInspection(
        filename=snapshot.filename,
        path=path,
        row_count=int(count_row[0]),
        columns=columns,
    )
