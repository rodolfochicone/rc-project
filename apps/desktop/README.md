# rc Desktop

Electron shell for the rc macOS control panel. Thin lifecycle owner that starts (or attaches to) the rc daemon and loads the daemon-served UI in a `BrowserWindow`.

## Stack

- Electron 42.x (latest stable)
- electron-builder 26.x (universal macOS builds)
- TypeScript 6, vitest, oxlint/oxfmt

## Development

```sh
# From repo root — install all workspace deps
bun install

# TypeCheck
bun run --filter @rc/desktop typecheck

# Unit tests
bun run --filter @rc/desktop test

# Build JS output
bun run --filter @rc/desktop build
```

## Packaging (universal macOS .app)

```sh
# Build the rc binary first
make build

# Package the .app + dmg (arm64 + x64 universal)
bun run --filter @rc/desktop package
```

Output lands in `apps/desktop/dist-packaged/`.

## Code signing

Set these environment variables before packaging:

| Variable           | Description                                            |
| ------------------ | ------------------------------------------------------ |
| `CSC_LINK`         | Base64-encoded `.p12` certificate, or path to the file |
| `CSC_KEY_PASSWORD` | Password for the `.p12` certificate                    |

electron-builder picks these up automatically.

## Notarization

After signing, notarize with Apple's `notarytool`:

```sh
xcrun notarytool submit \
  "dist-packaged/rc-<version>-universal.dmg" \
  --apple-id "$APPLE_ID" \
  --password "$APPLE_APP_SPECIFIC_PASSWORD" \
  --team-id "$APPLE_TEAM_ID" \
  --wait

xcrun stapler staple "dist-packaged/rc.app"
```

Required env:

| Variable                      | Description                            |
| ----------------------------- | -------------------------------------- |
| `APPLE_ID`                    | Apple ID email used for notarization   |
| `APPLE_APP_SPECIFIC_PASSWORD` | App-specific password for the Apple ID |
| `APPLE_TEAM_ID`               | Developer Team ID (10-char string)     |

## Smoke checklist (AC1/AC2/AC3)

Run these manually after `bun run --filter @rc/desktop package`:

**AC1 — Launch / attach, no duplicate daemon**

- [ ] Launch the built `.app` with no daemon running → rc UI renders within startup timeout (30s)
- [ ] Launch again while first instance is running → second window focuses, no new `rc daemon` process
- [ ] Verify via `ps aux | grep "rc daemon"` → exactly one `rc daemon` process

**AC2 — Live SSE, no Origin/Host/CSRF rejection**

- [ ] Start a workflow run, confirm live status events appear in the UI
- [ ] Check daemon logs → no `403` or `412` responses from Origin/CSRF middleware
- [ ] Kill network (lo0 down) and restore → SSE stream resumes via `Last-Event-ID` without manual reload

**AC3 — Bounded restart, graceful quit**

- [ ] Kill `rc daemon` externally → app auto-restarts within bounded backoff, UI recovers
- [ ] Quit the app (Cmd-Q) → daemon stops gracefully; `ps aux | grep "rc daemon"` shows no orphan; `~/.rc/daemon/daemon.lock` is absent or re-acquirable
- [ ] With a pre-existing daemon (not started by the app), quit → daemon survives quit
