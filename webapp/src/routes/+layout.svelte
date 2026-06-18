<script lang="ts">
  import { goto } from '$app/navigation';
  import { page } from '$app/stores';
  import Sidebar from '$lib/components/sidebar/Sidebar.svelte';
  import '../app.css';

  let { children } = $props();

  let baseUrl = (import.meta.env.VITE_API_URL || '').replace(/\/$/, '');
  let ready = $state(false);

  $effect(() => {
    const path = $page.url.pathname;
    if (!path) return;
    checkAuth(path);
  });

  async function checkAuth(p: string) {
    if (p === '/setup') {
      try {
        const res = await fetch(baseUrl + '/api/admin/setup/status');
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
        const res = await fetch(baseUrl + '/api/admin/setup/status');
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
        const res = await fetch(baseUrl + '/api/admin/setup/status');
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

  let showSidebar = $derived(ready && $page.url.pathname.startsWith('/login') === false && $page.url.pathname.startsWith('/setup') === false);
</script>

{#if ready}
  {#if showSidebar}
    <div class="min-h-screen bg-background">
      <Sidebar />
      <div class="ml-16 min-h-screen">
        {@render children()}
      </div>
    </div>
  {:else}
    {@render children()}
  {/if}
{/if}
