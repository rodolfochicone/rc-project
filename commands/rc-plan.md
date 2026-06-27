---
description: Run the rc planning pipeline — PRD, TechSpec, and task breakdown — in sequence for a feature.
argument-hint: [feature-name-or-idea]
disable-model-invocation: true
---

You are running the **rc planning pipeline** for: $ARGUMENTS

Run these rc skills in order, using the Skill tool. Each phase reads the artifact the previous one wrote under `.rc/tasks/<slug>/`, so do not skip or reorder them.

1. **PRD** — invoke the `rc-create-prd` skill with the feature name/idea above. It produces `_prd.md`.
2. **TechSpec** — once the PRD is complete, invoke `rc-create-techspec` for the same feature. It reads `_prd.md` and produces `_techspec.md`.
3. **Tasks** — once the TechSpec is complete, invoke `rc-create-tasks`. It reads the PRD + TechSpec and writes the task files.

Between phases, confirm the previous artifact exists before continuing. If a phase needs input from the user, ask before moving on. At the end, report the slug and the list of generated task files so the user can run `/rc-exec` next.
