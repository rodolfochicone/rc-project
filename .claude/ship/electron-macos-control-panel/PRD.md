# PRD — rc Electron macOS Control Panel

Product requirements for a native macOS desktop "control panel" that wraps the
existing rc daemon web UI, owns the daemon lifecycle, and adds config/workspace/
extension surfaces. This is the **why & what**; technical design lives in SPEC.md.
Source of truth for stack/conventions: `STACK.md` (same folder).

## Problem

rc's daemon, REST/SSE API, and React web UI already exist, but using rc as a
day-to-day desktop tool has friction:

- The web UI is only reachable after the user **manually** runs `rc daemon start`
  and opens a browser at the right `127.0.0.1:<port>`. There is no turnkey app.
- The daemon lifecycle (start / health / crash-restart / graceful stop) is the
  user's responsibility; crashes and orphan processes / stuck locks are easy to hit.
- **Config is CLI-only**: global `~/.rc/config.toml` and per-workspace
  `.rc/config.toml` cannot be read or written over HTTP, so there is no GUI to
  edit them.
- There is **no UI** for workspace management (register / unregister / rename) or
  for viewing installed extensions and reusable agents.

The result: rc is powerful but not approachable as a desktop product, and key
configuration/management tasks require the terminal.

## Goals

- One-double-click macOS `.app` that brings up rc (daemon + UI) with no terminal.
- The desktop app **owns** the daemon lifecycle robustly (start/attach, health,
  restart-on-crash, graceful stop) without orphans or stuck locks.
- Make config, workspace management, and extensions/agents viewable and (for
  config + workspaces) editable from a GUI — reusing the existing web UI, not
  rewriting it.
- Keep rc **single-binary / local-first**: Electron is a thin shell; the daemon
  and the artifacts under `~/.rc` / `.rc` remain the source of truth.

## Target users

- **rc power users / engineers on macOS** who run AI-assisted development
  workflows locally and want a persistent, always-available control surface
  instead of a terminal + browser tab.
- **Operators / maintainers** who need to inspect and adjust rc configuration,
  manage workspaces, and audit which extensions and reusable agents are active.

Out of audience: Windows/Linux desktop users (this release is macOS-only),
multi-user / remote-daemon scenarios.

## Functional requirements

### FR1 — Electron shell wraps the existing web UI

- FR1.1 The app loads the daemon-served web UI from `http://127.0.0.1:<port>`,
  reading the port/socket from `~/.rc/daemon/daemon.json` (never hardcoded).
- FR1.2 The `BrowserWindow` and its requests satisfy the daemon's existing
  Host / Origin / CSRF validation (`internal/api/httpapi/browser_middleware.go`)
  so all API and SSE calls succeed.
- FR1.3 Native macOS integration: main window, native menu bar, a tray /
  menu-bar item showing daemon status, dock re-open, and basic app shortcuts.
- FR1.4 No rewrite of `web/` or `packages/ui`: the renderer hosts no React of
  its own; it displays the daemon UI.

### FR2 — Daemon lifecycle owned by the app

- FR2.1 On launch, if no daemon is running, the app spawns the `rc` binary
  (`rc daemon start`) and waits until it is healthy.
- FR2.2 If a daemon is already running, the app **attaches** to it (respects
  `daemon.lock` / `daemon.json`) rather than starting a duplicate.
- FR2.3 The app polls `GET /api/daemon/health` and `GET /api/daemon/status`
  and surfaces daemon state (starting / healthy / unhealthy / stopped) in the UI
  and tray.
- FR2.4 On daemon crash (process exit while app is running), the app restarts it
  (with a bounded retry policy, not an infinite tight loop).
- FR2.5 On app quit, the daemon is stopped gracefully via `POST /api/daemon/stop`
  **only if the app started it**; an attached pre-existing daemon is left running.
  Quitting leaves no orphan processes and no stuck `daemon.lock`. (Default policy;
  see Open question Q4.)

### FR3 — Config HTTP endpoints (new Go) + UI

- FR3.1 New daemon HTTP handlers **read** global config (`~/.rc/config.toml`) and
  per-workspace config (`<root>/.rc/config.toml`) following the schema in
  `internal/core/workspace/config_types.go` (`Defaults`, `Tasks`/`tasks.run`,
  `FixReviews`, `FetchReviews`, `WatchReviews`, `Exec`, `Runs`, `Sound`).
- FR3.2 New daemon HTTP handlers **write** global and per-workspace config.
  Writes are **atomic** (temp file + rename, per `internal/daemon/info.go`),
  validated against the existing config validation before persisting, and the
  daemon reflects the new values after the write.
