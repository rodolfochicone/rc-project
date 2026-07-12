# SPEC — RC Electron macOS Control Panel

Technical design satisfying `PRD.md`, conforming to `STACK.md` (same folder).
This is the **how**. Scope: a new Electron shell at `apps/desktop/`, new Go
daemon HTTP endpoints for config + extensions/agents reads, and new shared web
UI pages — all reusing existing daemon, security, config, lifecycle, and SSE
machinery. No `web/` or `packages/ui` rewrite.

Verified against source (file:line cited inline). Where the PRD listed an Open
Question, the resolved decision is stated under **Decisions** with its evidence.

---

## 0. Decisions resolving PRD open questions

- **Q2 / Origin & CSRF (resolved by reading `internal/api/httpapi/browser_middleware.go`).**
  The Electron `BrowserWindow` loads the UI **directly from
  `http://127.0.0.1:<port>`** (the daemon-served UI), never `file://` for the
  app surface. Consequences proven by the middleware:
  - `originAllowed` (l.307–316) accepts only scheme `http` + a loopback host
    whose port equals `s.Port()`. A window at `http://127.0.0.1:<port>` sends
    `Origin: http://127.0.0.1:<port>` → **allowed**. A `file://` origin would be
    rejected (scheme not `http`). Therefore the renderer must be the remote
    localhost document, not a local file.
  - `hostValidationMiddleware` (l.26–41) requires `Host` ∈ {`127.0.0.1`,
    `localhost`, bind host} with matching port — satisfied automatically by the
    remote load.
  - CSRF (l.159–237) sets a `Secure; SameSite=Strict` cookie and, for mutating
    browser requests, requires the `X-RC-CSRF` header to equal the cookie. The
    existing `web/` `openapi-fetch` client already echoes this; **the new pages
    reuse that same client**, so CSRF is handled by the UI, not by Electron. The
    cookie is `Secure` — over plain `http://127.0.0.1` Chromium **does** treat
    loopback as a secure context, so the cookie is stored. This must be smoke-
    tested (AC2); if a future Chromium drops it, fall back to reading the
    `X-RC-CSRF` response header (already returned on every response, l.173).
  - **No custom `Origin`/header injection in Electron.** Do not override
    `Origin` via `webRequest` — the natural remote-load Origin already passes.
- **Q1 / binary discovery.** Resolution order in the Electron main process:
  1. `RC_BINARY` env var (absolute path) — dev override;
  2. bundled binary at `process.resourcesPath/bin/rc` (production `.app`);
  3. `rc` on `PATH` (fallback).
     First existing + executable wins. The packaged `.app` bundles the `rc` binary
     built by `make build` via electron-builder `extraResources`.
