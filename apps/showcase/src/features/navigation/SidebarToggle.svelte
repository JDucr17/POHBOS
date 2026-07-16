<script lang="ts">
  import { Menu, X } from "@lucide/svelte";

  let open = $state(false);

  $effect(() => {
    document
      .getElementById("sidebar-nav")
      ?.setAttribute("data-open", String(open));
  });
</script>

<button
  type="button"
  onclick={() => (open = !open)}
  aria-controls="sidebar-nav"
  aria-expanded={open}
  aria-label="Toggle navigation"
  class="text-muted-foreground transition-colors hover:text-foreground md:hidden"
>
  {#if open}
    <X class="size-5" />
  {:else}
    <Menu class="size-5" />
  {/if}
</button>
