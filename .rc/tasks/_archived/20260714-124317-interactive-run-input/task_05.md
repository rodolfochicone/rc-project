---
status: completed
title: Interactive exec turn loop
type: backend
complexity: high
dependencies:
  - task_02
  - task_03
  - task_04
---

# Task 5: Interactive exec turn loop

## Overview
Turn an interactive exec run into a multi-turn conversation. Instead of finalizing
when the agent's turn ends, the executor emits an `awaiting_input` event for the
question, blocks on the input coordinator, and re-prompts the live session with the
user's answer — looping until the run is cancelled. Non-interactive runs keep
finalizing at `end_turn` exactly as today.

<critical>
- ALWAYS READ the PRD and TechSpec before starting
- REFERENCE TECHSPEC for implementation details — do not duplicate here
- FOCUS ON "WHAT" — describe what needs to be accomplished, not how
- MINIMIZE CODE — show code only to illustrate current structure or problem areas
- TESTS REQUIRED — every task MUST include tests in deliverables
</critical>

<requirements>
- When `RuntimeConfig.Interactive` is true, on a non-cancelled turn end the
  executor MUST emit `session.awaiting_input` (kind `question`, text = the agent's
  last message) and block on `coordinator.Await` instead of finalizing.
- On a `UserResponse` with text/option, the executor MUST re-prompt the live
  session via the continue method (task_04) and resume streaming.
- On `Cancelled` response or run-context cancellation, the loop MUST exit and the
  run MUST terminate cleanly.
- When `Interactive` is false, the executor MUST finalize at `end_turn` with the
  current behavior (regression guard).
- The client/session MUST stay open across turns; cleanup MUST still run on exit
  (no leaked session/client).
</requirements>

## Subtasks
- [x] 5.1 Read `Interactive`/coordinator from the runtime config in the exec path.
- [x] 5.2 Wrap the single-turn execution in a loop that pauses on `end_turn` when
      interactive.
- [x] 5.3 Surface the last assistant text as the question `PendingInput` handed to
      `coordinator.Await` (event emission lives in the coordinator — task_06).
- [x] 5.4 Await a response and re-prompt the live session with the answer.
- [x] 5.5 Exit cleanly on cancel; preserve non-interactive finalize behavior.

## As-built notes (for downstream tasks)
- **Threading landed here** (deferred from task_04): `runshared.Config` gained
  `Interactive` + `InputCoordinator` (copied in `NewConfig`), and
  `acpshared.createACPClient` forwards them into `agent.ClientConfig`.
- **Loop location:** `runSingleExecAttempt` (exec.go). After turn 1 completes
  successfully AND `interactiveExec(cfg)` (Interactive && coordinator != nil),
  `runInteractiveTurns` runs — gated so the non-interactive path is byte-for-byte
  unchanged (regression test `...NonInteractiveNeverAwaits`).
- **Per-turn execution:** `acpshared.ContinueSessionExecution` type-asserts the
  client to `agent.SessionContinuer`, calls `Continue` on the same ACP session id,
  and builds a fresh handler reusing the owner's log writers. Turn≥2 executions
  have nil Client/OutFile/ErrFile so their `Close` is a no-op — the owning
  execution (`defer execution.Close()`) keeps the client alive across turns and
  closes it once at the end.
- **Activity-watchdog fix:** the open-ended `Await` is wrapped in
  `activity.BeginActivity()/EndActivity()` so `TimeSinceLastActivity()` reports 0
  while waiting — otherwise the inactivity watchdog would cancel the run during a
  user pause. (`ActivityMonitor` methods are nil-safe.)
- **`session.awaiting_input` is NOT emitted by the executor.** Per task_04, the
  single emission point is `coordinator.Await`. The executor passes a question
  `PendingInput` (Kind=question, Text = last assistant message via
  `renderAssistantOutput`); **task_06 wires the coordinator with the run journal to
  emit the event and set/clear `pending_input`.**
- **Completion semantics (per ADR):** an interactive run loops until the user
  declines/cancels (`UserResponse.Canceled` or ctx end); it returns the latest
  successful turn result. There is no natural "all done → completed" signal in the
  MVP — the user ends the conversation by canceling.
</requirements>

## Implementation Details
Modify the exec execution path around `executeExecJob`
(`internal/core/run/exec/exec.go:242`, called from `ExecuteExec` :230). The session
is created in `internal/core/run/internal/acpshared/command_io.go`
(`createACPSession` :301) and must remain open for the loop's lifetime — keep the
existing `defer state.close()` as the final cleanup. Reference the TechSpec
"System Architecture" data-flow and ADR-002 for the loop shape. Extract the
last assistant message from the session update stream to populate the question
text.

### Relevant Files
- `internal/core/run/exec/exec.go` — `ExecuteExec` (:230), `executeExecJob` (:242),
  `finalizeExecResult`; where the loop is introduced.
- `internal/core/run/internal/acpshared/command_io.go` — session setup/lifetime
  (`SetupSessionExecution` :119, `createACPSession` :301).
- `internal/core/agent/session.go` — session updates channel feeding last-message
  extraction; continue method from task_04.
- `pkg/rc/events/kinds/session.go` — `SessionAwaitingInputPayload` (task_02).

### Dependent Files
- `internal/daemon/run_manager.go` — supplies the coordinator/flag via the runtime
  config (task_06); finalization/terminal status interplay.

### Related ADRs
- [ADR-002: Block-in-callback permissions + live-session re-prompt](../adrs/adr-002.md) — defines the question-resume mechanism.
- [ADR-004: Interactivity is opt-in via a flag set at run start](../adrs/adr-004.md) — gates the loop.

## Deliverables
- Interactive turn loop in the exec path with event emission and live re-prompt.
- Preserved non-interactive finalize path.
- Unit tests with 80%+ coverage **(REQUIRED)**
- Integration tests driving multiple turns via a fake ACP **(REQUIRED)**

## Tests
- Unit tests:
  - [ ] Interactive run: on first `end_turn`, an `awaiting_input` (kind question)
        event is emitted carrying the last assistant message text.
  - [ ] Interactive run: a submitted text answer triggers exactly one live
        re-prompt with that text.
  - [ ] Interactive run: a `Cancelled` response exits the loop and terminates the
        run without another re-prompt.
  - [ ] Non-interactive run: finalizes at `end_turn` with no `awaiting_input` event
        (regression guard).
- Integration tests:
  - [ ] Fake ACP two-turn flow: turn 1 ends → await → submit "B" → turn 2 streams
        and the transcript contains both turns in order.
- Test coverage target: >=80%
- All tests must pass

## Success Criteria
- All tests passing (including `-race`)
- Test coverage >=80%
- Session/client are closed exactly once after the loop exits
- Non-interactive exec behavior is unchanged
