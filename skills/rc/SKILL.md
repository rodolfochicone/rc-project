---
name: rc
description: Explains rc capabilities, the core workflow skills, slash commands, artifact structure, and how the idea→PRD→techspec→tasks→execution→review pipeline works across Claude Code and OpenCode. Use when the user asks how to use rc, what commands are available, or how the workflow pipeline works. Do not use for executing workflow steps — use the specific rc-* skills instead.
model: haiku
effort: low
---

# rc Reference Guide

Comprehensive reference for the rc AI-assisted development workflow.

## What Is rc

rc is a bundle of **skills, slash commands, hooks, and agents** for **Claude Code** and **OpenCode**. It covers product ideation, technical specification, task decomposition, task execution, and PR review remediation — all driven from inside your coding agent.

Key characteristics:

- **Runs inside the agent.** Every phase is a skill or slash command executed by Claude Code or OpenCode. There is no separate binary or daemon.
- **Artifact-driven.** Planning and review artifacts live as markdown under `.rc/tasks/<slug>/`, versioned alongside the codebase. In monorepos a project may hold more than one `.rc` directory; the workflow skills discover the existing `.rc` directories and ask which one to use when several are present, falling back to a `.rc/` at the project root when none exists.
- **Markdown everywhere.** PRDs, specs, tasks, reviews, and ADRs are human-readable markdown — versioned, diffable, and editable between steps.

## Workflow Pipeline Overview

The standard development pipeline follows these phases in order. Each phase produces artifacts consumed by the next. Each is a slash command that invokes the matching skill.

1. **Requirements** — `/rc-create-prd` creates a business-focused Product Requirements Document at `.rc/tasks/<slug>/_prd.md` with ADRs.
2. **Technical Design** — `/rc-create-techspec` translates the PRD into a technical specification at `.rc/tasks/<slug>/_techspec.md` with ADRs.
3. **Task Decomposition** — `/rc-create-tasks` breaks down the PRD and TechSpec into independently implementable task files (`task_01.md`, `task_02.md`, …) and a master list at `_tasks.md`.
4. **Execution** — the `rc-execute-task` skill implements each task file in order, directly in the agent session — implement, validate, track, and (optionally) commit.
5. **Review** — `/rc-review-round` performs a comprehensive AI code review and writes review issue files under `reviews-NNN/`.
6. **Remediation** — the `rc-fix-reviews` skill triages, fixes, and verifies each review issue.

Repeat phases 5–6 until the review is clean, then merge.

For a detailed step-by-step walkthrough of each phase, read `references/workflow-guide.md`.

## Core Skills Summary

| Skill | Trigger | When To Use | Do Not Use For |
| --- | --- | --- | --- |
| `rc-create-prd` | `/rc-create-prd` | Building a Product Requirements Document | TechSpec, task breakdown, coding |
| `rc-create-techspec` | `/rc-create-techspec` | Translating PRD into technical design | PRD creation, task execution |
| `rc-create-tasks` | `/rc-create-tasks` | Decomposing PRD+TechSpec into task files | Execution, review |
| `rc-execute-task` | (invoked per task) | Executing a single PRD task end-to-end | Review work |
| `rc-review-round` | `/rc-review-round` | Performing comprehensive code review | Fetching external reviews, fixing |
| `rc-fix-reviews` | (invoked to remediate) | Fixing review issues | Task execution |
| `rc-final-verify` | `/rc-final-verify` | Enforcing verification before completion claims | Early planning, brainstorming |
| `rc-workflow-memory` | (invoked during execution) | Maintaining cross-task workflow memory | User preferences |
| `rc` | `/rc` | Learning how to use rc | Executing workflow steps |

## Artifact Directory Structure

```
.rc/
  tasks/
    <slug>/                            # One directory per workflow
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
      <timestamp>-<slug>/              # Archived completed workflows
```

## Common Patterns

- Follow the pipeline in order: PRD → TechSpec → Tasks → Execution → Review → Fix.
- Validate task frontmatter (required fields present) before executing tasks, to catch metadata issues early.
- Keep the tasks directory focused — move fully completed workflows under `_archived/`.

## Anti-Patterns

- **Skipping pipeline stages.** Executing tasks without a PRD and task files produces poor results.
- **Mixing workflow skills out of order.** Running `/rc-create-tasks` without a PRD and TechSpec leads to shallow task decomposition.
- **Confusing skills with slash commands.** Skills carry the workflow logic; slash commands (`/rc-create-prd`) are the entry points that invoke them.
- **Skipping verification.** Always use `rc-final-verify` before claiming task completion or creating commits.
