# Security policy

## Supported versions

Security fixes are applied to the default branch (`main`). Self-hosted deployments should track `main` or recent release tags when published.

## Reporting a vulnerability

**Please do not open a public GitHub issue for security bugs.**

Contact the maintainer via [GitHub Security Advisories](https://github.com/rigter/devitri/security/advisories/new) or [rigter.me](https://rigter.me) with:

- Description and impact
- Steps to reproduce
- Affected component (backend, frontend, plugin)

We will acknowledge receipt and aim to respond within a reasonable timeframe.

## Operator responsibilities

Devitri is self-hosted. Deployers are responsible for:

- **HTTPS** on any internet-exposed API
- Strong master password and protecting `.env` (`DEVITRI_MASTER_HASH`, `DEVITRI_JWT_SECRET`)
- **`DEVITRI_CORS_ORIGINS`** restricted to trusted dashboard origins
- **`DEVITRI_TRUST_PROXY_HEADERS=true`** only behind a reverse proxy they control
- Revoking compromised device tokens via the dashboard

## Known design notes

- Web dashboard sessions use JWT in `localStorage` (acceptable when the UI is not publicly hosted; see README deployment profile A).
- Obsidian plugin uses long-lived device tokens stored in plugin data; revoke via **Devices** if a device is lost.
