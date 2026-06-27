# SPEC — Fork & Rebrand rc → rc

> The how. Reads: PRD.md (why/what + ACs), STACK.md (detected source stack), REQUEST.md (authoritative request).
> Brownfield faithful fork: conform to the source's structure/conventions exactly. This SPEC defines the mechanical rename + visual rebrand, not a redesign.
> Source: `~/dev/rc`. Target: `~/dev/rc-harness-cli` (empty). Source facts in this doc were re-verified against the live tree.

## 0. Source-fact corrections to STACK.md (verified against the live source)

These supersede STACK.md where they differ; the rename logic below relies on them:

- **goreleaser** (`.goreleaser.yml`) actually has: `license: "BSL-1.1"`, `homepage: "https://github.com/rc/rc"`, `description: "rc CLI"`, `maintainers: ["rc Team <support@rc.com>"]`. STACK.md's "homepage rc.com / MIT" was approximate. The goreleaser `metadata.license` is a label string, **not** the repo LICENSE — leave the LICENSE-file license model untouched (it is MIT) and only relabel brand strings.
- **LICENSE** is MIT, `Copyright (c) 2026 NauckGroup LTDA`. It contains **no `rc` substring** — so the AC5 residual scan needs no allowlist entry for LICENSE. Attribution is by company name and is preserved verbatim (see D5 allowlist).
- **Welcome header** (`internal/cli/setup.go::printWelcomeHeader`, defined L490): renders a 2-line lipgloss box — `title="rc // SETUP"` + a fixed subtitle. It exposes **no** user / model / plan / org / cwd data. A second chrome string `"rc // INTERACTIVE INPUT"` lives in `internal/cli/theme.go` L79. Both are asserted by `internal/cli/theme_test.go` (L64, L84).
- **Version label** for the header comes from `internal/version.Version` (default `"dev"`, injected at release via ldflags). `rc vX.Y.Z` = `"rc " + "v" + version.Version` when not `"dev"`, else `"rc dev"`.
- **`rc.com`** appears in code/tests/docs (`cmd/rc/main*.go`, `go.mod` comment, `test/public_api_test.go`, `test/skills_bundle_test.go`, `Makefile`, `internal/core/...`, skill references). All → `rc.tech`.
- `skills/rc/` is a real skill whose **ID** `rc` is referenced by `skills-lock.json` and the `skills/embed.go` / `.agents/skills/` mirror.

## 1. Decision summary (resolves PRD Open Questions)

| #    | Question                      | Decision                                                                                                                                                                                                                                                                                                        | Rationale                                                                          |
| ---- | ----------------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ---------------------------------------------------------------------------------- |
| D-Q1 | Welcome header scope          | **Re-skin + relabel the existing simple box only.** Rounded orange border, title `rc vX.Y.Z`, equivalent subtitle. The richer `Welcome back <user>` / mascot / model·plan·org·cwd header → **Proposal P1** (data sources do not exist in the current CLI; building them violates AC7 "zero functional change"). | PRD §F3.4, AC6, AC7; REQUEST C4 vs CON1 tension; verified header has no such data. |
| D-Q2 | `cy-*` skill/extension prefix | **Keep `cy-*` as-is** (stable public IDs in `skills-lock.json` + embeds). Renamed only if it appears as a brand-display string. Added to D5 allowlist as intentional preservation.                                                                                                                              | PRD §O5; renaming changes public skill IDs (out of scope).                         |
| D-Q3 | `~/.rc` migration             | **No migration.** Fresh `~/.rc`; `.rc/` runtime dirs excluded from the copy (working data). Auto-migration → Proposal P2.                                                                                                                                                                                       | PRD §O3, §O6.                                                                      |
| D-Q4 | MIT attribution allowlist     | LICENSE has no `rc` token; attribution is `NauckGroup LTDA` (preserved verbatim). Allowlist = `cy-*` IDs + an explicit "forked from rc (github.com/rc/rc)" credit line we **add** to README/NOTICE (the only intentionally-retained `rc` strings).                                                              | PRD §N6, AC5; CON2.                                                                |
| D-Q5 | Copy exclusion set            | Locked in §2.1.                                                                                                                                                                                                                                                                                                 | PRD §F1.1.                                                                         |
| D-Q6 | Distribution publishing       | **Rename in config only**; no new publish pipelines.                                                                                                                                                                                                                                                            | PRD §O8, OQ6 default.                                                              |

