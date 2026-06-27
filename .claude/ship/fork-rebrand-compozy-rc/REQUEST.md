# REQUEST — Fork & Rebrand rc as "rc"

> Authoritative request for this plan. All downstream artifacts (PRD / SPEC / TASKS) derive from this file.
> Language note: source instructions were given in Portuguese; this spec is in English for tooling consistency. Brand-facing copy stays as specified.

## Objective

Faithfully fork the entire **rc** project (located at `~/dev/rc`) into the currently-empty target directory `~/dev/rc-harness-cli`, and rebrand it as **rc**, an rc-tech product. **Preserve 100% of functionality.** The only intended changes are:

1. Product **identity / naming** (module path, binary, all `rc` strings → rc equivalents).
2. **Visual identity** (logo asset, color palette, terminal brand colors).
3. The CLI **welcome header**.

This is a rebrand-only fork. No new features are built as part of implementation.

## Context

- **Source project:** rc — a CLI that orchestrates AI agents "from idea to delivered code." Go backend + web frontend (React / Vite / TypeScript, using `@ai-sdk/react` and `@assistant-ui/react`). ~37M, **MIT licensed** (fork is permitted).
- **Toolchain (confirmed in source):** Go module `github.com/rc/rc`; web build via **bun + vite**; tests via **vitest + playwright** (web) and **go test** (backend); release via **goreleaser**; lint via **golangci-lint**; **husky** git hooks; `bun.lock` present; shadcn UI (`web/components.json`, `web/src/styles.css`, `baseColor: neutral`, `iconLibrary: lucide`).
- **Target:** `~/dev/rc-harness-cli` (empty; not yet a git repo).
- **Project conventions (global CLAUDE.md):** Think before coding; simplicity first; surgical changes; goal-driven with verifiable success criteria; surface conflicts rather than blend; read before write; tests verify intent; match the codebase's existing conventions. → For this fork that means: **conform to rc's existing structure and style; rename mechanically and completely; change nothing else.**

### Concrete rename anchors discovered in source (non-exhaustive; SPEC must complete the inventory)

- **Go module:** `go.mod` → `module github.com/rc/rc`.
- **Public package entrypoint:** `rc.go` (`package rc`, imports `github.com/rc/rc/internal/...`).
- **CLI main:** `cmd/rc/` (dir + `main.go`); goreleaser `binary: rc`, `main: ./cmd/rc`, `project_name: rc`, ldflags inject `github.com/rc/rc/internal/version.*`.
- **Welcome header:** `printWelcomeHeader(cmd)` in `internal/cli/setup.go` (called ~line 126, defined ~line 490); covered by `internal/cli/theme_test.go`.
- **Terminal brand palette:** `internal/charmtheme/theme.go` — current green brand (`ColorBrand=#CAEA28`, accents `#A3E635/#84CC16/#65A30D`, progress gradient `#65A30D→#CAEA28`) must become the rc orange→amber palette.
- **Config home dir:** `internal/config/home.go` → `DirName = ".rc"` (the `~/.rc` user directory and all `.rc` path references across code + tests).
- **Web package identity:** root `package.json` `"name": "rc"`; `web/package.json`; shadcn aliases `@rc/ui`, `@rc/ui/utils` in `web/components.json`; web brand tokens in `web/src/styles.css`.
- **Scale of textual rename:** ~438 Go files and ~58 web TS/TSX files contain the string `rc` (case-insensitive). Plus: `.goreleaser.yml`, `Makefile`, `README.md`, `AGENTS.md`, `CLAUDE.md`, `CONTRIBUTING.md`, `CHANGELOG.md`/`RELEASE_NOTES.md`, `aur-pkg/`, `.github/`, `.husky/`, `cliff.toml`, `docs/`, `skills/`, `openapi/`, `sdk/`, `packages/`, `extensions/`, `zsh/`, badges, homepage URLs, support emails, telemetry identifiers.

## Approved Decisions (do NOT re-ask)

