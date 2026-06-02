# Contributing to Devitri

Thank you for considering a contribution. Devitri is a monorepo with three independent packages that communicate only over HTTP.

## Before you start

1. Read [`FOUNDATION.md`](./FOUNDATION.md) for the API contract and sync rules.
2. Read [`DESIGN.md`](./DESIGN.md) for UI tokens and layout rules.
3. Read [`AGENTS.md`](./AGENTS.md) for monorepo conventions.
4. Copy [`.env.example`](./.env.example) to `.env` for local backend runs.
5. Optional specs live under [`docs/tasks/`](./docs/tasks/).

## Development checks

Match what CI runs (see [`.github/workflows/ci.yml`](./.github/workflows/ci.yml)):

```bash
cd backend && go test ./...
cd frontend && pnpm install && pnpm exec svelte-kit sync && pnpm run check && pnpm run build
cd plugin-obsidian && npm ci && npm run check && npm run build
```

Pull requests should pass all three jobs before merge.

Before your first commit (or when changing `.gitignore`), run:

```bash
./scripts/validate-gitignore.sh
```

## Pull requests

- Keep changes focused; one concern per PR when possible.
- Do not edit applied SQLite migrations; add a new numbered file instead.
- If you change API JSON fields or endpoints, update `FOUNDATION.md` and all clients (`frontend/src/lib/api/client.ts`, `plugin-obsidian/src/sync/api.ts`).
- Use [Conventional Commits](https://www.conventionalcommits.org/) for commit messages (English).
- Do not commit `.env`, vault data, or `data/` / `vaults/` contents.

## Code style

- **Go:** stdlib-first; `gofmt` / usual Go idioms.
- **Frontend:** design tokens from `nano-zinc.css` only—no hardcoded hex in `.svelte` files.
- **Plugin:** HTTP via Obsidian `requestUrl` only (not raw `fetch`).

## Questions

Open a GitHub issue for design questions or a draft PR for early feedback.
