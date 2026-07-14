---
status: completed
title: Frontend input data layer
type: frontend
complexity: medium
dependencies:
  - task_07
---

# Task 8: Frontend input data layer

## Overview
Add the web data layer for answering a run: a typed adapter that POSTs to the new
input endpoint, a TanStack Query mutation hook that invalidates the affected run
queries, and the `interactive` flag on the exec start request. This mirrors the
existing cancel adapter/hook and keeps UI components (task_09) thin.

<critical>
- ALWAYS READ the PRD and TechSpec before starting
- REFERENCE TECHSPEC for implementation details — do not duplicate here
- FOCUS ON "WHAT" — describe what needs to be accomplished, not how
- MINIMIZE CODE — show code only to illustrate current structure or problem areas
- TESTS REQUIRED — every task MUST include tests in deliverables
</critical>

<requirements>
- MUST add `sendRunInput(params)` to the runs adapter using the generated client's
  `POST /api/runs/{run_id}/input`, mirroring `cancelRun`'s error handling.
- MUST add a `useSendRunInput` mutation hook that invalidates the run, snapshot,
  and transcript queries on success (mirroring `useCancelRun`).
- MUST expose `pending_input` and `RunInputRequest` via the runs `types.ts` from
  the regenerated OpenAPI types.
- MUST add an optional `interactive` flag to the exec start request/adapter so a
  run can be started interactively.
- MUST NOT contain UI rendering — that belongs to task_09.
</requirements>

## Subtasks
- [x] 8.1 Add `sendRunInput` to the runs API adapter.
- [x] 8.2 Add `useSendRunInput` with the right query invalidations.
- [x] 8.3 Re-export `pending_input`/`RunInputRequest` types from runs `types.ts`.
- [x] 8.4 Add `interactive` to the exec start params/adapter and hook.

## As-built notes (for downstream tasks)
- **`sendRunInput({ runId, input })`** in `runs-api.ts` mirrors `cancelRun`; the
  body is the generated `RunInputRequest` type. **`useSendRunInput()`** invalidates
  the run, snapshot, and transcript query keys (not `lists()` — answering input
  does not change the run list). Both exported from `@/systems/runs`.
- **Types re-exported from runs `types.ts`**: `RunPendingInput` (the
  `pending_input` shape), `RunInputRequest`, and `RunInputOption` (for task_09's
  option rendering). The snapshot type already resolves `pending_input` via
  `RunSnapshotPayload`.
- **`interactive` is an optional flag on `StartExecParams`** in `exec-api.ts`,
  added to the POST body only when truthy. The `useStartExec` hook needed no
  change — it forwards `StartExecParams` straight to `startExec`.

## Implementation Details
Mirror `cancelRun` (`web/src/systems/runs/adapters/runs-api.ts:67`) and
`useCancelRun` (`web/src/systems/runs/hooks/use-runs.ts:67`). Pull the new types
from `@/generated/rc-openapi` via `web/src/systems/runs/types.ts`. Extend the exec
start path in `web/src/systems/exec/adapters/exec-api.ts` and
`web/src/systems/exec/hooks/use-exec.ts` with the optional `interactive` flag (see
TechSpec "API Endpoints"). Use the shared typed client in
`web/src/lib/api-client.ts`.

### Relevant Files
- `web/src/systems/runs/adapters/runs-api.ts` — `cancelRun` (:67) pattern; add
  `sendRunInput`.
- `web/src/systems/runs/hooks/use-runs.ts` — `useCancelRun` (:67) pattern; add
  `useSendRunInput`.
- `web/src/systems/runs/types.ts` — re-export generated `pending_input`/
  `RunInputRequest` types.
- `web/src/systems/exec/adapters/exec-api.ts` / `hooks/use-exec.ts` / `types.ts` —
  add `interactive`.
- `web/src/lib/api-client.ts` — shared typed client.

### Dependent Files
- `web/src/systems/runs/components/run-detail-view.tsx` and the new response panel
  (task_09) consume the hook.
- `web/src/generated/rc-openapi.d.ts` — source of the new types (task_07).

### Related ADRs
- [ADR-004: Interactivity is opt-in via a flag set at run start](../adrs/adr-004.md) — the `interactive` start flag.

## Deliverables
- `sendRunInput` adapter + `useSendRunInput` hook.
- Re-exported `pending_input`/`RunInputRequest` types.
- `interactive` flag on the exec start path.
- Unit tests with 80%+ coverage **(REQUIRED)**
- Integration tests for the mutation/invalidations **(REQUIRED)**

## Tests
- Unit tests:
  - [x] `sendRunInput` POSTs to `/api/runs/{run_id}/input` with the `prompt_id` +
        `option_id`/`text` body and throws a descriptive error on non-ok response.
  - [x] `useSendRunInput` invalidates the run, snapshot, and transcript query keys
        on success.
  - [x] The exec start adapter includes `interactive: true` in the body only when
        the flag is set.
- Integration tests:
  - [x] With a mocked handler (fetch stub), a successful submit resolves and
        triggers the expected query invalidations (`use-runs.test.tsx`).
- Test coverage target: >=80%
- All tests must pass

## Success Criteria
- All tests passing
- Test coverage >=80%
- `bun run lint`, `typecheck`, and `vitest` pass
- Adapter/hook mirror the cancel pattern and use generated types
