---
name: rc-compact
description: Decides when and how to compact the conversation at logical task boundaries, driven by the session's real token usage rather than waiting for automatic mid-task compaction. Use during long multi-phase runs to compact deliberately so the right context survives. Do not use for context inventory of installed tooling (use rc-context-budget) or for persistent cross-session memory (use rc-workflow-memory / rc-project-memory).
model: haiku
effort: low
---

# Strategic Compaction

Automatic compaction fires at an arbitrary point — often mid-edit — and discards whatever it judges least recent, which is frequently the task state you still need. Compacting **deliberately at a logical boundary** keeps you in control of what survives. This skill decides *when* to suggest compaction (from the real usage signal) and *how* to do it so nothing load-bearing is lost.

## When to compact — read the real signal

Base the decision on actual token usage, not message count:

- Find the session's token usage (sum of input + cache-read + cache-creation from the latest turn's usage, when available).
- Detect the window: default 200k; a 1M-context model is flagged by a `[1m]` marker in the model id.
- **Threshold**: suggest compaction at ~80% of the window — ~160k of 200k, ~800k of 1M — and re-suggest every additional ~60k if the user defers.
- If usage is not available, fall back to phase boundaries: suggest compaction after completing a major phase when the conversation has clearly grown large.

**Always prefer a boundary.** Even past threshold, finish the current atomic step (don't compact between an Edit and its verification); compact at the seam between phases (after a task is implemented + verified, before starting the next).

## How to compact — what must survive

Before compacting, make sure the durable state is captured outside the conversation so it survives:

- The current objective and which phase/task is in progress.
- What has been done and **verified** (with evidence), what was tried and failed, what is not yet attempted.
- Key file paths, decisions/ADRs, and open follow-ups.

If a workflow-memory or project-memory skill is in use, flush this into it first (it persists across compaction and sessions). Then trigger `/compact` (or the harness equivalent) with a short instruction naming what to keep: "keep the current task objective, verified progress, and the open follow-ups; drop resolved tangents."

## What NOT to do

- Do not compact mid-edit or between an action and its verification.
- Do not rely on compaction as memory — it is lossy. Durable facts go to memory skills first.
- Do not auto-compact silently when the user is mid-conversation; suggest it and let them choose unless they asked you to manage it.

## Output

When the threshold is crossed, state: current estimated usage vs window, that a logical boundary has (or has not) been reached, what you will preserve, and the compaction action you recommend — then act on it if the user agrees or has delegated the decision.
