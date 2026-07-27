from pathlib import Path
from typing import Annotated

import typer

from pohbos_analytics.errors import ExtractionError, SnapshotInspectionError
from pohbos_analytics.extract import extract_raw
from pohbos_analytics.inspect import SnapshotInspection, inspect_raw
from pohbos_analytics.config import DEFAULT_RAW_DIR, resolve_postgres_url

RawDirOption = Annotated[
    Path,
    typer.Option(
        "--raw-dir",
        file_okay=False,
        dir_okay=True,
        help="Directory containing raw Parquet snapshots.",
    ),
]

PostgresURLOption = Annotated[
    str | None,
    typer.Option(
        "--postgres-url",
        help="Operational Postgres URL.",
    ),
]

app = typer.Typer(help="Extract and inspect Streamline analytics snapshots.")


@app.command()
def version() -> None:
    """Print the analytics package version."""
    typer.echo("pohbos-analytics 0.1.0")


@app.command("inspect-raw")
def inspect_raw_command(raw_dir: RawDirOption = DEFAULT_RAW_DIR) -> None:
    """Inspect existing raw Parquet snapshots."""
    try:
        inspections = inspect_raw(raw_dir)
    except SnapshotInspectionError as error:
        typer.echo(f"error: {error}", err=True)
        raise typer.Exit(code=1) from None

    _render_inspections(inspections)


@app.command()
def extract(
    postgres_url: PostgresURLOption = None,
    raw_dir: RawDirOption = DEFAULT_RAW_DIR,
) -> None:
    """Export operational Postgres data to raw Parquet snapshots."""
    try:
        resolved_url = resolve_postgres_url(postgres_url)
        if resolved_url is None:
            raise ExtractionError(
                "no Postgres URL provided; use --postgres-url, "
                "POHBOS_ANALYTICS_POSTGRES_URL, or DATABASE_URL"
            )
        inspections = extract_raw(resolved_url, raw_dir)
    except (ExtractionError, SnapshotInspectionError) as error:
        typer.echo(f"error: {error}", err=True)
        raise typer.Exit(code=1) from None

    _render_inspections(inspections)


def _render_inspections(inspections: tuple[SnapshotInspection, ...]) -> None:
    typer.echo("Raw snapshot inspection")
    for inspection in inspections:
        typer.echo()
        typer.echo(inspection.filename)
        typer.echo(f"  path: {inspection.path}")
        typer.echo(f"  rows: {inspection.row_count}")
        typer.echo("  columns:")
        for column in inspection.columns:
            typer.echo(f"    {column.name} {column.data_type}")
