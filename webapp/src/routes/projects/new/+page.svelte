<script lang="ts">
  import { goto } from '$app/navigation';
  import { createProject } from '$lib/services/api';

  let name = '';
  let multitenancyEnabled = $state(false);
  let error = '';
  let loading = false;

  async function handleCreate() {
    if (!name) {
      error = 'Please enter a project name.';
      return;
    }

    loading = true;
    error = '';

    try {
      const project = await createProject(name, multitenancyEnabled);
      goto(`/${project.slug}`);
    } catch (e) {
      error = 'Failed to create project. Please try again.';
    } finally {
      loading = false;
    }
  }
</script>

<main class="max-w-lg mx-auto p-8 pt-20">
  <div class="bg-surface p-8 rounded-xl shadow-sm border border-border">
    <h2 class="text-2xl font-bold text-text mb-2">New Project</h2>
    <p class="text-muted mb-6">Create a new project to manage tenants, collections, and users.</p>

    {#if error}
      <div class="bg-danger-soft border border-danger text-danger px-3 py-2 rounded mb-4">{error}</div>
    {/if}

    <div class="mb-6">
      <label class="block text-sm font-medium text-muted mb-1">Project Name</label>
      <input bind:value={name} type="text" placeholder="e.g. My SaaS"
             class="w-full p-3 border rounded-lg focus:ring-2 focus:ring-primary outline-none transition bg-surface text-text border-border" />
    </div>

    <div class="mb-6 flex items-center justify-between">
      <div>
        <label class="text-sm font-medium text-text">Multi-tenancy</label>
        <p class="text-xs text-muted">Enable for projects that need per-tenant data isolation</p>
      </div>
      <button onclick={() => { multitenancyEnabled = !multitenancyEnabled; }}
              role="switch" aria-checked={multitenancyEnabled}
              class="relative w-11 h-6 rounded-full transition cursor-pointer shrink-0 {multitenancyEnabled ? 'bg-primary' : 'bg-neutral-300 dark:bg-neutral-700'}">
        <span class="absolute top-0.5 left-0.5 w-5 h-5 bg-white rounded-full shadow transition {multitenancyEnabled ? 'translate-x-5' : ''}"></span>
      </button>
    </div>

    <button onclick={handleCreate} disabled={loading}
            class="w-full bg-primary text-white font-medium py-3 px-4 rounded-lg hover:bg-primary-hover transition duration-200 disabled:opacity-50">
      {loading ? 'Creating...' : 'Create Project'}
    </button>
  </div>
</main>
