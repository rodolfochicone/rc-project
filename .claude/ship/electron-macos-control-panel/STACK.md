# STACK.md — rc Electron macOS Control Panel

Single source of truth for stack, versions, layout, and conventions. Every later
phase (PRD/SPEC/TASKS) and subagent reads this. This is a **brownfield** target:
conform to what exists; do not substitute frameworks. The Electron shell is the
only genuinely new stack and the user named it explicitly — honor it.

## Repo identity

- Module: `github.com/rodolfochicone/rc-project` (path: `/Users/rodolfo.chicone/dev/rc-project`)
- **Not a git repo** at the worktree root (no `.git`). Git-safety rules still apply
  if/when one exists; never run destructive git commands without explicit permission.
- Monorepo: Go main module + bun/turbo JS/TS workspaces. No `apps/` dir yet.

## Stacks & EXACT versions (from manifests + lockfiles)

### Go (daemon, CLI, core) — source of truth

- Go `1.26.1` (`toolchain go1.26.4`); local toolchain `go1.26.4 darwin/arm64`.
- HTTP: `github.com/gin-gonic/gin v1.12.0` (Gin).
- TOML: **`github.com/pelletier/go-toml/v2 v2.3.0`** — already the project's TOML lib.
  Config writes MUST evaluate/reuse this; do not add a new TOML dependency.
- File lock: `github.com/gofrs/flock v0.13.0` (daemon singleton lock).
- File watch: `github.com/fsnotify/fsnotify v1.9.0`.
- CLI: `github.com/spf13/cobra v1.10.2`. TUI: charm bubbletea/v2 `2.0.2`, huh/v2, lipgloss/v2.
- SQLite (global store): `modernc.org/sqlite v1.49.0` (pure-Go).
- Lint: `golangci-lint` via `.golangci.yml`, **zero-issue tolerance**.

### Frontend (bun + turbo monorepo)

- Package manager: **bun `1.3.11`** (pinned `.bun-version`, `packageManager: bun@1.3.11`). Use `bun`, never npm/yarn/pnpm.
- Monorepo orchestrator: **turbo `^2.8.20`** (`turbo.json`).
- Workspaces: `packages/ui` (`@rodolfochicone/ui`), `web` (`rc-web`), `sdk/*`.
- TypeScript `^6.0.2`. Vite `^8.0.3`. Vitest `^4.1.0` (test runner).
- Lint/format: **oxlint `^1.60.0` + oxfmt `^0.46.0`** (NOT eslint/prettier). Configs: `.oxlintrc.json`, `.oxfmtrc.json`.
- Node available locally: `v22.14.0` (used by `scripts/*.mjs` codegen).

### `web` (rc-web — the UI Electron will wrap; DO NOT rewrite)

- React `^19.2.0` + react-dom `^19.2.0`.
- Routing: **TanStack Router `^1.168.22`** (file-based, generated routes via `@tanstack/router-plugin` / `router-generator`; codegen step `web/scripts/tsr-generate.mjs`).
- Data: **TanStack Query `^5.99.0`**; virtual: `@tanstack/react-virtual ^3.13.24`.
- HTTP client: `openapi-fetch ^0.17.0` typed against generated OpenAPI types.
- State: `zustand ^5.0.11`. Schema: `zod ^4.3.0`.
- Styling: Tailwind `^4.2.3` (v4, `@tailwindcss/vite`), `class-variance-authority`, `tailwind-merge`, `clsx`. Icons `lucide-react`. Toasts `sonner`.
- AI: `ai ^6.0.168`, `@ai-sdk/react`, `@assistant-ui/react`.
- Tests: Vitest + `@testing-library/react ^16.3.0` + `jsdom` + `msw ^2.13.4`. E2E: Playwright `^1.55.0` (`web/playwright.config.ts`). Storybook `^10.3.5`.
- Embedded into the Go binary via `web/embed.go` (`//go:embed all:dist`, pkg `webassets`). Dev proxy via daemon `--web-dev-proxy`.

### `packages/ui` (`@rodolfochicone/ui` — shared component lib; reuse, do not fork)

- Headless: **`@base-ui/react ^1.2.0`** (base-ui, NOT shadcn/radix). Markdown via `react-markdown` + `rehype-sanitize` + `remark-gfm`. `tailwind-variants`.
- Peer: react/react-dom `^19`.

### Electron shell (NEW — user-named stack, honor exactly)

- No Electron app exists yet (`apps/` absent). New app lives at `apps/desktop/` per REQUEST.
- Must integrate as a bun/turbo workspace (add `apps/*` to root `package.json` `workspaces` and turbo pipeline) using the SAME toolchain: bun, TypeScript `^6`, oxlint/oxfmt, vitest.
- Electron + electron-builder versions are greenfield: pick latest stable at build time, document exact pins in the app's own `package.json`. Target: universal macOS `.app` (arm64 + x64), code signing + notarization documented.
- The Electron renderer does NOT host React — it loads the daemon-served UI at `http://127.0.0.1:<port>`. Electron is a thin shell + lifecycle owner only.

