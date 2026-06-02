<script lang="ts">
  import { Folder as FolderIcon, File as FileIcon, ChevronRight } from 'lucide-svelte';
  import { cn } from '$lib/utils';

  export interface Folder {
    id: string;
    name: string;
    path: string;
    children: (Folder | File)[];
  }
  
  export interface File {
    id: string;
    name: string;
    path: string;
    content: string;
    hash: string;
    modified_at: string;
  }
  
  interface Props {
    folders: Folder[];
    selectedFile?: File | null;
    onFileSelect?: (file: File | null) => void;
  }

  let { 
    folders = [], 
    selectedFile = null, 
    onFileSelect = () => {} 
  }: Props = $props();
  
  let selectionPath = $state<string[]>([]);
  let container = $state<HTMLDivElement>();
  
  let columns = $derived.by(() => {
    const root = Array.isArray(folders) ? folders : [];
    const cols: (Folder | File)[][] = [root];
    let currentLevel: (Folder | File)[] = root;

    for (const segment of selectionPath) {
      const selectedFolder = currentLevel.find(
        (item) => 'children' in item && item.name === segment
      ) as Folder | undefined;
      const children = selectedFolder?.children;
      if (!Array.isArray(children)) {
        break;
      }
      cols.push(children);
      currentLevel = children;
    }

    return cols.map((col) => sortColumnItems(col)).filter((col) => Array.isArray(col));
  });

  // Auto-scroll to the rightmost column after selection or folder tree updates
  $effect(() => {
    selectionPath;
    const columnCount = columns.length;
    const el = container;
    if (!el || columnCount === 0) return;

    requestAnimationFrame(() => {
      if (!el.isConnected) return;
      el.scrollTo({
        left: el.scrollWidth,
        behavior: 'smooth',
      });
    });
  });

  function sortColumnItems(items: (Folder | File)[] | undefined): (Folder | File)[] {
    if (!items) return [];
    return [...items].sort((a, b) => {
      const aIsFolder = 'children' in a;
      const bIsFolder = 'children' in b;
      if (aIsFolder !== bIsFolder) {
        return aIsFolder ? -1 : 1;
      }
      return a.name.localeCompare(b.name, undefined, { sensitivity: 'base' });
    });
  }

  // Clear stale drill-down when the tree is replaced (e.g. background manifest refresh)
  $effect(() => {
    folders;
    selectionPath = [];
  });

  function handleItemClick(item: Folder | File, level: number) {
    if ('children' in item) {
      const newPath = selectionPath.slice(0, level);
      newPath[level] = item.name;
      selectionPath = newPath;
      onFileSelect(null);
    } else {
      // Keep only columns up to this file; drop any folder drill-down to the right
      selectionPath = selectionPath.slice(0, level);
      onFileSelect(item);
    }
  }

  function isSelected(item: Folder | File, level: number) {
    if ('children' in item) {
      return selectionPath[level] === item.name;
    }
    return selectedFile?.path === item.path;
  }
</script>

<div 
  bind:this={container}
  class="flex h-full w-full overflow-x-auto overflow-y-hidden border rounded-lg bg-background shadow-sm"
>
  {#each columns as columnItems, level}
    <div class="flex h-full w-64 flex-col border-r last:border-r-0 last:flex-1 min-w-[200px]">
      <div class="flex h-10 items-center border-b bg-muted/50 px-4 text-[10px] font-bold uppercase tracking-widest text-muted-foreground truncate">
        {level === 0 ? 'Vault' : selectionPath[level - 1]}
      </div>
      
      <div class="flex-1 overflow-y-auto p-1 space-y-1">
        {#each columnItems ?? [] as item (item.path)}
          <button 
            type="button"
            class={cn(
              "group flex w-full items-center gap-3 rounded-md px-3 py-2 text-sm transition-all hover:bg-muted focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring",
              isSelected(item, level) && "bg-primary text-primary-foreground hover:bg-primary/90"
            )}
            onclick={() => handleItemClick(item, level)}
          >
            <span class="shrink-0 opacity-70">
              {#if 'children' in item}
                <FolderIcon size={16} />
              {:else}
                <FileIcon size={16} />
              {/if}
            </span>
            <span class="flex-1 truncate text-left">{item.name}</span>
            {#if 'children' in item}
              <ChevronRight size={14} class={cn("shrink-0 opacity-40", isSelected(item, level) && "opacity-100")} />
            {/if}
          </button>
        {/each}
      </div>
    </div>
  {/each}
</div>
