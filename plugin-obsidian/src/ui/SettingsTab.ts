/**
 * Devitri Settings Tab
 */

import {App, ButtonComponent, Notice, PluginSettingTab, Setting} from 'obsidian';
import type {DevitriPluginHost} from '../types';
import {DevitriSyncEngine} from '../sync/engine';
import {DevitriApi} from '../sync/api';
import {formatTimestamp} from '../sync/conflicts';

type ConnectionState = 'connected' | 'disconnected' | 'checking';

export class DevitriSettingsTab extends PluginSettingTab {
  private plugin: DevitriPluginHost;
  private engine: DevitriSyncEngine;
  private pendingAccessKey = '';
  private connectionState: ConnectionState = 'disconnected';
  private connectionMessage = 'Not connected';
  private statusBannerEl: HTMLElement | null = null;
  private connectButton: ButtonComponent | null = null;
  private verifyButton: ButtonComponent | null = null;
  private manualSyncButton: ButtonComponent | null = null;
  private actionInProgress = false;

  constructor(
    app: App,
    plugin: DevitriPluginHost,
    _api: DevitriApi,
    engine: DevitriSyncEngine,
  ) {
    super(app, plugin);
    this.plugin = plugin;
    this.engine = engine;
  }

  display(): void {
    const {containerEl} = this;
    containerEl.empty();
    containerEl.addClass('devitri-settings');

    containerEl.createEl('h2', {text: 'Devitri'});

    this.statusBannerEl = this.createStatusBanner(containerEl);
    this.updateStatusBanner();

    this.createConnectionSection(containerEl);
    this.createSyncControlSection(containerEl);
    this.createBulkDeleteSection(containerEl);
    this.createConflictSection(containerEl);
    this.createDangerZoneSection(containerEl);

    if (this.plugin.isConnected()) {
      this.setConnectionState('connected', this.getLocalConnectionMessage());
      void this.verifyConnection(false);
    } else {
      this.setConnectionState('disconnected', this.getLocalConnectionMessage());
    }
  }

  private createStatusBanner(container: HTMLElement): HTMLElement {
    const banner = container.createDiv('devitri-status-banner');
    banner.createDiv('devitri-status-dot');
    const body = banner.createDiv('devitri-status-body');
    body.createDiv('devitri-status-label');
    body.createDiv('devitri-status-detail');
    return banner;
  }

  private updateStatusBanner(): void {
    if (!this.statusBannerEl) return;

    const dot = this.statusBannerEl.querySelector('.devitri-status-dot');
    const label = this.statusBannerEl.querySelector('.devitri-status-label');
    const detail = this.statusBannerEl.querySelector('.devitri-status-detail');

    dot?.removeClass(
      'devitri-status-dot--connected',
      'devitri-status-dot--disconnected',
      'devitri-status-dot--checking',
    );
    dot?.addClass(`devitri-status-dot--${this.connectionState}`);

    const title =
      this.connectionState === 'connected'
        ? 'Connected'
        : this.connectionState === 'checking'
          ? 'Checking connection…'
          : 'Not connected';

    if (label) label.textContent = title;
    if (detail) detail.textContent = this.connectionMessage;
  }

  private setConnectionState(state: ConnectionState, message: string): void {
    this.connectionState = state;
    this.connectionMessage = message;
    this.updateStatusBanner();
    this.updateActionButtonsState();
  }

  private areActionsDisabled(): boolean {
    return (
      this.plugin.isSyncing ||
      this.actionInProgress ||
      this.connectionState === 'checking'
    );
  }

  private updateActionButtonsState(): void {
    const disabled = this.areActionsDisabled();
    this.connectButton?.setDisabled(disabled);
    this.verifyButton?.setDisabled(disabled);
    this.manualSyncButton?.setDisabled(disabled);
  }

  private async withActionsLocked(fn: () => Promise<void>): Promise<void> {
    if (this.areActionsDisabled()) return;

    this.actionInProgress = true;
    this.updateActionButtonsState();
    try {
      await fn();
    } finally {
      this.actionInProgress = false;
      this.updateActionButtonsState();
    }
  }

  private async runSyncCycle(): Promise<{
    downloaded: number;
    uploaded: number;
  }> {
    const result = await this.plugin.startSyncCycle({ notify: false });
    if (!result) {
      throw new Error('Sync did not run');
    }
    return result;
  }

