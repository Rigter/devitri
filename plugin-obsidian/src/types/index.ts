/**
 * Devitri Obsidian Plugin - TypeScript Types
 */

export interface FileState {
  path: string;
  hash: string;
  modified_at: number;
  size: number;
}

export interface SyncManifest {
  vault_id: string;
  generated_at: number;
  files: FileState[];
}

/** Ensures manifest.files is always an array (API may return null for empty vaults). */
export function normalizeSyncManifest(
  manifest: SyncManifest | null | undefined,
  vaultId = ''
): SyncManifest {
  if (!manifest) {
    return { vault_id: vaultId, generated_at: 0, files: [] };
  }
  return {
    vault_id: manifest.vault_id || vaultId,
    generated_at: manifest.generated_at ?? 0,
    files: Array.isArray(manifest.files) ? manifest.files : [],
  };
}

export interface SyncDecision {
  path: string;
  action: 'upload' | 'download' | 'conflict' | 'skip' | 'delete';
  local?: FileState;
  base?: FileState;
  remote?: FileState;
  conflict_copy_path?: string;
}

export interface SyncResult {
  uploaded: number;
  downloaded: number;
  conflicts: number;
  errors: string[];
}

export interface PluginData {
  serverUrl: string;
  vaultId: string;
  token: string;
  tokenExpiresAt: number;
  deviceId: string;
  deviceName: string;
  syncInterval: number;
  lastSync: number;
  manifestB: SyncManifest;
  conflicts: ConflictRecord[];
}

export interface ConflictRecord {
  path: string;
  detected_at: number;
  device_id: string;
  conflict_copy_path?: string;
}

export interface SyncBatchRequest {
  device_id: string;
  files: Array<{ path: string; hash: string; modified_at: number; size: number }>;
  bulk_delete_confirmed?: boolean;
}

export interface SyncBatchResponse {
  to_upload: string[];
  to_download: string[];
  conflicts: string[];
  to_delete: string[];
  bulk_delete_warning?: boolean;
}

export interface AuthResponse {
  token: string;
  expires_at: number;
  device_id: string;
}

export interface SessionResponse {
  device_id: string;
  device_name: string;
  expires_at: number;
}

export interface SettingsResponse {
  security: {
    session_ttl_hours: number;
    device_token_ttl_days: number;
    login_rate_limit_per_minute: number;
  };
  sync: {
    delete_threshold_count: number;
    delete_threshold_percent: number;
  };
}


/**
 * Minimal surface the Settings UI expects from the plugin instance.
 * Keep this interface stable even if the plugin class name changes.
 */
export type DevitriPluginHost = import('obsidian').Plugin & {
  data: PluginData;
  isSyncing: boolean;
  saveData(data?: unknown): Promise<void>;
  restartSyncInterval(): void;
  isConnected(): boolean;
  getApi(): any;
  replaceApi(api: unknown): void;
  startSyncCycle(options?: { notify?: boolean }): Promise<import('./index').SyncResult | null>;
  updateStatusBarIdle(): void;
};
