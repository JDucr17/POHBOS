import asyncio
from time import perf_counter


class RequestScheduler:
    """Schedules outbound request dispatches to hold a target rate."""

    def __init__(self, events_per_second: float | None) -> None:
        if events_per_second is not None and events_per_second <= 0:
            raise ValueError("events_per_second must be positive")

        self._interval: float | None = (
            None if events_per_second is None else 1.0 / events_per_second
        )
        self._next_dispatch = perf_counter()

    async def wait(self) -> None:
        if self._interval is None:
            return

        now = perf_counter()
        delay = self._next_dispatch - now

        if delay > 0:
            await asyncio.sleep(delay)
            self._next_dispatch += self._interval
            return

        # Behind schedule: resume from now, don't burst to catch up.
        self._next_dispatch = now + self._interval
