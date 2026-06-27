# TASKS — Fork & Rebrand rc → rc

> The steps. Reads: SPEC.md (design + token map), PRD.md (ACs), STACK.md (source stack).
> Smallest ORDERED list that fully delivers the SPEC. Source: `~/dev/rc`. Target: `~/dev/rc-harness-cli`.
> Token map (SPEC §3.1) and allowlist (SPEC §3.1 / §7) are authoritative for every rename task.
> Each task: files it touches + a verifiable done-condition tied to a PRD AC. Tasks are sequential unless marked parallel-safe.
> Hard ordering constraint (SPEC §6): all Go path/string renames → `go mod tidy` → `go build`; then web renames → `bun install` → codegen → build. Lockfiles are regenerated, never hand-edited.

---

## Phase 1 — Faithful fork (copy / scaffold) → AC1, F1.x

### T1. Copy the rc tree into the target with the locked exclude set

- **Touches:** entire `~/dev/rc-harness-cli/` (new files); reads `~/dev/rc/`.
- **Do:** rsync `~/dev/rc/` → `~/dev/rc-harness-cli/` applying the SPEC §2.1 exclude set exactly: `.git/ bin/ dist/ web/dist/* (keep web/dist/.keep) node_modules/ **/node_modules/ .turbo/ .cache/ .tmp/ coverage/ coverage.out coverage.html *.tsbuildinfo web/src/routeTree.gen.ts skills-lock.json skills/*/autoresearch-*/ .DS_Store .env .codex/ .claude/ .rc/ .factory/ .resources/ .release-notes/ ai-docs/ RELEASE_NOTES.md RELEASE_BODY.md`. Copy `.agents/` (real dir). Copy `CHANGELOG.md` for now (reset in T20).
- **Done:** target tree mirrors source minus excludes; `web/dist/.keep` present; none of the excluded paths exist in target. `diff -rq` (excludes filtered) shows only excluded paths missing. (F1.1)

### T2. Initialize the target repo

- **Touches:** `~/dev/rc-harness-cli/.git/`.
- **Do:** `git init` in target. No upstream history. Stage/commit per workflow rules later (Developer).
- **Done:** `git status` works in target; no rc commit history present. (F1.2)

### T3. Verify the unmodified fork builds (pre-rename baseline)

- **Touches:** none (read/build only).
- **Do:** in target, `go build ./...` (with the still-`rc` module — it must still compile because all paths are internally consistent post-copy). Capture baseline. NOTE: module path is still `github.com/rc/rc`; this proves the copy is complete and self-consistent before renames.
- **Done:** `go build ./...` succeeds on the unmodified copy. (F1.1 / pre-AC1 gate)

---

## Phase 2 — Go backend identity rename → AC1, AC4, AC7, F2.1–F2.5

> Order within Go: pkg path before module root (most-specific first, SPEC §3.1). Path moves before string rewrites where embeds are involved.

### T4. Rename Go identity-bearing paths (git-mv)

- **Touches:** `cmd/rc/` → `cmd/rc/`; `rc.go` → `rc.go`; `pkg/rc/` → `pkg/rc/` (incl. `pkg/rc/{events,runs}` → `pkg/rc/{events,runs}`).
- **Do:** move directories/files only; do not edit contents yet.
- **Done:** paths exist at new locations; old paths gone. (F2.2, F2.3, F2.4, F2.9)

### T5. Rewrite Go import paths and module-scoped tokens across all `.go` files

- **Touches:** all `*.go` in target; `go.mod`.
- **Do:** textual replace in strict order (SPEC §3.1): `github.com/rc/rc/pkg/rc` → `github.com/rc-tech/rc-harness/pkg/rc`, then `github.com/rc/rc` → `github.com/rc-tech/rc-harness`. Set `go.mod` `module github.com/rc-tech/rc-harness` + rebrand its leading comment. Then `gofmt`/`goimports`. Do NOT touch `cy-*` IDs or `NauckGroup LTDA`.
- **Done:** no `github.com/rc/rc` substring remains in any `.go` or `go.mod`; `gofmt -l` clean. (F2.1, AC4)

### T6. Rename the public package declaration and its API test

