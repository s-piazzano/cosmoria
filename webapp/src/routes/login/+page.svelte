<script lang="ts">
  import { onMount } from 'svelte';
  import { goto } from '$app/navigation';
  import { setAuthToken, validateToken } from '../../lib/services/api';

  let email = '';
  let password = '';
  let error = '';
  let loading = false;
  let checking = true;

  onMount(async () => {
    const token = localStorage.getItem('cosmoria_token');
    if (token) {
      const admin = await validateToken();
      if (admin) {
        goto('/', { replaceState: true });
        return;
      }
    }
    checking = false;
  });

  async function handleLogin() {
    if (!email || !password) {
      error = 'Please enter both email and password.';
      return;
    }

    loading = true;
    error = '';

    try {
      const response = await fetch('/api/admin/login', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ email, password }),
      });

      if (!response.ok) {
        error = 'Login failed. Check your credentials.';
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

{#if !checking}
<div class="min-h-screen bg-background flex items-center justify-center p-4">
  <div class="max-w-md w-full bg-surface p-8 rounded-xl shadow-lg border border-border">
    <h1 class="text-3xl font-bold mb-2 text-primary">Cosmoria Login</h1>
    <p class="text-muted mb-8">Access your dashboard</p>

    {#if error}
      <div class="bg-danger-soft border border-danger text-danger px-3 py-2 rounded mb-4">
        {error}
      </div>
    {/if}

    <div class="space-y-4">
      <div>
        <label class="block text-sm font-medium text-muted mb-1">Email</label>
        <input bind:value={email} type="text"
               class="w-full p-3 border rounded-lg focus:ring-2 focus:ring-primary outline-none transition bg-surface text-text border-border" />
      </div>
      <div>
        <label class="block text-sm font-medium text-muted mb-1">Password</label>
        <input bind:value={password} type="password"
               class="w-full p-3 border rounded-lg focus:ring-2 focus:ring-primary outline-none transition bg-surface text-text border-border" />
      </div>
      <button onclick={handleLogin} disabled={loading}
              class="w-full bg-primary text-white font-bold py-3 px-4 rounded-lg hover:bg-primary-hover transition duration-200 disabled:opacity-50">
        {loading ? 'Signing in...' : 'Sign In'}
      </button>
    </div>
  </div>
</div>
{/if}
