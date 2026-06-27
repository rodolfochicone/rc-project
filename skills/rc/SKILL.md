---
name: rc
description: Explains rc capabilities, CLI commands, core workflow skills, optional extension skills, configuration, artifact structure, reusable agents, and extensions. Use when the user asks how to use rc, what commands are available, how the workflow pipeline works, or how to configure a workspace. Do not use for executing workflow steps — use the specific es- skills instead.
model: haiku
effort: low
---

# rc Reference Guide

Comprehensive reference for the rc CLI and its AI-assisted development workflow.

## What Is rc

rc is a Go CLI that orchestrates the full lifecycle of AI-assisted development. It covers product ideation, technical specification, task decomposition, automated execution via AI coding agents, and PR review remediation.

Key characteristics:

- **Agent-agnostic.** Supports claude, codex, copilot, cursor-agent, droid, gemini, opencode, and pi as ACP runtimes.
- **Skills-based.** Bundled skills (installed via `rc setup`) teach agents how to execute each workflow phase.
- **Artifact-driven.** Planning and review artifacts live as markdown under `.rc/tasks/<slug>/`, versioned alongside the codebase. In monorepos a project may hold more than one `.rc` directory; the workflow skills discover the existing `.rc` directories and ask which one to use when several are present, falling back to a `.rc/` at the project root when none exists.
- **Daemon-backed runtime.** A home-scoped daemon owns run state, workspace registration, snapshots, streams, and the synced `global.db` catalog under `~/.rc/`.
- **Single binary, local-first.** The daemon is launched from the same binary; there are no external control planes.

## Workflow Pipeline Overview

The standard development pipeline follows these phases in order. Each phase produces artifacts consumed by the next.

1. **Setup** -- `rc setup` installs core skills into target agents plus any setup assets shipped by enabled extensions. **Claude Code users** can instead install the skills and slash commands as a Claude Code plugin (`/plugin marketplace add rodolfochicone/rc-project` then `/plugin install rc@rc-project`), which auto-updates via `/plugin marketplace update`; plugin commands are namespaced as `/rc:rc-*`. The plugin is an additive, Claude-only channel -- pick either it or `rc setup` for Claude Code, not both. See the README "Install as a Claude Code plugin" section (the marketplace add clones the private repo, so GitHub read access via `gh auth login` or `GH_TOKEN` is required).
2. **Ideation** (optional) -- install and enable the first-party `rc-idea-factory` extension, run `rc setup`, then use `/rc-idea-factory` to expand a raw idea into a structured, research-backed spec at `.rc/tasks/<slug>/_idea.md`.
3. **Requirements** -- `/rc-create-prd` creates a business-focused Product Requirements Document at `.rc/tasks/<slug>/_prd.md` with ADRs.
4. **Technical Design** -- `/rc-create-techspec` translates the PRD into a technical specification at `.rc/tasks/<slug>/_techspec.md` with ADRs.
5. **Task Decomposition** -- `/rc-create-tasks` breaks down the PRD and TechSpec into independently implementable task files (`task_01.md`, `task_02.md`, etc.) and a master list at `_tasks.md`.
6. **Execution** -- `rc tasks run <slug> --ide <runtime>` dispatches task files sequentially to the configured AI agent for implementation.
7. **Review** -- `/rc-review-round` (manual AI review) or `rc reviews fetch <slug> --provider coderabbit --pr <N>` (external provider) produces review issue files under `reviews-NNN/`.
8. **Remediation** -- `rc reviews fix <slug>` processes review issues, triages, fixes, and verifies each one.
9. **Archive** -- `rc archive --name <slug>` moves fully completed workflows to `.rc/tasks/_archived/`.

Repeat phases 7-8 until the review is clean, then merge.

For a detailed step-by-step walkthrough of each phase, read `references/workflow-guide.md`.

## CLI Commands Quick Reference

