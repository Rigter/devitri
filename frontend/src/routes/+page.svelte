<script lang="ts">
  import type { VaultResponse } from '../lib/api/client';
  import { ApiClient } from '../lib/api/client';
  import StatsCard from '../lib/components/molecules/StatsCard.svelte';
  import { Package, Files, Folder, HardDrive, ArrowRight, Plus, RefreshCcw, Clock } from 'lucide-svelte';
  import { navigateTo } from '$lib/utils/navigation';
  import { formatVaultLastSync } from '$lib/utils/vault-display';
  import { startVaultPoll } from '$lib/utils/vault-poll';
  import {
    hasVaultListSyncChanged,
    vaultListSyncFingerprint,
  } from '$lib/utils/vault-sync';
  import { serverStatus } from '$lib/stores/server.svelte';
  import { authStore } from '$lib/stores/auth.svelte';
  
  const apiClient = new ApiClient();
  const POLL_INTERVAL_MS = 30_000;
  
  let vaults = $state<VaultResponse[]>([]);
  let isLoading = $state(false);
  let error = $state<string | null>(null);
  let listSyncFingerprint = $state('');
  
  async function fetchVaults(showLoading = true): Promise<void> {
    if (showLoading) {
      isLoading = true;
    }
    error = null;
    try {
      const next = await apiClient.getVaults();
      vaults = next;
      listSyncFingerprint = vaultListSyncFingerprint(next);
    } catch (err) {
      error = err instanceof Error ? err.message : 'An unknown error occurred';
    } finally {
      if (showLoading) {
        isLoading = false;
      }
    }
  }

  async function pollVaultsIfSyncChanged(): Promise<void> {
    try {
      const next = await apiClient.getVaults();
      if (!hasVaultListSyncChanged(listSyncFingerprint, next)) {
        return;
      }
      vaults = next;
      listSyncFingerprint = vaultListSyncFingerprint(next);
    } catch {
      // Silent background poll — keep current list on failure
    }
  }
  
  function navigateIfNeeded(target: string): void {
    if (typeof window !== 'undefined' && window.location.pathname === target) return;
    navigateTo(target);
  }

  $effect(() => {
    if (serverStatus.isLoading || authStore.isLoading || !serverStatus.checked) return;

    if (!authStore.isAuthenticated) {
      if (!serverStatus.ready) {
        navigateIfNeeded('/setup');
      } else {
        navigateIfNeeded('/login');
      }
      return;
    }

    void fetchVaults(true);

    const stopPoll = startVaultPoll({
      intervalMs: POLL_INTERVAL_MS,
      tick: pollVaultsIfSyncChanged,
    });

    return stopPoll;
  });
</script>

