# Bundled Traefik stack

`docker-compose.yml` in this folder is a **symlink** to the repository root [`docker-compose.yml`](../../docker-compose.yml).

## Required files (must be files, not directories)

Docker bind-mounts expect **files**. If they are missing when you first run `docker compose up`, Docker may create **empty directories** with these names and Traefik will fail to start.

From this directory:

```bash
cp traefik.yml.example traefik.yml
touch acme.json
chmod 600 acme.json
```

- **`traefik.yml`** — static Traefik config (start from `traefik.yml.example`).
- **`acme.json`** — Let's Encrypt storage (empty at first; never commit; see root `.gitignore`).

Then from the **repository root**:

```bash
docker compose up -d --build
```

For an **existing** Traefik on your VPS, use [`deploy/traefik-external/`](../traefik-external/) instead.
