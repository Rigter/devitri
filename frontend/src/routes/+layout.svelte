<script lang="ts">
  import { onMount } from 'svelte';
  import type { Snippet } from 'svelte';
  import { Sun, Moon, LayoutDashboard, Settings, LogOut, Smartphone, Plug, SlidersHorizontal } from 'lucide-svelte';
  import { serverStatus, refreshServerStatus } from '$lib/stores/server.svelte';
  import { authStore, initAuth } from '$lib/stores/auth.svelte';
  import { page } from '$app/state';
  import { navigateTo } from '$lib/utils/navigation';
  import Toaster from '$lib/components/organisms/Toaster.svelte';
  import '../lib/styles/themes/nano-zinc.css';

  interface Props {
    children: Snippet;
  }

  let { children }: Props = $props();
  let darkMode = $state(true);

  function toggleTheme(): void {
    darkMode = !darkMode;
    if (typeof document !== 'undefined') {
      document.documentElement.classList.toggle('light', !darkMode);
    }
  }

  async function handleLogout() {
    await authStore.logout();
    navigateTo('/login');
  }

  function isActive(href: string): boolean {
    const path = page.url.pathname;
    if (href === '/') return path === '/' || path.startsWith('/vault/');
    return path === href || path.startsWith(`${href}/`);
  }

  function navClass(href: string): string {
    const base = 'flex items-center space-x-2 transition-colors';
    return isActive(href) ? `${base} font-bold text-foreground` : `${base} text-muted-foreground hover:text-foreground`;
  }

  onMount(() => {
    void refreshServerStatus();
    void initAuth();
    const prefersLight = window.matchMedia('(prefers-color-scheme: light)').matches;
    darkMode = !prefersLight;
    document.documentElement.classList.toggle('light', !darkMode);
  });
</script>

<div class="flex min-h-screen flex-col">
  <header class="sticky top-0 z-50 w-full border-b bg-background/95 backdrop-blur supports-[backdrop-filter]:bg-background/60">
    <div class="container mx-auto flex h-14 max-w-7xl items-center px-4 md:px-8">
      <div class="mr-8 flex items-center space-x-0">
        <span class="font-display italic font-bold text-xl">d</span>
        <span class="font-logo text-xl uppercase tracking-tighter">evitri</span>
      </div>
      
      <nav class="flex flex-1 items-center space-x-6 text-sm font-medium">
        {#if serverStatus.ready && authStore.isAuthenticated}
          <a href="/" data-sveltekit-reload class={navClass('/')}>
            <LayoutDashboard size={16} />
            <span>Dashboard</span>
          </a>
          <a href="/connect" data-sveltekit-reload class={navClass('/connect')}>
            <Plug size={16} />
            <span>Connect</span>
          </a>
          <a href="/devices" data-sveltekit-reload class={navClass('/devices')}>
            <Smartphone size={16} />
            <span>Devices</span>
          </a>
          <a href="/settings" data-sveltekit-reload class={navClass('/settings')}>
            <SlidersHorizontal size={16} />
            <span>Settings</span>
          </a>
        {:else if !serverStatus.isLoading && !serverStatus.ready}
          <a href="/setup" data-sveltekit-reload class={navClass('/setup')}>
            <Settings size={16} />
            <span>Setup</span>
          </a>
        {/if}
      </nav>

      <div class="flex items-center space-x-2">
        {#if authStore.isAuthenticated}
          <button
            type="button"
            onclick={handleLogout}
            class="inline-flex h-9 items-center gap-2 rounded-md border border-border px-3 text-xs font-bold transition-colors hover:bg-muted"
          >
            <LogOut size={14} />
            <span>Sign Out</span>
          </button>
        {/if}
        <button
          onclick={toggleTheme} 
          class="inline-flex h-9 w-9 items-center justify-center rounded-md border border-border bg-background transition-colors hover:bg-muted"
        >
          {#if darkMode}
            <Sun size={18} />
          {:else}
            <Moon size={18} />
          {/if}
          <span class="sr-only">Toggle theme</span>
        </button>
      </div>
    </div>
  </header>

  <main class="flex-1 container mx-auto max-w-7xl px-4 py-8 md:px-8">
    {@render children()}
  </main>
  
  <footer class="border-t py-6 md:py-0">
    <div class="container mx-auto max-w-7xl flex flex-col items-center justify-between gap-4 md:h-16 md:flex-row px-4 md:px-8">
      <p class="text-center text-xs leading-loose text-muted-foreground md:text-left">
        © 2026 Devitri — Bidirectional Sync & Visualization System for Obsidian.
      </p>
    </div>
  </footer>
</div>

<Toaster />

<style>
  :global(html) {
    color-scheme: dark;
  }
  :global(html.light) {
    color-scheme: light;
  }
</style>
