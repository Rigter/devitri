/**
 * Manifest Builder
 * Constructs local manifest by scanning vault and computing SHA-256 hashes
 */

import { Vault } from 'obsidian';
import { FileState, SyncManifest } from '../types';

export async function computeHash(buffer: ArrayBuffer): Promise<string> {
  const hashBuffer = await crypto.subtle.digest('SHA-256', buffer);
  const hashArray = Array.from(new Uint8Array(hashBuffer));
  return hashArray.map((b) => b.toString(16).padStart(2, '0')).join('');
}

export async function buildLocalManifest(vault: Vault): Promise<SyncManifest> {
  const files: FileState[] = [];
  const markdownFiles = await vault.getMarkdownFiles();

  for (const file of markdownFiles) {
    // Skip .obsidian/ directory
    if (file.path.startsWith('.obsidian/')) {
      continue;
    }

    const stat = file.stat;
    if (stat.size > 0) {
      const content = await vault.readBinary(file);
      const hash = await computeHash(content);
      files.push({
        path: file.path,
        hash,
        modified_at: stat.mtime,
        size: stat.size,
      });
    }
  }

  return {
    vault_id: '', // Will be set by caller
    generated_at: Math.floor(Date.now() / 1000),
    files,
  };
}
