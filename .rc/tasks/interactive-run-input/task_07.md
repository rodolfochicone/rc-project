---
status: completed
title: HTTP input endpoint and OpenAPI surface
type: backend
complexity: medium
dependencies:
  - task_06
---

# Task 7: HTTP input endpoint and OpenAPI surface

## Inherited scope from task_06 (read first)
Task_06 deliberately stopped at the daemon boundary. This task additionally owns:
- Add `SendInput(ctx, runID, RunInput) error` to the `apicore.RunService`
  interface and add a stub to the ~4 api/core test fakes (handlers_test,
  handlers_contract_test, handlers_service_errors_test, transport_integration_test).
  `RunManager.SendInput` already exists with the matching signature.
- The Go contract types already exist: `contract.RunInput`, `contract.RunPendingInput`,
  `contract.RunInputOption`, and `RunSnapshot.PendingInput` (+ apicore aliases). Only
  the OpenAPI declarations remain: the `RunInput` request body, `pending_input` on
  the snapshot schema, the `interactive` flag on the exec request, the new endpoint,
  then `node scripts/codegen.mjs`.
- Map the daemon's `ErrRunNotAwaitingInput` to HTTP 409 and the GetRun
  not-found error to 404 in the handler.

## Overview
Expose the input path over HTTP so the web UI can answer a paused run. Add
`POST /api/runs/{run_id}/input`, declare it and the new request/response shapes in
the OpenAPI spec (plus the `interactive` exec flag and `pending_input` snapshot
field), and regenerate the typed client. The handler mirrors `CancelRun`.

<critical>
- ALWAYS READ the PRD and TechSpec before starting
- REFERENCE TECHSPEC for implementation details — do not duplicate here
- FOCUS ON "WHAT" — describe what needs to be accomplished, not how
- MINIMIZE CODE — show code only to illustrate current structure or problem areas
- TESTS REQUIRED — every task MUST include tests in deliverables
</critical>

<requirements>
- MUST add `POST /api/runs/{run_id}/input` routed to a new `SendInput` handler
  mirroring `CancelRun`, returning `202` on delivery, `404` for unknown run / no
  awaiting prompt, `409` when the run is not awaiting input, and the existing
  `403`/`500` problem responses.
- The `RunInputRequest` body MUST accept `prompt_id` plus one of `option_id`,
  `text`, or `cancelled`, validated before reaching the service.
- MUST declare the endpoint, `RunInputRequest`, the `interactive` field on the exec
  request, and `pending_input` on the snapshot in `openapi/rc-daemon.json`, then
  run codegen so `web/src/generated/rc-openapi.d.ts` is regenerated.
- The OpenAPI contract test MUST pass with the new route and schemas.
</requirements>

## Subtasks
- [x] 7.1 Add the `SendInput` handler that binds/validates `RunInputRequest` and
      maps service errors to status codes.
- [x] 7.2 Register the route in the runs route group.
- [x] 7.3 Declare the endpoint + `RunInputRequest` in the OpenAPI spec.
- [x] 7.4 Add `interactive` to the exec request schema and `pending_input` to the
      snapshot schema.
- [x] 7.5 Run `node scripts/codegen.mjs` and verify the generated client.

## As-built notes (for downstream tasks)
- **Request body field is `canceled`** (American spelling, matching the existing
  `contract.RunInput.Canceled` Go type from task_06), not `cancelled` as written
  in the task prose. The generated TS type `RunInputRequest` exposes `canceled?`.
- **`ErrRunNotAwaitingInput` moved to the `internal/api/core` package.** The HTTP
  handler maps it to 409 via `statusForError`; it cannot import `internal/daemon`
  (cycle), so the canonical sentinel now lives in apicore and `daemon` re-exports
  it (`var ErrRunNotAwaitingInput = apicore.ErrRunNotAwaitingInput`). `errors.Is`
  still matches in every existing daemon caller/test.
- **Validation returns HTTP 400** (`badRequestProblem`, code `invalid_request`),
  not 422, and runs before the service is reached.
- **`interactive` is wired end-to-end** (per task_06's "task_07 sets it from the
  API"): `contract.ExecRequest.Interactive` → `StartExecRun` handler →
  `RunManager` exec prepare → `RuntimeConfig.Interactive`. Only the `/api/exec`
  start path carries the flag; task/review run starts are unchanged.
- **`pending_input` added to BOTH `RunSnapshot` and `RunSnapshotPayload`** OpenAPI
  schemas (they both back `contract.RunSnapshot`); `RunSnapshotPayload` is the one
  the snapshot endpoint and contract test exercise.

## Implementation Details
Add the handler in `internal/api/core/handlers.go` next to `CancelRun` (:1415) and
the route in `internal/api/core/routes.go` (:72-81). Map the typed not-found /
not-awaiting errors from task_06 to 404/409. Edit `openapi/rc-daemon.json`
mirroring the `cancelRun` operation and `ExecRequest` schema (see TechSpec "API
Endpoints"). The codegen flow is `scripts/codegen.mjs` → `web/src/generated/
rc-openapi.d.ts`. The contract test lives in
`internal/api/httpapi/openapi_contract_test.go`.

### Relevant Files
- `internal/api/core/handlers.go` — `CancelRun` (:1415) as the pattern; new
  `SendInput` handler.
- `internal/api/core/routes.go` — runs route group (:72-81).
- `internal/api/core/interfaces.go` — `RunService.SendInput` (from task_06).
- `openapi/rc-daemon.json` — endpoint + schemas.
- `scripts/codegen.mjs` — regenerates the TS client.
- `internal/api/httpapi/openapi_contract_test.go` — route/schema contract test.

### Dependent Files
- `web/src/generated/rc-openapi.d.ts` — regenerated types consumed by the frontend
  (task_08).
- `web/src/systems/exec` — `interactive` flag becomes available (task_08).

### Related ADRs
- [ADR-003: Represent "awaiting input" as an event plus a snapshot field](../adrs/adr-003.md) — `pending_input` schema field.
- [ADR-004: Interactivity is opt-in via a flag set at run start](../adrs/adr-004.md) — `interactive` request field.

## Deliverables
- `SendInput` handler + route.
- OpenAPI declarations (endpoint, `RunInputRequest`, `interactive`,
  `pending_input`) and regenerated client.
- Unit tests with 80%+ coverage **(REQUIRED)**
- Integration tests via the OpenAPI contract test **(REQUIRED)**

## Tests
- Unit tests:
  - [x] `POST /runs/{id}/input` with a valid `option_id` returns 202 and calls the
        service once.
  - [x] Missing `prompt_id` (and no body fields) returns 400 before the service is
        called.
  - [x] Unknown run id maps the service not-found error to 404.
  - [x] A run that is not awaiting maps the not-awaiting error to 409.
- Integration tests:
  - [x] The OpenAPI contract test confirms the route and `RunInputRequest`/
        `pending_input` schemas are present and consistent.
  - [x] `node scripts/codegen.mjs --check` passes (generated client matches spec).
- Test coverage target: >=80%
- All tests must pass

## Success Criteria
- All tests passing
- Test coverage >=80%
- `make verify` passes (fmt/lint/test/build)
- Generated TS client includes the input endpoint and new fields
