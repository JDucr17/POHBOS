<script lang="ts">
  import type { Component } from "svelte";
  import {
    Cog,
    Info,
    LogIn,
    Target,
    Search,
    AudioWaveform,
    Inbox,
    Funnel,
  } from "@lucide/svelte";
  import { scale } from "svelte/transition";
  import { cubicOut } from "svelte/easing";
  import type { PipelineNode } from "./pipelineConfig";

  interface Props {
    node: PipelineNode;
    pulse?: number;
  }

  let { node, pulse = 0 }: Props = $props();

  let hovered = $state(false);
  let gearEl: HTMLSpanElement | undefined;

  const GLYPHS: Record<string, Component> = {
    login: LogIn,
    target: Target,
    search: Search,
    waveform: AudioWaveform,
    inbox: Inbox,
    funnel: Funnel,
  };

  const href = $derived(`/services/${node.slug}`);
  const peekBelow = $derived(node.y < 45);
  const Glyph = $derived((node.glyph && GLYPHS[node.glyph]) || Target);

  function prefersReducedMotion(): boolean {
    return (
      typeof window !== "undefined" &&
      window.matchMedia("(prefers-reduced-motion: reduce)").matches
    );
  }

  $effect(() => {
    const p = pulse;
    if (!gearEl || p === 0 || prefersReducedMotion()) return;
    gearEl.animate(
      [{ transform: "rotate(0deg)" }, { transform: "rotate(360deg)" }],
      { duration: 1400, easing: "linear" },
    );
  });
</script>

<div
  class="group relative w-[13rem]"
  onmouseenter={() => (hovered = true)}
  onmouseleave={() => (hovered = false)}
  onfocusin={() => (hovered = true)}
  onfocusout={() => (hovered = false)}
  role="group"
>
  <div class="node-card">
    <div class="node-inner">
      <div
        class="flex h-7 items-center justify-between border-b border-[color:var(--node-border)] px-3"
      >
        <span class="text-[color:var(--node-fg-muted)]" aria-hidden="true">
          <Info class="size-3" />
        </span>
        <span
          bind:this={gearEl}
          class="inline-flex text-[color:var(--node-fg-muted)]"
          aria-hidden="true"
        >
          <Cog class="size-3" />
        </span>
      </div>

      <div class="flex items-center gap-2.5 px-3.5 py-3">
        <div class="min-w-0 flex-1">
          <a
            href={href}
            class="block leading-tight font-semibold text-[color:var(--node-fg)] transition-opacity hover:opacity-80"
            style="font-size: 16px;"
          >
            {node.label}
          </a>
          <span
            class="mt-0.5 block font-mono text-[10px] text-[color:var(--node-fg-muted)]"
          >
            {node.role}
          </span>
        </div>

        <div
          class="flex size-10 shrink-0 items-center justify-center rounded-md border border-[color:var(--node-border)] text-[color:var(--node-fg)]"
        >
          <Glyph class="size-5" />
        </div>
      </div>
    </div>
  </div>

  {#if hovered}
    <div
      class="absolute left-1/2 z-50 -translate-x-1/2 {peekBelow
        ? 'top-full pt-3'
        : 'bottom-full pb-3'}"
    >
      <div
        in:scale={{ duration: 160, start: 0.96, opacity: 0, easing: cubicOut }}
        out:scale={{ duration: 110, start: 0.98, opacity: 0, easing: cubicOut }}
        class="w-60 border border-border bg-card p-3 text-left shadow-md {peekBelow
          ? 'origin-top'
          : 'origin-bottom'}"
      >
        <p
          class="mb-1 font-mono text-[10px] tracking-wider text-muted-foreground uppercase"
        >
          {node.label}
        </p>
        <p class="text-xs leading-relaxed text-foreground">{node.peek}</p>
        <a
          href={href}
          class="mt-2 inline-block font-mono text-[11px] font-medium text-primary hover:text-primary-hover"
        >
          Read more →
        </a>
      </div>
    </div>
  {/if}
</div>

<style>
  .node-card {
    --bevel: polygon(
      18px 0,
      calc(100% - 9px) 0,
      100% 9px,
      100% calc(100% - 18px),
      calc(100% - 18px) 100%,
      9px 100%,
      0 calc(100% - 9px),
      0 18px
    );
    clip-path: var(--bevel);
    padding: 1px;
    background: var(--node-border);
    filter: drop-shadow(0 1px 2px color-mix(in oklch, var(--color-foreground) 16%, transparent));
  }

  .node-inner {
    clip-path: var(--bevel);
    background: var(--node-surface-top);
  }
</style>
