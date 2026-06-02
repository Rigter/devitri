-- Track current state of files in each vault
CREATE TABLE IF NOT EXISTS files (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    vault_id    TEXT    NOT NULL REFERENCES vaults(vault_id),
    path        TEXT    NOT NULL,
    hash_sha256 TEXT    NOT NULL,
    size_bytes  INTEGER NOT NULL,
    modified_at INTEGER NOT NULL,
    deleted_at  INTEGER, -- Soft delete for sync tracking
    UNIQUE(vault_id, path)
);

-- Index for manifest lookups
CREATE INDEX IF NOT EXISTS idx_files_vault_path ON files(vault_id, path);

-- Detailed sync history
CREATE TABLE IF NOT EXISTS sync_log (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    vault_id    TEXT    NOT NULL REFERENCES vaults(vault_id),
    device_id   TEXT    NOT NULL,
    operation   TEXT    NOT NULL, -- 'upload', 'download', 'delete', 'conflict'
    file_path   TEXT    NOT NULL,
    status      TEXT    NOT NULL, -- 'success', 'error', 'pending'
    detail      TEXT,
    created_at  INTEGER NOT NULL
);