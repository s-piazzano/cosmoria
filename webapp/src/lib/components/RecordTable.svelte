<script lang="ts">
  import { onMount } =_svelte;
  import type { Record } from '../types';

  export let collectionName: string = "Records";
  
  let items: Record[] = [];
  let isLoading = false;
  let searchQuery = '';
  let filterType = 'all';

  // Il backend fornirà la lista degli item filtrabili.
  // Per ora, gestiamo il filtro lato client sulla serie di dati scaricata.
  $: filteredItems = items.filter(item => {
    if (!searchQuery) return true;
    const term = searchQuery.toLowerCase();
    return Object.values(item.data).some(val => 
      String(val).toLowerCase().includes(term)
    );
  });

  onMount(() => {
    // Simulation of data fetch from backend
    setTimeout(() => {
      items = [
        { id: '1', collection_id: 'c1', data: { name: 'Item A', status: 'active' } },
        { id: '2', collection_id: 'c1', data: { name: 'Item B', status: 'pending' } },
      ];
    }, 500);
  });
</script>

<div class="space-y-4">
  <div class="flex justify-between items-center">
    <h2 class="text-xl font-bold">{collectionName}</h2>
    <button class="bg-blue-600 text-white px-4 py-1 rounded hover:bg-blue-700 transition">
      + Create New
    </button>
  </div>

  <div class="relative">
    <input 
      type="text" 
      placeholder="Search items..." 
      bind:value={searchQuery}
      class="w-full p-2 pl-10 border rounded-lg focus:ring-2 focus:ring-blue-500 outline-none" 
    />
    <span class="absolute left-3 top-2.5 opacity-40">🔍</span>
  </div>

  <div class="overflow-x-auto border rounded-lg shadow-sm">
    <table class="min-w-full divide-y divide-gray-200 text-sm text-left">
      <thead class="bg-gray-50">
        <tr>
          <th class="px-4 py-3 font-semibold">ID</th>
          <th class="px-4 py-3 font-semibold">Content Preview</th>
          <th class="px-4 py-3 font-semibold">Status</th>
          <th class="px-4 py-3 font-semibold">Actions</th>
        </tr>
      </thead>
      <tbody class="divide-y divide-gray-100 bg-white">
        {#each filteredItems as item}
          <tr>
            <td class="px-4 py-3 text-gray-500">{item.id}</td>
            <td class="px-4 py-3 font-medium">{JSON.stringify(item.data)}</td>
            <td class="px-4 py-3">
              <span class="px-2 py-1 rounded-full bg-green-100 text-green-700 text-xs">
                {item.data.status || 'N/A'}
              </span>
            </td>
            <td class="px-4 py-3 space-x-2">
              <button class="text-blue-600 hover:underline">Edit</button>
              <button class="text-red-600 hover:underline">Delete</button>
            </td>
          </tr>
        {:else}
          <tr>
            <td colspan="4" class="px-4 py-10 text-center text-gray-400">No records found matches your search.</td>
          </tr>
        {/each}
      </tbody>
    </table>
  </div>
</div>
