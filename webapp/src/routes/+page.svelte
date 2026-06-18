<script lang="ts">
  import { onMount } from 'svelte';
  import { goto } from '$app/navigation';
  import { fetchProjects } from '../lib/services/api';
  import type { ProjectWithRole } from '../lib/types';

  let projects: ProjectWithRole[] = [];
  let loading = true;
  let error = '';

  onMount(async () => {
    try {
      projects = await fetchProjects();
    } catch (e) {
      error = 'Failed to load projects.';
    } finally {
      loading = false;
    }
  });

  function goToProject(slug: string) {
    goto(`/${slug}`);
  }

  function goToNewProject() {
    goto('/projects/new');
  }
</script>

<main class="max-w-5xl mx-auto p-8">
    <div class="flex items-center justify-between mb-8">
      <h2 class="text-2xl font-bold text-text">My Projects</h2>
      <button onclick={goToNewProject}
              class="bg-primary text-white px-4 py-2 rounded-lg hover:bg-primary-hover text-sm font-medium">
        + New Project
      </button>
    </div>

    {#if loading}
      <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
        {#each Array(3) as _}
          <div class="bg-surface p-6 rounded-xl shadow-sm border border-border animate-pulse">
            <div class="h-5 bg-neutral-300 dark:bg-neutral-700 rounded w-3/4 mb-3"></div>
            <div class="h-4 bg-neutral-300 dark:bg-neutral-700 rounded w-1/2 mb-2"></div>
            <div class="h-3 bg-neutral-300 dark:bg-neutral-700 rounded w-1/3"></div>
          </div>
        {/each}
      </div>
    {:else if error}
      <div class="bg-danger-soft border border-danger text-danger px-4 py-3 rounded mb-4">{error}</div>
    {:else if projects.length === 0}
      <div class="text-center py-20">
        <div class="text-muted text-6xl mb-4">📦</div>
        <h3 class="text-xl font-semibold text-text mb-2">No projects yet</h3>
        <p class="text-muted mb-6">Create your first project to get started.</p>
        <button onclick={goToNewProject}
                class="bg-primary text-white px-6 py-3 rounded-lg hover:bg-primary-hover font-medium">
          + New Project
        </button>
      </div>
    {:else}
      <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
        {#each projects as project}
          <button onclick={() => goToProject(project.slug)}
                  class="bg-surface p-6 rounded-xl shadow-sm border border-border hover:shadow-md hover:border-primary transition text-left cursor-pointer">
            <h3 class="text-lg font-semibold text-text mb-1">{project.name}</h3>
            <p class="text-sm text-muted mb-3">/{project.slug}</p>
            <p class="text-xs text-muted">Created {new Date(project.created_at).toLocaleDateString()}</p>
          </button>
        {/each}
      </div>
    {/if}
  </main>
