<script lang="ts">
  import * as Chart from "$lib/components/ui/chart/index.js";
  import { scaleUtc } from "d3-scale";
  import { curveLinear } from "d3-shape";
  import { AreaChart, Tooltip } from "layerchart";
  import ChartTooltipContent from "@/features/analytics/components/ChartTooltipContent.svelte";
  import DashboardPanel from "@/features/analytics/components/DashboardPanel.svelte";
  import {
    formatChartDate,
    formatCompact,
    formatFullUtcDate,
    formatInteger,
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
      isPartialDay: day.isPartialDay,
    })),
  );
  const xTicks = $derived(responsiveDateTicks(chartData.map((day) => day.date)));
  const requestTotal = $derived(data.reduce((total, day) => total + day.requestCount, 0));
  const chartConfig = {
    requests: { label: "Requests", color: "var(--primary)" },
  } satisfies Chart.ChartConfig;
</script>

<DashboardPanel title="Daily requests" description="Requests received each day.">
  {#if chartData.length > 0 && requestTotal > 0}
    <Chart.Container config={chartConfig} class="h-64 w-full aspect-auto">
      <AreaChart
        data={chartData}
        x="date"
        xScale={scaleUtc()}
        xPadding={[18, 18]}
        yPadding={[0, 18]}
        padding={{ top: 4, right: 20, bottom: 24, left: 44 }}
        points
        series={[
          {
            key: "requests",
            label: chartConfig.requests.label,
            color: "var(--color-requests)",
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
          yAxis: { format: formatCompact },
        }}
      >
        {#snippet tooltip({ context })}
          <Tooltip.Root {context} variant="none">
            {#snippet children({ data: point })}
              <ChartTooltipContent
                date={formatFullUtcDate(point.eventDate)}
                partialDay={point.isPartialDay}
                metrics={[{ label: "Requests", value: formatInteger(point.requests) }]}
              />
            {/snippet}
          </Tooltip.Root>
        {/snippet}
      </AreaChart>
    </Chart.Container>
  {:else}
    <p class="flex h-64 items-center justify-center text-sm text-muted-foreground">
      No requests were received during this date range.
    </p>
  {/if}
</DashboardPanel>
