# Caddy deployment

Greenfield stack: **Caddy** terminates HTTP/HTTPS and routes:

| Path | Service |
|------|---------|
| `/api/*` | Go backend (`:8080`) |
| `/health` | Go backend |
| everything else | Static frontend (Caddy in container, `:80`) |

Same-host routing avoids the Traefik `405` issue (API POST hitting the static dashboard).

## Quick start (HTTP only)

From repository root:

```bash
cp .env.example .env
# fill secrets; optional: CADDY_EMAIL= for future HTTPS

docker compose -f deploy/caddy/docker-compose.yml up -d --build
```

Default `Caddyfile` listens on **`:80`** only (no TLS). Suitable for LAN or behind another proxy.

## Production (HTTPS + domain)

1. Set in `.env`:
   - `CADDY_EMAIL` — Let's Encrypt contact
   - `DEVITRI_DOMAIN` — e.g. `devitri.example.com`
   - `DEVITRI_CORS_ORIGINS=https://devitri.example.com`
   - `DEVITRI_TRUST_PROXY_HEADERS=true`
2. Replace Caddyfile:
   ```bash
   cp deploy/caddy/Caddyfile.production.example deploy/caddy/Caddyfile
   ```
3. Rebuild and start:
   ```bash
   docker compose -f deploy/caddy/docker-compose.yml up -d --build
   ```

If API and dashboard share one domain, leave `DEVITRI_BACKEND_URL` unset; the frontend uses the browser origin. For a separate API hostname, set `DEVITRI_BACKEND_URL` and rebuild the frontend.

## Volumes

Backend data paths (aligned with the Docker entrypoint):

- `../../data` → `/data` (SQLite)
- `../../vaults` → `/vaults` (note files)

`caddy_data/` and `caddy_config/` are local TLS/state — do not commit.

## Existing reverse proxy on the VPS

Use [`deploy/traefik-external/`](../traefik-external/) instead of this stack.
