import { isImageFileName, resolveNoteAssetPath } from './files';

export interface WikiEmbed {
  target: string;
  display?: string;
  width?: number;
}

/** Parse Obsidian wiki embed: ![[file.png]], ![[file.png|alt]], ![[file.png|200]] */
export function parseWikiEmbed(inner: string): WikiEmbed {
  const pipe = inner.indexOf('|');
  let target = (pipe >= 0 ? inner.slice(0, pipe) : inner).trim();
  const param = pipe >= 0 ? inner.slice(pipe + 1).trim() : '';

  // Drop heading/block anchors (#heading, #^block-id)
  target = target.split('#')[0]!.trim();

  let display: string | undefined;
  let width: number | undefined;
  if (param) {
    if (/^\d+$/.test(param)) {
      width = parseInt(param, 10);
    } else {
      display = param;
    }
  }

  return { target, display, width };
}

/**
 * Resolve a wiki-link target to a vault file path.
 * 1. Relative to the note directory
 * 2. Fallback: unique match by filename anywhere in the vault (Obsidian-style)
 */
export function resolveWikiLinkPath(
  notePath: string,
  target: string,
  vaultFiles: string[]
): string | null {
  const normalized = target.replace(/\\/g, '/');
  const relative = resolveNoteAssetPath(notePath, normalized);

  if (vaultFiles.includes(relative)) {
    return relative;
  }
  if (vaultFiles.includes(normalized)) {
    return normalized;
  }

  const basename = normalized.includes('/')
    ? normalized.slice(normalized.lastIndexOf('/') + 1)
    : normalized;

  const matches = vaultFiles.filter((filePath) => {
    if (filePath === normalized || filePath === relative) return true;
    if (filePath.endsWith(`/${normalized}`)) return true;
    const fileBase = filePath.includes('/')
      ? filePath.slice(filePath.lastIndexOf('/') + 1)
      : filePath;
    return fileBase === basename;
  });

  if (matches.length === 0) return null;
  if (matches.length === 1) return matches[0]!;

  const noteDir = notePath.includes('/')
    ? notePath.slice(0, notePath.lastIndexOf('/'))
    : '';

  const inNoteDir = matches.find(
    (m) => noteDir && (m.startsWith(`${noteDir}/`) || m === basename)
  );
  if (inNoteDir) return inNoteDir;

  return matches[0]!;
}

export type FetchBlobFn = (path: string) => Promise<Blob>;

export interface ResolveEmbedsResult {
  markdown: string;
  blobUrls: string[];
}

/**
 * Resolve Obsidian image embeds in markdown for dashboard preview.
 * - Standard: ![alt](path)
 * - Wiki: ![[image.png]], ![[folder/img.png|alt]], ![[img.png|200]]
 */
export async function resolveObsidianImageEmbeds(
  markdown: string,
  notePath: string,
  vaultFiles: string[],
  fetchBlob: FetchBlobFn
): Promise<ResolveEmbedsResult> {
  const blobUrls: string[] = [];
  let result = markdown;

  const wikiPattern = /!\[\[([^\]]+)\]\]/g;
  for (const match of [...markdown.matchAll(wikiPattern)]) {
    const full = match[0];
    const embed = parseWikiEmbed(match[1]!);
    if (!isImageFileName(embed.target)) continue;

    const assetPath = resolveWikiLinkPath(notePath, embed.target, vaultFiles);
    if (!assetPath) continue;

    try {
      const blob = await fetchBlob(assetPath);
      const url = URL.createObjectURL(blob);
      blobUrls.push(url);
      const alt = embed.display ?? embed.target;
      const widthAttr = embed.width ? ` width="${embed.width}"` : '';
      const img = `<img src="${url}" alt="${escapeAttr(alt)}" class="max-w-full h-auto rounded-lg my-4"${widthAttr} />`;
      result = result.replace(full, img);
    } catch {
      // Missing asset on server — leave wiki syntax visible
    }
  }

  const mdPattern = /!\[([^\]]*)\]\(([^)]+)\)/g;
  for (const match of [...markdown.matchAll(mdPattern)]) {
    const full = match[0];
    if (result.indexOf(full) < 0) continue;

    const alt = match[1] ?? '';
    const src = match[2]!.trim();
    if (/^https?:\/\//i.test(src) || src.startsWith('data:') || src.startsWith('blob:')) {
      continue;
    }

    const assetPath = resolveNoteAssetPath(notePath, src);
    if (!isImageFileName(assetPath)) continue;

    try {
      const blob = await fetchBlob(assetPath);
      const url = URL.createObjectURL(blob);
      blobUrls.push(url);
      result = result.replace(full, `![${alt}](${url})`);
    } catch {
      // keep original
    }
  }

  return { markdown: result, blobUrls };
}

function escapeAttr(value: string): string {
  return value
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    .replace(/"/g, '&quot;')
    .replace(/'/g, '&#39;');
}
