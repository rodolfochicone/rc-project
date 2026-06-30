---
status: completed
title: "Docs — package tenet update and rc-project-memory skill"
type: docs
complexity: low
dependencies:
  - task_03
  - task_04
---

# Task 5: Docs — package tenet update and rc-project-memory skill

## Overview
Update the documentation to reflect that project memory now has a committed text mirror as its
shareable source of truth. This corrects the package doc comment's superseded "no markdown
mirror" tenet and teaches the export/import workflow in the `rc-project-memory` skill so agents
and users actually use it.

<critical>
- ALWAYS READ the TechSpec and ADR-001 before starting
- REFERENCE TECHSPEC for implementation details — do not duplicate here
- FOCUS ON "WHAT" — describe what needs to be accomplished, not how
- MINIMIZE CODE — show code only to illustrate current structure or problem areas
- TESTS REQUIRED — docs task: the "test" is that referenced commands/flags exist and match the shipped CLI
</critical>

<requirements>
- MUST update the package doc comment in `internal/core/projectmemory/store.go` so it no longer claims "there is no markdown mirror"; it MUST describe the committed `.rc/memory/` mirror and that the DB is a local cache/query index.
- MUST add `rc memory export` and `rc memory import` to the Command Reference in `skills/rc-project-memory/SKILL.md`, with a short "sharing across machines" workflow note (export → commit → import).
- MUST keep the doc accurate to the shipped commands from task_03/task_04 (flag names, output) — no aspirational behavior.
- SHOULD note the v1 limitation that deletions do not propagate automatically (per ADR-002).
- MUST NOT change runtime behavior — documentation only.
</requirements>

## Subtasks
- [x] 5.1 Rewrite the `projectmemory` package doc comment to describe the mirror and the DB-as-cache model.
- [x] 5.2 Add the two commands to the SKILL.md Command Reference.
- [x] 5.3 Add a short "Sharing across machines" section to the skill (export/commit/import; deletion caveat).
- [x] 5.4 Cross-check every documented flag/command against the shipped CLI.

## Implementation Details
Edit `internal/core/projectmemory/store.go` (top-of-file package comment, lines 1-6) and
`skills/rc-project-memory/SKILL.md` (the "Command Reference" list around lines 20-31 and a new
short workflow subsection). Keep the skill's existing tone and bullet style. Reference TechSpec
"Executive Summary" for the one-line framing of the mirror.

### Relevant Files
- `internal/core/projectmemory/store.go` — package doc comment (lines 1-6) with the superseded tenet.
- `skills/rc-project-memory/SKILL.md` — Command Reference (lines 20-31) and curation/workflow guidance.

### Dependent Files
- None — documentation only.

### Patterns to Mirror
```text
// SOURCE: skills/rc-project-memory/SKILL.md:22-23
- `rc memory search "<terms>" [--scope <s>] [--limit N] [--format json]` — ranked lookup.
- `rc memory add --scope <s> --title "<t>" --body "<b>" [--key <k>] [--tags a,b] [--source <origin>]` — insert or upsert.
```

### Related ADRs
- [ADR-001: Share project memory through a committed markdown-per-fact mirror](adrs/adr-001.md) — the superseded tenet and the new model.
- [ADR-002: Mirror identity and most-recent-wins conflict resolution](adrs/adr-002.md) — the deletion caveat to document.

## Deliverables
- Updated package doc comment in `store.go`.
- Updated `skills/rc-project-memory/SKILL.md` with the two commands and the sharing workflow.
- Verification that documented commands/flags match the shipped CLI **(REQUIRED)**.

## Tests
- Unit tests:
  - [ ] Not applicable (documentation only). `make verify` MUST still pass (fmt/lint/build over the changed Go comment).
- Integration tests:
  - [ ] Manual check: `rc memory export --help` and `rc memory import --help` match what the skill documents.
- Test coverage target: not applicable (no new code paths)
- All tests must pass

## Success Criteria
- All tests passing (`make verify` green)
- The package doc no longer claims "no markdown mirror"
- SKILL.md documents export/import and the sharing workflow accurately
- Deletion-propagation caveat is stated
