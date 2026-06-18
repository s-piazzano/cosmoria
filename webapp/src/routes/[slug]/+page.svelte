<script lang="ts">
  import { page } from '$app/stores';
  import { onMount } from 'svelte';
  import ProjectSidebar from '$lib/components/sidebar/ProjectSidebar.svelte';
  import OverviewTab from '$lib/components/tabs/OverviewTab.svelte';
  import TenantsTab from '$lib/components/tabs/TenantsTab.svelte';
  import type { PageData } from './$types';

  const { data } = $props() as { data: PageData };
  const project = $derived(data.project);

  const activeTab = $derived($page.url.searchParams.get('tab') || 'overview');

  onMount(() => {
    localStorage.setItem('cosmoria_last_project', project.slug);
  });
</script>

<div class="flex h-full">
  <ProjectSidebar slug={project.slug} active={activeTab} multitenancyEnabled={project.multitenancy_enabled} />

  <main class="flex-1 p-8 overflow-y-auto">
    {#if activeTab === 'overview'}
      <OverviewTab slug={project.slug} />
    {:else if activeTab === 'tenants'}
      <TenantsTab slug={project.slug} />
    {:else if activeTab === 'collections'}
      <div class="text-center py-20 text-muted">Collections tab — coming soon</div>
    {:else if activeTab === 'roles'}
      <div class="text-center py-20 text-muted">Roles tab — coming soon</div>
    {:else if activeTab === 'users'}
      <div class="text-center py-20 text-muted">Users tab — coming soon</div>
    {:else if activeTab === 'files'}
      <div class="text-center py-20 text-muted">Files tab — coming soon</div>
    {:else if activeTab === 'audit'}
      <div class="text-center py-20 text-muted">Audit Log tab — coming soon</div>
    {:else if activeTab === 'api-keys'}
      <div class="text-center py-20 text-muted">API Keys tab — coming soon</div>
    {:else}
      <div class="text-center py-20 text-muted">Tab not found</div>
    {/if}
  </main>
</div>
