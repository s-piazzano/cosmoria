<script lang="ts">
  import { setAuthToken } from '../../lib/services/api';

  let email = '';
  let password = '';
  let error = '';
  let loading = false;

  async function completeSetup() {
    if (!email || !password) {
      error = 'Please fill in all fields.';
      return;
    }

    loading = true;
    error = '';

    try {
      const response = await fetch('/api/admin/setup', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ email, password }),
      });

      if (!response.ok) {
        error = 'Setup failed. Please try again.';
        return;
      }

      const data = await response.json();
      setAuthToken(data.token);
      localStorage.setItem('cosmoria_email', data.admin.email);
      window.location.href = '/';
    } catch (e) {
      error = 'Network error. Please try again.';
    } finally {
      loading = false;
    }
  }
</script>

<div class="min-h-screen bg-background flex items-center justify-center p-4">
  <div class="max-w-md w-full bg-surface p-8 rounded-xl shadow-lg border border-border">
    <h1 class="text-2xl font-bold mb-6 text-primary">Cosmoria Setup</h1>

    {#if error}
      <div class="bg-danger-soft border border-danger text-danger px-3 py-2 rounded mb-4">
        {error}
      </div>
    {/if}

    <div>
      <p class="mb-4 text-muted">Create your super admin account:</p>
      <input bind:value={email} type="email" placeholder="admin@example.com"
             class="w-full p-2 border rounded mb-4 focus:ring-2 focus:ring-primary outline-none bg-surface text-text border-border" />
      <input bind:value={password} type="password" placeholder="Admin password"
             class="w-full p-2 border rounded mb-6 focus:ring-2 focus:ring-primary outline-none bg-surface text-text border-border" />
      <button onclick={completeSetup} disabled={loading}
              class="bg-primary text-white px-4 py-2 rounded hover:bg-primary-hover w-full disabled:opacity-50">
        {loading ? 'Setting up...' : 'Complete Setup'}
      </button>
    </div>
  </div>
</div>
