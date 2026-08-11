<script lang="ts">
  import * as Chart from "$lib/components/ui/chart/index.js";
  import { scaleUtc } from "d3-scale";
  import { curveLinear } from "d3-shape";
  import { AreaChart, Tooltip } from "layerchart";
  import ChartTooltipContent from "@/features/analytics/components/ChartTooltipContent.svelte";
  import DashboardPanel from "@/features/analytics/components/DashboardPanel.svelte";
  import {
    formatChartDate,
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
      requests: day.requestCount,
      evaluated: day.evaluatedWindowCount,
      evaluationRate:
        day.requestCount === 0 ? 0 : (day.evaluatedWindowCount / day.requestCount) * 100,
      isPartialDay: day.isPartialDay,
    })),
  );
  const xTicks = $derived(responsiveDateTicks(chartData.map((day) => day.date)));
  const yMaximum = $derived.by(() => {
    const observedMaximum = Math.max(0, ...chartData.map((day) => day.evaluationRate));
    return Math.min(100, Math.max(5, Math.ceil((observedMaximum * 1.15) / 5) * 5));
  });
  const chartConfig = {
    evaluationRate: { label: "Evaluation rate", color: "var(--cold-path)" },
  } satisfies Chart.ChartConfig;
</script>

<DashboardPanel
  title="Evaluation rate"
  description="Share of requests that triggered a detector evaluation."
>
  {#if chartData.length > 0}
    <Chart.Container config={chartConfig} class="h-64 w-full aspect-auto">
      <AreaChart
        data={chartData}
        x="date"
        xScale={scaleUtc()}
        xPadding={[18, 18]}
        yDomain={[0, yMaximum]}
        yPadding={[0, 18]}
        padding={{ top: 4, right: 20, bottom: 24, left: 44 }}
        points
        series={[
          {
            key: "evaluationRate",
            label: chartConfig.evaluationRate.label,
            color: "var(--color-evaluationRate)",
          },
        ]}
        props={{
          area: {
            curve: curveLinear,
            fillOpacity: 0.2,
            line: { class: "stroke-2" },
          },
          points: { r: 3.5 },
          xAxis: { format: formatChartDate, ticks: xTicks, tickLength: 6 },
          yAxis: { format: formatPercent },
        }}
      >
        {#snippet tooltip({ context })}
          <Tooltip.Root {context} variant="none">
            {#snippet children({ data: point })}
              <ChartTooltipContent
                date={formatFullUtcDate(point.eventDate)}
                partialDay={point.isPartialDay}
                metrics={[
                  { label: "Requests", value: formatInteger(point.requests) },
                  { label: "Evaluated", value: formatInteger(point.evaluated) },
                  { label: "Evaluation rate", value: formatPercent(point.evaluationRate) },
                ]}
              />
            {/snippet}
          </Tooltip.Root>
        {/snippet}
      </AreaChart>
    </Chart.Container>
  {:else}
    <p class="flex h-64 items-center justify-center text-sm text-muted-foreground">
      No evaluation data matches this date range.
    </p>
  {/if}
</DashboardPanel>
