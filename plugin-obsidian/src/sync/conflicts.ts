/**
 * Conflict Resolution
 * Three-way merge for Markdown files, conflict copy creation for conflicts
 */


export interface ThreeWayMergeResult {
  success: boolean;
  content?: string;
  conflict_copy_path?: string;
}

export function threeWayMerge(
  base: string,
  local: string,
  remote: string
): ThreeWayMergeResult {
  // Split content into paragraphs (split on \n\n)
  const baseParagraphs = base.split('\n\n');
  const localParagraphs = local.split('\n\n');
  const remoteParagraphs = remote.split('\n\n');

  // If lengths differ significantly, merge likely to fail
  if (
    Math.abs(baseParagraphs.length - localParagraphs.length) > 2 ||
    Math.abs(baseParagraphs.length - remoteParagraphs.length) > 2
  ) {
    return { success: false };
  }

  // For each paragraph, detect which version changed
  const merged: string[] = [];

  const maxLen = Math.max(
    baseParagraphs.length,
    localParagraphs.length,
    remoteParagraphs.length
  );
  for (let i = 0; i < maxLen; i++) {
    const b = baseParagraphs[i] ?? '';
    const l = localParagraphs[i] ?? '';
    const r = remoteParagraphs[i] ?? '';

    if (l === r) {
      // Both made same change or no change
      merged.push(l || b);
    } else if (l === b) {
      // Only remote changed
      merged.push(r);
    } else if (r === b) {
      // Only local changed
      merged.push(l);
    } else {
      // Both changed differently — conflict on this paragraph
      return { success: false };
    }
  }

  return { success: true, content: merged.filter((p) => p).join('\n\n') };
}

export function createConflictCopy(
  path: string,
  deviceId: string
): string {
  const now = new Date();
  const timestamp = now.toISOString()
    .replace(/[-:]/g, '')
    .slice(0, 12); // YYYYMMDDHHMM
  const parts = path.split('.');
  const ext = parts.pop() ?? '';
  const baseName = parts.join('.');
  const conflictPath = `${baseName} (Devitri Conflict - ${deviceId} - ${timestamp}).${ext}`;

  return conflictPath;
}

export function formatTimestamp(unixTimestamp: number): string {
  if (unixTimestamp === 0) return 'Never';
  const date = new Date(unixTimestamp * 1000);
  return date.toLocaleString();
}
