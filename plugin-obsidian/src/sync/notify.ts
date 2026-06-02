import type { SyncResult } from '../types';

/** Short relative time for status bar (English, compact). */
export function formatRelativeSync(unixSec: number): string {
  if (unixSec <= 0) return 'never';

  const elapsed = Math.floor(Date.now() / 1000) - unixSec;
  if (elapsed < 45) return 'just now';
  if (elapsed < 3600) return `${Math.floor(elapsed / 60)}m ago`;
  if (elapsed < 86400) return `${Math.floor(elapsed / 3600)}h ago`;
  return `${Math.floor(elapsed / 86400)}d ago`;
}

export function formatSyncSummary(result: SyncResult): string {
  const parts: string[] = [];
  if (result.uploaded > 0) parts.push(`${result.uploaded}↑`);
  if (result.downloaded > 0) parts.push(`${result.downloaded}↓`);
  if (result.conflicts > 0) parts.push(`${result.conflicts} conflict`);
  return parts.length > 0 ? ` · ${parts.join(', ')}` : '';
}
