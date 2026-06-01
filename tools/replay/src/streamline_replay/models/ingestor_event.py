from dataclasses import dataclass
from datetime import datetime


@dataclass(frozen=True, slots=True)
class IngestEvent:
    source_id: str
    visitor_id: str
    event_time: datetime
    http_method: str
    uri: str
    status_code: int
    referrer_present: bool
    user_agent: str | None = None
    bytes: int | None = None

    def to_payload(self) -> dict[str, object]:
        payload: dict[str, object] = {
            "source_id": self.source_id,
            "visitor_id": self.visitor_id,
            "event_time": self.event_time.isoformat(),
            "http_method": self.http_method,
            "uri": self.uri,
            "status_code": self.status_code,
            "referrer_present": self.referrer_present,
        }

        if self.user_agent:
            payload["user_agent"] = self.user_agent

        if self.bytes is not None:
            payload["bytes"] = self.bytes

        return payload
