# Devitri

*Pronunciation:* **DEH-vih-tree** · `/ˈdɛvɪtri/` (from *devitrification*)

[![CI](https://github.com/rigter/devitri/actions/workflows/ci.yml/badge.svg)](https://github.com/rigter/devitri/actions/workflows/ci.yml)

**Bidirectional sync and self-hosted web dashboard for Obsidian vaults.**

Devitri keeps your notes synchronized across devices using content hashes (SHA-256) and three-way merge, while giving you a minimal web UI to browse vaults and manage device access. You run the stack; your data stays on your infrastructure.

> Architecture and API contract: [`FOUNDATION.md`](./FOUNDATION.md) · UI: [`DESIGN.md`](./DESIGN.md) · Contributors: [`AGENTS.md`](./AGENTS.md)

## Features

- **3-way sync** — local, remote, and last-synced base; automatic Markdown merge when edits do not overlap
- **Conflict copies** — safe filenames when merge is not possible (`Devitri Conflict - device - timestamp`)
- **Bulk-delete protection** — blocks large deletion batches until explicitly confirmed
- **Obsidian plugin** — desktop and mobile via `requestUrl` (no browser CORS issues)
- **Web dashboard** — static SvelteKit app (Miller columns, markdown preview, device tokens)
- **First-run setup** — generates bcrypt hash and JWT secret; secrets are never written to disk by the server
- **Security defaults** — JWT sessions, rate-limited login, path validation, upload size limits, security headers

## Repository layout

```
devitri/
├── backend/           # Go API + sync engine + SQLite
├── frontend/          # SvelteKit static dashboard
├── plugin-obsidian/   # Obsidian community plugin (TypeScript)
├── deploy/            # Docker Compose (dev, Traefik, Caddy)
├── FOUNDATION.md      # Product & API specification (source of truth)
├── DESIGN.md          # Design system (Nano v1 / Zinc)
└── AGENTS.md          # Contributor / agent context
```

## Deployment models

| Model | API | Dashboard | CORS | Typical use |
|-------|-----|-----------|------|-------------|
| **A — API only (public)** | HTTPS on VPS | Local (`npm run dev` / preview) | List localhost origins + set `VITE_DEVITRI_BACKEND_URL` | Personal server, max privacy for UI |
| **B — API + dashboard (public)** | HTTPS | HTTPS (same or other domain) | `DEVITRI_CORS_ORIGINS` = dashboard URL(s) | Family team, browser access anywhere |
| **C — Local Docker** | `localhost:8080` | `localhost:3000` | Defaults in dev compose | Hacking, integration tests |

**Obsidian plugin:** always talks to your API URL with `Authorization: Bearer`. It does **not** use browser CORS; use **HTTPS** for any internet-exposed API.

See [`.env.example`](./.env.example) for copy-paste `DEVITRI_CORS_ORIGINS` values per profile.

## Quick start (Docker, local)

```bash
git clone https://github.com/rigter/devitri.git
cd devitri
cp .env.example .env
# Complete first-run: start stack, open setup UI, paste generated secrets into .env, restart

docker compose -f deploy/dev/docker-compose.yml up -d --build
```

- API: http://localhost:8080  
- Dashboard (container): http://localhost:3000  
- Health: http://localhost:8080/health  

First-run: until `DEVITRI_MASTER_HASH` and `DEVITRI_JWT_SECRET` are set, only `/api/setup/*` is available. Use the dashboard **Setup** flow or the setup API to generate values, then restart the backend.

## Development

### Backend

```bash
cd backend
go test ./...
go run ./cmd/devitri
```

Listens on `:8080` by default. SQLite under `./data`, vault files under `./vaults` (or paths configured in Docker).

### Frontend

```bash
cd frontend
npm install
cp .env.example .env.local   # optional: VITE_DEVITRI_BACKEND_URL for remote API
npm run dev                    # http://localhost:5173
npm run check && npm run build
```

### Obsidian plugin

```bash
cd plugin-obsidian
npm install
npm run build
```

Copy `main.js` and `manifest.json` into your vault’s `.obsidian/plugins/devitri-obsidian-plugin/` (folder name must match `manifest.json` `id`). Details: [`plugin-obsidian/README.md`](./plugin-obsidian/README.md).

## Configuration

| File | Purpose |
|------|---------|
| [`.env.example`](./.env.example) | Backend secrets, CORS, sync thresholds, proxy trust |
| [`frontend/.env.example`](./frontend/.env.example) | `VITE_DEVITRI_BACKEND_URL` when API is remote |

Important variables:

- **`DEVITRI_MASTER_HASH`** / **`DEVITRI_JWT_SECRET`** — required after setup  
- **`DEVITRI_CORS_ORIGINS`** — browser dashboard origins (comma-separated)  
- **`DEVITRI_TRUST_PROXY_HEADERS`** — `true` only behind your reverse proxy  
- **`DEVITRI_ALLOW_INSECURE_JWT`** — local dev only; never in production  

Production API **must** be served over **HTTPS**. That is the operator’s responsibility (Traefik, Caddy, etc. under `deploy/`).

## Production compose

```bash
# Bundled Traefik stack (recommended) — real file at repo root
docker compose up -d --build

# Same stack via deploy path (symlink to root docker-compose.yml)
docker compose -f deploy/traefik/docker-compose.yml up -d --build

# Caddy greenfield install
docker compose -f deploy/caddy/docker-compose.yml up -d

# Existing external Traefik (backend + frontend only)
docker compose -f deploy/traefik-external/docker-compose.yml up -d --build
```

Before the bundled Traefik stack, create **files** under `deploy/traefik/` (not directories — see [`deploy/traefik/README.md`](./deploy/traefik/README.md)):

```bash
cp deploy/traefik/traefik.yml.example deploy/traefik/traefik.yml
touch deploy/traefik/acme.json && chmod 600 deploy/traefik/acme.json
```

For an existing VPS Traefik network, prefer `deploy/traefik-external/`.

## Security

- All `/api/*` routes require `Authorization: Bearer <token>` except `/api/auth/login` and `/api/setup/*` (setup only before configured).
- Dashboard tokens are stored in `localStorage` when using the web UI; prefer profile **A** if you do not want a public dashboard attack surface.
- Revoke device tokens from **Devices** in the dashboard.
- Report vulnerabilities: see [`SECURITY.md`](./SECURITY.md).

## Testing

Local verification (Docker, split dev, plugin, sync): [`TESTING.md`](./TESTING.md).

## FAQ

### Only new notes sync after I change the server URL (local → VPS)

The Obsidian plugin keeps a **local sync base** (`manifestB`) in plugin data. After the first cycles it runs **incremental** sync: it re-scans paths you create, edit, delete, or rename—not necessarily every file in the vault on every run.

If you previously synced against **another server** (for example local Docker on `localhost:8080`) and then point **Server URL** at a **new remote** VPS with an **empty** vault:

- The plugin may still believe older notes are already synchronized (base manifest from the old server).
- Only files marked **dirty** since that change (typically **new** notes) are picked up and uploaded.
- The remote API can show `files: []` while your laptop still has a full vault.

**Fix:** In Obsidian → Settings → Devitri → **Reset Local Sync State**, then **Sync Now**. That clears the local base and performs a full Markdown scan toward the current server. Confirm on the server:

```bash
curl -sS -H "Authorization: Bearer TOKEN" https://your-api.example.com/api/vaults/YOUR_VAULT/sync/manifest
ls -la /vaults/YOUR_VAULT/   # inside the backend container or host mount
```

Use the same **Vault ID** and a fresh **access key** from the dashboard **Connect** page for the server you are on now.

### What file types sync?

The plugin syncs **Markdown** notes (`.md` via Obsidian’s markdown file list). Paths under `.obsidian/` are excluded. Attachments and other extensions are not synced in the current release.

### Server manifest is empty but the dashboard lists my vault

The vault row can exist after the first API contact while **no files** have been uploaded yet. See the local → remote case above, or check the plugin developer console after **Sync Now** for errors (401, bulk-delete blocked, upload failures).

## Contributing

We welcome issues and pull requests. Read [`CONTRIBUTING.md`](./CONTRIBUTING.md) and [`AGENTS.md`](./AGENTS.md) before larger changes. API JSON shapes in `FOUNDATION.md` are contractual—update them together with `backend`, `frontend`, and `plugin-obsidian` clients.

## Author

**Rigter** — [rigter.me](https://rigter.me) · [GitHub](https://github.com/rigter)

## License

MIT — see [`LICENSE`](./LICENSE). Copyright (c) 2026 [Rigter](https://rigter.me).
