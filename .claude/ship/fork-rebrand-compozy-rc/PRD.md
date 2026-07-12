# PRD — Fork & Rebrand RC → RC

> The why & what. No implementation detail (that lives in SPEC.md).
> Reads: REQUEST.md (authoritative request), STACK.md (detected source stack).
> Brand-facing copy stays exactly as specified in REQUEST; this doc is in English for tooling consistency.

## Problem

RC-tech needs its own branded CLI product. Rather than build from scratch, the request is to faithfully fork the existing MIT-licensed **RC** project (`~/dev/rc`) — a CLI that orchestrates AI agents "from idea to delivered code" (Go backend + bun/Vite/React/TypeScript web frontend) — into the empty target `~/dev/rc-harness-cli` and rebrand it as **RC**.

The product value is unchanged: the same agent-orchestration capability, now shipped under RC's identity (name, module path, binary, colors, logo, welcome header). The work is a faithful rebrand fork: **preserve 100% of RC's functionality**; change only identity, visual identity, and the CLI welcome header.

## Target users

- **RC-tech developers / end users** who run the `rc` CLI to orchestrate AI agents from idea to delivered code, and who use the accompanying web UI.
- **RC-tech maintainers** who build, test, release, and distribute the forked product under the RC name (goreleaser, brew/npm/AUR, CI).
- (Indirect) **Open-source consumers** who would see RC identity in docs, README, badges, package scopes, and the public API surface.

## Functional requirements

Grouped to match REQUEST sections A–C. All are functionality-preserving except where they explicitly re-skin/relabel brand surfaces.

### F1. Faithful fork (copy / scaffold)

- F1.1 The full RC tree is reproduced in `~/dev/rc-harness-cli`, preserving structure (all source dirs, dotfiles, tooling config), excluding only VCS/build/working cruft (precise exclude set defined in SPEC; see Open Question 5).
- F1.2 Target is initialized as its own repo state (no upstream RC git history required).
- F1.3 LICENSE remains MIT with original-author attribution preserved as required by the MIT license.

### F2. Identity / naming rename (functionality-preserving)

- F2.1 Go module path `github.com/rc/rc` → `github.com/rc-tech/rc-harness`, updated across every `.go` import.
- F2.2 CLI entrypoint `cmd/rc/` → `cmd/rc/`; binary `rc` → `rc`.
- F2.3 Public package `rc.go` / `package rc` renamed consistently (file + package), keeping the public API surface intact.
- F2.4 Public Go subpackage path `pkg/rc/` renamed to the RC-scoped equivalent; public API guarded by `public_api_test.go` stays equivalent.
- F2.5 User config directory `.rc` → `.rc` (`internal/config/home.go DirName` plus every `~/.rc` path/error-string reference in code and tests). Default behavior: fresh `.rc`, **no silent migration** (migration deferred to Proposals; see Open Question 3).
- F2.6 Web identity renamed: root + `web/package.json` `name`; shadcn workspace aliases `@rc/ui*` → RC scope; any `rc` in TS/TSX.
- F2.7 Release/distribution identity renamed: goreleaser `project_name`, build id, `binary`, `main`, ldflag version-package paths, homepage, description, maintainer email/domain; brew tap, npm scope, AUR packaging.
- F2.8 All remaining textual occurrences (`rc`/`RC`/`RC`, URLs `github.com/rc/rc` and `rc.com`, support emails, telemetry/user-agent/analytics identifiers) updated to the RC equivalents (domain `rc.tech`), across docs, README, AGENTS.md, CLAUDE.md, CONTRIBUTING, changelogs, `.github/`, `.husky/`, Makefile, `cliff.toml`, `aur-pkg/`, `zsh/`, `scripts/`, `openapi/`, `sdk/`, `skills/`, `extensions/`, badges.
- F2.9 Identity-bearing **paths and filenames** (not just string contents) renamed: zsh completion plugin, `openapi/rc-daemon.json`, generated `rc-openapi.d.ts`, design-mockup `rc-icon.png`, etc. — keeping all `go:embed` paths valid.
- F2.10 A complete inventory of every place the `rc` identity appears is produced (in SPEC) so the rename is provably complete and auditable.

### F3. Visual rebrand