| Command | Purpose | Key Flags |
| --- | --- | --- |
| **Setup & Config** | | |
| `rc setup` | Install core skills and enabled extension assets | `--agent`, `--skill`, `--global`, `--copy`, `--list`, `--all`, `--yes` |
| `rc upgrade` | Update CLI to latest release | |
| **Workflow Execution** | | |
| `rc daemon` | Manage the home-scoped daemon lifecycle | `start`, `status`, `stop` |
| `rc workspaces` | Inspect and manage daemon workspace registrations | `list`, `show`, `register`, `unregister`, `resolve` |
| `rc tasks run` | Execute PRD task files through the daemon | `--name`, `--attach`, `--ui`, `--stream`, `--detach`, `--task-runtime` |
| `rc exec` | Execute an ad hoc prompt | `--agent`, `--format`, `--prompt-file`, `--tui`, `--persist`, `--run-id` |
| `rc runs` | Attach, watch, and purge daemon-managed runs | `attach`, `watch`, `purge` |
| **Review** | | |
| `rc reviews fetch` | Fetch provider review comments | `--provider`, `--pr`, `--name`, `--round` |
| `rc reviews fix` | Process review issue files | `--name`, `--round`, `--concurrent`, `--batch-size`, `--ide` |
| **Utilities** | | |
| `rc tasks validate` | Validate task file metadata | `--name`, `--tasks-dir`, `--format` |
| `rc sync` | Reconcile workflow artifacts into daemon `global.db` | `--name`, `--root-dir`, `--tasks-dir` |
| `rc archive` | Move daemon-eligible completed workflows to archive | `--name`, `--root-dir`, `--tasks-dir` |
| `rc migrate` | Convert legacy artifacts to frontmatter | `--name`, `--dry-run`, `--reviews-dir` |
| **Agent Management** | | |
| `rc agents list` | List resolved reusable agents | |
| `rc agents inspect` | View agent definition and defaults | `<name>` |
| **Extensions** | | |
| `rc ext list` | List extensions | |
| `rc ext inspect` | View extension details | `<name>` |
| `rc ext install` | Install an extension from a local path or GitHub repo archive | `<source>`, `--remote`, `--ref`, `--subdir` |
| `rc ext uninstall` | Remove an extension | `<name>` |
| `rc ext enable/disable` | Toggle extension | `<name>` |
| `rc ext doctor` | Diagnose extension issues | |

Common flags shared by `tasks run`, `exec`, and `reviews fix`: `--ide`, `--model`, `--reasoning-effort`, `--add-dir`, `--auto-commit`, `--dry-run`.

For complete flag documentation, read `references/cli-reference.md`.

## Core Skills Summary

| Skill | Trigger | When To Use | Do Not Use For |
| --- | --- | --- | --- |
| `rc-create-prd` | `/rc-create-prd` | Building a Product Requirements Document | TechSpec, task breakdown, coding |
| `rc-create-techspec` | `/rc-create-techspec` | Translating PRD into technical design | PRD creation, task execution |
| `rc-create-tasks` | `/rc-create-tasks` | Decomposing PRD+TechSpec into task files | Execution, review |
| `rc-execute-task` | (internal) | Executing a single PRD task (called by `rc tasks run`) | Direct invocation, review work |
| `rc-review-round` | `/rc-review-round` | Performing comprehensive code review | Fetching external reviews, fixing |
| `rc-fix-reviews` | (internal) | Remediating review issues (called by `rc reviews fix`) | Fetching reviews, task execution |
| `rc-final-verify` | `/rc-final-verify` | Enforcing verification before completion claims | Early planning, brainstorming |
| `rc-workflow-memory` | (internal) | Maintaining cross-task workflow memory | PR reviews, user preferences |
| `rc` | `/rc` | Learning how to use rc | Executing workflow steps |

## Optional Extension Skills

| Skill | Trigger | When To Use | Install Flow |
| --- | --- | --- | --- |
| `rc-idea-factory` | `/rc-idea-factory` | Raw feature idea needs structured exploration before a PRD | `rc ext install --yes rc/rc --remote github --ref <tag> --subdir extensions/rc-idea-factory` -> `rc ext enable rc-idea-factory` -> `rc setup` |

For detailed skill descriptions and inputs/outputs, read `references/skills-reference.md`.

## Artifact Directory Structure

```
.rc/
  config.toml                          # Workspace configuration
  tasks/
    <slug>/                            # One directory per workflow
      _idea.md                         # Idea spec (from rc-idea-factory)
      _prd.md                          # Product Requirements Document
      _techspec.md                     # Technical Specification
      _tasks.md                        # Master task list
      task_01.md ... task_N.md         # Individual task files
      adrs/
        adr-001.md ... adr-NNN.md      # Architecture Decision Records
      reviews-NNN/
        issue_001.md ... issue_N.md    # Review issues with round metadata in frontmatter
      memory/
        MEMORY.md                      # Shared workflow memory
        task_01.md ... task_N.md       # Per-task memory
    _archived/
      <timestamp>-<slug>/             # Archived completed workflows
  agents/
    <name>/                            # Workspace-scoped reusable agents
      AGENT.md                         # Agent definition
      mcp.json                         # Optional MCP server config
  extensions/                          # Workspace-scoped extensions
```

