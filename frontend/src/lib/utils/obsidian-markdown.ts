/**
 * Obsidian-flavored markdown → HTML for dashboard preview (no external markdown deps).
 */

import { sanitizePreviewHtml } from './sanitize-preview-html';

const CALLOUT_TYPES = new Set([
  'note',
  'abstract',
  'summary',
  'tldr',
  'info',
  'todo',
  'tip',
  'hint',
  'important',
  'success',
  'check',
  'done',
  'question',
  'help',
  'faq',
  'warning',
  'caution',
  'attention',
  'failure',
  'fail',
  'missing',
  'danger',
  'error',
  'bug',
  'example',
  'quote',
  'cite',
]);

export function renderObsidianMarkdown(source: string): string {
  if (!source.trim()) return '';

  let md = stripFrontmatter(source);
  md = stripObsidianComments(md);

  const footnotes = collectFootnoteDefinitions(md);

  const { text: withoutFences, tokens: fenceTokens } = extractFencedBlocks(md);
  const { text: withoutInline, tokens: inlineTokens } = extractInlineCode(withoutFences);
  const { text: withoutImgs, tokens: imgTokens } = extractPreRenderedImgs(withoutInline);

  md = withoutImgs.replace(/^\[\^[\w-]+\]:\s*.+$/gm, '');
  md = preprocessCallouts(md);
  const body = parseBlocks(md.split('\n'));

  let html = body;
  if (footnotes.length > 0) {
    html += `<section class="devitri-footnotes"><hr /><ol>${footnotes
      .map(
        (fn) =>
          `<li id="fn-${escapeHtml(fn.id)}">${formatInline(fn.text)} <a href="#fnref-${escapeHtml(fn.id)}" class="devitri-footnote-back">↩</a></li>`
      )
      .join('')}</ol></section>`;
  }

  html = restoreTokens(html, inlineTokens, 'INLINE');
  html = restoreTokens(html, fenceTokens, 'FENCE');
  html = restoreTokens(html, imgTokens, 'IMG');

  return sanitizePreviewHtml(html);
}

function stripFrontmatter(md: string): string {
  if (!md.startsWith('---')) return md;
  const end = md.indexOf('\n---', 3);
  if (end === -1) return md;
  return md.slice(end + 4).replace(/^\s+/, '');
}

function stripObsidianComments(md: string): string {
  return md.replace(/%%[\s\S]*?%%/g, '');
}

interface TokenStore {
  text: string;
  tokens: string[];
}

function extractFencedBlocks(md: string): TokenStore {
  const tokens: string[] = [];
  const text = md.replace(/```([\w-]*)\n?([\s\S]*?)```/g, (full, lang, code) => {
    const language = (lang || 'text').toLowerCase();
    const label =
      language === 'mermaid'
        ? 'Mermaid diagram (not rendered in preview)'
        : language;
    const html = `<pre class="devitri-code-block" data-lang="${escapeHtml(language)}"><div class="devitri-code-block__label">${escapeHtml(label)}</div><code>${escapeHtml(code.replace(/\n$/, ''))}</code></pre>`;
    tokens.push(html);
    return `%%FENCE_${tokens.length - 1}%%`;
  });
  return { text, tokens };
}

