<script lang="ts">
  import DashboardPanel from "@/features/analytics/components/DashboardPanel.svelte";
  import {
    formatFullUtcDate,
    formatInteger,
    formatUtcTime,
    formatUtcTimestamp,
  } from "@/features/analytics/format";
  import type { BaselineUsageSegment } from "@/features/analytics/types";

  interface Props {
    data: BaselineUsageSegment[];
  }

  let { data }: Props = $props();
</script>

<DashboardPanel
  title="Baseline usage"
  description="Shows when each baseline was used and how many windows it evaluated."
>
  {#if data.length > 0}
    <div class="overflow-x-auto rounded-lg border border-border">
      <table class="w-full min-w-[58rem] border-collapse text-left text-xs">
        <thead class="bg-muted/70 text-muted-foreground">
          <tr>
            <th class="px-3 py-2.5 font-medium">Date</th>
            <th class="px-3 py-2.5 font-medium">Baseline run</th>
            <th class="px-3 py-2.5 font-medium">Active period</th>
            <th class="px-3 py-2.5 text-right font-medium">Evaluated</th>
            <th class="px-3 py-2.5 text-right font-medium">Scored</th>
            <th class="px-3 py-2.5 text-right font-medium">Insufficient history</th>
            <th class="px-3 py-2.5 font-medium">Fitted at</th>
          </tr>
        </thead>
        <tbody class="divide-y divide-border bg-card">
          {#each data as segment (`${segment.eventDate}-${segment.segmentNumber}`)}
            <tr
              class="transition-colors hover:bg-muted/40"
              title={`Baseline usage segment ${segment.segmentNumber}`}
            >
              <td class="px-3 py-3 whitespace-nowrap font-medium text-foreground">
                {formatFullUtcDate(segment.eventDate)}
              </td>
              <td class="px-3 py-3">
                <span
                  class="inline-flex rounded-full border border-primary/30 bg-primary/5 px-2 py-0.5 font-mono text-[11px] font-semibold tabular-nums text-primary"
                >
                  BL-{segment.baselineRunId}
                </span>
              </td>
              <td class="px-3 py-3 whitespace-nowrap text-muted-foreground">
                {formatUtcTime(segment.firstEventTime)}–{formatUtcTime(segment.lastEventTime)} UTC
              </td>
              <td class="px-3 py-3 text-right tabular-nums text-foreground">
                {formatInteger(segment.baselineUsageWindowCount)}
              </td>
              <td class="px-3 py-3 text-right tabular-nums text-risk-low">
                {formatInteger(segment.scoredWindowCount)}
              </td>
              <td class="px-3 py-3 text-right tabular-nums text-risk-medium">
                {formatInteger(segment.insufficientHistoryWindowCount)}
              </td>
              <td class="px-3 py-3 whitespace-nowrap text-muted-foreground">
                {formatUtcTimestamp(segment.baselineFitAt)}
              </td>
            </tr>
          {/each}
        </tbody>
      </table>
    </div>
  {:else}
    <p class="flex h-40 items-center justify-center text-sm text-muted-foreground">
      No baseline-backed decisions match this date range.
    </p>
  {/if}
</DashboardPanel>
