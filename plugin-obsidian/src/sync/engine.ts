/**
 * Sync Engine – 3-Way Sync State Machine
 */

import { App, normalizePath, TFile, Vault } from 'obsidian';
import { BulkDeleteBlockedError, DevitriApi } from './api';
import {
  FileState,
  normalizeSyncManifest,
  SyncDecision,
  SyncManifest,
  SyncResult,
} from '../types';
import { computeHash, buildLocalManifest } from './manifest';
import { threeWayMerge, createConflictCopy } from './conflicts';

function isTFile(file: unknown): file is TFile {
  return file instanceof TFile;
}

export class DevitriSyncEngine {
  private vault: Vault;
  private api: DevitriApi;
  private manifestB: SyncManifest;
  private conflictCount = 0;
  private conflicts: { path: string; detected_at: number }[] = [];
  private dirtyFiles = new Set<string>();
  private deletedFiles = new Set<string>();
  private deleteThresholdCount = 20;
  private deleteThresholdPercent = 5;
  private bulkDeleteConfirmed = false;
  private pendingBulkDelete: { count: number } | null = null;

  constructor(app: App, api: DevitriApi, manifestB: SyncManifest) {
    this.vault = app.vault;
    this.api = api;
    this.manifestB = normalizeSyncManifest(manifestB, api.vaultId);
  }

  private updateBaseManifest(file: string, hash: string, mtime: number): void {
    const existingIndex = this.manifestB.files.findIndex((f) => f.path === file);
    const entry: FileState = { path: file, hash, modified_at: mtime, size: 0 };
    if (existingIndex >= 0) {
      this.manifestB.files[existingIndex] = entry;
    } else {
      this.manifestB.files.push(entry);
    }
  }

  public markDirty(file: string): void {
    this.dirtyFiles.add(file);
    this.deletedFiles.delete(file);
  }

  public markDeleted(file: string): void {
    this.dirtyFiles.delete(file);
    this.deletedFiles.add(file);
  }

  public markRenamed(oldPath: string, newPath: string): void {
    this.dirtyFiles.delete(oldPath);
    this.dirtyFiles.add(newPath);
    this.deletedFiles.delete(newPath);
    this.deletedFiles.add(oldPath);
  }

  public getManifestB(): SyncManifest {
    return this.manifestB;
  }

  public setManifestB(manifest: SyncManifest): void {
    this.manifestB = normalizeSyncManifest(manifest, this.api.vaultId);
  }

  public setApi(api: DevitriApi): void {
    this.api = api;
  }

  public setBulkDeleteConfirmed(confirmed: boolean): void {
    this.bulkDeleteConfirmed = confirmed;
  }

  public getPendingBulkDelete(): { count: number } | null {
    return this.pendingBulkDelete;
  }

  private markPendingBulkDelete(count: number): void {
    if (count > 0) {
      this.pendingBulkDelete = { count };
    }
  }

  private clearBulkDeleteState(): void {
    this.bulkDeleteConfirmed = false;
    this.pendingBulkDelete = null;
  }

  private async scan(): Promise<SyncManifest> {
    // First sync (empty base): scan the whole vault. Incremental sync only
    // refreshes paths touched since the last cycle via dirtyFiles.
    if (this.manifestB.files.length === 0) {
      const full = await buildLocalManifest(this.vault);
      const files = full.files.filter((f) => !this.deletedFiles.has(f.path));
      return {
        vault_id: this.api.vaultId,
        generated_at: Math.floor(Date.now() / 1000),
        files,
      };
    }

    const files: FileState[] = this.manifestB.files.map((f) => ({ ...f }));

    for (const path of this.dirtyFiles) {
      const existing = files.find((f) => f.path === path);
      const file = this.vault.getAbstractFileByPath(path);
      if (isTFile(file) && file.stat.size > 0) {
        const content = await this.vault.readBinary(file);
        const hash = await computeHash(content);
        if (existing) {
          existing.hash = hash;
          existing.modified_at = file.stat.mtime;
          existing.size = file.stat.size;
        } else {
          files.push({
            path,
            hash,
            modified_at: file.stat.mtime,
            size: file.stat.size,
          });
        }
      }
    }

    for (const path of this.deletedFiles) {
      const idx = files.findIndex((f) => f.path === path);
      if (idx >= 0) {
        files.splice(idx, 1);
      }
    }

    return {
      vault_id: this.api.vaultId,
      generated_at: Math.floor(Date.now() / 1000),
      files,
    };
  }

