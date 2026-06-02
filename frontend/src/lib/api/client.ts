import { authStore } from '$lib/stores/auth.svelte';

export interface SetupCheckResponse {
  ready: boolean;
  missing_fields: string[];
}

export interface SetupGenerateResponse {
  hash: string;
  jwt_secret: string;
}

export interface GenerateHashResponse {
  hash: string;
}

export interface GenerateJWTSecretResponse {
  secret: string;
}

export interface AuthLoginResponse {
  token: string;
  expires_at: number;
  device_id: string;
}

export interface SettingsResponse {
  security: {
    session_ttl_hours: number;
    device_token_ttl_days: number;
    login_rate_limit_per_minute: number;
    bcrypt_cost: number;
    master_hash_configured: boolean;
    jwt_secret_configured: boolean;
  };
  sync: {
    delete_threshold_count: number;
    delete_threshold_percent: number;
  };
  operational: {
    tz: string;
    puid: number;
    pgid: number;
  };
}

export interface VaultResponse {
  id: string;
  name: string;
  path: string;
  last_sync?: number;
  last_sync_device_id?: string;
  last_sync_device_name?: string;
  stats: {
    files: number;
    folders: number;
    size_bytes: number;
  };
}

export interface FileResponse {
  id: string;
  name: string;
  path: string;
  content: string;
  hash: string;
  modified_at: string;
}

interface RawVault {
  vault_id?: string;
  id?: string | number;
  name: string;
  path: string;
  last_sync?: number;
  last_sync_device_id?: string;
  last_sync_device_name?: string;
  total_files?: number;
  total_size_bytes?: number;
  stats?: {
    files?: number;
    folders?: number;
    size_bytes?: number;
  };
}

function mapVault(raw: RawVault): VaultResponse {
  return {
    id: String(raw.vault_id ?? raw.id ?? ''),
    name: raw.name,
    path: raw.path,
    last_sync: raw.last_sync,
    last_sync_device_id: raw.last_sync_device_id,
    last_sync_device_name: raw.last_sync_device_name,
    stats: {
      files: raw.total_files ?? raw.stats?.files ?? 0,
      folders: raw.stats?.folders ?? 0,
      size_bytes: raw.total_size_bytes ?? raw.stats?.size_bytes ?? 0
    }
  };
}

/** API base URL for backend requests and Obsidian plugin Server URL field. */
export function getApiBaseUrl(): string {
  if (import.meta.env.VITE_DEVITRI_BACKEND_URL) {
    return import.meta.env.VITE_DEVITRI_BACKEND_URL;
  }
  if (typeof window !== 'undefined' && window.location.hostname === 'localhost') {
    return 'http://localhost:8080';
  }
  if (typeof window !== 'undefined') {
    return window.location.origin;
  }
  return '';
}

function authHeaders(): Record<string, string> {
  const headers: Record<string, string> = {};
  if (authStore.token) {
    headers.Authorization = `Bearer ${authStore.token}`;
  }
  return headers;
}

export class ApiClient {
  private baseUrl: string;

  constructor() {
    this.baseUrl = getApiBaseUrl();
  }

  private async request(path: string, init: RequestInit = {}): Promise<Response> {
    const headers = new Headers(init.headers);
    for (const [key, value] of Object.entries(authHeaders())) {
      headers.set(key, value);
    }
    return fetch(`${this.baseUrl}${path}`, { ...init, headers });
  }

  async checkSetup(): Promise<SetupCheckResponse> {
    const response = await fetch(`${this.baseUrl}/api/setup/check`);
    if (!response.ok) {
      throw new Error(`Setup check failed (${response.status})`);
    }
    return response.json();
  }

