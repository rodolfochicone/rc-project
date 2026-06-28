---
description: Run the full rc pipeline end to end — plan, execute, (quality), review, docs, ship — then distill learnings.
argument-hint: [feature-name-or-idea]
disable-model-invocation: true
---

You are running the **full rc pipeline** for: $ARGUMENTS

Drive every phase in order by invoking the underlying rc skills with the Skill tool. Confirm each phase's artifacts and verification before moving on; stop and report if a phase fails.

1. **Plan** — `rc-create-prd` → `rc-create-techspec` → `rc-create-tasks` (each reads the previous artifact under `.rc/tasks/<slug>/`).
2. **Execute** — implement each pending task with `rc-execute-task`, validating with the project's gate (e.g. `make verify`) after each.
3. **Quality (conditional)** — if the change has a subjective-quality surface (UI/UX, CLI ergonomics, user-facing copy), run `rc-gan` on that surface with a sensible threshold to drive it up before review. If the change has no such surface (e.g. backend/library only), **skip this step and say so** — do not run a quality loop on work `make verify` already covers.
4. **Review (≤3 rounds)** — first run `rc-simplify-review` once and apply the safe cuts it finds, then re-verify; then loop up to 3 times: `rc-review-round`; if it found issues, `rc-fix-reviews` then verify; stop when a round is clean or after 3 rounds.
5. **Docs** — `rc-readme`, then `rc-postman` and `rc-openapi` (skip the API docs if there is no HTTP API).
6. **Ship** — invoke `rc-git` to move the work onto a feature branch, push, and open a PR (it confirms each outward-facing step).
7. **Learn** — finally invoke `rc-instincts` to distill recurring corrections and patterns from this session into `.rc/instincts/INSTINCTS.md`. It is a no-op when nothing durable is worth capturing, so it is safe to always run. (Capture is richer when the session ran with `RC_INSTINCTS=1`.)

At the end, summarize what each phase produced, whether the quality loop ran or was skipped, the PR URL, and any instincts recorded.
