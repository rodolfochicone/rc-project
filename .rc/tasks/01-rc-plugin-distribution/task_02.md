---
status: completed
title: Sync the manifest version into the OSS GoReleaser release flow
type: infra
complexity: medium
dependencies:
  - task_01
---

# Task 2: Sync the manifest version into the OSS GoReleaser release flow

## Overview
Wire the plugin manifest `version` to the release tag so it tracks `vX.Y.Z` automatically
during a release. Releases always run through the OSS GoReleaser flow
(`.goreleaser.oss.yml`), which is the only release path; this task adds a step to that
flow that rewrites the `version` field in both manifests, preventing the silent
"forgotten bump" failure mode where users stop receiving updates.

<critical>
- ALWAYS READ the PRD and TechSpec before starting
- REFERENCE TECHSPEC for implementation details — do not duplicate here
- FOCUS ON "WHAT" — describe what needs to be accomplished, not how
- MINIMIZE CODE — show code only to illustrate current structure or problem areas
- TESTS REQUIRED — every task MUST include tests in deliverables
</critical>

<requirements>
- MUST update the manifest `version` in both `.claude-plugin/plugin.json` and `.claude-plugin/marketplace.json` to the release tag as part of the OSS GoReleaser flow (`.goreleaser.oss.yml`) — the only release path.
- MUST implement the rewrite as a small, testable Go helper invoked by a GoReleaser hook, NOT as an inline `sed`/`awk` edit (per the no-workarounds rule; the manifests are structured JSON).
- The helper MUST accept the target version (sourced from the GoReleaser version/tag) and write it verbatim into both manifests, preserving all other fields and JSON formatting.
- The helper MUST strip a leading `v` if the tag carries one, so the manifest `version` matches the `0.14.0` style used in the TechSpec data model.
- The helper MUST fail loudly (non-zero exit) if a manifest is missing or unparseable, so a release aborts rather than shipping a stale `version`.
- After the rewrite, the task_01 consistency test (`plugin.json` version == `marketplace.json` version) MUST still hold.
- `make verify` MUST pass with zero lint issues.
</requirements>

## Subtasks
- [x] 2.1 Add a small Go helper that rewrites the `version` field in both manifests given a target version string.
- [x] 2.2 Normalize the input version (strip a leading `v`) and validate both manifests parse before writing.
- [x] 2.3 Invoke the helper from a `.goreleaser.oss.yml` hook so the release tag flows into the manifests.
- [x] 2.4 Add unit tests for the helper: correct rewrite, version normalization, and error on missing/malformed manifest.
- [x] 2.5 Verify a GoReleaser dry-run/snapshot leaves both manifests carrying the expected version.

## Implementation Details
Add the helper under `cmd/` (a tiny standalone `main` package) or `internal/`, wired into
`.goreleaser.oss.yml`. The existing flow injects `internal/version.Version={{ .Version }}`
via `ldflags` and runs `before: hooks: - make frontend-build`; add the manifest-sync step
to the hook chain (or an equivalent `before` hook) so the GoReleaser version template
reaches the helper. Reference the TechSpec "Integration Points" (OSS GoReleaser release
flow) and "Development Sequencing" build order step 4. The JSON shape the helper edits is
defined in the TechSpec "Data Models" section.

Keep the change additive and minimal: the helper only rewrites the `version` field; it
does not introduce manifest-generation logic or new manifest fields.

### Relevant Files
- `.goreleaser.oss.yml` — OSS release config; `builds[].ldflags` and `before.hooks` define where the version-sync step is added.
- `internal/version/version.go` — `Version` default `"dev"`, stamped at build time; reference for how the tag is consumed elsewhere.
- `.claude-plugin/plugin.json` — `version` field rewritten by the helper (from task_01).
- `.claude-plugin/marketplace.json` — `version` field rewritten by the helper (from task_01).

### Dependent Files
- `Makefile` — may gain a target that the GoReleaser hook calls (e.g. a `plugin-version-sync` target); `make verify` runs the helper's unit tests.
- `test/plugin_consistency_test.go` — its version-equality assertion must keep holding after a sync.

### Related ADRs
- [ADR-002: Pin plugin updates to release tags via the `version` field](adrs/adr-002.md) — this task implements the per-release bump folded into the OSS release flow.

## Deliverables
- A tested Go helper that rewrites the `version` field in both manifests from a target version.
- A `.goreleaser.oss.yml` hook step that invokes the helper with the release tag during the OSS flow.
- Unit tests with 80%+ coverage of the helper's rewrite, normalization, and error paths **(REQUIRED)**.
- Integration evidence: a GoReleaser snapshot/dry-run showing both manifests carry the resolved version **(REQUIRED)**.

## Tests
- Unit tests:
  - [ ] Given version `1.2.3`, the helper writes `"version": "1.2.3"` to both manifests and leaves all other fields unchanged.
  - [ ] Given tag `v1.2.3`, the helper normalizes to `1.2.3` (leading `v` stripped).
  - [ ] A missing `.claude-plugin/plugin.json` causes the helper to exit non-zero with a descriptive error.
  - [ ] A malformed (invalid JSON) manifest causes the helper to exit non-zero without partially writing.
- Integration tests:
  - [ ] Run the helper, then assert `plugin.json` version equals `marketplace.json` version (reuses the task_01 invariant).
  - [ ] A `goreleaser release --snapshot --clean -f .goreleaser.oss.yml` dry-run completes with the hook executed and both manifests carrying the snapshot version.
- Test coverage target: >=80%
- All tests must pass

## Success Criteria
- All tests passing
- Test coverage >=80%
- A release through `.goreleaser.oss.yml` writes the release tag into both manifests' `version` fields.
- Running the helper twice is idempotent for the same version.
- `make verify` (fmt + lint + test + build) passes at 100% with zero lint issues.
