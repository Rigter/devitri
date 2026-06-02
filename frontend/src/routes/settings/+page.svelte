<script lang="ts">
  import { ApiClient, type SettingsResponse } from '$lib/api/client';
  import { authStore } from '$lib/stores/auth.svelte';
  import { navigateTo } from '$lib/utils/navigation';
  import { ChevronLeft, RefreshCcw, AlertCircle, Shield, SlidersHorizontal, Info } from 'lucide-svelte';

  const apiClient = new ApiClient();

  let settings = $state<SettingsResponse | null>(null);
  let isLoading = $state(true);
  let error = $state<string | null>(null);

  async function fetchSettings() {
    isLoading = true;
    error = null;
    try {
      settings = await apiClient.getSettings();
    } catch (err) {
      error = err instanceof Error ? err.message : 'Failed to load settings';
    } finally {
      isLoading = false;
    }
  }

  function formatConfigured(value: boolean): string {
    return value ? 'Configured' : 'Not set';
  }

  $effect(() => {
    if (authStore.isLoading) return;
    if (!authStore.isAuthenticated) {
      navigateTo('/login');
    } else {
      void fetchSettings();
    }
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
        <h1 class="font-display italic text-4xl font-bold tracking-tight">Settings</h1>
        <p class="text-muted-foreground font-sans">Read-only view of the platform configuration loaded from your environment.</p>
      </div>
      <button
        type="button"
        onclick={fetchSettings}
        class="inline-flex h-10 w-10 items-center justify-center rounded-md border border-border bg-background transition-colors hover:bg-muted"
        title="Refresh settings"
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
  {:else if isLoading && !settings}
    <div class="grid gap-4">
      {#each Array(3) as _}
        <div class="h-32 rounded-xl border border-dashed animate-pulse bg-muted/20"></div>
      {/each}
    </div>
  {:else if settings}
    <section class="rounded-xl border bg-card shadow-sm overflow-hidden">
      <div class="flex items-center gap-3 border-b px-6 py-4">
        <Shield size={18} />
        <h2 class="font-sans text-sm font-bold uppercase tracking-widest">Security</h2>
      </div>
      <dl class="divide-y">
        <div class="grid gap-1 px-6 py-4 sm:grid-cols-[1fr_auto] sm:gap-4">
          <dt class="text-sm text-muted-foreground">Dashboard session TTL</dt>
          <dd class="font-mono text-sm">{settings.security.session_ttl_hours} hours <span class="text-muted-foreground">· DEVITRI_SESSION_TTL_HOURS</span></dd>
        </div>
        <div class="grid gap-1 px-6 py-4 sm:grid-cols-[1fr_auto] sm:gap-4">
          <dt class="text-sm text-muted-foreground">Device token TTL</dt>
          <dd class="font-mono text-sm">{settings.security.device_token_ttl_days} days <span class="text-muted-foreground">· DEVITRI_DEVICE_TOKEN_TTL_DAYS</span></dd>
        </div>
        <div class="grid gap-1 px-6 py-4 sm:grid-cols-[1fr_auto] sm:gap-4">
          <dt class="text-sm text-muted-foreground">Login rate limit</dt>
          <dd class="font-mono text-sm">{settings.security.login_rate_limit_per_minute} / min <span class="text-muted-foreground">· DEVITRI_LOGIN_RATE_LIMIT</span></dd>
        </div>
        <div class="grid gap-1 px-6 py-4 sm:grid-cols-[1fr_auto] sm:gap-4">
          <dt class="text-sm text-muted-foreground">Bcrypt cost</dt>
          <dd class="font-mono text-sm">{settings.security.bcrypt_cost}</dd>
        </div>
        <div class="grid gap-1 px-6 py-4 sm:grid-cols-[1fr_auto] sm:gap-4">
          <dt class="text-sm text-muted-foreground">Master password hash</dt>
          <dd class="font-mono text-sm">{formatConfigured(settings.security.master_hash_configured)} <span class="text-muted-foreground">· DEVITRI_MASTER_HASH</span></dd>
        </div>
        <div class="grid gap-1 px-6 py-4 sm:grid-cols-[1fr_auto] sm:gap-4">
          <dt class="text-sm text-muted-foreground">JWT secret</dt>
          <dd class="font-mono text-sm">{formatConfigured(settings.security.jwt_secret_configured)} <span class="text-muted-foreground">· DEVITRI_JWT_SECRET</span></dd>
        </div>
      </dl>
    </section>

    <section class="rounded-xl border bg-card shadow-sm overflow-hidden">
      <div class="flex items-center gap-3 border-b px-6 py-4">
        <SlidersHorizontal size={18} />
        <h2 class="font-sans text-sm font-bold uppercase tracking-widest">Sync</h2>
      </div>
      <dl class="divide-y">
        <div class="grid gap-1 px-6 py-4 sm:grid-cols-[1fr_auto] sm:gap-4">
          <dt class="text-sm text-muted-foreground">Bulk delete file threshold</dt>
          <dd class="font-mono text-sm">{settings.sync.delete_threshold_count} files <span class="text-muted-foreground">· DEVITRI_DELETE_THRESHOLD_COUNT</span></dd>
        </div>
        <div class="grid gap-1 px-6 py-4 sm:grid-cols-[1fr_auto] sm:gap-4">
          <dt class="text-sm text-muted-foreground">Bulk delete percent threshold</dt>
          <dd class="font-mono text-sm">{settings.sync.delete_threshold_percent}% <span class="text-muted-foreground">· DEVITRI_DELETE_THRESHOLD_PERCENT</span></dd>
        </div>
      </dl>
    </section>

    <section class="rounded-xl border bg-card shadow-sm overflow-hidden">
      <div class="flex items-center gap-3 border-b px-6 py-4">
        <Info size={18} />
        <h2 class="font-sans text-sm font-bold uppercase tracking-widest">Operational</h2>
      </div>
      <dl class="divide-y">
        <div class="grid gap-1 px-6 py-4 sm:grid-cols-[1fr_auto] sm:gap-4">
          <dt class="text-sm text-muted-foreground">Timezone</dt>
          <dd class="font-mono text-sm">{settings.operational.tz} <span class="text-muted-foreground">· TZ / DEVITRI_TZ</span></dd>
        </div>
        <div class="grid gap-1 px-6 py-4 sm:grid-cols-[1fr_auto] sm:gap-4">
          <dt class="text-sm text-muted-foreground">Process user / group</dt>
          <dd class="font-mono text-sm">{settings.operational.puid}:{settings.operational.pgid} <span class="text-muted-foreground">· PUID/PGID or DEVITRI_*</span></dd>
        </div>
      </dl>
    </section>

    <div class="rounded-xl border bg-muted/30 p-8 space-y-4">
      <div class="flex items-center gap-3 text-foreground font-bold">
        <Info size={20} />
        <h2>How to change settings</h2>
      </div>
      <p class="text-sm text-muted-foreground leading-relaxed">
        Values are read from your <code class="rounded bg-muted px-1">.env</code> file at container startup. Edit the variables above in your project <code class="rounded bg-muted px-1">.env</code>, then restart the backend container for changes to take effect.
      </p>
      <code class="block rounded-lg border bg-background/50 px-4 py-3 font-mono text-[10px] leading-relaxed">
        docker compose -f deploy/dev/docker-compose.yml up -d --force-recreate backend
      </code>
    </div>
  {/if}
</div>
