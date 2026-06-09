import logging


def configure_logging(level: int = logging.INFO) -> None:
    """Configure root logging for the replay CLI.

    Plain console format for now; a structured JSON formatter matching the
    Go services' slog output is a later refinement.
    """
    logging.basicConfig(
        level=level,
        format="%(asctime)s %(levelname)s %(name)s %(message)s",
    )
