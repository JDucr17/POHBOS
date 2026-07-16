export type RiskLevel = "low" | "medium" | "high" | "critical";

export type DecisionStatus = "scored" | "no_baseline" | "insufficient_history";

export type DecisionAction = "allow" | "log" | "challenge" | "block";

export interface Decision {
  decided_at: string;
  score_raw?: number;
  score_normalized?: number;
  risk_level?: RiskLevel;
  action: DecisionAction;
  status: DecisionStatus;
  policy_version: string;
  baseline_run_id?: number;
}

export interface DecisionEnvelope {
  decision_id: string;
  schema_version: string;
  source_id: string;
  visitor_id: string;
  event_id: string;
  payload: Decision;
}

export type FlowHandler = (decision: DecisionEnvelope) => void;

export interface FlowFeed {
  subscribe(handler: FlowHandler): () => void;
  start(): void;
  stop(): void;
}
