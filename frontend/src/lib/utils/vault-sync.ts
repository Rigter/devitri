import type { VaultResponse } from '$lib/api/client';

/** Stable fingerprint for detecting remote sync activity (last_sync metadata only). */
export function vaultSyncFingerprint(vault: VaultResponse): string {
  const sync = vault.last_sync ?? 0;
  const deviceId = vault.last_sync_device_id?.trim() ?? '';
  const deviceName = vault.last_sync_device_name?.trim() ?? '';
  return `${vault.id}:${sync}:${deviceId}:${deviceName}`;
}

export function vaultListSyncFingerprint(vaults: VaultResponse[]): string {
  return [...vaults]
    .map(vaultSyncFingerprint)
    .sort()
    .join('|');
}

export function hasVaultListSyncChanged(
  previous: string,
  vaults: VaultResponse[]
): boolean {
  return vaultListSyncFingerprint(vaults) !== previous;
}

export function hasVaultSyncChanged(
  previousSync: number | undefined,
  vault: VaultResponse
): boolean {
  return (vault.last_sync ?? 0) !== (previousSync ?? 0);
}

export interface ManifestFileEntry {
  path: string;
  hash: string;
}

/** Fingerprint of vault file hashes — detects content changes even if last_sync is unchanged. */
export function manifestFingerprint(
  files: ManifestFileEntry[] | undefined
): string {
  if (!files?.length) return '';
  return [...files]
    .map((f) => `${f.path}:${f.hash}`)
    .sort()
    .join('|');
}
