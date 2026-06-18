<script lang="ts">
  import { goto } from '$app/navigation';
  import { Orbit, ListTree,  ScrollText, Settings } from '@lucide/svelte';
  import { page } from '$app/stores';

  let active = $derived.by(() => {
    const p = $page.url.pathname;
    if (p === '/') return 'dashboard';
    if (p === '/projects' || p.startsWith('/projects/')) return 'projects';
    if (p === '/settings') return 'settings';
    if (p.startsWith('/logs')) return 'logs';
    return 'projects';
  });

  function goToProjects() {
    goto('/projects');
  }

  function goToLogs() {
    const last = localStorage.getItem('cosmoria_last_project');
    if (last) {
      goto(`/${last}?tab=audit`);
    } else {
      goto('/');
    }
  }
</script>

<aside class="w-16 h-screen bg-surface border-r border-border flex flex-col items-center py-4 gap-1 shrink-0 fixed left-0 top-0 z-30">
  <button onclick={() => goto('/')}
          title="Dashboard"
          class="w-10 h-10 rounded-full flex items-center justify-center transition {active === 'dashboard' ? 'bg-primary-500 dark:bg-primary-950/30 text-primary-50' : 'text-muted hover:bg-surface-soft hover:text-text'}">
    <Orbit size={20} />
  </button>

  <button onclick={goToProjects}
          title="Projects"
          class="w-10 h-10 rounded-full flex items-center justify-center transition {active === 'projects' ? 'bg-primary-500 dark:bg-primary-950/30 text-primary-50' : 'text-muted hover:bg-surface-soft hover:text-text'}">
    <ListTree size={20} />
  </button>

  <button onclick={goToLogs}
          title="Logs"
          class="w-10 h-10 rounded-full flex items-center justify-center transition {active === 'logs' ? 'bg-primary-500 dark:bg-primary-950/30 text-primary-50' : 'text-muted hover:bg-surface-soft hover:text-text'}">
    <ScrollText size={20} />
  </button>

  <div class="flex-1"></div>

  <button onclick={() => goto('/settings')}
          title="Settings"
          class="w-10 h-10 rounded-full flex items-center justify-center transition {active === 'settings' ? 'bg-primary-500 dark:bg-primary-950/30 text-primary' : 'text-muted hover:bg-surface-soft hover:text-text'}">
    <Settings size={20} />
  </button>
</aside>
