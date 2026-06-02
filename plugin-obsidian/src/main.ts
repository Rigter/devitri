/**
 * Devitri Obsidian Plugin - Main Entry Point
 */

import { Notice, Platform, Plugin } from 'obsidian';
import styles from './styles.css';
import { DevitriSettingsTab } from './ui/SettingsTab';
import { DevitriSyncEngine } from './sync/engine';
import { DevitriApi } from './sync/api';
import { formatRelativeSync, formatSyncSummary } from './sync/notify';
import { normalizeSyncManifest, SyncManifest, PluginData, SyncResult } from './types';

function slugifyVaultId(name: string): string {
  return name
    .trim()
    .toLowerCase()
    .replace(/[^a-z0-9]+/g, '-')
    .replace(/^-+|-+$/g, '');
}

class DevitriPlugin extends Plugin {
  data!: PluginData;
  private api!: DevitriApi;
  private engine!: DevitriSyncEngine;
  private settingsTab!: DevitriSettingsTab;
  private syncIntervalId: number | null = null;
  private statusBarItem: HTMLElement | null = null;
  private statusBarResetTimer: number | null = null;
  isSyncing = false;

  async onload() {
    this.injectStyles();

    this.data = (await this.loadData()) ?? this.getInitialData();
    if (this.data.tokenExpiresAt === undefined) {
      this.data.tokenExpiresAt = 0;
    }

    // Vault ID is derived from the currently open Obsidian vault.
    // Obsidian plugins run per-vault; listing all local vaults is not supported.
    if (!this.data.vaultId) {
      this.data.vaultId = slugifyVaultId(this.app.vault.getName());
      await this.saveData(this.data);
    }

    this.api = new DevitriApi(
      this.data.serverUrl,
      this.data.vaultId,
      this.data.token,
      this.data.deviceId
    );

    this.engine = new DevitriSyncEngine(
      this.app,
      this.api,
      normalizeSyncManifest(
        this.data.manifestB || this.getInitialManifest(),
        this.data.vaultId
      )
    );

    this.settingsTab = new DevitriSettingsTab(this.app, this, this.api, this.engine);
    this.addSettingTab(this.settingsTab);

    this.app.workspace.onLayoutReady(() => {
      this.setupStatusBar();
      this.registerEvent(this.app.vault.on('create', this.onFileCreate.bind(this)));
      this.registerEvent(this.app.vault.on('modify', this.onFileModify.bind(this)));
      this.registerEvent(this.app.vault.on('delete', this.onFileDelete.bind(this)));
      this.registerEvent(this.app.vault.on('rename', this.onFileRename.bind(this)));
      this.startSyncInterval();
    });
  }

  async onExternalSettingsChange() {
    const loaded = await this.loadData();
    if (loaded) {
      this.data = loaded;
    }
    this.api.updateCredentials(
      this.data.serverUrl,
      this.data.vaultId,
      this.data.token,
      this.data.deviceId
    );
    this.engine.setManifestB(
      normalizeSyncManifest(
        this.data.manifestB || this.getInitialManifest(),
        this.data.vaultId
      )
    );
  }

  async onunload() {
    if (this.syncIntervalId !== null) {
      clearInterval(this.syncIntervalId);
      this.syncIntervalId = null;
    }
    if (this.statusBarResetTimer !== null) {
      clearTimeout(this.statusBarResetTimer);
      this.statusBarResetTimer = null;
    }
    this.statusBarItem = null;

    this.data.manifestB = this.engine.getManifestB();
    await this.saveData(this.data);
  }

  restartSyncInterval(): void {
    if (this.syncIntervalId !== null) {
      clearInterval(this.syncIntervalId);
      this.syncIntervalId = null;
    }
    this.startSyncInterval();
  }

  getApi(): DevitriApi {
    return this.api;
  }

  replaceApi(api: DevitriApi): void {
    this.api = api;
    this.engine.setApi(api);
  }

  isConnected(): boolean {
    if (!this.data.token || !this.data.serverUrl || !this.data.vaultId) {
      return false;
    }
    if (
      this.data.tokenExpiresAt > 0 &&
      this.data.tokenExpiresAt < Math.floor(Date.now() / 1000)
    ) {
      return false;
    }
    return true;
  }

  private injectStyles(): void {
    if (document.getElementById('devitri-plugin-styles')) return;
    const el = document.createElement('style');
    el.id = 'devitri-plugin-styles';
    el.textContent = styles;
    document.head.appendChild(el);
  }

  private onFileCreate(file: { path: string }): void {
    if (!this.isFileSyncable(file)) return;
    this.engine.markDirty(file.path);
  }

  private onFileModify(file: { path: string }): void {
    if (!this.isFileSyncable(file)) return;
    this.engine.markDirty(file.path);
  }

