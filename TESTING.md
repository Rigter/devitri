# Testing Guide

How to run Devitri locally and verify backend, dashboard, and Obsidian plugin behavior.

See also: [README.md](./README.md) (overview), [`.env.example`](./.env.example) (configuration profiles), [FOUNDATION.md](./FOUNDATION.md) (API contract).

---

## Prerequisites

- Docker and Docker Compose
- [Go](https://go.dev/) 1.24+ (backend development)
- [Node.js](https://nodejs.org/) 20+ and [pnpm](https://pnpm.io/) (frontend; lockfile is `pnpm-lock.yaml`)
- [npm](https://www.npmjs.com/) (only if you test the [Obsidian plugin](https://github.com/rigter/devitri-obsidian-plugin) from its repo)

---

## Choose a test layout

| Layout | Backend | Dashboard | When to use |
|--------|---------|-----------|-------------|
| **Docker stack** | `localhost:8080` | `localhost:3000` | Full stack in containers (profile C) |
| **Split dev** | `localhost:8080` (Docker or `go run`) | `localhost:5173` (`pnpm run dev`) | Fast UI iteration (profile A) |
| **Traefik compose** | Via proxy | Via proxy | Production-like routing |

---

## 1. Docker stack (backend + dashboard)

From the repository root:

```bash
cp .env.example .env
mkdir -p data vaults
docker compose -f deploy/dev/docker-compose.yml up -d --build
```

| Service | URL |
|---------|-----|
| API | http://localhost:8080 |
| Health | http://localhost:8080/health |
| Dashboard (container) | http://localhost:3000 |

Logs:

```bash
docker compose -f deploy/dev/docker-compose.yml logs -f backend
docker compose -f deploy/dev/docker-compose.yml logs -f frontend
```

Stop / rebuild:

```bash
docker compose -f deploy/dev/docker-compose.yml down
docker compose -f deploy/dev/docker-compose.yml up -d --build
```

---

## 2. First-run setup

Until `DEVITRI_MASTER_HASH` and `DEVITRI_JWT_SECRET` are set in `.env`, protected `/api/*` returns **503**; only `/api/setup/*` works.

1. Open http://localhost:3000/setup (Docker) or http://localhost:5173/setup (Vite dev).
2. Generate master password hash and JWT secret in the wizard (server-side only).
3. Paste values into `.env` at the repo root.
4. Restart the backend:
   ```bash
   docker compose -f deploy/dev/docker-compose.yml restart backend
   ```
5. The UI polls `GET /api/setup/check` until `ready: true`, then redirects to login.

For local backend without Docker:

```bash
cd backend && go run ./cmd/devitri
```

Optional dev-only shortcut (never in production): `DEVITRI_ALLOW_INSECURE_JWT=true` in `.env` — see `.env.example`.

---

## 3. Split development (recommended for frontend work)

**Terminal A — API**

```bash
# Option 1: Docker backend only
docker compose -f deploy/dev/docker-compose.yml up -d backend

# Option 2: native Go
cd backend
go run ./cmd/devitri
```

**Terminal B — dashboard**

```bash
cd frontend
pnpm install
cp .env.example .env.local
# If API is remote or on 8080 while Vite is on 5173:
# echo 'VITE_DEVITRI_BACKEND_URL=http://localhost:8080' >> .env.local
pnpm run dev
```

Open http://localhost:5173. CORS defaults allow `http://localhost:5173` when `DEVITRI_CORS_ORIGINS` is unset.

Checks:

```bash
cd frontend && pnpm run check && pnpm run build
cd backend && go test ./...
```

---

## 4. Dashboard smoke test

1. **Setup** — complete first-run if needed.
2. **Login** — http://localhost:3000/login (or `:5173`) with master password.
3. **Home** — vault list and stats load.
4. **Vault** — open a vault; Miller columns and markdown preview work.
5. **Connect** — generate a device access key (master password required); copy token for the plugin.
6. **Devices** — list sessions; revoke a test device if needed.

API sanity (with token from login):

```bash
curl -s http://localhost:8080/health
curl -s -H "Authorization: Bearer YOUR_TOKEN" http://localhost:8080/api/vaults
```

---

## 5. Obsidian plugin

The plugin is maintained in [devitri-obsidian-plugin](https://github.com/rigter/devitri-obsidian-plugin). Clone that repo and follow its README for build and install steps.

Outputs `main.js` and uses `manifest.json` (`id`: `devitri-obsidian-plugin`).

### Install

Folder name **must** match manifest `id`:

```
<VAULT>/.obsidian/plugins/devitri-obsidian-plugin/
  main.js
  manifest.json
```

### Configure

In Obsidian → Settings → Devitri:

| Field | Local dev value |
|-------|-----------------|
| Server URL | `http://localhost:8080` (API, not the dashboard port) |
| Vault ID | e.g. `personal` (slug you use on the server) |
| Access key | Token from dashboard **Connect** (not the master password) |

Use **Verify** / **Sync Now** in plugin settings. Bulk deletes above thresholds require **Confirm once** in settings, then sync again.

---

## 6. Sync end-to-end

1. Create or edit `.md` files in the Obsidian vault.
2. Trigger sync in the plugin (or wait for the interval).
3. Confirm files under `vaults/<vault_id>/` on the host (Docker volume) or via dashboard vault browser.
4. Edit a file in the dashboard (read) and in Obsidian; verify conflict copy or merge per [FOUNDATION.md](./FOUNDATION.md) §7.

---

## 7. Production-like compose (Traefik)

```bash
docker compose up -d
```

Requires Traefik network and labels configured in root `docker-compose.yml` (or the symlink at `deploy/traefik/docker-compose.yml`). Use your real domain and HTTPS for plugin testing; the plugin should use the **API** base URL (`https://api.example.com`), not the dashboard origin unless they are the same host with `/api` routed.

---

## Troubleshooting

| Issue | What to check |
|-------|----------------|
| `permission denied` on `data/` or `vaults/` | `mkdir -p data vaults && chmod 755 data vaults` |
| Dashboard cannot reach API | API on `:8080`; set `VITE_DEVITRI_BACKEND_URL`; check `DEVITRI_CORS_ORIGINS` for custom ports |
| `503` on `/api/*` | First-run: finish setup and restart backend with `.env` filled |
| Plugin “connection failed” | Server URL = API base; token from **Connect**; HTTPS in production |
| CORS errors in browser | Add exact origin (e.g. `http://localhost:5173`) to `DEVITRI_CORS_ORIGINS` |
| Plugin not listed | Folder name = `devitri-obsidian-plugin`; enable community plugins |

**Logs (Docker dev stack)**

```bash
docker compose -f deploy/dev/docker-compose.yml logs backend
docker compose -f deploy/dev/docker-compose.yml logs frontend
docker compose -f deploy/dev/docker-compose.yml ps
```

---

## Automated tests

Run locally before opening a PR:

```bash
cd backend && go test ./...
cd frontend && pnpm install && pnpm exec svelte-kit sync && pnpm run check && pnpm run build
```

Plugin checks: see [devitri-obsidian-plugin](https://github.com/rigter/devitri-obsidian-plugin).

### Continuous integration

GitHub Actions workflow [`.github/workflows/ci.yml`](./.github/workflows/ci.yml) runs on every push to `main`/`master` and on pull requests:

| Job | Command |
|-----|---------|
| Backend | `go test ./...` (CGO enabled for SQLite) |
| Frontend | `pnpm run check` + `pnpm run build` |

Ad-hoc Go scripts in `backend/` root use `//go:build ignore` and are excluded from `go test ./...`.

---

*Last updated: 2026-06-07*