The skill directory `skills/rc/` (and `.agents/skills/rc/`) is a brand-name dir, **not** a `cy-*` stable ID → it **is** renamed to `skills/rc/` (see §5.3). The `cy-*` dirs are kept.

## 2. Phase 1 — Faithful fork (copy / scaffold)

### 2.1 Copy set & exclusions (D-Q5)

Copy the entire `~/dev/rc` tree into `~/dev/rc-harness-cli` **except** (rsync-style excludes, derived from `.gitignore` + working-data dirs):

```
.git/                              bin/                  dist/
web/dist/*  (keep web/dist/.keep)  node_modules/         **/node_modules/
.turbo/  .cache/  .tmp/            coverage/ coverage.out coverage.html
*.tsbuildinfo                      web/src/routeTree.gen.ts
skills-lock.json                   skills/*/autoresearch-*/
.DS_Store  .env
# source working/agent data (NOT shippable product state):
.codex/  .claude/  .rc/  .factory/  .resources/  .release-notes/
ai-docs/  RELEASE_NOTES.md  RELEASE_BODY.md  CHANGELOG.md(*)
```

(\*) `CHANGELOG.md` is copied but **truncated/reset** — it is upstream rc history (PRD §N4 lockfiles-are-regenerated spirit; carrying it would create thousands of `rc` residuals from history). Replace with a single "Initial rc fork" entry. `cliff.toml` config is kept + rebranded.

`.agents/` is a real directory (skills mirror, not a symlink per STACK.md) → copy it.

### 2.2 Repo init

- `git init` in target (no upstream history). Stage nothing as part of planning; the Developer commits per workflow rules.
- Keep `LICENSE` (MIT) verbatim; add a `NOTICE`/README credit line (§7).

### 2.3 Trade-off

Copying then mass-renaming (vs. `git filter-repo`) is chosen because the target needs no upstream history and the rename touches non-git-tracked generated paths too. Lockfiles/generated files are regenerated post-rename (§6), never hand-edited (PRD §N4).

## 3. Phase 2 — Identity rename: data shapes & contracts

### 3.1 Canonical token map (the single substitution table)

Order matters: apply most-specific first to avoid partial overlaps.

| From                      | To                                     | Scope                                                     |
| ------------------------- | -------------------------------------- | --------------------------------------------------------- |
| `github.com/rc/rc/pkg/rc` | `github.com/rc-tech/rc-harness/pkg/rc` | Go import paths                                           |
| `github.com/rc/rc`        | `github.com/rc-tech/rc-harness`        | Go module + all imports, URLs                             |
| `cmd/rc`                  | `cmd/rc`                               | path + any string refs (goreleaser `main`)                |
| `pkg/rc`                  | `pkg/rc`                               | path                                                      |
| `@rc/ui`                  | `@rc-tech/ui`                          | bun workspace pkg name + shadcn aliases + imports         |
| `@rc/cli`                 | `@rc-tech/cli`                         | goreleaser npm scope                                      |
| `homebrew-rc`             | `homebrew-rc`                          | goreleaser brew tap                                       |
| `rc.com`                  | `rc.tech`                              | all domains/emails (`support@rc.com` → `support@rc.tech`) |
| `package rc` / `rc.go`    | `package rc` / `rc.go`                 | public Go pkg + filename                                  |
| `.rc` (DirName)           | `.rc`                                  | `internal/config/home.go` + every `~/.rc` ref             |
| `rc`                      | `rc`                                   | display strings (`rc // SETUP` → `rc // SETUP`, etc.)     |
| `rc`                      | `rc`                                   | prose / display                                           |
| `rc` (bare, lowercase)    | `rc`                                   | identifiers, package.json names, paths, strings           |

**Preserved (NOT substituted) — D5 allowlist:** `cy-*` skill/extension IDs; `NauckGroup LTDA` (LICENSE); the single added "forked from rc" credit line(s); the regenerated lockfiles (excluded from scan anyway).

### 3.2 Go backend (PRD F2.1–F2.4)

