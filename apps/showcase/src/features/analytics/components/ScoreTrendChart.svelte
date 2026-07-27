<script lang="ts">
  import * as Chart from "$lib/components/ui/chart/index.js";
  import { scaleUtc } from "d3-scale";
  import { curveLinear } from "d3-shape";
  import { LineChart, Tooltip } from "layerchart";
  import ChartTooltipContent from "@/features/analytics/components/ChartTooltipContent.svelte";
  import DashboardPanel from "@/features/analytics/components/DashboardPanel.svelte";
  import {
    formatChartDate,
    formatFullUtcDate,
    formatInteger,
    formatScore,
    toUtcDate,
  } from "@/features/analytics/format";
  import type { DailyRiskSummary } from "@/features/analytics/types";

  interface Props {
    data: DailyRiskSummary[];
  }

  let { data }: Props = $props();

  const chartData = $derived(
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
  const chartConfig = {
    average: { label: "Average", color: "var(--risk-low)" },
    p90: { label: "P90", color: "var(--primary)" },
    p99: { label: "P99", color: "var(--sink)" },
  } satisfies Chart.ChartConfig;
</script>

<DashboardPanel
  title="Score trend"
  description="Daily average, p90, and p99 anomaly scores for scored windows."
>
  <div class="mb-3 flex flex-wrap gap-x-4 gap-y-2 text-[11px] text-muted-foreground">
    {#each Object.entries(chartConfig) as [key, item] (key)}
      <span class="inline-flex items-center gap-1.5">
        <span class="h-0.5 w-3 rounded-full" style={`background: ${item.color}`} aria-hidden="true"
        ></span>
        {item.label}
      </span>
    {/each}
  </div>
  {#if chartData.length > 0}
    <Chart.Container config={chartConfig} class="h-60 w-full aspect-auto">
      <LineChart
        data={chartData}
        x="date"
        xScale={scaleUtc()}
        yDomain={[0, 1]}
        yPadding={[0, 18]}
        points
        series={[
          { key: "average", label: chartConfig.average.label, color: "var(--color-average)" },
          { key: "p90", label: chartConfig.p90.label, color: "var(--color-p90)" },
          { key: "p99", label: chartConfig.p99.label, color: "var(--color-p99)" },
        ]}
        props={{
          spline: { curve: curveLinear, strokeWidth: 2 },
          points: { r: 3.5 },
          xAxis: { format: formatChartDate, ticks: chartData.length },
          yAxis: { format: formatScore, ticks: 5 },
        }}
      >
        {#snippet tooltip({ context })}
          <Tooltip.Root {context} variant="none">
            {#snippet children({ data: point })}
              <ChartTooltipContent
                date={formatFullUtcDate(point.eventDate)}
                partialDay={point.isPartialDay}
                metrics={[
                  {
                    label: "Average score",
                    value: formatScore(point.average),
                    color: "var(--risk-low)",
                  },
                  { label: "P90 score", value: formatScore(point.p90), color: "var(--primary)" },
                  { label: "P99 score", value: formatScore(point.p99), color: "var(--sink)" },
                  { label: "Scored windows", value: formatInteger(point.scored) },
                ]}
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
