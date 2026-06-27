---
status: completed
title: Document the plugin install channel and integration runbook
type: docs
complexity: low
dependencies:
  - task_01
---

# Task 3: Document the plugin install channel and integration runbook

## Overview
Document the new Claude Code plugin as an additive install path alongside `rc setup`, and
capture the manual integration runbook used to validate it. Users need to know how to add
the marketplace, install the plugin, the `GH_TOKEN` prerequisite for the private repo, and
that the plugin channel serves Claude Code only.

<critical>
- ALWAYS READ the PRD and TechSpec before starting
- REFERENCE TECHSPEC for implementation details — do not duplicate here
- FOCUS ON "WHAT" — describe what needs to be accomplished, not how
- MINIMIZE CODE — show code only to illustrate current structure or problem areas
- TESTS REQUIRED — every task MUST include tests in deliverables
</critical>

<requirements>
- MUST document, in `README.md`, the plugin install flow: `/plugin marketplace add rodolfochicone/rc-project` then `/plugin install rc@rc-project`, and `/plugin marketplace update` to pull new tagged versions.
- MUST state the `GH_TOKEN` (or configured git credentials) prerequisite, since `rodolfochicone/rc-project` is private and the marketplace add/auto-update clones it.
- MUST present the plugin as ADDITIVE to `rc setup` and note it is Claude-Code-only; other agents (codex, cursor, droid, gemini, opencode, pi) keep using `rc setup`.
- MUST note that skills surface as `rc:rc-*` under the plugin namespace and recommend picking ONE channel (plugin OR setup) to avoid duplicate listings.
- MUST update `CLAUDE.md` and `skills/rc/SKILL.md` so the plugin path is discoverable from project guidance and the root skill description.
- MUST include the manual integration runbook (local-path add → install → verify `/rc:rc-*` resolve → bumped-version update) as the documented validation procedure for this feature.
</requirements>

## Subtasks
- [x] 3.1 Add a plugin install section to `README.md` near the existing install/quick-reference area.
- [x] 3.2 Document the `GH_TOKEN` prerequisite and the Claude-only, additive scope.
- [x] 3.3 Update `CLAUDE.md` and `skills/rc/SKILL.md` to reference the plugin channel.
- [x] 3.4 Write the manual integration runbook (local-path add, install, verify, update).

## Implementation Details
Update the three documentation surfaces. In `README.md`, add the plugin install path
next to the existing CLI install / "Quick Reference" content so both channels sit
together. In `CLAUDE.md`, note the plugin as an alternative distribution for Claude Code.
In `skills/rc/SKILL.md` (the root `rc` skill that explains rc capabilities), mention the
plugin install option. Reference the TechSpec "API Endpoints" section for the exact slash
commands and the "Integration Points" / "Testing Approach → Integration Tests" sections
for the runbook steps and the `GH_TOKEN` prerequisite. Do not duplicate ADR rationale into
the docs; link or summarize at a user-facing level.

The integration validation is manual by design (per the TechSpec); the automated gate for
this feature is the task_01 consistency test under `make verify`.

### Relevant Files
- `README.md` — primary user-facing install/usage doc; add the plugin section here.
- `CLAUDE.md` — project guidance; note the additive Claude-only plugin channel.
- `skills/rc/SKILL.md` — root `rc` skill describing rc capabilities and install paths.
- `.claude-plugin/marketplace.json` — source of the marketplace/plugin names referenced in docs (from task_01).

### Dependent Files
- `.claude-plugin/plugin.json` — the documented plugin `name` (`rc`) must match (from task_01).

### Related ADRs
- [ADR-001: Distribute rc skills and commands as a Claude Code plugin hosted in rc-project](adrs/adr-001.md) — explains the `GH_TOKEN`/heavy-clone prerequisite documented here.
- [ADR-003: Keep the `rc-` name prefix; plugin is an additive Claude-only channel](adrs/adr-003.md) — basis for the "additive, Claude-only, pick one channel" guidance.

## Deliverables
- A `README.md` plugin install section covering add/install/update commands and the `GH_TOKEN` prerequisite.
- `CLAUDE.md` and `skills/rc/SKILL.md` updated to reference the plugin channel.
- A documented manual integration runbook for validating the plugin end to end.
- Validation that documented commands and plugin/marketplace names match the manifests from task_01 **(REQUIRED)**.

## Tests
- Unit tests:
  - [ ] Documentation review confirms the marketplace name (`rc-project`), plugin name (`rc`), and slash commands in the docs exactly match `.claude-plugin/marketplace.json` and `.claude-plugin/plugin.json`.
  - [ ] No automated unit tests apply to Markdown docs; correctness is verified by the name/command cross-check above and the runbook below.
- Integration tests (manual runbook — per TechSpec, not automated):
  - [ ] In a scratch Claude Code session, `/plugin marketplace add ./` (local path) succeeds.
  - [ ] `/plugin install rc@rc-project` installs the plugin.
  - [ ] `/rc:rc-create-prd` and `/rc:rc-execute-task` resolve as plugin skills.
  - [ ] After bumping the manifest `version`, `/plugin marketplace update` reports and applies the update.
- Test coverage target: not applicable to documentation; the feature's automated coverage is provided by task_01.
- All tests must pass (`make verify` stays green; no code changes in this task)

## Success Criteria
- All tests passing (`make verify` remains green; this task adds no code)
- `README.md`, `CLAUDE.md`, and `skills/rc/SKILL.md` describe the plugin install path, the `GH_TOKEN` prerequisite, and the additive Claude-only scope.
- Documented plugin/marketplace names and slash commands match the manifests from task_01.
- The manual integration runbook is complete enough to execute end to end without further guidance.