Global paths:
- `~/.rc/agents/<name>/` -- global reusable agents (workspace overrides global)
- `~/.rc/extensions/` -- user-scoped extensions
- `~/.rc/runs/<run-id>/` -- daemon-managed run artifacts and persisted exec sessions
- `~/.rc/global.db` -- daemon workspace, workflow, task, and review catalog

## Configuration

Workspace defaults live in `.rc/config.toml`. CLI flags always override config values.

```toml
[defaults]
ide = "claude"
model = "opus"
auto_commit = true
reasoning_effort = "high"
add_dirs = ["../shared-lib"]

[tasks]
types = ["frontend", "backend", "docs", "test", "infra", "refactor", "chore", "bugfix"]

[tasks.run]
include_completed = false

[fix_reviews]
concurrent = 2
batch_size = 3

[fetch_reviews]
provider = "coderabbit"
nitpicks = false

[exec]
verbose = false
tui = false
persist = false
```

For all fields, types, and defaults, read `references/config-reference.md`.

## Reusable Agents and the Council Pattern

Reusable agents are standalone personas that can be invoked via `rc exec --agent <name>` or referenced by skills through `run_agent`.

**Discovery order:** workspace (`.rc/agents/<name>/`) overrides global (`~/.rc/agents/<name>/`).

**Agent definition:** Each agent has an `AGENT.md` with YAML frontmatter (`title`, `description`) and optional `mcp.json` for MCP server configuration.

**Council agents shipped by the optional `rc-idea-factory` extension**:

| Agent | Perspective |
| --- | --- |
| `pragmatic-engineer` | Execution-focused, delivery speed, maintenance burden |
| `architect-advisor` | Long-term system coherence, boundaries, coupling |
| `security-advocate` | Attack vectors, compliance, data protection |
| `product-mind` | User impact, business value, opportunity cost |
| `devils-advocate` | Challenges assumptions, surfaces risks, stress-tests |
| `the-thinker` | Cross-domain patterns, structural reframing |

Install flow: `rc ext install --yes rc/rc --remote github --ref <tag> --subdir extensions/rc-idea-factory` -> `rc ext enable rc-idea-factory` -> `rc setup`.

The `rc-idea-factory` skill uses these agents in a council debate to challenge feature scope and surface risks. The `council` skill can also orchestrate multi-advisor debates on demand.

Management commands: `rc agents list`, `rc agents inspect <name>`.

## Extensions

Executable plugins that extend rc at runtime via JSON-RPC 2.0 on stdin/stdout.

- **Three scopes:** bundled (shipped with rc), user (`~/.rc/extensions/`), workspace (`.rc/extensions/`). Workspace overrides user overrides bundled.
- **Disabled by default.** Enable explicitly with `rc ext enable <name>` or `--extensions` flag on `exec`.
- **Capabilities:** lifecycle observation, prompt decoration, plan injection, agent session modification, review provider registration.
- **SDKs:** TypeScript (`@rc/extension-sdk`), Go (`sdk/extension`).
- **Scaffolding:** `npx @rc/create-extension` generates extension boilerplate.

Management: `rc ext list`, `rc ext inspect <name>`, `rc ext install <source>`, `rc ext uninstall <name>`, `rc ext enable/disable <name>`, `rc ext doctor`.

## Common Patterns

- Run `rc setup` before starting any workflow to ensure core skills and enabled extension assets are installed.
- Follow the pipeline in order: idea (optional) -> PRD -> TechSpec -> Tasks -> Execution -> Review -> Fix.
- Configure workspace defaults in `.rc/config.toml` to reduce repetitive CLI flags.
- Run `rc tasks validate --name <slug>` before `rc tasks run` to catch metadata issues early.
- Use `rc archive` to clean up fully completed workflows and keep the tasks directory focused.
- Use `rc exec --agent <name>` for ad hoc prompts with a specific advisor perspective.
- Use `rc exec --persist` to save session artifacts for later resumption with `--run-id`.

## Anti-Patterns

- **Skipping pipeline stages.** Running `rc tasks run` without a PRD and task files produces poor results.
- **Invoking `rc-execute-task` directly.** Use `rc tasks run`, which handles dispatch, sequencing, memory, and tracking.
- **Mixing workflow skills out of order.** Running `/rc-create-tasks` without a PRD and TechSpec leads to shallow task decomposition.
- **Editing task file frontmatter manually.** Use `rc migrate` or `rc tasks validate` to fix metadata issues programmatically.
- **Confusing skills with CLI commands.** Skills (slash commands like `/rc-create-prd`) run inside an agent session. CLI commands (`rc tasks run`) run in the terminal.
- **Skipping verification.** Always use `rc-final-verify` before claiming task completion or creating commits.