## Architecture & integration points (load-bearing)

### Daemon HTTP API (Gin) — `internal/api/`

- Server: `internal/api/httpapi/server.go` — binds `127.0.0.1`, functional-options `Option` pattern, `PortUpdater` persists chosen port.
- Routes (shared): `internal/api/core/routes.go` (`RegisterRoutes(router gin.IRouter, handlers *core.Handlers)`). Handlers in `internal/api/core/handlers.go`.
- Route contract inventory: `internal/api/contract/routes.go` (`RouteInventory []RouteSpec{Method, Path, ResponseType, TimeoutClass}`). **Every new route MUST be added here** — `openapi_contract_test.go` enforces parity.
- Existing daemon endpoints: `GET /api/daemon/status` (probe), `GET /api/daemon/health` (probe), `GET /api/daemon/metrics`, `POST /api/daemon/stop` (mutate). Workspaces: `POST/GET /api/workspaces`, `POST /api/workspaces/sync`, `GET/PATCH/DELETE /api/workspaces/:id`, `POST /api/workspaces/resolve`, `GET /api/workspaces/:id/ws`. Plus tasks/reviews/runs/exec/sync/ui.
- **SSE**: `GET /api/runs/:run_id/stream` with `Last-Event-ID`, heartbeat, overflow, cursor. Event kinds in `pkg/rc/events/event.go`. Durable journal: `~/.rc/runs/<run-id>/events.jsonl`.

### Security middleware (Electron BrowserWindow MUST satisfy these) — `internal/api/httpapi/browser_middleware.go`

- **Host validation**: `hostValidationMiddleware` — only `127.0.0.1`/`localhost` hosts allowed.
- **Origin validation**: `originValidationMiddleware` — non-empty `Origin` must pass `originAllowed`. Configure Electron requests/window so Origin is a permitted localhost origin or absent as appropriate.
- **CSRF**: cookie-based (`csrfCookieLifetime = 24h`).
- **Active workspace**: `X-rc-Workspace-ID` header (`core.HeaderActiveWorkspaceID`) required on workspace-scoped routes; missing → `412 workspace_context_missing`.

### Timeout classes — `internal/api/contract/timeout.go`

`probe` (2s), `read` (15s), `mutate` (30s), `long_mutate` (120s), `stream` (0/none). New config read routes → `TimeoutRead`; config write routes → `TimeoutMutate`.

### Error format — `internal/api/core/errors.go` / `contract`

`*Problem` via `core.NewProblem(status, code, message, details, err)`; `core.RespondError(c, err)`. JSON shape carries `code` / `message` / `request_id` (`RequestIDFromContext`). Reuse this; do not invent a new error envelope.

### Config (the new endpoints' domain) — `internal/core/workspace/`

- Types: `config_types.go` — `ProjectConfig{ Defaults, Tasks, FixReviews, FetchReviews, WatchReviews, Exec, Runs, Sound }` (all fields pointer-optional TOML).
- Load/merge/validate: `config.go`, `config_merge.go`, `config_validate.go`. Resolution via `Resolve(ctx, startDir)` → `Context{ ConfigPath, WorkspaceConfigPath, GlobalConfigPath, Config }`. Uses `pelletier/go-toml/v2`.
- Global config: `~/.rc/config.toml`; workspace: `<root>/.rc/config.toml`.
- **Writes must be atomic** (temp file + rename — pattern already used by `daemon.WriteInfo` in `internal/daemon/info.go`) and preserve comments/structure where feasible with go-toml/v2.

### Daemon lifecycle & discovery (Electron owns this) — `internal/daemon/`, `internal/config/home.go`

- Discovery record `daemon.json` (`internal/daemon/info.go`: `Info{ PID, Version, SocketPath, HTTPPort, StartedAt, State }`), written atomically. Path resolved via `internal/config/home.go`: `~/.rc/daemon/daemon.json` (`InfoPath`), socket `daemon.sock`, lock `daemon.lock`. Defaults: `DefaultHTTPPort = 2323`, `EphemeralHTTPPort = -1`.
- Singleton lock: `internal/daemon/lock.go` (flock, `ErrAlreadyRunning`). Electron must **attach** to a running daemon, not duplicate — respect the lock.
- Spawn binary `rc daemon start` (entry `cmd/rc/main.go`, command `internal/cli/daemon_commands.go`, default startup timeout 10s). Electron reads port/socket from `daemon.json`, never hardcodes.