- **Touches:** `rc.go` (`package rc` → `package rc`), every file in `pkg/rc/` with `package rc`, `test/public_api_test.go`.
- **Do:** rewrite `package rc` → `package rc`; update `test/public_api_test.go` to import `pkg/rc` / `package rc` while asserting the SAME exported symbols (type aliases like `type Mode = core.Mode` stay name-preserving — public surface identical per SPEC §3.2).
- **Done:** package decls are `rc`; `public_api_test.go` references the new import and the symbol set is unchanged from source. (F2.3, F2.4, AC7)

### T7. Rename the config directory brand + error strings

- **Touches:** `internal/config/home.go` (`DirName = ".rc"`), all code/test/fixture literals `~/.rc` → `~/.rc`, error strings `"create rc directory"` → `"create rc directory"`, and the config/home test(s).
- **Do:** change `DirName` and every `.rc` brand literal. Keep sub-dir constants (`agents state daemon db runs logs cache extensions`) UNCHANGED (not brand-bearing, config-key stability per SPEC §3.2).
- **Done:** `DirName == ".rc"`; no `~/.rc` / `.rc` literal remains in `.go`; home test asserts `.rc`. (F2.5, AC4, AC7)

### T8. Rewrite remaining bare `rc`/`rc`/`rc` tokens in Go (non-display logic)

- **Touches:** all remaining `*.go` strings/identifiers/filenames not covered above, including `internal/core/.../testdata` mock-extension module paths (SPEC R3, ~438 files mechanical), `cmd/rc/main*.go`, skill-ID literals (deferred display strings handled in T10/T15), `rc.com` → `rc.tech`, `support@rc.com` → `support@rc.tech`.
- **Do:** apply token map (`rc`→`rc`, `rc`→`rc`, `rc`→`rc`, `rc.com`→`rc.tech`) to all non-allowlisted Go occurrences. Skip `cy-*`, `NauckGroup LTDA`.
- **Done:** `grep -ri rc` over `*.go` returns only `cy-*` IDs (no other matches). (F2.8, AC5 partial)

### T9. Update goreleaser & version ldflag wiring for Go build

- **Touches:** `.goreleaser.yml`, `.goreleaser.release-header.md.tmpl`, `.goreleaser.release-footer.md.tmpl`, `aur-pkg/*`.
- **Do:** set `project_name`, build `id`, `binary: rc`, `main: ./cmd/rc`, ldflags `-X github.com/rc-tech/rc-harness/internal/version.{Version,Commit,Date}`, `homepage`, `description: "rc CLI"`, `maintainers` (email/domain → rc.tech), brew tap `homebrew-rc` → `homebrew-rc`, npm scope `@rc/cli` → `@rc-tech/cli`. AUR pkgname/strings. No new publishing wired (D-Q6).
- **Done:** goreleaser config and AUR contain no `rc`; ldflag path points at the new module. (F2.7)

### T10. Build the rich rc welcome header (USER DECISION — net-new feature, in-scope; overrides Proposal P1 / Open Question 1)

- **Touches:** `internal/cli/setup.go` (`printWelcomeHeader`, ~L490), `internal/cli/theme.go` (~L79); add a small mascot/peak ASCII art constant.
- **USER DECISION:** implement the rich header like reference image 2 (Claude Code style), NOT just a re-skin. This is a deliberate net-new feature; AC7 "zero functional change" carries a documented exception for the welcome header only.
- **Do:** rounded orange box (`lipgloss.RoundedBorder()` in rc orange via existing theme tokens) with: top-left label `rc vX.Y.Z` (from `internal/version.Version`); centered `Welcome back <user>!`; centered ASCII art of the rc peak/triangle mark (orange→gold if the renderer supports gradient, else solid rc orange); and a dimmed bottom line of `model · plan · org · cwd`. SOURCE THE DATA from what the CLI already has: `cwd` from `os.Getwd`; `<user>` from OS user / git config (fallback to a generic greeting if unavailable); `model`/`provider` from the resolved rc/rc config if present; `plan`/`org` ONLY if a real data source exists — otherwise OMIT those fields (do NOT fabricate). Decide field availability during implementation and document which fields were sourced vs omitted.
- **Done:** header renders a rounded orange box with `rc vX.Y.Z`, `Welcome back <user>!`, the peak mascot, and the dimmed context line (only for fields with real data sources). Also relabel the secondary `"rc // INTERACTIVE INPUT"` → `"rc // INTERACTIVE INPUT"`. (F3.4, AC6, AC7-exception)

