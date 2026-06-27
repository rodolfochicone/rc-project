---
provider: manual
pr:
round: 1
round_created_at: 2026-06-25T16:43:59Z
status: resolved
file: .gitignore
line: 1
severity: medium
author: claude-code
provider_ref:
---

# Issue 002: Stray rc-plugin-sync binary is not gitignored

## Review Comment

A compiled `rc-plugin-sync` Mach-O binary sits untracked at the repository root
(`git status` shows `?? rc-plugin-sync`), and `.gitignore` has no entry that
covers it. `cmd/rc-plugin-sync` is designed to be invoked via `go run` (the
release hook and the documented manual invocation both use `go run`), so a
built binary should never live in the repo. Because it is untracked rather than
ignored, a routine `git add .` or `git add -A` would commit a
platform-specific, architecture-specific binary into the source tree.

Suggested fix: remove the stray artifact and add an ignore entry so it cannot be
committed accidentally — e.g. add `/rc-plugin-sync` (and, if not already
covered, `/rc`) to `.gitignore`, matching however the existing root-level
binaries are ignored. Confirm `git status` is clean of build artifacts
afterward.

## Triage

- Decision: `VALID`
- Root cause: `cmd/rc-plugin-sync` is meant to run via `go run`, but a compiled
  `rc-plugin-sync` arm64 binary was left at the repo root and `.gitignore` had no
  entry covering it. The same applies to the `rc` main binary if built at root
  (`go build ./cmd/rc`), though `make` already routes that to `bin/`.
- Fix approach: add `/rc` and `/rc-plugin-sync` to `.gitignore` (root-level
  build artifacts, mirroring the existing `bin/`/`dist/` ignores) and delete the
  stray `rc-plugin-sync` artifact so `git status` is clean of build output.
- Notes: only a build artifact is removed (`rm` of an untracked Mach-O file);
  no tracked source or working-tree edits are discarded.