  private onFileDelete(file: { path: string }): void {
    if (!this.isFileSyncable(file)) return;
    this.engine.markDeleted(file.path);
  }

  private onFileRename(file: { path: string }, oldPath: string): void {
    if (!this.isFileSyncable(file)) return;
    this.engine.markRenamed(oldPath, file.path);
  }

  private isFileSyncable(file: { path: string }): boolean {
    return !file.path.startsWith('.obsidian/');
  }

  private startSyncInterval(): void {
    const intervalMs = (this.data.syncInterval ?? 900) * 1000;
    if (intervalMs <= 0) return;

    this.syncIntervalId = window.setInterval(() => {
      void this.startSyncCycle();
    }, intervalMs);
  }

  async startSyncCycle(options?: { notify?: boolean }): Promise<SyncResult | null> {
    if (this.isSyncing || !this.isConnected()) return null;

    if (
      this.data.tokenExpiresAt > 0 &&
      this.data.tokenExpiresAt < Math.floor(Date.now() / 1000)
    ) {
      console.warn('Devitri: Token expired. Reconnect with a new access key.');
      return null;
    }

    const notify = options?.notify ?? true;
    this.isSyncing = true;
    this.setStatusBarSyncing();
    this.settingsTab.display();

    try {
      const result = await this.engine.run();
      this.data.lastSync = Math.floor(Date.now() / 1000);
      this.data.manifestB = this.engine.getManifestB();
      await this.saveData(this.data);

      if (notify) {
        this.showSyncSuccess(result);
      } else {
        this.updateStatusBarIdle();
      }

      return result;
    } catch (err) {
      console.error('Devitri: Sync cycle failed', err);
      if (notify) {
        this.showSyncFailure(err);
      } else {
        this.updateStatusBarIdle();
      }
      return null;
    } finally {
      this.isSyncing = false;
      this.settingsTab.display();
    }
  }

  private setupStatusBar(): void {
    this.statusBarItem = this.addStatusBarItem();
    this.statusBarItem.addClass('devitri-status-bar');
    this.statusBarItem.setText('Devitri');
    this.updateStatusBarIdle();
  }

  private setStatusBarText(text: string): void {
    this.statusBarItem?.setText(text);
  }

  private setStatusBarSyncing(): void {
    if (!this.isConnected()) return;
    this.setStatusBarText('Devitri · syncing…');
  }

  updateStatusBarIdle(): void {
    if (!this.statusBarItem) return;

    if (!this.isConnected()) {
      this.setStatusBarText('Devitri · offline');
      return;
    }

    if (this.isSyncing) {
      this.setStatusBarText('Devitri · syncing…');
      return;
    }

    const last = this.data.lastSync || 0;
    if (last === 0) {
      this.setStatusBarText('Devitri · connected');
      return;
    }

    this.setStatusBarText(`Devitri · synced ${formatRelativeSync(last)}`);
  }

  private scheduleStatusBarIdleReset(delayMs = 5000): void {
    if (this.statusBarResetTimer !== null) {
      clearTimeout(this.statusBarResetTimer);
    }
    this.statusBarResetTimer = window.setTimeout(() => {
      this.statusBarResetTimer = null;
      this.updateStatusBarIdle();
    }, delayMs);
  }

  private showSyncSuccess(result: SyncResult): void {
    const summary = formatSyncSummary(result);
    this.setStatusBarText(`Devitri · synced${summary}`);
    this.scheduleStatusBarIdleReset();

    // Brief toast — visible while editing without opening settings
    new Notice(`Devitri: vault synced${summary}`, 2200);
  }

  private showSyncFailure(err: unknown): void {
    const message = err instanceof Error ? err.message : String(err);
    this.setStatusBarText('Devitri · sync failed');
    this.scheduleStatusBarIdleReset(6000);
    new Notice(`Devitri: sync failed — ${message}`, 4000);
  }

  private getInitialData(): PluginData {
    const deviceId =
      localStorage.getItem('devitri_device_id') || this.generateDeviceId();
    localStorage.setItem('devitri_device_id', deviceId);

    return {
      serverUrl: '',
      vaultId: '',
      token: '',
      tokenExpiresAt: 0,
      deviceId,
      deviceName: this.getDeviceName(),
      syncInterval: 900,
      lastSync: 0,
      manifestB: this.getInitialManifest(),
      conflicts: [],
    };
  }

  private getInitialManifest(): SyncManifest {
    return { vault_id: '', generated_at: 0, files: [] };
  }

  private generateDeviceId(): string {
    return `d${Date.now().toString(36)}-${Math.random().toString(36).slice(2, 11)}`;
  }

  private getDeviceName(): string {
    const platform = Platform.isMobile ? 'mobile' : 'desktop';
    return `${this.app.vault.getName()}-${platform}`;
  }
}

// Obsidian loads community plugins via CommonJS `module.exports = PluginClass`.
module.exports = DevitriPlugin;
