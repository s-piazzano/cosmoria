<script lang="ts">
  import { onMount } from 'svelte';
  import { fetchOverview, updateProject } from '$lib/services/api';
  import type { OverviewData } from '$lib/types';

  let { slug }: { slug: string } = $props();

  let data: OverviewData | null = $state(null);
  let loading = $state(true);
  let error = $state('');
  let toggling = $state(false);
  let toggleError = $state('');

  onMount(async () => {
    try {
      data = await fetchOverview(slug);
    } catch (e) {
      error = 'Failed to load overview.';
    } finally {
      loading = false;
    }
  });

  async function toggleMultitenancy() {
    if (!data) return;
    toggling = true;
    toggleError = '';
    try {
      const updated = await updateProject(slug, { multitenancy_enabled: !data.multitenancy_enabled });
      data = { ...data, multitenancy_enabled: updated.multitenancy_enabled };
    } catch (e: any) {
      toggleError = e.message || 'Failed to toggle multitenancy.';
    } finally {
      toggling = false;
    }
  }
</script>

{#if loading}
  <div class="grid grid-cols-1 md:grid-cols-3 gap-6 mb-8">
    {#each Array(6) as _}
      <div class="bg-surface p-6 rounded-xl shadow-sm border border-border animate-pulse">
        <div class="h-4 bg-neutral-300 dark:bg-neutral-700 rounded w-1/2 mb-3"></div>
        <div class="h-8 bg-neutral-300 dark:bg-neutral-700 rounded w-1/3"></div>
      </div>
    {/each}
  </div>
{:else if error}
  <div class="bg-danger-soft border border-danger text-danger px-4 py-3 rounded">{error}</div>
{:else if data}
  <div class="grid grid-cols-1 md:grid-cols-3 gap-6 mb-8">
    <div class="bg-surface p-6 rounded-xl shadow-sm border border-border">
      <h3 class="text-muted text-sm font-semibold uppercase tracking-wider">Tenants</h3>
      <p class="text-4xl font-bold mt-2 text-text">{data.stats.tenants}</p>
    </div>
    <div class="bg-surface p-6 rounded-xl shadow-sm border border-border">
      <h3 class="text-muted text-sm font-semibold uppercase tracking-wider">Collections</h3>
      <p class="text-4xl font-bold mt-2 text-text">{data.stats.collections}</p>
    </div>
    <div class="bg-surface p-6 rounded-xl shadow-sm border border-border">
      <h3 class="text-muted text-sm font-semibold uppercase tracking-wider">Records</h3>
      <p class="text-4xl font-bold mt-2 text-text">{data.stats.records}</p>
    </div>
    <div class="bg-surface p-6 rounded-xl shadow-sm border border-border">
      <h3 class="text-muted text-sm font-semibold uppercase tracking-wider">Users</h3>
      <p class="text-4xl font-bold mt-2 text-text">{data.stats.users}</p>
    </div>
    <div class="bg-surface p-6 rounded-xl shadow-sm border border-border">
      <h3 class="text-muted text-sm font-semibold uppercase tracking-wider">Files</h3>
      <p class="text-4xl font-bold mt-2 text-text">{data.stats.files}</p>
    </div>
    <div class="bg-surface p-6 rounded-xl shadow-sm border border-border">
      <h3 class="text-muted text-sm font-semibold uppercase tracking-wider">Roles</h3>
      <p class="text-4xl font-bold mt-2 text-text">{data.stats.roles}</p>
    </div>
  </div>

  <div class="bg-surface rounded-xl shadow-sm border border-border p-6 mb-6">
    <div class="flex items-center justify-between">
      <div>
        <h3 class="text-sm font-semibold text-text">Multi-tenancy</h3>
        <p class="text-xs text-muted mt-0.5">When enabled, data is isolated per tenant</p>
      </div>
      <button onclick={toggleMultitenancy} disabled={toggling}
              role="switch" aria-checked={data.multitenancy_enabled}
              class="relative w-11 h-6 rounded-full transition cursor-pointer shrink-0 disabled:opacity-50 {data.multitenancy_enabled ? 'bg-primary' : 'bg-neutral-300 dark:bg-neutral-700'}">
        <span class="absolute top-0.5 left-0.5 w-5 h-5 bg-white rounded-full shadow transition {data.multitenancy_enabled ? 'translate-x-5' : ''}"></span>
      </button>
    </div>
    {#if toggleError}
      <div class="bg-danger-soft border border-danger text-danger px-3 py-2 rounded mt-3 text-sm">{toggleError}</div>
    {/if}
  </div>

  <section class="bg-surface rounded-xl shadow-sm border border-border">
    <div class="px-6 py-4 border-b border-border">
      <h2 class="text-lg font-semibold text-text">Recent Activity</h2>
    </div>
    {#if !data.recent_audit || data.recent_audit.length === 0}
      <div class="px-6 py-8 text-center text-muted">No recent activity recorded.</div>
    {:else}
      <table class="w-full text-sm">
        <thead class="bg-surface-soft text-muted uppercase text-xs">
          <tr>
            <th class="text-left px-6 py-3 font-medium">Action</th>
            <th class="text-left px-6 py-3 font-medium">Resource</th>
            <th class="text-left px-6 py-3 font-medium">Time</th>
          </tr>
        </thead>
        <tbody class="divide-y divide-border">
          {#each data.recent_audit as entry}
            <tr class="hover:bg-surface-soft">
              <td class="px-6 py-3 text-text font-medium">{entry.action}</td>
              <td class="px-6 py-3 text-muted">{entry.resource || '-'}</td>
              <td class="px-6 py-3 text-muted">{new Date(entry.created_at).toLocaleString()}</td>
            </tr>
          {/each}
        </tbody>
      </table>
    {/if}
  </section>
{/if}
