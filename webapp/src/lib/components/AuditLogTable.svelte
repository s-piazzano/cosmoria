<script lang="ts">
  import type { AuditEntry } from '../types';
  
  export let history = [] as AuditEntry[];

  const getRiskColor = (level: string) => {
    if (level === 'high') return 'text-red-600 bg-red-100 border-red-200';
    if (level === 'medium') return 'text-orange-600 bg-orange-100 border-orange-200';
    return 'text-green-600 bg-green-100 border-green-200';
  }
</script>

<div class="bg-white rounded-xl shadow-sm border border-gray-200">
  <div class="p-4 border-b border-gray-100 flex justify-between items-center">
    <h3 class="text-lg font-bold text-gray-800">Audit Logs</h3>
    <span class="text-xs bg-indigo-50 text-indigo-600 px-2 py-1 rounded">Security Stream</span>
  </div>
  <div class="overflow-x-auto">
    <table class="min-w-full divide-y divide-gray-100">
      <thead class="bg-gray-50 text-left text-xs uppercase tracking-wider font-semibold">
        <tr>
          <th class="px-4 py-3">Action</th>
          <th class="px-4 py-3">Resource</th>
          <th class="px-4 py-3">Risk</th>
          <th class="px-4 py-3">Time</th>
        </tr>
      </thead>
      <tbody class="divide-y divide-gray-100 text-sm">
        {#if history.length === 0}
          <tr>
            <td colspan="4" class="px-4 py-8 text-center text-gray-400 italic">No audit activity recorded.</td>
          </tr>
        {:else}
          {#each history as log}
            <tr>
              <td class="px-4 py-3 font-medium">{@1}</strong></td>
              <td class="px-4 py-3 text-gray-600">{log.resource_id || 'N/A'}</td>
              <td class="px-4 py-3">
                <span class="px-2 py-1 rounded border ${getRiskColor(log.risk_level)} font-bold">
                  {log.risk_level}
                </span>
              </td>
              <td class="px-4 py-3 text-gray-400">{log.timestamp}</td>
            </tr>
          {/each}
        {/if}
      </tbody>
    </table>
  </div>
</div>
