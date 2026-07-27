export interface AnalyticsDashboardFilters {
  sourceId: string;
  startDate: string;
  endDate: string;
}

export interface AnalyticsSourceDateRange {
  sourceId: string;
  firstEventDate: string;
  lastEventDate: string;
}

export interface AnalyticsFilterOptions {
  sourceRanges: AnalyticsSourceDateRange[];
}

export interface DailyRiskSummary {
  sourceId: string;
  eventDate: string;
  requestCount: number;
  evaluatedWindowCount: number;
  scoredWindowCount: number;
  noBaselineWindowCount: number;
  insufficientHistoryWindowCount: number;
  lowRiskCount: number;
  mediumRiskCount: number;
  highRiskCount: number;
  criticalRiskCount: number;
  allowActionCount: number;
  logActionCount: number;
  challengeActionCount: number;
  blockActionCount: number;
  averageNormalizedScore: number | null;
  p90NormalizedScore: number | null;
  p99NormalizedScore: number | null;
  isPartialDay: boolean;
}

export interface BaselineUsageSegment {
  sourceId: string;
  eventDate: string;
  segmentNumber: number;
  baselineRunId: number;
  firstEventTime: string;
  lastEventTime: string;
  baselineUsageWindowCount: number;
  scoredWindowCount: number;
  insufficientHistoryWindowCount: number;
  baselineFitAt: string;
}

export interface AnalyticsDashboardData {
  daily: DailyRiskSummary[];
  baselineUsage: BaselineUsageSegment[];
}
