---
description: Review-and-fix loop for a rc feature — a one-shot simplify pass, then up to 3 rounds of review-round plus fix-reviews.
argument-hint: [slug]
disable-model-invocation: true
---

You are running the **rc review loop** for the feature slug: $ARGUMENTS — a one-shot simplify pass, then a maximum of **3 rounds** of quality review.

First, once, before the loop:

0. **Simplify (one-shot)** — invoke the `rc-simplify-review` skill for the slug. It writes a ranked delete-list (over-engineering only) to `.rc/tasks/<slug>/simplify-review-NNN.md`. Apply the cuts you can prove safe — dead code and single-caller abstractions confirmed via Serena; the skill never flags validation, error handling, security, or concurrency — then re-run the project's verification gate. Leave judgment calls (`shrink:`/`yagni:`) for the user. Running this first means the rounds below review already-lean code.

Then repeat up to 3 times:

1. **Review** — invoke the `rc-review-round` skill for the slug. It writes a review round directory (`reviews-NNN/`) with issue files.
2. **Check** — if the round found **no issues**, stop the loop and report success.
3. **Fix** — otherwise invoke the `rc-fix-reviews` skill to remediate every issue in that round, then validate with the project's verification gate.
4. Increment the round counter. After **3 rounds**, stop even if issues remain.

At the end, report the simplify net (lines/deps cut), how many rounds ran, what was fixed, and any issues still open. Suggest `/rc-docs` next.
