---
description: Review-and-fix loop for a rc feature — a one-shot simplify pass, then up to 3 rounds of review-round plus fix-reviews.
argument-hint: [slug]
disable-model-invocation: true
---

You are running the **rc review loop** for the feature slug: $ARGUMENTS — a one-shot simplify pass, then a maximum of **3 rounds** of quality review.

First, once, before the loop:

0. **Simplify (one-shot)** — invoke the `rc-simplify-review` skill for the slug. It writes a ranked delete-list (over-engineering only) to `.rc/tasks/<slug>/simplify-review-NNN.md`. Apply the cuts you can prove safe — dead code and single-caller abstractions confirmed via Serena; the skill never flags validation, error handling, security, or concurrency — then re-run the project's verification gate. Leave judgment calls (`shrink:`/`yagni:`) for the user. Running this first means the rounds below review already-lean code.

Then repeat up to 3 times:

1. **Review (fresh eyes)** — invoke the `rc-review-round` skill for the slug. Review each round from scratch — re-derive findings against the current code; do not merely re-check the previous round's issues (that anchors you to the last framing and misses what the fix introduced). The skill already excludes issues tracked in prior rounds, so a round surfaces only _new_ problems.
2. **Convergence check** — the loop is converged only when a full fresh round yields **zero new blocking issues** (no `critical`/`high`). On a converged round, stop and report success.
3. **Fix** — otherwise invoke the `rc-fix-reviews` skill to remediate every issue in that round, then validate with the project's verification gate.
4. Increment the round counter. After **3 rounds**, stop. If blocking issues remain, do **not** declare success — stop and escalate to the user, listing the open blockers, instead of looping further.

At the end, report the simplify net (lines/deps cut), how many rounds ran, what was fixed, whether the loop converged (clean round) or hit the cap, and any issues still open. Suggest `/rc-docs` next.
