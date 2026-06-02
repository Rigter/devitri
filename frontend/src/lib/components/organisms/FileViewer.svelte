<script lang="ts">
  import { FileText, Calendar, Hash, Image as ImageIcon, FileCode, File as FileGeneric } from 'lucide-svelte';
  import { isImageFileName } from '$lib/utils/files';
  import { renderObsidianMarkdown } from '$lib/utils/obsidian-markdown';
  import '$lib/styles/devitri-markdown.css';

  interface Props {
    content?: string;
    fileName?: string;
    filePath?: string;
  }

  let {
    content = '',
    fileName = '',
    filePath = '',
  }: Props = $props();

  const isImage = $derived(isImageFileName(fileName) || content.startsWith('blob:'));
  const isMarkdown = $derived(/\.md$/i.test(fileName));
  const isCode = $derived(/\.(js|ts|json|go|css|html|yaml|yml)$/i.test(fileName));

  let htmlContent = $derived(isMarkdown ? renderObsidianMarkdown(content) : '');
</script>

<!-- Mobile: content flows with page scroll. Desktop: fixed panel with internal scroll. -->
<div class="flex flex-col bg-background lg:h-full lg:min-h-0 lg:overflow-hidden">
  {#if fileName}
    <div
      class="sticky top-14 z-10 flex shrink-0 flex-col gap-2 border-b bg-muted/30 p-4 backdrop-blur-sm supports-[backdrop-filter]:bg-muted/80 md:p-6 lg:static lg:top-auto lg:z-auto lg:bg-muted/30 lg:backdrop-blur-none"
    >
      <div class="flex items-center gap-3 min-w-0">
        <div class="shrink-0 rounded-md bg-primary/10 p-2 text-primary">
          {#if isImage}
            <ImageIcon size={20} />
          {:else if isMarkdown}
            <FileText size={20} />
          {:else if isCode}
            <FileCode size={20} />
          {:else}
            <FileGeneric size={20} />
          {/if}
        </div>
        <h2 class="truncate text-lg font-bold tracking-tight md:text-xl">{fileName}</h2>
      </div>
      <div class="flex flex-wrap items-center gap-4 text-[10px] uppercase tracking-wider text-muted-foreground">
        <div class="flex min-w-0 items-center gap-1.5">
          <Hash size={12} class="shrink-0" />
          <span class="truncate">{filePath}</span>
        </div>
      </div>
    </div>
  {/if}

  <div class="p-4 pb-12 md:p-8 lg:min-h-0 lg:flex-1 lg:overflow-y-auto lg:overscroll-contain lg:p-12">
    {#if isImage}
      <div class="flex items-center justify-center p-4 lg:min-h-[200px] lg:h-full">
        <img
          src={content}
          alt={fileName}
          class="max-h-[70dvh] max-w-full object-contain rounded-lg border border-border shadow-lg lg:max-h-full"
        />
      </div>
    {:else if isMarkdown}
      <!-- htmlContent is sanitized in renderObsidianMarkdown (allowlist DOM sanitizer) -->
      <article class="devitri-md max-w-none selection:bg-primary selection:text-primary-foreground">
        {@html htmlContent}
      </article>
    {:else}
      <pre
        class="bg-muted p-6 rounded-lg font-mono text-sm leading-relaxed overflow-x-auto border border-border"
      ><code>{content}</code></pre>
    {/if}
  </div>
</div>
