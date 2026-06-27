# TASKS — rc Electron macOS Control Panel

Ordered, concrete build steps derived from `SPEC.md` (read `PRD.md` + `STACK.md`
for context). Each task lists the files it touches and a **done-condition** that
is verifiable by real execution, with the PRD acceptance criterion it satisfies.

Ordering rule: the Go daemon (config write → services → handlers → contract →
OpenAPI) lands first because the web UI and Electron shell both depend on the
generated OpenAPI types and live endpoints. Within Go, build leaf logic before
wiring. The Electron shell is last (it only needs the live daemon + served UI).

Conventions for every Go task: `golang-pro` skill first; `%w` wraps, `log/slog`,
`context.Context` first arg, no `panic`/`log.Fatal`, small interfaces, compile-time
`var _ Iface = (*T)(nil)`, no hand-edited `go.mod` (`go get` only). For every
test task: `testing-anti-patterns` skill; table-driven `t.Run`, `t.Parallel()`,
`t.TempDir()`, `-race`. For frontend: `react`/`typescript-advanced`/`tanstack`/
`tailwindcss` (+ `vitest`); oxlint/oxfmt 4-space, explicit `: ReactElement`. Run
`rc-final-verify` + `make verify` before declaring any Go-touching task done.

The global gate (run after the relevant phase, not per micro-step):
`make verify` green; `bun run frontend:typecheck`, `bun run frontend:test`,
`bun run codegen-check` green.

---

## Phase 1 — Go: atomic config write (leaf logic)

### T1. `workspace.WriteConfig` — validated atomic typed TOML write

- **Files:** `internal/core/workspace/config_write.go` (new).
- **Do:** Add `func WriteConfig(ctx context.Context, configPath string, cfg ProjectConfig, scope configScope) error` (scope param so the caller selects `globalConfigScope` vs `workspaceConfigScope`). Call `cfg.validate(scope)` first; on failure return before touching disk. Marshal with `toml.Marshal` (go-toml/v2 v2.3.0 — the only TOML lib). Write atomically using the exact temp-file + `fsync` + rename + dir-`fsync`, perms `0o600` pattern from `daemon.WriteInfo` (`internal/daemon/info.go` l.95–123).
- **Done:** `go build ./internal/core/workspace/...` compiles; the function exists with the signature above and reuses `cfg.validate`. (Feeds AC4.)

### T2. `WriteConfig` tests — atomicity, round-trip, validation

- **Files:** `internal/core/workspace/config_write_test.go` (new).
- **Do:** Table-driven, `t.TempDir()`: (a) round-trip equality — `LoadConfig(ctx, WriteConfig(cfg))` deep-equals `cfg` for a representative non-empty `ProjectConfig`; (b) atomicity — an injected failing writer (or pre-rename failure) leaves any pre-existing file byte-identical and writes no partial file; (c) validation — an invalid `ProjectConfig` is rejected and the on-disk file is unchanged / not created.
- **Done:** `make test` passes these with `-race`. (AC4: invalid edit rejected, file unchanged; saved values re-read.)

---

## Phase 2 — Go: contract types + route inventory + OpenAPI

### T3. Contract response types

- **Files:** `internal/api/contract/types.go` (edit).
- **Do:** Add `ConfigDocumentResponse`, `ExtensionListResponse`, `AgentListResponse`, and their element types (extension item: `Ref`, enabled, source, manifest summary; agent item: name, scope/source, description, warnings), following existing `*Response` naming and JSON tags. `ConfigDocument` mirrors the JSON projection of `ProjectConfig` (`config_types.go` l.14–97): `defaults`, `tasks`, `fix_reviews`, `fetch_reviews`, `watch_reviews`, `exec`, `runs`, `sound`, all optional.
- **Done:** Package compiles; types exported. (Feeds AC6/AC7.)

### T4. Route inventory entries

