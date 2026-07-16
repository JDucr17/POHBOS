import type {
  Decision,
  DecisionAction,
  DecisionEnvelope,
  DecisionStatus,
  FlowFeed,
  FlowHandler,
  RiskLevel,
} from "@/lib/types";

export const FEED_TUNING = {
  emitIntervalMs: 2100,
  jitterMs: 1100,
};

const SOURCE_IDS = ["eclog-bookstore", "eclog-travel", "eclog-market"] as const;

function randRange(min: number, max: number): number {
  return min + Math.random() * (max - min);
}

function pick<T>(values: readonly T[]): T {
  return values[Math.floor(Math.random() * values.length)]!;
}

function id(prefix: string): string {
  return `${prefix}-${Math.random().toString(16).slice(2, 10)}`;
}

function riskBand(): RiskLevel {
  const r = Math.random();
  if (r < 0.56) return "low";
  if (r < 0.8) return "medium";
  if (r < 0.93) return "high";
  return "critical";
}

function scoreForBand(band: RiskLevel): number {
  switch (band) {
    case "low":
      return randRange(0.18, 0.7);
    case "medium":
      return randRange(0.7, 0.9);
    case "high":
      return randRange(0.9, 0.98);
    case "critical":
      return randRange(0.98, 0.999);
  }
}

function actionForBand(band: RiskLevel): DecisionAction {
  switch (band) {
    case "low":
      return "allow";
    case "medium":
      return "log";
    case "high":
      return "challenge";
    case "critical":
      return "block";
  }
}

function pickStatus(): DecisionStatus {
  const r = Math.random();
  if (r < 0.82) return "scored";
  if (r < 0.92) return "no_baseline";
  return "insufficient_history";
}

function generateDecision(): DecisionEnvelope {
  const status = pickStatus();
  const baselineRunId = 1000 + Math.floor(Math.random() * 40);

  let payload: Decision;
  if (status === "scored") {
    const band = riskBand();
    const normalized = scoreForBand(band);
    payload = {
      decided_at: new Date().toISOString(),
      score_raw: randRange(4, 42),
      score_normalized: normalized,
      risk_level: band,
      action: actionForBand(band),
      status,
      policy_version: "v1",
      baseline_run_id: baselineRunId,
    };
  } else if (status === "insufficient_history") {
    payload = {
      decided_at: new Date().toISOString(),
      action: "log",
      status,
      policy_version: "v1",
      baseline_run_id: baselineRunId,
    };
  } else {
    payload = {
      decided_at: new Date().toISOString(),
      action: "log",
      status,
      policy_version: "v1",
    };
  }

  return {
    decision_id: id("dec"),
    schema_version: "v1",
    source_id: pick(SOURCE_IDS),
    visitor_id: id("v"),
    event_id: id("evt"),
    payload,
  };
}

export function createFakeFlowFeed(): FlowFeed {
  const handlers = new Set<FlowHandler>();
  let timer: ReturnType<typeof setTimeout> | null = null;

  function scheduleNext() {
    const delay = FEED_TUNING.emitIntervalMs + Math.random() * FEED_TUNING.jitterMs;
    timer = setTimeout(() => {
      const decision = generateDecision();
      handlers.forEach((handler) => handler(decision));
      scheduleNext();
    }, delay);
  }

  return {
    subscribe(handler) {
      handlers.add(handler);
      return () => handlers.delete(handler);
    },
    start() {
      if (timer === null) scheduleNext();
    },
    stop() {
      if (timer !== null) {
        clearTimeout(timer);
        timer = null;
      }
    },
  };
}