1. **Naming**
   - Binary / CLI command: **`rc`**.
   - Go module path: **`github.com/rc-tech/rc-harness`** — replace ALL references to `github.com/rc/rc` and the bare string `rc` (internal and user-visible).
   - Brand domain: **rc.tech**.
   - Displayed product name: **"rc"**.
2. **Scope** — complete faithful fork: copy the entire project (`cmd/`, `internal/`, `pkg/`, engine, `web/`, `docs/`, `skills/`, `scripts/`, `openapi/`, `sdk/`, `packages/`, `extensions/`, etc.), rebrand, and keep ALL functionality identical to rc. rc is MIT; fork is allowed.
3. **Improvements** — proposed enhancements / researched features go in a **separate "Proposals — requires approval"** section only. They are **NOT** part of the implementation TASKS. Implementation TASKS cover **only** the faithful fork + rebrand. Nothing new is built without explicit user approval.

## rc Visual Identity (source of truth — subagents have no access to the reference images)

- **Logo:** a rounded peak / triangle mark (mountain / stylized "A"), open stroke / outline, with a **gradient from warm orange (top-left) to golden-yellow (bottom-right)**. Beside it, the wordmark **"rc"** in lowercase, geometric sans-serif, bold weight, black / near-black.
- **Color palette (theme tokens derived from the logo):**
  - Primary orange: `~#F26B21` / `#F37021`
  - Amber / gold: `~#FBB034` / `#FDB813`
  - Orange→gold gradient for accents and the icon.
  - Text near-black: `#1C1C1C`
  - Light backgrounds: `#FFFFFF` / `#FAFAFA`; dark theme also supported.
  - Replace rc's current (green) palette with this in the web theme (Tailwind / `components.json` / CSS tokens) and in every CLI brand color (colored terminal output).
- **Welcome header (CLI):** bordered terminal box, rounded single-line border in rc **orange**; top-left a label with app name + version (e.g. `rc vX.Y.Z`); centered `Welcome back <user name>!`; below center a mascot / icon (the rc peak/logo rendered as ASCII / pixel art); bottom lines in dimmed text: model · plan · email/organization · current directory (cwd). Reproduce this layout as the equivalent of rc's existing welcome header, with frame/accents in the rc brand color.

## Requirements

### A. Copy / scaffold the fork

- A1. Copy the full rc tree from `~/dev/rc` into `~/dev/rc-harness-cli`, preserving structure (all listed dirs + dotfiles + config). Exclude only VCS/build cruft as appropriate (e.g. `.git/`, build output) per SPEC.
- A2. Initialize the target as its own repo state if needed (no upstream rc history required); LICENSE remains MIT with attribution preserved as required by the MIT license.

### B. Identity / naming rename (functionality-preserving)

- B1. Go module path `github.com/rc/rc` → `github.com/rc-tech/rc-harness`; update every import path across all `.go` files.
- B2. Rename CLI entrypoint dir `cmd/rc/` → `cmd/rc/`; binary `rc` → `rc`; goreleaser `project_name`, `binary`, `main`, ldflag version package paths, homepage, description, maintainer email/domain.
- B3. Public package `rc.go` / `package rc` renamed consistently (package name + file), keeping the public API surface intact.
- B4. User config directory `.rc` → `.rc` (`internal/config/home.go DirName` and every `.rc` path reference in code AND tests). Decide & document migration/back-compat behavior for existing `~/.rc` dirs (default: fresh `.rc`, no silent migration unless trivially safe — flag in Proposals if migration is desired).
- B5. Web identity: root + `web/package.json` `name`, shadcn aliases `@rc/ui*` → rc-scoped equivalents, and any `rc` in TS/TSX.
- B6. All remaining textual occurrences (case-sensitive `rc`, `rc`, `rc`): docs, README, AGENTS.md, CLAUDE.md, CONTRIBUTING, changelogs, `.github/`, `.husky/`, `Makefile`, `cliff.toml`, `aur-pkg/`, `zsh/`, `scripts/`, `openapi/`, `sdk/`, `skills/`, `extensions/`, badges, URLs (`github.com/rc/rc`, `rc.com`, support emails), and any telemetry / user-agent / analytics identifiers.
- B7. Produce and include in SPEC a **complete inventory** of every place the `rc` identity appears (binary, module, configs, docs, badges, telemetry, `~/.rc` paths, package scopes, asset filenames) so the rename is provably complete.

