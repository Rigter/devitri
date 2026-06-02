<script lang="ts">
  import { onMount } from 'svelte';
  import { ApiClient } from '../../lib/api/client';
  import { serverStatus, refreshServerStatus } from '$lib/stores/server.svelte';
  import { toast } from '$lib/stores/toast.svelte';
  import { copyToClipboard } from '$lib/utils/clipboard';
  import { navigateTo } from '$lib/utils/navigation';
  import {
    CheckCircle2,
    Copy,
    Check,
    RefreshCcw,
    ShieldCheck,
    Info
  } from 'lucide-svelte';

  const apiClient = new ApiClient();

  let password = $state('');
  let showPassword = $state(false);
  let envBlock = $state<string | null>(null);
  let isGenerating = $state(false);
  let isPolling = $state(false);
  let pollSucceeded = $state(false);
  let copied = $state(false);
  let error = $state<string | null>(null);

  let pollingInterval: ReturnType<typeof setInterval> | null = null;

  function buildEnvBlock(hash: string, jwtSecret: string): string {
    return `DEVITRI_MASTER_HASH=${hash}\nDEVITRI_JWT_SECRET=${jwtSecret}`;
  }

  function goToLogin(): void {
    navigateTo('/login');
  }

  function scheduleLoginRedirect(): void {
    pollSucceeded = true;
    stopPolling();
    setTimeout(goToLogin, 1200);
  }

  async function pollServerStatus(showReadyToast = false): Promise<boolean> {
    const status = await refreshServerStatus();
    if (status?.ready) {
      if (showReadyToast) toast.success('Server is ready');
      scheduleLoginRedirect();
      return true;
    }
    return false;
  }

  async function handleGenerate(event: SubmitEvent) {
    event.preventDefault();
    error = null;

    if (password.length < 8) {
      error = 'Password must be at least 8 characters';
      return;
    }

    isGenerating = true;
    try {
      const { hash, jwt_secret } = await apiClient.generateSetupConfig(password);
      envBlock = buildEnvBlock(hash, jwt_secret);
      toast.success('Configuration generated');
    } catch (e) {
      error = e instanceof Error ? e.message : 'Failed to generate configuration';
      toast.error(error);
    } finally {
      isGenerating = false;
    }
  }

  async function copyEnvBlock() {
    if (!envBlock) return;
    const ok = await copyToClipboard(envBlock, '.env block copied');
    copied = ok;
    if (ok) setTimeout(() => { copied = false; }, 2000);
  }

  function startPolling() {
    stopPolling();
    pollSucceeded = false;
    isPolling = true;
    toast.info('Waiting for server restart…');

    void pollServerStatus(true).then((ready) => {
      if (ready) return;
      pollingInterval = setInterval(() => {
        void pollServerStatus(true);
      }, 3000);
    });
  }

  function stopPolling() {
    if (pollingInterval !== null) {
      clearInterval(pollingInterval);
      pollingInterval = null;
    }
    isPolling = false;
  }

  onMount(() => {
    void refreshServerStatus();
    return () => stopPolling();
  });
</script>

