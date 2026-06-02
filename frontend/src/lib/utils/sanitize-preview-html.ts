/**
 * Allowlist HTML sanitizer for markdown preview ({@html}).
 * Runs in the browser only (kit ssr: false). Strips tags/attrs outside our renderer output.
 */

const ALLOWED_TAGS = new Set([
  'article',
  'section',
  'h1',
  'h2',
  'h3',
  'h4',
  'h5',
  'h6',
  'p',
  'div',
  'span',
  'ul',
  'ol',
  'li',
  'blockquote',
  'pre',
  'code',
  'table',
  'thead',
  'tbody',
  'tr',
  'th',
  'td',
  'hr',
  'a',
  'img',
  'strong',
  'em',
  'del',
  'mark',
  'sup',
  'br',
]);

const ALLOWED_ATTRS = new Set([
  'class',
  'id',
  'href',
  'alt',
  'src',
  'loading',
  'target',
  'rel',
  'title',
  'data-lang',
  'data-target',
]);

const SAFE_HREF = /^(?:https?:|mailto:|#)/i;
const SAFE_IMG_SRC = /^(?:https:|blob:)/i;

function isSafeHref(value: string): boolean {
  const v = value.trim();
  if (!v || v.startsWith('javascript:') || v.startsWith('data:')) return false;
  return SAFE_HREF.test(v);
}

function isSafeImgSrc(value: string): boolean {
  const v = value.trim();
  if (!v || v.startsWith('javascript:') || v.startsWith('data:')) return false;
  return SAFE_IMG_SRC.test(v);
}

function sanitizeElement(el: Element): void {
  const tag = el.tagName.toLowerCase();
  if (!ALLOWED_TAGS.has(tag)) {
    el.remove();
    return;
  }

  for (const attr of [...el.attributes]) {
    const name = attr.name.toLowerCase();
    if (name.startsWith('on') || name === 'style' || !ALLOWED_ATTRS.has(name)) {
      el.removeAttribute(attr.name);
      continue;
    }

    if (name === 'href' && !isSafeHref(attr.value)) {
      el.removeAttribute(attr.name);
    } else if (name === 'src' && !isSafeImgSrc(attr.value)) {
      el.removeAttribute(attr.name);
    } else if (name === 'target' && attr.value !== '_blank') {
      el.removeAttribute(attr.name);
    } else if (name === 'rel' && attr.value !== 'noopener noreferrer') {
      el.setAttribute('rel', 'noopener noreferrer');
    }
  }

  for (const child of [...el.childNodes]) {
    if (child.nodeType === Node.ELEMENT_NODE) {
      sanitizeElement(child as Element);
    } else if (child.nodeType === Node.COMMENT_NODE) {
      child.remove();
    }
  }
}

export function sanitizePreviewHtml(html: string): string {
  if (!html) return '';
  if (typeof DOMParser === 'undefined') {
    return '';
  }

  const doc = new DOMParser().parseFromString(html, 'text/html');
  for (const child of [...doc.body.childNodes]) {
    if (child.nodeType === Node.ELEMENT_NODE) {
      sanitizeElement(child as Element);
    }
  }
  return doc.body.innerHTML;
}
