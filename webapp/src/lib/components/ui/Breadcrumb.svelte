<script lang="ts">
  let {
    items,
  }: {
    items: { label: string; href?: string }[];
  } = $props();
</script>

<nav aria-label="Breadcrumb" class="flex items-center gap-2 min-w-0">
  <a href="/" class="text-xl font-bold text-primary whitespace-nowrap shrink-0">Cosmoria</a>
  {#each items as item}
    <span class="text-muted shrink-0">/</span>
    {#if item.href}
      <a href={item.href} class="text-sm text-primary hover:text-primary-hover shrink-0">{item.label}</a>
    {:else}
      <span class="text-sm text-text font-medium truncate">{item.label}</span>
    {/if}
  {/each}
</nav>
