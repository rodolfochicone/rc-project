# Workflow Guide

End-to-end walkthrough of the RC development pipeline from setup through archive.

## Prerequisites

1. **Install RC.** Ensure the `rc` binary is available in the system PATH.
2. **Install the plugin.** Install RC through your host's plugin/marketplace mechanism (Claude Code: `/plugin marketplace add rodolfochicone/rc-project` then `/plugin install rc@rc-project`). Skills, commands, agents, and hooks are auto-discovered.
3. **Ideation extension.** `/rc-idea-factory` ships bundled under `extensions/rc-idea-factory`; its skills and council agents are available once the plugin is installed.
4. **Configure workspace (optional).** Create `.rc/config.toml` to set default IDE, model, and other preferences. Read `config-reference.md` for all fields.

## Phase 1: Ideation (Optional)

**Skill:** `/rc-idea-factory [feature-idea]`

Use when a raw idea needs structured exploration before committing to a PRD.

Bundled: `extensions/rc-idea-factory` (skills + council agents), available with the plugin.

1. Invoke `/rc-idea-factory` inside an agent session with the feature idea.
2. Answer 3-6 clarifying questions (one per message, multiple-choice preferred).
3. The skill runs parallel codebase exploration and web research.
4. A business analyst persona evaluates viability with KPIs and scoring.
5. A council debate (3-5 advisors from the extension-shipped council agents) challenges scope and surfaces risks.
6. A product strategist scans for higher-leverage alternatives.
7. Review and approve the draft idea spec.
8. Output: `.rc/tasks/<slug>/_idea.md` + ADRs.

**Skip when:** The requirements are already well-understood and a PRD can be written directly.

## Phase 2: Requirements

**Skill:** `/rc-create-prd [feature-name-or-idea] [idea-file]`

1. Invoke `/rc-create-prd` with the feature name. If `_idea.md` exists, it is used as primary context.
2. The skill runs parallel codebase and market research.
3. Answer clarifying questions focused on WHAT and WHY (not HOW).
4. Choose from 2-3 product approaches. An ADR is created for the decision.
5. Review and approve the complete PRD draft.
6. Output: `.rc/tasks/<slug>/_prd.md` + ADRs.

**Key rule:** The PRD describes user capabilities and business outcomes only. No databases, APIs, frameworks, or architecture.

## Phase 3: Technical Design

**Skill:** `/rc-create-techspec [feature-name]`

1. Invoke `/rc-create-techspec` with the feature name.
2. The skill reads the existing `_prd.md` and explores the codebase architecture.
3. Answer technical clarifying questions.
4. Technical ADRs are created for architecture decisions.
5. Review and approve the TechSpec draft.
6. Output: `.rc/tasks/<slug>/_techspec.md` + ADRs.

**Contains:** System architecture, data models, core interfaces, API design, development sequencing.

## Phase 4: Task Decomposition

**Skill:** `/rc-create-tasks [feature-name]`

1. Invoke `/rc-create-tasks` with the feature name.
2. The skill loads the PRD and TechSpec, then breaks them into granular tasks.
3. Review the proposed task breakdown.
4. Task files are generated with YAML frontmatter: `status`, `title`, `type`, `complexity`, `dependencies`.
5. Tasks are enriched with codebase patterns and implementation context.
6. Validation runs via `node "$CLAUDE_PLUGIN_ROOT/scripts/validate-tasks.mjs" --slug <slug>`.
7. Output: `task_01.md` through `task_N.md`, `_tasks.md` master list.

**Task types:** `frontend`, `backend`, `docs`, `test`, `infra`, `refactor`, `chore`, `bugfix`. Custom types can be registered in `.rc/config.toml` under `[tasks].types`.

## Phase 5: Execution

**Run:** `/rc-tasks-workflow <slug>` (Claude Code) or the `rc-execute-task` skill per task in dependency order (any host)

1. RC reads task files from `.rc/tasks/<slug>/` in order, respecting dependencies.
2. The CLI auto-starts the home-scoped daemon when needed and starts the run through daemon transport.
3. For each pending task, RC constructs a prompt including the task spec, PRD, TechSpec, ADRs, and workflow memory.
4. The configured ACP runtime executes the task using the `rc-execute-task` skill.
5. Each task: read spec -> implement -> validate with `rc-final-verify` -> update tracking -> optional commit.
6. Workflow memory is maintained across tasks via `rc-workflow-memory`.

**Key flags:**
- `--auto-commit` -- create a local commit after each task completes cleanly.
- `--dry-run` -- generate prompts without running the IDE tool.
- `--include-completed` -- re-process tasks already marked as completed.

**Interactive mode:** In interactive terminals, `tasks run` attaches to the TUI by default; use `--ui`, `--stream`, `--detach`, or `--attach` to override that behavior.

## Phase 6: Review

Two paths are available:

### Path A: Manual AI Review

**Skill:** `/rc-review-round [feature-name]`

Invoke inside an agent session. The skill performs a comprehensive code review of the implementation and generates issue files under `.rc/tasks/<slug>/reviews-NNN/`.

### Path B: External Provider Review

**Review:** `/rc-review-round` (manual multi-lens review). External-provider auto-fetch (e.g. CodeRabbit) has no plugin-native equivalent.

Fetches review comments from an external provider (currently CodeRabbit) and writes them as issue markdown files under `reviews-NNN/`.

**Both paths produce:** `issue_*.md` files with YAML frontmatter containing round metadata (`provider`, `pr`, `round`, `round_created_at`) plus issue metadata (`status`, `severity`, `file`, `line`).

## Phase 7: Remediation

**Fix:** `/rc-fix-reviews`

1. RC reads issue files from the latest (or specified) review round.
2. For each batch of issues, the configured ACP runtime executes the `rc-fix-reviews` skill.
3. Each issue is triaged (valid/invalid), fixed if valid (in severity order), and verified with `rc-final-verify`.
4. Issue file frontmatter is updated: `pending` -> `valid`/`invalid` -> `resolved`.

**Key flags:**
- `--concurrent <N>` -- process N batches in parallel.
- `--batch-size <N>` -- group N file scopes per batch.
- `--include-resolved` -- re-process already-resolved issues.

**Iterate:** Repeat phases 6-7 until all reviews are clean, then merge.

## Phase 8: Archive

**Archive:** move `.rc/tasks/<slug>/` into `.rc/tasks/_archived/`

Moves fully completed workflows from `.rc/tasks/<slug>/` to `.rc/tasks/_archived/<timestamp>-<slug>/`.

**Eligibility:** Run `rc sync` first. Archive eligibility is computed from synced daemon state: all task items must be completed and all synced review issues must be resolved.

## Ad Hoc Execution

**Exec:** prompt the agent/host directly (there is no `rc exec` wrapper)

Execute a single prompt outside the pipeline workflow.

- **Reusable agents:** invoke a specialist agent directly (e.g. the `rc-oracle` agent) in your host.
- **Persistence:** `--persist` saves session artifacts. Resume with `--run-id <id>`.
- **TUI mode:** `--tui` opens an interactive terminal UI.
- **Output formats:** `--format text` (default), `json` (lean JSONL), `raw-json` (full event stream).

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
