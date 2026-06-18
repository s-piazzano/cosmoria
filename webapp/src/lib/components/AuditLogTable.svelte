<script lang="ts">
  import type { AuditEntry } from '../types';
  
  export let history = [] as AuditEntry[];

  const getRiskColor = (level: string) => {
    if (level === 'high') return 'text-danger bg-danger-soft border-danger';
    if (level === 'medium') return 'text-warning bg-warning-soft border-warning';
    return 'text-success bg-success-soft border-success';
  }
</script>

<div class="bg-surface rounded-xl shadow-sm border border-border">
  <div class="p-4 border-b border-border flex justify-between items-center">
    <h3 class="text-lg font-bold text-text">Audit Logs</h3>
    <span class="text-xs bg-secondary-50 text-secondary px-2 py-1 rounded">Security Stream</span>
  </div>
  <div class="overflow-x-auto">
    <table class="min-w-full divide-y divide-border">
      <thead class="bg-surface-soft text-left text-xs uppercase tracking-wider font-semibold">
        <tr>
          <th class="px-4 py-3">Action</th>
          <th class="px-4 py-3">Resource</th>
          <th class="px-4 py-3">Risk</th>
          <th class="px-4 py-3">Time</th>
        </tr>
      </thead>
      <tbody class="divide-y divide-border text-sm">
        {#if history.length === 0}
          <tr>
            <td colspan="4" class="px-4 py-8 text-center text-muted italic">No audit activity recorded.</td>
          </tr>
        {:else}
          {#each history as log}
            <tr>
              <td class="px-4 py-3 font-medium">{log.action}</td>
              <td class="px-4 py-3 text-muted">{log.resource_id || 'N/A'}</td>
              <td class="px-4 py-3">
                <span class="px-2 py-1 rounded border ${getRiskColor(log.risk_level)} font-bold">
                  {log.risk_level}
                </span>
              </td>
              <td class="px-4 py-3 text-muted">{log.timestamp}</td>
            </tr>
          {/each}
        {/if}
      </tbody>
    </table>
  </div>
</div>