- FR3.3 The new endpoints conform to `internal/api` conventions: registered in
  `internal/api/contract/routes.go` and `internal/api/core/routes.go`, correct
  timeout class (read = `TimeoutRead`, write = `TimeoutMutate`), `X-rc-Workspace-ID`
  header for workspace-scoped routes, and the `code`/`message`/`request_id`
  error envelope (`core.NewProblem` / `core.RespondError`).
- FR3.4 `openapi/rc-daemon.json` is updated for the new routes and TS types are
  regenerated (`bun run codegen`); `bun run codegen-check` is diff-clean.
- FR3.5 New **shared web UI pages** (usable in both browser and Electron),
  following the existing TanStack Router + `systems/<feature>` + `openapi-fetch`
  - TanStack Query + `@rodolfochicone/ui` patterns:
  * Config editing — global and per-workspace.
  * Workspace management — register / unregister / rename, via the existing
    workspace endpoints.
  * Extensions / reusable agents — read-only view at minimum.

### FR4 — Real-time updates

- FR4.1 The app follows runs live by reusing the existing SSE endpoint
  `GET /api/runs/:run_id/stream`; event kinds from `pkg/rc/events/event.go`
  (run / job / session / tool_call / usage / task / review) update the UI live.
- FR4.2 Cursor-based reconnection (`Last-Event-ID`) works inside the Electron
  context after transient disconnects.

### FR5 — Packaging & distribution

- FR5.1 The app builds a macOS `.app` via electron-builder (or equivalent),
  **universal** (arm64 + x64).
- FR5.2 The app integrates as a bun/turbo workspace at `apps/desktop/` using the
  repo toolchain (bun, TypeScript `^6`, oxlint/oxfmt, vitest); root
  `package.json` workspaces and `turbo.json` are updated accordingly.
- FR5.3 Code signing and notarization steps are documented and the build is
  reproducible from a documented command.

## Non-functional requirements

- **NFR1 — Conformance over preference.** All new Go follows CLAUDE.md +
  golangci-lint (zero-issue): `%w`-wrapped errors, `log/slog`, `context.Context`
  first arg, no `panic`/`log.Fatal` in production paths, every goroutine with
  explicit ownership + `ctx.Done()` shutdown, small interfaces, functional
  options, compile-time interface checks. Frontend follows oxlint/oxfmt (4-space),
  strict TS, and reuses `@rodolfochicone/ui` (base-ui) — no shadcn/radix, no
  `web/` rewrite.
- **NFR2 — Local-first / single-binary.** Electron is a shell only; no new
  sidecars or external control planes. Source of truth stays the daemon +
  `~/.rc` / `.rc`. No hardcoded ports — always read `daemon.json`.
- **NFR3 — Dependencies via tooling.** No hand-edited `go.mod` (use `go get`);
  reuse `pelletier/go-toml/v2 v2.3.0` for TOML (no new TOML lib); JS deps via bun.
- **NFR4 — Atomicity & safety.** Config writes are atomic and never corrupt the
  file on partial failure. No destructive git commands without explicit user
  permission.
- **NFR5 — Robust lifecycle.** Daemon restart uses bounded retries (no infinite
  tight loop, no `time.Sleep`-based orchestration in Go paths); app quit cleans
  up its owned process and lock.
- **NFR6 — Verification is a blocking gate.** `make verify`
  (fmt → lint zero-issues → `test -race` → build) green; frontend
  `bun run frontend:typecheck` + `bun run frontend:test` + `bun run codegen-check`
  green; the `.app` builds reproducibly.

## Out of scope

- Rewriting or restyling the existing `web/` UI or `packages/ui` components.
- Windows and Linux desktop builds (macOS only this release).
- Remote / multi-machine daemons, multi-user auth, or exposing the daemon beyond
  `127.0.0.1`.
- Editing extensions or reusable agents from the UI (extensions/agents view is
  **read-only** in this release; install/enable/disable is out of scope).
- New event kinds, new run/task/review semantics, or changes to the SSE protocol.
- Auto-update / Sparkle, App Store distribution, telemetry/analytics.
- Replacing or extending the TOML config schema beyond what
  `config_types.go` already defines.

## Acceptance criteria (verifiable by real execution)

Each criterion is checkable by running the app, the daemon, the test suites, or
inspecting files — not by inspection of intent alone.

- **AC1 (FR1, FR2.1, FR2.2):** Launching the built `.app` on macOS, with no
  daemon running and without the user invoking `rc daemon start`, results in the
  rc web UI rendered in the app window within the daemon's startup timeout; a
  second launch (or launch while a daemon is already running) attaches to the
  existing daemon and starts no duplicate process (verify a single `rc daemon`
  process via `ps`, single owner of `daemon.lock`).
