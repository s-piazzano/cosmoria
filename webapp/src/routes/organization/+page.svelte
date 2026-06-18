<script lang="ts">
  import OrgManagement_Switch_New from '../../lib/components/OrgManagement_Switch_New.svelte';
</script>

<main class="p-8">
  <div class="max-w-5xl">
    <header class="mb-8">
      <h1 class="text-3xl font-bold text-text">Organization Control</h1>
      <p class="text-muted">Manage users, roles, and access permissions within your organization.</p>
    </header>

  <div class="grid grid-cols-1 md:grid-cols-2 gap-8 items-start">
    <section class="bg-surface p-6 rounded-xl shadow-sm border border-border">
      <h2 class="text-lg font-bold mb-4 text-primary">Active Roles</h2>
      <div class="space-y-3">
        <div class="p-3 bg-surface-soft rounded flex justify-between items-center">
          <span>Editor Role</span>
          <span class="text-xs px-2 py-1 bg-success-soft text-success rounded">Active</span>
        </div>
        <div class="p-3 bg-surface-soft rounded flex justify-between items-center">
          <span>Viewer Role</span>
          <span class="text-xs px-2 py-1 bg-success-soft text-success rounded">Active</span>
        </div>
      </div>
    </section>

    <OrgManagement_Switch_New roleName="Detailed Roles" options={[{id: '1', name: 'Admin'}, {id: '2', name: 'Manager'}]} />
  </div>
</main>
