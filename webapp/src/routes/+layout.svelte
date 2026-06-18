<script lang="ts">
  import { goto } from '$app/navigation';
  import { page } from '$app/stores';
  import '../app.css';

  let ready = false;

  $: path = $page.url.pathname;

  $: if (path) {
    checkAuth(path);
  }

  async function checkAuth(p: string) {
    if (p === '/setup') {
      try {
        const res = await fetch('/api/admin/setup/status');
        const data = await res.json();
        if (!data.needs_setup) {
          goto('/login');
          return;
        }
      } catch (e) {
        console.error('Auth guard: setup status check failed', e);
      }
      ready = true;
      return;
    }

    if (p === '/login') {
      try {
        const res = await fetch('/api/admin/setup/status');
        const data = await res.json();
        if (data.needs_setup) {
          goto('/setup');
          return;
        }
      } catch (e) {
        console.error('Auth guard: setup status check failed', e);
      }
      ready = true;
      return;
    }

    const token = localStorage.getItem('cosmoria_token');
    if (!token) {
      try {
        const res = await fetch('/api/admin/setup/status');
        const data = await res.json();
        if (data.needs_setup) {
          goto('/setup');
          return;
        }
      } catch (e) {
        console.error('Auth guard: setup status check failed', e);
      }
      goto('/login');
      return;
    }

    ready = true;
  }
</script>

{#if ready}
  <slot />
{/if}
