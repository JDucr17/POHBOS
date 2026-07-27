<script lang="ts">
  import { RotateCcw } from "@lucide/svelte";
  import { onMount } from "svelte";
  import BaselineUsageTable from "@/features/analytics/components/BaselineUsageTable.svelte";
  import DailyRequestsChart from "@/features/analytics/components/DailyRequestsChart.svelte";
  import DetectorCoverageChart from "@/features/analytics/components/DetectorCoverageChart.svelte";
  import KpiCard from "@/features/analytics/components/KpiCard.svelte";
  import RecommendedActionsChart from "@/features/analytics/components/RecommendedActionsChart.svelte";
  import RiskCompositionChart from "@/features/analytics/components/RiskCompositionChart.svelte";
  import ScoreTrendChart from "@/features/analytics/components/ScoreTrendChart.svelte";
  import {
    openSourceDailyRiskDashboard,
    type AnalyticsLoadStage,
    type AnalyticsDashboardSession,
  } from "@/features/analytics/sourceDailyRisk";
  import type {
    AnalyticsDashboardData,
    AnalyticsDashboardFilters,
    AnalyticsFilterOptions,
  } from "@/features/analytics/types";

  type LoadState = "loading" | "ready" | "error";

  let loadState = $state<LoadState>("loading");
  let loadStage = $state<AnalyticsLoadStage>("duckdb");
  let isRefreshing = $state(false);
  let errorMessage = $state("");
  let options = $state.raw<AnalyticsFilterOptions | null>(null);
  let dashboardData = $state.raw<AnalyticsDashboardData | null>(null);
  let sourceId = $state("");
  let startDate = $state("");
  let endDate = $state("");

  let dashboardSession: AnalyticsDashboardSession | undefined;
  let componentActive = false;
  let querySequence = 0;

  const daily = $derived(dashboardData?.daily ?? []);
  const baselineUsage = $derived(dashboardData?.baselineUsage ?? []);
  const selectedSourceRange = $derived(
    options?.sourceRanges.find((range) => range.sourceId === sourceId) ?? null,
  );
  const totals = $derived.by(() => ({
    requests: daily.reduce((total, day) => total + day.requestCount, 0),
    evaluated: daily.reduce((total, day) => total + day.evaluatedWindowCount, 0),
    scored: daily.reduce((total, day) => total + day.scoredWindowCount, 0),
    challenges: daily.reduce((total, day) => total + day.challengeActionCount, 0),
    blocked: daily.reduce((total, day) => total + day.blockActionCount, 0),
  }));

  async function refresh(filters: AnalyticsDashboardFilters): Promise<void> {
    if (!dashboardSession) return;
    if (filters.startDate > filters.endDate) {
      errorMessage = "The start date must be on or before the end date.";
      return;
    }

    const sequence = ++querySequence;
    isRefreshing = true;
    errorMessage = "";

    try {
      const nextData = await dashboardSession.query(filters);
      if (!componentActive || sequence !== querySequence) return;
      dashboardData = nextData;
    } catch (error) {
      if (!componentActive || sequence !== querySequence) return;
      errorMessage =
        error instanceof Error ? error.message : "The selected analytics data could not be read.";
    } finally {
      if (componentActive && sequence === querySequence) isRefreshing = false;
    }
  }

  function currentFilters(): AnalyticsDashboardFilters {
    return { sourceId, startDate, endDate };
  }

  function handleSourceChange(event: Event): void {
    sourceId = (event.currentTarget as HTMLSelectElement).value;
    const sourceRange = options?.sourceRanges.find((range) => range.sourceId === sourceId);
    if (!sourceRange) return;
    startDate = sourceRange.firstEventDate;
    endDate = sourceRange.lastEventDate;
    void refresh(currentFilters());
  }

  function handleStartDateChange(event: Event): void {
    startDate = (event.currentTarget as HTMLInputElement).value;
    void refresh(currentFilters());
  }

  function handleEndDateChange(event: Event): void {
    endDate = (event.currentTarget as HTMLInputElement).value;
    void refresh(currentFilters());
  }

  function resetFilters(): void {
    if (!selectedSourceRange) return;
    startDate = selectedSourceRange.firstEventDate;
    endDate = selectedSourceRange.lastEventDate;
    void refresh(currentFilters());
  }

  async function initialize(): Promise<void> {
    loadState = "loading";
    errorMessage = "";

    const artifactUrl = new URL(
      `${import.meta.env.BASE_URL}data/source_daily_risk_summary.parquet`,
      window.location.origin,
    ).href;

    try {
      const session = await openSourceDailyRiskDashboard(artifactUrl, (stage) => {
        if (componentActive) loadStage = stage;
      });
      if (!componentActive) {
        await session.close();
        return;
      }

      dashboardSession = session;
      options = session.filterOptions;
      const initialRange =
        options.sourceRanges.find((range) => range.sourceId.toLowerCase().startsWith("eclog")) ??
        options.sourceRanges[0];
      if (!initialRange) throw new Error("The analytics artifact contains no source date ranges.");
      sourceId = initialRange.sourceId;
      startDate = initialRange.firstEventDate;
      endDate = initialRange.lastEventDate;
      dashboardData = await session.query(currentFilters());

      if (!componentActive) return;
      loadState = "ready";
    } catch (error) {
      if (!componentActive) return;
      await dashboardSession?.close();
      dashboardSession = undefined;
      errorMessage =
        error instanceof Error ? error.message : "The analytics dashboard could not be loaded.";
      loadState = "error";
    }
  }

  onMount(() => {
    componentActive = true;
    void initialize();

    return () => {
      componentActive = false;
      querySequence += 1;
      if (dashboardSession) void dashboardSession.close();
    };
  });
