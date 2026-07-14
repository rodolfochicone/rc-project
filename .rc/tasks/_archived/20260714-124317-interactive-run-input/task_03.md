---
status: completed
title: Concrete input coordinator (channel mailbox)
type: backend
complexity: medium
dependencies:
  - task_01
---

# Task 3: Concrete input coordinator (channel mailbox)

## Overview
Implement the per-run `InputCoordinator` as a channel-based mailbox: one side
(`Await`) blocks until a response for a specific pending prompt arrives or the
context is cancelled; the other side (`Submit`) delivers a response from the HTTP
layer. This is the synchronization primitive that lets a blocked agent turn resume
on user input.

<critical>
- ALWAYS READ the PRD and TechSpec before starting
- REFERENCE TECHSPEC for implementation details — do not duplicate here
- FOCUS ON "WHAT" — describe what needs to be accomplished, not how
- MINIMIZE CODE — show code only to illustrate current structure or problem areas
- TESTS REQUIRED — every task MUST include tests in deliverables
</critical>

<requirements>
- MUST implement `model.InputCoordinator` with a compile-time assertion
  (`var _ model.InputCoordinator = (*T)(nil)`).
- `Await` MUST block until a `UserResponse` whose `PromptID` matches the awaited
  `PendingInput.ID` is submitted, or until `ctx` is done (returning `ctx.Err()`).
- `Submit` MUST return a descriptive error when no prompt with the given id is
  currently awaiting, and MUST NOT block.
- MUST be safe under concurrent `Await`/`Submit` (guard shared state; pass
  `-race`).
- MUST register exactly one awaiting slot per `PendingInput.ID` and clean it up on
  resolution or cancellation.
</requirements>

## Subtasks
- [x] 3.1 Implement the mailbox with per-prompt-id waiter registration.
- [x] 3.2 Implement `Await` with `select` on the response channel and `ctx.Done()`.
- [x] 3.3 Implement `Submit` with id matching and a no-waiter error.
- [x] 3.4 Ensure cleanup of the waiter on both resolution and cancellation.
- [x] 3.5 Add the compile-time interface assertion.

## As-built notes (for downstream tasks)
- Implemented as unexported `inputCoordinator` in `internal/daemon/run_input.go`
  (the daemon owns `activeRun` and constructs it via `newInputCoordinator`; the
  executor/agent consume it through the `model.InputCoordinator` interface, so no
  import cycle). Buffered (cap 1) channel per waiter; `sync.Mutex`-guarded waiter
  map; one waiter per `PendingInput.ID` (duplicate `Await` errors).
- Tests use scheduler-yield polling (`runtime.Gosched`) instead of sleeps to keep
  the concurrent Await/Submit cases deterministic under `-race`.

## Implementation Details
Place the implementation in the package that owns per-run runtime wiring so the
daemon can construct it and store it on `activeRun` (task_06). A small `sync.Mutex`
guarding a `map[string]chan UserResponse` is sufficient — see the TechSpec
"Component Overview" (Input coordinator). Do not add a new top-level package; add a
file to an existing runtime package.

### Relevant Files
- `internal/core/model` — the interface and value types from task_01.
- `internal/daemon/run_manager.go` — the eventual owner that constructs and stores
  the coordinator (task_06); keep the impl importable from here.

### Dependent Files
- `internal/core/agent/client.go` — calls `Await` from the permission callback
  (task_04).
- `internal/core/run/exec/exec.go` — calls `Await` from the turn loop (task_05).
- `internal/api/core` / `internal/daemon` — call `Submit` from the HTTP path
  (task_06/07).

### Related ADRs
- [ADR-002: Block-in-callback permissions + live-session re-prompt](../adrs/adr-002.md) — the coordinator is the broker that unblocks the callback/loop.

## Deliverables
- Concrete `InputCoordinator` implementation with constructor.
- Compile-time interface assertion.
- Unit tests with 80%+ coverage **(REQUIRED)**
- Integration tests covering concurrent submit/await **(REQUIRED)**

## Tests
- Unit tests:
  - [ ] `Await` returns the submitted `UserResponse` when `Submit` matches the
        awaited prompt id.
  - [ ] `Await` returns `ctx.Err()` when the context is cancelled before any submit.
  - [ ] `Submit` with an id that is not awaiting returns a descriptive error.
  - [ ] A second `Submit` for an already-resolved id returns the no-waiter error
        (no panic on closed channel).
- Integration tests:
  - [ ] Concurrent `Await` (prompt A) and `Submit` (prompt A) from separate
        goroutines resolves deterministically under `-race`.
- Test coverage target: >=80%
- All tests must pass

## Success Criteria
- All tests passing (including `-race`)
- Test coverage >=80%
- No goroutine leak: every `Await` returns on resolution or cancellation
- Implements `model.InputCoordinator` (assertion compiles)
