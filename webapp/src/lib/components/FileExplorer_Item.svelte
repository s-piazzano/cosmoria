<script lang="ts">
  export let name = '';
  export let size = 0;
  export let type = 'file';
</script>

<div class="flex items-center p-2 hover:bg-gray-50 rounded transition cursor-pointer">
  <span class="mr-3 text-xl">{type === 'folder' ? '📁' : '📄'}</span>
  <span class="text-sm font-medium text-gray-800">{name}</span>
  <span class="ml-auto text-xs text-gray-400">{(size / 1024).toFixed(1)} KB</span>
</div>
