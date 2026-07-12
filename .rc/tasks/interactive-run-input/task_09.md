---
status: completed
title: Frontend response panel and option parsing
type: frontend
complexity: high
dependencies:
  - task_08
---

# Task 9: Frontend response panel and option parsing

## Overview
Build the UI that lets a user answer a paused run inside the run detail view. When
a run is awaiting input, a response panel appears showing the prompt and a
response area: option buttons (ACP options for permissions, or A/B/C/D parsed from
a skill question) plus a free-text box that is always available. Submitting calls
the mutation from task_08 and the live stream shows the continuation.

<critical>
- ALWAYS READ the PRD and TechSpec before starting
- REFERENCE TECHSPEC for implementation details — do not duplicate here
- FOCUS ON "WHAT" — describe what needs to be accomplished, not how
- MINIMIZE CODE — show code only to illustrate current structure or problem areas
- TESTS REQUIRED — every task MUST include tests in deliverables
</critical>

<requirements>
- MUST render a response panel in the run detail when the run is awaiting input,
  derived from `snapshot.pending_input` and/or the live `session.awaiting_input`
  event; the panel MUST be hidden when the run is not awaiting.
- For permission prompts, MUST render the ACP-supplied options as buttons.
- For skill questions, MUST heuristically parse labeled options (e.g. `A) … B) …`)
  into clickable buttons, with a free-text box ALWAYS available as a fallback.
- Submitting a button or text MUST call `useSendRunInput` with the matching
  `prompt_id` and the selected `option_id`/`text`; the panel MUST disable while the
  mutation is in flight.
- MUST render `user_message_chunk` content in transcript order so the answer shows
  in the conversation.
- The option parser MUST be a pure, unit-tested function isolated from rendering.
</requirements>

## Subtasks
- [x] 9.1 Add a pure option-parser util (text → ordered options + remainder).
- [x] 9.2 Build the response panel: prompt text, option buttons, text box, submit.
- [x] 9.3 Wire the panel into the run detail, gated on awaiting state.
- [x] 9.4 Render `user_message_chunk` in the transcript.
- [x] 9.5 Handle submit/disabled/error states via the mutation.

## As-built notes
- **Two pure utils under `lib/`**: `option-parser.ts` (`parseQuestionOptions` —
  needs ≥2 `A)`/`B.` markers to avoid false positives on prose) and
  `pending-input.ts` (`resolvePendingInput` + `parseAwaitingInput` — live
  `session.awaiting_input` events take precedence over `snapshot.pending_input`
  and are cleared by a later `session.update` or terminal run event, ADR-003).
- **`RunInputPanel`** (`components/run-input-panel.tsx`) is presentational with
  local text state; permission prompts submit `option_id`, parsed question options
  submit the letter as `text`, and the free-text box is ALWAYS rendered.
- **Gating** lives in `run-detail-view.tsx`: the panel is hidden when the run
  status is terminal (guards a stale snapshot `pending_input` on a canceled run)
  and when no prompt is outstanding. The route wires `useSendRunInput` and passes
  `onSendInput`/`isSendingInput`/`sendInputError`.
- **`user_message_chunk`** is coalesced into a user-role message in
  `run-transcript-panel.tsx`'s `mergeTranscriptMessages`. Required a fix to
  `convertMessage`: assistant-ui only accepts `status` on assistant messages, so
  user messages now render without it.

## Implementation Details
Add the panel and parser under `web/src/systems/runs/components` and wire it into
`run-detail-view.tsx`. Derive the awaiting state from `snapshot.pending_input` and
the live event in `run-transcript-panel.tsx`'s event handling
(`runUIMessageFromLiveEvent`). For permissions, options come structured from the
event payload; for questions, parse the text (frontend heuristic per ADR decision).
Reuse `@escaletech/ui` primitives (Button, SurfaceCard) consistent with existing
run components. See TechSpec "User Experience"/"API Endpoints".

### Relevant Files
- `web/src/systems/runs/components/run-detail-view.tsx` — host for the response
  panel.
- `web/src/systems/runs/components/run-transcript-panel.tsx` — live event handling
  (`runUIMessageFromLiveEvent`), `user_message_chunk` rendering.
- `web/src/systems/runs/hooks/use-run-stream.ts` / `lib/event-store.ts` — live
  `awaiting_input` events.
- `web/src/systems/runs/hooks/use-runs.ts` — `useSendRunInput` (task_08).
- `web/src/systems/runs/types.ts` — `pending_input` type (task_08).

### Dependent Files
- `web/src/systems/runs/components/run-detail-view.test.tsx` — extended coverage.
- `@escaletech/ui` — shared Button/Card primitives (consumed, not modified).

### Related ADRs
- [ADR-001: Pause-and-resume interactive runs inside the run detail view](../adrs/adr-001.md) — the response surface lives in the run detail.
- [ADR-003: Represent "awaiting input" as an event plus a snapshot field](../adrs/adr-003.md) — drives the panel from event/snapshot, not a status.

## Deliverables
- Pure option-parser util.
- Response panel component wired into the run detail.
- `user_message_chunk` transcript rendering.
- Unit tests with 80%+ coverage **(REQUIRED)**
- Integration tests for the awaiting → submit flow **(REQUIRED)**

## Tests
- Unit tests:
  - [x] Parser: `"A) Keep B) Plan C) Both"` yields three options with labels
        Keep/Plan/Both in order.
  - [x] Parser: free-form text with no options yields zero options (text box only).
  - [x] Panel renders ACP permission options as buttons from the event payload.
  - [x] Clicking option "B" calls `useSendRunInput` with the prompt id and the
        matching `option_id`/`text`.
  - [x] Submitting free text calls the mutation with `text` and disables while in
        flight.
  - [x] Panel is not rendered when the run is not awaiting input.
- Integration tests:
  - [x] Awaiting event arrives → panel shows → submit → mutation called and panel
        clears when `pending_input` resolves; `user_message_chunk` appears in
        transcript order (`run-detail-view.test.tsx`).
- Test coverage target: >=80%
- All tests must pass

## Success Criteria
- All tests passing
- Test coverage >=80%
- `bun run lint`, `typecheck`, and `vitest` pass
- The free-text box is always available even when option buttons render