- **Q3 / TOML fidelity.** Config writes **re-marshal the typed `ProjectConfig`**
  with `pelletier/go-toml/v2 v2.3.0` (the project's only TOML lib, per STACK).
  Arbitrary user comments are **not** preserved; structure/key-ordering follows
  go-toml's marshaller. This is acceptable per PRD Q3 default. Writes are atomic
  (temp+rename, mirroring `daemon.WriteInfo`, `internal/daemon/info.go` l.77–123).
- **Q4 / quit policy.** Default: graceful stop (`POST /api/daemon/stop`) of the
  **app-started** daemon on quit; an **attached** pre-existing daemon is left
  running. Ownership is tracked by an in-memory flag set only when the app
  itself spawned the process. No user toggle this release.
- **Q5 / Electron versions.** Pin **Electron `38.x` (latest stable line)** and
  **electron-builder `25.x`** in `apps/desktop/package.json`; exact patch
  resolved by `bun install` at build time and committed in `bun.lock`. Universal
  `.app` via electron-builder `mac.target: { target: dmg/zip, arch: universal }`.
- **Q6 / CI coverage.** Electron unit checks (oxlint/oxfmt/vitest/tsc) run in the
  normal frontend pipeline (turbo). The `.app` **package** build is a separate
  documented command (`bun run --filter @rc/desktop package`), not part of
  `make verify`.
- **Q7 / extensions & agents data source.** No HTTP endpoint currently exposes
  installed extensions or reusable agents (verified: only CLI/runtime callers of
  `extension.Discovery.Discover` and `agents.Registry.Discover`). Two new
  **read-only** endpoints are added (see §3.3), backed by
  `internal/core/extension` (`Discovery.Discover`) and `internal/core/agents`
  (`Registry.Discover`).

---

## 1. Component overview

```
apps/desktop/              NEW Electron shell (TS, bun/turbo workspace)
  src/main/                main process: lifecycle owner, daemon supervisor, tray, menu
  src/preload/             minimal preload (contextIsolation) — status bridge only
  (no renderer React)      window loads http://127.0.0.1:<port> served by daemon

internal/api/core/         + new config + catalog handlers, interfaces, routes
internal/api/contract/     + new RouteSpec entries + response types
internal/core/workspace/   + WriteConfig (atomic typed marshal) ; reuse Load*/validate
internal/core/configsvc/   NEW thin service: read/write global+workspace config
internal/core/catalog/     NEW thin service: list extensions + reusable agents (read-only)

openapi/rc-daemon.json     + new paths/schemas ; regenerate web/src/generated
web/src/routes/_app/       + config.tsx, workspaces.tsx, extensions.tsx (TanStack Router)
web/src/systems/config/    NEW feature module (Query hooks + base-ui forms)
web/src/systems/workspaces/ NEW feature module
web/src/systems/extensions/ NEW feature module (read-only)
```

No new sidecars, no new TOML/HTTP libs, no new state libs. Electron is shell +
supervisor only.

---

## 2. Electron shell (`apps/desktop/`)

### 2.1 Process model

- `contextIsolation: true`, `nodeIntegration: false`, `sandbox: true` for the
  renderer that hosts the daemon UI. The renderer is the **remote** document; it
  has no Node access. Preload exposes only a read-only `window.rcdesktop`
  channel for daemon status (used by an optional native status strip, not the UI).
- Single instance lock (`app.requestSingleInstanceLock()`); second launch focuses
  the existing window (complements the daemon's own flock singleton).

### 2.2 Daemon supervisor (`src/main/daemon.ts`)

Owns FR2. State machine: `starting → healthy → unhealthy → stopped`.

Data source = `~/.rc/daemon/daemon.json` (never hardcoded port). Path computed in
TS as `path.join(os.homedir(), '.rc', 'daemon', 'daemon.json')` — mirrors
`internal/config/home.go` `InfoPath` (l.85). Schema mirrors `daemon.Info`
(`internal/daemon/info.go` l.29–36): `{ pid, http_port, socket_path, state,
started_at, version }`.

Algorithm on launch:

1. Read `daemon.json`. If present, `Validate`-equivalent (pid>0, port 0–65535)
   and probe `GET http://127.0.0.1:<http_port>/api/daemon/health`.
2. If healthy → **attach** (`ownsDaemon = false`). Do not spawn. (FR2.2, AC1).
   The OS-level flock in `internal/daemon/lock.go` guarantees a spawn would fail
   anyway; we avoid it proactively.
3. If absent/unhealthy → spawn the resolved `rc` binary as `rc daemon start`
   (Q1 discovery), `ownsDaemon = true`. Poll `/api/daemon/health` every 250ms up
   to the daemon's startup window (`defaultDaemonStartupTimeout = 10s`,
   `internal/cli/daemon_commands.go` l.31) ×3 = 30s ceiling, then surface a
   start-failure UI state. No infinite loop.
4. On healthy, read the (possibly updated, ephemeral-port) `daemon.json` again
   and load the window at `http://127.0.0.1:<http_port>`.

Health/status polling: `GET /api/daemon/health` + `GET /api/daemon/status` every
3s while running; result drives tray icon + window state (FR2.3). Uses
`AbortController` with the `probe` 2s budget.

Crash handling (FR2.4): subscribe to the spawned `child_process` `exit`. If
`ownsDaemon` and exit is unexpected, restart with **bounded exponential backoff**
(max 5 attempts, 1s→2s→4s→8s→16s caps). After the cap, stop and show an
unhealthy/error state; never tight-loop. Backoff is timer-based in JS (not Go),
satisfying NFR5 (no `time.Sleep` orchestration in Go paths).

Quit (FR2.5, Q4, AC3): on `before-quit`, if `ownsDaemon`, call
`POST /api/daemon/stop` (mutate, 30s budget) and await child exit; remove our
reference. If `!ownsDaemon`, do nothing (leave attached daemon running). The
daemon itself removes `daemon.lock`/`daemon.json` on graceful stop, so quitting
the app leaves no orphan process and no stuck lock (verified path:
`internal/daemon` lock+info lifecycle).

### 2.3 Native integration (FR1.3)

- `BrowserWindow` main window; macOS native menu (`Menu.setApplicationMenu`).
- Tray / menu-bar item (`Tray`) reflecting `starting/healthy/unhealthy/stopped`;
  menu: Show, Restart daemon, Quit.
- Dock re-open (`app.on('activate')` recreates/show window).
- Standard shortcuts (Cmd-R reload, Cmd-Q quit through the supervisor).

### 2.4 Interfaces / contracts (TS)

```ts
interface DaemonInfo {
  pid: number;
  http_port: number;
  socket_path?: string;
  state: "starting" | "ready" | "stopped";
  started_at: string;
  version?: string;
}
type DaemonUiState = "starting" | "healthy" | "unhealthy" | "stopped";
interface Supervisor {
  start(): Promise<DaemonInfo>; // attach-or-spawn, resolves when healthy
  state(): DaemonUiState;
  onState(cb: (s: DaemonUiState) => void): () => void;
  stop(): Promise<void>; // graceful, only if owned
}
```

Binary resolver and `daemon.json` reader are injected (interfaces) so they are
unit-testable under vitest with temp dirs and a fake `child_process`.

### 2.5 Packaging (FR5)

- `electron-builder` config in `apps/desktop/electron-builder.yml`:
  `mac.arch: universal`, `extraResources: [{ from: ../../bin/rc, to: bin/rc }]`,
  hardened runtime, `entitlements` for the spawned child.
- Signing/notarization documented in `apps/desktop/README.md`:
  `CSC_LINK`/`CSC_KEY_PASSWORD` for signing, `notarytool` via
  `APPLE_ID`/`APPLE_APP_SPECIFIC_PASSWORD`/`APPLE_TEAM_ID`. Build command:
  `bun run --filter @rc/desktop package`.
- Workspace wiring (FR5.2): add `"apps/*"` to root `package.json` `workspaces`;
  add `apps/desktop` tasks to `turbo.json` (`build`, `lint`, `test`, `typecheck`,
  `package`). oxlint/oxfmt configs inherited from repo root; `tsconfig` extends
  `tsconfig.base.json`.

---

## 3. Go daemon changes

All new code conforms to CLAUDE.md/golangci-lint (NFR1): `%w` wraps, `log/slog`,
`context.Context` first arg, no `panic`/`log.Fatal`, small interfaces (accept
interface, return struct), functional options where a constructor grows,
compile-time `var _ Iface = (*T)(nil)`. No hand-edited `go.mod`.

### 3.1 New service interfaces (`internal/api/core/interfaces.go`)

Add two interfaces and wire them onto `Handlers`/`HandlerConfig` (l.16–32, l.34–53)
exactly like the existing services:

```go
// ConfigService reads and writes RC global and per-workspace config.
type ConfigService interface {
    GetGlobal(ctx context.Context) (ConfigDocument, error)
    PutGlobal(ctx context.Context, doc ConfigDocument) (ConfigDocument, error)
    GetWorkspace(ctx context.Context, workspaceID string) (ConfigDocument, error)
    PutWorkspace(ctx context.Context, workspaceID string, doc ConfigDocument) (ConfigDocument, error)
}

// CatalogService lists installed extensions and reusable agents (read-only).
type CatalogService interface {
    Extensions(ctx context.Context, workspaceID string) (ExtensionList, error)
    Agents(ctx context.Context, workspaceID string) (AgentList, error)
}
```

`Handlers` gains `Config ConfigService` and `Catalog CatalogService` fields (and
`HandlerConfig` mirrors, plus `Clone()` propagation, l.104–120).

### 3.2 Config endpoints (FR3.1–3.3)

New handlers in `internal/api/core/handlers.go`, registered in
`internal/api/core/routes.go` (l.11) and inventoried in
`internal/api/contract/routes.go` (l.16):

| Method | Path                    | Response                 | Timeout         | Workspace hdr |
| ------ | ----------------------- | ------------------------ | --------------- | ------------- |
| GET    | `/api/config/global`    | `ConfigDocumentResponse` | `TimeoutRead`   | no            |
| PUT    | `/api/config/global`    | `ConfigDocumentResponse` | `TimeoutMutate` | no            |
| GET    | `/api/config/workspace` | `ConfigDocumentResponse` | `TimeoutRead`   | **yes**       |
| PUT    | `/api/config/workspace` | `ConfigDocumentResponse` | `TimeoutMutate` | **yes**       |

Workspace-scoped routes require `X-RC-Workspace-ID`. **Add their paths to
`requiresActiveWorkspace`** in `internal/api/httpapi/browser_middleware.go`
(l.247–267) so a missing header yields the existing
`412 workspace_context_missing`. Handlers read the ID via
`core.ActiveWorkspaceIDFromContext` (set by the middleware at l.118) and resolve
the workspace root through the existing `WorkspaceService.Get`.

Handler shape mirrors existing handlers (e.g. `UpdateWorkspace`, l.508–533):
nil-service guard → `bindJSON` → validate → call service → respond with the
contract type. Errors flow through `h.respondError` → `core.RespondError` →
`*Problem` envelope (`internal/api/core/errors.go`), giving the standard
`code`/`message`/`request_id` shape (FR3.3). Validation failures return
`400 config_invalid`; the file is **not** written.

### 3.3 Catalog endpoints (FR3.5 extensions/agents read-only, Q7)

| Method | Path                      | Response                | Timeout       | Workspace hdr |
| ------ | ------------------------- | ----------------------- | ------------- | ------------- |
| GET    | `/api/catalog/extensions` | `ExtensionListResponse` | `TimeoutRead` | **yes**       |
| GET    | `/api/catalog/agents`     | `AgentListResponse`     | `TimeoutRead` | **yes**       |

Workspace-scoped (extensions/agents resolve workspace + global scopes). Add both
paths to `requiresActiveWorkspace`.

### 3.4 ConfigService implementation (`internal/core/configsvc/`)

Thin adapter over `internal/core/workspace`:

- `GetGlobal` → `workspace.LoadGlobalConfig(ctx)` (`config.go` l.191–206) →
  map `ProjectConfig` to `ConfigDocument`.
- `GetWorkspace` → resolve root from workspace store, then
  `workspace.LoadConfig(ctx, root)` (l.183–189).
- `PutGlobal`/`PutWorkspace` → **new `workspace.WriteConfig(ctx, path, cfg)`**
  (§3.5): validate via the existing `cfg.validate(scope)` (already called inside
  loaders, `config.go` l.301), then atomic write. After write, re-load to return
  the persisted document (proves the daemon reflects new values, FR3.2/AC4).

`ConfigDocument` is the JSON-facing shape. To avoid re-deriving the schema we
serialize the **typed `ProjectConfig`** as JSON (all fields pointer-optional, so
JSON `null`/absent = "unset"). The OpenAPI schema is the JSON projection of
`ProjectConfig` (`config_types.go` l.14–97): `defaults`, `tasks`, `fix_reviews`,
`fetch_reviews`, `watch_reviews`, `exec`, `runs`, `sound`. Round-trip: JSON →
`ProjectConfig` → `WriteConfig` (TOML) and TOML → `ProjectConfig` → JSON.

### 3.5 `workspace.WriteConfig` (FR3.2, NFR4, Q3)

New exported function in `internal/core/workspace/`. Contract:

```go
func WriteConfig(ctx context.Context, configPath string, cfg ProjectConfig) error
```

Steps: `cfg.validate(scope)` first (reject before touching disk); marshal with
`toml.Marshal` (go-toml/v2 v2.3.0); write atomically via temp file + `fsync` +
rename + dir `fsync` — **identical pattern to `daemon.WriteInfo`**
(`internal/daemon/info.go` l.95–123), perms `0o600`. On any failure before
rename, the original file is untouched (NFR4/AC4). Scope (`globalConfigScope` vs
`workspaceConfigScope`) chosen by caller so validation matches the target.

### 3.6 CatalogService implementation (`internal/core/catalog/`)

- `Extensions` → build `extension.Discovery` for the workspace root and call
  `Discover(ctx)` (`internal/core/extension/discovery.go` l.92). Project
  `DiscoveryResult.Extensions` (effective, `DiscoveredExtension` l.… with `Ref`,
  `Manifest`, `Enabled`, `Source`) to `ExtensionList`. Read-only: no enable/
  disable (out of scope). Include non-fatal `Failures`/`Overrides` counts as
  metadata, not errors.
- `Agents` → `agents.New().Discover(ctx, root)`
  (`internal/core/agents/agents.go` l.207) → project `Catalog.Agents`
  (`ResolvedAgent`: name, scope/source, description) to `AgentList`; surface
  `Catalog.Problems` as per-item warnings.

### 3.7 Wiring

Both services are constructed where the daemon assembles `HandlerConfig` (the
daemon transport bootstrap that builds `core.NewHandlers`). `RegisterRoutes`
guards on nil handlers already; new handlers guard on nil service like every
existing handler (`serviceUnavailableProblem`). Compile-time checks added:
`var _ core.ConfigService = (*configsvc.Service)(nil)` etc.

### 3.8 Contract/OpenAPI (FR3.4, AC6/AC7)

- Add the six new `RouteSpec` entries to `RouteInventory`
  (`internal/api/contract/routes.go`) — `openapi_contract_test.go` enforces
  parity, so this is mandatory for `make verify`.
- Add response Go types (`ConfigDocumentResponse`, `ExtensionListResponse`,
  `AgentListResponse`, and element types) to `internal/api/contract` (`types.go`)
  following existing `*Response` naming.
- Update `openapi/rc-daemon.json` with the six paths + schemas; run
  `bun run codegen`; `bun run codegen-check` must be diff-clean.

---

## 4. Web UI (shared, browser + Electron) (FR3.5)

Reuse existing patterns verbatim (STACK §web): TanStack Router file-based routes,
`systems/<feature>` modules, `openapi-fetch` typed against generated OpenAPI,
TanStack Query for server state, zustand only if local state is needed, zod for
form schemas, `@escaletech/ui` (base-ui) components. **No shadcn/radix, no new
styling system, no `web/` rewrite.** 4-space indent, explicit `: ReactElement`
return types, oxlint/oxfmt clean.

New routes under `web/src/routes/_app/` (alongside `workflows.tsx`, `runs.tsx`):

- `config.tsx` — tabbed global vs per-workspace config editor.
- `workspaces.tsx` — register / rename / unregister (reuses existing
  `POST/PATCH/DELETE /api/workspaces` — no new Go).
- `extensions.tsx` — read-only extensions + reusable agents lists.

New feature modules:

- `web/src/systems/config/` — Query hooks (`useGlobalConfig`,
  `useWorkspaceConfig`, `useSaveConfig`) over the new endpoints; a base-ui form
  driven by a zod schema mirroring `ProjectConfig`'s optional fields; on save,
  show the standard error envelope message on `config_invalid`; invalidate the
  query on success so the UI reflects persisted values.
- `web/src/systems/workspaces/` — list/register/rename/delete using existing
  workspace endpoints + `X-RC-Workspace-ID` plumbing already present in the
  client.
- `web/src/systems/extensions/` — read-only tables from
  `/api/catalog/extensions` and `/api/catalog/agents`.

Generated route tree is regenerated via `web/scripts/tsr-generate.mjs` (not
hand-edited). All API calls go through the existing `openapi-fetch` client, which
already carries the CSRF header + `X-RC-Workspace-ID`, satisfying FR1.2 with no
Electron-specific code.

## 5. Real-time updates (FR4)

No protocol or event-kind changes (out of scope). The new pages and the wrapped
UI consume the existing `GET /api/runs/:run_id/stream` SSE with `Last-Event-ID`
reconnection exactly as `web/` does today. Because the Electron window is the
remote document, EventSource/fetch-stream behaves identically to a browser;
`Last-Event-ID` resume works unchanged (AC2). Verification is a smoke test, not
new code.

---

## 6. Trade-offs & key decisions

- **Remote-load over `file://` shell.** Chosen so Origin/Host/CSRF pass with zero
  Electron-side header hacking and zero UI duplication (FR1.4). Cost: the window
  is blank until the daemon is healthy — handled by a native "starting" splash in
  the main process before the window navigates.
- **Typed re-marshal over comment-preserving TOML.** Conforms to the single TOML
  lib (NFR3); accepts comment loss (Q3). Avoids adding a CST/round-trip TOML dep.
- **Catalog endpoints reuse discovery, stay read-only.** Smallest surface that
  satisfies FR3.5 without pulling enable/disable lifecycle (out of scope).
- **Supervision in JS, not Go.** Keeps the daemon unchanged and single-binary;
  backoff timers live in the shell, honoring NFR5's "no `time.Sleep` in Go
  orchestration".
- **`.app` build outside `make verify`.** Packaging is slow + macOS-only (Q6);
  unit checks still gate every change.

## 7. Test approach (NFR6, AC6/AC7/AC8)

- **Go (`make verify`, `-race`).** Table-driven `t.Run`, `t.Parallel()`,
  `t.TempDir()`:
  - `workspace.WriteConfig`: atomicity (no partial file on marshal/rename
    failure via injected failing writer), round-trip equality
    (`Load(Write(cfg)) == cfg` for representative configs), validation rejects
    bad input and leaves the file unchanged (AC4).
  - `configsvc`/`catalog`: mock `WorkspaceService`/discovery via interfaces;
    assert mapping and error propagation. No production test-only methods.
  - Handlers: extend `handlers_test.go`/`handlers_error_paths_test.go` with the
    new routes — success, `config_invalid` 400, missing-workspace 412,
    service-unavailable. `openapi_contract_test.go` proves inventory↔OpenAPI
    parity (AC6).
- **Frontend (`bun run frontend:typecheck` / `frontend:test` / `codegen-check`).**
  Vitest + Testing Library + MSW mocking the new endpoints: config load/save/
  invalid-error rendering; workspace register/rename/unregister; extensions/
  agents read-only render. `codegen-check` diff-clean (AC7).
- **Electron (`apps/desktop`, vitest + oxlint/oxfmt + tsc).** Unit-test the
  supervisor with a fake `child_process` and temp `daemon.json`: attach-vs-spawn
  decision (AC1), bounded-backoff restart and give-up (AC3), graceful-stop-only-
  if-owned (AC3), binary resolution order (Q1). The `.app` build + a manual smoke
  checklist (launch with/without running daemon, live SSE, kill+recover, quit
  leaves no orphan/lock) cover AC1–AC3/AC8 by real execution.

## 8. File-touch summary (feeds TASKS.md)

- New Go: `internal/core/configsvc/*.go`, `internal/core/catalog/*.go`,
  `internal/core/workspace/config_write.go` (+tests).
- Edited Go: `internal/api/core/interfaces.go`, `handlers.go`, `routes.go`;
  `internal/api/contract/routes.go`, `types.go`;
  `internal/api/httpapi/browser_middleware.go` (`requiresActiveWorkspace`);
  daemon transport bootstrap (wire `Config`/`Catalog`).
- Spec/codegen: `openapi/rc-daemon.json`, regenerated
  `web/src/generated/rc-openapi.d.ts`.
- New web: `web/src/routes/_app/{config,workspaces,extensions}.tsx`,
  `web/src/systems/{config,workspaces,extensions}/*` (+tests).
- New app: `apps/desktop/**` (main/preload/electron-builder/tests/README).
- Edited root: `package.json` (`workspaces: apps/*`), `turbo.json`, Makefile
  frontend targets if needed.

## 9. Risks carried into build

- **CSRF cookie persistence over loopback `http`** in the bundled Chromium must
  be smoke-verified (Decision Q2); header fallback is the contingency.
- **electron-builder universal + notarization** on the available macOS toolchain
  is greenfield (Q5) — verify the signing/notarytool flow early.
- **`DiscoveredExtension`/`ResolvedAgent` exported field surface** for the
  catalog projection should be confirmed against the structs at build time
  (fields cited from `discovery.go`/`agents.go`); adjust the response schema to
  the actual exported fields.