<div class="mx-auto max-w-2xl py-12">
  <header class="mb-12 space-y-2 text-center">
    <h1 class="font-display text-4xl font-bold italic tracking-tight">Setup</h1>
    <p class="font-sans text-muted-foreground">Choose your master password and copy the generated <code class="rounded bg-muted px-1">.env</code> values.</p>
  </header>

  <div class="overflow-hidden rounded-xl border bg-card shadow-lg">
    {#if serverStatus.isLoading}
      <div class="flex flex-col items-center justify-center gap-4 p-20">
        <RefreshCcw class="animate-spin text-muted-foreground" size={32} />
        <p class="font-mono text-sm uppercase tracking-widest text-muted-foreground">Checking status…</p>
      </div>
    {:else if isPolling || pollSucceeded}
      <div class="space-y-6 p-12 text-center">
        {#if pollSucceeded}
          <div class="mb-4 inline-flex h-20 w-20 items-center justify-center rounded-full bg-emerald-500/10 text-emerald-500">
            <CheckCircle2 size={40} />
          </div>
          <h2 class="text-xl font-bold">Server is ready</h2>
          <p class="text-sm text-muted-foreground">Redirecting to sign in…</p>
          <button
            type="button"
            onclick={goToLogin}
            class="inline-flex h-11 items-center justify-center rounded-md bg-foreground px-8 text-sm font-bold text-background transition-opacity hover:opacity-90"
          >
            Continue now
          </button>
        {:else}
          <div class="relative mb-4 inline-flex items-center justify-center">
            <div class="absolute inset-0 animate-spin rounded-full border-4 border-muted/30 border-t-foreground"></div>
            <div class="flex h-20 w-20 items-center justify-center rounded-full bg-muted/30">
              <RefreshCcw size={32} class="text-muted-foreground" />
            </div>
          </div>
          <h2 class="text-xl font-bold">Waiting for server…</h2>
          <p class="mx-auto max-w-xs font-mono text-sm uppercase tracking-tight text-muted-foreground">
            Checking every 3 seconds
          </p>
          <button
            type="button"
            onclick={() => void pollServerStatus(true)}
            class="text-xs font-bold uppercase tracking-widest underline underline-offset-4"
          >
            Check now
          </button>
        {/if}
      </div>
    {:else if serverStatus.ready}
      <div class="space-y-6 p-12 text-center">
        <div class="mb-4 inline-flex h-20 w-20 items-center justify-center rounded-full bg-emerald-500/10 text-emerald-500">
          <CheckCircle2 size={40} />
        </div>
        <h2 class="text-2xl font-bold tracking-tight">Server is ready</h2>
        <p class="mx-auto max-w-sm text-muted-foreground">Configuration detected. Sign in with your master password.</p>
        <button
          type="button"
          onclick={goToLogin}
          class="inline-flex h-11 items-center justify-center rounded-md bg-foreground px-8 text-sm font-bold text-background transition-opacity hover:opacity-90"
        >
          Go to Sign In
        </button>
      </div>
    {:else}
      <div class="space-y-6 p-8">
        {#if !envBlock}
          <form onsubmit={handleGenerate} class="space-y-6">
            <p class="text-sm leading-relaxed text-muted-foreground">
              This password protects your Devitri instance. It is <strong class="text-foreground">not</strong> stored on the server — only a bcrypt hash goes in your <code class="rounded bg-muted px-1">.env</code>.
            </p>

            <div class="space-y-2">
              <label for="password" class="text-xs font-bold uppercase tracking-widest text-muted-foreground">
                Master password
              </label>
              <div class="relative">
                <input
                  id="password"
                  type={showPassword ? 'text' : 'password'}
                  bind:value={password}
                  autocomplete="new-password"
                  minlength={8}
                  required
                  class="flex h-11 w-full rounded-md border bg-background px-3 pr-24 font-mono text-sm outline-none focus-visible:ring-2 focus-visible:ring-muted"
                  placeholder="At least 8 characters"
                />
                <button
                  type="button"
                  onclick={() => { showPassword = !showPassword; }}
                  class="absolute right-2 top-1/2 -translate-y-1/2 rounded px-2 py-1 text-[10px] font-bold uppercase tracking-wider text-muted-foreground hover:text-foreground"
                >
                  {showPassword ? 'Hide' : 'Show'}
                </button>
              </div>
            </div>

            {#if error}
              <p class="text-sm text-destructive">{error}</p>
            {/if}

            <button
              type="submit"
              disabled={isGenerating || password.length < 8}
              class="inline-flex h-11 w-full items-center justify-center gap-2 rounded-md bg-foreground text-sm font-bold text-background transition-opacity hover:opacity-90 disabled:opacity-50"
            >
              <ShieldCheck size={18} />
              {isGenerating ? 'Generating…' : 'Generate .env values'}
            </button>
          </form>
        {:else}
          <p class="text-sm text-muted-foreground">
            Paste these lines into your project <code class="rounded bg-muted px-1">.env</code>, then restart the backend container.
          </p>

          <div class="relative">
            <textarea
              readonly
              rows={4}
              class="w-full resize-none rounded-lg border bg-muted p-4 font-mono text-xs leading-relaxed"
              value={envBlock}
            ></textarea>
            <button
              type="button"
              onclick={copyEnvBlock}
              class="absolute right-2 top-2 rounded-md border bg-background p-2 shadow-sm transition-colors hover:bg-muted"
              aria-label="Copy .env block"
            >
              {#if copied}
                <Check size={14} class="text-emerald-500" />
              {:else}
                <Copy size={14} />
              {/if}
            </button>
          </div>

          <div class="flex gap-3 rounded-lg border border-amber-500/20 bg-amber-500/10 p-4">
            <Info size={18} class="mt-0.5 shrink-0 text-amber-500" />
            <p class="text-xs leading-relaxed text-amber-200/80">
              After updating <code class="font-bold">.env</code>, run:
              <code class="mt-1 block rounded bg-background/50 px-2 py-1 font-mono text-[10px]">docker compose -f deploy/dev/docker-compose.yml up -d --force-recreate backend</code>
            </p>
          </div>

          <div class="flex flex-col gap-3 sm:flex-row">
            <button
              type="button"
              onclick={startPolling}
              class="inline-flex h-11 flex-1 items-center justify-center rounded-md bg-foreground text-sm font-bold text-background transition-opacity hover:opacity-90"
            >
              I've updated .env — wait for server
            </button>
            <button
              type="button"
              onclick={() => { envBlock = null; password = ''; }}
              class="inline-flex h-11 items-center justify-center rounded-md border px-4 text-sm font-bold transition-colors hover:bg-muted"
            >
              Start over
            </button>
          </div>
        {/if}
      </div>
    {/if}
  </div>
</div>
