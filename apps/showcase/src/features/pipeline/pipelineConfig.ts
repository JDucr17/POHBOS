export type NodeType = "service" | "topic" | "storage";

export type NodeTone = "primary" | "cold" | "sink";

export interface PipelineNode {
  id: string;
  type: NodeType;
  label: string;
  x: number;
  y: number;
  slug?: string;
  tone?: NodeTone;
  typeTag?: string;
  glyph?: string;
  role?: string;
  peek?: string;
}

export type ConnectionKind = "hot" | "branch" | "write" | "read" | "signal";

export type ConnectionRoute = "straight" | "elbowV";

export interface PipelineConnection {
  id: string;
  from: string;
  to: string;
  kind: ConnectionKind;
  route?: ConnectionRoute;
}

export const NODES: PipelineNode[] = [
  {
    id: "ingestor",
    type: "service",
    label: "Ingestor",
    slug: "ingestor",
    tone: "primary",
    typeTag: "service",
    glyph: "login",
    role: "entrypoint",
    peek: "Entrypoint of the system. Receives events over HTTP and publishes an envelope — the raw event plus ingestion metadata — to the events topic.",
    x: 12,
    y: 18,
  },
  {
    id: "events",
    type: "topic",
    label: "events topic",
    typeTag: "kafka",
    x: 29,
    y: 18,
  },
  {
    id: "detector",
    type: "service",
    label: "Detector",
    slug: "detector",
    tone: "primary",
    typeTag: "service",
    glyph: "target",
    role: "hot path",
    peek: "Hot-path service. Maintains each visitor's live trailing window, scores it against the source's baseline through a cadence gate, applies the active policy, and publishes a decision.",
    x: 50,
    y: 18,
  },
  {
    id: "decisions",
    type: "topic",
    label: "decisions topic",
    typeTag: "kafka",
    x: 71,
    y: 18,
  },
  {
    id: "query-api",
    type: "service",
    label: "Query API",
    slug: "query-api",
    tone: "primary",
    typeTag: "service",
    glyph: "search",
    role: "read layer",
    peek: "Read layer. Consumes decisions into fast in-memory views and serves them to clients over HTTP and a live SSE stream.",
    x: 88,
    y: 18,
  },
  {
    id: "events-sink",
    type: "service",
    label: "Events Sink",
    slug: "sink",
    tone: "sink",
    typeTag: "service",
    glyph: "inbox",
    role: "events → store",
    peek: "Projects consumed records into durable Postgres storage in idempotent batches, advancing offsets only once a batch is persisted or dead-lettered. This instance projects events.",
    x: 29,
    y: 52,
  },
  {
    id: "decisions-sink",
    type: "service",
    label: "Decisions Sink",
    slug: "sink",
    tone: "sink",
    typeTag: "service",
    glyph: "funnel",
    role: "decisions → store",
    peek: "Projects consumed records into durable Postgres storage in idempotent batches, advancing offsets only once a batch is persisted or dead-lettered. This instance projects decisions.",
    x: 71,
    y: 52,
  },
  {
    id: "postgres",
    type: "storage",
    label: "Postgres",
    typeTag: "storage",
    x: 50,
    y: 80,
  },
  {
    id: "baseline-worker",
    type: "service",
    label: "Baseline Worker",
    slug: "baseline-worker",
    tone: "cold",
    typeTag: "cold-path",
    glyph: "waveform",
    role: "fits baselines",
    peek: "Cold-path service. Reads a source's event history from Postgres, fits a per-source HBOS baseline, and signals the detector that a new baseline is available.",
    x: 50,
    y: 44,
  },
];

export const CONNECTIONS: PipelineConnection[] = [
  { id: "c-ing-evt", from: "ingestor", to: "events", kind: "hot" },
  { id: "c-evt-det", from: "events", to: "detector", kind: "hot" },
  { id: "c-det-dec", from: "detector", to: "decisions", kind: "hot" },
  { id: "c-dec-api", from: "decisions", to: "query-api", kind: "hot" },
  { id: "c-evt-esink", from: "events", to: "events-sink", kind: "branch" },
  { id: "c-dec-dsink", from: "decisions", to: "decisions-sink", kind: "branch" },
  { id: "c-esink-pg", from: "events-sink", to: "postgres", kind: "write", route: "elbowV" },
  { id: "c-dsink-pg", from: "decisions-sink", to: "postgres", kind: "write", route: "elbowV" },
  { id: "c-pg-api", from: "postgres", to: "query-api", kind: "read" },
  { id: "c-pg-bw", from: "postgres", to: "baseline-worker", kind: "read" },
  { id: "c-bw-det", from: "baseline-worker", to: "detector", kind: "signal" },
];

export const HOT_PATH_ORDER = [
  "ingestor",
  "events",
  "detector",
  "decisions",
  "query-api",
];