- **Files:** `internal/api/contract/routes.go` (edit, l.16 inventory).
- **Do:** Add six `RouteSpec` entries: `GET/PUT /api/config/global` (`TimeoutRead`/`TimeoutMutate`), `GET/PUT /api/config/workspace` (`TimeoutRead`/`TimeoutMutate`), `GET /api/catalog/extensions` (`TimeoutRead`), `GET /api/catalog/agents` (`TimeoutRead`), each with the matching `ResponseType` from T3.
- **Done:** `internal/api/contract` compiles; entries present. (AC6.)

### T5. OpenAPI spec + regenerate TS types

- **Files:** `openapi/rc-daemon.json` (edit); `web/src/generated/rc-openapi.d.ts` (regenerated, do not hand-edit).
- **Do:** Add the six paths and their request/response schemas (ConfigDocument, ExtensionList, AgentList) to `openapi/rc-daemon.json`. Run `bun run codegen`.
- **Done:** `bun run codegen-check` is diff-clean; generated TS contains the new operations/types. (AC7.)

---

## Phase 3 — Go: service interfaces + implementations

### T6. Service interfaces on the Handlers facade

- **Files:** `internal/api/core/interfaces.go` (edit).
- **Do:** Add `ConfigService` (`GetGlobal`/`PutGlobal`/`GetWorkspace`/`PutWorkspace`) and `CatalogService` (`Extensions`/`Agents`) interfaces exactly as in SPEC §3.1. Add `Config ConfigService` and `Catalog CatalogService` fields to `Handlers` and `HandlerConfig` (l.16–53) and propagate them in `Clone()` (l.104–120).
- **Done:** `internal/api/core` compiles; both fields exist and are cloned. (Feeds AC4/AC5.)

### T7. ConfigService implementation

- **Files:** `internal/core/configsvc/service.go` (new); `internal/core/configsvc/service_test.go` (new).
- **Do:** Thin adapter: `GetGlobal` → `workspace.LoadGlobalConfig`; `GetWorkspace` → resolve root via injected `WorkspaceService.Get` then `workspace.LoadConfig`; `PutGlobal`/`PutWorkspace` → `workspace.WriteConfig` (T1) with the correct scope, then re-load and return the persisted doc. Map `ProjectConfig` ↔ `ConfigDocument` both ways. Add `var _ core.ConfigService = (*Service)(nil)`. Tests mock the workspace store via interface; assert mapping, re-read-after-write, and error propagation (no production test-only methods).
- **Done:** `make test` green for the package with `-race`. (AC4.)

### T8. CatalogService implementation

