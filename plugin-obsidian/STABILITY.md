# Stability and security

Behavioural guarantees of the Obsidian plugin. Installation: [`INSTALL.md`](INSTALL.md).

## Conflict handling

- **3-way sync** compares local (L), base (B), and remote (R) using SHA-256 hashes.
- Non-overlapping Markdown edits: automatic paragraph-level merge when possible.
- Overlapping edits or binary files: conflict copy  
  `Name (Devitri Conflict - deviceId - YYYYMMDDHHMM).ext`

## Sync loop prevention

After a download, the engine avoids treating the write as a new local edit when `H(local) == H(remote)`, so Obsidian `modify` events do not cause upload loops.

## Bulk-delete protection

Client and server enforce thresholds (defaults: more than 20 files or more than 5% of the vault):

1. Sync is blocked and a notice is shown.
2. Open **Settings → Devitri → Bulk delete** and tap **Confirm once**.
3. Run **Sync Now** again.

Server-side deletes may also require the `X-Bulk-Delete-Confirmed` header (handled by the plugin after confirmation).

## Authentication

- Connect with a **device access key** from the web dashboard (**Connect**), not the master password.
- The access key is stored in Obsidian plugin data as a JWT for API calls.
- On **401**, generate a new key in the dashboard and reconnect; there is no silent password re-login.

## Network (mobile)

All HTTP uses Obsidian **`requestUrl`**. Do not use `fetch` or `XMLHttpRequest` (CORS issues on iOS WebView).

## Filesystem rules

- Paths under `.obsidian/` are never synced.
- Server rejects `.obsidian` segments in user paths.

## Performance

- Dirty/deleted paths tracked in `Set`s for incremental scans.
- Full vault scan mainly on first sync or after **Reset Local Sync State**.

## Errors

- Per-file failures are logged; the rest of the cycle continues when possible.
- Critical connection errors surface in settings and notices.

## Operator checklist

- Use **HTTPS** for any API exposed on the internet.
- Revoke lost devices in dashboard **Devices**.
- Prefer API-only exposure (dashboard can stay local); see [../README.md](../README.md) deployment profiles.

## Verify

1. Obsidian developer console: `Devitri: Plugin loaded with data`
2. Settings show **Connected**
3. **Sync Now** updates files under `vaults/{vault_id}/` on the server