- `go.mod`: `module github.com/rc-tech/rc-harness`; rebrand the leading comment.
- Mass import rewrite via `go mod edit -module` semantics is insufficient (it only rewrites the module line); use a path-aware rewrite (e.g. `gofmt -r` is too narrow for strings) → **strategy: textual replace of the two import-path tokens across all `.go` files, then `goimports`/`gofmt` + `go build` to prove correctness.** Order: pkg path first, then module root.
- `rc.go` → `rc.go`, `package rc` → `package rc`. Public re-exported type aliases (`type Mode = core.Mode`, etc.) are name-preserving — **the public API surface is identical** (PRD F2.3, AC7); only the package name and import path change. `test/public_api_test.go` is updated to import `pkg/rc` / `package rc` and assert the same symbols.
- `pkg/rc/{events,runs}` → `pkg/rc/{events,runs}`; internal imports updated.
- `cmd/rc/` → `cmd/rc/`; `main.go` import of `internal/version` path updated; goreleaser `main: ./cmd/rc`, `binary: rc`, ldflags `-X github.com/rc-tech/rc-harness/internal/version.{Version,Commit,Date}`.
- `internal/config/home.go`: `DirName = ".rc"`. The sub-dir name constants (`agents`, `state`, `daemon`, `db`, `runs`, `logs`, `cache`, `extensions`) are **not brand-bearing → unchanged** (config-key stability, PRD N1/AC7). Update error strings like `"create rc directory"` → `"create rc directory"`. All `~/.rc` literals in code/tests/fixtures → `~/.rc`.

### 3.3 Web / TS (PRD F2.6)

- Root `package.json` `"name": "rc"` → `"rc-harness"`; keep `packageManager: bun@1.3.11` (PRD N4).
- `web/package.json` `"rc-web"` → `"rc-web"`.
- `packages/ui/package.json` `"@rc/ui"` → `"@rc-tech/ui"`; all `@rc/ui` / `@rc/ui/utils` imports across `web/` + `packages/ui/` updated; `web/components.json` aliases updated to `@rc-tech/ui`.
- `web/tsconfig*`/`tsconfig.base.json` path aliases referencing `@rc/ui` updated.
- Any `rc` in TS/TSX strings/identifiers → `rc` per token map.

### 3.4 Release / distribution (PRD F2.7)

`.goreleaser.yml`: `project_name`, build `id`, `binary`, `main`, ldflag paths, `homepage`, `description`, `maintainers` email/domain, brew tap repo, npm scope. `.goreleaser.release-{header,footer}.md.tmpl` rebranded. `aur-pkg/` strings + any pkgname. **No new publishing wired** (D-Q6).

### 3.5 Tooling / CI / docs (PRD F2.8)

Rebrand textual `rc` in: `Makefile` (targets `check-go-version`/`check-bun-version` keep behavior; only display/paths), `.github/` workflows, `.husky/{commit-msg,pre-commit}`, `.pre-commit-config.yaml`, `.commitlintrc.yaml`, `cliff.toml`, `.coderabbit.yaml`, `.editorconfig` (if any), `scripts/` (incl. `scripts/link-skills.sh`), `README.md`, `AGENTS.md`, `CLAUDE.md`, `CONTRIBUTING.md`, badges/shields URLs, `docs/`, `openapi/`, `sdk/`, `extensions/`, `zsh/`, `skills/`.

## 4. Phase 2 — Visual rebrand: data shapes

### 4.1 rc palette (single source, from REQUEST Visual Identity)

| Role            | rc value              | Replaces (green) |
| --------------- | --------------------- | ---------------- |
| Primary orange  | `#F26B21`             | —                |
| Orange (alt)    | `#F37021`             | —                |
| Amber/gold      | `#FBB034`             | —                |
| Gold (alt)      | `#FDB813`             | —                |
| Text near-black | `#1C1C1C`             | —                |
| Light bg        | `#FFFFFF` / `#FAFAFA` | —                |

CLI `internal/charmtheme/theme.go` mapping (keep variable names + dark neutral bg/semantic colors; swap only brand hues):

- `ProgressGradientStart "#65A30D"` → `"#F26B21"`; `ProgressGradientEnd "#CAEA28"` → `"#FBB034"` (orange→amber gradient).
- `ColorBrand #CAEA28` → `#F26B21`; `ColorAccent #A3E635` → `#FBB034`; `ColorAccentAlt #84CC16` → `#FDB813`; `ColorAccentDeep #65A30D` → `#F37021`.
- **Unchanged** (not brand-bearing, PRD N1): `ColorBg*`, `ColorSuccess/Error/Warning/Info`, `ColorFgBright/Muted/Dim`, `ColorBorder`. `ColorBorderFocus = ColorAccent` keeps the indirection (now resolves to amber).
- Any other green-hex source: `internal/core/run/ui/view_test.go` assertions updated to the new tokens.

### 4.2 Web tokens (PRD F3.2) — Tailwind v4 CSS-first

