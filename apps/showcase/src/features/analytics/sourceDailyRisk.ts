import type { AsyncDuckDBConnection } from "@duckdb/duckdb-wasm";
import { createBrowserDatabase, DuckDBDataProtocol } from "@/features/analytics/duckdb/database";
import type {
  AnalyticsDashboardData,
  AnalyticsDashboardFilters,
  AnalyticsFilterOptions,
  BaselineUsageSegment,
  DailyRiskSummary,
} from "@/features/analytics/types";

const artifactName = "source_daily_risk_summary.parquet";
const relationName = "source_daily_risk_summary";

interface QueryRow {
  toJSON(): unknown;
}

export interface AnalyticsDashboardSession {
  filterOptions: AnalyticsFilterOptions;
  query(filters: AnalyticsDashboardFilters): Promise<AnalyticsDashboardData>;
  close(): Promise<void>;
}

export type AnalyticsLoadStage = "duckdb" | "artifact";

function toRecords(rows: Iterable<QueryRow>): Record<string, unknown>[] {
  return Array.from(rows, (row) => row.toJSON() as Record<string, unknown>);
}

function requiredString(record: Record<string, unknown>, key: string): string {
  const value = record[key];
  if (typeof value !== "string") {
    throw new Error(`Expected ${key} to be a string.`);
  }
  return value;
}

function requiredNumber(record: Record<string, unknown>, key: string): number {
  const value = record[key];
  if (typeof value === "number") return value;
  if (typeof value === "bigint") return Number(value);
  throw new Error(`Expected ${key} to be numeric.`);
}

function nullableNumber(record: Record<string, unknown>, key: string): number | null {
  if (record[key] === null) return null;
  return requiredNumber(record, key);
}

function requiredBoolean(record: Record<string, unknown>, key: string): boolean {
  const value = record[key];
  if (typeof value !== "boolean") {
    throw new Error(`Expected ${key} to be boolean.`);
  }
  return value;
}

async function queryFilterOptions(
  connection: AsyncDuckDBConnection,
): Promise<AnalyticsFilterOptions> {
  const sourceResult = await connection.query(`
    select
      source_id,
      min(event_date)::varchar as first_event_date,
      max(event_date)::varchar as last_event_date
    from ${relationName}
    group by source_id
    order by source_id
  `);
  const sourceRanges = toRecords(sourceResult).map((record) => ({
    sourceId: requiredString(record, "source_id"),
    firstEventDate: requiredString(record, "first_event_date"),
    lastEventDate: requiredString(record, "last_event_date"),
  }));

  if (sourceRanges.length === 0) {
    throw new Error("The published analytics artifact contains no source-day data.");
  }

  return { sourceRanges };
}

async function queryDailySummaries(
  connection: AsyncDuckDBConnection,
  filters: AnalyticsDashboardFilters,
): Promise<DailyRiskSummary[]> {
  const statement = await connection.prepare(`
    select *
    from (
      select
        source_id,
        event_date::varchar as event_date,
        request_count,
        evaluated_window_count,
        scored_window_count,
        no_baseline_window_count,
        insufficient_history_window_count,
        low_risk_count,
        medium_risk_count,
        high_risk_count,
        critical_risk_count,
        allow_action_count,
        log_action_count,
        challenge_action_count,
        block_action_count,
        average_normalized_score,
        p90_normalized_score,
        p99_normalized_score,
        event_date = min(event_date) over ()
          or event_date = max(event_date) over () as is_partial_day
      from ${relationName}
      where source_id = ?
    ) as source_days
    where event_date::date between ?::date and ?::date
    order by event_date
  `);

  try {
    const result = await statement.query(filters.sourceId, filters.startDate, filters.endDate);
    return toRecords(result).map((record) => ({
      sourceId: requiredString(record, "source_id"),
      eventDate: requiredString(record, "event_date"),
      requestCount: requiredNumber(record, "request_count"),
      evaluatedWindowCount: requiredNumber(record, "evaluated_window_count"),
      scoredWindowCount: requiredNumber(record, "scored_window_count"),
      noBaselineWindowCount: requiredNumber(record, "no_baseline_window_count"),
      insufficientHistoryWindowCount: requiredNumber(record, "insufficient_history_window_count"),
      lowRiskCount: requiredNumber(record, "low_risk_count"),
      mediumRiskCount: requiredNumber(record, "medium_risk_count"),
      highRiskCount: requiredNumber(record, "high_risk_count"),
      criticalRiskCount: requiredNumber(record, "critical_risk_count"),
      allowActionCount: requiredNumber(record, "allow_action_count"),
      logActionCount: requiredNumber(record, "log_action_count"),
      challengeActionCount: requiredNumber(record, "challenge_action_count"),
      blockActionCount: requiredNumber(record, "block_action_count"),
      averageNormalizedScore: nullableNumber(record, "average_normalized_score"),
      p90NormalizedScore: nullableNumber(record, "p90_normalized_score"),
      p99NormalizedScore: nullableNumber(record, "p99_normalized_score"),
      isPartialDay: requiredBoolean(record, "is_partial_day"),
    }));
  } finally {
    await statement.close();
  }
}

