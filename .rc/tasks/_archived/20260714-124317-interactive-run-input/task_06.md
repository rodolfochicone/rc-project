---
status: completed
title: Run manager wiring, SendInput, and snapshot pending_input
type: backend
complexity: high
dependencies:
  - task_01
  - task_03
  - task_05
---

# Task 6: Run manager wiring, SendInput, and snapshot pending_input

## Overview
Wire the coordinator into the daemon's run lifecycle so a live run can receive
input. The `RunManager` creates a coordinator per run, stores it on `activeRun`,
threads it into the runtime config, exposes `SendInput` to deliver responses (the
write-path analog of `Cancel`), and surfaces the outstanding prompt as
`pending_input` on the run snapshot.

<critical>
- ALWAYS READ the PRD and TechSpec before starting
- REFERENCE TECHSPEC for implementation details — do not duplicate here
- FOCUS ON "WHAT" — describe what needs to be accomplished, not how
- MINIMIZE CODE — show code only to illustrate current structure or problem areas
- TESTS REQUIRED — every task MUST include tests in deliverables
</critical>

<requirements>
- `RunManager` MUST construct an `InputCoordinator` per run, store it on
  `activeRun`, and set it (plus the `Interactive` flag) on the run's
  `RuntimeConfig` before execution starts.
- MUST add `SendInput(ctx, runID, model.UserResponse) error` to the run service,
  mirroring `Cancel`: look up `getActive(runID)`; return a typed not-found error
  when the run is unknown/terminal and a typed not-awaiting error when no prompt is
  outstanding; otherwise call `coordinator.Submit`.
- The run snapshot MUST include a `pending_input` value while a prompt is
  outstanding and clear it once a response is delivered or the run terminates.
- MUST NOT introduce a new run status value (per ADR-003).
- Non-interactive runs MUST behave exactly as before (no coordinator effect).
</requirements>

## Subtasks
- [x] 6.1 Add an `inputCoordinator` field to `activeRun`; construct it in the start
      path and set it on the runtime config.
- [x] 6.2 Implement `SendInput` on the run manager with a typed not-awaiting error.
- [x] 6.3 Track the outstanding `PendingInput` per run (set on Await, clear on
      resolve); the coordinator emits `session.awaiting_input` on Await.
- [x] 6.4 Populate `pending_input` in the snapshot builder (event-driven).
- [x] 6.5 Add a backend end-to-end test exercising SendInput → awaiting run.

## As-built notes (for downstream tasks)
- **The coordinator now emits the event (the linchpin from tasks 04/05).**
  `newRunInputCoordinator(runID, scope.RunJournal(), logger)` wraps the task_03
  mailbox: `Await` emits `session.awaiting_input` via `submitSyntheticEvent` and
  records the outstanding prompt; `PendingInput()` exposes it; resolution clears
  it. The pure `newInputCoordinator()` mailbox is retained for task_03 tests.
- **Wiring:** `activeRun.inputCoordinator` is constructed in `startRun` from
  `scope.RunJournal()` and set on `runtimeCfg.InputCoordinator` before `runAsync`.
  (The `Interactive` flag still defaults false until task_07 sets it from the API.)
- **Snapshot `pending_input` is event-driven**, not live-state: the snapshot
  builder handles `session.awaiting_input` (sets `pendingInput`) and clears it on
  the next `session.update`. Known minor edge: a run canceled while waiting may
  still show `pending_input` in the snapshot (the UI keys off run status); ADR-003
  accepted "clear on next update / termination".
- **`RunManager.SendInput(ctx, runID, apicore.RunInput)`** mirrors `Cancel`:
  unknown run → GetRun error; terminal / not active / no outstanding prompt →
  `ErrRunNotAwaitingInput`; otherwise routes to `coordinator.Submit`.
- **DEFERRED TO TASK_07 (HTTP):** adding `SendInput` to the `RunService` interface
  (+ fixing the ~4 api/core test fakes), the `POST /runs/{id}/input` handler/route,
  and ALL OpenAPI work — `RunInput`, `RunPendingInput`/`pending_input` on the
  snapshot schema, the `interactive` exec-request flag, and codegen. The Go
  contract types (`contract.RunInput`, `contract.RunPendingInput`, the
  `RunSnapshot.PendingInput` field) already exist; only the OpenAPI spec + client
  regen remain.

## Implementation Details
Modify `internal/daemon/run_manager.go`: `activeRun` (:121), `newActiveRun`
(:1363), `startRun` (:1089), `getActive`/`setActive` (:1644), and add `SendInput`
mirroring `Cancel` (:1701). Populate `pending_input` in
`internal/daemon/run_snapshot.go`. Track the outstanding prompt either on the
coordinator or `activeRun` (whichever the TechSpec "Data Models" favors) so the
snapshot can read it. Reference the TechSpec "System Architecture" data flow.

### Relevant Files
- `internal/daemon/run_manager.go` — `activeRun`, start/registration, `Cancel`
  (analog), `getActive`.
- `internal/daemon/run_snapshot.go` — snapshot assembly; add `pending_input`.
- `internal/api/core/interfaces.go` — `RunService` interface (:111); add
  `SendInput`.
- `internal/core/model` — `InputCoordinator`/`UserResponse`/`PendingInput`
  (task_01).

### Dependent Files
- `internal/api/core/handlers.go` — the HTTP handler calls `SendInput` (task_07).
- `internal/core/run/exec/exec.go` — consumes the coordinator/flag set here
  (task_05).
- `openapi/rc-daemon.json` — snapshot schema gains `pending_input` (task_07).

### Related ADRs
- [ADR-003: Represent "awaiting input" as an event plus a snapshot field](../adrs/adr-003.md) — `pending_input` on the snapshot.
- [ADR-004: Interactivity is opt-in via a flag set at run start](../adrs/adr-004.md) — flag threaded into the runtime config.

## Deliverables
- `activeRun.inputCoordinator` + runtime-config wiring.
- `RunService.SendInput` with typed not-found / not-awaiting errors.
- `pending_input` populated/cleared in the snapshot.
- Unit tests with 80%+ coverage **(REQUIRED)**
- Integration (backend e2e) tests for SendInput → resume **(REQUIRED)**

## Tests
- Unit tests:
  - [ ] `SendInput` for an unknown run id returns the typed not-found error.
  - [ ] `SendInput` for a run with no outstanding prompt returns the typed
        not-awaiting error.
  - [ ] `SendInput` for an awaiting run calls `Submit` and returns success.
  - [ ] Snapshot shows `pending_input` while awaiting and nil after resolution.
- Integration tests:
  - [ ] Backend e2e: start an interactive run (fake ACP) → it awaits → `SendInput`
        with an option/text → the run resumes and `pending_input` clears.
- Test coverage target: >=80%
- All tests must pass

## Success Criteria
- All tests passing (including `-race`)
- Test coverage >=80%
- `SendInput` mirrors `Cancel` semantics (idempotent/typed errors)
- No new run status value added; snapshot reflects the waiting state
