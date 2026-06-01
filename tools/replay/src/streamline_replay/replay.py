from collections.abc import Callable, Iterator
from functools import partial
from itertools import islice
from pathlib import Path
from time import perf_counter

from streamline_replay.adapters.eclog import iter_eclog_events
from streamline_replay.client import open_ingest_client
from streamline_replay.dispatch import ProgressCallback, dispatch_events
from streamline_replay.models import IngestEvent
from streamline_replay.rate import RequestScheduler
from streamline_replay.stats import ReplayStats, ReplaySummary
from streamline_replay.errors import EmptyInputError


async def replay_eclog(
    input_path: Path,
    target_url: str,
    source_id: str,
    limit: int | None = None,
    timeout_seconds: float = 5.0,
    concurrency: int = 1,
    rate: float | None = None,
    loop: bool = False,
    on_progress: ProgressCallback | None = None,
) -> ReplaySummary:
    if concurrency <= 0:
        raise ValueError("concurrency must be positive")

    if loop and limit is not None:
        raise ValueError("limit and loop cannot be combined")

    stats = ReplayStats()
    scheduler = RequestScheduler(rate)
    started_at = perf_counter()

    start_pass = partial(iter_eclog_events, input_path, source_id)
    events = replay_passes(start_pass, loop=loop)
    if limit is not None:
        events = islice(events, limit)

    async with open_ingest_client(
        target_url=target_url,
        timeout_seconds=timeout_seconds,
        concurrency=concurrency,
    ) as client:
        await dispatch_events(
            client,
            events,
            scheduler,
            stats,
            concurrency,
            on_progress=on_progress,
        )

    elapsed_seconds = perf_counter() - started_at
    return stats.summary(elapsed_seconds)


def replay_passes(
    start_pass: Callable[[], Iterator[IngestEvent]],
    *,
    loop: bool,
) -> Iterator[IngestEvent]:
    """Yield one pass over the source, then repeat while looping.

    The first pass must yield events; an empty first pass means the input
    is faulty, so fail fast rather than send nothing or loop forever.
    """
    first_pass_was_empty = True
    for event in start_pass():
        first_pass_was_empty = False
        yield event

    if first_pass_was_empty:
        raise EmptyInputError()

    while loop:
        yield from start_pass()
