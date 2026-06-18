<script lang="ts">
  import { goto } from '$app/navigation';
  import { Plus } from '@lucide/svelte';
  import { fetchProjects } from '$lib/services/api';
  import type { ProjectWithRole } from '$lib/types';

  let projects: ProjectWithRole[] = $state([]);
  let loading = $state(true);
  let error = '';

  $effect(() => {
    fetchProjects()
      .then((p) => { projects = p; loading = false; })
      .catch((err) => { error = err.message; loading = false; });
  });
</script>

<main class="p-8">
  <div class="flex items-center justify-between mb-8">
    <div>
      <h1 class="text-3xl font-bold text-text">Projects</h1>
      <p class="text-muted mt-1">Manage your projects</p>
    </div>
    <button onclick={() => goto('/projects/new')}
            class="inline-flex items-center gap-2 bg-primary text-white px-5 py-2.5 rounded-lg hover:bg-primary-hover transition text-sm font-medium cursor-pointer">
      <Plus size={18} />
      New Project
    </button>
  </div>

  {#if loading}
    <div class="space-y-3">
      {#each [1, 2, 3] as _}
        <div class="h-16 bg-surface-soft rounded-xl animate-pulse" />
      {/each}
    </div>
  {:else if error}
    <div class="bg-danger-soft border border-danger text-danger px-4 py-3 rounded-lg">{error}</div>
  {:else if projects.length === 0}
    <div class="flex flex-col items-center justify-center py-24 text-center">
      <div class="text-5xl mb-4 text-muted opacity-30">📦</div>
      <h2 class="text-xl font-semibold text-text mb-2">No projects yet</h2>
      <p class="text-muted mb-6">Create your first project to get started.</p>
      <button onclick={() => goto('/projects/new')}
              class="bg-primary text-white px-6 py-2.5 rounded-lg hover:bg-primary-hover transition text-sm font-medium cursor-pointer">
        Create Project
      </button>
    </div>
  {:else}
    <div class="bg-surface border border-border rounded-xl overflow-hidden">
      <table class="w-full text-sm">
        <thead>
          <tr class="border-b border-border text-left text-muted text-xs uppercase tracking-wider">
            <th class="px-5 py-3 font-medium">Name</th>
            <th class="px-5 py-3 font-medium">Slug</th>
            <th class="px-5 py-3 font-medium">Role</th>
            <th class="px-5 py-3 font-medium">Multi-tenancy</th>
            <th class="px-5 py-3 font-medium">Created</th>
            <th class="px-5 py-3 w-10"></th>
          </tr>
        </thead>
        <tbody>
          {#each projects as project}
            <tr onclick={() => goto(`/${project.slug}`)}
                class="border-b border-border last:border-0 hover:bg-surface-soft transition cursor-pointer">
              <td class="px-5 py-4 font-medium text-text">{project.name}</td>
              <td class="px-5 py-4 text-muted font-mono text-xs">{project.slug}</td>
              <td class="px-5 py-4">
                <span class="inline-block px-2 py-0.5 text-xs rounded bg-primary-500/10 text-primary font-medium">
                  {project.role}
                </span>
              </td>
              <td class="px-5 py-4">
                {#if project.multitenancy_enabled}
                  <span class="inline-block px-2 py-0.5 text-xs rounded bg-success-soft text-success font-medium">Enabled</span>
                {:else}
                  <span class="text-muted text-xs">—</span>
                {/if}
              </td>
              <td class="px-5 py-4 text-muted text-xs">{new Date(project.created_at).toLocaleDateString()}</td>
              <td class="px-5 py-4 text-muted text-right">
                <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="m9 18 6-6-6-6"/></svg>
              </td>
            </tr>
          {/each}
        </tbody>
      </table>
    </div>
  {/if}
</main>