### T11. Swap CLI brand color tokens

- **Touches:** `internal/charmtheme/theme.go`.
- **Do:** per SPEC §4.1: `ProgressGradientStart #65A30D`→`#F26B21`; `ProgressGradientEnd #CAEA28`→`#FBB034`; `ColorBrand #CAEA28`→`#F26B21`; `ColorAccent #A3E635`→`#FBB034`; `ColorAccentAlt #84CC16`→`#FDB813`; `ColorAccentDeep #65A30D`→`#F37021`. Keep variable names, `ColorBg*`, semantic colors, `ColorBorder`, and the `ColorBorderFocus = ColorAccent` indirection unchanged.
- **Done:** none of `#CAEA28 #A3E635 #84CC16 #65A30D` remain in `theme.go`; brand tokens hold the orange/amber values. (F3.1, AC6)

### T12. Update Go brand-intent tests

- **Touches:** `internal/charmtheme/theme_test.go`, `internal/cli/theme_test.go`, `internal/core/run/ui/view_test.go`, `test/release_config_test.go`, `cmd/rc/main_test.go`.
- **Do:** update assertions ONLY for brand intent: new hexes (`theme_test`, `view_test`), `rc // SETUP` + `rc // INTERACTIVE INPUT` (`cli/theme_test` L64/L84), goreleaser identity (`release_config_test`), binary/domain (`main_test`). No behavioral test logic changes.
- **Done:** these tests assert rc identity/palette; only brand assertions changed. (AC6, AC7)

### T13. Rename Go-adjacent identity paths under embed roots

