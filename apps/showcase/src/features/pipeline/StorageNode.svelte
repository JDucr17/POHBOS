<script lang="ts">
  import { siPostgresql } from "simple-icons";
  import type { PipelineNode } from "./pipelineConfig";

  interface Props {
    node: PipelineNode;
  }

  let { node }: Props = $props();
</script>

<div
  class="relative flex cursor-default flex-col items-center gap-1.5 select-none"
>
  <svg viewBox="0 0 72 84" class="pg-cyl w-[4.75rem]" aria-hidden="true">
    <defs>
      <linearGradient id="pg-body" x1="0" y1="0" x2="0" y2="1">
        <stop offset="0%" stop-color="var(--node-surface-top)" />
        <stop offset="100%" stop-color="var(--node-surface)" />
      </linearGradient>
    </defs>
    <path
      d="M8 14 V64 A28 9 0 0 0 64 64 V14"
      fill="url(#pg-body)"
      stroke="var(--node-border)"
      stroke-width="1.5"
    />
    <ellipse
      cx="36"
      cy="31"
      rx="28"
      ry="9"
      fill="none"
      stroke="var(--node-border)"
      stroke-width="1"
      opacity="0.35"
    />
    <ellipse
      cx="36"
      cy="47"
      rx="28"
      ry="9"
      fill="none"
      stroke="var(--node-border)"
      stroke-width="1"
      opacity="0.28"
    />
    <ellipse
      cx="36"
      cy="14"
      rx="28"
      ry="9"
      fill="var(--node-surface-top)"
      stroke="var(--node-border)"
      stroke-width="1.5"
    />
  </svg>

  <svg
    viewBox="0 0 24 24"
    class="size-6 text-primary"
    fill="currentColor"
    role="img"
    aria-label={node.label}
  >
    <path d={siPostgresql.path} />
  </svg>
</div>

<style>
  .pg-cyl {
    filter: drop-shadow(0 1px 2px color-mix(in oklch, var(--color-foreground) 16%, transparent));
  }
</style>
