<script lang="ts">
  import type { Role } from '../types';
  export let roleName = '';
  export let options: any[] = [];

  let selectedIndex = 0;
</script>

<div class="bg-white p-6 rounded-xl border border-gray-200 shadow-sm">
  <h3 class="text-lg font-bold mb-4 text-gray-800">{roleName} Management</h3>
  <p class="text-sm text-gray-500 mb-6">Configure permissions and manage internal roles.</p>
  
  <div class="space-y-6">
    <div>
      <label class="block text-xs font-semibold uppercase text-gray-400 mb-2">Selection</label>
      <select class="w-full p-3 border rounded-lg bg-gray-50 focus:ring-2 focus:ring-blue-600 outline-none">
        {#each options as opt}
          <option value={opt.id}>{opt.name}</option>
        {/each}
      </select>
    </div>

    <div class="flex flex-wrap gap-2">
      <!-- Representation of permission tags -->
      <span class="px-3 py-1 bg-blue-100 text-blue-700 rounded-full text-xs font-medium">Read: All</span>
      <span class="px-3 py-1 bg-green-100 text-green-700 rounded-full text-xs font-medium">Write: Allowed</span>
    </div>
  </div>
</div>
