---
description: rc planning pipeline — PRD → TechSpec → Tasks, each on its own model
agent: rc
---
Run the rc planning pipeline for: $ARGUMENTS

Delegate each phase to its specialized subagent via the task tool, in order, confirming the artifact exists before the next:

1. `rc-prd` subagent → create the PRD (skill rc-create-prd) → `.rc/tasks/<slug>/_prd.md`.
2. `rc-techspec` subagent → produce the TechSpec from that PRD (rc-create-techspec) → `_techspec.md`.
3. `rc-tasks` subagent → break it into task files (rc-create-tasks).

At the end, report the slug and the generated task files.
