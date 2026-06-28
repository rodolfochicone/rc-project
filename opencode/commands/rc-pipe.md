---
description: rc full pipeline — plan → exec → (quality) → review → fix → learn, each phase on its own model
agent: rc
---

Run the full rc delivery pipeline for: $ARGUMENTS

Delegate each phase to its specialized subagent via the task tool, in order, confirming success before the next:

1. Planning: `rc-prd` → `rc-techspec` → `rc-tasks` (produce the task files).
2. Execution: for each task, `rc-exec` (or `rc-exec-bulk` when running independent tasks in parallel).
3. Quality (conditional): if the change has a subjective-quality surface (UI/UX, CLI ergonomics, user-facing copy), delegate to `rc-gan` to drive that surface up to a threshold. Skip — and say so — for backend/library-only work that `make verify` already covers.
4. Review: `rc-review` (must differ from the executor model — it does).
5. Fix: `rc-fix` for any issues the review raised; re-verify.
6. Learn: run the `rc-instincts` skill to distill recurring corrections and patterns from this session into `.rc/instincts/INSTINCTS.md` (no-op when nothing durable is worth capturing).

Stop and report if any phase fails. At the end, summarize what shipped, whether the quality loop ran or was skipped, and any instincts recorded.
