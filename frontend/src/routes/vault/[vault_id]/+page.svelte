<script lang="ts">
  import type { VaultResponse } from '../../../lib/api/client';
  import { ApiClient } from '../../../lib/api/client';
  import MillerColumns from '../../../lib/components/organisms/MillerColumns.svelte';
  import FileViewer from '../../../lib/components/organisms/FileViewer.svelte';
  import { ChevronLeft, HardDrive, Info, RefreshCcw } from 'lucide-svelte';
  import { serverStatus } from '$lib/stores/server.svelte';
  import { authStore } from '$lib/stores/auth.svelte';
  import { page } from '$app/state';
  import { navigateTo } from '$lib/utils/navigation';
  import { isImageFileName } from '$lib/utils/files';
  import { resolveObsidianImageEmbeds } from '$lib/utils/markdown-embeds';
  import { startVaultPoll } from '$lib/utils/vault-poll';
  import { hasVaultSyncChanged, manifestFingerprint } from '$lib/utils/vault-sync';
  
  const apiClient = new ApiClient();
  const POLL_INTERVAL_MS = 10_000;
  
  interface Props {
    data: { vault_id?: string };
  }

  let { data }: Props = $props();
  const vaultId = $derived(data.vault_id ?? page.params.vault_id ?? '');
  let vault = $state<VaultResponse | null>(null);
  let manifest = $state<any>(null);
  let folders = $state<any[]>([]);
  let selectedFile = $state<any>(null);
  let isLoading = $state(false);
  let error = $state<string | null>(null);
  let knownLastSync = $state<number | undefined>(undefined);
  let knownManifestFingerprint = $state('');
  let blobUrls: string[] = [];

  function revokeBlobUrls(): void {
    for (const url of blobUrls) {
      URL.revokeObjectURL(url);
    }
    blobUrls = [];
  }

  function getVaultFilePaths(): string[] {
    if (!manifest?.files || !Array.isArray(manifest.files)) return [];
    return manifest.files.map((f: { path: string }) => f.path);
  }

  async function enrichMarkdown(markdown: string, notePath: string): Promise<string> {
    const { markdown: enriched, blobUrls: newUrls } = await resolveObsidianImageEmbeds(
      markdown,
      notePath,
      getVaultFilePaths(),
      (path) => apiClient.getFileBlob(vaultId, path)
    );
    blobUrls.push(...newUrls);
    return enriched;
  }
  
  function applyManifest(nextManifest: { files?: Array<{ path: string; hash: string; modified_at?: number }> }): void {
    manifest = nextManifest;
    const files = Array.isArray(nextManifest?.files) ? nextManifest.files : [];
    knownManifestFingerprint = manifestFingerprint(files);
    folders = buildTree(files);
  }

  async function loadManifest(): Promise<void> {
    if (!vaultId) return;
    applyManifest(await apiClient.getManifest(vaultId));
  }

  async function fetchVaultData(showLoading = true): Promise<void> {
    if (!vaultId) {
      error = 'Missing vault ID in URL';
      return;
    }

    if (showLoading) {
      isLoading = true;
    }
    error = null;
    try {
      vault = await apiClient.getVault(vaultId);
      knownLastSync = vault.last_sync;
      await loadManifest();
    } catch (err) {
      error = err instanceof Error ? err.message : 'An unknown error occurred';
    } finally {
      if (showLoading) {
        isLoading = false;
      }
    }
  }

  async function loadFileContent(file: { path: string; name: string; hash?: string }): Promise<void> {
    if (isImageFileName(file.name)) {
      const blob = await apiClient.getFileBlob(vaultId, file.path);
      const url = URL.createObjectURL(blob);
      blobUrls.push(url);
      selectedFile = { ...file, content: url };
      return;
    }

    let content = await apiClient.getFileContent(vaultId, file.path);
    if (/\.md$/i.test(file.name)) {
      content = await enrichMarkdown(content, file.path);
    }
    selectedFile = { ...file, content };
  }

  async function refreshOpenFileIfChanged(
    files: Array<{ path: string; hash: string; modified_at?: number }>
  ): Promise<void> {
    const open = selectedFile;
    if (!open?.path || open.children) return;

    const entry = files.find((f) => f.path === open.path);
    if (!entry) {
      selectedFile = null;
      return;
    }
    if (entry.hash === open.hash) return;

    revokeBlobUrls();
    selectedFile = { ...open, content: '', hash: entry.hash, modified_at: entry.modified_at };
    try {
      await loadFileContent({ ...open, hash: entry.hash, modified_at: entry.modified_at });
    } catch (err) {
      console.error('Failed to refresh file content:', err);
    }
  }

  async function pollVaultIfSyncChanged(): Promise<void> {
    if (!vaultId) return;

    try {
      const [nextVault, nextManifest] = await Promise.all([
        apiClient.getVault(vaultId),
        apiClient.getManifest(vaultId),
      ]);

      const files = Array.isArray(nextManifest?.files) ? nextManifest.files : [];
      const nextFingerprint = manifestFingerprint(files);
      const syncChanged = hasVaultSyncChanged(knownLastSync, nextVault);
      const manifestChanged = nextFingerprint !== knownManifestFingerprint;

      if (!syncChanged && !manifestChanged) {
        return;
      }

      knownLastSync = nextVault.last_sync;
      vault = nextVault;
      applyManifest(nextManifest);
      await refreshOpenFileIfChanged(files);
    } catch {
      // Silent background poll
    }
  }

  function buildTree(files: any[]): any[] {
    const root: any = { id: 'root', name: 'Vault', path: '', children: [] };
    const nodeMap: Record<string, any> = { '': root };

    const sorted = [...files]
      .filter((file) => typeof file?.path === 'string' && file.path.length > 0)
      .sort((a, b) => a.path.length - b.path.length);

    for (const file of sorted) {
      const parts = file.path.split('/').filter(Boolean);
      let currentPath = '';

      for (let index = 0; index < parts.length; index++) {
        const part = parts[index];
        const parentPath = currentPath;
        currentPath = currentPath ? `${currentPath}/${part}` : part;
        const isLast = index === parts.length - 1;

        ensureParentChildren(nodeMap, parentPath);

        let node = nodeMap[currentPath];
        if (!node) {
          node = {
            id: currentPath,
            name: part,
            path: currentPath,
          };

          if (isLast) {
            node.hash = file.hash;
            node.modified_at = file.modified_at;
            node.content = '';
          } else {
            node.children = [];
          }

          nodeMap[currentPath] = node;
          nodeMap[parentPath].children.push(node);
          continue;
        }

        if (isLast) {
          node.hash = file.hash;
          node.modified_at = file.modified_at;
          node.content = node.content ?? '';
        } else if (!node.children) {
          // Path was a file; a deeper path forces it to act as a folder
          node.children = [];
          delete node.hash;
          delete node.modified_at;
          delete node.content;
        }
      }
    }

    return root.children ?? [];
  }

  function ensureParentChildren(
    nodeMap: Record<string, any>,
    parentPath: string
  ): void {
    const parent = nodeMap[parentPath];
    if (!parent) return;
    if (!Array.isArray(parent.children)) {
      parent.children = [];
    }
  }
  
  async function handleFileSelect(file: any | null): Promise<void> {
    revokeBlobUrls();

    if (!file) {
      selectedFile = null;
      return;
    }

    selectedFile = { ...file, content: '' };

    if (file.children) {
      return;
    }

    try {
      await loadFileContent(file);
    } catch (err) {
      console.error('Failed to fetch file content:', err);
      selectedFile = { ...file, content: '' };
    }
  }
  
  $effect(() => {
    if (serverStatus.isLoading || authStore.isLoading) return;

    if (!serverStatus.ready) {
      navigateTo('/setup');
      return;
    }
    if (!authStore.isAuthenticated) {
      navigateTo('/login');
      return;
    }
    if (!vaultId) return;

    void fetchVaultData(true);

    const stopPoll = startVaultPoll({
      intervalMs: POLL_INTERVAL_MS,
      runOnStart: true,
      tick: pollVaultIfSyncChanged,
    });

    return stopPoll;
  });
