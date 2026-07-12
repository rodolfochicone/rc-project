# STACK — Fork & Rebrand RC → RC

> Single source of truth for the stack, layout, conventions, and rename surface. PRD / SPEC / TASKS all read this.
> **Brownfield fork**: target `~/dev/rc-harness-cli` is empty; the profile below is taken from the **source** at `~/dev/rc` (the tree being forked). Conform to the source's conventions exactly — this is a faithful rebrand, not a redesign.

## Project type

Monorepo: **Go CLI backend** + **bun/Vite/React/TypeScript web frontend**, glued by **Turborepo** + bun workspaces. ~37M source tree (`.git` ~12M of it). Source module: `github.com/rc/rc`. MIT licensed (fork permitted; attribution must be preserved).

## Languages, runtimes, EXACT versions (from manifests + lockfiles)

### Go backend

- **Go:** `1.26.1` (from `go.mod`; enforced by `make check-go-version`).
- **Module path:** `github.com/rc/rc` → target `github.com/rc-tech/rc-harness`.
- **Key direct deps (pinned in `go.mod`):**
  - TUI/theme: `charm.land/bubbles/v2 v2.1.0`, `charm.land/bubbletea/v2 v2.0.2`, `charm.land/huh/v2 v2.0.3`, `charm.land/lipgloss/v2 v2.0.2`, `github.com/charmbracelet/x/ansi v0.11.6` (brand colors live here — `lipgloss.Color`).
  - CLI: `github.com/spf13/cobra v1.10.2` (+ `pflag` indirect).
  - HTTP: `github.com/gin-gonic/gin v1.12.0`.
  - DB: `modernc.org/sqlite v1.49.0` (pure-Go sqlite).
  - MCP/agents: `github.com/modelcontextprotocol/go-sdk v1.5.0`, `github.com/coder/acp-go-sdk v0.6.3`, `github.com/coder/websocket v1.8.14`.
  - Config: `github.com/pelletier/go-toml/v2 v2.3.0`, `gopkg.in/yaml.v3 v3.0.1`.
  - Self-update/release: `github.com/creativeprojects/go-selfupdate v1.5.2`, `github.com/Masterminds/semver/v3 v3.4.0`.
  - Misc: `github.com/fsnotify/fsnotify v1.9.0`, `github.com/gofrs/flock v0.13.0`, `github.com/atotto/clipboard v0.1.4`, `golang.org/x/sys v0.42.0`.
- **`go.sum` present** — lockfile must be regenerated after module-path rename (`go mod tidy`).

### JS/TS frontend

- **Package manager:** `bun@1.3.11` (`.bun-version` = `1.3.11`; root `package.json` `"packageManager": "bun@1.3.11"`). Lockfile: `bun.lock` (261K) — must be regenerated after package renames. Enforced by `make check-bun-version`.
- **Monorepo orchestrator:** `turbo ^2.8.20` (`turbo.json`).
- **Build/dev:** `vite ^8.0.3` (root devDep) / web uses Vite via `@vitejs/plugin-react ^6.0.0` + `@tailwindcss/vite ^4.2.3`.
- **TypeScript:** `typescript ^6.0.2` (root); web `tsconfig.json` extends `tsconfig.base.json` (`target ES2022`, `strict`, `verbatimModuleSyntax`, `noUncheckedIndexedAccess`, `moduleResolution: bundler`).
- **React:** `react ^19.2.0` / `react-dom ^19.2.0`.
- **AI/chat UI:** `@ai-sdk/react ^3.0.170`, `@assistant-ui/react ^0.12.26`, `@assistant-ui/react-ai-sdk ^1.3.20`, `ai ^6.0.168`.
- **Routing/data:** `@tanstack/react-router ^1.168.22` (+ generator/plugin/devtools), `@tanstack/react-query ^5.99.0`, `@tanstack/react-virtual ^3.13.24`.
- **State:** `zustand ^5.0.11`. **Validation:** `zod ^4.3.0`.
- **Styling:** `tailwindcss ^4.2.3` (v4, CSS-first config — no `tailwind.config`; tokens in CSS). `class-variance-authority ^0.7.1`, `clsx ^2.1.1`, `tailwind-merge ^3.5.0`.
- **UI kit:** shadcn (`web/components.json`, `style: default`, `baseColor: neutral`, `iconLibrary: lucide` → `lucide-react ^1.8.0`), consumed via workspace pkg `@rc/ui`.
- **API client:** `openapi-fetch ^0.17.0` + `openapi-typescript ^7.9.1` (codegen from `openapi/`). `sonner` toasts.
- **Storybook:** `^10.3.5` (`@storybook/react-vite`, addon-a11y/docs/themes).