### Web UI surface (new shared pages live here) — `web/src/`

- Layout: `web/src/routes/` (file-based TanStack Router; `_app.tsx` layout + `_app/` children like `workflows.tsx`, `runs.tsx`, `reviews.tsx`). Feature modules under `web/src/systems/*` (`app-shell`, `dashboard`, `workflows`, `runs`, `reviews`, `spec`, `memory`).
- New pages (config edit global/workspace, workspace management, extensions/agents read-only) follow the route + `systems/<feature>` pattern, typed via generated OpenAPI types and `openapi-fetch` + TanStack Query.

### OpenAPI codegen

- Source of truth: `openapi/rc-daemon.json`. Generated TS: `web/src/generated/rc-openapi.d.ts` via `scripts/codegen.mjs` (uses `openapi-typescript`). Router types via `web/scripts/tsr-generate.mjs`.
- After adding endpoints: update `openapi/rc-daemon.json`, run `bun run codegen`; `bun run codegen-check` must be green (diff-clean).

## Conventions (the rules to fit in)

### Go (enforced by CLAUDE.md + golangci-lint, zero tolerance)

- Wrap errors `fmt.Errorf("context: %w", err)`; match with `errors.Is`/`errors.As`, never string compare.
- `log/slog` for logging; no `log.Printf`/`fmt.Println` for operational output.
- `context.Context` first arg across runtime boundaries; avoid `context.Background()` outside `main`/focused tests.
- No `panic`/`log.Fatal` in production paths.
- Small interfaces (accept interfaces, return structs); functional options for complex constructors; compile-time checks `var _ Iface = (*T)(nil)`.
- Every goroutine: explicit ownership + `ctx.Done()` shutdown; no fire-and-forget; no `time.Sleep` in orchestration.
- No `interface{}`/`any` when concrete type known; no reflection without justification.
- Never hand-edit `go.mod` — use `go get`.

### TypeScript / frontend

- Strict TS (`tsconfig.base.json`), explicit return types in existing components (e.g. `: ReactElement`).
- oxlint + oxfmt (4-space indent per existing files); no eslint/prettier.
- TanStack Router file-based routes (generated, do not hand-edit generated route tree). TanStack Query for server state, zustand for client state, zod for schemas.
- Reuse `@rodolfochicone/ui` (base-ui) components; do not introduce shadcn/radix or rewrite `web/`.

### Tests

- Go: table-driven `t.Run` subtests, `t.Parallel()`, `t.TempDir()`, `t.Helper()`, `-race` mandatory. Mock via interfaces. Apply `testing-anti-patterns` skill.
- Frontend: Vitest + Testing Library + MSW; Playwright for E2E. Tests encode intent.

## Verification gates (BLOCKING — task not complete until green)

- Go: `make verify` = `fmt -> lint (zero issues) -> test -race -> build`. **Known gotcha (MEMORY):** under `rtk`, daemon tests can fail in bare-repo mode — prefix with `GIT_CONFIG_PARAMETERS='safe.bareRepository=all'` when needed.
- Frontend: `bun run frontend:typecheck`, `bun run frontend:test`, `bun run codegen-check` (all green).
- Electron app: must build a reproducible universal macOS `.app`; signing/notarization documented.
- Activate `rc-final-verify` before claiming done; `golang-pro` before Go code; `react`/`typescript-advanced`/`tanstack`/`tailwindcss` (+ `vitest`, `testing-anti-patterns`) before frontend; `systematic-debugging` + `no-workarounds` for bugs.

## Risks / ambiguities to resolve before build

1. **Electron + electron-builder versions** are unpinned (greenfield) — pick latest stable at SPEC time and document exact pins; verify electron-builder universal (arm64+x64) + notarization flow on the available macOS toolchain.
2. **Origin/CSRF from Electron**: must confirm precisely which `Origin` the BrowserWindow sends and how `originAllowed`/CSRF cookie behave for `file://`-loaded preload vs. remote `http://127.0.0.1` load. Needs reading `originAllowed`/`security_headers.go` during SPEC.
3. **Daemon binary discovery**: how Electron locates the `rc` binary (bundled in `.app` resources vs. PATH vs. dev build) is unspecified — decide and document.
4. **Quit policy**: REQUEST allows "stop gracefully OR keep daemon, per preference" — pick a default (graceful stop with no orphans/locks) and confirm with user.
5. **TOML comment preservation**: go-toml/v2 marshalling does not round-trip arbitrary comments; clarify expected fidelity ("where possible") — likely write only the typed `ProjectConfig` structure atomically.
6. **`apps/*` workspace wiring**: adding `apps/desktop` requires updating root `package.json` workspaces + turbo pipeline + Makefile frontend targets; confirm CI/verify should cover the Electron build.