</script>

<div class="flex flex-col gap-6 lg:gap-8">
  <header class="flex flex-col gap-4 border-b pb-6">
    <a
      href="/"
      data-sveltekit-reload
      class="inline-flex items-center gap-2 text-xs font-bold uppercase tracking-widest text-muted-foreground hover:text-foreground transition-colors w-fit"
    >
      <ChevronLeft size={14} />
      <span>Back to Dashboard</span>
    </a>

    <div class="flex flex-col gap-4 sm:flex-row sm:items-end sm:justify-between">
    <div class="space-y-1 min-w-0">
      <div class="flex items-center gap-2 text-muted-foreground uppercase tracking-widest text-[10px] font-bold">
        <HardDrive size={12} />
        <span>Vault Explorer</span>
      </div>
      <h1 class="font-display italic text-3xl font-bold tracking-tight">{vault?.name ?? 'Loading...'}</h1>
    </div>
    
    {#if vault}
      <div class="flex gap-8 font-mono text-[10px] uppercase tracking-wider text-muted-foreground">
        <div class="flex flex-col items-end gap-1">
          <span class="text-foreground font-bold">{vault.stats.files}</span>
          <span>Files</span>
        </div>
        <div class="flex flex-col items-end gap-1">
          <span class="text-foreground font-bold">{vault.stats.folders}</span>
          <span>Folders</span>
        </div>
        <div class="flex flex-col items-end gap-1">
          <span class="text-foreground font-bold">{(vault.stats.size_bytes / (1024 * 1024)).toFixed(1)}MB</span>
          <span>Size</span>
        </div>
      </div>
    {/if}
    </div>
  </header>
  
  {#if isLoading}
    <div class="flex min-h-[50dvh] lg:h-[600px] flex-col items-center justify-center gap-4 rounded-xl border border-dashed text-muted-foreground">
      <RefreshCcw class="animate-spin" size={32} />
      <p class="text-xs font-mono uppercase tracking-widest">Analyzing vault structure...</p>
    </div>
  {:else if error}
    <div class="flex min-h-[50dvh] lg:h-[600px] flex-col items-center justify-center gap-4 rounded-xl border border-dashed border-destructive p-8 text-center">
      <p class="text-destructive font-bold">Error loading vault</p>
      <p class="text-xs text-muted-foreground">{error}</p>
      <button onclick={() => fetchVaultData(true)} class="mt-4 text-xs font-bold underline">Try again</button>
    </div>
  {:else}
    <div class="flex flex-col gap-4 max-lg:overflow-visible lg:grid lg:grid-cols-12 lg:gap-6 lg:h-[700px]">
      <section
        class="h-[min(36dvh,16rem)] shrink-0 overflow-hidden lg:col-span-5 lg:h-full lg:min-h-0"
        aria-label="Vault file tree"
      >
        <MillerColumns 
          folders={folders}
          selectedFile={selectedFile}
          onFileSelect={handleFileSelect}
        />
      </section>
      
      <section
        class="flex flex-col rounded-lg border bg-muted/10 shadow-inner max-lg:overflow-visible lg:col-span-7 lg:h-full lg:min-h-0 lg:overflow-hidden"
        aria-label="File preview"
      >
        {#if selectedFile}
          <FileViewer 
            content={selectedFile.content} 
            fileName={selectedFile.name} 
            filePath={selectedFile.path} 
          />
        {:else}
          <div class="flex flex-col items-center justify-center gap-4 py-16 text-center text-muted-foreground lg:min-h-0 lg:h-full lg:py-8">
            <div class="rounded-full bg-muted p-4">
              <Info size={32} strokeWidth={1.5} />
            </div>
            <div class="space-y-1">
              <p class="font-sans font-bold text-foreground">No file selected</p>
              <p class="text-xs max-w-[240px]">Navigate through the columns on the left and select a file to view its content here.</p>
            </div>
          </div>
        {/if}
      </section>
    </div>
  {/if}
</div>
