---
description: rc full pipeline — plan → exec → review → fix, each phase on its own model
agent: rc
---
Run the full rc delivery pipeline for: $ARGUMENTS

Delegate each phase to its specialized subagent via the task tool, in order, confirming success before the next:

1. Planning: `rc-prd` → `rc-techspec` → `rc-tasks` (produce the task files).
2. Execution: for each task, `rc-exec` (or `rc-exec-bulk` when running independent tasks in parallel).
3. Review: `rc-review` (must differ from the executor model — it does).
4. Fix: `rc-fix` for any issues the review raised; re-verify.

Stop and report if any phase fails. At the end, summarize what shipped.
