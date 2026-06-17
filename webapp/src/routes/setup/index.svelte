<script lang="ts">
  import { onMount } from 'svelte';
  let step = 1;
  let initialData = { projectName: '', adminEmail: '', adminPass: '' };

  function nextStep() {
    if (step === 1) step = 2;
  }
</script>

<div class="min-h-screen bg-slate-50 flex items-center justify-center p-4">
  <div class="max-w-md w-full bg-white p-8 rounded-xl shadow-lg border border-grayed-200">
    <h1 class="text-2xl font-bold mb-6 text-blue-600">Configurazione Iniziale Cosmoria</h1>
    
    {#if step === 1}
      <div>
        <p class="mb-4">Inserisci il nome del tuo progetto:</p>
        <input bind:value={initialData.projectName} placeholder="Esempio: My SaaS" 
               class="w-full p-2 border rounded mb-4 focus:ring-2 focus:ring-blue-500 outline-none" />
        <button on:click={nextStep} class="bg-blue-600 text-white px-4 py-2 rounded hover:bg-blue-700 w-full">Avanti</button>
      </div>
    {:else}
      <div>
        <p class="mb-4">Configura le credenziali del super admin:</p>
        <input bind:value={initialData.adminEmail} type="email" placeholder="email@esempio.com" 
               class="w-full p-2 border rounded mb-4 focus:ring-2 focus:ring-blue-500 outline-none" />
        <input bind:value={initialData.adminPass} type="password" placeholder="Password amministratore" 
               class="w-full p-2 border rounded mb-6 focus:ring-2 focus:ring-blue-500 outline-none" />
        <button class="bg-green-600 text-white px-4 py-2 rounded hover:bg-green-700 w-full">Completa Setup</button>
      </div>
    {/if}
  </div>
</div>