- F3.1 CLI terminal palette: RC's green brand tokens in `internal/charmtheme/theme.go` (and any other color source) replaced with the RC orange→amber palette and gradient (per Visual Identity in REQUEST). Affected brand-intent tests updated.
- F3.2 Web theme: RC's green brand tokens (`packages/ui/src/tokens.css`, `web/src/styles.css`) replaced with the RC palette; shadcn `baseColor: neutral` and `iconLibrary: lucide` kept; light + dark theme support kept.
- F3.3 Logo asset: an RC logo (orange→gold gradient peak/triangle mark + lowercase "rc" wordmark, per REQUEST) created and used in place of RC's brand assets (`imgs/`, web `public/`, favicons, README hero) wherever the RC logo is referenced.
- F3.4 CLI welcome header re-skinned/relabeled to the RC brand (rounded orange box; `RC vX.Y.Z` label). **Scope of the richer header layout** (`Welcome back <user>!`, ASCII peak mascot, dimmed model · plan · email/org · cwd line) is constrained by AC6 "zero functional change" — see Open Question 1; any net-new data/behavior moves to Proposals.

## Non-functional requirements

- N1. **Zero functional change** vs. RC beyond identity + visuals: same commands, flags, behavior, public APIs, config keys (except the brand dir name).
- N2. **Conformance over taste**: match RC's existing structure, style, and conventions (Go `internal`/`pkg` boundaries, Tailwind v4 CSS-first tokens, oxfmt/oxlint, golangci-lint v2, conventional commits, husky + pre-commit hooks). No refactors, no adjacent "improvements" (CLAUDE.md Rules 2, 3, 8).
- N3. **Surgical changes only**: rename mechanically and completely; change nothing else.
- N4. **Exact-version fidelity**: keep the toolchain/dependency versions recorded in STACK.md (Go 1.26.1, bun 1.3.11, turbo, vite 8, etc.). Lockfiles (`go.sum`, `bun.lock`, generated route tree / openapi types / skills lock) are **regenerated**, not hand-edited.
- N5. **Brand-facing copy** uses "RC" / lowercase "rc" wordmark exactly as specified; domain `rc.tech`; module `github.com/rc-tech/rc-harness`.
- N6. **License compliance**: MIT LICENSE and any legally-required attribution to the original RC authors preserved; preserved `rc`/author strings enumerated as an allowlist (see Open Question 4).
- N7. **Subagents have no access to reference images**: the Visual Identity section in REQUEST is the single source of truth for colors, logo, and header layout.

## Out of scope

- O1. Any new product feature or capability beyond the fork + rebrand.
- O2. Any refactor, restructure, or "improvement" of RC code, even where conventions are debatable (surface, do not change).
- O3. `~/.rc → ~/.rc` auto-migration / back-compat (deferred to Proposals).
- O4. A richer net-new welcome header that introduces data sources the current CLI does not already expose (deferred to Proposals unless resolved otherwise — Open Question 1).
- O5. Renaming the `cy-*` skill/extension prefix if it is a stable public identifier (default: keep; Open Question 2).
- O6. Porting source-repo working data (`.rc/tasks`, `.rc/research`, `.codex/`, `.claude/`, etc.) — these are not shippable product state.
- O7. Telemetry policy changes / opt-in re-review (Proposals).
- O8. CI matrix changes beyond re-skinning the existing workflows to the new name (Proposals).
- O9. Implementing anything in the "Proposals — requires approval" section.

## Acceptance criteria (verifiable by real execution)

