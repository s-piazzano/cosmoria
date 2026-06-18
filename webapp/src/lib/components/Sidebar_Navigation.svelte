<script lang="ts">
  export let activeItem: string = 'dashboard';
</script>

<nav class="w-64 bg-slate-900 text-white h-screen p-6 flex flex-col">
  <div class="text-2xl font-bold mb-10">Cosmoria</div>
  <ul class="space-y-4 flex-grow">
    <li><a href="/" class="hover:text-blue-400 {activeItem === 'dashboard' ? 'text-blue-400' : ''}">Dashboard</a></li>
    <li><a href="/resources" class="hover:text-blue-400 {activeItem === 'resources' ? 'text-blue-100' : ''}">Resource Center</a></li>
    <li><a href="/organization" class="hover:text-blue-400 {activeItem === 'org' ? 'text-blue-100' : ''}">Organization</a></li>
    <li><a href="/files" class="hover:text-blue-400 {activeItem === 'files' ? 'text-blue-100' : ''}">File Manager</a></li>
  </ul>
  <div class="border-t border-gray-700 pt-6">
    <a href="/settings" class="hover:text-blue-400">Settings</a>
  </div>
</nav>