</script>

<section
  class="min-h-[38rem] border-b border-border bg-background"
  aria-label="Source daily analytics dashboard"
  aria-busy={loadState === "loading" || isRefreshing}
>
  <div class="w-full px-4 py-6 sm:px-6 lg:px-8 lg:py-8">
    <div>
      <div class="mb-6 max-w-3xl">
        <h1 class="text-2xl font-semibold tracking-tight text-foreground sm:text-3xl">
          Traffic Analytics
        </h1>
        <p class="mt-2 text-sm leading-6 text-muted-foreground">
          Explore anomaly scoring results from real website traffic captured in HTTP access logs,
          including total requests, risk levels, score trends, recommended actions, and baseline
          usage.
        </p>
      </div>

      {#if options}
        <div class="rounded-lg border border-primary/15 bg-primary/5 px-4 py-4 sm:px-[18px]">
          <div
            class="grid gap-4 md:grid-cols-[14rem_minmax(0,28rem)] md:items-end lg:flex lg:items-end"
          >
            <label class="grid gap-2 text-xs font-medium text-foreground lg:w-56 lg:shrink-0">
              <span class="pl-1">Source</span>
              <select
                value={sourceId}
                onchange={handleSourceChange}
                class="h-10 w-full rounded-md border border-input bg-background px-3 text-sm text-foreground shadow-xs focus-visible:ring-2 focus-visible:ring-ring focus-visible:outline-none"
              >
                {#each options.sourceRanges as sourceRange (sourceRange.sourceId)}
                  <option value={sourceRange.sourceId}>{sourceRange.sourceId}</option>
                {/each}
              </select>
            </label>

            <div class="grid gap-2 lg:w-[30rem] lg:shrink-0">
              <span
                id="analytics-date-range-label"
                class="pl-1 text-xs font-medium text-foreground"
              >
                Date range
              </span>
              <div
                class="grid grid-cols-1 items-center gap-3 sm:grid-cols-[minmax(0,1fr)_auto_minmax(0,1fr)]"
                role="group"
                aria-labelledby="analytics-date-range-label"
              >
                <label class="sr-only" for="analytics-start-date">Start date</label>
                <input
                  id="analytics-start-date"
                  type="date"
                  value={startDate}
                  min={selectedSourceRange?.firstEventDate}
                  max={endDate}
                  onchange={handleStartDateChange}
                  class="h-10 min-w-0 w-full rounded-md border border-input bg-background px-3 text-sm text-foreground shadow-xs focus-visible:ring-2 focus-visible:ring-ring focus-visible:outline-none"
                />
                <span class="justify-self-center text-xs text-muted-foreground" aria-hidden="true"
                  >to</span
                >
                <label class="sr-only" for="analytics-end-date">End date</label>
                <input
                  id="analytics-end-date"
                  type="date"
                  value={endDate}
                  min={startDate}
                  max={selectedSourceRange?.lastEventDate}
                  onchange={handleEndDateChange}
                  class="h-10 min-w-0 w-full rounded-md border border-input bg-background px-3 text-sm text-foreground shadow-xs focus-visible:ring-2 focus-visible:ring-ring focus-visible:outline-none"
                />
              </div>
            </div>

            <button
              type="button"
              onclick={resetFilters}
              class="inline-flex h-10 w-full items-center justify-center gap-2 self-end rounded-md border border-primary/30 bg-primary/10 px-4 text-sm font-medium text-primary shadow-xs transition-colors hover:border-primary/40 hover:bg-primary/15 focus-visible:ring-2 focus-visible:ring-ring focus-visible:outline-none md:col-span-2 md:w-[10.5rem] md:justify-self-start lg:ml-auto"
            >
              <RotateCcw class="size-3.5" aria-hidden="true" />
              Reset filters
            </button>
          </div>

          {#if isRefreshing}
            <p class="mt-3 text-xs text-primary" aria-live="polite">Updating dashboard…</p>
          {:else if errorMessage && loadState === "ready"}
            <p class="mt-3 text-xs text-destructive" role="alert">
              {errorMessage}
            </p>
          {/if}
        </div>
      {/if}
    </div>

    {#if loadState === "loading"}
      <div class="mt-7 space-y-5" aria-live="polite">
        <p class="text-sm text-muted-foreground">
          {loadStage === "duckdb" ? "Loading DuckDB-Wasm…" : "Loading the analytics artifact…"}
        </p>
        <div class="grid grid-cols-2 gap-3 lg:grid-cols-5">
          {#each Array(5) as _, index (index)}
            <div
              class="h-32 animate-pulse rounded-xl border border-border bg-card motion-reduce:animate-none"
            ></div>
          {/each}
        </div>
        <div class="grid gap-4 lg:grid-cols-3">
          {#each Array(3) as _, index (index)}
            <div
              class="h-80 animate-pulse rounded-xl border border-border bg-card motion-reduce:animate-none"
            ></div>
          {/each}
        </div>
      </div>
    {:else if loadState === "error"}
      <div
        class="mx-auto mt-7 max-w-xl rounded-xl border border-destructive/30 bg-card p-6 text-center"
        role="alert"
      >
        <p class="font-semibold text-foreground">Daily analytics could not be loaded.</p>
        <p class="mt-2 text-sm leading-6 text-muted-foreground">{errorMessage}</p>
        <button
          type="button"
          class="mt-4 inline-flex h-9 items-center rounded-md bg-primary px-4 text-sm font-medium text-primary-foreground transition-colors hover:bg-primary-hover focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 focus-visible:outline-none"
          onclick={() => void initialize()}
        >
          Try again
        </button>
      </div>
    {:else if options && dashboardData && daily.length === 0}
      <div class="mt-7 rounded-xl border border-border bg-card p-8 text-center shadow-sm">
        <p class="font-semibold text-foreground">No records match this date range.</p>
        <p class="mt-2 text-sm text-muted-foreground">
          Choose dates within the available range for {sourceId}, or reset the filters.
        </p>
      </div>
    {:else if options && dashboardData}
      <div class="mt-7 flex items-center gap-3">
        <h2 class="shrink-0 text-sm font-semibold text-foreground">Selected period</h2>
        <div class="h-px flex-1 bg-border/70" aria-hidden="true"></div>
      </div>

      <div class="mt-3 grid grid-cols-1 gap-4 sm:grid-cols-2 md:grid-cols-3 lg:grid-cols-5">
        <KpiCard
          label="Total requests"
          value={totals.requests}
          detail="Requests received during the selected period."
          values={daily.map((day) => day.requestCount)}
          tone="primary"
        />
        <KpiCard
          label="Evaluated windows"
          value={totals.evaluated}
          detail="Requests that triggered a detector evaluation."
          values={daily.map((day) => day.evaluatedWindowCount)}
          tone="teal"
        />
        <KpiCard
          label="Scored windows"
          value={totals.scored}
          detail="Evaluations that had enough recent activity to score."
          values={daily.map((day) => day.scoredWindowCount)}
          tone="green"
        />
        <KpiCard
          label="Challenges"
          value={totals.challenges}
          detail="Evaluations where the policy recommended a challenge."
          values={daily.map((day) => day.challengeActionCount)}
          tone="challenge"
        />
        <KpiCard
          label="Blocked"
          value={totals.blocked}
          detail="Evaluations where the policy recommended blocking."
          values={daily.map((day) => day.blockActionCount)}
          tone="block"
        />
      </div>

      <div class="mt-5 grid gap-4 xl:grid-cols-3">
        <DailyRequestsChart data={daily} />
        <DetectorCoverageChart data={daily} />
        <RiskCompositionChart data={daily} />
      </div>

      <div class="mt-4 grid gap-4 xl:grid-cols-[minmax(18rem,0.8fr)_minmax(0,1.2fr)]">
        <RecommendedActionsChart data={daily} />
        <ScoreTrendChart data={daily} />
      </div>

      <div class="mt-4">
        <BaselineUsageTable data={baselineUsage} />
      </div>
    {/if}
  </div>
</section>