async function queryBaselineUsage(
  connection: AsyncDuckDBConnection,
  filters: AnalyticsDashboardFilters,
): Promise<BaselineUsageSegment[]> {
  const statement = await connection.prepare(`
    select
      daily.source_id,
      daily.event_date::varchar as event_date,
      segment.segment_number as segment_number,
      segment.baseline_run_id as baseline_run_id,
      segment.first_event_time::varchar as first_event_time,
      segment.last_event_time::varchar as last_event_time,
      segment.baseline_usage_window_count as baseline_usage_window_count,
      segment.scored_window_count as scored_window_count,
      segment.insufficient_history_window_count as insufficient_history_window_count,
      segment.baseline_fit_at::varchar as baseline_fit_at
    from ${relationName} as daily,
    unnest(daily.baseline_usage_segments) as segments(segment)
    where daily.source_id = ?
      and daily.event_date between ?::date and ?::date
    order by daily.event_date, segment.segment_number
  `);

  try {
    const result = await statement.query(filters.sourceId, filters.startDate, filters.endDate);
    return toRecords(result).map((record) => ({
      sourceId: requiredString(record, "source_id"),
      eventDate: requiredString(record, "event_date"),
      segmentNumber: requiredNumber(record, "segment_number"),
      baselineRunId: requiredNumber(record, "baseline_run_id"),
      firstEventTime: requiredString(record, "first_event_time"),
      lastEventTime: requiredString(record, "last_event_time"),
      baselineUsageWindowCount: requiredNumber(record, "baseline_usage_window_count"),
      scoredWindowCount: requiredNumber(record, "scored_window_count"),
      insufficientHistoryWindowCount: requiredNumber(record, "insufficient_history_window_count"),
      baselineFitAt: requiredString(record, "baseline_fit_at"),
    }));
  } finally {
    await statement.close();
  }
}

export async function openSourceDailyRiskDashboard(
  artifactUrl: string,
  onLoadStage?: (stage: AnalyticsLoadStage) => void,
): Promise<AnalyticsDashboardSession> {
  onLoadStage?.("duckdb");
  const database = await createBrowserDatabase();
  let connection: AsyncDuckDBConnection | undefined;
  let closed = false;

  const close = async (): Promise<void> => {
    if (closed) return;
    closed = true;
    await connection?.close();
    await database.terminate();
  };

  try {
    onLoadStage?.("artifact");
    await database.registerFileURL(artifactName, artifactUrl, DuckDBDataProtocol.HTTP, false);
    connection = await database.connect();
    await connection.query("set allow_community_extensions = false");
    await connection.query("load parquet");
    await connection.query(
      `create view ${relationName} as select * from read_parquet('${artifactName}')`,
    );

    const filterOptions = await queryFilterOptions(connection);

    return {
      filterOptions,
      async query(filters: AnalyticsDashboardFilters): Promise<AnalyticsDashboardData> {
        if (closed || !connection) throw new Error("The analytics query session is closed.");

        const daily = await queryDailySummaries(connection, filters);
        const baselineUsage = await queryBaselineUsage(connection, filters);
        return { daily, baselineUsage };
      },
      close,
    };
  } catch (error) {
    await close();
    throw error;
  }
}
