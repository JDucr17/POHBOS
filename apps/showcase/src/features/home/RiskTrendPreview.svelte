<script lang="ts">
  import * as Chart from "$lib/components/ui/chart/index.js";
  import { BarChart } from "layerchart";

  interface RiskTrendPoint {
    date: Date;
    low: number;
    medium: number;
    high: number;
    critical: number;
  }

  const chartData: RiskTrendPoint[] = [
    { date: new Date("2020-01-18"), low: 2279, medium: 579, high: 200, critical: 44 },
    { date: new Date("2020-01-19"), low: 4596, medium: 1721, high: 668, critical: 176 },
    { date: new Date("2020-01-20"), low: 6332, medium: 1858, high: 916, critical: 364 },
    { date: new Date("2020-01-21"), low: 6992, medium: 2119, high: 942, critical: 239 },
    { date: new Date("2020-01-22"), low: 8065, medium: 1669, high: 795, critical: 291 },
    { date: new Date("2020-01-23"), low: 3590, medium: 1121, high: 533, critical: 141 },
  ];

  const chartConfig = {
    low: { label: "Low", color: "var(--risk-low)" },
    medium: { label: "Medium", color: "var(--risk-medium)" },
    high: { label: "High", color: "var(--risk-high)" },
    critical: { label: "Critical", color: "var(--risk-critical)" },
  } satisfies Chart.ChartConfig;
</script>

<div role="img" aria-label="Daily risk-level composition for EClog traffic">
  <Chart.Container config={chartConfig} class="h-24 w-full aspect-auto">
    <BarChart
      data={chartData}
      x="date"
      yPadding={[0, 18]}
      bandPadding={0.28}
      axis={false}
      grid={false}
      rule={false}
      highlight={false}
      tooltipContext={false}
      pointerEvents={false}
      seriesLayout="stack"
      series={[
        {
          key: "low",
          label: chartConfig.low.label,
          color: "var(--color-low)",
        },
        {
          key: "medium",
          label: chartConfig.medium.label,
          color: "var(--color-medium)",
        },
        {
          key: "high",
          label: chartConfig.high.label,
          color: "var(--color-high)",
        },
        {
          key: "critical",
          label: chartConfig.critical.label,
          color: "var(--color-critical)",
        },
      ]}
    />
  </Chart.Container>
</div>
