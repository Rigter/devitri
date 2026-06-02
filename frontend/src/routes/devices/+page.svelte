<script lang="ts">
  import { onMount } from 'svelte';
  import { ApiClient } from '$lib/api/client';
  import { Smartphone, Shield, Trash2, Clock, Calendar, ChevronLeft, RefreshCcw, AlertCircle } from 'lucide-svelte';
  import { navigateTo } from '$lib/utils/navigation';
  import { authStore } from '$lib/stores/auth.svelte';

  const apiClient = new ApiClient();
  
  let sessions = $state<any[]>([]);
  let currentSessionId = $state<number | null>(null);
  let isLoading = $state(true);
  let isRevoking = $state<number | null>(null);
  let error = $state<string | null>(null);

  async function fetchSessions() {
    isLoading = true;
    error = null;
    try {
      // Get current session first to highlight it
      const current = await apiClient.getCurrentSession();
      currentSessionId = current.id;

      sessions = await apiClient.getSessions();
    } catch (err) {
      error = err instanceof Error ? err.message : 'Failed to fetch sessions';
    } finally {
      isLoading = false;
    }
  }

  async function handleRevoke(id: number) {
    if (!confirm('Are you sure you want to revoke access for this device? It will be immediately logged out.')) {
      return;
    }

    let password: string | undefined;
    if (id !== currentSessionId) {
      password = prompt('Enter your master password to revoke this device:') ?? '';
      if (!password) return;
    }

    isRevoking = id;
    try {
      await apiClient.revokeSession(id, password);
      if (id === currentSessionId) {
        await authStore.logout();
        navigateTo('/login');
      } else {
        sessions = sessions.filter(s => s.id !== id);
      }
    } catch (err) {
      alert(err instanceof Error ? err.message : 'Failed to revoke session');
    } finally {
      isRevoking = null;
    }
  }

  function formatDate(timestamp: number) {
    return new Date(timestamp * 1000).toLocaleString();
  }

  function getTimeAgo(timestamp: number) {
    const seconds = Math.floor(Date.now() / 1000 - timestamp);
    if (seconds < 60) return 'Just now';
    const minutes = Math.floor(seconds / 60);
    if (minutes < 60) return `${minutes}m ago`;
    const hours = Math.floor(minutes / 60);
    if (hours < 24) return `${hours}h ago`;
    return `${Math.floor(hours / 24)}d ago`;
  }

  // Redirection logic
  $effect(() => {
    if (authStore.isLoading) return;
    if (!authStore.isAuthenticated) {
      navigateTo('/login');
    } else {
      fetchSessions();
    }
  });

  onMount(() => {
    // Current session ID and device ID are handled within fetchSessions or via crypto
  });
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
        <h1 class="font-display italic text-4xl font-bold tracking-tight">Authorized Devices</h1>
        <p class="text-muted-foreground font-sans">Manage devices and applications that have access to your Devitri instance.</p>
      </div>
      <button 
        onclick={fetchSessions}
        class="inline-flex h-10 w-10 items-center justify-center rounded-md border border-border bg-background transition-colors hover:bg-muted"
        title="Refresh list"
      >
        <RefreshCcw size={18} class={isLoading ? 'animate-spin' : ''} />
      </button>
    </div>
  </header>

  {#if error}
    <div class="flex items-center gap-3 text-sm font-bold text-destructive bg-destructive/5 p-6 rounded-xl border border-destructive/20">
      <AlertCircle size={20} />
      <span>Error: {error}</span>
    </div>
  {:else if isLoading && sessions.length === 0}
    <div class="grid gap-4">
      {#each Array(3) as _}
        <div class="h-24 rounded-xl border border-dashed animate-pulse bg-muted/20"></div>
      {/each}
    </div>
  {:else if sessions.length === 0}
    <div class="flex h-60 flex-col items-center justify-center gap-4 rounded-xl border border-dashed text-center">
      <Smartphone size={40} class="text-muted-foreground/30" />
      <p class="text-muted-foreground">No active sessions found.</p>
    </div>
  {:else}
    <div class="grid gap-4">
      {#each sessions as session}
        <div class="group relative flex items-center justify-between rounded-xl border bg-card p-6 shadow-sm transition-all hover:shadow-md">
          <div class="flex items-center gap-5">
            <div class="bg-muted p-3 rounded-lg text-muted-foreground group-hover:text-foreground transition-colors">
              <Smartphone size={24} />
            </div>
            <div class="space-y-1">
              <div class="flex items-center gap-3">
                <h3 class="font-sans font-bold">{session.device_name}</h3>
                {#if session.id === currentSessionId}
                  <span class="inline-flex items-center rounded-full bg-emerald-500/10 px-2 py-0.5 text-[10px] font-bold uppercase tracking-wider text-emerald-500">
                    This Device
                  </span>
                {/if}
              </div>
              <div class="flex items-center gap-4 text-[10px] font-mono uppercase tracking-widest text-muted-foreground">
                <span class="flex items-center gap-1">
                  <Clock size={10} />
                  Last seen {getTimeAgo(session.last_seen)}
                </span>
                <span class="flex items-center gap-1">
                  <Calendar size={10} />
                  Expires {new Date(session.expires_at * 1000).toLocaleDateString()}
                </span>
                <span class="hidden sm:inline opacity-50">ID: {session.device_id}</span>
              </div>
            </div>
          </div>

          <div class="flex items-center gap-2">
            <button 
              onclick={() => handleRevoke(session.id)}
              disabled={isRevoking === session.id}
              class="inline-flex h-9 items-center gap-2 rounded-md border border-destructive/20 px-3 text-xs font-bold text-destructive transition-colors hover:bg-destructive hover:text-destructive-foreground disabled:opacity-50"
            >
              {#if isRevoking === session.id}
                <RefreshCcw size={14} class="animate-spin" />
              {:else}
                <Trash2 size={14} />
              {/if}
              <span class="hidden sm:inline">Revoke Access</span>
            </button>
          </div>
        </div>
      {/each}
    </div>
  {/if}

  <div class="rounded-xl border bg-muted/30 p-8 space-y-4">
    <div class="flex items-center gap-3 text-foreground font-bold">
      <Shield size={20} />
      <h2>Security & Access</h2>
    </div>
    <p class="text-sm text-muted-foreground leading-relaxed">
      Each device or application connected to Devitri uses a unique access key. If you lose a device or suspect a key has been compromised, you should revoke its access immediately. Revoking access will invalidate the token and stop all synchronization from that device.
    </p>
    <div class="pt-2">
      <a 
        href="/connect"
        data-sveltekit-reload
        class="inline-flex items-center gap-2 text-xs font-bold uppercase tracking-widest text-foreground underline underline-offset-4"
      >
        Connect a new device
      </a>
    </div>
  </div>
</div>
