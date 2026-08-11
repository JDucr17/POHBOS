<script lang="ts">
  import * as Chart from "$lib/components/ui/chart/index.js";
  import { BarChart, Tooltip } from "layerchart";
  import ChartTooltipContent from "@/features/analytics/components/ChartTooltipContent.svelte";
  import DashboardPanel from "@/features/analytics/components/DashboardPanel.svelte";
  import {
    formatChartDate,
    formatCompact,
    formatFullUtcDate,
    formatInteger,
    formatPercent,
    responsiveDateTicks,
    toUtcDate,
  } from "@/features/analytics/format";
  import type { DailyRiskSummary } from "@/features/analytics/types";

  interface Props {
    data: DailyRiskSummary[];
  }

  let { data }: Props = $props();

  const chartData = $derived(
    data.map((day) => ({
      date: toUtcDate(day.eventDate),
      eventDate: day.eventDate,
      scored: day.scoredWindowCount,
      low: day.lowRiskCount,
      medium: day.mediumRiskCount,
      high: day.highRiskCount,
      critical: day.criticalRiskCount,
      isPartialDay: day.isPartialDay,
    })),
  );
  const xTicks = $derived(responsiveDateTicks(chartData.map((day) => day.date)));
  const riskTotal = $derived(data.reduce((total, day) => total + day.scoredWindowCount, 0));
  const chartConfig = {
    low: { label: "Low", color: "var(--risk-low)" },
    medium: { label: "Medium", color: "var(--risk-medium)" },
    high: { label: "High", color: "var(--risk-high)" },
    critical: { label: "Critical", color: "var(--risk-critical)" },
  } satisfies Chart.ChartConfig;

  function riskShare(count: number, scored: number): string {
    return formatPercent(scored === 0 ? 0 : (count / scored) * 100);
  }
</script>

<DashboardPanel
  title="Risk levels over time"
  description="Risk levels assigned to scored windows each day."
>
  <div
    class="mb-3 grid grid-cols-2 gap-x-3 gap-y-2 text-[11px] text-muted-foreground sm:grid-cols-4 xl:grid-cols-2 2xl:grid-cols-4"
  >
    {#each Object.entries(chartConfig) as [key, item] (key)}
      <span class="inline-flex min-w-0 items-center gap-1.5 whitespace-nowrap">
        <span class="size-2 rounded-sm" style={`background: ${item.color}`} aria-hidden="true"
        ></span>
        {item.label}
      </span>
    {/each}
  </div>
  {#if chartData.length > 0 && riskTotal > 0}
    <Chart.Container config={chartConfig} class="h-56 w-full aspect-auto">
      <BarChart
        data={chartData}
        x="date"
        yPadding={[0, 18]}
        padding={{ top: 4, right: 20, bottom: 24, left: 44 }}
        bandPadding={0.32}
        highlight={{ bar: { y: "scored" }, opacity: 0.06 }}
        tooltipContext={{ hideDelay: 100 }}
        seriesLayout="stack"
        series={[
          { key: "low", label: chartConfig.low.label, color: "var(--color-low)" },
          { key: "medium", label: chartConfig.medium.label, color: "var(--color-medium)" },
          { key: "high", label: chartConfig.high.label, color: "var(--color-high)" },
          {
            key: "critical",
            label: chartConfig.critical.label,
            color: "var(--color-critical)",
          },
        ]}
        props={{
          xAxis: { format: formatChartDate, ticks: xTicks, tickLength: 6 },
          yAxis: { format: formatCompact },
        }}
      >
        {#snippet tooltip({ context })}
          <Tooltip.Root
            {context}
            variant="none"
            anchor="bottom"
            xOffset={0}
            yOffset={12}
            contained="window"
            pointerEvents={true}
            classes={{ root: "select-text" }}
          >
            {#snippet children({ data: point })}
              <ChartTooltipContent
                date={formatFullUtcDate(point.eventDate)}
                partialDay={point.isPartialDay}
                metrics={[
                  { label: "Scored windows", value: formatInteger(point.scored) },
                  {
                    label: "Low risk",
                    value: `${formatInteger(point.low)} · ${riskShare(point.low, point.scored)}`,
                    color: "var(--risk-low)",
                  },
                  {
                    label: "Medium risk",
                    value: `${formatInteger(point.medium)} · ${riskShare(point.medium, point.scored)}`,
                    color: "var(--risk-medium)",
                  },
                  {
                    label: "High risk",
                    value: `${formatInteger(point.high)} · ${riskShare(point.high, point.scored)}`,
                    color: "var(--risk-high)",
                  },
                  {
                    label: "Critical risk",
                    value: `${formatInteger(point.critical)} · ${riskShare(point.critical, point.scored)}`,
                    color: "var(--risk-critical)",
                  },
                ]}
              />
            {/snippet}
          </Tooltip.Root>
        {/snippet}
      </BarChart>
    </Chart.Container>
  {:else}
    <p class="flex h-56 items-center justify-center text-sm text-muted-foreground">
      No scored windows are available for this date range.
    </p>
  {/if}
</DashboardPanel>