- **Files:** `internal/core/catalog/service.go` (new); `internal/core/catalog/service_test.go` (new).
- **Do:** `Extensions` → build `extension.Discovery` for the workspace root, `Discover(ctx)`, project `DiscoveryResult.Extensions` (`Ref`, `Manifest`, `Enabled`, `Source`) to `ExtensionList`; include `Failures`/`Overrides` as metadata, not errors. `Agents` → `agents.New().Discover(ctx, root)` → project `Catalog.Agents` (`ResolvedAgent`) to `AgentList`; surface `Catalog.Problems` as per-item warnings. **Confirm exact exported field names against `discovery.go`/`agents.go` at build time** (SPEC risk #3) and adjust the projection. Add compile-time check. Tests mock discovery via interface; assert read-only mapping and that non-fatal failures do not error the call.
- **Done:** `make test` green for the package; read-only (no enable/disable). (AC5.)

---

## Phase 4 — Go: handlers, routes, middleware, wiring

### T9. Config + catalog handlers

- **Files:** `internal/api/core/handlers.go` (edit).
- **Do:** Add handlers for the six routes mirroring `UpdateWorkspace` (l.508–533): nil-service guard → (PUT) `bindJSON` → call service → respond with the contract type; errors via `h.respondError`. Validation failures from `WriteConfig` surface as `400 config_invalid` (file not written). Workspace-scoped handlers read the ID via `core.ActiveWorkspaceIDFromContext` and resolve root through `WorkspaceService.Get`.
- **Done:** Package compiles; handlers return the standard `*Problem` envelope on error. (AC4: invalid → standard envelope.)

### T10. Route registration

- **Files:** `internal/api/core/routes.go` (edit, l.11).
- **Do:** Register the six routes in `RegisterRoutes` with the timeout classes from T4, guarding on nil handlers like existing routes.
- **Done:** Routes resolve at runtime; package compiles. (AC6.)

### T11. Workspace-context middleware

- **Files:** `internal/api/httpapi/browser_middleware.go` (edit, `requiresActiveWorkspace` l.247–267).
- **Do:** Add `/api/config/workspace`, `/api/catalog/extensions`, `/api/catalog/agents` to `requiresActiveWorkspace` so a missing `X-rc-Workspace-ID` yields `412 workspace_context_missing`. Do **not** add the global-config paths. Do not alter Origin/Host/CSRF logic.
- **Done:** A workspace-scoped request without the header returns 412; with the header it reaches the handler. (AC2: no Origin/Host/CSRF regressions; AC4/AC5.)

### T12. Daemon transport wiring + contract parity test

- **Files:** daemon transport bootstrap that builds `core.NewHandlers` (the `HandlerConfig` assembly site); `internal/api/core/handlers_test.go` + `internal/api/core/handlers_error_paths_test.go` (edit).
- **Do:** Construct `configsvc.Service` and `catalog.Service` and pass them into `HandlerConfig`. Extend handler tests with the new routes: success, `config_invalid` 400, missing-workspace 412, service-unavailable (nil service). `openapi_contract_test.go` must pass inventory↔OpenAPI parity unchanged.
- **Done:** `make verify` passes 100% (fmt, zero-issue lint, `test -race`, build); `openapi_contract_test.go` green. (AC6.)

---

## Phase 5 — Web UI (shared, browser + Electron)

### T13. Config feature module

- **Files:** `web/src/systems/config/*` (new — Query hooks + zod schema + base-ui form).
- **Do:** `useGlobalConfig`, `useWorkspaceConfig`, `useSaveConfig` (TanStack Query) over the new endpoints via the existing `openapi-fetch` client (already carries CSRF + `X-rc-Workspace-ID`). A `@rodolfochicone/ui` (base-ui) form driven by a zod schema mirroring `ProjectConfig`'s optional fields; on `config_invalid`, render the standard error-envelope `message`; on success, invalidate the query so the UI shows persisted values. No shadcn/radix, no `web/` rewrite, 4-space, explicit `: ReactElement`.
- **Done:** `bun run frontend:typecheck` green; Vitest + MSW tests cover load / save / invalid-error render. (AC4.)

### T14. Workspaces feature module

- **Files:** `web/src/systems/workspaces/*` (new).
- **Do:** list / register / rename / unregister using the **existing** `POST/GET/PATCH/DELETE /api/workspaces` endpoints (no new Go) through `openapi-fetch` + Query, with `X-rc-Workspace-ID` plumbing already in the client.
- **Done:** Vitest + MSW tests cover register/rename/unregister against mocked existing endpoints; typecheck green. (AC5.)

### T15. Extensions feature module (read-only)

- **Files:** `web/src/systems/extensions/*` (new).
- **Do:** Read-only tables from `/api/catalog/extensions` and `/api/catalog/agents` via Query hooks. No mutate actions (out of scope).
- **Done:** Vitest test renders both lists from mocked endpoints; typecheck green. (AC5.)

### T16. Routes + generated route tree

- **Files:** `web/src/routes/_app/config.tsx`, `web/src/routes/_app/workspaces.tsx`, `web/src/routes/_app/extensions.tsx` (new); regenerated TanStack route tree via `web/scripts/tsr-generate.mjs` (do not hand-edit generated output).
- **Do:** Add the three file-based routes under `_app/` (alongside `workflows.tsx`/`runs.tsx`), wiring the feature modules from T13–T15. Run `tsr-generate.mjs`.
- **Done:** `bun run frontend:typecheck`, `bun run frontend:test`, `bun run codegen-check` all green; routes navigable. (AC5/AC7.)

---

## Phase 6 — Electron shell (`apps/desktop/`)

### T17. Workspace scaffold + toolchain wiring

- **Files:** `apps/desktop/package.json` (new — pin Electron `38.x`, electron-builder `25.x`, TS `^6`, vitest, oxlint/oxfmt; deps via `bun add`), `apps/desktop/tsconfig.json` (extends `tsconfig.base.json`), root `package.json` (add `"apps/*"` to `workspaces`), `turbo.json` (add `apps/desktop` `build`/`lint`/`test`/`typecheck`/`package` tasks).
- **Do:** Stand up the bun/turbo workspace inheriting repo oxlint/oxfmt configs. JS deps via `bun add`, never hand-edited.
- **Done:** `bun install` resolves and writes `bun.lock`; `bun run --filter @rc/desktop typecheck` succeeds on an empty entry; `apps/desktop` appears in turbo's task graph. (AC8.)

### T18. Binary resolver + `daemon.json` reader (injectable)

- **Files:** `apps/desktop/src/main/binary.ts`, `apps/desktop/src/main/info.ts` (new) + their `*.test.ts`.
- **Do:** Binary resolver with order `RC_BINARY` → `process.resourcesPath/bin/rc` → `rc` on `PATH` (first existing+executable wins). `info.ts` reads `~/.rc/daemon/daemon.json` (`path.join(os.homedir(), '.rc', 'daemon', 'daemon.json')`, mirroring `InfoPath`), parsing the `daemon.Info` shape `{ pid, http_port, socket_path, state, started_at, version }`. Both exposed via interfaces for injection.
- **Done:** Vitest (temp dir, fake fs/path) proves resolution order and `daemon.json` parse/validate. (AC1.)

### T19. Daemon supervisor

- **Files:** `apps/desktop/src/main/daemon.ts` (new) + `daemon.test.ts`.
- **Do:** Implement the `Supervisor` interface (SPEC §2.4) with the `starting → healthy → unhealthy → stopped` state machine: attach-or-spawn (probe `/api/daemon/health`; attach with `ownsDaemon=false` if healthy, else spawn `rc daemon start` with `ownsDaemon=true`, poll every 250ms up to the 30s ceiling); 3s health/status polling with `AbortController` 2s budget; crash restart with bounded exponential backoff (max 5: 1→2→4→8→16s) then give-up to unhealthy; graceful `POST /api/daemon/stop` on quit **only if owned**, attached daemon left running. Inject the binary resolver, `daemon.json` reader, and a fake `child_process` for tests.
- **Done:** Vitest proves: attach-vs-spawn decision (AC1), bounded-backoff restart + give-up (AC3), graceful-stop-only-if-owned (AC3). `bun run --filter @rc/desktop test` green.

### T20. Main process: window, native integration, preload, single-instance

- **Files:** `apps/desktop/src/main/index.ts`, `apps/desktop/src/main/tray.ts`, `apps/desktop/src/main/menu.ts`, `apps/desktop/src/preload/index.ts` (new).
- **Do:** `app.requestSingleInstanceLock()` (second launch focuses existing window). On ready, start the supervisor, show a native "starting" splash, then load the **remote** `http://127.0.0.1:<http_port>` document in a `BrowserWindow` with `contextIsolation:true`, `nodeIntegration:false`, `sandbox:true`. Native menu (`Menu.setApplicationMenu`), `Tray` reflecting `starting/healthy/unhealthy/stopped` with Show / Restart daemon / Quit, dock re-open (`app.on('activate')`), Cmd-R/Cmd-Q. Minimal preload exposing read-only `window.rcdesktop` status only. On `before-quit`, route through `supervisor.stop()`. **No Origin/header injection** — the natural remote-load Origin passes the daemon middleware.
- **Done:** `bun run --filter @rc/desktop typecheck` + oxlint/oxfmt clean. (Sets up AC1/AC2/AC3; manual smoke in T22.)

### T21. Packaging config + docs

- **Files:** `apps/desktop/electron-builder.yml` (new), `apps/desktop/README.md` (new), `Makefile` (edit only if frontend targets need the new workspace).
- **Do:** electron-builder config: `mac.arch: universal` (dmg + zip), hardened runtime + child entitlements, `extraResources: [{ from: ../../bin/rc, to: bin/rc }]` (bundles the `make build` binary). README documents signing (`CSC_LINK`/`CSC_KEY_PASSWORD`), notarization (`notarytool` via `APPLE_ID`/`APPLE_APP_SPECIFIC_PASSWORD`/`APPLE_TEAM_ID`), and the reproducible build command `bun run --filter @rc/desktop package`. Packaging stays out of `make verify`.
- **Done:** `bun run --filter @rc/desktop package` produces a universal (arm64+x64) `.app`/dmg from the documented command; README lists signing+notarization steps. (AC8.)

### T22. End-to-end smoke verification (real execution)

- **Files:** `apps/desktop/README.md` (smoke checklist section).
- **Do:** Execute and record: (a) launch built `.app` with no daemon running and without `rc daemon start` → rc UI renders within startup timeout; second launch / launch-with-running-daemon attaches, no duplicate (verify single `rc daemon` via `ps`, single `daemon.lock` owner) (AC1); (b) launch a workflow/exec/review, confirm live SSE updates (job/session/tool_call/usage) and **no Origin/Host/CSRF rejections in daemon logs**; kill+restore connectivity → stream resumes via `Last-Event-ID` (AC2); (c) kill daemon while app runs → bounded auto-restart + UI recovers; quit app → app-started daemon stopped gracefully, `ps` shows no orphan `rc daemon`, `daemon.lock` absent/re-acquirable; attached pre-existing daemon survives quit (AC3).
- **Done:** All three smoke groups pass on macOS and are recorded in the checklist. (AC1, AC2, AC3.)

---

## Final gate

### T23. Full verification sweep

- **Do:** `rc-final-verify` skill, then: `make verify` (fmt → zero-issue lint → `test -race` → build) green — use `GIT_CONFIG_PARAMETERS='safe.bareRepository=all'` prefix if the rtk bare-repo daemon-test gotcha appears; `bun run frontend:typecheck`, `bun run frontend:test`, `bun run codegen-check` green; `bun run --filter @rc/desktop test` + oxlint/oxfmt green; `bun run --filter @rc/desktop package` builds the universal `.app`.
- **Done:** Every gate green. (AC6, AC7, AC8 — and AC1–AC5 covered by their phase tasks.)

---

## Acceptance-criteria coverage map

- **AC1** (launch/attach, no duplicate): T18, T19, T20, T22(a).
- **AC2** (live SSE, no Origin/Host/CSRF rejection, `Last-Event-ID` resume): T11, T20, T22(b).
- **AC3** (bounded restart, graceful owned-only quit, no orphan/lock): T19, T20, T22(c).
- **AC4** (config read/save, atomic, validated, invalid→envelope): T1, T2, T7, T9, T13.
- **AC5** (workspace register/rename/unregister, extensions/agents read-only): T8, T14, T15, T16.
- **AC6** (Go `make verify` + contract parity): T3, T4, T9, T10, T12, T23.
- **AC7** (frontend typecheck/test/codegen-check diff-clean): T5, T13–T16, T23.
- **AC8** (universal `.app` from one command, signing/notarization documented, Electron checks): T17, T21, T22, T23.
