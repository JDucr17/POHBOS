#!/usr/bin/env bash
set -euo pipefail

RAW_EVENTS_TOPIC="${KAFKA_RAW_EVENTS_TOPIC:-raw_events}"
DECISIONS_TOPIC="${KAFKA_DECISIONS_TOPIC:-decisions}"
DLQ_TOPIC="${KAFKA_DEAD_LETTER_TOPIC:-dead_letter_events}"
BASELINE_SIGNALS_TOPIC="${KAFKA_BASELINE_SIGNALS_TOPIC:-baseline_signals}"

DEV_RETENTION_MS="${KAFKA_DEV_RETENTION_MS:-21600000}"   # 6 hours
DLQ_RETENTION_MS="${KAFKA_DLQ_RETENTION_MS:-604800000}"  # 7 days

create_topic() {
  local topic="$1"
  local partitions="$2"
  local retention_ms="$3"

  echo "Ensuring topic '${topic}' exists..."

  docker compose exec -T redpanda \
    rpk topic create "${topic}" \
      --partitions "${partitions}" \
      --replicas 1 \
    >/dev/null 2>&1 || true

  docker compose exec -T redpanda \
    rpk topic alter-config "${topic}" \
      --set "retention.ms=${retention_ms}" \
    >/dev/null

  docker compose exec -T redpanda \
    rpk topic alter-config "${topic}" \
      --set "cleanup.policy=delete" \
    >/dev/null
}

create_topic "${RAW_EVENTS_TOPIC}" 3 "${DEV_RETENTION_MS}"
create_topic "${DECISIONS_TOPIC}" 3 "${DEV_RETENTION_MS}"
create_topic "${DLQ_TOPIC}" 1 "${DLQ_RETENTION_MS}"
create_topic "${BASELINE_SIGNALS_TOPIC}" 1 "${DEV_RETENTION_MS}"

echo "topics are ready:"
docker compose exec -T redpanda rpk topic list