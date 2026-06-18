<script lang="ts">
  import { setAuthToken } from '../../lib/services/api';

  let step = 1;
  let initialData = { projectName: '', adminEmail: '', adminPass: '' };
  let error = '';
  let loading = false;

  function nextStep() {
    if (step === 1) step = 2;
  }

  async function completeSetup() {
    if (!initialData.projectName || !initialData.adminEmail || !initialData.adminPass) {
      error = 'Please fill in all fields.';
      return;
    }

    loading = true;
    error = '';

    try {
      const response = await fetch('/api/admin/setup', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          email: initialData.adminEmail,
          password: initialData.adminPass,
          project_name: initialData.projectName,
        }),
      });

      if (!response.ok) {
        error = 'Setup failed. Please try again.';
        return;
      }

      const data = await response.json();
      setAuthToken(data.token);
      window.location.href = '/';
    } catch (e) {
      error = 'Network error. Please try again.';
    } finally {
      loading = false;
    }
  }
</script>

<div class="min-h-screen bg-slate-50 flex items-center justify-center p-4">
  <div class="max-w-md w-full bg-white p-8 rounded-xl shadow-lg border border-gray-200">
    <h1 class="text-2xl font-bold mb-6 text-blue-600">Cosmoria Setup</h1>

    {#if error}
      <div class="bg-red-100 border border-red-400 text-red-700 px-3 py-2 rounded mb-4">
        {error}
      </div>
    {/if}

    {#if step === 1}
      <div>
        <p class="mb-4">Enter your project name:</p>
        <input bind:value={initialData.projectName} placeholder="e.g. My SaaS"
               class="w-full p-2 border rounded mb-4 focus:ring-2 focus:ring-blue-500 outline-none" />
        <button onclick={nextStep} class="bg-blue-600 text-white px-4 py-2 rounded hover:bg-blue-700 w-full">Next</button>
      </div>
    {:else}
      <div>
        <p class="mb-4">Configure super admin credentials:</p>
        <input bind:value={initialData.adminEmail} type="email" placeholder="admin@example.com"
               class="w-full p-2 border rounded mb-4 focus:ring-2 focus:ring-blue-500 outline-none" />
        <input bind:value={initialData.adminPass} type="password" placeholder="Admin password"
               class="w-full p-2 border rounded mb-6 focus:ring-2 focus:ring-blue-500 outline-none" />
        <button onclick={completeSetup} disabled={loading}
                class="bg-green-600 text-white px-4 py-2 rounded hover:bg-green-700 w-full disabled:opacity-50">
          {loading ? 'Setting up...' : 'Complete Setup'}
        </button>
      </div>
    {/if}
  </div>
</div>
