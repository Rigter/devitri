---
project: Devitri
type: Architecture
status: active
tags: [obsidian, sync, self-hosted, golang, sveltekit, open-source]
created: 2026-05-29
updated: 2026-06-02
theme: Nano v1 — Zinc
---

# Devitri: Local-First Sync and Self-Hosted Web Dashboard for Obsidian

**Devitri** is a bidirectional sync and self-hosted web visualization system for Obsidian vaults. It addresses limitations of iCloud sync, orphan `.icloud` files, and Git friction on mobile.

**Related docs:** [`DESIGN.md`](./DESIGN.md) (UI) · [`README.md`](./README.md) (getting started) · [`.env.example`](./.env.example) (configuration)

---

## 1. Name and Concept

*Pronunciation:* **DEH-vih-tree** · `/ˈdɛvɪtri/` (from *devitrification*)

In geology, **devitrification** is the process by which volcanic glass (obsidian) reorganizes into stable microcrystals.

| State | Meaning |
|-------|---------|
| Amorphous glass | Scattered notes and unstructured fragments |
| Devitrification | Consistent sync, integrity checks, change control |
| Ordered crystal | Structured, accessible PKM on every device |

---

## 2. General Architecture

Devitri typically runs as **two containers** (frontend + backend) on a Docker network. Each service stays focused and replaceable.

```
┌─────────────────────────────────────────────────────────────────┐
│                     DOCKER NETWORK (devitri)                    │
│  ┌──────────────────────┐        ┌──────────────────────────┐  │
│  │  FRONTEND (SvelteKit) │◄──────►│  BACKEND (Go API)        │  │
│  │  static / :80         │  HTTP  │  :8080                   │  │
│  └──────────────────────┘        └──────────┬───────────────┘  │
│                                             │                   │
│                              ┌──────────────▼───────────────┐  │
│                              │  /vaults  → user notes        │  │
│                              │  /data    → SQLite + state    │  │
│                              └──────────────────────────────┘  │
└─────────────────────────────────────────────────────────────────┘
                                      │
                         Reverse proxy (Traefik or Caddy)
                                      │
                               Internet / Tailscale
```

A reverse proxy (Traefik or Caddy) is the single TLS entry point in production. For development, expose ports directly (`deploy/dev/docker-compose.yml`).

### Deployment profiles

| Profile | API | Dashboard | Notes |
|---------|-----|-----------|-------|
| **A** | Public HTTPS | Local (`npm run dev`) | Set `VITE_DEVITRI_BACKEND_URL`; list localhost origins in `DEVITRI_CORS_ORIGINS` |
| **B** | Public HTTPS | Public HTTPS | Set `DEVITRI_CORS_ORIGINS` to dashboard origin(s) |
| **C** | Local Docker | Local Docker `:3000` | Default CORS includes `:3000`, `:5173`, `:4173` |

The Obsidian plugin always uses `Authorization: Bearer` via `requestUrl` — **not** browser CORS. Use **HTTPS** for any internet-exposed API.

---

## 3. Backend (Go)

### Why Go

- Low idle memory (≈10–30 MB)
- Single static binary
- Native concurrency for parallel syncs
- Strong stdlib for HTTP, crypto, and files

### Manifest database (SQLite)

SQLite lives under `/data`, separate from `/vaults`, so note directories stay clean.

#### Schema

```sql
CREATE TABLE vaults (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    vault_id    TEXT    NOT NULL UNIQUE,
    name        TEXT    NOT NULL,
    description TEXT,
    path        TEXT    NOT NULL,
    created_at  INTEGER NOT NULL,
    last_sync   INTEGER
);

CREATE TABLE files (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    vault_id    TEXT    NOT NULL REFERENCES vaults(vault_id),
    path        TEXT    NOT NULL,
    hash_sha256 TEXT    NOT NULL,
    size_bytes  INTEGER NOT NULL,
    modified_at INTEGER NOT NULL,
    deleted_at  INTEGER,
    UNIQUE(vault_id, path)
);

CREATE TABLE sessions (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    token_hash  TEXT    NOT NULL UNIQUE,
    device_id   TEXT    NOT NULL,
    device_name TEXT,
    vault_id    TEXT    REFERENCES vaults(vault_id),
    created_at  INTEGER NOT NULL,
    expires_at  INTEGER NOT NULL,
    last_seen   INTEGER
);

CREATE TABLE sync_log (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    vault_id    TEXT    NOT NULL REFERENCES vaults(vault_id),
    device_id   TEXT    NOT NULL,
    operation   TEXT    NOT NULL,
    file_path   TEXT    NOT NULL,
    status      TEXT    NOT NULL,
    detail      TEXT,
    created_at  INTEGER NOT NULL
);

CREATE TABLE plugin_webhooks (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    plugin_id   TEXT    NOT NULL UNIQUE,
    name        TEXT    NOT NULL,
    webhook_url TEXT    NOT NULL,
    secret      TEXT    NOT NULL,
    events      TEXT    NOT NULL,
    enabled     INTEGER NOT NULL DEFAULT 1,
    created_at  INTEGER NOT NULL
);
```

