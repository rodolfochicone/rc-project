# PRD: Interactive Run Input

## Overview

rc agents frequently need a human before they can keep going: to approve a
sensitive action, or to answer a question a workflow skill has asked. The escale
web UI watches runs but cannot answer them — it streams events read-only, and the
only write action is cancel. As a result the most valuable interactive
workflows — PRD ideation, TechSpec clarification, and approving the agent's
actions — cannot be driven from the web at all. A user who starts a "create PRD"
run sees the agent ask "Which problem should this solve first? A/B/C/D" and then
has no way to reply.

This feature gives a run a first-class **"Awaiting your response"** state and lets
the user answer the agent directly in the run detail view. When the agent
pauses — for a permission request or a skill question — the prompt appears with a
response area (clickable buttons when the agent offers discrete options, plus a
free-text box), the user responds, and the same run continues.

It is for anyone running rc workflows from the web UI (product/eng users driving
ideation, spec, and execution) who today must drop to the terminal to interact
with the agent.

## Goals

- Let a user answer an agent's question and approve or reject an agent's action
  without leaving the run detail view.
- Make "this run is waiting for me" obvious and unmissable.
- Cover both interaction types — permission requests and free-text skill
  questions — under one consistent interaction model in the MVP.
- Unblock the PRD, TechSpec, and ideation workflows on the web UI.

## User Stories

- As a product user running "create PRD" from the web, I want to answer the
  agent's clarifying questions in the run view so that the brainstorming actually
  progresses instead of stalling.
- As a user driving a TechSpec run, I want to pick one of the labeled options the
  agent offers with a single click so that answering is fast and unambiguous.
- As a user supervising an execution run, I want to approve or reject the agent's
  request to run a command or write a file so that I stay in control of sensitive
  actions instead of everything being auto-approved.
- As a user who answered with free text, I want the run to resume immediately and
  show the agent's next message so that the conversation feels continuous.
- As a user who stepped away, I want a paused run to keep waiting for me so that I
  do not lose progress, with the option to cancel if I no longer want it.

## Core Features

### 1. "Awaiting your response" run state (must-have)

A run enters a clearly labeled waiting state whenever the agent needs the user.
The run detail view shows the pending prompt prominently. The state persists
until the user responds or cancels the run.

- Functional: the run surfaces the question/permission prompt text the agent
  emitted.
- Functional: the run's status communicates "waiting on you" distinctly from
  running, completed, or failed.

### 2. Permission approval prompt (must-have)

When the agent requests permission for an action, the user sees the action being
requested and the available choices (e.g., allow once, allow always, reject) as
buttons, and selects one. The run proceeds according to the choice. This replaces
the current always-auto-approve behavior for user-observed runs.

### 3. Skill-question answering (must-have)

When a workflow skill asks a question and yields, the user can answer in the run
detail view and the run continues from that answer.

- Functional: when the question presents discrete labeled options (A/B/C/D), the
  UI renders them as clickable buttons.
- Functional: a free-text box is always available as a fallback, including when
  options are not detected and for open-ended questions.

### 4. Response affordance (must-have)

A single, consistent response area: option buttons when the agent offers discrete
choices, plus a free-text field. Submitting sends the answer to the running agent
and the live event stream shows the continuation.

### 5. Continuous transcript (must-have)

The user's answers and the agent's continued output appear in order in the run
transcript, so a paused-and-resumed run reads as one coherent conversation.

## User Experience

Primary persona: a rc web user driving a workflow (ideation/spec/execution).

Primary flow:

1. The user starts a run (e.g., create PRD) and watches it in the run detail
   view.
2. The agent asks a question or requests permission; the run enters "Awaiting
   your response" and the prompt is shown with a response area.
3. If the agent offered discrete options, the user clicks one; otherwise (or
   additionally) the user types an answer and submits.
4. The run resumes; the answer and the agent's next output stream into the
   transcript in order.
5. Steps 2–4 repeat as needed until the run completes.

UX considerations:

- The waiting state must be visually distinct and draw attention (the run is
  blocked on the user).
- Option buttons reduce friction for the common A/B/C/D case; the text box
  guarantees the user can always respond.
