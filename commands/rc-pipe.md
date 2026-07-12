---
description: Run the full RC pipeline end to end — plan, execute, review, docs — then open a PR with rc-git.
argument-hint: [feature-name-or-idea]
disable-model-invocation: true
---

You are running the **full RC pipeline** for: $ARGUMENTS

Drive every phase in order by invoking the underlying RC skills with the Skill tool. Confirm each phase's artifacts and verification before moving on; stop and report if a phase fails.

0. **Warm-up (optional)** — on a large or unfamiliar codebase, invoke `rc-codemap` first so every later phase reads cheap, per-directory structure maps instead of re-exploring from scratch. Skip on a small repo where the maps cost more than they save.
1. **Plan** — `rc-create-prd` → `rc-create-techspec` → `rc-create-tasks` (each reads the previous artifact under `.rc/tasks/<slug>/`).
2. **Execute** — implement each pending task with `rc-execute-task`, validating with the project's gate (e.g. `make verify`) after each.
3. **Review (≤3 rounds)** — first run `rc-simplify-review` once and apply the safe cuts it finds, then re-verify; then loop up to 3 times: `rc-review-round`; if it found issues, `rc-fix-reviews` then verify; stop when a round is clean or after 3 rounds.
4. **Docs** — `rc-readme`, then `rc-postman` and `rc-openapi` (skip the API docs if there is no HTTP API).
5. **Ship** — finally invoke `rc-git` to move the work onto a feature branch, push, and open a PR (it confirms each outward-facing step).

At the end, summarize what each phase produced and the PR URL.
