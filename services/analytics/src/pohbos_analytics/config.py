import os
from collections.abc import Mapping
from pathlib import Path


DEFAULT_RAW_DIR = Path("data/raw")
ANALYTICS_POSTGRES_URL_ENV = "POHBOS_ANALYTICS_POSTGRES_URL"
DATABASE_URL_ENV = "DATABASE_URL"


def resolve_postgres_url(
    cli_value: str | None, environ: Mapping[str, str] | None = None
) -> str | None:
    environment = os.environ if environ is None else environ
    return (
        cli_value
        or environment.get(ANALYTICS_POSTGRES_URL_ENV)
        or environment.get(DATABASE_URL_ENV)
    )
