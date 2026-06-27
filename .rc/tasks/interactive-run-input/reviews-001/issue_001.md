---
provider: manual
pr:
round: 1
round_created_at: 2026-06-15T10:54:55Z
status: resolved
file: web/src/systems/runs/components/run-detail-view.tsx
line: 182
severity: medium
author: claude-code
provider_ref:
---

# Issue 001: RunInputPanel keeps stale free-text across prompt transitions

## Review Comment

`RunInputPanel` holds the free-text answer in local state
(`const [text, setText] = useState("")` at `run-input-panel.tsx:41`), but
`run-detail-view.tsx:182` renders it **without a `key`**:

```tsx
{pendingInput ? (
  <RunInputPanel pendingInput={pendingInput} onSubmit={onSendInput} ... />
) : null}
```

When one prompt is immediately superseded by another **without the panel
unmounting in between** (e.g. an agent that hits two consecutive permission
requests in the same turn, so no intervening `session.update` clears the first
prompt), `resolvePendingInput` returns the new prompt while React reuses the same
component instance. The `text` state typed for prompt A then carries over to
prompt B, and a stale free-text answer can be submitted against the wrong
`prompt_id`.

Suggested fix: reset the panel's local state per prompt by keying the element on
the prompt id, which is the idiomatic "reset state on prop change" pattern:

```tsx
<RunInputPanel key={pendingInput.prompt_id} pendingInput={pendingInput} ... />
```

This guarantees a fresh text box for every distinct prompt with no effect logic.

## Triage

- Decision: `VALID`
- Notes: Confirmed — `RunInputPanel` owns the free-text via `useState` and the
  detail view rendered it without a `key`, so React reuses the instance across
  back-to-back prompts and carries stale text. Root cause: missing per-prompt
  state reset. Fix applied: `key={pendingInput.prompt_id}` on the `<RunInputPanel>`
  in `run-detail-view.tsx` (the idiomatic "reset state on prop change" pattern).
  Regression guard added in `run-input-panel.test.tsx` ("…reset the text box when
  keyed by a new prompt id…"), which mirrors the detail view's keying and asserts
  the textarea clears when the prompt id changes.