- Accessibility: the prompt and response controls are keyboard-navigable and use
  semantic controls; submitting via keyboard is supported.
- The response area is disabled or hidden when the run is not waiting, to avoid
  implying input is possible mid-stream.

## High-Level Technical Constraints

- Must integrate with the existing run detail view and live event stream rather
  than introducing a separate surface.
- Must preserve existing behavior for runs nobody is interacting with: a paused
  run waits indefinitely (it does not auto-fail or auto-cancel).
- Answers must reach the specific in-flight run reliably and resume it; ordering
  of the user answer relative to agent output must be preserved in the
  transcript.

## Non-Goals (Out of Scope)

- A separate chat/console surface distinct from the run detail view.
- Timeouts or automatic fallback behavior for unanswered prompts (deferred; MVP
  waits indefinitely).
- Special handling to keep headless/automated runs from blocking on input
  (Open Question; MVP targets user-observed runs).
- Editing or retracting an answer after it has been submitted.
- Multi-user coordination (two people answering the same run at once) and
  conflict resolution.
- Per-action permission policies, allowlists, or remembered approvals beyond what
  the agent itself offers as options.
- Bringing this interaction to the CLI/TUI (already interactive in the terminal).

## Phased Rollout Plan

### MVP (Phase 1)

- "Awaiting your response" run state surfaced in the run detail view.
- Permission approval prompts answered via ACP option buttons.
- Skill questions answered via detected option buttons plus a free-text fallback.
- Paused runs wait indefinitely; cancel remains available.
- Answers resume the run and stream into a continuous transcript.
- Success criteria to proceed: a user can complete an end-to-end "create PRD" or
  "create TechSpec" run entirely from the web UI, answering every prompt, and can
  approve/reject at least one agent action in an execution run.

### Phase 2

- Timeout and fallback policy for unanswered prompts (e.g., auto-proceed after a
  configurable delay), so automated/headless runs are safe.
- Handling for headless/unobserved runs (decide wait vs. auto-approve when nobody
  is attached).
- Success criteria to proceed: interactive runs are safe to start in automated
  contexts without hanging.

### Phase 3

- Richer permission UX (remembered approvals, per-workspace policies).
- Notifications when a run starts waiting (so the user can step away and be
  pulled back).

## Success Metrics

- Share of PRD/TechSpec runs started in the web UI that reach completion (vs.
  stalling at an unanswered question) increases materially after launch.
- Users can answer a prompt and see the run resume within a few seconds of
  submitting.
- Reduction in users dropping from the web UI to the terminal to drive
  interactive workflows.
- Of agent permission requests on observed runs, the share explicitly decided by
  a human (rather than silently auto-approved) rises from zero.

## Risks and Mitigations

- **Users do not notice a run is waiting and abandon it.** Mitigation: make the
  waiting state prominent in the run detail and run lists; consider notifications
  in a later phase.
- **Heuristic option detection misreads a question and shows wrong/no buttons.**
  Mitigation: the free-text box is always available, so the user can answer
  regardless; treat buttons as an enhancement.
- **Interactive runs started in automated contexts hang indefinitely.**
  Mitigation: scoped out of MVP and called out as an Open Question; Phase 2 adds
  timeout/headless policy.
- **Adoption risk if the interaction feels slower than the terminal.**
  Mitigation: one-click option buttons for the common case; keep the response
  path lightweight.

## Architecture Decision Records

- [ADR-001: Pause-and-resume interactive runs inside the run detail view](adrs/adr-001.md)
  — Surface and answer both permission requests and skill questions in the
  existing run detail view, with the run waiting indefinitely until answered.

## Open Questions

- How should runs with no human attached (headless, scheduled, automated) behave
  when the agent asks for input — keep waiting, or fall back to auto-approve?
  (Targeted for Phase 2.)
- Should free-text skill questions ever be answered with structured buttons the
  *client* defines, or only buttons the agent explicitly offered? (Affects how
  aggressive option detection should be.)
- For permission requests, which option set should be shown when the agent
  offers many — all of them, or a curated subset?
- Does a paused run need a visible indicator in the runs *list* (not just the
  detail) so users can find waiting runs? (Likely yes; confirm for MVP.)