  private negotiate(L: SyncManifest, R: SyncManifest): SyncDecision[] {
    const decisions: SyncDecision[] = [];
    const localManifest = normalizeSyncManifest(L, L.vault_id);
    const remoteManifest = normalizeSyncManifest(R, L.vault_id);
    const B = normalizeSyncManifest(this.manifestB, localManifest.vault_id);
    const bMap = new Map(B.files.map((f) => [f.path, f]));
    const rMap = new Map(remoteManifest.files.map((f) => [f.path, f]));

    for (const local of localManifest.files) {
      const base = bMap.get(local.path);
      const remoteFile = rMap.get(local.path);
      const HL = local.hash;
      const HB = base?.hash ?? null;
      const HR = remoteFile?.hash ?? null;

      if (HB === null && HR === null) {
        decisions.push({ path: local.path, action: 'upload', local });
      } else if (HB !== null && HR !== null && HL === HB && HR === HB) {
        decisions.push({ path: local.path, action: 'skip', local, base, remote: remoteFile });
      } else if (HB !== null && HL !== HB && HR === HB) {
        decisions.push({ path: local.path, action: 'upload', local, base, remote: remoteFile });
      } else if (HB !== null && HR !== HB && HL === HB) {
        decisions.push({ path: local.path, action: 'download', local, base, remote: remoteFile });
      } else if (HB !== null && HL !== HB && HR !== HB && HL !== HR) {
        decisions.push({ path: local.path, action: 'conflict', local, base, remote: remoteFile });
      } else if (HB !== null && HL === HB && HR !== HB) {
        decisions.push({ path: local.path, action: 'skip', local, base, remote: remoteFile });
      }
    }

    for (const remoteFile of remoteManifest.files) {
      if (!localManifest.files.find((f) => f.path === remoteFile.path)) {
        const base = bMap.get(remoteFile.path);
        decisions.push({
          path: remoteFile.path,
          action: 'download',
          base,
          remote: remoteFile,
        });
      }
    }

    return decisions;
  }

  private async ensureParentFolder(path: string): Promise<void> {
    const dir = path.split('/').slice(0, -1).join('/');
    if (dir && this.vault.getAbstractFileByPath(dir) === null) {
      await this.vault.createFolder(dir);
    }
  }

  private async pull(toDownload: string[], R: SyncManifest): Promise<number> {
    let downloaded = 0;
    const remoteManifest = normalizeSyncManifest(R, this.api.vaultId);

    for (const path of toDownload) {
      const remote = remoteManifest.files.find((f) => f.path === path);
      if (!remote) continue;

      try {
        const response = await this.api.downloadFile(path);
        const content = response.arrayBuffer;
        const hash = await computeHash(content);
        if (hash !== remote.hash) {
          console.error(`Devitri: Hash mismatch for ${path}, skipping download`);
          continue;
        }

        await this.ensureParentFolder(path);
        await this.vault.adapter.writeBinary(normalizePath(path), content);

        const file = this.vault.getAbstractFileByPath(path);
        if (isTFile(file)) {
          this.updateBaseManifest(path, hash, file.stat.mtime);
        } else {
          this.updateBaseManifest(path, hash, remote.modified_at);
        }

        downloaded++;
      } catch (err) {
        console.error(`Devitri: Failed to download ${path}`, err);
      }
    }

    for (const path of toDownload) {
      this.dirtyFiles.delete(path);
    }

    return downloaded;
  }

  private async push(
    toUpload: FileState[],
    toDelete: string[]
  ): Promise<{ uploaded: number; deleted: number }> {
    let uploaded = 0;
    let deleted = 0;

    for (const fileState of toUpload) {
      const file = this.vault.getAbstractFileByPath(fileState.path);
      if (!isTFile(file) || file.stat.size === 0) {
        console.warn(`Devitri: File not found or empty for upload: ${fileState.path}`);
        continue;
      }

      try {
        const content = await this.vault.readBinary(file);
        await this.api.uploadFile(fileState.path, fileState.hash, content);
        this.updateBaseManifest(fileState.path, fileState.hash, fileState.modified_at);
        uploaded++;
      } catch (err) {
        console.error(`Devitri: Failed to upload ${fileState.path}`, err);
      }
    }

    for (const path of toDelete) {
      try {
        await this.api.deleteFile(path, this.bulkDeleteConfirmed);
        this.manifestB.files = this.manifestB.files.filter((f) => f.path !== path);
        deleted++;
      } catch (err) {
        if (err instanceof BulkDeleteBlockedError) {
          this.markPendingBulkDelete(toDelete.length);
          throw err;
        }
        console.error(`Devitri: Failed to delete ${path}`, err);
      }
    }

    if (deleted > 0) {
      this.clearBulkDeleteState();
    }

    return { uploaded, deleted };
  }