<div class="space-y-10">
  <header class="flex flex-col gap-2">
    <h1 class="font-display italic text-4xl font-bold tracking-tight">Dashboard</h1>
    <p class="text-muted-foreground font-sans">Self-hosted web visualization for your Obsidian vaults.</p>
  </header>
  
  {#if isLoading && vaults.length === 0}
    <div class="flex h-60 flex-col items-center justify-center gap-4 text-muted-foreground">
      <RefreshCcw class="animate-spin" size={32} />
      <p class="text-xs font-mono uppercase tracking-widest">Fetching Vaults...</p>
    </div>
  {:else}
    <section class="space-y-6">
      <div class="flex items-center justify-between">
        <h2 class="font-sans text-xs font-bold uppercase tracking-widest text-muted-foreground">Quick Stats</h2>
      </div>
      <div class="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-4">
        <StatsCard 
          title="Vaults" 
          value={vaults.length} 
          icon={Package} 
        />
        <StatsCard 
          title="Files" 
          value={vaults.reduce((sum, vault) => sum + vault.stats.files, 0)} 
          icon={Files} 
        />
        <StatsCard 
          title="Folders" 
          value={vaults.reduce((sum, vault) => sum + vault.stats.folders, 0)} 
          icon={Folder} 
        />
        <StatsCard 
          title="Size (MB)" 
          value={(vaults.reduce((sum, vault) => sum + vault.stats.size_bytes, 0) / (1024 * 1024)).toFixed(1)} 
          icon={HardDrive} 
        />
      </div>
    </section>
    
    <section class="space-y-6">
      <div class="flex items-center justify-between">
        <h2 class="font-sans text-xs font-bold uppercase tracking-widest text-muted-foreground text-foreground">Vaults</h2>
        <button 
          onclick={() => fetchVaults(true)}
          class="inline-flex items-center gap-2 rounded-md border px-3 py-1.5 text-xs font-bold transition-colors hover:bg-muted"
        >
          <RefreshCcw size={14} class={isLoading ? 'animate-spin' : ''} />
          <span>Refresh</span>
        </button>
      </div>

      {#if error}
        <div class="flex h-40 items-center justify-center rounded-xl border border-dashed border-destructive/50 text-sm text-destructive">
          Error: {error}
        </div>
      {:else if vaults.length === 0}
        <div class="flex flex-col items-center justify-center gap-6 rounded-xl border border-dashed text-center p-8 py-24 sm:py-32">
          <div class="bg-muted p-4 rounded-full">
            <Package size={40} class="text-muted-foreground" />
          </div>
          <div class="space-y-2">
            <h3 class="font-sans text-xl font-bold tracking-tight">No vaults detected</h3>
            <p class="text-sm text-muted-foreground max-w-sm mx-auto">
              You haven't synchronized any Obsidian vaults yet. To start seeing your notes here, you need to connect the Devitri plugin from Obsidian.
            </p>
          </div>
          <div class="flex flex-col sm:flex-row gap-3 mt-4">
            <a 
              href="/connect"
              data-sveltekit-reload
              class="inline-flex items-center gap-2 rounded-md bg-foreground px-6 py-2.5 text-sm font-bold text-background transition-colors hover:bg-foreground/90"
            >
              <Plus size={16} />
              <span>Connect Obsidian</span>
            </a>
            <button 
              onclick={() => fetchVaults(true)}
              class="inline-flex items-center gap-2 rounded-md border px-6 py-2.5 text-sm font-bold transition-colors hover:bg-muted"
            >
              <RefreshCcw size={16} />
              <span>Check again</span>
            </button>
          </div>
        </div>
      {:else}
        <div class="grid grid-cols-1 gap-6 md:grid-cols-2 lg:grid-cols-3">
          {#each vaults as vault}
            <div class="group relative flex flex-col justify-between rounded-xl border bg-card p-6 shadow-sm transition-all hover:shadow-md hover:border-muted">
              <div>
                <h3 class="font-sans text-lg font-bold tracking-tight">{vault.name}</h3>
                <p class="mt-1 font-mono text-[10px] text-muted-foreground truncate">{vault.path}</p>
                
                <div class="mt-6 flex gap-6 font-mono text-[10px] uppercase tracking-wider text-muted-foreground">
                  <div class="flex flex-col gap-1">
                    <span class="text-foreground font-bold">{vault.stats.files}</span>
                    <span>Files</span>
                  </div>
                  <div class="flex flex-col gap-1">
                    <span class="text-foreground font-bold">{vault.stats.folders}</span>
                    <span>Folders</span>
                  </div>
                  <div class="flex flex-col gap-1">
                    <span class="text-foreground font-bold">{(vault.stats.size_bytes / (1024 * 1024)).toFixed(1)}MB</span>
                    <span>Size</span>
                  </div>
                </div>

                <div class="mt-5 border-t border-border pt-4">
                  <div class="flex items-center gap-1.5 text-[10px] font-bold uppercase tracking-wider text-muted-foreground">
                    <Clock size={12} />
                    <span>Last sync</span>
                  </div>
                  {#if formatVaultLastSync(vault)}
                    <p class="mt-1.5 font-mono text-[11px] leading-relaxed text-foreground normal-case tracking-normal">
                      {formatVaultLastSync(vault)}
                    </p>
                  {:else}
                    <p class="mt-1.5 font-mono text-[11px] text-muted-foreground normal-case tracking-normal">
                      No sync recorded yet
                    </p>
                  {/if}
                </div>
              </div>

              <div class="mt-8">
                <a 
                  href="/vault/{vault.id}"
                  data-sveltekit-reload
                  class="inline-flex items-center gap-2 text-xs font-bold uppercase tracking-widest text-foreground underline underline-offset-4 transition-all group-hover:gap-3"
                >
                  <span>Open Vault</span>
                  <ArrowRight size={14} />
                </a>
              </div>
            </div>
          {/each}
        </div>
      {/if}
    </section>
  {/if}
</div>
