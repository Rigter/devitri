/**
 * Devitri API Wrapper
 * All HTTP calls go through this layer to ensure consistent error handling
 * and mobile compatibility (requestUrl only).
 */

import { requestUrl, RequestUrlResponse } from 'obsidian';
import {
  AuthResponse,
  normalizeSyncManifest,
  SessionResponse,
  SettingsResponse,
  SyncBatchRequest,
  SyncBatchResponse,
  SyncManifest,
} from '../types';

/** Returned when the server rejects a delete that exceeds bulk-delete thresholds. */
export class BulkDeleteBlockedError extends Error {
  constructor(message = 'Bulk delete blocked by server safety limits') {
    super(message);
    this.name = 'BulkDeleteBlockedError';
  }
}

export class DevitriApi {
  private serverUrl: string;
  private _vaultId: string;
  private token: string;
  private _deviceId: string;

  constructor(
    serverUrl: string,
    vaultId: string,
    token: string,
    deviceId: string
  ) {
    this.serverUrl = serverUrl.replace(/\/$/, '');
    this._vaultId = vaultId;
    this.token = token;
    this._deviceId = deviceId;
  }

  get vaultId(): string {
    return this._vaultId;
  }

  get deviceId(): string {
    return this._deviceId;
  }

  updateCredentials(
    serverUrl: string,
    vaultId: string,
    token: string,
    deviceId: string
  ): void {
    this.serverUrl = serverUrl.replace(/\/$/, '');
    this._vaultId = vaultId;
    this.token = token;
    this._deviceId = deviceId;
  }

  private normalizeResponse(response: RequestUrlResponse) {
    if (response.status >= 200 && response.status < 300) {
      if (response.json !== undefined) {
        return response.json;
      }
      if (response.arrayBuffer !== undefined) {
        return response.arrayBuffer;
      }
      return response.text;
    }
    throw new Error(`HTTP ${response.status}`);
  }

  private async request(
    method: string,
    path: string,
    options: {
      headers?: Record<string, string>;
      body?: string | ArrayBuffer;
      auth?: boolean;
    } = {}
  ) {
    const url = `${this.serverUrl}${path}`;
    const headers: Record<string, string> = { ...options.headers };

    if (options.auth !== false && this.token) {
      headers['Authorization'] = `Bearer ${this.token}`;
    }

    try {
      const response = await requestUrl({
        url,
        method,
        headers,
        body: options.body,
        throw: false,
      });

      if (response.status === 401) {
        throw new Error(
          'Session expired. Generate a new access key in the dashboard and reconnect.'
        );
      }

      if (response.status === 404) {
        const hint =
          this.serverUrl.includes('localhost:3000')
            ? ' (Hint: in local dev, use backend http://localhost:8080, not the dashboard :3000)'
            : '';
        throw new Error(`HTTP 404 ${method} ${path}${hint}`);
      }

      return this.normalizeResponse(response);
    } catch (err) {
      const message = err instanceof Error ? err.message : String(err);
      throw new Error(`Devitri API error: ${message}`);
    }
  }

  /** Validate the current token against the server. */
  public async verifySession(): Promise<SessionResponse> {
    if (!this.token) {
      throw new Error('No token configured');
    }
    const session = (await this.request('GET', '/api/auth/session')) as SessionResponse;
    this._deviceId = session.device_id;
    return session;
  }

  /** Validate an access key from /connect and store session metadata. */
  public async connectWithAccessKey(accessKey: string): Promise<SessionResponse> {
    this.token = accessKey.trim();
    const session = (await this.request('GET', '/api/auth/session')) as SessionResponse;
    this._deviceId = session.device_id;
    return session;
  }

  /** Optional password login (short-lived session). */
  public async authenticateWithPassword(
    password: string,
    deviceId: string,
    deviceName: string
  ): Promise<AuthResponse> {
    const response = await requestUrl({
      url: `${this.serverUrl}/api/auth/login`,
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        password,
        device_id: deviceId,
        device_name: deviceName,
      }),
      throw: false,
    });

    if (response.status !== 200) {
      throw new Error('Authentication failed');
    }

    const data = response.json as AuthResponse;
    this.token = data.token;
    this._deviceId = data.device_id;
    return data;
  }

  public async getSettings(): Promise<SettingsResponse> {
    return this.request('GET', '/api/settings') as Promise<SettingsResponse>;
  }


  public async getManifest(): Promise<SyncManifest> {
    const raw = (await this.request(
      'GET',
      `/api/vaults/${this._vaultId}/sync/manifest`
    )) as SyncManifest;
    return normalizeSyncManifest(raw, this._vaultId);
  }

  public async uploadFile(
    path: string,
    hash: string,
    content: ArrayBuffer
  ): Promise<void> {
    await this.request('POST', `/api/vaults/${this._vaultId}/sync/upload`, {
      headers: {
        'X-File-Path': path,
        'X-File-Hash': hash,
        'Content-Type': 'application/octet-stream',
      },
      body: content,
    });
  }

  public async downloadFile(path: string): Promise<{ arrayBuffer: ArrayBuffer }> {
    const url = `${this.serverUrl}/api/vaults/${this._vaultId}/sync/download?path=${encodeURIComponent(path)}`;
    const headers: Record<string, string> = {
      Accept: 'application/octet-stream',
    };
    if (this.token) {
      headers.Authorization = `Bearer ${this.token}`;
    }

    try {
      const response = await requestUrl({
        url,
        method: 'GET',
        headers,
        throw: false,
      });

      if (response.status === 401) {
        throw new Error(
          'Session expired. Generate a new access key in the dashboard and reconnect.'
        );
      }

      if (response.status === 404) {
        throw new Error(`File not found: ${path}`);
      }

      if (response.status < 200 || response.status >= 300) {
        throw new Error(`HTTP ${response.status}`);
      }

      // Do not use normalizeResponse — it prefers .json and breaks on binary (PNG, etc.)
      if (response.arrayBuffer) {
        return { arrayBuffer: response.arrayBuffer };
      }

      throw new Error(`No binary body in download response for ${path}`);
    } catch (err) {
      const message = err instanceof Error ? err.message : String(err);
      throw new Error(`Devitri API error: ${message}`);
    }
  }

  public async deleteFile(path: string, bulkDeleteConfirmed = false): Promise<void> {
    const url = `${this.serverUrl}/api/vaults/${this._vaultId}/sync/file`;
    const headers: Record<string, string> = { 'Content-Type': 'application/json' };
    if (this.token) {
      headers.Authorization = `Bearer ${this.token}`;
    }
    if (bulkDeleteConfirmed) {
      headers['X-Bulk-Delete-Confirmed'] = 'true';
    }

    const response = await requestUrl({
      url,
      method: 'DELETE',
      headers,
      body: JSON.stringify({ path }),
      throw: false,
    });

    if (response.status === 401) {
      throw new Error(
        'Session expired. Generate a new access key in the dashboard and reconnect.'
      );
    }

    if (response.status === 403) {
      const body = response.json as { error?: string; message?: string } | undefined;
      if (body?.error === 'bulk_delete_blocked') {
        throw new BulkDeleteBlockedError(body.message);
      }
    }

    if (response.status < 200 || response.status >= 300) {
      throw new Error(`HTTP ${response.status}`);
    }
  }

  public async batchSync(files: SyncBatchRequest): Promise<SyncBatchResponse> {
    return this.request('POST', `/api/vaults/${this._vaultId}/sync/batch`, {
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(files),
    }) as Promise<SyncBatchResponse>;
  }

  public updateToken(token: string): void {
    this.token = token;
  }

  public getToken(): string {
    return this.token;
  }
}
