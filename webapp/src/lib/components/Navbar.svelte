<script lang="ts">
  import { page } from '$app/stores';
  import ThemeToggle from './ThemeToggle.svelte';
  import UserMenu from './UserMenu.svelte';
  import Breadcrumb from './ui/Breadcrumb.svelte';

  let isProjectPage = $derived(!!$page.data.project);
  let isNewProject = $derived($page.url.pathname.startsWith('/projects/new'));

  let breadcrumbItems = $derived.by(() => {
    if (isProjectPage) {
      return [
        { label: 'Dashboard', href: '/' },
        { label: $page.data.project?.name || '' },
      ];
    }
    if (isNewProject) {
      return [{ label: 'New Project' }];
    }
    return [];
  });
</script>

<nav class="bg-surface border-b border-border px-6 py-3 flex items-center justify-between sticky top-0 z-40">
  <div class="flex items-center gap-2 min-w-0">
    <Breadcrumb items={breadcrumbItems} />
  </div>

  <div class="flex items-center gap-1 shrink-0">
    <ThemeToggle />
    <UserMenu />
  </div>
</nav>
