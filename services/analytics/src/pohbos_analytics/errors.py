class SnapshotInspectionError(Exception):
    """A required raw snapshot is missing or unreadable."""


class ExtractionError(Exception):
    """Raw snapshot extraction could not complete."""
