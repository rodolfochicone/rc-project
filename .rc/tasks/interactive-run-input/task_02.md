---
status: completed
title: Awaiting-input event kind and payload
type: backend
complexity: low
dependencies: []
---

# Task 2: Awaiting-input event kind and payload

## Overview
Introduce the `session.awaiting_input` event so a paused run can tell observers
exactly what it is waiting for. This event carries the pending prompt — its
correlation id, kind (permission or question), text, and, for permissions, the
ACP-supplied options — and is the signal the UI renders.

<critical>
- ALWAYS READ the PRD and TechSpec before starting
- REFERENCE TECHSPEC for implementation details — do not duplicate here
- FOCUS ON "WHAT" — describe what needs to be accomplished, not how
- MINIMIZE CODE — show code only to illustrate current structure or problem areas
- TESTS REQUIRED — every task MUST include tests in deliverables
</critical>

<requirements>
- MUST add `EventKindSessionAwaitingInput = "session.awaiting_input"` alongside the
  existing session event kinds in `pkg/rc/events`.
- MUST add `SessionAwaitingInputPayload` (and any nested option/request types) to
  `pkg/rc/events/kinds/session.go` with fields per the TechSpec "Data Models".
- MUST follow the existing payload conventions in that package (JSON tags,
  documented exported fields, `Index` for ordering).
- MUST NOT change or repurpose existing event kinds; this is additive.
</requirements>

## Subtasks
- [x] 2.1 Add the `EventKindSessionAwaitingInput` constant in the events kind list.
- [x] 2.2 Define `SessionAwaitingInputPayload` with correlation id, kind, text, and
      options.
- [x] 2.3 Wire the payload into the documented payload API the same way sibling
      session payloads are exposed.
- [x] 2.4 Confirm `ToolCallStateWaitingForConfirmation` is reused where a tool-call
      waiting state is relevant (do not add a duplicate state).

## As-built notes (for downstream tasks)
- Kind values exposed as constants in the public `kinds` package:
  `kinds.AwaitingInputKindPermission` / `kinds.AwaitingInputKindQuestion`. The
  `internal/core/model` side has equivalents (`model.PendingInputKind*`); the
  agent/executor map model → kinds string by these matching values (the public
  `kinds` package must not import `internal`).
- `SessionAwaitingInputPayload` lives in `pkg/rc/events/kinds/session.go`; the new
  kind is documented in `docs/events.md` and enforced by `docs_test.go`.
- `ToolCallStateWaitingForConfirmation` was left as-is (not duplicated). The
  awaiting-input pause is modeled as a session-level event, distinct from a tool
  call's lifecycle state, so no tool-call-state change was needed.

## Implementation Details
Add the constant in the session events block of `pkg/rc/events/event.go` and the
payload type in `pkg/rc/events/kinds/session.go`, mirroring `SessionUpdatePayload`
and friends. The payload field shapes are specified in the TechSpec "Data Models"
section — reference it rather than redefining shapes here.

### Relevant Files
- `pkg/rc/events/event.go` — `EventKind` constants (session block) where the new
  kind is declared.
- `pkg/rc/events/kinds/session.go` — session payload types; new payload added here
  next to `SessionUpdatePayload`.

### Dependent Files
- `internal/core/agent/client.go` — emits the event in task_04.
- `internal/core/run/exec/exec.go` — emits the event for questions in task_05.
- `web/src/systems/runs` — consumes the event shape in task_09 (via the raw event
  stream).

### Related ADRs
- [ADR-003: Represent "awaiting input" as an event plus a snapshot field](../adrs/adr-003.md) — this event is half of that representation.

## Deliverables
- `EventKindSessionAwaitingInput` constant.
- `SessionAwaitingInputPayload` type with nested option type.
- Unit tests with 80%+ coverage **(REQUIRED)**
- Integration tests for payload (de)serialization **(REQUIRED)**

## Tests
- Unit tests:
  - [ ] `SessionAwaitingInputPayload` with `Kind="permission"` and two options
        marshals to JSON with the expected `option_id`/`label` fields.
  - [ ] `SessionAwaitingInputPayload` with `Kind="question"` and empty options
        marshals without an options array (omitempty honored).
  - [ ] A marshaled payload unmarshals back to an equal struct (round-trip).
- Integration tests:
  - [ ] An event envelope with `kind="session.awaiting_input"` carrying the payload
        passes the events package's envelope encode/decode path.
- Test coverage target: >=80%
- All tests must pass

## Success Criteria
- All tests passing
- Test coverage >=80%
- The new kind and payload are exported and documented
- No existing event kind or payload is modified
