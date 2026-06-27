---
description: Execute the implemented rc tasks for a feature using the rc-execute-task skill.
argument-hint: [slug]
disable-model-invocation: true
---

You are running the **rc execution phase** for the feature slug: $ARGUMENTS

1. Resolve the workflow directory `.rc/tasks/<slug>/` (ask for the slug if it was not provided).
2. Identify the pending task files, in order.
3. For each pending task, invoke the `rc-execute-task` skill to implement it end-to-end, then validate with the project's verification gate (e.g. `make verify`) before moving to the next task. Stop and report if a task cannot be completed.
4. Let the skill reflect progress in the task tracking files.

At the end, report which tasks were completed and the verification result. Suggest `/rc-review` next.
