<script lang="ts">
  import { onMount } from 'svelte';
  import { ApiClient, getApiBaseUrl } from '$lib/api/client';
  import { Shield, Smartphone, Terminal, Copy, Check, ChevronLeft, AlertCircle } from 'lucide-svelte';
  import { navigateTo } from '$lib/utils/navigation';
  import { authStore } from '$lib/stores/auth.svelte';

  const apiClient = new ApiClient();
  
  let deviceName = $state('');
  let deviceId = $state('');
  let masterPassword = $state('');
  let generatedToken = $state<string | null>(null);
  let expiresAt = $state<number | null>(null);
  let isLoading = $state(false);
  let error = $state<string | null>(null);
  let copied = $state(false);
  let pluginServerUrl = $state('http://localhost:8080');
  const isLocalDev = $derived(
    typeof window !== 'undefined' && window.location.hostname === 'localhost'
  );

  // Redirection logic
  $effect(() => {
    if (authStore.isLoading) return;
    if (!authStore.isAuthenticated) {
      navigateTo('/login');
    }
  });

  onMount(() => {
    deviceId = crypto.randomUUID();
    pluginServerUrl = getApiBaseUrl() || 'http://localhost:8080';
  });

  async function handleGenerateToken() {
    if (!deviceName) {
      error = 'Please enter a device name';
      return;
    }
    if (!masterPassword) {
      error = 'Enter your master password to authorize this device';
      return;
    }

    isLoading = true;
    error = null;
    try {
      const response = await apiClient.generateToken(deviceId, deviceName, masterPassword);
      masterPassword = '';
      generatedToken = response.token;
      expiresAt = response.expires_at;
    } catch (err) {
      error = err instanceof Error ? err.message : 'Failed to generate token';
    } finally {
      isLoading = false;
    }
  }

  function copyToClipboard() {
    if (!generatedToken) return;
    navigator.clipboard.writeText(generatedToken);
    copied = true;
    setTimeout(() => { copied = false; }, 2000);
  }

  function formatDate(timestamp: number) {
    return new Date(timestamp * 1000).toLocaleDateString();
  }
</script>

