# Workflow Guide

End-to-end walkthrough of the rc development pipeline, from requirements through review.

## Prerequisites

1. **Install the skills.** For Claude Code, install the plugin (`/plugin marketplace add rodolfochicone/rc-project` then `/plugin install rc@rc-project`). For OpenCode, copy the `opencode/` bundle into your OpenCode config (see the README).
2. **Work inside your agent.** Every phase below is a slash command or skill run from within Claude Code or OpenCode.

## Phase 1: Requirements

**Skill:** `/rc-create-prd [feature-name-or-idea]`

1. Invoke `/rc-create-prd` with the feature name.
2. The skill runs parallel codebase and market research.
3. Answer clarifying questions focused on WHAT and WHY (not HOW).
4. Choose from 2-3 product approaches. An ADR is created for the decision.
5. Review and approve the complete PRD draft.
6. Output: `.rc/tasks/<slug>/_prd.md` + ADRs.

**Key rule:** The PRD describes user capabilities and business outcomes only. No databases, APIs, frameworks, or architecture.

## Phase 2: Technical Design

**Skill:** `/rc-create-techspec [feature-name]`

1. Invoke `/rc-create-techspec` with the feature name.
2. The skill reads the existing `_prd.md` and explores the codebase architecture.
3. Answer technical clarifying questions.
4. Technical ADRs are created for architecture decisions.
5. Review and approve the TechSpec draft.
6. Output: `.rc/tasks/<slug>/_techspec.md` + ADRs.

**Contains:** System architecture, data models, core interfaces, API design, development sequencing.

## Phase 3: Task Decomposition

**Skill:** `/rc-create-tasks [feature-name]`

1. Invoke `/rc-create-tasks` with the feature name.
2. The skill loads the PRD and TechSpec, then breaks them into granular tasks.
3. Review the proposed task breakdown.
4. Task files are generated with YAML frontmatter: `status`, `title`, `type`, `complexity`, `dependencies`.
5. Tasks are enriched with codebase patterns and implementation context.
6. Validate that every task file has the required frontmatter fields before executing.
7. Output: `task_01.md` through `task_N.md`, `_tasks.md` master list.

**Task types:** `frontend`, `backend`, `docs`, `test`, `infra`, `refactor`, `chore`, `bugfix`.

## Phase 4: Execution

**Skill:** `rc-execute-task`

1. Read task files from `.rc/tasks/<slug>/` in order, respecting `dependencies`.
2. For each pending task, gather context: the task spec, PRD, TechSpec, ADRs, and workflow memory.
3. Execute the task directly in the agent session using the `rc-execute-task` skill.
4. Each task: read spec → implement → validate with `rc-final-verify` → update tracking → optional commit.
5. Maintain workflow memory across tasks via `rc-workflow-memory`.

## Phase 5: Review

**Skill:** `/rc-review-round [feature-name]`

Invoke inside an agent session. The skill performs a comprehensive code review of the implementation and generates issue files under `.rc/tasks/<slug>/reviews-NNN/`.

**Produces:** `issue_*.md` files with YAML frontmatter containing round metadata (`round`, `round_created_at`) plus issue metadata (`status`, `severity`, `file`, `line`).

## Phase 6: Remediation

**Skill:** `rc-fix-reviews`

1. Read issue files from the latest (or specified) review round.
2. For each issue, triage (valid/invalid), fix if valid (in severity order), and verify with `rc-final-verify`.
3. Update issue file frontmatter: `pending` → `valid`/`invalid` → `resolved`.

**Iterate:** Repeat phases 5-6 until all reviews are clean, then merge.

## Archiving

Move fully completed workflows from `.rc/tasks/<slug>/` to `.rc/tasks/_archived/<slug>/` once all task items are completed and all review issues are resolved, to keep the tasks directory focused.

## Workflow Memory

The `rc-workflow-memory` skill maintains two tiers of context during task execution:

| File | Purpose |
| --- | --- |
| `.rc/tasks/<slug>/memory/MEMORY.md` | Shared cross-task memory: architecture decisions, discovered patterns, open risks, handoffs |
| `.rc/tasks/<slug>/memory/task_NN.md` | Per-task memory: objective snapshot, files touched, errors hit, next steps |

- Memory files are scaffolded before task execution and updated during the run.
- Agents read both files as mandatory context before editing code.
- Promotion from task to shared memory requires: needed by other tasks, durable across runs, and not derivable from existing artifacts.
- Auto-compaction triggers when files exceed size limits.

## Architecture Decision Records

ADRs are created during the PRD and TechSpec phases to document significant decisions.

- **Location:** `.rc/tasks/<slug>/adrs/adr-NNN.md` (zero-padded 3-digit numbers).
- **Structure:** Status, Date, Context, Decision, Alternatives Considered, Consequences.
- **Referenced by:** PRDs and TechSpecs include an "Architecture Decision Records" section linking to all ADRs.