- **AC2 (FR1.2, FR4):** From the app, a workflow / exec / review can be launched,
  and its events (job / session / tool_call / usage) update **live** in the UI
  via SSE; no Origin/Host/CSRF rejections appear in daemon logs for the app's
  requests. Killing and restoring connectivity resumes the stream via
  `Last-Event-ID` without losing events.
- **AC3 (FR2.4, FR2.5):** Killing the daemon process while the app runs triggers
  an automatic restart (bounded) and the UI recovers. Quitting the app stops the
  app-started daemon gracefully (`POST /api/daemon/stop`) with **no orphan
  `rc daemon` process** (`ps` shows none) and **no stuck `daemon.lock`**
  (file absent or re-acquirable). An attached pre-existing daemon survives app quit.
- **AC4 (FR3.1, FR3.2):** Global and per-workspace config can be read and saved
  from the app; after saving, the corresponding `config.toml` on disk reflects
  the change, the file remains valid TOML parseable by the existing loader, and a
  subsequent read returns the saved values. An invalid edit is rejected with the
  standard error envelope and leaves the file unchanged.
- **AC5 (FR3.5):** A workspace can be registered, renamed, and unregistered from
  the UI, with the change reflected by the existing workspace endpoints /
  `~/.rc` store. Extensions and reusable agents are listed (read-only) in the UI.
- **AC6 (FR3.3, NFR1, NFR6 — Go):** `make verify` passes 100% (fmt, zero-issue
  lint, `test -race`, build). New routes appear in
  `internal/api/contract/routes.go` and pass `openapi_contract_test.go` parity.
- **AC7 (FR3.4, NFR6 — frontend):** `bun run frontend:typecheck`,
  `bun run frontend:test`, and `bun run codegen-check` are all green;
  `openapi/rc-daemon.json` and generated TS types are updated and diff-clean.
- **AC8 (FR5):** The universal macOS `.app` (arm64 + x64) builds from a single
  documented command; signing and notarization steps are documented and the
  build is reproducible. The Electron workspace passes its own oxlint/oxfmt +
  vitest checks.

## Open questions

These are genuine ambiguities for the user to resolve before/while building; the
listed default is what the build will assume if not corrected.

- **Q1 — Daemon binary discovery.** How does the Electron app locate the `rc`
  binary? Options: bundled inside the `.app` Resources, found on `PATH`, or a
  dev-build path. _Default assumption:_ bundle the `rc` binary in the `.app` for
  production, with a configurable override (env var / setting) for dev. (STACK
  risk #3.)
- **Q2 — Origin/CSRF exact behavior from Electron.** The precise `Origin` the
  `BrowserWindow` sends (and CSRF cookie handling) when loading remote
  `http://127.0.0.1:<port>` vs. a `file://` preload must be confirmed against
  `originAllowed` / `security_headers.go` during SPEC. _Default assumption:_ load
  the UI directly from `http://127.0.0.1:<port>` so Origin is a permitted
  localhost origin and CSRF cookies behave as in a browser. (STACK risk #2.)
- **Q3 — TOML comment fidelity.** `go-toml/v2` does not round-trip arbitrary
  comments. REQUEST says "preserve comments/structure where possible". _Default
  assumption:_ write the typed `ProjectConfig` structure atomically (full
  re-marshal); arbitrary user comments are **not** guaranteed to survive. Confirm
  this fidelity level is acceptable. (STACK risk #5.)
- **Q4 — Quit policy default.** REQUEST allows "stop gracefully OR keep daemon,
  per preference." _Default assumption:_ graceful stop of the app-started daemon
  on quit (no orphans/locks); a pre-existing attached daemon is left running.
  Confirm whether a user-facing toggle is wanted. (STACK risk #4.)
- **Q5 — Electron / electron-builder versions.** Greenfield/unpinned; latest
  stable will be chosen at SPEC time and pinned in `apps/desktop/package.json`.
  Confirm whether a specific Electron major is mandated. (STACK risk #1.)
- **Q6 — CI / verify coverage for the Electron build.** Should `make verify` /
  the turbo pipeline include building the `.app` (slow, macOS-only), or should
  packaging stay a separate documented step? _Default assumption:_ Electron unit
  checks (oxlint/oxfmt/vitest) run in the normal frontend pipeline; the `.app`
  package build stays a separate documented command. (STACK risk #6.)
- **Q7 — Extensions/agents data source.** The read-only extensions / reusable
  agents view needs a daemon data source. If no HTTP endpoint currently exposes
  installed extensions / reusable agents, a new read endpoint may be required
  (in scope per FR3.5) — confirm the intended source
  (`internal/core/extension`, `internal/core/agents`) during SPEC.
