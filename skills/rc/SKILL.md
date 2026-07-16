---
name: rc
description: Explains RC — a pure agent plugin (skills, commands, agents, hooks) for Claude Code, OpenCode and other tools — its workflow pipeline, artifact structure, bundled specialist agents, hooks, and configuration. Use when the user asks how to use RC, what skills or commands exist, or how the workflow pipeline works. Do not use to execute workflow steps — use the specific rc-* skills instead.
model: haiku
effort: low
---

# RC Reference Guide

Reference for RC, an agent plugin for AI-assisted development. RC is installed as a plugin — no
CLI, no binary, no daemon — and works across Claude Code, OpenCode, and other agent hosts.

## What Is RC

RC orchestrates the full lifecycle of AI-assisted development — product ideation, technical
specification, task decomposition, task execution, and PR-review remediation — entirely through
**skills, commands, agents, and hooks** that run inside your agent host.

- **Host-agnostic.** Ships for Claude Code, OpenCode, and other tools; each host loads the same
  skills, commands, agents, and hooks.
- **Skills-based.** Each workflow phase is a skill the agent follows.
- **Artifact-driven.** Planning and review artifacts live as markdown under `.rc/tasks/<slug>/`,
  versioned alongside the code. In monorepos a project may hold more than one `.rc`; the workflow
  skills discover them and ask which to use when several exist, falling back to a `.rc/` at the
  project root when none exists.
- **Local-first.** All state is plain files under `.rc/` — no external services.

## Installation

Install the RC plugin through your host's plugin/marketplace mechanism; the skills, commands,
agents, and hooks are auto-discovered. For **Claude Code**:
`/plugin marketplace add rodolfochicone/rc-project`, then `/plugin install rc@rc-project`
(auto-updates via `/plugin marketplace update`; commands are namespaced `/rc:rc-*`). GitHub read
access is required for the private repo (`gh auth login` or `GH_TOKEN`).

## Workflow Pipeline

Phases run in order; each produces artifacts consumed by the next.

1. **Requirements** — `/rc-create-prd` → `.rc/tasks/<slug>/_prd.md` + ADRs.
2. **Technical Design** — `/rc-create-techspec` → `_techspec.md` + ADRs.
3. **Task Decomposition** — `/rc-create-tasks` → `task_01.md … task_N.md` + master `_tasks.md`. Validate with `node "$CLAUDE_PLUGIN_ROOT/scripts/validate-tasks.mjs" --slug <slug>`.
4. **Execution** — `/rc-tasks-workflow <slug>` (Claude Code; drives the `Workflow` tool) or the `rc-execute-task` skill per task in dependency order (any host).
5. **Review** — `/rc-review-round` (manual multi-lens review) or `/rc-review-workflow` (Claude Code review→fix loop) → issue files under `reviews-NNN/`.
6. **Remediation** — `/rc-fix-reviews` triages, fixes, and verifies each review issue.
8. **Archive** — move a fully completed `.rc/tasks/<slug>/` into `.rc/tasks/_archived/`.

Repeat review/remediation until clean, then ship with `/rc-git`. For a step-by-step walkthrough,
read `references/workflow-guide.md`.

**Autonomous loop (opt-in, replaces phases 4-7).** For a migration or a large mechanical build-out,
`/rc-loop` walks `.rc/ROADMAP.md` (authored by `/rc-roadmap`) one phase at a time — planning,
executing and verifying each phase without per-step human gates, recording lessons (`rc-lessons`)
as it goes. It runs only behind a readiness gate (`skills/rc-loop/references/loop-readiness.md`):
a loop does not fix a weak harness, it compounds its errors. Ordinary feature work stays in the
human-gated pipeline above. Outward-facing actions (PR, push, Linear writes) are never autonomous.

## Core Skills

| Skill | Trigger | When To Use | Do Not Use For |
| --- | --- | --- | --- |
| `rc-create-prd` | `/rc-create-prd` | Building a Product Requirements Document | TechSpec, task breakdown, coding |
| `rc-create-techspec` | `/rc-create-techspec` | Translating a PRD into technical design | PRD creation, task execution |
| `rc-create-tasks` | `/rc-create-tasks` | Decomposing PRD+TechSpec into task files | Execution, review |
| `rc-execute-task` | (skill) | Executing a single task, in dependency order (portable, any host) | Review work, PR batches |
| `rc-tasks-workflow` | `/rc-tasks-workflow` | Running a slug's tasks via the Claude `Workflow` tool (Claude Code only) | Non-Claude hosts (use `rc-execute-task` per task), authoring tasks |
| `rc-review-round` | `/rc-review-round` | Comprehensive multi-lens code review | Fixing issues, task execution |
| `rc-review-workflow` | `/rc-review-workflow` | Review→fix→re-review loop via the Claude `Workflow` tool (Claude Code only) | Single-pass review, non-Claude hosts |
| `rc-fix-reviews` | `/rc-fix-reviews` | Remediating review-round issues | Fetching reviews, task execution |
| `rc-analyze` | `/rc-analyze` | Deep, evidence-based analysis of existing code | Reviewing a diff, editing code |
| `rc-final-verify` | `/rc-final-verify` | Enforcing verification before completion claims | Early planning, brainstorming |
| `rc-memory` | `/rc-memory` | Durable cross-session memory + learnings (`.rc/memory/`) | Task-scoped notes (use rc-workflow-memory) |
| `rc-workflow-memory` | (skill) | Cross-task workflow memory under `.rc/tasks/` | PR reviews, user preferences |
| `rc-roadmap` | `/rc-roadmap` | Authoring/reading `.rc/ROADMAP.md` — the epic phases a loop walks | Planning one phase's tasks (use rc-create-tasks) |
| `rc-loop` | `/rc-loop` | Walking the roadmap unattended, after the readiness gate passes (Claude Code only) | A single task (rc-execute-task), a fixed task set (rc-tasks-workflow), a weak harness |
| `rc-lessons` | (skill) | Recording/loading grounded lessons so a loop stops repeating its own bugs | Durable project facts (use rc-memory) |
| `rc` | `/rc` | Learning how to use RC | Executing workflow steps |
| `rc-fullstack-axum-svelte` | (skill) | Umbrella Axum + SQLx + SvelteKit (Bun) | React/Next, PRD/task pipeline phases |
| `rc-axum` | (skill) | Axum 0.8+ APIs, WS, middleware, security/tests | SvelteKit-only or SQLx-only work |
| `rc-sqlx` | (skill) | SQLx 0.8 + Postgres access, migrations, DB tests | HTTP routing alone, non-Rust ORMs |
| `rc-sveltekit` | (skill) | SvelteKit 2 SSR, forms, hooks, adapter-node, Bun | React/Next, Axum-only APIs |