#### Volume layout

```
/vaults/
  personal/
    Notes/
    .obsidian/
  work/
    ...
```

The backend validates `vault_id` against the `vaults` table and on-disk path before any file operation.

### Security layers

1. **Bcrypt master password** — only hash in `.env` (`DEVITRI_MASTER_HASH`).
2. **JWT (HS256)** — short-lived sessions; hash stored in `sessions`.
3. **Login rate limit** — default 5 attempts/minute/IP (`DEVITRI_LOGIN_RATE_LIMIT`).
4. **`.obsidian/` isolation** — writes/deletes under `.obsidian` rejected; dashboard may read.
5. **Bulk delete protection** — thresholds via `DEVITRI_DELETE_THRESHOLD_*`; confirmation via `bulk_delete_confirmed` (batch) or `X-Bulk-Delete-Confirmed` (DELETE).
6. **Path validation** — no traversal; no `.obsidian` segments in user paths.
7. **HTTP security headers** — `nosniff`, `X-Frame-Options`, HSTS behind TLS, CSP on API responses.
8. **CORS** — `DEVITRI_CORS_ORIGINS` for browser clients only.
9. **Upload limit** — `DEVITRI_MAX_UPLOAD_BYTES` (default 50 MB).

---

## 4. REST API Reference

All `/api/*` routes require `Authorization: Bearer <TOKEN>` except:

- `POST /api/auth/login`
- `/api/setup/*` (only while not configured; guarded when configured)

Sync and vault routes are namespaced: `/api/vaults/:vault_id/...`.

### Vaults

| Method | Endpoint | Description |
|--------|----------|-------------|
| `GET` | `/api/vaults` | List vaults |
| `POST` | `/api/vaults` | Register vault |
| `GET` | `/api/vaults/:vault_id` | Vault metadata |
| `DELETE` | `/api/vaults/:vault_id` | Delete vault (explicit confirmation) |

**`GET /api/vaults`** — response 200:

```json
{
  "vaults": [
    {
      "vault_id": "personal",
      "name": "Personal",
      "description": "Personal notes",
      "path": "/vaults/personal",
      "total_files": 318,
      "total_size_bytes": 12994560,
      "last_sync": 1748476800
    }
  ]
}
```

**`POST /api/vaults`** — request / response 201:

```json
{ "vault_id": "archive", "name": "Archive", "description": "Historical notes" }
```

```json
{ "vault_id": "archive", "path": "/vaults/archive", "created_at": 1748476800 }
```

### Authentication

| Method | Endpoint | Description |
|--------|----------|-------------|
| `POST` | `/api/auth/login` | Master password → JWT (dashboard session) |
| `POST` | `/api/auth/logout` | Invalidate current session |
| `GET` | `/api/auth/session` | Validate token; device info |
| `POST` | `/api/auth/token` | Generate device token (requires master password in body) |

**`POST /api/auth/login`**

```json
{
  "password": "your-master-password",
  "device_id": "web-dashboard-uuid",
  "device_name": "Devitri Web Dashboard"
}
```

```json
{
  "token": "eyJhbGciOiJIUzI1NiIs...",
  "expires_at": 1748563200,
  "device_id": "web-dashboard-uuid"
}
```

401: `{ "error": "invalid_credentials" }`

### Sync

| Method | Endpoint | Description |
|--------|----------|-------------|
| `GET` | `/api/vaults/:vault_id/sync/manifest` | Full manifest (paths + hashes) |
| `POST` | `/api/vaults/:vault_id/sync/upload` | Upload file (headers + body) |
| `GET` | `/api/vaults/:vault_id/sync/download` | Download file (`?path=`) |
| `DELETE` | `/api/vaults/:vault_id/sync/file` | Delete file (bulk safety + optional header) |
| `POST` | `/api/vaults/:vault_id/sync/batch` | Batch negotiation |

**`GET /api/vaults/:vault_id/sync/manifest`**

```json
{
  "vault_id": "personal",
  "generated_at": 1748476800,
  "files": [
    {
      "path": "Notes/Ideas.md",
      "hash": "e3b0c44298fc1c149afb...",
      "size": 2048,
      "modified_at": 1748390400
    }
  ]
}
```

