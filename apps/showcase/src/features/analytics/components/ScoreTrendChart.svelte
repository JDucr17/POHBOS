<script lang="ts">
  import * as Chart from "$lib/components/ui/chart/index.js";
  import { scaleUtc } from "d3-scale";
  import { curveLinear } from "d3-shape";
  import { LineChart, Tooltip, type ChartState } from "layerchart";
  import { tick } from "svelte";
  import ChartTooltipContent from "@/features/analytics/components/ChartTooltipContent.svelte";
  import DashboardPanel from "@/features/analytics/components/DashboardPanel.svelte";
  import {
    formatChartDate,
    formatFullUtcDate,
    formatInteger,
    formatScore,
    responsiveDateTicks,
    toUtcDate,
  } from "@/features/analytics/format";
  import type { DailyRiskSummary } from "@/features/analytics/types";

  interface Props {
    data: DailyRiskSummary[];
  }

  interface ScorePoint {
    date: Date;
    eventDate: string;
    average: number;
    p90: number;
    p99: number;
    scored: number;
    isPartialDay: boolean;
  }

  interface TooltipMetric {
    label: string;
    value: string;
    color?: string;
  }

  let { data }: Props = $props();
  let activeSeries = $state<string | null>(null);
  let tooltipLocked = $state(false);
  let chartContext = $state<ChartState<ScorePoint>>();
  let clearActiveSeriesTimeout: ReturnType<typeof setTimeout> | undefined;

  const chartData = $derived<ScorePoint[]>(
    data
      .filter(
        (day) =>
          day.averageNormalizedScore !== null &&
          day.p90NormalizedScore !== null &&
          day.p99NormalizedScore !== null,
      )
      .map((day) => ({
        date: toUtcDate(day.eventDate),
        eventDate: day.eventDate,
        average: day.averageNormalizedScore as number,
        p90: day.p90NormalizedScore as number,
        p99: day.p99NormalizedScore as number,
        scored: day.scoredWindowCount,
        isPartialDay: day.isPartialDay,
      })),
  );
  const xTicks = $derived(responsiveDateTicks(chartData.map((day) => day.date)));
  const chartConfig = {
    average: { label: "Average", color: "var(--risk-low)" },
    p90: { label: "P90", color: "var(--primary)" },
    p99: { label: "P99", color: "var(--sink)" },
  } satisfies Chart.ChartConfig;

  function selectSeries(seriesKey?: string): void {
    clearTimeout(clearActiveSeriesTimeout);
    activeSeries = seriesKey ?? null;
  }

  function leaveSeriesPoint(): void {
    if (tooltipLocked) {
      if (chartContext) {
        chartContext.series.highlightKey = activeSeries;
      }
      return;
    }

    clearTimeout(clearActiveSeriesTimeout);
    clearActiveSeriesTimeout = setTimeout(() => {
      activeSeries = null;
    }, 100);
  }

  function pinTooltip(event: MouseEvent): void {
    event.stopPropagation();
    clearTimeout(clearActiveSeriesTimeout);
    tooltipLocked = true;
  }

  async function dismissTooltip(): Promise<void> {
    clearTimeout(clearActiveSeriesTimeout);
    tooltipLocked = false;
    activeSeries = null;

    if (chartContext) {
      chartContext.series.highlightKey = null;
      await tick();
      chartContext.tooltip.hide();
    }
  }

  function handleKeydown(event: KeyboardEvent): void {
    if (event.key === "Escape" && tooltipLocked) {
      void dismissTooltip();
    }
  }

  function tooltipMetrics(point: ScorePoint): TooltipMetric[] {
    const scoreMetrics = {
      average: {
        label: "Average score",
        value: formatScore(point.average),
        color: "var(--risk-low)",
      },
      p90: { label: "P90 score", value: formatScore(point.p90), color: "var(--primary)" },
      p99: { label: "P99 score", value: formatScore(point.p99), color: "var(--sink)" },
    };

    if (activeSeries === "average" || activeSeries === "p90" || activeSeries === "p99") {
      return [scoreMetrics[activeSeries]];
    }

    return [
      scoreMetrics.average,
      scoreMetrics.p90,
      scoreMetrics.p99,
      { label: "Scored windows", value: formatInteger(point.scored) },
    ];
  }
</script>

<svelte:window onkeydown={handleKeydown} />

<DashboardPanel
  title="Score trend"
  description="Daily average, p90, and p99 anomaly scores for scored windows."
>
  <div class="mb-3 grid grid-cols-3 gap-x-3 gap-y-2 text-[11px] text-muted-foreground">
    {#each Object.entries(chartConfig) as [key, item] (key)}
      <span class="inline-flex min-w-0 items-center gap-1.5 whitespace-nowrap">
        <span class="h-0.5 w-3 rounded-full" style={`background: ${item.color}`} aria-hidden="true"
        ></span>
        {item.label}
      </span>
    {/each}
  </div>
  {#if chartData.length > 0}
    <Chart.Container config={chartConfig} class="h-60 w-full aspect-auto">
      <LineChart
        bind:context={chartContext}
        data={chartData}
        x="date"
        xScale={scaleUtc()}
        xPadding={[18, 18]}
        yDomain={[0, 1]}
        yPadding={[0, 18]}
        padding={{ top: 4, right: 20, bottom: 24, left: 44 }}
        points
        highlight={{
          lines: true,
          points: true,
          onPointEnter: (_event, { point }) => selectSeries(point.seriesKey),
          onPointLeave: leaveSeriesPoint,
        }}
        onPointClick={pinTooltip}
        tooltipContext={{
          hideDelay: 100,
          locked: tooltipLocked,
          onclick: () => void dismissTooltip(),
        }}
        series={[
          { key: "average", label: chartConfig.average.label, color: "var(--color-average)" },
          { key: "p90", label: chartConfig.p90.label, color: "var(--color-p90)" },
          { key: "p99", label: chartConfig.p99.label, color: "var(--color-p99)" },
        ]}
        props={{
          spline: { curve: curveLinear, strokeWidth: 2 },
          points: { r: 3.5 },
          xAxis: { format: formatChartDate, ticks: xTicks, tickLength: 6 },
          yAxis: { format: formatScore, ticks: 5 },
        }}
      >
        {#snippet tooltip({ context })}
          <Tooltip.Root
            {context}
            variant="none"
            anchor={context.tooltip.x < context.containerWidth / 2 ? "right" : "left"}
            xOffset={14}
            yOffset={0}
            contained="window"
            pointerEvents={tooltipLocked}
            props={{
              root: {
                style: tooltipLocked ? "user-select: text;" : undefined,
                onclick: (event) => event.stopPropagation(),
              },
            }}
          >
            {#snippet children({ data: point })}
              <ChartTooltipContent
                date={formatFullUtcDate(point.eventDate)}
                partialDay={point.isPartialDay}
                metrics={tooltipMetrics(point)}
              />
            {/snippet}
          </Tooltip.Root>
        {/snippet}
      </LineChart>
    </Chart.Container>
  {:else}
    <p class="flex h-60 items-center justify-center text-sm text-muted-foreground">
      No scored windows are available for this date range.
    </p>
  {/if}
</DashboardPanel>
