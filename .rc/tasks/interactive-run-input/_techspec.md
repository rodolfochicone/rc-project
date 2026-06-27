# TechSpec: Interactive Run Input

## Executive Summary

This feature lets a user answer a running ACP agent from the escale run detail
view — both permission requests and skill questions — and have the run continue.
The design rests on one fact from the codebase: the ACP connection and session
object stay alive for the duration of a job. Permission requests are handled by
**blocking inside the `RequestPermission` callback** while the agent's turn is in
flight; skill questions (which end the turn) are handled by **re-prompting the
same live session** with the user's answer in an executor turn loop. A per-run
**input coordinator** (a channel mailbox stored on the daemon's `activeRun`)
brokers responses from a new `POST /api/runs/{run_id}/input` endpoint into the
blocked callback / paused loop — mirroring how `Cancel` reaches a live run today.
The "waiting" state is carried by a new `session.awaiting_input` event plus a
`pending_input` snapshot field, with no new run status enum value.

The primary trade-off: interactivity is **opt-in** per run and an interactive run
does not reach `completed` naturally in the MVP — every `end_turn` pauses, and the
run ends when the user cancels. This keeps the lifecycle change contained and
leaves automated runs untouched, at the cost of a less "finished-looking" terminal
state for interactive runs (tracked in Known Risks).

Maps to PRD goals: answering questions and approving actions in the run detail
(Core Features 1–4), continuous transcript (Core Feature 5), opt-in/observed-run
scope (Non-Goals, Phased Rollout).

## System Architecture

### Component Overview

- **Input coordinator** (`internal/core/model` interface + concrete impl on the
  daemon side): a per-run mailbox. `Await(ctx, PendingInput)` blocks the
  caller (permission callback or turn loop) until a matching response arrives or
  ctx is cancelled; `Submit(response)` delivers a response from the HTTP layer.
- **ACP client** (`internal/core/agent`): interactive `RequestPermission` blocks
  on the coordinator instead of auto-approving; a new `Session.Continue` re-prompts
  the live session for the next turn.
- **Exec executor** (`internal/core/run/exec`): for interactive runs, a turn loop
  that emits `awaiting_input` on `end_turn`, blocks on the coordinator, and
  re-prompts the live session with the answer; ends on cancel.
- **Run manager** (`internal/daemon`): creates the coordinator, stores it on
  `activeRun`, threads it through `RuntimeConfig`, and exposes `SendInput`.
- **HTTP API** (`internal/api/core`): `POST /runs/{run_id}/input` handler →
  `RunService.SendInput` → `RunManager` → coordinator.
- **Events/snapshot** (`pkg/rc/events`, snapshot builder + OpenAPI): the
  `session.awaiting_input` event and the `pending_input` snapshot field.
- **Web UI** (`web/src/systems/runs`): a response panel in the run detail that
  reads `pending_input` / the live event, renders option buttons (ACP options or
  parsed A/B/C/D) plus a text box, and POSTs the answer.

### Data Flow

Agent pauses → (permission: `RequestPermission` blocks; question: turn ends and
loop blocks) → `awaiting_input` event emitted + `pending_input` set on snapshot →
streamed to UI → user submits → `POST /runs/{id}/input` → `RunManager.SendInput`
→ `coordinator.Submit` → blocked `Await` returns → agent continues (same turn for
permission; `Session.Continue` for questions) → session updates stream into the
transcript.

## Implementation Design

### Core Interfaces

Per-run broker between the HTTP layer and the paused agent (interface in
`internal/core/model`, keeping `model` dependency-free):

```go
// InputCoordinator brokers user responses to a run paused for input.
type InputCoordinator interface {
    // Await blocks until a response for prompt.ID arrives or ctx is done.
    // Called from the permission callback and the exec turn loop.
    Await(ctx context.Context, prompt PendingInput) (UserResponse, error)
    // Submit delivers a response from the HTTP layer; errors if no prompt
    // with that ID is currently awaiting.
    Submit(resp UserResponse) error
}

type PendingInput struct {
    ID      string         // correlation id, unique per pause
    Kind    string         // "permission" | "question"
    Text    string         // permission/question text shown to the user
    Options []InputOption  // permission options (ACP); empty for questions
}

type InputOption struct {
    OptionID string // ACP OptionId
    Label    string
}

type UserResponse struct {
    PromptID  string // must match the PendingInput.ID being awaited
    OptionID  string // selected permission option (optional)
    Text      string // free-text answer (optional)
    Cancelled bool
}
```

Live-session continuation on the agent session (in `internal/core/agent`):

```go
// Continue sends another prompt turn on the already-open live session and
// streams updates into the same session channel. Used for multi-turn input.
Continue(ctx context.Context, prompt model.Prompt) error
```

### Data Models

- `PendingInput`, `InputOption`, `UserResponse`, `InputCoordinator` — above
  (`internal/core/model`).
- `model.RuntimeConfig` gains `Interactive bool` and `InputCoordinator
  InputCoordinator` (nil/false ⇒ current behavior).
- `activeRun` (`internal/daemon/run_manager.go`) gains an `inputCoordinator`
  field, created in `newActiveRun`/`startRun`.
- Event payload `SessionAwaitingInputPayload` in
  `pkg/rc/events/kinds/session.go` (Index, ACPSessionID, Kind, Text, Options,
  PromptID). New constant `EventKindSessionAwaitingInput = "session.awaiting_input"`.
- Snapshot payload gains `pending_input` (PromptID, Kind, Text, Options),
  populated while a prompt is outstanding, cleared on response/termination.
- Exec/start request gains `interactive bool`.

### API Endpoints

- **POST `/api/runs/{run_id}/input`** — submit a response to a paused run.
  - Request `RunInputRequest`: `{ "prompt_id": string, "option_id"?: string,
    "text"?: string, "cancelled"?: bool }`. One of `option_id` / `text` /
    `cancelled` required.
  - Responses: `202 AcceptedResponse` (delivered); `404` (no run / no matching
    awaiting prompt); `409` (run not currently awaiting input); `403`/`500` per
    existing conventions. Declared in `openapi/rc-daemon.json` mirroring
    `cancelRun`, then `node scripts/codegen.mjs` regenerates the TS client.
- **GET `/api/runs/{run_id}/snapshot`** — unchanged path; response now includes
  optional `pending_input`.
- **POST `/api/exec`** — request body gains optional `interactive` (default
  false).

## Integration Points

No new external services. Internal boundaries: HTTP (`internal/api/core`) →
`RunService` (`internal/api/core/interfaces.go`, add `SendInput`) →
`RunManager` (`internal/daemon`) → `model.RuntimeConfig` → executor
(`internal/core/run/exec`) → ACP client (`internal/core/agent`). The ACP SDK
(`github.com/coder/acp-go-sdk`) permission types are consumed only inside the
agent client and mapped to `InputOption`/`UserResponse`.

## Impact Analysis

| Component | Impact Type | Description and Risk | Required Action |
|-----------|-------------|----------------------|-----------------|
| `internal/core/model` | modified | New interface + types; `RuntimeConfig` fields. Low risk (additive). | Add `InputCoordinator`, `PendingInput`, `UserResponse`, config fields |
| `pkg/rc/events` (+ kinds) | modified | New event kind + payload. Low risk (additive). | Add `EventKindSessionAwaitingInput` + `SessionAwaitingInputPayload` |
| `internal/core/agent/client.go` | modified | Interactive `RequestPermission` blocks on coordinator; new `Continue`. Medium risk — touches the live ACP turn. | Gate on `Interactive`; emit event; block on `Await`; add live re-prompt |
| `internal/core/run/exec` | modified | Turn loop holding session open; no finalize on `end_turn` when interactive. Medium/high risk — lifecycle change. | Add interactive loop around `executeExecJob` |
| `internal/daemon/run_manager.go` | modified | `activeRun.inputCoordinator`; `SendInput`. Low/medium risk. | Construct coordinator; implement `SendInput`; thread into config |
| `internal/api/core` (handlers/routes/interfaces) | modified | New endpoint + `RunService.SendInput`. Low risk (mirrors cancel). | Add handler, route, interface method |
| snapshot builder + `openapi/rc-daemon.json` | modified | `pending_input` field + `RunInputRequest` schema + `interactive`. Low risk. | Edit spec; populate snapshot; run codegen |
| `web/src/systems/runs` | modified | Response panel, adapter, hook, option parser. Low risk (additive UI). | Build input UI + mutation |
| `web/src/systems/exec` | modified | Optional `interactive` on start. Low risk. | Add flag to start request |

## Testing Approach

### Unit Tests

- **Input coordinator (Go, table-driven, `-race`)**: `Await`/`Submit`
  happy path; `Submit` with no awaiting prompt → error; ctx cancellation unblocks
  `Await`; mismatched `PromptID`; concurrent submit/await.
- **Interactive `RequestPermission`**: with coordinator returns the selected
  option; cancelled → `Cancelled` outcome; without coordinator (non-interactive)
  keeps current auto-approve. Mock the coordinator via the interface.
- **Exec turn loop**: on `end_turn` with `interactive`, emits `awaiting_input`
  and calls `Continue` with the answer; on cancel, exits cleanly; non-interactive
  finalizes at `end_turn` (regression guard).
- **`RunManager.SendInput`**: routes to the right `activeRun`; unknown run / not
  awaiting → typed errors mapped to 404/409.
- **Snapshot**: `pending_input` present while awaiting, cleared after response.
- **Frontend (vitest)**: option parser detects `A) / B)` lists and ignores
  non-option text; response panel renders ACP option buttons, parsed buttons, and
  the text fallback; submitting calls the mutation with the right payload;
  `user_message_chunk` renders in order.

### Integration Tests

- **OpenAPI contract test** (`internal/api/httpapi/openapi_contract_test.go`):
  the new route and `RunInputRequest` schema match the spec.
- **End-to-end (Go, fake ACP client)**: drive an interactive exec run through a
  permission pause and a question pause, submit responses via `SendInput`, assert
  the turn continues and the transcript is ordered.

## Development Sequencing

### Build Order

1. **Model types** — `InputCoordinator`, `PendingInput`, `InputOption`,
   `UserResponse`, and `RuntimeConfig.{Interactive,InputCoordinator}` in
   `internal/core/model`. No dependencies.
2. **Event kind + payload** — `EventKindSessionAwaitingInput` and
   `SessionAwaitingInputPayload` in `pkg/rc/events`. No dependencies.
3. **Concrete coordinator** — channel-mailbox implementation + unit tests.
   Depends on step 1.
4. **ACP client changes** — interactive `RequestPermission` (emit event from step
   2, block via step 1) and `Session.Continue` live re-prompt. Depends on steps
   1, 2.
5. **Exec turn loop** — interactive pause/resume around `executeExecJob`. Depends
   on steps 2, 3, 4.
6. **RunManager wiring** — create coordinator, store on `activeRun`, thread into
   `RuntimeConfig`, implement `SendInput`; populate `pending_input` in the
   snapshot. Depends on steps 1, 3, 5.
7. **HTTP + OpenAPI** — `RunService.SendInput`, handler, route, `RunInputRequest`
   schema, `pending_input` + `interactive` in the spec; run codegen. Depends on
   step 6.
8. **Frontend** — adapter `sendRunInput`, `useSendRunInput`, response panel +
   option parser, `interactive` on exec start, `user_message_chunk` rendering.
   Depends on step 7.
9. **Integration tests + `make verify` (Go) and frontend lint/typecheck/test.**
   Depends on all prior steps.

### Technical Dependencies

- ACP SDK permission types (`github.com/coder/acp-go-sdk`) — already vendored.
- `make verify` (fmt/lint/test/build) and the frontend codegen toolchain.

## Monitoring and Observability

- Structured `slog` events (with `run_id`, `prompt_id`, `kind`) when a run starts
  awaiting input, when a response is delivered, and when an awaited prompt is
  cancelled.
- The `session.awaiting_input` event is itself observable in the run event feed.
- Metric-friendly fields: time spent awaiting per prompt (emit start/resolve
  timestamps) to inform a future timeout policy (Phase 2).

## Technical Considerations

### Key Decisions

- **Decision**: Block-in-callback for permissions, live-session re-prompt for
  questions ([ADR-002](adrs/adr-002.md)).
  - **Rationale**: The connection is already open; avoids persisted-resume and a
    new run status.
  - **Trade-offs**: Executor holds the session open across turns; interactive runs
    end via cancel rather than a natural `completed`.
  - **Alternatives rejected**: persisted `ResumeSession` per turn;
    everything-through-permission.
- **Decision**: Event + snapshot field, not a new run status
  ([ADR-003](adrs/adr-003.md)).
- **Decision**: Opt-in `interactive` flag ([ADR-004](adrs/adr-004.md)).
- **Decision**: Option detection in the frontend; permissions carry ACP options
  in the event payload.

### Known Risks

- **Interactive runs never reach `completed`.** MVP treats every `end_turn` as a
  pause; the run ends on cancel. Likelihood: certain. Mitigation: document the
  behavior; a future signal (artifact written / agent "done" marker) can finalize
  naturally — Open Question.
- **A blocked permission callback ties up the ACP connection indefinitely.**
  Mitigation: the wait is bound to the run context; cancel unblocks it.
- **Stale `pending_input`** if the resolve signal is missed. Mitigation: clear on
  the next session update and on termination; newer events supersede.
- **Frontend option parsing is heuristic.** Mitigation: text box is always the
  fallback (PRD-accepted).
- **Headless interactive runs hang.** Out of MVP scope via opt-in; Phase 2 adds a
  timeout/headless policy.

## Architecture Decision Records

- [ADR-001: Pause-and-resume interactive runs inside the run detail view](adrs/adr-001.md)
  — Surface and answer both permission requests and skill questions in the run
  detail; wait indefinitely.
- [ADR-002: Block-in-callback permissions + live-session re-prompt for multi-turn](adrs/adr-002.md)
  — Block the permission callback within the turn; re-prompt the live session for
  questions instead of persisted resume.
- [ADR-003: Represent "awaiting input" as an event plus a snapshot field](adrs/adr-003.md)
  — No new run status; a `session.awaiting_input` event and `pending_input`
  snapshot field drive the UI.
- [ADR-004: Interactivity is opt-in via a flag set at run start](adrs/adr-004.md)
  — Only runs started with `interactive=true` block for input; all others keep
  current auto-approve/finalize behavior.
