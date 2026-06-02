# Devitri Obsidian Plugin (package)

TypeScript plugin for Obsidian — syncs a vault with a self-hosted [Devitri](../README.md) API.

## Documentation

| File | Description |
|------|-------------|
| [`README.md`](README.md) | User-facing overview |
| [`INSTALL.md`](INSTALL.md) | Build and install steps |
| [`STABILITY.md`](STABILITY.md) | Sync safety and auth behaviour |

## Layout

```
plugin-obsidian/
├── src/
│   ├── main.ts
│   ├── index.ts
│   ├── types/index.ts
│   ├── sync/
│   │   ├── api.ts          # requestUrl HTTP client
│   │   ├── engine.ts       # Scan → Negotiate → Pull → Push
│   │   ├── manifest.ts
│   │   ├── conflicts.ts
│   │   └── notify.ts
│   └── ui/SettingsTab.ts
├── manifest.json           # id must match install folder name
├── esbuild.config.mjs
└── dist/                   # build output (gitignored)
```

## Features (implementation)

- SHA-256 manifests and 3-way decisions
- Bulk-delete guard + **Confirm once** UI
- `onExternalSettingsChange` reloads credentials when `data.json` syncs
- Default sync interval **900 s** (`syncInterval` in plugin data)

## Development

```bash
npm run check
npm run build
```

CI runs the same checks from the monorepo root (see [`.github/workflows/ci.yml`](../.github/workflows/ci.yml)).

## License

MIT — [Rigter](https://rigter.me)
