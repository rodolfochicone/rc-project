---
provider: manual
pr:
round: 1
round_created_at: 2026-06-15T10:54:55Z
status: resolved
file: web/src/systems/runs/components/run-detail-view.tsx
line: 429
severity: low
author: claude-code
provider_ref:
---

# Issue 003: Duplicated terminal-run-status helper across detail view and route

## Review Comment

`isTerminalRunStatus` in `run-detail-view.tsx:429-440` is a near-verbatim copy of
`isTerminalStatus` in `web/src/routes/_app/runs_.$runId.tsx:176-188` — both
lowercase the status and compare against the same list (`completed`, `succeeded`,
`failed`, `crashed`, `canceled`, `cancelled`). The two copies can drift (e.g. a
new terminal status added in one place but not the other), which would desync the
input-panel gating from the route's `runTerminated` logic.

Suggested fix: extract one shared helper (e.g. `isTerminalRunStatus` exported
from `systems/runs/lib`) and import it in both the route and the detail view, so
the terminal-status definition lives in a single place. This mirrors how
`isTerminalKind` is already shared for event kinds.

## Triage

- Decision: `VALID`
- Notes: Confirmed near-verbatim duplication between the route and the detail
  view. Fix applied: extracted a single `isTerminalRunStatus` into
  `web/src/systems/runs/lib/run-status.ts`, exported it from the runs barrel, and
  imported it in both `run-detail-view.tsx` and `routes/_app/runs_.$runId.tsx`
  (the local copies were removed). Mirrors how `isTerminalKind` is already shared.
  Touching the route file (outside the issue's `file:` path) was the minimum
  needed to remove the duplicate definition. Added a behavior test in
  `run-detail-view.test.tsx` asserting a terminated run hides the input panel even
  with a stale snapshot `pending_input`, which exercises the shared helper.
