# CLAUDE.md

This file provides project guidance for coding agents working in this repository.

## HIGH PRIORITY

- **IF YOU DON'T CHECK SKILLS** your task will be invalidated and we will generate rework
- **NEVER** use workarounds — always use the `no-workarounds` skill for any fix/debug task
- **NEVER** use web search tools to search local project code — for local code, use Grep/Glob instead
- This repository ships **plain markdown and small scripts** — there is no build, no binary, no `make`. "Done" means the content is coherent, every referenced file exists, and frontmatter is valid.

## Project Overview

RC is an **agent plugin** — skills, commands, agents, and hooks — that drives the full lifecycle
of AI-assisted development: optional ideation → PRD → TechSpec → tasks → execution → review →
remediation. It runs inside the agent host (Claude Code, OpenCode, and other tools); every
workflow artifact lives as plain markdown under a consumer project's `.rc/`. There is no CLI,
no daemon, and no database.

Distribution: the Claude Code plugin marketplace (manifests under `.claude-plugin/`). OpenCode
assets live under `opencode/`.

## Repository Layout

| Path             | Responsibility                                                             |
| ---------------- | -------------------------------------------------------------------------- |
| `skills/`        | All skills — one directory per skill: `SKILL.md` (+ `references/`, `assets/`, `scripts/`) |
| `commands/`      | Slash commands (`/rc-plan`, `/rc-exec`, `/rc-review`, `/rc-pipe`, …)       |
| `agents/`        | Reusable agents: cost-tiered leaf workers (`rc-explorer`, `rc-librarian`, `rc-fixer`, `rc-oracle`) and the council archetypes (`pragmatic-engineer`, `architect-advisor`, `security-advocate`, `product-mind`, `devils-advocate`, `the-thinker`) |
| `hooks/`         | Hook manifest (`hooks.json`) + guard/format/observe scripts                |
| `opencode/`      | OpenCode-specific assets (agents, commands)                                |
| `.claude-plugin/`| Claude Code plugin + marketplace manifests                                 |
| `docs/`          | Operational docs (install channels, plugin usage)                          |
| `scripts/`       | Utility scripts shipped to consumers (e.g. `validate-tasks.mjs`)           |
| `extensions/`    | Extension definitions                                                      |

## Authoring rules (skills, commands, agents, hooks)

- One directory per skill; the `SKILL.md` frontmatter needs `name` (kebab-case, matches the
  directory) and `description`. Deep content goes in `references/` — loaded on demand, never
  inlined into the description.
- **Descriptions are a context budget** (see the `rc-context-budget` skill): every skill/agent
  description is loaded in every session. State the trigger ("Use when…") and the anti-trigger
  ("Do not use for…") in a few lines; no marketing prose.
- Agents: frontmatter `name`, `description`, `tools` (least privilege), optional `model`
  (omit to inherit the session model) and `color`. One agent = one responsibility.
- Hooks must fail open (environment problems never block the tool call) and guard by file
  extension/command pattern — see `hooks/scripts/` for the house style.
- Before editing any skill, read it fully — including its `references/` — and keep every
  internal link valid. If a rule must hold every time, it belongs in a hook, not in prose.
- Never reference the retired `rc` CLI (`rc setup`, `rc exec`, `rc tasks run`, `rc reviews`) —
  hosts load the plugin directly; execution happens through host-owned tools.

## Verification before completion

1. Every file you touched parses: valid frontmatter, no broken relative links, referenced
   `references/`/`assets/` files exist.
2. `grep` the repo for the thing you renamed or removed — no dangling mentions.
3. Activate `rc-final-verify` before claiming any task is done.

## CRITICAL: Git Commands Restriction

- **ABSOLUTELY FORBIDDEN**: **NEVER** run `git restore`, `git checkout`, `git reset`, `git clean`, `git rm`, or any other git commands that modify or discard working directory changes **WITHOUT EXPLICIT USER PERMISSION**
- **DATA LOSS RISK**: These commands can **PERMANENTLY LOSE CODE CHANGES** and cannot be easily recovered
- **REQUIRED ACTION**: If you need to revert or discard changes, **YOU MUST ASK THE USER FIRST**
- If the worktree contains unexpected edits, read them and work around them; do not revert them

## Code Search and Discovery

- **TOOL HIERARCHY**: Grep/Glob for repository content; the Context7 MCP for external
  library documentation; web search only for web research — **never** for local project code.

## Anti-Patterns for Agents

1. **Skip skill activation** because "it's a small change" — every domain change requires its skill
2. **Bloat a description** — long frontmatter descriptions cost every session, everywhere
3. **Inline deep content in SKILL.md** that belongs in `references/` (progressive disclosure)
4. **Leave dangling references** after renaming/removing a skill, agent, or script
5. **Reintroduce CLI-era instructions** (`rc setup`, `rc exec`, `~/.rc/agents` provisioning)
6. **Run destructive git commands without permission** — `git restore`, `git reset`, `git clean` require explicit user approval
7. **Duplicate an existing skill** — search `skills/` before creating; extend or reference instead
