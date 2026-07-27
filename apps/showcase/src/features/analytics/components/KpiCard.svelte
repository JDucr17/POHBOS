<script lang="ts">
  import { formatInteger } from "@/features/analytics/format";

  type KpiTone = "primary" | "teal" | "green" | "challenge" | "block";

  interface Props {
    label: string;
    value: number;
    detail: string;
    values: number[];
    tone: KpiTone;
  }

  const toneClasses: Record<KpiTone, { accent: string; text: string; surface: string }> = {
    primary: {
      accent: "bg-primary",
      text: "text-primary",
      surface: "bg-primary/5",
    },
    teal: {
      accent: "bg-cold-path",
      text: "text-cold-path",
      surface: "bg-cold-path/5",
    },
    green: {
      accent: "bg-risk-low",
      text: "text-risk-low",
      surface: "bg-risk-low/5",
    },
    challenge: {
      accent: "bg-risk-medium",
      text: "text-risk-medium",
      surface: "bg-risk-medium/5",
    },
    block: {
      accent: "bg-risk-critical",
      text: "text-risk-critical",
      surface: "bg-risk-critical/5",
    },
  };

  let { label, value, detail, values, tone }: Props = $props();

  const classes = $derived(toneClasses[tone]);
  const sparkline = $derived.by(() => {
    if (values.length === 0) return "";
    if (values.length === 1) return "0,20 100,20";

    const minimum = Math.min(...values);
    const maximum = Math.max(...values);
    const spread = maximum - minimum || 1;

    return values
      .map((point, index) => {
        const x = (index / (values.length - 1)) * 100;
        const y = 35 - ((point - minimum) / spread) * 28;
        return `${x},${y}`;
      })
      .join(" ");
  });
</script>

<article
  class={`relative overflow-hidden rounded-xl border border-border bg-card p-4 shadow-sm ${classes.surface}`}
>
  <span class={`absolute inset-x-0 top-0 h-0.5 ${classes.accent}`} aria-hidden="true"></span>
  <div class="flex items-start justify-between gap-3">
    <div class="min-w-0">
      <p class="text-xs font-medium text-muted-foreground">{label}</p>
      <p class={`mt-2 text-2xl font-semibold tracking-tight ${classes.text}`}>
        {formatInteger(value)}
      </p>
    </div>
    <svg
      class={`mt-5 h-10 w-20 shrink-0 ${classes.text}`}
      viewBox="0 0 100 40"
      fill="none"
      aria-hidden="true"
    >
      <polyline
        points={sparkline}
        stroke="currentColor"
        stroke-width="2.5"
        stroke-linecap="round"
        stroke-linejoin="round"
      ></polyline>
    </svg>
  </div>
  <p class="mt-2 text-[11px] leading-4 text-muted-foreground">{detail}</p>
</article>