## Workspace layout (bun workspaces, from root `package.json`)

`workspaces: ["packages/ui", "web", "sdk/*"]`.

```
~/dev/rc/
├── go.mod / go.sum / rc.go        # public Go package `package rc` (re-exports internal/core types)
├── cmd/rc/                        # CLI main (main.go, main_test.go) → binary `rc`
├── internal/                           # api/ charmtheme/ cli/ config/ contentblock/ core/ daemon/ logger/ setup/ store/ update/ version/
├── pkg/rc/                        # public Go subpkgs: events/ runs/
├── web/                                # React app: src/ (generated/ lib/ routes/ storybook/ systems/ test/), e2e/, public/, .storybook/, embed.go (go:embed of dist)
├── packages/ui/                        # @rc/ui shadcn kit: src/ (assets/ components/ lib/ index.ts tokens.css)
├── sdk/                                # create-extension/ extension/ extension-sdk-ts/
├── extensions/                         # rc-idea-factory/ rc-qa-workflow/
├── skills/                             # rc/ + cy-* skills, embed.go (go:embed); skills-lock.json at root
├── agents/                             # embed.go (go:embed reusable agents)
├── openapi/                            # rc-daemon.json (OpenAPI spec → TS codegen source)
├── docs/ scripts/ imgs/ zsh/ aur-pkg/ test/ (mixed *.test.ts + *_test.go)
└── tooling dotfiles (see below)
```

- **Asset embedding:** `go:embed` in `web/embed.go`, `skills/embed.go`, `agents/embed.go`, `internal/core/extension/discovery_bundled.go`, `internal/core/prompt/templates.go` — renames must keep embed paths valid.
- `.agents/` is a **real directory** (skills mirror, not a symlink).

## Test frameworks & how tests run

- **Go:** standard `go test ./...` (table tests; `_test.go` throughout). `make test` / `test-coverage` / `test-nocache`.
- **Web unit:** **Vitest** `^4.1.0` (root `vitest.config.ts` for `test/*.test.ts`; web `web/vitest.config.ts`; pkg `packages/ui` has its own). `jsdom ^27.1.0`, `@testing-library/react ^16.3.0`, `@testing-library/jest-dom`, `@testing-library/user-event`, `msw ^2.13.4` for mocking. Coverage via `@vitest/coverage-v8`.
- **Web e2e:** **Playwright** `^1.55.0` (`web/playwright.config.ts`).
- **Aggregate verify:** `make verify` = `frontend-verify fmt lint test go-build frontend-e2e`. `make frontend-verify` = `frontend-lint frontend-typecheck frontend-test frontend-build`.
- Contract tests exist in `test/` that assert config/codegen integrity (`frontend-workspace-config.test.ts`, `codegen-script.test.ts`, `frontend-verification-contract.test.ts`, `release_config_test.go`, `public_api_test.go`, `skills_bundle_test.go`) — these will likely encode `rc` identity and must be updated to assert RC intent.

## Lint / format / hooks / CI / release

