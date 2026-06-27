---
status: completed
title: Add plugin manifests (plugin.json + marketplace.json) and consistency test
type: infra
complexity: medium
dependencies: []
---

# Task 1: Add plugin manifests (plugin.json + marketplace.json) and consistency test

## Overview
Add the two Claude Code plugin manifests at the repository root and a Go test that
guards their validity and keeps them in sync with the embedded `skills/` and `commands/`
assets. This is the foundation of the plugin distribution channel: once the manifests
exist, the repo doubles as a single-plugin marketplace without duplicating any skill
content.

<critical>
- ALWAYS READ the PRD and TechSpec before starting
- REFERENCE TECHSPEC for implementation details — do not duplicate here
- FOCUS ON "WHAT" — describe what needs to be accomplished, not how
- MINIMIZE CODE — show code only to illustrate current structure or problem areas
- TESTS REQUIRED — every task MUST include tests in deliverables
</critical>

<requirements>
- MUST create `.claude-plugin/plugin.json` at the repo root with `name == "rc"`, a non-empty `version`, a `description`, and an `author` object, per the TechSpec "Data Models" section.
- MUST create `.claude-plugin/marketplace.json` at the repo root declaring marketplace `rc-project` listing a single plugin `rc` with `source == "."` and a non-empty `version` matching `plugin.json`.
- MUST NOT modify `skills/embed.go` or `commands/embed.go`; the new `.claude-plugin/` directory MUST stay outside their embed globs (`*/SKILL.md */references/*` and `*.md`).
- MUST add a table-driven Go test that parses both manifests, asserts the field invariants above, and verifies coherence with `skills.FS` (every embedded `rc-*` skill directory has a `SKILL.md`) and `commands.FS`.
- The test MUST reuse `skills.FS` / `commands.FS` rather than re-globbing the working directory, and MUST run with no network and no `rc setup` invocation.
- `make verify` MUST pass with zero lint issues.
</requirements>

## Subtasks
- [x] 1.1 Create `.claude-plugin/plugin.json` with the manifest fields defined in the TechSpec.
- [x] 1.2 Create `.claude-plugin/marketplace.json` referencing the `rc` plugin with `source: "."`.
- [x] 1.3 Confirm the binary still builds and the embed globs do not pick up `.claude-plugin/`.
- [x] 1.4 Add `test/plugin_consistency_test.go` parsing both manifests and asserting field invariants.
- [x] 1.5 Extend the test to assert manifest/asset coherence against `skills.FS` and `commands.FS`.

## Implementation Details
Create the two manifest files at the repository root under `.claude-plugin/`. The plugin
root is the repo root, so Claude Code auto-discovers the existing `skills/` and
`commands/` directories — no content is copied. See the TechSpec "Data Models" section
for the exact JSON shape and the "Core Interfaces" section for the Go structs the test
should use to parse the manifests (these structs live only in the test package; no
production type is added).

Place the test in the existing `test/` package alongside `test/skills_bundle_test.go`,
which already imports `github.com/rodolfochicone/rc-project/skills` and demonstrates the
table-driven + `t.Parallel()` + `t.Run()` style to follow. Read the manifests from disk
relative to the repo root and decode them with `encoding/json`. Use `skills.FS` and
`commands.FS` to enumerate embedded assets for the coherence check.

### Relevant Files
- `.claude-plugin/plugin.json` — new plugin manifest (create).
- `.claude-plugin/marketplace.json` — new marketplace catalog (create).
- `test/plugin_consistency_test.go` — new consistency test (create).
- `test/skills_bundle_test.go` — existing reference for test style and `skills.FS` usage.
- `skills/embed.go` — defines `skills.FS` (glob `*/SKILL.md */references/*`); read-only.
- `commands/embed.go` — defines `commands.FS` (glob `*.md`); read-only.

### Dependent Files
- `skills/embed.go`, `commands/embed.go` — must remain unchanged; verify the new directory is not embedded.
- `Makefile` — `make verify` runs the new test; no edit expected.

### Related ADRs
- [ADR-001: Distribute rc skills and commands as a Claude Code plugin hosted in rc-project](adrs/adr-001.md) — defines the single-plugin-marketplace layout these manifests implement.
- [ADR-003: Keep the `rc-` name prefix; plugin is an additive Claude-only channel](adrs/adr-003.md) — the manifests expose skills as `rc:rc-*` without renaming directories.

## Deliverables
- `.claude-plugin/plugin.json` at the repo root, valid against the TechSpec data model.
- `.claude-plugin/marketplace.json` at the repo root, valid and consistent with `plugin.json`.
- `test/plugin_consistency_test.go` with table-driven subtests covering manifest validity and asset coherence.
- Confirmation that `make build` still succeeds with the manifests present.
- Unit tests with 80%+ coverage of the new test's parsing/validation logic **(REQUIRED)**.
- Integration tests for manifest/asset coherence against the embedded FS **(REQUIRED)**.

## Tests
- Unit tests:
  - [ ] `plugin.json` parses as valid JSON and `name == "rc"`.
  - [ ] `plugin.json` `version` is non-empty.
  - [ ] `marketplace.json` parses as valid JSON and lists exactly one plugin named `rc`.
  - [ ] The `rc` marketplace entry has `source == "."` and a non-empty `version`.
  - [ ] `marketplace.json` plugin `version` equals `plugin.json` `version`.
- Integration tests:
  - [ ] Every directory in `skills.FS` whose name starts with `rc` (e.g. `rc-execute-task`) contains a `SKILL.md`, so a renamed/added skill that breaks discovery fails the test.
  - [ ] `commands.FS` exposes the expected `rc-*.md` command files (e.g. `rc-exec.md`, `rc-review.md`).
- Test coverage target: >=80%
- All tests must pass

## Success Criteria
- All tests passing
- Test coverage >=80%
- Both manifest files exist at the repo root and parse cleanly with the documented invariants.
- `make verify` (fmt + lint + test + build) passes at 100% with zero lint issues.
- `make build` produces a working binary, confirming the embed globs ignore `.claude-plugin/`.