## Bundled specialist agents

RC ships **leaf-worker agents** (under `agents/`, discovered as `rc:<name>`) to delegate to, each
on a cost-appropriate model tier. They carry no `Task`/`Agent` tool, so they cannot spawn further
subagents (the recursion cap).

| Agent | Lane | Model |
| --- | --- | --- |
| `rc-explorer` | read-only codebase recon (compressed map) | haiku |
| `rc-librarian` | external docs / web research | haiku |
| `rc-oracle` | architecture, review, hard debugging | opus |
| `rc-fixer` | bounded implementation | sonnet |

See `references/delegation-contract.md` for the routing table, the task-prompt contract, and
write-ownership rules for parallel fan-out.

## Bundled hooks

Configured in `hooks/hooks.json` (bash scripts under `hooks/scripts/`, gated by
`RC_HOOK_PROFILE=minimal|standard|strict`):

- **Guardrails** — `git-guard`/`commit-guard` (never auto-commit/push), `gateguard` (force
  grounding before the first edit of a file), `db-guard` (read-only database access by default).
- **Observability** — `observe` appends tool observations to `.rc/memory/observations.jsonl`
  for the `rc-memory` loop; `notify` on stop/notification.
- **Resilience** — `repair-guidance` injects corrective guidance when an `Edit`/`Task` returns a
  repairable failure, so the agent fixes the root cause instead of retrying the same failing call.

For detailed skill descriptions and inputs/outputs, read `references/skills-reference.md`.

## Artifact Directory Structure

```
.rc/
  tasks/
    <slug>/                            # One directory per workflow
      _prd.md                          # Product Requirements Document
      _techspec.md                     # Technical Specification
      _tasks.md                        # Master task list
      task_01.md ... task_N.md         # Individual task files
      adrs/                            # Architecture Decision Records
      reviews-NNN/                     # Review issues (round metadata in frontmatter)
      memory/                          # Per-workflow (task-scoped) memory
    _archived/<slug>/                  # Archived completed workflows
  memory/                              # Single durable memory + learning store (rc-memory)
    INDEX.md                           # Index of curated facts
    <slug>.md                          # One durable fact per file
    LEARNINGS.md                       # Distilled trigger→action learnings
    observations.jsonl                 # Raw tool observations (observe hook)
  analysis/                            # rc-analyze reports
```

## Reusable agents and the council pattern

Beyond the bundled specialist agents above, the plugin ships **council advisor** agents
used in a multi-perspective debate (see the `rc-council` skill):

| Agent | Perspective |
| --- | --- |
| `pragmatic-engineer` | Execution speed, maintenance burden |
| `architect-advisor` | Long-term coherence, boundaries, coupling |
| `security-advocate` | Attack vectors, compliance, data protection |
| `product-mind` | User impact, business value, opportunity cost |
| `devils-advocate` | Challenges assumptions, surfaces risks |
| `the-thinker` | Cross-domain patterns, structural reframing |

## Configuration

There is no runtime config file. Model and reasoning effort are set in each agent's and skill's
frontmatter (`model:` / `effort:`); hook behavior is set via the `RC_HOOK_PROFILE`
(minimal|standard|strict) and `RC_DISABLED_HOOKS` env vars. See `references/config-reference.md`.

## Common Patterns

- Follow the pipeline in order: idea (optional) → PRD → TechSpec → Tasks → Execution → Review → Fix.
- Validate a task set (`validate-tasks.mjs`) before running it, to catch metadata issues early.
- Delegate broad recon to `rc-explorer` and hard review/decisions to `rc-oracle` to keep the main
  context lean and costs down.
- Curate durable facts with `rc-memory`; let the `observe` hook + `rc-memory` capture
  recurring patterns.

## Anti-Patterns

- **Skipping pipeline stages.** Executing tasks without a PRD and task files produces poor results.
- **Mixing workflow skills out of order.** Running `/rc-create-tasks` without a PRD and TechSpec
  leads to shallow decomposition.
- **Editing task frontmatter blindly.** Run the task validator to catch metadata issues.
- **Skipping verification.** Always use `rc-final-verify` before claiming completion or committing.