- **Go lint:** `golangci-lint` v2 config (`.golangci.yml`): enables `bodyclose errcheck funlen goconst gocritic gocyclo gosec govet ineffassign …`. `make lint` runs it.
- **JS/TS lint+format:** **oxlint** `^1.60.0` (`.oxlintrc.json`) + **oxfmt** `^0.46.0` (`.oxfmtrc.json`). `lint-staged` routes `*.go → make fmt`, `*.{ts,tsx,js,jsx} → oxfmt + oxlint`, `*.{css,json,yaml,md} → oxfmt`.
- **Commits:** commitlint `^20.5.0` conventional config (`.commitlintrc.yaml`); `cliff.toml` for git-cliff changelogs.
- **Git hooks:** **husky** `^9.1.7` (`.husky/commit-msg`, `.husky/pre-commit`) and a parallel `.pre-commit-config.yaml` (pre-commit framework). `prepare: husky`; `postinstall: bash scripts/link-skills.sh`.
- **Release:** **goreleaser** (`.goreleaser.yml`) — `project_name: rc`, build id `rc`, `main: ./cmd/rc`, `binary: rc`, ldflags inject `github.com/rc/rc/internal/version.{Version,Commit,Date}`; brew tap `homebrew-rc`, npm `@rc/cli`, AUR `aur-pkg/` (commented), homepage `https://rc.com` / `github.com/rc/rc`, maintainer `RC Team <support@rc.com>`. Plus `.goreleaser.release-{header,footer}.md.tmpl`.
- **CI:** `.github/` (workflows) — must be re-skinned to the new name.
- **Editor/misc:** `.editorconfig`, `.coderabbit.yaml`, `.yamlfix.toml`, `.prettierignore`, `tsconfig*.json`.

## Brand surface (what the rebrand touches)

### CLI terminal palette — `internal/charmtheme/theme.go` (source of truth, asserted by `theme_test.go`)

Current **green** brand → must become RC **orange→amber**:

- `ProgressGradientStart = "#65A30D"`, `ProgressGradientEnd = "#CAEA28"`
- `ColorBrand = #CAEA28`, `ColorAccent = #A3E635`, `ColorAccentAlt = #84CC16`, `ColorAccentDeep = #65A30D`
- Backgrounds (likely kept dark-neutral): `ColorBgBase #0C0A09`, `ColorBgSurface #1C1917`, `ColorBgOverlay #292524`; semantic `Success #10B981 / Error #EF4444 / Warning #F59E0B / Info #3B82F6`; fg `#E7E5E4 / Muted #A8A29E / Dim #78716C`; `ColorBorder #44403C`, `ColorBorderFocus = ColorAccent`.
- Tests pinning these: `internal/charmtheme/theme_test.go` (asserts each hex), `internal/cli/theme_test.go` (asserts `ColorBrand` usage + `"RC // SETUP"` / `"RC // INTERACTIVE INPUT"` strings).

### Web palette — Tailwind v4 CSS tokens

- `web/src/styles.css`: radial-gradient brand accent `rgba(214,242,74,...)` (= green `#d6f24a`).
- `packages/ui/src/tokens.css`: `--brand: #d6f24a;` (+ ~8.8K of token defs, light+dark). **This is the web brand source of truth.**
- shadcn aliases in `web/components.json`: `@rc/ui`, `@rc/ui/utils` → must rename to RC scope.

### Welcome header — `internal/cli/setup.go`

- `printWelcomeHeader(cmd)` called at ~L126, defined at ~L490. **Current implementation renders a simple box: title `"RC // SETUP"` + subtitle** (lipgloss box, `newCLIChromeStyles()`).
- **AMBIGUITY (flag for orchestrator):** the REQUEST describes a much richer header (rounded orange box, `RC vX.Y.Z` label, `Welcome back <user>!`, ASCII peak mascot, dimmed model·plan·email/org·cwd line). The source header today does **not** contain that user/model/plan/cwd data. PRD/SPEC must resolve whether C4 = (a) faithful re-skin+relabel of the existing simple box, or (b) building a new richer header (which is net-new behavior beyond a rebrand and would conflict with CON1/AC6). The REQUEST text says both "re-implement to the RC layout" and "preserve the header's existing data sources/behavior; only re-skin + relabel" — these two are in tension.

### Logo / image assets

- `imgs/`: `screenshot.png`, `how-it-works-flow.{png,svg}`, `how-it-works.drawio` (SVG/drawio contain green hexes).
- `web/public/`: `symbol.png` (140K — likely the logo), `mockServiceWorker.js`.
- `docs/design/daemon-mockup/`: `assets/rc-icon.png`, `colors_and_type.css`, `src/*.jsx` (contain green hexes).
- README hero / badges reference RC logo + shields.

## Rename inventory — scale (case-insensitive `rc`, excl `node_modules` + `.git`)

