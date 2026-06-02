# Devitri Obsidian Plugin

Bidirectional sync client for Obsidian vaults, using the [Devitri](https://github.com/rigter/devitri) self-hosted API.

**Author:** [Rigter](https://rigter.me)

## Features

- Bidirectional sync with SHA-256 change detection
- 3-way conflict resolution (automatic Markdown merge + conflict copies)
- Bulk-delete protection with explicit **Confirm once** in settings
- Mobile-safe HTTP via Obsidian `requestUrl` only (no `fetch`)
- Device access keys from the web dashboard (**Connect**)

## Prerequisites

- Obsidian 1.0+ (desktop or mobile)
- A running Devitri API (HTTPS recommended on the public internet)

## Quick install

```bash
cd plugin-obsidian
npm install
npm run build
cp dist/main.js dist/manifest.json \
  "$VAULT/.obsidian/plugins/devitri-obsidian-plugin/"
```

The plugin folder name must be **`devitri-obsidian-plugin`** (same as `id` in `manifest.json`).

Full steps: [`INSTALL.md`](INSTALL.md). Stability and security: [`STABILITY.md`](STABILITY.md).

## Configure in Obsidian

| Field | Value |
|-------|--------|
| **Server URL** | API base URL only, e.g. `https://api.example.com` or `http://localhost:8080` (not the dashboard port) |
| **Vault ID** | Slug configured on the server, e.g. `personal` |
| **Access key** | Generate in dashboard → **Connect**, paste once, then **Connect & Sync** |

Default automatic sync interval: **900 seconds (15 minutes)**. Set `0` for manual sync only.

## Bulk delete

If a sync would delete more than 20 files or more than 5% of the vault, sync is blocked until you tap **Confirm once** in plugin settings and run **Sync Now** again.

## Development

```bash
npm run check   # TypeScript
npm run build   # → dist/main.js + dist/manifest.json
```

Monorepo docs: [`../TESTING.md`](../TESTING.md), [`../FOUNDATION.md`](../FOUNDATION.md).

## License

MIT — see [`../LICENSE`](../LICENSE).
