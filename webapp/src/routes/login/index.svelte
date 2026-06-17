<script lang="ts">
  let email = '';
  let password = '';
  let error = '';

  async function handleLogin() {
    // Logic for authentication will go here, 
    // eventually connecting to the backend auth service
    if (email && password) {
      console.log("Logging in with:", email);
    } else {
      error = "Please enter both email and password.";
    }
  }
</script>

<div class="min-h-screen bg-slate-50 flex items-center justify-center p-4">
  <div class="max-w-md w-full bg-white p-8 rounded-xl shadow-lg border border-gray-200">
    <h1 class="text-3xl font-bold mb-2 text-blue-600">Cosmoria Login</h1>
    <p class="text-gray-500 mb-8">Access your dashboard</p>
    
    {#if error}
      <div class="bg-red-100 border border-red-400 text-red-700 px-3 py-2 rounded mb-4">
        {error}
      </div>
    {/if}

    <div class="space-y-4">
      <div>
        <label class="block text-sm font-medium text-gray-700 mb-1">Email</label>
        <input bind:value={email} type="text" 
               class="w-full p-3 border rounded-lg focus:ring-2 focus:ring-blue-500 outline-none transition" />
      </div>
      <div>
        <label class="block text-sm font-medium text-gray-700 mb-1">Password</label>
        <input bind:value={password} type="password" 
               class="w-full p-3 border rounded-lg focus:ring-2 focus:ring-blue-500 outline-none transition" />
      </div>
      <button on:click={handleLogin} 
              class="w-full bg-blue-600 text-white font-bold py-3 px-4 rounded-lg hover:bg-blue-700 transition duration-200">
        Sign In
      </button>
    </div>
    <div class="mt-6 text-center text-sm">
      <a href="/setup" class="text-blue-600 hover:underline">Start a new project setup</a>
    </div>
  </div>
</div>
