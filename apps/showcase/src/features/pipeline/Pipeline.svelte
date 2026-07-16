<script lang="ts">
  import type { DecisionEnvelope } from "@/lib/types";
  import { createFakeFlowFeed } from "./flowFeed";
  import {
    CONNECTIONS,
    HOT_PATH_ORDER,
    NODES,
    type PipelineConnection,
  } from "./pipelineConfig";
  import ServiceNode from "./ServiceNode.svelte";
  import StorageNode from "./StorageNode.svelte";
  import TopicNode from "./TopicNode.svelte";

  const TIMING = {
    hotPulseMs: 1600,
    signalPulseMs: 2100,
    ambientSeconds: 2.6,
  };

  interface Pulse {
    id: number;
    path: "hot" | "signal";
    durMs: number;
  }

  let pulses = $state<Pulse[]>([]);
  let pulseSeq = 0;
  let decisionCount = 0;

  let nodePulse = $state<Record<string, number>>({});
  let nodePulseSeq = 0;
  let hitTimers: ReturnType<typeof setTimeout>[] = [];

  const HIT_SCHEDULE = [
    { id: "ingestor", delay: 0 },
    { id: "events-sink", delay: 480 },
    { id: "detector", delay: 800 },
    { id: "decisions-sink", delay: 1320 },
    { id: "postgres", delay: 1460 },
    { id: "query-api", delay: 1600 },
  ];

  function hitNode(id: string) {
    nodePulseSeq += 1;
    nodePulse = { ...nodePulse, [id]: nodePulseSeq };
  }

  function scheduleHits(steps: { id: string; delay: number }[]) {
    for (const step of steps) {
      hitTimers.push(setTimeout(() => hitNode(step.id), step.delay));
    }
  }

  function nodeById(id: string) {
    return NODES.find((n) => n.id === id)!;
  }

  function pathFor(conn: PipelineConnection): string {
    const a = nodeById(conn.from);
    const b = nodeById(conn.to);
    if (conn.route === "elbowV") {
      return `M ${a.x} ${a.y} L ${a.x} ${b.y} L ${b.x} ${b.y}`;
    }
    return `M ${a.x} ${a.y} L ${b.x} ${b.y}`;
  }

  const hotPathD = HOT_PATH_ORDER.map((id, i) => {
    const n = nodeById(id);
    return `${i === 0 ? "M" : "L"} ${n.x} ${n.y}`;
  }).join(" ");

  const signalPathD = pathFor(CONNECTIONS.find((c) => c.id === "c-bw-det")!);

  function prefersReducedMotion(): boolean {
    return (
      typeof window !== "undefined" &&
      window.matchMedia("(prefers-reduced-motion: reduce)").matches
    );
  }

  function addPulse(path: "hot" | "signal", durMs: number) {
    const id = ++pulseSeq;
    pulses = [...pulses, { id, path, durMs }];
    if (pulses.length > 16) pulses = pulses.slice(-16);
  }

  function removePulse(id: number) {
    pulses = pulses.filter((p) => p.id !== id);
  }

  function handleDecision(_decision: DecisionEnvelope) {
    if (prefersReducedMotion()) return;
    addPulse("hot", TIMING.hotPulseMs);
    scheduleHits(HIT_SCHEDULE);
    decisionCount += 1;
    if (decisionCount % 5 === 0) {
      addPulse("signal", TIMING.signalPulseMs);
      scheduleHits([
        { id: "baseline-worker", delay: 0 },
        { id: "detector", delay: TIMING.signalPulseMs },
      ]);
    }
  }

  $effect(() => {
    const feed = createFakeFlowFeed();
    const unsub = feed.subscribe(handleDecision);
    feed.start();
    return () => {
      unsub();
      feed.stop();
      hitTimers.forEach(clearTimeout);
      hitTimers = [];
    };
  });
</script>

<div class="w-full overflow-x-auto">
  <div
    class="pipeline-canvas relative min-h-[72vh] w-full min-w-[960px] overflow-hidden"
    style="--ambient-dur: {TIMING.ambientSeconds}s"
  >
  <div class="grid-layer absolute inset-0"></div>

  <svg
    class="pointer-events-none absolute inset-0 z-20 h-full w-full"
    viewBox="0 0 100 100"
    preserveAspectRatio="none"
    aria-hidden="true"
  >
    <g>
      {#each CONNECTIONS as conn (conn.id)}
        <path
          d={pathFor(conn)}
          fill="none"
          stroke="var(--primary)"
          stroke-width="1.25"
          stroke-linecap="round"
          opacity="0.32"
          vector-effect="non-scaling-stroke"
        />
      {/each}
    </g>

    {#each pulses as p (p.id)}
      <path
        d={p.path === "hot" ? hotPathD : signalPathD}
        class="packet"
        pathLength="100"
        fill="none"
        vector-effect="non-scaling-stroke"
        style="animation-duration: {p.durMs}ms"
        onanimationend={() => removePulse(p.id)}
      />
    {/each}
  </svg>

  <div class="pointer-events-none absolute inset-0 z-30">
    {#each NODES as node (node.id)}
      <div
        class="pointer-events-auto absolute -translate-x-1/2 -translate-y-1/2 hover:z-50 focus-within:z-50"
        style="left: {node.x}%; top: {node.y}%"
      >
        {#if node.type === "service"}
          <ServiceNode {node} pulse={nodePulse[node.id] ?? 0} />
        {:else if node.type === "topic"}
          <TopicNode {node} />
        {:else}
          <StorageNode {node} />
        {/if}
      </div>
    {/each}
    </div>
  </div>
</div>

<style>
  .pipeline-canvas {
    background-color: var(--canvas-bg);
  }

  .grid-layer {
    --grid-size: 34px;
    --grid-line-color: var(--color-border);
    --grid-dot-color: var(--color-muted-foreground);
    --grid-opacity: 0.5;
    background-image:
      linear-gradient(to right, var(--grid-line-color) 1px, transparent 1px),
      linear-gradient(to bottom, var(--grid-line-color) 1px, transparent 1px),
      radial-gradient(circle, var(--grid-dot-color) 1.5px, transparent 2px);
    background-size:
      var(--grid-size) var(--grid-size),
      var(--grid-size) var(--grid-size),
      var(--grid-size) var(--grid-size);
    background-position:
      0 0,
      0 0,
      calc(var(--grid-size) / 2) calc(var(--grid-size) / 2);
    opacity: var(--grid-opacity);
  }

  .packet {
    stroke: var(--primary);
    stroke-width: 3;
    stroke-linecap: round;
    stroke-dasharray: 1.5 100;
    animation-name: pulse-travel;
    animation-timing-function: linear;
    animation-fill-mode: forwards;
  }

  @keyframes pulse-travel {
    from {
      stroke-dashoffset: 104;
    }
    to {
      stroke-dashoffset: 4;
    }
  }

  @media (prefers-reduced-motion: reduce) {
    .packet {
      animation: none;
    }
  }
</style>
