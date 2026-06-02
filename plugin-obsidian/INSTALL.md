# Install and run

## Prerequisites

- Obsidian 1.0+
- Node.js 18+
- Devitri API reachable from your device (see [../TESTING.md](../TESTING.md))

## Build

```bash
cd plugin-obsidian
npm install
npm run build
```

Output: `dist/main.js` and `dist/manifest.json`.

## Install in a vault

Plugins live at `{vault}/.obsidian/plugins/devitri-obsidian-plugin/`. The folder name **must** match `manifest.json` → `id`.

```bash
VAULT="$HOME/Documents/MyVault"   # your vault path

mkdir -p "$VAULT/.obsidian/plugins/devitri-obsidian-plugin"
cp dist/main.js dist/manifest.json \
  "$VAULT/.obsidian/plugins/devitri-obsidian-plugin/"
```

Restart Obsidian → **Settings** → **Community plugins** → enable **Devitri**.

## Configure

| Setting | Example |
|---------|---------|
| Server URL | `http://localhost:8080` (local API) or `https://api.yourdomain.com` |
| Vault ID | `personal` |
| Access key | From dashboard **Connect** (paste in the access key field) |

Then **Connect & Sync** or **Verify**, and optionally **Sync Now**.

- Default sync interval: **900** seconds (15 min). Use `0` for manual only.
- In production, point **Server URL** at the **API** host, not the static dashboard URL.

## Documentation in this folder

| File | Purpose |
|------|---------|
| [`README.md`](README.md) | Overview and quick reference |
| [`STABILITY.md`](STABILITY.md) | Safety behaviour and troubleshooting |
| [`PROJECT.md`](PROJECT.md) | Repo layout for contributors |

## Troubleshooting

| Issue | What to check |
|-------|----------------|
| Plugin not listed | Folder name = `devitri-obsidian-plugin`; both `main.js` and `manifest.json` present |
| Connection fails | API up, URL without trailing slash, HTTPS in production |
| `localhost:3000` fails | Use API on `:8080`, not the dashboard dev server |
| Sync blocked (bulk delete) | **Confirm once** in settings, then **Sync Now** |
| Session expired | New access key from dashboard **Connect** |

## Platforms

macOS, Windows, Linux, iOS, and Android (via `requestUrl`).
