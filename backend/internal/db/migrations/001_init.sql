-- Vaults registered in the server
-- The vault_id is defined by the user when configuring the plugin (e.g. "personal", "work")
-- The path on disk will be /vaults/{vault_id}/
CREATE TABLE vaults (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    vault_id    TEXT    NOT NULL UNIQUE,   -- Slug defined by the user (e.g. "personal")
    name        TEXT    NOT NULL,          -- Human-readable name (e.g. "Personal")
    description TEXT,
    path        TEXT    NOT NULL,          -- Absolute path in the container (/vaults/personal)
    created_at  INTEGER NOT NULL,
    last_sync   INTEGER                    -- Unix timestamp of the last successful sync
);

-- Sessions table for tracking JWT tokens
CREATE TABLE sessions (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    token_hash  TEXT    NOT NULL UNIQUE,   -- SHA-256 of the JWT
    device_id   TEXT    NOT NULL,          -- Client-provided or generated ID
    device_name TEXT,                      -- User-friendly device name
    vault_id    TEXT    REFERENCES vaults(vault_id), -- Optional: scope to specific vault
    created_at  INTEGER NOT NULL,          -- Unix timestamp
    expires_at  INTEGER NOT NULL,          -- Unix timestamp (24h TTL)
    last_seen   INTEGER                    -- Updated on every authorized request
);