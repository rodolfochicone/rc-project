---
description: Review-and-fix loop for a RC feature — a one-shot simplify pass, then up to 3 rounds of review-round plus fix-reviews.
argument-hint: [slug]
disable-model-invocation: true
---

You are running the **RC review loop** for the feature slug: $ARGUMENTS — a one-shot simplify pass, then a maximum of **3 rounds** of quality review.

First, once, before the loop:

0. **Simplify (one-shot)** — invoke the `rc-simplify-review` skill for the slug. It writes a ranked delete-list (over-engineering only) to `.rc/tasks/<slug>/simplify-review-NNN.md`. Apply the cuts you can prove safe — dead code and single-caller abstractions confirmed via Serena; the skill never flags validation, error handling, security, or concurrency — then re-run the project's verification gate. Leave judgment calls (`shrink:`/`yagni:`) for the user. Running this first means the rounds below review already-lean code.

Then repeat up to 3 times:

1. **Review** — invoke the `rc-review-round` skill for the slug. It writes a review round directory (`reviews-NNN/`) with issue files.
2. **Fix** — if the round found any issues, invoke the `rc-fix-reviews` skill to remediate **every** issue in that round (all severities, root cause first), then validate with the project's verification gate.
3. **Check convergence** — stop the loop when the round surfaced **no new high- or critical-severity issues** (any medium/low issues were already fixed in step 2, but they don't earn another expensive round). Also stop if the round found nothing at all.
4. Increment the round counter. After **3 rounds**, stop even if high/critical issues remain, and report them — don't keep looping.

At the end, report the simplify net (lines/deps cut), how many rounds ran, what was fixed, and any issues still open. Suggest `/rc-docs` next.
