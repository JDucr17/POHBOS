<script lang="ts">
  import type { Snippet } from "svelte";

  interface Props {
    title: string;
    description?: string;
    class?: string;
    children: Snippet;
  }

  let { title, description, class: className = "", children }: Props = $props();
</script>

<section class={`rounded-xl border border-border bg-card p-5 shadow-sm ${className}`}>
  <header class="mb-4">
    <h2 class="text-sm font-semibold tracking-tight text-foreground">{title}</h2>
    {#if description}
      <p class="mt-1 text-xs leading-5 text-muted-foreground">{description}</p>
    {/if}
  </header>
  {@render children()}
</section>
