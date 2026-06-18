<script lang="ts">
  import { onMount } from 'svelte';
  import { listTenants, createTenant, deleteTenant } from '$lib/services/api';
  import type { Tenant } from '$lib/types';
  import Modal from '../Modal.svelte';

  let { slug }: { slug: string } = $props();

  let tenants: Tenant[] = $state([]);
  let loading = $state(true);
  let error = $state('');

  let showModal = $state(false);
  let newName = $state('');
  let saving = $state(false);
  let modalError = $state('');

  let deleteConfirm = $state<string | null>(null);

  async function load() {
    try {
      tenants = await listTenants(slug);
    } catch (e) {
      error = 'Failed to load tenants.';
    } finally {
      loading = false;
    }
  }

  onMount(load);

  async function handleCreate() {
    if (!newName.trim()) return;
    saving = true;
    modalError = '';
    try {
      await createTenant(slug, newName.trim());
      showModal = false;
      newName = '';
      await load();
    } catch (e) {
      modalError = 'Failed to create tenant.';
    } finally {
      saving = false;
    }
  }

  async function handleDelete(id: string) {
    try {
      await deleteTenant(slug, id);
      deleteConfirm = null;
      await load();
    } catch (e) {
      error = 'Failed to delete tenant.';
    }
  }
</script>

<div class="flex items-center justify-between mb-6">
  <h2 class="text-xl font-bold text-text">Tenants</h2>
  <button onclick={() => { showModal = true; modalError = ''; newName = ''; }}
          class="bg-primary text-white px-4 py-2 rounded-lg hover:bg-primary-hover text-sm font-medium">
    + New Tenant
  </button>
</div>

{#if loading}
  <div class="bg-surface rounded-xl shadow-sm border border-border">
    <div class="p-6 space-y-4 animate-pulse">
      {#each Array(3) as _}
        <div class="h-12 bg-surface-soft rounded"></div>
      {/each}
    </div>
  </div>
{:else if error}
  <div class="bg-danger-soft border border-danger text-danger px-4 py-3 rounded">{error}</div>
{:else if tenants.length === 0}
  <div class="bg-surface rounded-xl shadow-sm border border-border p-12 text-center">
    <div class="text-muted text-5xl mb-4">🏢</div>
    <h3 class="text-lg font-semibold text-text mb-2">No tenants yet</h3>
    <p class="text-muted mb-6">Create your first tenant to organize your users.</p>
    <button onclick={() => { showModal = true; modalError = ''; newName = ''; }}
            class="bg-primary text-white px-6 py-3 rounded-lg hover:bg-primary-hover font-medium">
      + Create Tenant
    </button>
  </div>
{:else}
  <div class="bg-surface rounded-xl shadow-sm border border-border overflow-hidden">
    <table class="w-full text-sm">
      <thead class="bg-surface-soft text-muted uppercase text-xs">
        <tr>
          <th class="text-left px-6 py-3 font-medium">Name</th>
          <th class="text-left px-6 py-3 font-medium">ID</th>
          <th class="text-left px-6 py-3 font-medium">Created</th>
          <th class="text-right px-6 py-3 font-medium">Actions</th>
        </tr>
      </thead>
      <tbody class="divide-y divide-border">
        {#each tenants as tenant}
          <tr class="hover:bg-surface-soft">
            <td class="px-6 py-3 text-text font-medium">{tenant.name}</td>
            <td class="px-6 py-3 text-muted font-mono text-xs">{tenant.id.slice(0, 8)}...</td>
            <td class="px-6 py-3 text-muted">{new Date(tenant.created_at).toLocaleDateString()}</td>
            <td class="px-6 py-3 text-right">
              {#if deleteConfirm === tenant.id}
                <div class="flex items-center justify-end gap-2">
                  <span class="text-xs text-muted">Confirm?</span>
                  <button onclick={() => handleDelete(tenant.id)}
                          class="bg-danger text-white px-2 py-1 rounded text-xs">Delete</button>
                  <button onclick={() => { deleteConfirm = null; }}
                          class="bg-surface-soft text-muted px-2 py-1 rounded text-xs hover:bg-neutral-200 dark:hover:bg-neutral-700">Cancel</button>
                </div>
              {:else}
                <button onclick={() => { deleteConfirm = tenant.id; }}
                        class="text-danger text-xs font-medium">Delete</button>
              {/if}
            </td>
          </tr>
        {/each}
      </tbody>
    </table>
  </div>
{/if}

<Modal open={showModal} title="New Tenant" onclose={() => { showModal = false; }}>
  {#snippet children()}
    {#if modalError}
      <div class="bg-danger-soft border border-danger text-danger px-3 py-2 rounded mb-4 text-sm">{modalError}</div>
    {/if}
    <div class="mb-4">
      <label class="block text-sm font-medium text-muted mb-1">Tenant Name</label>
      <input bind:value={newName} type="text" placeholder="e.g. acme-corp"
             class="w-full p-3 border rounded-lg focus:ring-2 focus:ring-primary outline-none transition bg-surface text-text border-border" />
    </div>
    <div class="flex justify-end gap-3">
      <button onclick={() => { showModal = false; }}
              class="px-4 py-2 text-sm text-muted hover:text-text">Cancel</button>
      <button onclick={handleCreate} disabled={saving || !newName.trim()}
              class="bg-primary text-white px-4 py-2 rounded-lg text-sm font-medium hover:bg-primary-hover disabled:opacity-50">
        {saving ? 'Creating...' : 'Create'}
      </button>
    </div>
  {/snippet}
</Modal>