- **Touches:** `openapi/rc-daemon.json` → `openapi/rc-daemon.json`; `zsh/rc-completion/rc-completion.plugin.zsh` → `zsh/rc-completion/rc-completion.plugin.zsh` (+ that dir's `README.md`); any code/script referencing these. Embed roots per SPEC §5.2.
- **Do:** git-mv the paths; update referencing code/scripts (codegen target for openapi handled in T18).
- **Done:** renamed files exist; no `rc` in their names. (F2.9)

### T14. Regenerate `go.sum` and prove the Go build/tests pass

- **Touches:** `go.sum` (regenerated), build only.
- **Do:** `go mod tidy` → `go build ./...` → `go test ./...` → `golangci-lint run`.
- **Done:** `go build ./...` succeeds; `go test ./...` passes; `golangci-lint` clean to upstream baseline. (AC1, AC3, AC7) — **GATE: Go phase complete before web phase.**

---

## Phase 3 — Skill directory rename → F2.9, AC5, AC7

### T15. Rename the `rc` skill directory and its embeds/IDs

- **Touches:** `skills/rc/` → `skills/rc/`; `.agents/skills/rc/` → `.agents/skills/rc/`; `skills/embed.go` patterns; discovery code referencing skill ID `"rc"`; `test/skills_bundle_test.go`.
- **Do:** git-mv the skill dir + mirror; update embed patterns and any hardcoded `"rc"` skill-ID literal → `"rc"`. KEEP all `cy-*` skill/extension dirs (D-Q2). `skills-lock.json` is regenerated later (T19), not hand-edited.
- **Done:** `go build ./...` still succeeds (embeds resolve); `skills_bundle_test.go` asserts the `rc` skill ID; `cy-*` dirs untouched. (F2.9, AC7, R4)

---

## Phase 4 — Web / TS identity rename → AC2, F2.6, F3.2

### T16. Rename web package names and workspace scope

- **Touches:** root `package.json` (`"name": "rc"`→`"rc-harness"`, keep `packageManager: bun@1.3.11`), `web/package.json` (`rc-web`→`rc-web`), `packages/ui/package.json` (`@rc/ui`→`@rc-tech/ui`), `web/components.json` (aliases→`@rc-tech/ui`), `tsconfig.base.json` / `web/tsconfig*` path aliases.
- **Do:** rename package names and `@rc/ui` aliases per token map; keep `bun@1.3.11`, shadcn `baseColor: neutral`, `iconLibrary: lucide`.
- **Done:** no `@rc/ui` or `rc`/`rc-web` name in any `package.json`/`components.json`/`tsconfig*`. (F2.6, N4)

### T17. Rewrite `@rc/ui` imports and remaining TS/TSX brand tokens

- **Touches:** all `*.ts`/`*.tsx` under `web/` + `packages/ui/` importing `@rc/ui` or `@rc/ui/utils`; any `rc`/`rc`/`rc` strings/identifiers in TS/TSX; `rc.com`→`rc.tech`.
- **Do:** apply token map to imports + strings; skip `cy-*` and the §7 credit line.
- **Done:** `grep -ri rc` over `*.ts`/`*.tsx` returns only `cy-*` IDs. (F2.6, AC5 partial)

### T18. Rename the generated-openapi codegen target

- **Touches:** `web/src/generated/rc-openapi.d.ts` target name → `rc-openapi.d.ts`; the codegen script (`scripts/` / `test/codegen-script.test.ts` expectations) that points at `openapi/rc-daemon.json` (renamed in T13).
- **Do:** update the codegen script's input (`rc-daemon.json`) and output (`rc-openapi.d.ts`) names; the `.d.ts` itself is regenerated (T19), not hand-edited.
- **Done:** codegen config references the rc-named input/output; no `rc-openapi` reference remains in scripts/config. (F2.9)

### T19. Swap web brand color tokens

- **Touches:** `packages/ui/src/tokens.css`, `web/src/styles.css`.
- **Do:** `--brand: #d6f24a;`→`--brand: #f26b21;` (both light + dark blocks); shift any other green-derived brand token (brand-foreground/ring) to the orange/amber scale; `web/src/styles.css` radial-gradient `rgba(214,242,74,...)`→`rgba(242,107,33,...)` preserving alpha. Keep `baseColor: neutral`, lucide.
- **Done:** no `#d6f24a` / `rgba(214,242,74` remains in `tokens.css`/`styles.css`; orange brand tokens present in light+dark. (F3.2, AC6)

### T20. Regenerate web lockfile + generated artifacts and prove web build/tests pass

- **Touches:** `bun.lock`, `web/src/routeTree.gen.ts`, `web/src/generated/rc-openapi.d.ts`, `skills-lock.json` (all regenerated).
- **Do (SPEC §6 order):** `bun install` → openapi codegen (from `openapi/rc-daemon.json`) → TanStack route-tree generation → skills generator / `scripts/link-skills.sh` (postinstall) → `bun run build`. Then `vitest` unit suites (`test/frontend-workspace-config.test.ts`, `test/codegen-script.test.ts`, `test/frontend-verification-contract.test.ts`, `packages/ui` + `web`) and `playwright` e2e; update any selector/text asserting rc brand.
- **Done:** `bun install` + Vite build succeed; vitest passes; playwright passes to upstream baseline; regenerated `skills-lock.json` contains skill ID `rc` not `rc`. (AC2)

---

## Phase 5 — Tooling, CI, docs, assets, attribution → AC3, AC5, AC6, F2.8, F3.3, N6

### T21. Rebrand tooling / CI / hooks / changelog config

- **Touches:** `Makefile`, `.github/` workflows, `.husky/{commit-msg,pre-commit}`, `.pre-commit-config.yaml`, `.commitlintrc.yaml`, `cliff.toml`, `.coderabbit.yaml`, `.editorconfig` (if present), `scripts/*` (incl. `scripts/link-skills.sh`).
- **Do:** apply token map to textual `rc` only; keep target behavior (`check-go-version`/`check-bun-version` logic unchanged — display/paths only). Keep `cliff.toml` config, rebranded.
- **Done:** no non-allowlisted `rc` in these files; husky + pre-commit hooks run under the new name. (F2.8, AC3)

### T22. Rebrand docs / prose / badges / sdk / extensions / design mockup

- **Touches:** `README.md`, `AGENTS.md`, `CLAUDE.md`, `CONTRIBUTING.md`, badge/shield URLs, `docs/`, `openapi/` prose, `sdk/`, `extensions/` (non-`cy-*` strings), `docs/design/daemon-mockup/` (`colors_and_type.css` green hexes + `src/*.jsx` → rc palette).
- **Do:** apply token map to prose/URLs; swap design-mockup green hexes to rc palette; rename `docs/design/daemon-mockup/assets/rc-icon.png` → `rc-icon.png` and update references.
- **Done:** no non-allowlisted `rc` in docs/sdk/extensions; design mockup uses rc palette + renamed icon. (F2.8, F3.3)

### T23. Create the rc logo and replace rc brand assets

- **Touches:** new canonical rc SVG; `imgs/` (hero/`screenshot.png`, flow-diagram green hexes), `web/public/symbol.png`, favicons, README hero `<img>`.
- **Do:** create rounded peak/triangle mark, orange→gold gradient (`#F26B21`→`#FDB813`), lowercase bold "rc" wordmark in `#1C1C1C` (SPEC §4.3). Replace rc logo references; swap green hexes in flow diagrams. For raster assets embedding live CLI output (`screenshot.png`), regenerate from the rebuilt binary where feasible, else replace brand chrome only (R2).
- **Done:** rc logo present and referenced wherever rc's was; no green brand hexes in flow-diagram/asset sources. (F3.3, AC6, R2)

### T24. Rewrite LICENSE to rc; reset CHANGELOG — ZERO rc residuals (USER DECISION, overrides SPEC §7/N6)

- **Touches:** `README.md`, `CHANGELOG.md`, `LICENSE`.
- **USER DECISION (overrides the original MIT-attribution plan):** zero traces of `rc` ANYWHERE, including LICENSE. The user explicitly accepts the MIT-compliance risk of removing the upstream copyright/attribution. NO "forked from rc" credit line is added.
- **Do:** rewrite `LICENSE` to `Copyright (c) 2026 rc` (keep the MIT license body text, replace the copyright holder line; remove `NauckGroup LTDA` / any rc reference). Add NO upstream credit line in README/NOTICE. Reset `CHANGELOG.md` to a single "Initial rc release" entry.
- **Done:** `LICENSE` reads `Copyright (c) 2026 rc` with no rc/NauckGroup string; no credit line anywhere; CHANGELOG reset. After this, the ONLY allowed residuals are the `cy-*` public IDs (which do not contain the substring `rc`). (USER DECISION, R5)

---

## Phase 6 — Verification & sign-off → AC1–AC8

### T25. AC5 residual scan against the enumerated allowlist

- **Touches:** none (scan only).
- **Do:** `grep -ri rc` over the fork excluding `.git/`, `node_modules/`, and regenerated lockfiles (`go.sum`, `bun.lock`, `routeTree.gen.ts`, `*-openapi.d.ts`, `skills-lock.json`). Per USER DECISION (T24) there is NO credit-line exception and NO LICENSE exception — the scan must return **zero** matches. `cy-*` IDs are fine because they do not contain the substring `rc`.
- **Done:** `grep -ri rc` returns ZERO matches across the fork (LICENSE included). (AC5)

### T26. AC6 palette / no-green-hex assertion

- **Touches:** none (scan only).
- **Do:** assert none of `#CAEA28 #A3E635 #84CC16 #65A30D #d6f24a` / `rgba(214,242,74` remain in `internal/charmtheme/theme.go`, `packages/ui/src/tokens.css`, `web/src/styles.css`, or other brand color sources.
- **Done:** zero matches for green brand hexes in brand sources. (AC6)

### T27. AC3 aggregate `make verify`

- **Touches:** none (run only).
- **Do:** `make verify` (`frontend-verify fmt lint test go-build frontend-e2e`).
- **Done:** `make verify` is green end-to-end. (AC3)

### T28. AC8 smoke test of the built binary

- **Touches:** none (run only).
- **Do:** build `rc`; run it; confirm the rc welcome header (orange rounded box, `rc // SETUP`); confirm `~/.rc` is created/read (not `~/.rc`); exercise a core idea→code flow equivalent to rc; confirm `rc --version` shows the rc label.
- **Done:** binary is named `rc`, prints the rc header, reads/writes `~/.rc`, completes a core flow. (AC4, AC8)

---

## Notes for the orchestrator

- **GATE order is load-bearing:** T14 (Go green) must complete before Phase 4; T20 (web green) before Phase 6. SPEC §6.
- **Risks carried (SPEC §9):** R1 toolchain availability (Go 1.26.1 / bun 1.3.11 / playwright browsers) gates AC1/AC2/AC8; R2 raster logo regenerability; R4 `rc`→`rc` skill-ID is a public-ID change (confirm acceptable); R5 CHANGELOG reset is a judgment call (confirm vs. keep+allowlist).
- **No Proposal (P1–P6) is implemented here** (AC9). Richer header, `~/.rc` migration, `cy-*` rename, telemetry, publishing pipelines, new features remain parked in PRD §Proposals.
