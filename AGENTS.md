# Devitri — Agent Context

**Devitri** is a bidirectional sync and self-hosted web dashboard for Obsidian. This monorepo contains two packages (`backend`, `frontend`) that communicate with the [Obsidian plugin](https://github.com/rigter/devitri-obsidian-plugin) only over HTTP.

---

## Documentation

| Document | Purpose |
|----------|---------|
| [`FOUNDATION.md`](./FOUNDATION.md) | Product architecture, API contract, sync protocol (source of truth) |
| [`DESIGN.md`](./DESIGN.md) | Nano v1 (Zinc) design system |
| [`README.md`](./README.md) | User-facing overview and deployment |
| [`.env.example`](./.env.example) | Environment variables and deployment profiles |
| [`docs/tasks/`](./docs/tasks/) | Optional feature specs and implementation notes |

When changing API shapes, update `FOUNDATION.md` and all clients (`frontend/src/lib/api/client.ts`, and the plugin API wrapper in [devitri-obsidian-plugin](https://github.com/rigter/devitri-obsidian-plugin)).

---

## Tech Stack

| Layer | Tech |
|-------|------|
| Backend | Go 1.24+, SQLite (in `/data`), JWT `HS256`, Bcrypt cost 14 |
| Frontend | SvelteKit (SSG), CSS tokens only |
| Obsidian plugin | TypeScript, `requestUrl` (not `fetch`), esbuild — [separate repo](https://github.com/rigter/devitri-obsidian-plugin) |
| Infra | Docker, Traefik or Caddy reverse proxy |

---

## Monorepo Structure

```
devitri/
├── backend/              # Go — REST API + sync
├── frontend/             # SvelteKit static dashboard
├── deploy/               # docker-compose variants (dev, traefik, caddy)
├── FOUNDATION.md
├── DESIGN.md
└── .env.example
```

Obsidian plugin: [github.com/rigter/devitri-obsidian-plugin](https://github.com/rigter/devitri-obsidian-plugin)

---

## Key Conventions

1. **Independent packages** — backend and frontend in this repo; plugin in its own repo; HTTP REST only at runtime.
2. **Migrations** — new numbered file `NNN_description.sql` per schema change; never edit applied migrations.
3. **CSS tokens** — frontend uses `nano-zinc.css` custom properties only; no hardcoded hex in `.svelte` files.
4. **Plugin `manifest.json`** — `id` must match the plugin folder name exactly (`devitri-obsidian-plugin`).
5. **API auth** — `Authorization: Bearer <TOKEN>` on `/api/*` except `/api/auth/login` and `/api/setup/*` (setup only before configured).
6. **JSON contract** — field names and types in `FOUNDATION.md` §4 are stable across backend, frontend, and plugin.

---

## Critical Architecture

### First-run

Missing `DEVITRI_MASTER_HASH` or `DEVITRI_JWT_SECRET` → first-run mode: `/api/*` returns 503 except `/api/setup/*`. Server generates secrets; user copies into `.env` and restarts.

### Sync

- SHA-256 content identity; 3-way sync **L** (local), **B** (base), **R** (remote).
- Markdown merge when edits do not overlap; conflict copies otherwise.
- Bulk delete: blocked above count/percent thresholds until `bulk_delete_confirmed` (batch) or `X-Bulk-Delete-Confirmed` (single DELETE).

### Security

- Bcrypt master password in `.env`; JWT sessions in DB (`sessions.token_hash`).
- Rate-limited login; `.obsidian/` write/delete blocked; path validation on all file ops.
- `DEVITRI_CORS_ORIGINS` for browser dashboard only (plugin uses `requestUrl`, not browser CORS).

### Volumes

- `/vaults/{vault_id}` — note files
- `/data` — SQLite and server state

---

## Workflow for Agents

### Before any change

1. Read `FOUNDATION.md` for the affected area.
2. If API responses change, update frontend client and the plugin API wrapper in [devitri-obsidian-plugin](https://github.com/rigter/devitri-obsidian-plugin).
3. New schema → new migration file only.
4. Plugin manifest `id` unchanged unless folder is renamed too.

### Commands

```bash
cd backend && go test ./...
cd frontend && npm run check && npm run build
docker compose -f deploy/dev/docker-compose.yml up -d
```

Plugin build: see [devitri-obsidian-plugin](https://github.com/rigter/devitri-obsidian-plugin).

### CI

GitHub Actions: [`.github/workflows/ci.yml`](./.github/workflows/ci.yml) — backend `go test ./...`, frontend `pnpm run check` + build.

---

## Environment Variables

See [`.env.example`](./.env.example) for full list and deployment profiles (API-only public, API + public dashboard, local Docker).

---

## Notes for Agents

- Prefer Go stdlib; avoid new backend dependencies unless necessary.
- Never generate secrets in the browser; use `/api/setup/*` on the server.
- Plugin: **only** `requestUrl` for HTTP (mobile WebView CORS).
- Ask before renaming API fields, endpoints, or migration files.

---

*Last updated: 2026-06-07*