- `packages/ui/src/tokens.css`: `--brand: #d6f24a;` → `--brand: #f26b21;`. Any other green-derived brand token (e.g. brand-foreground/ring) shifted to the orange/amber scale, **light + dark blocks both updated**. shadcn `baseColor: neutral`, `iconLibrary: lucide` **kept** (PRD N2).
- `web/src/styles.css`: radial-gradient `rgba(214,242,74,...)` (green `#d6f24a`) → `rgba(242,107,33,...)` (`#f26b21`), preserving alpha values.

### 4.3 Logo / assets (PRD F3.3)

- Create rc SVG: rounded peak/triangle outline mark, orange→gold gradient (`#F26B21`→`#FDB813`), lowercase bold "rc" wordmark in `#1C1C1C`. This is the canonical source asset.
- Replace where rc logo is referenced: `imgs/` (regenerate/replace `screenshot.png` hero usage in README; flow diagrams' green hexes swapped), `web/public/symbol.png` (→ rc mark), favicons, README hero `<img>`.
- `docs/design/daemon-mockup/assets/rc-icon.png` → `rc-icon.png` (+ green hexes in `colors_and_type.css` / `src/*.jsx` → rc palette).
- **Trade-off:** raster assets (`screenshot.png`) that embed live CLI output cannot be perfectly reproduced without running the rebranded CLI; the Developer regenerates them from the rebuilt binary where feasible, otherwise replaces brand chrome only. Flagged as a build risk (§9).

### 4.4 Welcome header (PRD F3.4, D-Q1)

`internal/cli/setup.go::printWelcomeHeader` — re-skin only:

- Border: rounded single-line (`lipgloss.RoundedBorder()`) in rc orange (`ColorBrand`) via `newCLIChromeStyles()`.
- Title: `"rc // SETUP"` (mechanical `rc`→`rc`); optionally append version: `fmt.Sprintf("rc %s // SETUP", versionLabel())`. **Decision:** keep the mechanical `rc // SETUP` title to honor AC7 (no new data/behavior); the `rc vX.Y.Z` brand label is satisfied by the box title + the existing `rc --version`. `versionLabel` uses existing `internal/version.Version` only.
- `internal/cli/theme.go` L79 `"rc // INTERACTIVE INPUT"` → `"rc // INTERACTIVE INPUT"`; box style border/accent → orange via the theme tokens (already inherited from §4.1).
- No `Welcome back <user>` / mascot / model·plan·org·cwd line (→ Proposal P1).

## 5. Identity-bearing paths to rename (PRD F2.9, F2.10)

### 5.1 Directories / files (git-mv style; keep `go:embed` valid)

- `cmd/rc/` → `cmd/rc/`
- `rc.go` → `rc.go`
- `pkg/rc/` → `pkg/rc/`
- `zsh/rc-completion/rc-completion.plugin.zsh` → `zsh/rc-completion/rc-completion.plugin.zsh` (+ `zsh/rc-completion/README.md`)
- `openapi/rc-daemon.json` → `openapi/rc-daemon.json` (update codegen script refs)
- `web/src/generated/rc-openapi.d.ts` → `rc-openapi.d.ts` (regenerated by codegen; rename the codegen output target)
- `docs/design/daemon-mockup/assets/rc-icon.png` → `rc-icon.png`
- `skills/rc/` → `skills/rc/` and `.agents/skills/rc/` → `.agents/skills/rc/`

### 5.2 `go:embed` integrity

After any path rename under an embed root, the embed directives must still resolve. Embed roots per STACK.md: `web/embed.go`, `skills/embed.go`, `agents/embed.go`, `internal/core/extension/discovery_bundled.go`, `internal/core/prompt/templates.go`. The skill dir rename (§5.3) requires updating `skills/embed.go` patterns + any code referencing the `rc` skill ID. Verified by `go build` + `skills_bundle_test.go`.

### 5.3 Skill-ID consequence

Renaming `skills/rc/` → `skills/rc/` changes the skill ID `rc` → `rc` in `skills-lock.json` (regenerated, §6) and in `skills/embed.go` / discovery code / `test/skills_bundle_test.go`. The `cy-*` skill/extension dirs are **kept** (D-Q2). If any code hardcodes the literal skill name `"rc"`, it becomes `"rc"`.

## 6. Lockfiles & generated artifacts — regenerate, never hand-edit (PRD N4)

1. `go.sum` ← `go mod tidy` (after module rename).
2. `bun.lock` ← `bun install` (after package renames).
3. `web/src/routeTree.gen.ts` ← TanStack router generator (build/dev).
4. `web/src/generated/rc-openapi.d.ts` ← `openapi-typescript` codegen from renamed `openapi/rc-daemon.json`.
5. `skills-lock.json` ← skills generator / `scripts/link-skills.sh` (`postinstall`).

Order: Go rename → `go mod tidy` → `go build ./...`; then web renames → `bun install` → codegen → `bun run build`.

## 7. License / attribution (PRD N6, CON2)

- `LICENSE` unchanged (MIT, NauckGroup LTDA).
- Add to `README.md` (and optionally a `NOTICE`) one credit line: _"rc is a fork of rc (https://github.com/rc/rc), MIT-licensed, © NauckGroup LTDA."_ This is the **only** intentionally-retained `rc` string outside `cy-*` IDs and is the documented AC5 allowlist entry.

## 8. Test approach (maps to PRD ACs; conform to source frameworks)

- **Go** (`go test ./...`): faithful pass. Brand-intent test updates only — `internal/charmtheme/theme_test.go` (new hexes), `internal/cli/theme_test.go` (`rc // SETUP` / `rc // INTERACTIVE INPUT`), `internal/core/run/ui/view_test.go` (palette), `test/public_api_test.go` (new import path, same symbols), `test/skills_bundle_test.go` (rc skill ID), `test/release_config_test.go` (goreleaser identity), `cmd/rc/main_test.go` (binary/domain), config/home tests (`.rc`). → **AC1, AC4, AC7.**
- **Web unit** (vitest): `test/frontend-workspace-config.test.ts`, `test/codegen-script.test.ts`, `test/frontend-verification-contract.test.ts`, `packages/ui` + `web` suites — update identity assertions (`@rc-tech/ui`, rc package names, renamed openapi file). → **AC2.**
- **Web e2e** (playwright): run to upstream baseline; update any selector/text asserting rc brand. → **AC2.**
- **Aggregate**: `make verify` (`frontend-verify fmt lint test go-build frontend-e2e`) green; `golangci-lint` clean to baseline; husky + pre-commit hooks operate under the new name. → **AC3.**
- **Residual scan (AC5)**: `grep -ri rc` over the fork excluding `.git/`, `node_modules/`, regenerated lockfiles (`go.sum`, `bun.lock`, `routeTree.gen.ts`, `*-openapi.d.ts`, `skills-lock.json`) returns **only**: `cy-*` IDs and the §7 credit line. The Developer enumerates the exact scan output and confirms it equals the allowlist. → **AC5.**
- **Smoke (AC8)**: build `rc`; run it; confirm rc welcome header (orange rounded box, `rc // SETUP`), `~/.rc` created/read, and a core idea→code flow equivalent to rc.
- **Palette check (AC6)**: assert no green brand hexes (`#CAEA28 #A3E635 #84CC16 #65A30D #d6f24a` / `rgba(214,242,74`) remain in `theme.go`, `tokens.css`, `styles.css`, or other brand sources.

## 9. Risks / trade-offs for the orchestrator

- **R1 (build env):** `go test ./...` + playwright e2e require a working toolchain (Go 1.26.1, bun 1.3.11, browsers). If unavailable, AC1/AC2/AC8 cannot be executed — flag to user.
- **R2 (raster logo assets):** `screenshot.png` / embedded-output PNGs may not be perfectly regenerable without running the rebuilt CLI; brand-chrome replacement is the fallback (§4.3).
- **R3 (residual scan noise):** fixtures/testdata under `internal/core/.../testdata` contain many `rc` strings (module paths in mock extensions) — these **must** be renamed (they compile/build) and are not allowlisted. High volume (~438 Go files) but mechanical.
- **R4 (skill-ID rename):** renaming `skills/rc/`→`skills/rc/` is a public-ID change; it is in scope (brand dir, not a `cy-*` stable ID) but breaks any external reference to the `rc` skill. Confirm acceptable.
- **R5 (CHANGELOG reset):** truncating upstream `CHANGELOG.md` (§2.1) is a judgment call to avoid thousands of historical `rc` residuals — confirm with user vs. keeping + allowlisting history.

## 10. Out of scope (carried from PRD) / Proposals

No Proposal (P1 richer header, P2 migration, P3 `cy-*` rename, P4 telemetry, P5 publishing, P6 features) is implemented in TASKS. They remain in PRD §Proposals pending sign-off.
