---
description: Execute the implemented RC tasks for a feature using the rc-execute-task skill.
argument-hint: [slug]
disable-model-invocation: true
---

You are running the **RC execution phase** for the feature slug: $ARGUMENTS

1. Resolve the workflow directory `.rc/tasks/<slug>/` (ask for the slug if it was not provided).
2. Identify the pending task files, in order.
3. For each pending task, invoke the `rc-execute-task` skill to implement it end-to-end. The skill validates against the project's verification gate (e.g. `make verify`) as the until-condition, running a bounded **verify→fix loop**: on a red gate it iterates up to **3 fix cycles**, escalates a stubborn failure to `rc-oracle`, and never marks a task complete on a red gate. If a task is still red after the cap, stop and report it as blocked instead of moving to the next task.
4. Let the skill reflect progress in the task tracking files.

At the end, report which tasks were completed and the verification result. Suggest `/rc-review` next.
