---
name: rc-lessons
description: Deterministic loop-lessons machine — records grounded engineering lessons from verification signals (AC gaps, surviving mutants, spec deviations, gate failures) and loads the corroborated ones at plan/design time, so autonomous or repeated loops stop re-making the same mistake. All bookkeeping (IDs, distinct-feature recurrence, candidate→confirmed promotion, quarantine, prune) is owned by scripts/lessons.mjs, not prose. Use to record a lesson after a verify/review failure, or to load confirmed lessons before planning a task or roadmap phase. Do not use for user-correction trigger→action learnings (rc-memory), per-task working notes (rc-workflow-memory), or on-demand reflection on a diff (rc-lesson-learned).
user-invocable: false
model: sonnet
effort: low
---

# Loop Lessons

The learning layer that makes RC loops *safe to repeat*. Every time verification catches a
defect, the reason it slipped through is captured as a **grounded lesson**; before the next
task or roadmap phase is planned, the **corroborated** lessons are loaded as guidance. A
one-off is never trusted — a lesson is only promoted after it recurs across **two distinct
features**, and is quarantined if it ever fails when applied.

**The script owns the bookkeeping.** IDs, recurrence counting, promotion, quarantine, and
pruning are mechanical and rot when done by hand, so they live in `scripts/lessons.mjs`
(deterministic, dependency-free) — never re-implement them in prose or edit the store by hand.

- Canonical store: `.rc/lessons.json` (machine-owned — do **not** hand-edit)
- Rendered playbook: `.rc/LESSONS.md` (regenerated on every write)

Invoke the script via the plugin root:

```bash
node "$CLAUDE_PLUGIN_ROOT/scripts/lessons.mjs" <command> --root <project-root> [flags]
```

`--root` is the directory that contains `.rc/` (default `.`). Run `--selftest` to verify the
lifecycle offline.

## When to RECORD a lesson (write)

Record **only** at a verification/review boundary, and **only** with real grounding — a lesson
with no `file:line` / AC id / mutant id / SPEC_DEVIATION reference is an opinion, and the script
refuses it. Record when a `rc-execute-task` / `rc-final-verify` / `rc-code-review` /
`rc-review-round` pass catches:

| Signal | Use when |
| --- | --- |
| `ac_gap` | An acceptance criterion had no test / was not covered. |
| `surviving_mutant` | A discrimination/adversarial check survived — the test was too weak. |
| `spec_precision_gap` | The spec did not define a precise enough outcome. |
| `spec_deviation` | The implementation diverged from spec/design (SPEC_DEVIATION). |
| `gate_fail` | A build-level gate failed (build / lint / typecheck / test). |

```bash
node "$CLAUDE_PLUGIN_ROOT/scripts/lessons.mjs" add \
  --feature <slug> \
  --signal <ac_gap|surviving_mutant|spec_precision_gap|spec_deviation|gate_fail> \
  --source "<file:line | AC-id | mutant-id | SPEC_DEVIATION ref>" \
  --text "<one terse, actionable, canonical sentence>" \
  --scope "<optional path/layer/tag for retrieval>"
```

Phrase `--text` **tersely and canonically** — dedup is exact-after-normalization, so two
wordings of the same lesson only merge (and thus promote) if they normalize alike. State the
fix, not the incident: *"Assert world x/z per placement, not just prop count"*, not *"the
environment test failed again"*.

## When to LOAD lessons (read)

Load **confirmed** lessons before planning — at the start of `rc-create-tasks`, `rc-create-prd`
technical clarification, a `rc-roadmap` phase plan, or the Planner step of a loop. Confirmed
lessons are corroborated guidance; candidates are tracked but **not** trusted, so do not load
them as instructions.

```bash
node "$CLAUDE_PLUGIN_ROOT/scripts/lessons.mjs" list --root <project-root>            # confirmed only
node "$CLAUDE_PLUGIN_ROOT/scripts/lessons.mjs" list --root <project-root> --scope server/room
node "$CLAUDE_PLUGIN_ROOT/scripts/lessons.mjs" status --root <project-root>
```

Apply a confirmed lesson as a planning/design constraint for the matching scope. If applying a
lesson leads to a failure, penalize it so it heads toward quarantine instead of misleading the
next loop:

```bash
node "$CLAUDE_PLUGIN_ROOT/scripts/lessons.mjs" penalize --root <project-root> --id L-007
```

## Rules

- **Grounding or nothing.** No `--source`, no lesson. The script enforces this; do not work around it.
- **Never hand-edit** `.rc/lessons.json` or `.rc/LESSONS.md` — the next script write overwrites them.
- **One lesson = one terse sentence.** Not a paragraph, not a stack trace, not the task spec.
- **Confirmed guides, candidates observe.** Only load `confirmed` at plan/design time.
- **Distinct-feature promotion.** `--feature` is the slug the signal came from; re-recording within
  the same feature accrues evidence but never promotes.

## Not this skill

- **User corrections / workflow trigger→action learnings, cross-session** → `rc-memory` (`.rc/memory/LEARNINGS.md`).
- **Per-task working notes, decisions, handoff within one run** → `rc-workflow-memory` (`.rc/tasks/<slug>/memory/`).
- **On-demand "what did I learn from this diff?" reflection** (no store) → `rc-lesson-learned`.
