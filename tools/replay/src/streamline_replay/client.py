from collections.abc import AsyncIterator
from contextlib import asynccontextmanager
from dataclasses import dataclass
from time import perf_counter

import aiohttp
import orjson

from streamline_replay.models import IngestEvent

JSON_HEADERS = {"Content-Type": "application/json"}


@dataclass(frozen=True, slots=True)
class IngestResult:
    ok: bool
    status_code: int | None
    latency_ms: float
    error: str | None = None


class IngestClient:
    """Sends events to the Streamline ingestor.

    The wrapped aiohttp.ClientSession is borrowed; its lifecycle is owned
    by the open_ingest_client factory.
    """

    def __init__(self, target_url: str, session: aiohttp.ClientSession) -> None:
        self._target_url = target_url
        self._session = session

    async def send(self, event: IngestEvent) -> IngestResult:
        started_at = perf_counter()
        body = orjson.dumps(event.to_payload())

        try:
            async with self._session.post(self._target_url, data=body) as response:
                ok = 200 <= response.status < 300
                error = None if ok else (await response.text())[:500]

                return IngestResult(
                    ok=ok,
                    status_code=response.status,
                    latency_ms=_elapsed_ms(started_at),
                    error=error,
                )

        except aiohttp.ClientError as err:
            return IngestResult(
                ok=False,
                status_code=None,
                latency_ms=_elapsed_ms(started_at),
                error=str(err),
            )


@asynccontextmanager
async def open_ingest_client(
    target_url: str,
    timeout_seconds: float = 5.0,
    concurrency: int = 200,
) -> AsyncIterator[IngestClient]:
    """Open an IngestClient backed by a pooled aiohttp.ClientSession.

    The connection pool is sized to the concurrency ceiling so it does not
    silently throttle in-flight sends.
    """
    connector = aiohttp.TCPConnector(
        limit=concurrency,
        limit_per_host=concurrency,
    )
    timeout = aiohttp.ClientTimeout(total=timeout_seconds)

    async with aiohttp.ClientSession(
        connector=connector,
        timeout=timeout,
        headers=JSON_HEADERS,
    ) as session:
        yield IngestClient(target_url, session)


def _elapsed_ms(started_at: float) -> float:
    return (perf_counter() - started_at) * 1000