### C. Visual rebrand

- C1. CLI terminal palette: replace green brand tokens in `internal/charmtheme/theme.go` (and any other color source) with the rc orange→amber palette and gradient. Update affected tests (`internal/charmtheme/theme_test.go`, `internal/cli/theme_test.go`) to assert the new brand intent.
- C2. Web theme: replace rc's palette in `web/src/styles.css` CSS variables / shadcn tokens (`components.json`) and anywhere brand colors are defined, with the rc palette; keep light + dark theme support.
- C3. Logo asset: create the rc logo (SVG peak mark + gradient + "rc" wordmark) and replace rc's brand assets (`imgs/`, web `public/`, favicons, README hero) where the rc logo is referenced.
- C4. CLI welcome header: re-implement `printWelcomeHeader` to the rc layout described above (rounded orange box, `rc vX.Y.Z` label, `Welcome back <user>!`, ASCII/pixel rc peak mascot, dimmed model · plan · email/org · cwd line). Preserve the header's existing data sources/behavior; only re-skin + relabel.

### D. Verification strategy

- D1. `go build ./...` succeeds; `go test ./...` passes (faithful behavior; only brand-intent assertions changed).
- D2. Web builds via bun/vite (`bun install` + vite build) succeeds; **vitest** unit tests pass; **playwright** e2e pass (or run to the same baseline as upstream).
- D3. golangci-lint clean (to upstream baseline); husky/pre-commit hooks operate under the new name.
- D4. Smoke test: built `rc` binary runs, prints the new welcome header, reads/writes `~/.rc`, and exercises a core "idea→code" flow equivalent to rc.
- D5. A residual-string check: no unexpected `rc` (case-insensitive) remains except where intentionally preserved (e.g. MIT attribution to original author / upstream credit), and those exceptions are documented.

## Acceptance Criteria

- AC1. `~/dev/rc-harness-cli` contains the full faithful fork; `go build ./...` and `go test ./...` pass.
- AC2. Web `bun install` + vite build pass; vitest + playwright pass to the upstream baseline.
- AC3. Module path is `github.com/rc-tech/rc-harness`; binary is `rc`; user config dir is `~/.rc`; product displays as "rc".
- AC4. No unintended `rc` identity remains (verified by D5); intentional preservations (license attribution) are listed.
- AC5. CLI terminal output and web UI use the rc orange→amber palette; the rc logo replaces rc's; the welcome header matches the described rc layout.
- AC6. **Zero functional change** vs. rc beyond identity + visuals (same commands, flags, behavior, APIs).
- AC7. A clearly-separated **"Proposals — requires approval"** section exists, with NO proposal folded into implementation TASKS.

## Constraints

- CON1. Faithful fork only — **no new functionality, no refactors, no adjacent "improvements"** in the implementation TASKS (CLAUDE.md Rules 2 & 3). Match rc's existing conventions even where debatable (Rule 8); surface—don't silently change—anything questionable (Rule 5).
- CON2. Preserve the MIT LICENSE and any legally-required attribution to the original rc authors.
- CON3. Subagents have **no access to the reference images** — the Visual Identity section above is the single source of truth for colors, logo, and header layout.
- CON4. Keep all behavior, command surface, config keys (other than the brand dir name), and public API stable.
- CON5. Brand-facing copy uses "rc" / lowercase "rc" wordmark exactly as specified; domain rc.tech; module `github.com/rc-tech/rc-harness`.
- CON6. Do not build any "Proposals" item without explicit user approval.

## Proposals — Requires Approval (NOT part of implementation TASKS)

> Placeholder for the planning phase to populate. Researched enhancements and suggested improvements (e.g. `~/.rc → ~/.rc` auto-migration, telemetry opt-in re-review, CI matrix updates, new rc-specific features) live HERE only and require explicit user sign-off before any are scheduled. Nothing in this section is implemented during the fork+rebrand.
