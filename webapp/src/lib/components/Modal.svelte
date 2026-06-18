<script lang="ts">
  let { open = false, title = '', onclose, children }: { open?: boolean; title?: string; onclose?: () => void; children?: any } = $props();

  function handleBackdropClick(e: MouseEvent) {
    if (e.target === e.currentTarget && onclose) onclose();
  }

  function handleKeydown(e: KeyboardEvent) {
    if (e.key === 'Escape' && onclose) onclose();
  }
</script>

<svelte:window onkeydown={handleKeydown} />

{#if open}
  <div class="fixed inset-0 z-50 flex items-center justify-center bg-neutral-950/40" onclick={handleBackdropClick}>
    <div class="bg-surface rounded-xl shadow-xl w-full max-w-md mx-4 p-6" role="dialog" aria-modal="true">
      {#if title}
        <h2 class="text-lg font-semibold text-text mb-4">{title}</h2>
      {/if}
      {@render children?.()}
    </div>
  </div>
{/if}
