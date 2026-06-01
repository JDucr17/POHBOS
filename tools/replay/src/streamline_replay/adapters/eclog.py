"""Adapter for the EClog HTTP access-log dataset.

Dataset: EClog, Harvard Dataverse, DOI 10.7910/DVN/Z834IK.
Source: https://dataverse.harvard.edu/dataset.xhtml?persistentId=doi:10.7910/DVN/Z834IK

The EClog TimeStamp column is encoded as .NET ticks. We convert it to an
aware datetime normalized to UTC so payload serialization includes an explicit
RFC3339 offset.
"""

import csv
import logging
from collections.abc import Iterator, Mapping
from datetime import UTC, datetime, timedelta
from pathlib import Path

from streamline_replay.models import IngestEvent

logger = logging.getLogger(__name__)

_DOTNET_EPOCH = datetime(1, 1, 1, tzinfo=UTC)


def dotnet_ticks_to_datetime(ticks: int) -> datetime:
    """Convert .NET ticks to an aware datetime normalized to UTC."""
    return _DOTNET_EPOCH + timedelta(microseconds=ticks // 10)


def choose_visitor_id(ip_id: str, user_id: str) -> str:
    """Prefer UserId when present; otherwise fall back to anonymized IpId."""
    user_id = user_id.strip()
    if user_id and user_id != "-":
        return user_id

    return ip_id.strip()


def has_referrer(referrer: str) -> bool:
    referrer = referrer.strip()
    return referrer != "" and referrer != "-"


def required_field(row: Mapping[str, str | None], name: str) -> str:
    value = row[name]
    if value is None:
        raise ValueError(f"{name} is missing")

    return value


def parse_eclog_row(row: Mapping[str, str | None], source_id: str) -> IngestEvent:
    """Map one EClog CSV row into Streamline's raw ingestor event contract."""
    return IngestEvent(
        source_id=source_id,
        visitor_id=choose_visitor_id(
            required_field(row, "IpId"),
            required_field(row, "UserId"),
        ),
        event_time=dotnet_ticks_to_datetime(int(required_field(row, "TimeStamp"))),
        http_method=required_field(row, "HttpMethod").strip().upper(),
        uri=required_field(row, "Uri").strip(),
        status_code=int(required_field(row, "ResponseCode")),
        referrer_present=has_referrer(required_field(row, "Referrer")),
        user_agent=required_field(row, "UserAgent").strip() or None,
        bytes=int(required_field(row, "Bytes")),
    )


def iter_eclog_events(path: Path, source_id: str) -> Iterator[IngestEvent]:
    """Stream valid EClog events without loading the full file into memory."""
    with path.open(newline="", encoding="utf-8", errors="replace") as file:
        reader = csv.DictReader(file)

        for line_number, row in enumerate(reader, start=2):
            try:
                yield parse_eclog_row(row, source_id)
            except (KeyError, TypeError, ValueError) as err:
                logger.warning(
                    "skipping malformed eclog row",
                    extra={
                        "line_number": line_number,
                        "error": str(err),
                    },
                )
                continue
