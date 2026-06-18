<script lang="ts">
  import { goto } from '$app/navigation';
  import { ChevronDown, LogOut, UserCog } from '@lucide/svelte';
  import Dropdown from './ui/Dropdown.svelte';

  let email = $state(localStorage.getItem('cosmoria_email') || '');

  function handleLogout() {
    localStorage.removeItem('cosmoria_token');
    localStorage.removeItem('cosmoria_email');
    goto('/login');
  }
</script>

<Dropdown>
  {#snippet children({ open, toggle, close })}
    <button onclick={toggle}
            class="flex items-center gap-1.5 px-3 py-1.5 rounded-lg hover:bg-surface-soft text-sm text-muted transition cursor-pointer">
      <span class="font-medium">{email}</span>
      <ChevronDown size={14} class="text-muted {open ? 'rotate-180' : ''}" />
    </button>

    {#if open}
      <div onclick={(e) => e.stopPropagation()}
           class="absolute right-0 top-full mt-1 w-56 bg-surface border border-border rounded-lg shadow-xl z-50 py-1">
        <div class="px-4 py-2 text-xs text-muted border-b border-border truncate">
          {email}
        </div>
        <button onclick={() => { close(); goto('/settings'); }}
                class="w-full flex items-center gap-2 px-4 py-2 text-sm text-muted hover:bg-surface-soft transition">
          <UserCog size={16} class="text-muted" />
          Manage Super Admin
        </button>
        <hr class="border-border">
        <button onclick={() => { close(); handleLogout(); }}
                class="w-full flex items-center gap-2 px-4 py-2 text-sm text-danger hover:bg-surface-soft transition">
          <LogOut size={16} />
          Logout
        </button>
      </div>
    {/if}
  {/snippet}
</Dropdown>
