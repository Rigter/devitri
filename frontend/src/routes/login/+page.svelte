<script lang="ts">
  import { onMount } from 'svelte';
  import { navigateTo } from '$lib/utils/navigation';
  import { authStore } from '$lib/stores/auth.svelte';
  import { serverStatus, refreshServerStatus } from '$lib/stores/server.svelte';
  import { toast } from '$lib/stores/toast.svelte';
  import { Lock, LogIn, RefreshCcw } from 'lucide-svelte';

  let password = $state('');
  let isSubmitting = $state(false);
  let error = $state<string | null>(null);

  async function confirmServerReady(): Promise<void> {
    while (serverStatus.isLoading) {
      await new Promise((resolve) => setTimeout(resolve, 50));
    }
    if (serverStatus.ready) return;

    for (let attempt = 0; attempt < 3; attempt++) {
      await new Promise((resolve) => setTimeout(resolve, 1500));
      const status = await refreshServerStatus();
      if (status?.ready) return;
    }
  }

  onMount(() => {
    void confirmServerReady();
  });

  async function handleLogin(event: SubmitEvent) {
    event.preventDefault();
    error = null;
    isSubmitting = true;

    try {
      await authStore.login(password);
      toast.success('Signed in');
      password = '';
      navigateTo('/');
    } catch (e) {
      error = e instanceof Error ? e.message : 'Login failed';
      toast.error(error);
    } finally {
      isSubmitting = false;
    }
  }
</script>

<div class="mx-auto flex min-h-[60vh] max-w-md flex-col justify-center py-12">
  <header class="mb-10 space-y-2 text-center">
    <div class="mx-auto mb-4 inline-flex h-14 w-14 items-center justify-center rounded-full border bg-muted/30">
      <Lock size={24} />
    </div>
    <h1 class="font-display text-3xl font-bold italic tracking-tight">Sign In</h1>
    <p class="text-sm text-muted-foreground">
      Enter your master password to access the dashboard.
    </p>
  </header>

  <form onsubmit={handleLogin} class="space-y-6 rounded-xl border bg-card p-8 shadow-lg">
    {#if serverStatus.isLoading || !serverStatus.checked}
      <div class="flex flex-col items-center gap-3 py-6 text-muted-foreground">
        <RefreshCcw class="animate-spin" size={24} />
        <p class="text-sm">Checking server status…</p>
      </div>
    {:else if !serverStatus.ready}
      <p class="text-sm text-muted-foreground">
        Server setup is incomplete. Finish configuration first.
      </p>
      <button
        type="button"
        onclick={() => navigateTo('/setup')}
        class="inline-flex h-11 w-full items-center justify-center rounded-md bg-foreground text-sm font-bold text-background"
      >
        Go to Setup
      </button>
    {:else}
      <div class="space-y-2">
        <label for="password" class="text-xs font-bold uppercase tracking-widest text-muted-foreground">
          Master Password
        </label>
        <input
          id="password"
          type="password"
          bind:value={password}
          autocomplete="current-password"
          required
          class="flex h-11 w-full rounded-md border bg-background px-3 font-mono text-sm outline-none ring-offset-background focus-visible:ring-2 focus-visible:ring-muted"
          placeholder="Your master password"
        />
      </div>

      {#if error}
        <p class="text-sm text-destructive">{error}</p>
      {/if}

      <button
        type="submit"
        disabled={isSubmitting || !password}
        class="inline-flex h-11 w-full items-center justify-center gap-2 rounded-md bg-foreground text-sm font-bold text-background transition-opacity hover:opacity-90 disabled:opacity-50"
      >
        <LogIn size={18} />
        {isSubmitting ? 'Signing in…' : 'Sign In'}
      </button>
    {/if}
  </form>
</div>
