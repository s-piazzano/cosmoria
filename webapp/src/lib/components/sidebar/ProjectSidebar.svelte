<script lang="ts">
  import { goto } from "$app/navigation";

  let {
    slug,
    active,
    multitenancyEnabled = false,
  }: { slug: string; active: string; multitenancyEnabled?: boolean } = $props();

  let tabs = $derived.by(() => {
    const all = [
      { id: "overview", label: "Overview", icon: "📊" },
      { id: "tenants", label: "Tenants", icon: "🏢" },
      { id: "collections", label: "Collections", icon: "📁" },
      { id: "roles", label: "Roles", icon: "🔑" },
      { id: "users", label: "Users", icon: "👥" },
      { id: "files", label: "Files", icon: "📄" },
      { id: "audit", label: "Audit Log", icon: "📋" },
      { id: "api-keys", label: "API Keys", icon: "🔌" },
    ];
    return multitenancyEnabled ? all : all.filter((t) => t.id !== "tenants");
  });

  function goToTab(tabId: string) {
    goto(`/${slug}?tab=${tabId}`);
  }
</script>

<aside class="w-64 bg-surface border-r border-border flex shrink-0">
  <nav class="flex-1 py-4">
    {#each tabs as tab}
      <button
        onclick={() => goToTab(tab.id)}
        class="w-full text-left px-4 py-2.5 text-sm flex items-center gap-3 transition
               {active === tab.id
          ? 'bg-primary-50 dark:bg-primary-950/30 text-primary-700 dark:text-primary-400 font-medium border-r-2 border-primary'
          : 'text-muted hover:bg-surface-soft hover:text-text'}"
      >
        <span class="text-base">{tab.icon}</span>
        <span>{tab.label}</span>
      </button>
    {/each}
  </nav>
</aside>
