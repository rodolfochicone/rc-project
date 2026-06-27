---
status: completed
title: Interactive ACP permission callback and live-session continue
type: backend
complexity: high
dependencies:
  - task_01
  - task_02
---

# Task 4: Interactive ACP permission callback and live-session continue

## Overview
Make the ACP client capable of pausing for the user. When a run is interactive,
`RequestPermission` blocks on the input coordinator (emitting an
`awaiting_input` event) instead of auto-approving, and a new live-session continue
method re-prompts the already-open session so a multi-turn conversation can
proceed without persisted resume.

<critical>
- ALWAYS READ the PRD and TechSpec before starting
- REFERENCE TECHSPEC for implementation details — do not duplicate here
- FOCUS ON "WHAT" — describe what needs to be accomplished, not how
- MINIMIZE CODE — show code only to illustrate current structure or problem areas
- TESTS REQUIRED — every task MUST include tests in deliverables
</critical>

<requirements>
- When `Interactive` is true and a coordinator is present, `RequestPermission`
  MUST emit a `session.awaiting_input` event (kind `permission`) carrying the ACP
  options mapped to `InputOption`, then block on `coordinator.Await`, and return
  the selected option as `NewRequestPermissionOutcomeSelected`, or
  `NewRequestPermissionOutcomeCancelled` when the response is `Cancelled`.
- When `Interactive` is false (or no coordinator), `RequestPermission` MUST keep
  the current auto-approve-first-option behavior unchanged.
- MUST add a method that sends another prompt turn on the existing live session
  (`Session.Continue` / equivalent) that streams updates into the same session
  channel, per ADR-002.
- MUST cancel/return cleanly if the run context ends while awaiting (no leaked
  goroutine, no deadlock on the ACP connection).
- MUST map ACP permission request fields (option id, label) into `PendingInput`.
</requirements>

## Subtasks
- [x] 4.1 Thread the coordinator + interactive flag into the client (via
      `ClientConfig` → `clientImpl`; RuntimeConfig→ClientConfig wiring deferred to
      task_05/06 — see notes).
- [x] 4.2 Branch `RequestPermission`: interactive path awaits the coordinator;
      non-interactive path unchanged.
- [x] 4.3 Map `acp.RequestPermissionRequest.Options` ↔ `model.InputOption` and the
      response ↔ outcome.
- [x] 4.4 Add the live-session continue method that re-prompts the same `SessionId`.
- [x] 4.5 Honor context cancellation while blocked in the callback.

## As-built notes (deviations for downstream tasks)
- **`Continue` is exposed via a separate `agent.SessionContinuer` interface, NOT
  added to `agent.Client`.** Adding it to `Client` broke ~10 fake clients across
  the executor/exec/cli test suites. The optional-capability interface (idiomatic
  Go) keeps those fakes compiling; the task_05 turn loop type-asserts
  `client.(agent.SessionContinuer)`.
- **The client does NOT emit the `session.awaiting_input` event.** The ACP client
  has no journal/event handle (events are emitted by the acpshared
  `SessionUpdateHandler`). Interactive `RequestPermission` only calls
  `coordinator.Await`. The single emission point is `coordinator.Await` itself —
  the daemon wires the coordinator with the run journal in task_06, so both the
  permission path (client → Await) and the question path (executor → Await) emit
  one consistent event. **Task_06 must give the coordinator the journal and emit
  `session.awaiting_input` on Await (and clear `pending_input` on Submit).**
- **RuntimeConfig → ClientConfig threading is deferred to task_05/06.** This task
  added `agent.ClientConfig.{Interactive,InputCoordinator}` and wires them into
  `clientImpl`; the chain `model.RuntimeConfig → runshared.Config →
  agent.ClientConfig` (in `acpshared.createACPClient`) is wired when the exec loop
  and run-manager land.
- **`Continue` returns a NEW `Session` per turn** for the same ACP session id
  (the prior turn's wrapper finishes and closes its channel in `runPrompt`). It
  mirrors `ResumeSession` minus `LoadSession`, re-prompting the live connection.
- Spelling: `UserResponse.Canceled` (one `l`); tests avoid the ACP SDK's British
  `Outcome.Cancelled` token (misspell US) via a `selectedOption` helper.

## Implementation Details
Modify `RequestPermission` in `internal/core/agent/client.go:523` and add the
continue method near `runPrompt`/`CreateSession` (`client.go:277`, `:785`). The
permission callback is invoked synchronously while `c.conn.Prompt()` is in flight
(see TechSpec "System Architecture"), so blocking there pauses the agent turn. The
continue method calls `c.conn.Prompt()` again on the live `SessionId` and relies on
the existing `SessionUpdate` routing (`client.go:539`) to stream into the same
session. Reference the TechSpec "Core Interfaces" for the continue signature.

### Relevant Files
- `internal/core/agent/client.go` — `RequestPermission` (:523), `runPrompt` (:785),
  `CreateSession` (:277), `SessionUpdate` (:539).
- `internal/core/agent/session.go` — `sessionImpl`/`Session`; the continue method
  surface and where updates are published.
- `internal/core/model` — `InputCoordinator`, `PendingInput`, `UserResponse`
  (task_01).
- `pkg/rc/events/kinds/session.go` — `SessionAwaitingInputPayload` (task_02).

### Dependent Files
- `internal/core/run/exec/exec.go` — drives `Continue` in the turn loop (task_05).
- `internal/core/run/internal/acpshared/command_io.go` — constructs the client and
  passes session options (`SetupSessionExecution` :119, `createACPSession` :301);
  must forward the coordinator/flag.

### Related ADRs
- [ADR-002: Block-in-callback permissions + live-session re-prompt](../adrs/adr-002.md) — the core mechanism this task implements.
- [ADR-004: Interactivity is opt-in via a flag set at run start](../adrs/adr-004.md) — the gate around the interactive branch.

## Deliverables
- Interactive `RequestPermission` branch with event emission and `Await`.
- Live-session continue method.
- ACP option ↔ `InputOption` mapping.
- Unit tests with 80%+ coverage **(REQUIRED)**
- Integration tests with a fake/mock ACP connection **(REQUIRED)**

## Tests
- Unit tests:
  - [ ] Interactive `RequestPermission` returns `Selected(optionID)` when the
        coordinator yields a `UserResponse` with that `OptionID`.
  - [ ] Interactive `RequestPermission` returns `Cancelled` when the response has
        `Cancelled=true`.
  - [ ] Non-interactive `RequestPermission` still selects the first option (no
        coordinator interaction).
  - [ ] `RequestPermission` returns/cancels cleanly when the run context is
        cancelled while awaiting.
  - [ ] ACP options with two entries map to two `InputOption`s preserving id/label.
- Integration tests:
  - [ ] With a fake ACP connection, `Continue` sends a second prompt on the same
        session id and the resulting updates arrive on the same session channel in
        order.
- Test coverage target: >=80%
- All tests must pass

## Success Criteria
- All tests passing (including `-race`)
- Test coverage >=80%
- Non-interactive behavior is byte-for-byte unchanged
- A blocked permission callback unblocks on submit and on context cancel