  private async handleConflicts(
    conflicts: SyncDecision[]
  ): Promise<{ conflictsResolved: number; newConflicts: number }> {
    let conflictsResolved = 0;
    let newConflicts = 0;

    for (const decision of conflicts) {
      if (!decision.path || !decision.local || !decision.remote) {
        continue;
      }

      try {
        if (decision.path.endsWith('.md')) {
          const file = this.vault.getAbstractFileByPath(decision.path);
          if (!isTFile(file)) {
            newConflicts++;
            continue;
          }

          const localContent = await this.vault.read(file);
          const remoteResponse = await this.api.downloadFile(decision.path);
          const remoteContent = new TextDecoder().decode(remoteResponse.arrayBuffer);

          if (localContent === remoteContent) {
            continue;
          }

          const baseContent = '';
          const result = threeWayMerge(baseContent, localContent, remoteContent);

          if (result.success && result.content) {
            await this.vault.modify(file, result.content);
            const newContent = await this.vault.readBinary(file);
            const newHash = await computeHash(newContent);
            await this.api.uploadFile(decision.path, newHash, newContent);
            this.updateBaseManifest(decision.path, newHash, file.stat.mtime);
            conflictsResolved++;
          } else {
            const conflictPath = createConflictCopy(decision.path, this.api.deviceId);
            await this.vault.create(conflictPath, remoteContent);
            this.conflicts.push({
              path: decision.path,
              detected_at: Math.floor(Date.now() / 1000),
            });
            newConflicts++;
          }
        } else {
          newConflicts++;
          this.conflicts.push({
            path: decision.path,
            detected_at: Math.floor(Date.now() / 1000),
          });
        }
      } catch (err) {
        console.error(`Devitri: Conflict handling failed for ${decision.path}`, err);
        newConflicts++;
      }
    }

    this.conflictCount = this.conflicts.length;
    return { conflictsResolved, newConflicts };
  }

  private checkBulkDelete(toDelete: string[], totalFiles: number): boolean {
    if (toDelete.length === 0) return false;
    return (
      toDelete.length > this.deleteThresholdCount ||
      toDelete.length > totalFiles * (this.deleteThresholdPercent / 100)
    );
  }

  private async loadDeleteThresholds(): Promise<void> {
    try {
      const settings = await this.api.getSettings();
      this.deleteThresholdCount = settings.sync.delete_threshold_count;
      this.deleteThresholdPercent = settings.sync.delete_threshold_percent;
    } catch {
      // Keep defaults if settings endpoint is unavailable
    }
  }

  public async run(): Promise<SyncResult> {
    const result: SyncResult = {
      uploaded: 0,
      downloaded: 0,
      conflicts: 0,
      errors: [],
    };

    try {
      await this.loadDeleteThresholds();

      const L = await this.scan();
      const R = await this.api.getManifest();
      const decisions = this.negotiate(L, R);

      const toUpload = decisions
        .filter((d) => d.action === 'upload')
        .map((d) => d.local!)
        .filter(Boolean);
      const toDownload = decisions
        .filter((d) => d.action === 'download')
        .map((d) => d.path);
      const conflicts = decisions.filter((d) => d.action === 'conflict');
      const toDelete = decisions
        .filter((d) => d.action === 'delete')
        .map((d) => d.path);

      if (this.checkBulkDelete(toDelete, L.files.length) && !this.bulkDeleteConfirmed) {
        this.markPendingBulkDelete(toDelete.length);
        const errorMsg = `Devitri: Bulk delete of ${toDelete.length} files blocked. Confirm once in Devitri settings, then sync again.`;
        console.error(errorMsg);
        result.errors.push(errorMsg);
        return result;
      }

      result.downloaded = await this.pull(toDownload, R);

      const conflictResult = await this.handleConflicts(conflicts);
      result.conflicts = conflictResult.newConflicts;

      try {
        const pushResult = await this.push(toUpload, toDelete);
        result.uploaded = pushResult.uploaded;
      } catch (err) {
        if (err instanceof BulkDeleteBlockedError) {
          const message =
            'Devitri: Server blocked bulk delete. Confirm once in Devitri settings, then sync again.';
          result.errors.push(message);
          return result;
        }
        throw err;
      }

      this.manifestB.generated_at = Math.floor(Date.now() / 1000);
      this.dirtyFiles.clear();
      this.deletedFiles.clear();

      return result;
    } catch (err) {
      const message = err instanceof Error ? err.message : String(err);
      console.error('Devitri: Sync cycle failed', err);
      result.errors.push(message);
      return result;
    }
  }

  public getConflictCount(): number {
    return this.conflictCount;
  }

  public getConflicts(): { path: string; detected_at: number }[] {
    return this.conflicts;
  }
}
