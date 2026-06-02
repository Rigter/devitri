import type { VaultResponse } from '$lib/api/client';

export function formatVaultLastSync(vault: VaultResponse): string | null {
  if (!vault.last_sync || vault.last_sync <= 0) {
    return null;
  }

  const date = new Date(vault.last_sync * 1000);
  const formatted = date.toLocaleString(undefined, {
    dateStyle: 'medium',
    timeStyle: 'short',
  });

  const device =
    vault.last_sync_device_name?.trim() ||
    vault.last_sync_device_id?.trim() ||
    'Unknown device';

  return `${formatted} · ${device}`;
}