  async generateSetupConfig(password: string): Promise<SetupGenerateResponse> {
    const response = await fetch(`${this.baseUrl}/api/setup/generate`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ password })
    });
    if (!response.ok) {
      throw new Error(
        response.status === 400 ? 'Password must be at least 8 characters' : `Setup generate failed (${response.status})`
      );
    }
    return response.json();
  }

  async generateMasterHash(password: string): Promise<GenerateHashResponse> {
    const response = await fetch(`${this.baseUrl}/api/setup/generate-master-hash`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ password })
    });
    if (!response.ok) {
      throw new Error(`Generate master hash failed (${response.status})`);
    }
    return response.json();
  }

  async generateJWTSecret(): Promise<GenerateJWTSecretResponse> {
    const response = await fetch(`${this.baseUrl}/api/setup/generate-jwt-secret`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' }
    });
    if (!response.ok) {
      throw new Error(`Generate JWT secret failed (${response.status})`);
    }
    return response.json();
  }

  async login(
    password: string,
    deviceId: string,
    deviceName = 'Devitri Web Dashboard'
  ): Promise<AuthLoginResponse> {
    const response = await fetch(`${this.baseUrl}/api/auth/login`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ password, device_id: deviceId, device_name: deviceName })
    });
    if (!response.ok) {
      throw new Error(response.status === 401 ? 'Invalid password' : `Login failed (${response.status})`);
    }
    return response.json();
  }

  async logout(): Promise<void> {
    const response = await this.request('/api/auth/logout', { method: 'POST' });
    if (!response.ok) {
      throw new Error(`Logout failed (${response.status})`);
    }
  }

  async getSession(): Promise<any> {
    const response = await this.request('/api/auth/session');
    if (!response.ok) {
      throw new Error(`Session invalid (${response.status})`);
    }
    return response.json();
  }

  async getCurrentSession(): Promise<any> {
    return this.getSession();
  }

  async generateToken(
    deviceId: string,
    deviceName: string,
    password: string
  ): Promise<AuthLoginResponse> {
    const response = await this.request('/api/auth/token', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        device_id: deviceId,
        device_name: deviceName,
        password,
      }),
    });
    if (!response.ok) {
      throw new Error(`Generate token failed (${response.status})`);
    }
    return response.json();
  }

  async getSessions(): Promise<any[]> {
    const response = await this.request('/api/auth/sessions');
    if (!response.ok) {
      throw new Error(`Get sessions failed (${response.status})`);
    }
    return response.json();
  }

  async revokeSession(id: number, password?: string): Promise<void> {
    const response = await this.request(`/api/auth/sessions/${id}`, {
      method: 'DELETE',
      headers: { 'Content-Type': 'application/json' },
      body: password ? JSON.stringify({ password }) : undefined,
    });
    if (!response.ok) {
      throw new Error(`Revoke session failed (${response.status})`);
    }
  }

  async getSettings(): Promise<SettingsResponse> {
    const response = await this.request('/api/settings');
    if (!response.ok) {
      throw new Error(`Get settings failed (${response.status})`);
    }
    return response.json();
  }

  async getVaults(): Promise<VaultResponse[]> {
    const response = await this.request('/api/vaults');
    if (!response.ok) {
      throw new Error(`Get vaults failed: ${response.statusText}`);
    }
    const data = await response.json();
    const list: RawVault[] = Array.isArray(data) ? data : (data.vaults ?? []);
    return list.map(mapVault);
  }

  async getVault(id: string): Promise<VaultResponse> {
    const response = await this.request(`/api/vaults/${id}`);
    if (!response.ok) {
      throw new Error(`Get vault failed: ${response.statusText}`);
    }
    return mapVault(await response.json());
  }

  async getManifest(vaultId: string): Promise<{ files: Array<{ path: string; hash: string; modified_at: number }> }> {
    const response = await this.request(`/api/vaults/${vaultId}/sync/manifest`);
    if (!response.ok) {
      throw new Error(`Get manifest failed: ${response.statusText}`);
    }
    const data = await response.json();
    const files = Array.isArray(data?.files) ? data.files : [];
    return { ...data, files };
  }

  async getFileContent(vaultId: string, path: string): Promise<string> {
    const response = await this.request(
      `/api/vaults/${vaultId}/sync/download?path=${encodeURIComponent(path)}`
    );
    if (!response.ok) {
      throw new Error(`Get file failed: ${response.statusText}`);
    }
    return response.text();
  }

  /** Download raw file bytes (images, attachments, etc.). */
  async getFileBlob(vaultId: string, path: string): Promise<Blob> {
    const response = await this.request(
      `/api/vaults/${vaultId}/sync/download?path=${encodeURIComponent(path)}`
    );
    if (!response.ok) {
      throw new Error(`Get file failed: ${response.statusText}`);
    }
    return response.blob();
  }
}