**`POST /api/vaults/:vault_id/sync/upload`**

```
Authorization: Bearer <TOKEN>
X-File-Path: Notes/Ideas.md
X-File-Hash: e3b0c44298fc1c149afb...
Content-Type: application/octet-stream

<raw bytes>
```

**`POST /api/vaults/:vault_id/sync/batch`**

```json
{
  "device_id": "iphone-15-pro",
  "bulk_delete_confirmed": false,
  "files": [
    { "path": "Notes/Ideas.md", "hash": "abc123...", "modified_at": 1748390400, "size": 1024 }
  ]
}
```

```json
{
  "to_upload": ["Notes/Ideas.md"],
  "to_download": [],
  "conflicts": [],
  "to_delete": [],
  "bulk_delete_warning": false
}
```

When deletion count exceeds thresholds and `bulk_delete_confirmed` is false, `to_delete` is empty and `bulk_delete_warning` is true.

**`DELETE /api/vaults/:vault_id/sync/file`**

```json
{ "path": "Notes/Old.md" }
```

Optional header: `X-Bulk-Delete-Confirmed: true` when rolling-window delete limits would block the operation.

403 example:

```json
{
  "error": "bulk_delete_blocked",
  "message": "...",
  "bulk_delete_warning": true
}
```

### Dashboard and statistics

| Method | Endpoint | Description |
|--------|----------|-------------|
| `GET` | `/api/stats` | Global stats |
| `GET` | `/api/vaults/:vault_id/stats` | Per-vault stats |
| `GET` | `/api/vaults/:vault_id/stats/activity` | Recent sync activity |
| `GET` | `/api/vaults/:vault_id/tree` | File tree |
| `GET` | `/api/vaults/:vault_id/file` | File content (text) |
| `GET` | `/api/settings` | Non-secret server settings |

**`GET /api/stats`**

```json
{
  "vaults": [
    { "vault_id": "personal", "total_files": 318, "total_size_bytes": 12994560, "conflicts_pending": 1, "last_sync": 1748476800 }
  ],
  "devices": [
    { "device_id": "macbook-pro", "device_name": "MacBook Pro", "last_seen": 1748476800 }
  ],
  "totals": { "files": 460, "size_bytes": 21488230, "conflicts_pending": 2, "conflicts_resolved_total": 14 }
}
```

### Plugin webhooks

| Method | Endpoint | Description |
|--------|----------|-------------|
| `GET` | `/api/plugins` | List webhooks |
| `POST` | `/api/plugins` | Register webhook |
| `DELETE` | `/api/plugins/:id` | Remove webhook |
| `PUT` | `/api/plugins/:id/toggle` | Enable/disable |

Events include: `sync.completed`, `sync.batch_completed`, `conflict.detected`, `conflict.resolved`, `auth.login`, `vault.bulk_delete_blocked`.

Signed delivery:

```
X-Devitri-Signature: sha256=<hmac_hex>
X-Devitri-Event: sync.completed
X-Devitri-Delivery: <uuid>
```

### Obsidian plugin configuration (per vault)

```json
{
  "serverUrl": "https://api.example.com",
  "vaultId": "personal",
  "deviceId": "macbook-pro",
  "deviceName": "MacBook Pro",
  "token": "eyJhbGciOiJIUzI1NiIs...",
  "syncInterval": 900
}
```

---

## 5. Design System

UI follows **Nano v1 (Zinc)**. Full specification: [`DESIGN.md`](./DESIGN.md).

Implementation: `frontend/src/lib/styles/themes/nano-zinc.css`.

---

## 6. Frontend (SvelteKit)

- **SSG** via `@sveltejs/adapter-static`; `ssr: false` for SPA behavior.
- **Miller Columns** vault browser; Obsidian-flavored markdown preview with allowlist HTML sanitization.
- **Auth** — JWT in `localStorage` for web sessions; acceptable when dashboard runs locally (profile A). See README security notes.
- **CSP** — configured in `svelte.config.js` (`kit.csp`, hash mode for inline bootstrap).
- **Setup** — polls `GET /api/setup/check` until `ready: true`.

---

## 7. Sync Protocol and Conflicts

Files are identified by **SHA-256** of content.

### 3-way sync

| Condition | Action |
|-----------|--------|
| `H(L) = H(B) = H(R)` | No-op |
| `H(L) ≠ H(B)`, `H(R) = H(B)` | Upload local |
| `H(R) ≠ H(B)`, `H(L) = H(B)` | Download remote |
| `H(L) ≠ H(B)`, `H(R) ≠ H(B)`, `H(L) ≠ H(R)` | Conflict handling |