- **~438 `.go` files** contain `rc`.
- **~82 web `.ts`/`.tsx` files** contain `rc`.
- **~1299 files total** contain `rc` across the tree (docs, configs, skills, fixtures, ledgers).
- **Green brand hexes** appear in: `internal/charmtheme/theme.go` + `theme_test.go`, `internal/core/run/ui/view_test.go`, `packages/ui/src/tokens.css`, `web/src/styles.css` (via rgba), `imgs/*.svg`/`.drawio`, `docs/design/daemon-mockup/*`.

### Identity-bearing PATHS (rename, not just string-replace)

- `cmd/rc/` → `cmd/rc/`; `rc.go` (`package rc`) → RC-named file+package; `pkg/rc/` → RC-scoped path.
- `zsh/rc-completion/rc-completion.plugin.zsh`; `openapi/rc-daemon.json`; `web/src/generated/rc-openapi.d.ts`; `docs/design/daemon-mockup/assets/rc-icon.png`; `skills/rc/`; `.agents/skills/rc/`.
- Package names: root `package.json "name": "rc"`; `web/package.json "rc-web"`; `packages/ui "@rc/ui"`; goreleaser npm `@rc/cli`.
- User config dir: `internal/config/home.go` `DirName = ".rc"` → `.rc` (plus every `~/.rc` ref in code, tests, fixtures, error strings like `"create rc directory"`).

### Lockfiles to regenerate (do NOT hand-edit)

`go.sum` (via `go mod tidy`), `bun.lock` (via `bun install`), `web/src/routeTree.gen.ts` (TanStack, gitignored/regenerated), `web/src/generated/rc-openapi.d.ts` (via codegen), `skills-lock.json` (gitignored; generated).

### Conventions to honor (do not "improve")

- shadcn `baseColor: neutral`, `iconLibrary: lucide` — keep; only swap brand hue tokens.
- Tailwind v4 CSS-first (tokens in `tokens.css` / `styles.css`, no JS config).
- Go: `internal/` for private, `pkg/` + root `rc.go` for public API surface — keep boundaries; `public_api_test.go` guards it.
- `oxfmt`/`oxlint` formatting (NOT prettier/eslint); `golangci-lint v2`.
- Conventional commits; husky + pre-commit dual hooks.

## Open ambiguities (for orchestrator to resolve before SPEC)

1. **Welcome header scope (above):** faithful re-skin of the existing simple `RC // SETUP` box vs. building the richer `Welcome back <user>` mascot header. The latter is net-new behavior and conflicts with the "zero functional change" constraint. **Recommend:** treat the richer layout as the brand target only if its data (user/model/plan/org/cwd) already exists in the CLI context; otherwise faithfully re-skin the current box and move the richer header to "Proposals."
2. **`cy-*` skill/extension prefix** (`skills/cy-*`, `extensions/cy-*`, `.agents/skills/cy-*`): is `cy` an abbreviation of "RC" (rename to an RC prefix) or a stable public skill identifier (keep, to avoid breaking skill IDs in `skills-lock.json` and embeds)? Renaming changes public skill names; not renaming leaves a RC-derived token. **Recommend:** SPEC decides explicitly; default to keep (stable IDs) and document as an intentional preservation under D5.
3. **`~/.rc` → `~/.rc` migration:** REQUEST defers auto-migration to Proposals; default is fresh `.rc` with no silent migration. Confirm no committed `.rc/` runtime content needs porting (the source's `.rc/tasks`, `.rc/research` are source-repo working data, not shippable product state — exclude from the fork copy).
4. **MIT attribution:** which exact `rc`/author strings are _intentionally preserved_ in `LICENSE` (and any NOTICE) — must be enumerated so the D5 residual-`rc` check has a documented allowlist.
5. **Copy exclusions:** `.git/` (~12M), `bin/`, `dist/`, `web/dist/*`, `node_modules`, `.turbo/`, `coverage*`, `*.tsbuildinfo`, generated `routeTree.gen.ts`, and source working dirs (`.codex/`, `.claude/`, `.rc/`, `.factory/`, `.resources/`, `ai-docs`) per `.gitignore` — SPEC must define the precise copy/exclude set.
