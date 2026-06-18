<script lang="ts">
  import { Sun, Moon } from '@lucide/svelte';

  let dark = $state(localStorage.getItem('theme') === 'dark');

  function toggle() {
    dark = !dark;
    if (dark) {
      document.documentElement.classList.add('dark');
      localStorage.setItem('theme', 'dark');
    } else {
      document.documentElement.classList.remove('dark');
      localStorage.setItem('theme', 'light');
    }
  }
</script>

<button onclick={toggle} class="p-2 rounded-lg hover:bg-surface-soft text-muted transition cursor-pointer"
        aria-label="Toggle theme">
  {#if dark}
    <Moon size={18} />
  {:else}
    <Sun size={18} />
  {/if}
</button>