  private getLocalConnectionMessage(): string {
    const {serverUrl, vaultId, token, deviceId, tokenExpiresAt} =
      this.plugin.data;

    if (!serverUrl || !vaultId) {
      return 'Configure Server URL and Vault ID below.';
    }
    if (!token) {
      return 'Paste an access key from Dashboard → Connect.';
    }
    if (tokenExpiresAt > 0 && tokenExpiresAt < Math.floor(Date.now() / 1000)) {
      return 'Session expired. Generate a new access key.';
    }

    const parts = [`Server: ${serverUrl}`, `Vault: ${vaultId}`];
    if (deviceId) parts.push(`Device: ${deviceId}`);
    return parts.join(' · ');
  }

  private async verifyConnection(notify: boolean): Promise<boolean> {
    if (!this.plugin.isConnected()) {
      this.setConnectionState('disconnected', this.getLocalConnectionMessage());
      if (notify) {
        new Notice('Devitri: Not connected. Check your settings.', 3000);
      }
      return false;
    }

    this.setConnectionState('checking', this.getLocalConnectionMessage());

    try {
      const api = this.plugin.getApi();
      api.updateCredentials(
        this.plugin.data.serverUrl,
        this.plugin.data.vaultId,
        this.plugin.data.token,
        this.plugin.data.deviceId,
      );

      const session = await api.verifySession();
      this.plugin.data.deviceId = session.device_id;
      this.plugin.data.tokenExpiresAt = session.expires_at;
      await this.plugin.saveData(this.plugin.data);

      this.setConnectionState('connected', this.getLocalConnectionMessage());
      if (notify) {
        new Notice('Devitri: Connection verified', 2500);
      }
      return true;
    } catch (err) {
      const message = err instanceof Error ? err.message : String(err);
      this.setConnectionState('disconnected', message);
      if (notify) {
        new Notice(`Devitri: Connection failed — ${message}`, 5000);
      }
      return false;
    }
  }

  private createConnectionSection(container: HTMLElement): void {
    const section = container.createDiv('devitri-section');
    section.createEl('h3', {text: 'Server'});

    new Setting(section)
      .setName('Server URL')
      .setDesc(
        'API base URL. Dev: http://localhost:8080. Production: your Devitri API base URL (ie. https://api.devitri.yourdomain.com)',
      )
      .addText(text =>
        text
          .setValue(this.plugin.data.serverUrl || '')
          .setPlaceholder('https://api.devitri.yourdomain.com')
          .onChange(async value => {
            this.plugin.data.serverUrl = value.trim();
            await this.plugin.saveData(this.plugin.data);
            this.setConnectionState(
              'disconnected',
              this.getLocalConnectionMessage(),
            );
          }),
      );

    new Setting(section)
      .setName('Vault ID')
      .setDesc(
        'Derived from the currently open Obsidian vault name. Devitri uses it as the server vault identifier.',
      )
      .addText(text =>
        text.setValue(this.plugin.data.vaultId || '').setDisabled(true),
      )
      .addExtraButton(btn =>
        btn
          .setIcon('rotate-ccw')
          .setTooltip('Reset to current vault name')
          .onClick(async () => {
            const name = this.app.vault.getName();
            const vaultId = name
              .trim()
              .toLowerCase()
              .replace(/[^a-z0-9]+/g, '-')
              .replace(/^-+|-+$/g, '');
            this.plugin.data.vaultId = vaultId;
            await this.plugin.saveData(this.plugin.data);
            this.setConnectionState(
              'disconnected',
              this.getLocalConnectionMessage(),
            );
            new Notice(`Devitri: Vault ID set to ${vaultId}`, 2500);
            this.display();
          }),
      );

    new Setting(section)
      .setName('Access Key')
      .setDesc(
        'Generate in Dashboard → Connect. Never stored — only the JWT token is saved.',
      )
      .addText(text => {
        text.inputEl.type = 'password';
        text.setPlaceholder('Paste access key from dashboard');
        text.onChange(value => {
          this.pendingAccessKey = value;
        });
      });

    new Setting(section)
      .setName('Connect')
      .setDesc('Validate credentials and run an initial sync')
      .addButton(button => {
        this.connectButton = button;
        button
          .setButtonText('Connect & Sync')
          .setCta()
          .setDisabled(this.areActionsDisabled())
          .onClick(async () => {
            await this.handleConnect(true);
          });
      })
      .addButton(button => {
        this.verifyButton = button;
        button
          .setButtonText('Verify')
          .setTooltip('Check server connection without syncing')
          .setDisabled(this.areActionsDisabled())
          .onClick(async () => {
            if (this.pendingAccessKey) {
              await this.handleConnect(false);
              return;
            }
            await this.withActionsLocked(async () => {
              await this.verifyConnection(true);
            });
          });
      });
  }

