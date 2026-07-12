---
name: rc-loop
description: Autonomous creator-loop driver (Claude Code only) — walks .rc/ROADMAP.md one phase at a time, planning, executing, and verifying each phase without per-step human gates, until the roadmap is exhausted or a phase cannot reach a green gate. Use to advance a roadmap unattended for a long-running build/migration on a Claude Code host, after the loop-readiness gate passes. Do not use without a passing readiness gate, without a .rc/ROADMAP.md (author it with rc-roadmap first), on non–Claude Code hosts (run each phase's tasks via rc-execute-task), or for a single task (rc-execute-task) or a fixed pre-authored task set (rc-tasks-workflow).
argument-hint: "[roadmap-or-milestone]"
user-invocable: true
model: sonnet
effort: high
---

# Loop (creator)

Advance `.rc/ROADMAP.md` autonomously, one phase at a time: plan the phase, execute it, verify
it, and — only on a recorded PASS — flip its checkbox and move to the next. This is the
**creator loop** (it produces side effects and builds each phase on the last), so it is only as
safe as the harness under it. It reuses the existing RC skills; it does not re-implement plan,
execute, or verify.

## Stop and defer — do not start if

- **No `Workflow` tool** (non–Claude Code host) → stop; tell the user to run each phase's tasks
  via `rc-execute-task` in dependency order. This driver is Claude Code only, like `rc-tasks-workflow`.
- **No `.rc/ROADMAP.md`** → stop; point at `rc-roadmap` (`create`). A loop cannot invent intent.
- **Readiness gate fails** → stop. Read `references/loop-readiness.md` and answer its four
  questions first. If the answer is not "yes" to all four, stay in human-gated spec-driven
  (`rc-pipe` / `rc-card`) until the harness is ready. Autonomy on a weak harness perpetuates bugs.

## The cycle (per phase)

Resolve the next actionable phase via `rc-roadmap` (`next`). Then, for that phase:

1. **Load guidance** — `rc-lessons` (`list`, confirmed only) filtered to the phase's scope, plus
   the shared workflow memory (`rc-workflow-memory`) and any STATE decisions. Carry these into planning.
2. **Plan** — mark the phase `[~]`, then author the phase's tasks with `rc-create-tasks` scoped to
   the phase goal and its `> Done when:` gate. **Autonomous mode:** resolve ambiguities as explicit
   spec assumptions and record them; do not open a user confirmation gate mid-loop.
3. **Execute** — run the phase's tasks via `rc-tasks-workflow` (the `Workflow`-tool engine, one
   subagent per task in dependency order). Each task follows the `rc-execute-task` contract:
   explore → implement → run the project gate → update tracking.
4. **Verify** — the project's gate (`rc-final-verify`) is the until-condition. On a red gate,
   iterate `gather → fix root cause → re-verify` up to the bounded fix cap; escalate a stubborn
   failure to `rc-oracle`. Never advance on a red gate.
5. **Record lessons** — for every defect verification caught (AC gap, surviving mutant, spec
   deviation, precision gap, gate failure), record a grounded lesson via `rc-lessons` (`add`) so
   the next phase does not repeat it. This is what stops a creator loop from perpetuating errors.
6. **Close the phase** — on PASS: flip the roadmap checkbox to `[x]`, write the phase outcome +
   handoff to shared workflow memory (`rc-workflow-memory`), and commit the phase locally. Then
   return to step 1 for the next actionable phase.

## Reliable stop condition

The loop is **not** infinite — it stops and hands back to the human when:

- **Roadmap exhausted** — every phase is `[x]`. Report completion; the human decides the next
  batch (`rc-roadmap` `create`) or ends the milestone.
- **A phase cannot reach green** within the fix cap (and `rc-oracle` escalation did not resolve
  it). Leave the phase `[~]`, record the failing evidence and any lessons, and stop. Do not skip
  ahead to a dependent phase on a red gate.
- **Intent runs out** — the next phase needs a human decision the loop cannot assume (a
  product/architecture fork, external research, an asset the project does not have). Stop and
  surface the specific question.

Persist enough in shared workflow memory (`## Handoff`) that a re-invocation resumes from the
`[~]` phase without re-deriving context.

## Human-gated boundary (never autonomous)

The loop implements, verifies, and commits **locally**. Outward-facing actions stay human-gated
and are **not** performed by the loop: opening a PR (`rc-git`), pushing, and any Linear write or
state move (`rc-card` / `rc-linear`). The loop's job is to leave a green, committed working tree
per phase; shipping is the human's confirmed step. Match the `rc-card` guardrail — never offload a
confirmation-bearing action into unattended automation.

## Not this skill

- **Single task** → `rc-execute-task`. **Fixed, already-authored task set (one pass)** → `rc-tasks-workflow`.
- **Human-gated feature pipeline** → `rc-pipe`. **Human-gated Linear sub-issue loop** → `rc-card`.
- **Authoring/reordering the roadmap** → `rc-roadmap` (the intent step).
