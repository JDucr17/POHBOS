<script lang="ts">
  import { scale } from "svelte/transition";
  import { cubicOut } from "svelte/easing";
  import type { PipelineNode } from "./pipelineConfig";

  interface Props {
    node: PipelineNode;
  }

  let { node }: Props = $props();

  let hovered = $state(false);
  const peekBelow = $derived(node.y < 45);

  interface Schema {
    title: string;
    fields: string[];
    payload: string[];
  }

  const SCHEMAS: Record<string, Schema> = {
    events: {
      title: "EventEnvelope",
      fields: [
        "event_id",
        "ingested_at",
        "schema_version",
        "source_id",
        "visitor_id",
      ],
      payload: [
        "source_id",
        "visitor_id",
        "event_time",
        "http_method",
        "uri",
        "status_code",
        "referrer_present",
        "user_agent",
        "bytes",
        "resource_type",
      ],
    },
    decisions: {
      title: "DecisionEnvelope",
      fields: [
        "decision_id",
        "schema_version",
        "source_id",
        "visitor_id",
        "event_id",
      ],
      payload: [
        "decided_at",
        "score_raw",
        "score_normalized",
        "risk_level",
        "action",
        "status",
        "policy_version",
        "baseline_run_id",
      ],
    },
  };

  const schema = $derived(SCHEMAS[node.id]);
</script>

<div
  class="group relative w-[7.5rem] cursor-default select-none"
  onmouseenter={() => (hovered = true)}
  onmouseleave={() => (hovered = false)}
  role="group"
>
  <div class="relative h-[2.6rem] w-[7.5rem]">
    <svg
      class="absolute inset-0 h-full w-full"
      viewBox="0 0 120 42"
      preserveAspectRatio="none"
    >
      <path
        d="M14 2 H119 L106 40 H1 Z"
        fill="var(--color-muted)"
        stroke="var(--color-border)"
        stroke-width="1.5"
      />
    </svg>
    <div class="absolute inset-0 flex items-center justify-center">
      <span
        class="font-mono text-[10px] tracking-wide text-muted-foreground lowercase"
      >
        {node.label}
      </span>
    </div>
  </div>

  {#if hovered && schema}
    <div
      class="absolute left-1/2 z-50 -translate-x-1/2 {peekBelow
        ? 'top-full pt-3'
        : 'bottom-full pb-3'}"
    >
      <div
        in:scale={{ duration: 160, start: 0.96, opacity: 0, easing: cubicOut }}
        out:scale={{ duration: 110, start: 0.98, opacity: 0, easing: cubicOut }}
        class="schema-panel w-56 rounded-md p-3 text-left {peekBelow
          ? 'origin-top'
          : 'origin-bottom'}"
      >
        <p class="schema-title mb-1.5 font-mono text-[11px] font-semibold">
          {schema.title}
        </p>
        <div
          class="font-mono text-[10px] leading-relaxed text-[color:var(--node-fg)]"
        >
          {#each schema.fields as f (f)}
            <div>{f}</div>
          {/each}
          <div class="schema-payload mt-1">payload</div>
          {#each schema.payload as p (p)}
            <div class="pl-4">{p}</div>
          {/each}
        </div>
      </div>
    </div>
  {/if}
</div>

<style>
  .schema-panel {
    background: var(--node-surface);
    border: 1px solid var(--node-border);
    box-shadow: 0 8px 20px color-mix(in oklch, var(--color-foreground) 30%, transparent);
  }
  .schema-title {
    color: var(--node-fg);
  }
  .schema-payload {
    color: color-mix(in oklch, var(--flow) 70%, var(--node-fg));
  }
</style>
