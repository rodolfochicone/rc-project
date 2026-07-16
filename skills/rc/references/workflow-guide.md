# Workflow Guide

End-to-end walkthrough of the RC development pipeline from setup through archive.

## Prerequisites

1. **Install the plugin.** Install RC through your host's plugin/marketplace mechanism (Claude Code: `/plugin marketplace add rodolfochicone/rc-project` then `/plugin install rc@rc-project`). Skills, commands, agents, and hooks are auto-discovered. There is no binary, daemon, or CLI.
2. **Configure (optional).** Model and reasoning effort live in each skill's/agent's frontmatter; hook behavior is set via `RC_HOOK_PROFILE` / `RC_DISABLED_HOOKS`. Read `config-reference.md`.

## Phase 1: Requirements

**Skill:** `/rc-create-prd [feature-name-or-idea] [idea-file]`

1. Invoke `/rc-create-prd` with the feature name. If `_idea.md` exists, it is used as primary context.
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
6. Validation runs via `node "$CLAUDE_PLUGIN_ROOT/scripts/validate-tasks.mjs" --slug <slug>`.
7. Output: `task_01.md` through `task_N.md`, `_tasks.md` master list.

**Task types:** `frontend`, `backend`, `docs`, `test`, `infra`, `refactor`, `chore`, `bugfix` — the conventional set `rc-create-tasks` writes. The validator requires the `type` field to be present but does not constrain its value, so a project-specific type is allowed.

## Phase 4: Execution

**Run:** `/rc-tasks-workflow <slug>` (Claude Code) or the `rc-execute-task` skill per task in dependency order (any host)

1. Validate the set first: `node "$CLAUDE_PLUGIN_ROOT/scripts/validate-tasks.mjs" --slug <slug>`.
2. Task files are read from `.rc/tasks/<slug>/` and ordered topologically by their `dependencies` frontmatter. Tasks already `status: completed` are skipped.
3. On **Claude Code**, `/rc-tasks-workflow <slug>` drives the `Workflow` tool — one subagent per task, sequentially (tasks share the working tree; no parallelism, no worktree isolation). On **any other host**, run each task through the `rc-execute-task` skill in dependency order.
4. Each task carries its spec plus the PRD, TechSpec, ADRs, and workflow memory as context, and follows the `rc-execute-task` contract: explore -> implement -> validate with `rc-final-verify` -> update tracking -> optional local commit.
5. A task that fails verification is marked `failed`; tasks depending on it are skipped with the reason recorded. Independent tasks still run.
6. Workflow memory is maintained across tasks via `rc-workflow-memory`.

## Phase 4b: Autonomous loop (optional, replaces phases 3-6)

For a migration or a large mechanical build-out, `/rc-loop` walks `.rc/ROADMAP.md` (authored by `/rc-roadmap`) one phase at a time: load lessons -> plan the phase's tasks -> execute -> verify -> record lessons -> flip the checkbox on a PASS. It runs only behind the four readiness questions in `skills/rc-loop/references/loop-readiness.md`, and it stops on its own when the roadmap is exhausted, a phase cannot reach green, or a decision the loop cannot assume comes up. Outward-facing actions (PR, push, Linear writes) are never autonomous.

## Phase 5: Review

Two paths are available:

### Path A: Manual AI Review

**Skill:** `/rc-review-round [feature-name]`

Invoke inside an agent session. The skill performs a comprehensive code review of the implementation and generates issue files under `.rc/tasks/<slug>/reviews-NNN/`.

### Path B: External Provider Review

**Review:** `/rc-review-round` (manual multi-lens review). External-provider auto-fetch (e.g. CodeRabbit) has no plugin-native equivalent.

Fetches review comments from an external provider (currently CodeRabbit) and writes them as issue markdown files under `reviews-NNN/`.

**Both paths produce:** `issue_*.md` files with YAML frontmatter containing round metadata (`provider`, `pr`, `round`, `round_created_at`) plus issue metadata (`status`, `severity`, `file`, `line`).

## Phase 6: Remediation

**Fix:** `/rc-fix-reviews`

1. The skill reads issue files from the latest (or specified) review round.
2. Each issue is triaged (valid/invalid), fixed if valid (in severity order), and verified with `rc-final-verify`.
3. Issue file frontmatter is updated: `pending` -> `valid`/`invalid` -> `resolved`.

**Iterate:** Repeat phases 6-7 until all reviews are clean, then merge.

## Phase 7: Archive

**Archive:** move `.rc/tasks/<slug>/` into `.rc/tasks/_archived/<timestamp>-<slug>/`.

**Eligibility:** every `task_NN.md` is `status: completed` and every issue under `reviews-NNN/` is `resolved` — read from the task and issue files themselves, which are the source of truth.

## Ad Hoc Execution

Prompt the agent/host directly — there is no `rc exec` wrapper. To keep the main context lean, delegate: `rc-explorer` for recon, `rc-librarian` for external docs, `rc-oracle` for hard review/decisions, `rc-fixer` for bounded implementation (see `delegation-contract.md`).

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

ADRs are created during ideation, PRD, and TechSpec phases to document significant decisions.

- **Location:** `.rc/tasks/<slug>/adrs/adr-NNN.md` (zero-padded 3-digit numbers).
- **Structure:** Status, Date, Context, Decision, Alternatives Considered, Consequences.
- **Referenced by:** PRDs, TechSpecs, and idea specs include an "Architecture Decision Records" section linking to all ADRs.
