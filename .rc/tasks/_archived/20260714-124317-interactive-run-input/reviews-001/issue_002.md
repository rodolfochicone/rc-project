---
provider: manual
pr:
round: 1
round_created_at: 2026-06-15T10:54:55Z
status: resolved
file: web/src/systems/runs/components/run-detail-view.tsx
line: 95
severity: low
author: claude-code
provider_ref:
---

# Issue 002: resolvePendingInput recomputed on every render

## Review Comment

`run-detail-view.tsx:95-97` calls `resolvePendingInput(snapshot.pending_input,
liveEvents)` directly in the render body. `RunDetailView` re-renders on every
live stream event, and `resolvePendingInput` scans the whole `liveEvents` buffer
(bounded at 500 entries by `createRunEventStore`) on each render. The adjacent
transcript merge in `run-transcript-panel.tsx` already guards the equivalent
work behind `useMemo`; this derivation does not, so it recomputes even when the
inputs are unchanged.

Suggested fix: memoize on the same dependencies used elsewhere:

```tsx
const pendingInput = useMemo(
  () =>
    isTerminalRunStatus(run.status)
      ? null
      : resolvePendingInput(snapshot.pending_input ?? null, liveEvents),
  [run.status, snapshot.pending_input, liveEvents]
);
```

Low severity — the buffer is bounded, so the cost is small — but it is a cheap
consistency win with the surrounding code.

## Triage

- Decision: `VALID`
- Notes: Confirmed — the derivation ran in the render body while the adjacent
  transcript merge is memoized. Fix applied: wrapped `pendingInput` in `useMemo`
  keyed on `[run.status, snapshot.pending_input, liveEvents]` in
  `run-detail-view.tsx`. Pure perf/consistency change with no behavior change, so
  the existing `pendingInput`-derivation tests (panel shows/hides from snapshot
  and live events) remain the coverage; no new test added.
