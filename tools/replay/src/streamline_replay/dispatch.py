import asyncio
from collections.abc import Callable, Iterator

from streamline_replay.client import IngestClient, IngestResult
from streamline_replay.models import IngestEvent
from streamline_replay.rate import RequestScheduler
from streamline_replay.stats import ReplayStats

DEFAULT_PROGRESS_EVERY = 10000

# Progress reporting callback, caller defines reporting shape
ProgressCallback = Callable[[ReplayStats], None]


async def dispatch_events(
    client: IngestClient,
    events: Iterator[IngestEvent],
    scheduler: RequestScheduler,
    stats: ReplayStats,
    concurrency: int,
    on_progress: ProgressCallback | None = None,
    progress_every: int = DEFAULT_PROGRESS_EVERY,
) -> None:
    in_flight: set[asyncio.Task[IngestResult]] = set()
    next_report = progress_every

    for event in events:
        # blocks until the scheduler allows the next dispatch
        await scheduler.wait()

        in_flight.add(asyncio.create_task(client.send(event)))

        # cap amount of in-flight requests given the concurrency limit
        if len(in_flight) >= concurrency:
            done, in_flight = await asyncio.wait(
                in_flight,
                return_when=asyncio.FIRST_COMPLETED,
            )
            record_completed(stats, done)

            if on_progress is not None and stats.attempted >= next_report:
                on_progress(stats)
                next_report += progress_every

    if in_flight:
        done, _ = await asyncio.wait(in_flight)
        record_completed(stats, done)


def record_completed(
    stats: ReplayStats,
    tasks: set[asyncio.Task[IngestResult]],
) -> None:
    for task in tasks:
        if task.cancelled():
            stats.record_failure("send task cancelled")
            continue

        exc = task.exception()
        if exc is not None:
            stats.record_failure(repr(exc))
            continue

        stats.record(task.result())
