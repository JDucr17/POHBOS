<script lang="ts">
  interface Metric {
    label: string;
    value: string;
    color?: string;
  }

  interface Props {
    date: string;
    metrics: Metric[];
    partialDay?: boolean;
  }

  let { date, metrics, partialDay = false }: Props = $props();
</script>

<div class="min-w-52 rounded-lg border border-border bg-card p-3 text-xs shadow-lg">
  <p class="font-semibold text-foreground">{date}</p>
  {#if partialDay}
    <p class="mt-1 max-w-60 text-[10px] leading-4 text-risk-medium">
      Partial day — available traffic does not cover all 24 hours.
    </p>
  {/if}
  <dl class="mt-2 grid gap-1.5">
    {#each metrics as metric (metric.label)}
      <div class="flex items-center justify-between gap-5">
        <dt class="inline-flex items-center gap-1.5 text-muted-foreground">
          {#if metric.color}
            <span class="size-2 rounded-sm" style={`background: ${metric.color}`} aria-hidden="true"
            ></span>
          {/if}
          {metric.label}
        </dt>
        <dd class="font-medium tabular-nums text-foreground">{metric.value}</dd>
      </div>
    {/each}
  </dl>
</div>
