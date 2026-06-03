# External Traefik deployment

Deploy only Devitri **backend** and **frontend** on an existing Traefik Docker network.

## Commands

From repository root:

```bash
cp .env.example .env
# Set DEVITRI_API_HOST, DEVITRI_WEB_HOST, DEVITRI_BACKEND_URL, secrets, CORS

docker compose -f deploy/traefik-external/docker-compose.yml up -d --build
```

After changing the backend image or entrypoint, force a rebuild:

```bash
docker compose -f deploy/traefik-external/docker-compose.yml build --no-cache backend
docker compose -f deploy/traefik-external/docker-compose.yml up -d backend
```

## 405 on `/api/setup/generate`

The backend only accepts **POST**. A **405** on setup usually means Traefik sent the request to the **frontend** (static Caddy), not the Go API.

**Same hostname for UI and API** (e.g. `devitri.rigter.space`): the backend router must include `PathPrefix(\`/api\`)` and a higher **priority** than the frontend router. See labels in `docker-compose.yml`.

**Quick check:**

```bash
curl -sS -X POST https://devitri.example.com/api/setup/generate \
  -H 'Content-Type: application/json' \
  -d '{"password":"testpassword12"}'
```

Expect JSON with `hash` and `jwt_secret`, not HTML or 405.

## Database error on first start

If logs show `unable to open database file` on `/data`:

1. Confirm volumes are mounted (`devitri-data:/data`, not a single volume on two paths).
2. Do **not** set `user:` on the backend service (entrypoint must chown volumes as root).
3. Rebuild backend with `--no-cache` (see above).
4. One-off fix on a running stack:

   ```bash
   docker compose -f deploy/traefik-external/docker-compose.yml exec -u root backend \
     sh -c 'mkdir -p /data /vaults && chown -R devitri:devitri /data /vaults'
   docker compose -f deploy/traefik-external/docker-compose.yml restart backend
   ```

5. Check logs for `[devitri-entrypoint]` lines — if missing, the old image is still running.