- AC1. `~/dev/rc-harness-cli` contains the full faithful fork; `go build ./...` succeeds and `go test ./...` passes.
- AC2. Web `bun install` + Vite build succeed; **vitest** unit tests pass; **playwright** e2e pass to the upstream baseline.
- AC3. `golangci-lint` is clean to the upstream baseline; `make verify` (`frontend-verify fmt lint test go-build frontend-e2e`) passes; husky/pre-commit hooks operate under the new name.
- AC4. Module path is `github.com/rc-tech/rc-harness`; the built binary is named `rc`; running it creates/reads `~/.rc` (not `~/.rc`); product displays as "RC".
- AC5. **Residual-string check**: a case-insensitive scan for `rc` across the fork (excluding `.git`, `node_modules`, regenerated lockfiles) returns no matches except the documented allowlist (MIT attribution / upstream credit, and any intentionally-preserved public identifiers such as `cy-*` if Open Question 2 resolves to keep). The allowlist is enumerated in SPEC and matches the scan output exactly.
- AC6. CLI terminal output and web UI render the RC orange→amber palette (no green brand hexes remain in `internal/charmtheme/theme.go`, `packages/ui/src/tokens.css`, `web/src/styles.css`, or other brand color sources); the RC logo replaces RC's; the welcome header renders the RC brand (rounded orange box + `RC vX.Y.Z`) per the resolved scope of Open Question 1.
- AC7. **Zero functional change**: command surface, flags, behavior, public API (`public_api_test.go`), and config keys (other than the brand dir name) are equivalent to RC. Only brand-intent test assertions changed.
- AC8. Smoke test: the built `rc` binary runs, prints the RC welcome header, reads/writes `~/.rc`, and exercises a core "idea→code" flow equivalent to RC.
- AC9. A clearly-separated **"Proposals — requires approval"** section exists (in PRD and carried through planning); **no** proposal is folded into the implementation TASKS.

## Open questions

> Genuine ambiguities for the orchestrator/user to resolve before SPEC. Defaults below are recommendations; the build proceeds on the default if unresolved.

1. **Welcome header scope (tension in REQUEST C4 vs CON1/AC6/N1).** REQUEST simultaneously asks to "re-implement to the RC layout" (rounded box, `Welcome back <user>!`, ASCII peak mascot, dimmed model · plan · email/org · cwd line) and to "preserve the header's existing data sources/behavior; only re-skin + relabel." The current source header is a simple box titled `"RC // SETUP"` and does **not** expose user/model/plan/org/cwd data. Building the richer header is net-new behavior and conflicts with "zero functional change."
   **Default:** faithfully re-skin + relabel the existing box to RC brand (rounded orange border, `RC vX.Y.Z`, equivalent title). Move the richer mascot/user/model header to Proposals. Only build the richer layout in-scope if all its data already exists in the current CLI context.

2. **`cy-*` prefix on skills/extensions** (`skills/cy-*`, `extensions/cy-*`, `.agents/skills/cy-*`). Is `cy` a RC abbreviation to rename, or a stable public skill/extension identifier referenced by `skills-lock.json` and embeds? Renaming changes public skill IDs.
   **Default:** keep `cy-*` (stable IDs), document as an intentional preservation in the AC5 allowlist.

3. **`~/.rc → ~/.rc` migration.** Default is fresh `.rc` with no silent migration (per REQUEST, migration is a Proposal). Confirm no committed runtime `.rc/` content must be ported (treated as source working data, excluded from the fork).

4. **MIT attribution allowlist.** Exactly which `rc`/original-author strings are intentionally preserved (LICENSE, any NOTICE, "forked from" credit)? This list must be enumerated so AC5's residual scan has a precise allowlist.

5. **Copy exclusion set.** Confirm exclusions: `.git/`, `bin/`, `dist/`, `web/dist/*`, `node_modules`, `.turbo/`, `coverage*`, `*.tsbuildinfo`, generated `routeTree.gen.ts`, and source working dirs (`.codex/`, `.claude/`, `.rc/`, `.factory/`, `.resources/`, `ai-docs`). SPEC will lock the precise set.

6. **Distribution targets.** Should brew tap (`homebrew-rc`), npm scope (`@rc-tech/cli` or similar), and AUR packaging be actively published, or only renamed-in-config (not published)? **Default:** rename in config only; no new publishing pipelines wired up.

## Proposals — requires approval (NOT part of implementation TASKS)

> Researched enhancements parked here per REQUEST. None is implemented during the fork+rebrand without explicit user sign-off.

- P1. Richer welcome header (`Welcome back <user>!`, ASCII RC peak mascot, dimmed model · plan · email/org · cwd line) if it requires net-new data sources not currently in the CLI (see Open Question 1).
- P2. `~/.rc → ~/.rc` automatic migration / back-compat shim for existing users.
- P3. Rename of the `cy-*` skill/extension prefix to an RC-scoped prefix (if a clean public ID change is desired).
- P4. Telemetry / analytics opt-in re-review under the RC identity.
- P5. CI matrix and distribution-pipeline updates (active brew/npm/AUR publishing under RC, new release workflows) beyond re-skinning existing config.
- P6. Any new RC-specific features.
