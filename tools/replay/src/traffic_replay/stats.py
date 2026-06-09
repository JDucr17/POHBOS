from dataclasses import dataclass

from traffic_replay.client import IngestResult


@dataclass(frozen=True, slots=True)
class ReplaySummary:
    attempted: int
    accepted: int
    failed: int
    elapsed_seconds: float
    events_per_second: float
    last_error: str | None = None


@dataclass(slots=True)
class ReplayStats:
    attempted: int = 0
    accepted: int = 0
    failed: int = 0
    last_error: str | None = None

    def record(self, result: IngestResult) -> None:
        self.attempted += 1

        if result.ok:
            self.accepted += 1
            return

        self.failed += 1
        self.last_error = result.error

    def record_failure(self, error: str) -> None:
        self.attempted += 1
        self.failed += 1
        self.last_error = error

    def summary(self, elapsed_seconds: float) -> ReplaySummary:
        events_per_second = (
            self.attempted / elapsed_seconds if elapsed_seconds > 0 else 0.0
        )

        return ReplaySummary(
            attempted=self.attempted,
            accepted=self.accepted,
            failed=self.failed,
            elapsed_seconds=elapsed_seconds,
            events_per_second=events_per_second,
            last_error=self.last_error,
        )
