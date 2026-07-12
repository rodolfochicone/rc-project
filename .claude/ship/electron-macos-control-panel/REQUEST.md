# RC Electron macOS Control Panel

## Objective

Build a native macOS desktop app (Electron) that acts as a "RC control panel".
It **wraps the existing React web UI** served by the RC daemon (no frontend
rewrite), **manages the daemon lifecycle** (spawn / health / restart / stop),
and adds **new Go HTTP endpoints** for reading and writing config, plus **new
shared web UI pages** for config editing, workspace management, and a read-only
view of extensions / reusable agents. Live updates reuse the daemon's existing
SSE stream.

## Context (current codebase state)

- **HTTP daemon** (Gin) lives in `internal/api/httpapi/server.go`; routes are
  declared in `internal/api/core/routes.go` and `internal/api/contract/routes.go`.
  It binds to `127.0.0.1` with Host/Origin (localhost) validation + CSRF. The
  port/socket are published to `~/.rc/daemon/daemon.json`.
- **REST + SSE** API already covers: runs (list / detail / snapshot / events /
  `stream` SSE / cancel), tasks (list / detail / board / items / spec / memory /
  start-run / validate / archive), reviews (fetch / watch / rounds / issues /
  start-run), workspaces (register / update / delete / resolve / sync), exec, sync.
- **Real time**: `GET /api/runs/{id}/stream` (SSE) with `Last-Event-ID`,
  heartbeat, overflow, and cursor support. Event types live in
  `pkg/rc/events/event.go` (~25 kinds: run / job / session / tool_call / usage /
  task / review). The durable journal is `~/.rc/runs/<run-id>/events.jsonl`.
- **Web UI**: React 19 + Vite 8 + TanStack Router/Query in `web/` (pkg `rc-web`,
  ~70-75% complete): dashboard, workflows, task board, runs with live stream,
  reviews, workspace selector. Embedded in the binary via `web/embed.go`
  (embed.FS); dev mode via the `--web-dev-proxy` flag. TS types are generated
  from `openapi/rc-daemon.json`.
- **Current gaps**: config (`~/.rc/config.toml` + workspace `.rc/config.toml`)
  is CLI-only (no HTTP surface); no UI for workspace management, extensions, or
  reusable agents; no desktop app to drive the daemon.
- **Stack**: Go (main module) verified via `make verify`. Frontend is a
  bun + turbo monorepo with workspaces `packages/ui` (`@escaletech/ui`, React 19
  - Tailwind 4 + base-ui), `web` (`rc-web`), and `sdk/*`. Frontend checks run via
    `bun run frontend:*` and `bun run codegen` / `codegen-check`.

## Requirements

### 1. Electron shell (new app, e.g. `apps/desktop/`)

- Load the daemon-served web UI at `http://127.0.0.1:<port>`, reading the
  port/socket from `~/.rc/daemon/daemon.json` (never hardcode).
- Satisfy the daemon's Origin/Host/CSRF validation (configure the
  `BrowserWindow` / requests so they pass the existing localhost checks).
- Main window + native macOS menu, tray / menu-bar item showing daemon status,
  deep-link / shortcuts, re-open from the dock.
- Package a macOS `.app` (electron-builder or equivalent) with universal
  architecture (arm64 + x64). Document code signing and notarization steps.

### 2. Daemon lifecycle (owned by Electron)

- Spawn the `rc daemon start` binary (discover / point to the binary), poll
  `GET /api/daemon/health` and `/status`, restart on crash, and stop gracefully
  (`POST /api/daemon/stop`) on app quit.
- Handle an already-running daemon: **attach** rather than duplicate — respect
  `daemon.lock`.

### 3. Config endpoints (new Go HTTP) + UI

- Implement daemon HTTP handlers to **read and write** global config
  (`~/.rc/config.toml`) and per-workspace config (`.rc/config.toml`), following
  the schema in `internal/core/workspace/config_types.go` (defaults, `tasks.run`,
  `fix_reviews`, `fetch_reviews`, `watch_reviews`, `exec`, `runs`, `sound`).
- Update `openapi/rc-daemon.json` and regenerate TS types (`bun run codegen`).
- Add shared web UI pages (benefiting both browser and Electron): config editing
  (global / workspace), workspace management (register / unregister / rename),
  and an extensions / reusable agents view (read-only at minimum).

### 4. Real time

- Reuse the existing SSE (`/api/runs/{id}/stream`) to follow runs live; ensure
  cursor-based reconnection works inside the Electron context.

## Acceptance criteria

- [ ] The `.app` opens on macOS, auto-starts/attaches the daemon, and shows the
      web UI without the user running `rc daemon start` manually.
- [ ] A workflow / exec / review can be launched from the app, and its events
      (job / session / tool_call / usage) update live via SSE.
- [ ] On app close, the daemon stops gracefully (or is kept, per preference)
      with no orphan processes and no stuck locks.
- [ ] Global and workspace config can be read and saved from the app; the
      `config.toml` file reflects changes and the daemon reloads.
- [ ] Workspaces can be registered / removed / renamed from the UI.
- [ ] New Go endpoints pass `make verify` (fmt + lint zero-issues + test -race + build); frontend passes `bun run frontend:typecheck` and
      `bun run frontend:test`.
- [ ] OpenAPI is updated and TS types regenerated (`bun run codegen-check` green).
- [ ] App build is documented (signing / notarization) and reproducible.

## Constraints

- **Do not rewrite** the existing web UI; reuse `web/` and `packages/ui`.
- New daemon endpoints follow `internal/api` conventions: per-class timeouts,
  `X-RC-Workspace-ID` header, error format with `code` / `message` /
  `request_id`, localhost/CSRF validation.
- **Go conventions (per CLAUDE.md)**: wrap errors with `%w`; use `log/slog`;
  pass `context.Context` as the first arg across runtime boundaries; no `panic`
  / `log.Fatal` in production paths; every goroutine has explicit ownership and
  `ctx.Done()`-based shutdown; small interfaces (accept interfaces, return
  structs); functional options for complex constructors; compile-time interface
  checks. Activate `golang-pro` before writing Go and `testing-anti-patterns`
  before writing/modifying tests.
- **Do not** add Go deps by hand — use `go get`. Follow the CLAUDE.md skill
  dispatch protocol for every domain touched.
- Config writes must be **atomic** and preserve TOML comments/structure where
  possible — evaluate the TOML library already used in the project before
  adopting a new one.
- Keep everything **single-binary / local-first**: Electron is a shell; the
  source of truth stays the daemon and the artifacts under `~/.rc` / `.rc`.
- **Verification is a blocking gate**: `make verify` must pass at 100% (fmt,
  zero-issue lint, `-race` tests, build) and frontend `bun run frontend:*` /
  `codegen-check` must be green before any task is considered complete. Activate
  `rc-final-verify` before claiming done.
- **Git safety**: never run destructive git commands (`restore`, `reset`,
  `checkout`, `clean`, `rm`) without explicit user permission.
