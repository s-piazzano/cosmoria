<script lang="ts">
  import type { Snippet } from 'svelte';

  let {
    children,
    class: cls = '',
  }: {
    children: Snippet<[{ open: boolean; toggle: (e: Event) => void; close: () => void }]>;
    class?: string;
  } = $props();

  let open = $state(false);
  let container: HTMLDivElement | undefined = $state();

  function toggle(e: Event) {
    e.stopPropagation();
    open = !open;
  }

  function close() {
    open = false;
  }

  function handleWindowClick(e: MouseEvent) {
    if (!open) return;
    if (container && !container.contains(e.target as Node)) {
      open = false;
    }
  }

  function handleKeydown(e: KeyboardEvent) {
    if (e.key === 'Escape') open = false;
  }
</script>

<svelte:window onclick={handleWindowClick} onkeydown={handleKeydown} />

<div bind:this={container} class="relative {cls}">
  {@render children({ open, toggle, close })}
</div>