function extractInlineCode(md: string): TokenStore {
  const tokens: string[] = [];
  const text = md.replace(/`([^`\n]+)`/g, (full, code) => {
    tokens.push(`<code class="devitri-inline-code">${escapeHtml(code)}</code>`);
    return `%%INLINE_${tokens.length - 1}%%`;
  });
  return { text, tokens };
}

function sanitizeMarkdownImgTag(tag: string): string | null {
  const srcMatch = tag.match(/\ssrc=["']([^"']+)["']/i);
  const src = srcMatch?.[1]?.trim() ?? '';
  if (!src.startsWith('blob:') && !/^https:\/\//i.test(src)) {
    return null;
  }
  const altMatch = tag.match(/\salt=["']([^"']*)["']/i);
  const alt = escapeHtml(altMatch?.[1] ?? '');
  return `<img src="${escapeHtml(src)}" alt="${alt}" class="devitri-md-img" loading="lazy" />`;
}

function extractPreRenderedImgs(md: string): TokenStore {
  const tokens: string[] = [];
  let text = md.replace(/<img\s[^>]*\/?>/gi, (tag) => {
    const safe = sanitizeMarkdownImgTag(tag);
    if (!safe) {
      return '';
    }
    tokens.push(safe);
    return `%%IMG_${tokens.length - 1}%%`;
  });
  text = text.replace(/%%DEVITRI_IMG_(\d+)%%/g, (_, i) => {
    return `%%IMG_${i}%%`;
  });
  return { text, tokens };
}

function restoreTokens(html: string, tokens: string[] | undefined, prefix: string): string {
  if (!tokens?.length) return html;
  let out = html;
  for (let i = 0; i < tokens.length; i++) {
    out = out.split(`%%${prefix}_${i}%%`).join(tokens[i] ?? '');
  }
  return out;
}

function preprocessCallouts(md: string): string {
  const lines = md.split('\n');
  const out: string[] = [];
  let i = 0;

  while (i < lines.length) {
    const match = lines[i].match(/^>\s*\[!([\w-]+)\]\s*(.*)$/i);
    if (match) {
      const type = match[1].toLowerCase();
      const safeType = CALLOUT_TYPES.has(type) ? type : 'note';
      const title = match[2].trim();
      const body: string[] = [];
      i++;
      while (i < lines.length && /^>\s?/.test(lines[i])) {
        body.push(lines[i].replace(/^>\s?/, ''));
        i++;
      }
      const bodyHtml = body.length
        ? body.map((line) => `<p>${formatInline(line)}</p>`).join('')
        : '';
      const titleHtml = title
        ? `<div class="devitri-callout__title">${formatInline(title)}</div>`
        : `<div class="devitri-callout__title">${escapeHtml(safeType)}</div>`;
      out.push(
        `<div class="devitri-callout devitri-callout--${safeType}">${titleHtml}<div class="devitri-callout__body">${bodyHtml}</div></div>`
      );
      continue;
    }
    out.push(lines[i]);
    i++;
  }
  return out.join('\n');
}

function parseBlocks(lines: string[]): string {
  const html: string[] = [];
  let i = 0;

  while (i < lines.length) {
    const line = lines[i];

    if (!line.trim()) {
      i++;
      continue;
    }

    const fenceMatch = line.match(/^%%FENCE_(\d+)%%$/);
    if (fenceMatch) {
      html.push(line);
      i++;
      continue;
    }

    const heading = line.match(/^(#{1,6})\s+(.+)$/);
    if (heading) {
      const level = heading[1].length;
      html.push(`<h${level}>${formatInline(heading[2])}</h${level}>`);
      i++;
      continue;
    }

    if (/^(-{3,}|\*{3,}|_{3,})$/.test(line.trim())) {
      html.push('<hr class="devitri-hr" />');
      i++;
      continue;
    }

    if (isTableStart(lines, i)) {
      const { block, next } = parseTable(lines, i);
      html.push(block);
      i = next;
      continue;
    }

    if (/^[-*+]\s+\[[ xX]\]\s/.test(line)) {
      const { block, next } = parseTaskList(lines, i);
      html.push(block);
      i = next;
      continue;
    }

    if (/^[-*+]\s+/.test(line) || /^\d+\.\s+/.test(line)) {
      const { block, next } = parseList(lines, i);
      html.push(block);
      i = next;
      continue;
    }

    if (line.startsWith('> ')) {
      const { block, next } = parseBlockquote(lines, i);
      html.push(block);
      i = next;
      continue;
    }

    if (line.startsWith('$$')) {
      const { block, next } = parseMathBlock(lines, i);
      html.push(block);
      i = next;
      continue;
    }

    const para = parseParagraph(lines, i);
    html.push(para.html);
    i = para.next;
  }

  return html.join('\n');
}

function parseParagraph(lines: string[], start: number): { html: string; next: number } {
  const parts: string[] = [];
  let i = start;
  while (i < lines.length && lines[i].trim() && !isBlockStarter(lines, i)) {
    parts.push(formatInline(lines[i]));
    i++;
  }
  if (parts.length === 1) return { html: `<p>${parts[0]}</p>`, next: i };
  return {
    html: `<p>${parts.join('<br />')}</p>`,
    next: i,
  };
}

function isBlockStarter(lines: string[], i: number): boolean {
  const line = lines[i];
  if (!line.trim()) return true;
  if (/^%%(CALLOUT|FENCE)_/.test(line)) return true;
  if (/^#{1,6}\s/.test(line)) return true;
  if (/^[-*+]\s/.test(line) || /^\d+\.\s/.test(line)) return true;
  if (line.startsWith('> ')) return true;
  if (isTableStart(lines, i)) return true;
  if (line.startsWith('$$')) return true;
  return false;
}

function isTableStart(lines: string[], i: number): boolean {
  if (i + 1 >= lines.length) return false;
  return lines[i].includes('|') && /^\|?[\s|:-]+\|?$/.test(lines[i + 1].trim());
}

function parseTable(lines: string[], start: number): { block: string; next: number } {
  const rows: string[][] = [];
  let i = start;
  while (i < lines.length && lines[i].includes('|')) {
    const cells = lines[i]
      .trim()
      .replace(/^\|/, '')
      .replace(/\|$/, '')
      .split('|')
      .map((c) => c.trim());
    rows.push(cells);
    i++;
  }
  if (rows.length < 2) {
    return { block: `<p>${formatInline(lines[start])}</p>`, next: start + 1 };
  }
  const header = rows[0];
  const body = rows.slice(2);
  const thead = `<thead><tr>${header.map((c) => `<th>${formatInline(c)}</th>`).join('')}</tr></thead>`;
  const tbody = `<tbody>${body
    .map((row) => `<tr>${row.map((c) => `<td>${formatInline(c)}</td>`).join('')}</tr>`)
    .join('')}</tbody>`;
  return {
    block: `<div class="devitri-table-wrap"><table class="devitri-table">${thead}${tbody}</table></div>`,
    next: i,
  };
}

function parseTaskList(lines: string[], start: number): { block: string; next: number } {
  const items: string[] = [];
  let i = start;
  while (i < lines.length && /^[-*+]\s+\[[ xX]\]\s/.test(lines[i])) {
    const m = lines[i].match(/^[-*+]\s+\[([ xX])\]\s+(.*)$/);
    if (m) {
      const checked = m[1].toLowerCase() === 'x';
      items.push(
        `<li class="devitri-task${checked ? ' devitri-task--done' : ''}"><span class="devitri-task__box">${checked ? '✓' : ''}</span> ${formatInline(m[2])}</li>`
      );
    }
    i++;
  }
  return { block: `<ul class="devitri-task-list">${items.join('')}</ul>`, next: i };
}

function parseList(lines: string[], start: number): { block: string; next: number } {
  const ordered = /^\d+\.\s+/.test(lines[start]);
  const tag = ordered ? 'ol' : 'ul';
  const items: string[] = [];
  let i = start;
  const pattern = ordered ? /^\d+\.\s+(.*)$/ : /^[-*+]\s+(.*)$/;
  while (i < lines.length && pattern.test(lines[i])) {
    const m = lines[i].match(pattern);
    if (m) items.push(`<li>${formatInline(m[1])}</li>`);
    i++;
  }
  return { block: `<${tag} class="devitri-list">${items.join('')}</${tag}>`, next: i };
}

function parseBlockquote(lines: string[], start: number): { block: string; next: number } {
  const inner: string[] = [];
  let i = start;
  while (i < lines.length && lines[i].startsWith('> ')) {
    inner.push(lines[i].slice(2));
    i++;
  }
  const body = inner.map((l) => `<p>${formatInline(l)}</p>`).join('');
  return { block: `<blockquote class="devitri-blockquote">${body}</blockquote>`, next: i };
}

function parseMathBlock(lines: string[], start: number): { block: string; next: number } {
  if (lines[start].trim() === '$$' && start + 1 < lines.length) {
    const parts: string[] = [];
    let i = start + 1;
    while (i < lines.length && lines[i].trim() !== '$$') {
      parts.push(escapeHtml(lines[i]));
      i++;
    }
    return {
      block: `<div class="devitri-math devitri-math--block"><code>${parts.join('\n')}</code></div>`,
      next: i < lines.length ? i + 1 : i,
    };
  }
  const m = lines[start].match(/^\$\$(.+)\$\$$/);
  if (m) {
    return {
      block: `<div class="devitri-math devitri-math--block"><code>${escapeHtml(m[1])}</code></div>`,
      next: start + 1,
    };
  }
  return { block: `<p>${formatInline(lines[start])}</p>`, next: start + 1 };
}

interface FootnoteDef {
  id: string;
  text: string;
}

function collectFootnoteDefinitions(md: string): FootnoteDef[] {
  const defs: FootnoteDef[] = [];
  const re = /^\[\^([\w-]+)\]:\s*(.+)$/gm;
  let m: RegExpExecArray | null;
  while ((m = re.exec(md)) !== null) {
    defs.push({ id: m[1], text: m[2] });
  }
  return defs;
}

function formatInline(text: string): string {
  let s = escapeHtml(text);

  s = s.replace(/!\[([^\]]*)\]\(([^)]+)\)/g, (_, alt, src) => {
    const safeSrc =
      src.startsWith('blob:') || /^https:\/\//i.test(src) ? src : '';
    if (!safeSrc) return '';
    return `<img src="${escapeHtml(safeSrc)}" alt="${escapeHtml(alt)}" class="devitri-md-img" loading="lazy" />`;
  });

  s = s.replace(/(?<!!)\[\[([^\]]+)\]\]/g, (_, inner: string) => {
    const pipe = inner.indexOf('|');
    let target = pipe >= 0 ? inner.slice(0, pipe) : inner;
    const alias = pipe >= 0 ? inner.slice(pipe + 1) : undefined;
    const hashIdx = target.indexOf('#');
    let heading = '';
    if (hashIdx >= 0) {
      heading = target.slice(hashIdx + 1);
      target = target.slice(0, hashIdx);
    }
    const label = alias?.trim() || target + (heading ? ` › ${heading}` : '');
    return `<span class="devitri-wiki-link" data-target="${escapeHtml(target.trim())}" title="${escapeHtml(target.trim())}">${escapeHtml(label)}</span>`;
  });

  s = s.replace(/(^|[\s([{>])#([a-zA-Z][\w/-]*)/g, (_, before, tag) => {
    return `${before}<span class="devitri-tag">#${escapeHtml(tag)}</span>`;
  });

  s = s.replace(/\[\^([\w-]+)\]/g, (_, id) => {
    return `<sup class="devitri-footnote-ref"><a id="fnref-${escapeHtml(id)}" href="#fn-${escapeHtml(id)}">[${escapeHtml(id)}]</a></sup>`;
  });

  s = s.replace(/\^([a-zA-Z0-9-]+)/g, '<span class="devitri-block-ref" title="Block reference">^$1</span>');

  s = s.replace(/==([^=\n]+)==/g, '<mark class="devitri-highlight">$1</mark>');
  s = s.replace(/~~([^~]+)~~/g, '<del class="devitri-strike">$1</del>');
  s = s.replace(/\*\*([^*]+)\*\*/g, '<strong>$1</strong>');
  s = s.replace(/__([^_]+)__/g, '<strong>$1</strong>');
  s = s.replace(/(?<!\*)\*([^*\n]+)\*(?!\*)/g, '<em>$1</em>');
  s = s.replace(/(?<!_)_([^_\n]+)_(?!_)/g, '<em>$1</em>');

  s = s.replace(/\[([^\]]+)\]\(([^)]+)\)/g, (_, label, href) => {
    const safe =
      href.startsWith('http://') || href.startsWith('https://') || href.startsWith('mailto:')
        ? href
        : '#';
    const external = safe.startsWith('http');
    return `<a href="${escapeHtml(safe)}"${external ? ' target="_blank" rel="noopener noreferrer"' : ''} class="devitri-link">${label}</a>`;
  });

  s = s.replace(/(?<!\$)\$([^$\n]+)\$(?!\$)/g, '<span class="devitri-math devitri-math--inline"><code>$1</code></span>');

  return s;
}

function escapeHtml(value: string): string {
  return value
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    .replace(/"/g, '&quot;');
}