### Resolution

1. **Auto-merge** — Markdown, non-overlapping paragraph edits.
2. **Conflict copy** — `OriginalName (Devitri Conflict - DeviceID - YYYYMMDDHHMM).md`
3. **Bulk delete guard** — see §3.
4. **Loop avoidance** — ignore `modify` events when local hash already matches remote.

---

## 8. Obsidian Plugin (TypeScript)

- HTTP only via **`requestUrl`** (not `fetch`) for mobile WebView compatibility.
- Binary uploads as `ArrayBuffer` with `X-File-Path` / `X-File-Hash`.
- Register vault listeners after `onLayoutReady`.
- **`onExternalSettingsChange`** — reload credentials when `data.json` syncs from another device.
- **`manifest.json` `id`** must equal install folder name (`devitri-obsidian-plugin`).
- Bulk delete: client-side threshold check + server confirmation; **Confirm once** in plugin settings.

See [`plugin-obsidian/README.md`](./plugin-obsidian/README.md).

---

## 9. Community Plugins (Webhooks)

External services register for events; payloads are HMAC-SHA256 signed. Core stays small; extensions run out-of-process.

---

## 10. Reference Deployment

See `deploy/caddy/`, `deploy/dev/`, and `deploy/traefik-external/` for alternate compose files. Root `docker-compose.yml` is the bundled Traefik production stack; `deploy/traefik/docker-compose.yml` symlinks to it.

Production checklist:

- TLS on public API (operator responsibility)
- `DEVITRI_CORS_ORIGINS` set for public dashboards
- Block `/api/setup/*` from the internet after first-run (proxy rules in compose examples)
- `DEVITRI_TRUST_PROXY_HEADERS=true` only behind your reverse proxy

Environment variables: [`.env.example`](./.env.example).

---

## 11. First-Run Onboarding

When `DEVITRI_MASTER_HASH` and/or `DEVITRI_JWT_SECRET` are missing:

- Protected `/api/*` returns **503** (first-run middleware).
- Only `/api/setup/*` is usable until configured.

The server **never writes** secrets to the host filesystem. The user copies generated values into `.env` and restarts.

| Method | Endpoint | Description |
|--------|----------|-------------|
| `GET` | `/api/setup/check` | `{ "ready": bool, "missing": [...] }` |
| `POST` | `/api/setup/generate` | Combined setup helper |
| `POST` | `/api/setup/generate-master-hash` | Bcrypt hash from password |
| `POST` | `/api/setup/generate-jwt-secret` | Random JWT secret |

Frontend polls `/api/setup/check` every 3 seconds after the user updates `.env`.

Security:

- Passwords are hashed in memory, not stored.
- Setup routes are rate-limited; disabled once configured (`SetupGuard`).
- Secrets are generated server-side only.

---

## 12. Repository Structure

```
devitri/
├── backend/           # Go API + sync + SQLite migrations
├── frontend/          # SvelteKit dashboard
├── plugin-obsidian/   # Obsidian plugin
├── deploy/            # Docker variants
├── docs/tasks/        # Optional specs (see README there)
├── FOUNDATION.md
├── DESIGN.md
├── AGENTS.md
└── .env.example
```

### Conventions

- **Monorepo, three packages** — HTTP only at runtime.
- **Immutable migrations** — `NNN_description.sql`.
- **CSS tokens only** in Svelte components.
- **CI** — `go test`, `npm run check`, plugin build (when workflows are present).

---

## 13. Roadmap

### Phase 1 — Core (Backend + Sync)

- [x] REST API with JWT sessions
- [x] First-run `/api/setup/*`
- [x] 3-way sync + SQLite manifest
- [x] Vault CRUD, upload/download/delete/batch
- [x] Conflict handling + bulk delete protection
- [x] Rate limiting, CORS, security headers, path validation

### Phase 2 — Dashboard

- [x] Setup onboarding flow
- [x] Miller Columns + markdown preview
- [x] Nano Zinc theme (dark/light)
- [x] Stats and device management
- [ ] PWA polish

### Phase 3 — Obsidian Plugin

- [x] TypeScript plugin with `requestUrl`
- [x] Background sync + settings UI
- [x] Mobile-safe HTTP
- [ ] Community plugin store listing

### Phase 4 — Extension ecosystem

- [ ] Webhook dispatch hardening
- [ ] Official renderer plugin (optional)
- [ ] Developer documentation for webhooks

### Phase 5 — Distribution

- [ ] Published images on `ghcr.io`
- [ ] Deploy guides (Coolify, Portainer, Unraid)

---

*This document is the source of truth for architecture and API contracts. Update it when making breaking or behavioral changes.*
