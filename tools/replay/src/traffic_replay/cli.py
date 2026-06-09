import asyncio
import logging
from pathlib import Path
from typing import Annotated

import typer

from traffic_replay.observability import configure_logging
from traffic_replay.replay import replay_eclog
from traffic_replay.stats import ReplayStats
from traffic_replay.errors import EmptyInputError

logger = logging.getLogger(__name__)

DEFAULT_TIMEOUT_SECONDS = 5.0

InputPathOption = Annotated[
    Path,
    typer.Option(
        "--input",
        "-i",
        exists=True,
        file_okay=True,
        dir_okay=False,
        readable=True,
        help="Path to the EClog CSV/tabular file.",
    ),
]

TargetURLOption = Annotated[
    str,
    typer.Option(
        "--target-url",
        help="Full Streamline ingestor endpoint (e.g. http://localhost:8080/events).",
    ),
]

SourceIDOption = Annotated[
    str,
    typer.Option(
        "--source-id",
        help="Source identifier attached to replayed events.",
    ),
]

LimitOption = Annotated[
    int | None,
    typer.Option(
        "--limit",
        min=1,
        help="Maximum number of events to replay. Cannot be combined with --loop.",
    ),
]

TimeoutSecondsOption = Annotated[
    float,
    typer.Option(
        "--timeout-seconds",
        min=0.1,
        help="HTTP timeout per request, in seconds.",
    ),
]

ConcurrencyOption = Annotated[
    int,
    typer.Option(
        "--concurrency",
        min=1,
        help="Maximum number of in-flight HTTP requests.",
    ),
]

RateOption = Annotated[
    float | None,
    typer.Option(
        "--rate",
        min=0.1,
        help="Target dispatch rate in events/sec. Omit to send as fast as concurrency allows.",
    ),
]

LoopOption = Annotated[
    bool,
    typer.Option(
        "--loop",
        help="Replay the dataset repeatedly until interrupted. Cannot be combined with --limit.",
    ),
]

app = typer.Typer(help="Replay HTTP traffic into the Streamline ingestor.")


@app.command()
def version() -> None:
    """Print the replay driver version."""
    typer.echo("streamline-replay 0.1.0")


@app.command()
def eclog(
    input_path: InputPathOption,
    target_url: TargetURLOption,
    source_id: SourceIDOption,
    limit: LimitOption = None,
    timeout_seconds: TimeoutSecondsOption = DEFAULT_TIMEOUT_SECONDS,
    concurrency: ConcurrencyOption = 1,
    rate: RateOption = None,
    loop: LoopOption = False,
) -> None:
    """Replay EClog rows into the Streamline ingestor."""
    if loop and limit is not None:
        typer.echo("--limit and --loop cannot be combined", err=True)
        raise typer.Exit(code=2)

    configure_logging()

    try:
        summary = asyncio.run(
            replay_eclog(
                input_path=input_path,
                target_url=target_url,
                source_id=source_id,
                limit=limit,
                timeout_seconds=timeout_seconds,
                concurrency=concurrency,
                rate=rate,
                loop=loop,
                on_progress=_log_progress,
            )
        )

    except KeyboardInterrupt:
        typer.echo("interrupted", err=True)
        raise typer.Exit(code=130) from None

    except EmptyInputError:
        typer.echo(
            f"error: no events parsed from {input_path} — "
            f"the file may be empty, header-only, or in an unexpected format",
            err=True,
        )
        raise typer.Exit(code=1) from None

    typer.echo(
        f"attempted={summary.attempted} "
        f"accepted={summary.accepted} "
        f"failed={summary.failed} "
        f"elapsed={summary.elapsed_seconds:.2f}s "
        f"rate={summary.events_per_second:.1f}/s"
    )

    if summary.last_error:
        typer.echo(f"last_error={summary.last_error}", err=True)

    if summary.attempted == 0:
        typer.echo("no events were replayed", err=True)
        raise typer.Exit(code=1)

    if summary.failed > 0:
        raise typer.Exit(code=1)


def _log_progress(stats: ReplayStats) -> None:
    logger.info(
        "replay progress attempted=%d accepted=%d failed=%d",
        stats.attempted,
        stats.accepted,
        stats.failed,
    )
