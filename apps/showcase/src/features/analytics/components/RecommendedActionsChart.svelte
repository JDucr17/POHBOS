<script lang="ts">
  import DashboardPanel from "@/features/analytics/components/DashboardPanel.svelte";
  import { formatInteger, formatPercent } from "@/features/analytics/format";
  import type { DailyRiskSummary } from "@/features/analytics/types";

  interface Props {
    data: DailyRiskSummary[];
  }

  interface ActionSummary {
    key: string;
    label: string;
    value: number;
    color: string;
  }

  let { data }: Props = $props();
  let activeActionKey = $state<string | null>(null);

  const actionData = $derived<ActionSummary[]>([
    {
      key: "allow",
      label: "Allow",
      value: data.reduce((total, day) => total + day.allowActionCount, 0),
      color: "var(--risk-low)",
    },
    {
      key: "log",
      label: "Log",
      value: data.reduce((total, day) => total + day.logActionCount, 0),
      color: "var(--primary)",
    },
    {
      key: "challenge",
      label: "Challenge",
      value: data.reduce((total, day) => total + day.challengeActionCount, 0),
      color: "var(--risk-medium)",
    },
    {
      key: "block",
      label: "Block",
      value: data.reduce((total, day) => total + day.blockActionCount, 0),
      color: "var(--risk-critical)",
    },
  ]);
  const actionTotal = $derived(actionData.reduce((total, action) => total + action.value, 0));
  const activeAction = $derived(
    actionData.find((action) => action.key === activeActionKey) ?? null,
  );

  function actionShare(value: number): number {
    return actionTotal === 0 ? 0 : (value / actionTotal) * 100;
  }
</script>

<DashboardPanel title="Recommended actions" description="Actions recommended by the active policy.">
  {#if actionTotal > 0}
    <div class="space-y-5">
      <p class="text-xs text-muted-foreground">
        <span class="font-semibold tabular-nums text-foreground">{formatInteger(actionTotal)}</span>
        total actions
      </p>

      <div
        class="flex h-11 w-full overflow-hidden rounded-lg border border-border bg-muted"
        role="img"
        aria-label={`Recommended action distribution across ${formatInteger(actionTotal)} evaluated windows`}
      >
        {#each actionData as action (action.key)}
          {#if action.value > 0}
            <button
              type="button"
              class="relative min-w-1 transition-[filter,opacity] hover:brightness-110 focus-visible:z-10 focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-inset focus-visible:outline-none"
              class:opacity-60={activeActionKey !== null && activeActionKey !== action.key}
              style={`width: ${actionShare(action.value)}%; background: ${action.color}`}
              aria-label={`${action.label}: ${formatInteger(action.value)}, ${formatPercent(actionShare(action.value))}`}
              onpointerenter={() => (activeActionKey = action.key)}
              onpointerleave={() => (activeActionKey = null)}
              onfocus={() => (activeActionKey = action.key)}
              onblur={() => (activeActionKey = null)}
              onclick={() => (activeActionKey = action.key)}
            ></button>
          {/if}
        {/each}
      </div>

      <div class="min-h-5" aria-live="polite">
        {#if activeAction}
          <p class="text-xs text-foreground">
            <span class="font-semibold">{activeAction.label}</span>
            <span class="ml-2 tabular-nums text-muted-foreground">
              {formatInteger(activeAction.value)} · {formatPercent(actionShare(activeAction.value))}
            </span>
          </p>
        {:else}
          <p class="text-xs text-muted-foreground">Hover, focus, or tap a segment for details.</p>
        {/if}
      </div>

      <ul class="grid gap-3 sm:grid-cols-2">
        {#each actionData as action (action.key)}
          <li class="rounded-lg border border-border bg-background/60 p-3">
            <div class="flex items-center justify-between gap-3">
              <span class="inline-flex items-center gap-2 text-xs text-muted-foreground">
                <span
                  class="size-2.5 rounded-sm"
                  style={`background: ${action.color}`}
                  aria-hidden="true"
                ></span>
                {action.label}
              </span>
              <span class="text-xs font-semibold tabular-nums text-foreground">
                {formatPercent(actionShare(action.value))}
              </span>
            </div>
            <p class="mt-2 text-lg font-semibold tabular-nums text-foreground">
              {formatInteger(action.value)}
            </p>
          </li>
        {/each}
      </ul>
    </div>
  {:else}
    <p class="flex h-60 items-center justify-center text-sm text-muted-foreground">
      No actions were recommended during this date range.
    </p>
  {/if}
</DashboardPanel>
