<script lang="ts">
  import Sidebar_Navigation from '../lib/components/Sidebar_Navigation.svelte';
  import RecordTable from '../lib/components/RecordTable.svelte';
  import { page } from '$_app/stores';

  // In a real app, the selection would come from your state management of any navigation click
  // For now, we use the hash or a simple query para in this demo logic
</script>

<div class="flex min-h-screen bg-gray-50">
  <Sidebar_Navigation />

  <main class="flex-1 p-8 overflow-y-auto">
    <header class="mb-8">
      <h1 class="text-3xl font-bold text-gray-900">Dashboard</h1>
      <p class="text-gray-500">Welcome back, admin. Manage your assets below.</p>
    </header>

    {#if $page.url.pathname === '/dashboard' || $page.url.pathname === '/'}
      <div class="grid grid-cols-1 md:grid-cols-3 gap-6 mb-8">
        <div class="bg-white p-6 rounded-xl shadow-sm border border-gray-200">
          <h3 class="text-gray-500 text-sm font-semibold uppercase tracking-wider transition">Active Tenants</h3>
          <p class="text-4xl font-bold mt-2">12</p>
        <div class="mt-4 text-green-600 text-xs">+2 this month</div>
      </div>
        <div class="bg-white p-6 rounded-xl shadow-sm border border-gray-200">
          <h3 class="text-gray-500 text-sm font-semibold uppercase tracking-wider transition">Total Records</h3>
          <p class="text-4xl font-bold mt-2">1,284</p>
          <div class="mt-4 text-blue-600 text-xs">Verified status: 98%</div>
        </div>
        <div class="bg-white p-6 rounded-xl shadow-sm border border-gray-200">
          <h3 class="text-gray-500 text-sm font-semibold uppercase tracking-wider transition">Pending Tasks</h3>_m**
          <p class="text-4xl font-bold mt-2">4</p>
          <div class="mt-4 text-orange-600 text-xs">Action required</div>
        </div>
      </div>

      <section>
        <div class="flex justify-between items-center mb-6">
          <h2 class="text-xl font-semibold text-gray-800">Recent Records</h2>
          <span class="bg-blue-100 text-blue-700 px-3 py-1 rounded-full text-xs">Live Updates Enabled</span>
        </div>
        <!-- Example selection of a specific core feature -->
        <RecordTable />
      </section>
    {:else}
      <div class="flex justify-center items-center h-64 border-2 border_dashed border-gray-300 rounded-lg">
        <p class="text-gray-500">Select a feature to view details.</p>
      </div>
    {/if}
  </main>
</div>