<div class="max-w-5xl mx-auto space-y-10">
  <header class="flex flex-col gap-4">
    <a 
      href="/" 
      data-sveltekit-reload
      class="inline-flex items-center gap-2 text-xs font-bold uppercase tracking-widest text-muted-foreground hover:text-foreground transition-colors"
    >
      <ChevronLeft size={14} />
      <span>Back to Dashboard</span>
    </a>
    <div class="flex items-center justify-between">
      <div class="space-y-2">
        <h1 class="font-display italic text-4xl font-bold tracking-tight">Connect Obsidian</h1>
        <p class="text-muted-foreground font-sans">Generate an access key to synchronize your Obsidian vault with Devitri.</p>
      </div>
    </div>
  </header>

  <div class="grid gap-4">
    <section class="rounded-xl border bg-card p-8 space-y-6">
      <div class="flex items-center gap-4">
        <div class="bg-muted p-3 rounded-lg">
          <Smartphone size={24} class="text-foreground" />
        </div>
        <div>
          <h2 class="font-sans text-lg font-bold">1. Name your Device</h2>
          <p class="text-sm text-muted-foreground">This helps you identify which Obsidian instance is syncing.</p>
        </div>
      </div>

      {#if !generatedToken}
        <div class="space-y-4 pt-4">
          <div class="space-y-2">
            <label for="deviceName" class="text-xs font-bold uppercase tracking-widest text-muted-foreground">Device Name</label>
            <input 
              id="deviceName"
              type="text" 
              bind:value={deviceName}
              placeholder="e.g. MacBook Pro, iPhone, My Obsidian"
              class="w-full rounded-md border bg-background px-4 py-2 font-sans focus:outline-none focus:ring-2 focus:ring-foreground/10"
              onkeydown={(e) => e.key === 'Enter' && handleGenerateToken()}
            />
          </div>

          <div class="space-y-2">
            <label for="masterPassword" class="text-xs font-bold uppercase tracking-widest text-muted-foreground">Master Password</label>
            <input
              id="masterPassword"
              type="password"
              bind:value={masterPassword}
              autocomplete="current-password"
              placeholder="Required to issue a device access key"
              class="w-full rounded-md border bg-background px-4 py-2 font-sans focus:outline-none focus:ring-2 focus:ring-foreground/10"
              onkeydown={(e) => e.key === 'Enter' && handleGenerateToken()}
            />
          </div>

          {#if error}
            <div class="flex items-center gap-2 text-xs font-bold text-destructive bg-destructive/5 p-3 rounded-md border border-destructive/20">
              <AlertCircle size={14} />
              <span>{error}</span>
            </div>
          {/if}

          <button 
            onclick={handleGenerateToken}
            disabled={isLoading}
            class="w-full inline-flex items-center justify-center gap-2 rounded-md bg-foreground px-6 py-3 text-sm font-bold text-background transition-colors hover:bg-foreground/90 disabled:opacity-50"
          >
            {#if isLoading}
              <div class="h-4 w-4 animate-spin rounded-full border-2 border-background border-t-transparent"></div>
            {:else}
              <Shield size={16} />
            {/if}
            <span>Generate Access Key</span>
          </button>
        </div>
      {:else}
        <div class="space-y-6 pt-4 animate-in fade-in slide-in-from-bottom-4 duration-500">
          <div class="space-y-2">
            <span class="text-xs font-bold uppercase tracking-widest text-muted-foreground">Your Access Key</span>
            <div class="group relative">
              <div class="w-full break-all rounded-md border bg-muted/50 px-4 py-4 font-mono text-xs pr-12">
                {generatedToken}
              </div>
              <button 
                onclick={copyToClipboard}
                class="absolute right-3 top-3 p-2 rounded-md hover:bg-background border shadow-sm transition-all"
                title="Copy to clipboard"
              >
                {#if copied}
                  <Check size={16} class="text-emerald-500" />
                {:else}
                  <Copy size={16} class="text-muted-foreground" />
                {/if}
              </button>
            </div>
            <p class="text-[10px] text-muted-foreground italic">
              Copy this key and paste it in the Devitri plugin settings in Obsidian. This key expires on <strong>{formatDate(expiresAt!)}</strong>.
            </p>
          </div>

          <div class="bg-emerald-500/5 border border-emerald-500/20 p-4 rounded-lg flex gap-3">
            <Shield size={20} class="text-emerald-500 shrink-0" />
            <p class="text-xs text-emerald-600 leading-relaxed">
              <strong>Security Note:</strong> Treat this key like a password. Do not share it. It gives full access to synchronize and read your vaults.
            </p>
          </div>

          <button 
            onclick={() => { generatedToken = null; deviceName = ''; }}
            class="text-xs font-bold uppercase tracking-widest text-muted-foreground hover:text-foreground underline underline-offset-4"
          >
            Generate another key
          </button>
        </div>
      {/if}
    </section>

    <section class="rounded-xl border bg-card p-8 space-y-6">
      <div class="flex items-center gap-4">
        <div class="bg-muted p-3 rounded-lg">
          <Terminal size={24} class="text-foreground" />
        </div>
        <div>
          <h2 class="font-sans text-lg font-bold">2. How to use</h2>
          <p class="text-sm text-muted-foreground">Follow these steps to complete the integration.</p>
        </div>
      </div>

      <div class="space-y-4 pt-4">
        <ol class="space-y-4 list-decimal list-inside text-sm text-muted-foreground font-sans">
          <li class="pl-2">Build and copy the plugin to your vault: <code class="bg-muted px-1 rounded text-foreground">{`{vault}/.obsidian/plugins/devitri-obsidian-plugin/`}</code> (see <a href="https://github.com/rigter/devitri-obsidian-plugin" class="underline text-foreground hover:text-primary" target="_blank" rel="noopener noreferrer">devitri-obsidian-plugin</a>).</li>
          <li class="pl-2">Open <strong>Obsidian</strong> → <strong>Settings</strong> → <strong>Community Plugins</strong> → enable <strong>Devitri</strong>.</li>
          <li class="pl-2">
            In plugin settings, set <strong>Server URL</strong> to the API base URL:
            <code class="bg-muted px-1 rounded text-foreground font-bold">{pluginServerUrl}</code>
            {#if isLocalDev}
              <span class="block mt-1 text-xs italic">Local dev uses port <strong>8080</strong> (backend), not the dashboard on :3000.</span>
            {/if}
          </li>
          <li class="pl-2">Set <strong>Vault ID</strong> to match the vault configured on the server.</li>
          <li class="pl-2">Paste the <strong>Access Key</strong> you generated above.</li>
          <li class="pl-2">Click <strong>Connect & Sync</strong>.</li>
        </ol>
      </div>
    </section>
  </div>
</div>
