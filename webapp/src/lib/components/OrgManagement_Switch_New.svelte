<script lang="ts">
  export let roleName = '';
  export let options: any[] = [];
</script>

<div class="bg-surface p-6 rounded-xl border border-border shadow-sm">
  <h3 class="text-lg font-bold mb-4 text-text">{roleName} Management</h3>
  <p class="text-sm text-muted mb-6">Configure permissions and manage internal roles.</p>
  
  <div class="space-y-6">
    <div>
      <label for="org-select" class="block text-xs font-semibold uppercase text-muted mb-2">Selection</label>
      <select id="org-select" class="w-full p-3 border rounded-lg bg-surface-soft focus:ring-2 focus:ring-primary outline-none">
        {#each options as opt}
          <option value={opt.id}>{opt.name}</option>
        {/each}
      </select>
    </div>

    <div class="flex flex-wrap gap-2">
      <!-- Representation of permission tags -->
      <span class="px-3 py-1 bg-primary-100 text-primary-700 rounded-full text-xs font-medium">Read: All</span>
      <span class="px-3 py-1 bg-success-soft text-success rounded-full text-xs font-medium">Write: Allowed</span>
    </div>
  </div>
</div>