  private async handleConnect(runSync: boolean): Promise<void> {
    if (this.areActionsDisabled()) return;

    if (!this.pendingAccessKey && !this.plugin.data.token) {
      new Notice('Devitri: Please paste an access key', 3000);
      return;
    }
    if (!this.plugin.data.serverUrl || !this.plugin.data.vaultId) {
      new Notice('Devitri: Configure Server URL and Vault ID first', 3000);
      return;
    }

    await this.withActionsLocked(async () => {
      try {
        const freshApi = new DevitriApi(
          this.plugin.data.serverUrl,
          this.plugin.data.vaultId,
          this.plugin.data.token,
          this.plugin.data.deviceId || '',
        );

        if (this.pendingAccessKey) {
          const session = await freshApi.connectWithAccessKey(
            this.pendingAccessKey,
          );
          this.plugin.data.token = freshApi.getToken();
          this.plugin.data.deviceId = session.device_id;
          this.plugin.data.tokenExpiresAt = session.expires_at;
          this.pendingAccessKey = '';
        } else {
          const session = await freshApi.verifySession();
          this.plugin.data.deviceId = session.device_id;
          this.plugin.data.tokenExpiresAt = session.expires_at;
        }

        this.plugin.replaceApi(freshApi);
        await this.plugin.saveData(this.plugin.data);

        this.setConnectionState('connected', this.getLocalConnectionMessage());
        this.plugin.updateStatusBarIdle();
        new Notice('Devitri: Connected successfully!', 3000);

        if (runSync) {
          const result = await this.runSyncCycle();
          new Notice(
            `Devitri: Sync complete — ${result.downloaded} downloaded, ${result.uploaded} uploaded`,
            4000,
          );
        }

        this.display();
      } catch (err) {
        const message = err instanceof Error ? err.message : String(err);
        console.error('Devitri: Connection failed', err);
        this.setConnectionState('disconnected', message);
        new Notice(`Devitri: Connection failed — ${message}`, 5000);
      }
    });
  }

  private createSyncControlSection(container: HTMLElement): void {
    const section = container.createDiv('devitri-section');
    section.createEl('h3', {text: 'Sync'});

    new Setting(section)
      .setName('Sync Interval')
      .setDesc(
        'Seconds between sync cycles (0 = manual only). Default is 900 seconds (15 minutes).',
      )
      .addText(text =>
        text
          .setValue(String(this.plugin.data.syncInterval || 900))
          .setPlaceholder('900')
          .onChange(async value => {
            const interval = parseInt(value, 10) || 0;
            this.plugin.data.syncInterval = interval;
            await this.plugin.saveData(this.plugin.data);
            this.plugin.restartSyncInterval();
          }),
      );

    new Setting(section)
      .setName('Manual Sync')
      .setDesc(`Last sync: ${formatTimestamp(this.plugin.data.lastSync || 0)}`)
      .addButton(button => {
        this.manualSyncButton = button;
        button
          .setButtonText('Sync Now')
          .setTooltip('Trigger immediate sync')
          .setDisabled(this.areActionsDisabled())
          .onClick(async () => {
            if (!this.plugin.isConnected()) {
              new Notice(
                'Devitri: Not connected. Configure server and access key.',
                3000,
              );
              this.setConnectionState(
                'disconnected',
                this.getLocalConnectionMessage(),
              );
              return;
            }

            await this.withActionsLocked(async () => {
              try {
                new Notice('Devitri: Syncing...', 1000);
                const result = await this.runSyncCycle();
                new Notice(
                  `Devitri: Sync complete — ${result.downloaded} downloaded, ${result.uploaded} uploaded`,
                  4000,
                );
                this.display();
              } catch (err) {
                const message =
                  err instanceof Error ? err.message : String(err);
                new Notice(`Devitri: Sync failed — ${message}`, 5000);
              }
            });
          });
      });

    new Setting(section)
      .setName('Activity')
      .addText(text => text.setValue(this.getActivityText()).setDisabled(true));
  }

