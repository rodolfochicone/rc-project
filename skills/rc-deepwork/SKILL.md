---
name: rc-deepwork
description: A discipline for heavy, multi-phase, or risky coding sessions where the driver acts as a scheduler — drafting a plan, getting it reviewed before implementing, tracking progress in a durable file, and advancing phase by phase with verification gates — instead of diving straight into edits. Use for broad, multi-file, or risky work that spans several phases. Do not use for trivial edits, quick docs changes, or simple one-file bug fixes.
model: opus
effort: high
user-invocable: true
argument-hint: "[slug]"
---

# Deepwork

For heavy sessions, manage the work as a scheduler, not as the default implementer. The point is to stay oriented across many phases: plan first, review the plan, track progress durably, and gate each phase on verification.

Use it when the work is broad, risky, multi-file, or likely to span several implementation phases. Skip it for trivial edits, quick docs, or simple bug fixes — the overhead would exceed the value.

## Progress file — `.rc/deepwork/<slug>.md`

Create and maintain one markdown progress file for the session. It is the durable memory of the effort and survives compaction:

```markdown
# Deepwork: <slug>

## Goal
<one paragraph: what "done" means>

## Plan
- [ ] Phase 1 — <name>
- [ ] Phase 2 — <name>

## Confirmed research
<facts discovered and reconciled during the session — not speculation>

## Log
<dated one-line entries: decisions, blockers, phase transitions>
```

Keep the checklist and the session todos aligned with the active phase at all times.

## Protocol

1. **Draft a plan** before implementing. Break the work into ordered, verifiable phases with clear acceptance criteria.
2. **Get the plan reviewed** — have an independent reviewer (the `rc-review` agent, or the `rc-code-review` skill on the plan) critique it, and revise until it holds. Do not start implementing an unreviewed plan for risky work.
3. **Write confirmed research** into the progress file as it is reconciled, so later phases build on facts, not memory.
4. **Execute phase by phase.** Implement one phase, then run the project's verification gate (`rc-final-verify`) before advancing. Keep changes scoped to the active phase.
5. **Delegate** heavy or parallel phases to the phase agents (`rc-exec`, `rc-exec-bulk`), and isolate risky ones with `rc-worktrees`.
6. **Reconcile** the checklist and log after each phase; update the goal if scope legitimately changes (and say why).

## Rules

- No phase is "done" without fresh verification evidence (`rc-final-verify`).
- Never let the todos, the progress file, and the actual state diverge.
- If the plan is proven wrong mid-flight, stop and re-plan — do not push a broken plan to completion.