  private createBulkDeleteSection(container: HTMLElement): void {
    const pending = this.engine.getPendingBulkDelete();
    if (!pending) {
      return;
    }

    const section = container.createDiv('devitri-section devitri-warning');
    section.createEl('h3', {text: 'Bulk delete'});

    new Setting(section)
      .setName('Confirmation required')
      .setDesc(
        `The next sync would remove ${pending.count} file(s) from this device (above the safety threshold). ` +
          'Confirm once, then run Sync Now.',
      )
      .addButton(button =>
        button.setButtonText('Confirm once').setWarning().onClick(() => {
          this.engine.setBulkDeleteConfirmed(true);
          new Notice(
            'Devitri: Bulk delete confirmed for the next sync. Tap Sync Now.',
            6000,
          );
          this.display();
        }),
      );
  }

  private createConflictSection(container: HTMLElement): void {
    const section = container.createDiv('devitri-section');
    const conflictCount = this.engine.getConflictCount();
    section.createEl('h3', {text: `Conflicts (${conflictCount})`});

    if (conflictCount === 0) {
      section.createEl('p', {
        cls: 'devitri-muted',
        text: 'No conflicts detected.',
      });
      return;
    }

    for (const conflict of this.engine.getConflicts()) {
      const row = section.createDiv('devitri-conflict-row');
      row.createEl('div', {cls: 'devitri-conflict-name', text: conflict.path});
      row.createEl('div', {
        cls: 'devitri-muted',
        text: `Detected: ${formatTimestamp(conflict.detected_at)}`,
      });
    }
  }

  private createDangerZoneSection(container: HTMLElement): void {
    const section = container.createDiv('devitri-section devitri-danger');
    section.createEl('h3', {text: 'Danger Zone'});

    new Setting(section)
      .setName('Reset Local Sync State')
      .setDesc(
        'Clears the stored manifest. The next sync will re-download all files from the server.',
      )
      .addButton(button =>
        button
          .setButtonText('Reset')
          .setWarning()
          .onClick(async () => {
            this.plugin.data.manifestB = {
              vault_id: this.plugin.data.vaultId || '',
              generated_at: 0,
              files: [],
            };
            await this.plugin.saveData(this.plugin.data);
            this.engine.setManifestB(this.plugin.data.manifestB);
            new Notice('Devitri: Sync state reset.', 5000);
            this.display();
          }),
      );

    if (this.plugin.isConnected()) {
      new Setting(section)
        .setName('Disconnect')
        .setDesc('Clears stored token. Generate a new access key to reconnect.')
        .addButton(button =>
          button
            .setButtonText('Disconnect')
            .setWarning()
            .onClick(async () => {
              this.plugin.data.token = '';
              this.plugin.data.tokenExpiresAt = 0;
              this.plugin.data.lastSync = 0;
              this.plugin
                .getApi()
                .updateCredentials(
                  this.plugin.data.serverUrl,
                  this.plugin.data.vaultId,
                  '',
                  this.plugin.data.deviceId,
                );
              await this.plugin.saveData(this.plugin.data);
              this.pendingAccessKey = '';
              this.setConnectionState(
                'disconnected',
                this.getLocalConnectionMessage(),
              );
              new Notice('Devitri: Disconnected.', 3000);
              this.display();
            }),
        );
    }
  }

  private getActivityText(): string {
    if (!this.plugin.isConnected()) return 'Waiting for connection';
    if (this.plugin.isSyncing) return 'Syncing…';

    const lastSync = this.plugin.data.lastSync || 0;
    if (lastSync === 0) return 'Connected — no sync yet';

    const elapsed = Math.floor((Date.now() / 1000 - lastSync) / 60);
    if (elapsed < 1) return 'Last sync: just now';
    if (elapsed < 60) return `Last sync: ${elapsed} min ago`;
    if (elapsed < 1440) return `Last sync: ${Math.floor(elapsed / 60)} h ago`;
    return `Last sync: ${Math.floor(elapsed / 1440)} d ago`;
  }
}
