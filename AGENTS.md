# AGENTS.md

Guidance for coding agents working in **this repository**. RC is a **pure agent plugin** — a
collection of skills, commands, agents, and hooks (plus a couple of small Node/Bash helper
scripts). There is no compiled binary, no Go, no web/desktop app, and no build step.

## Project Overview

RC orchestrates the full lifecycle of AI-assisted development (ideation → PRD → TechSpec → tasks
→ execution → review → remediation) through skills the agent host loads. It ships for **Claude
Code, OpenCode, and other agent tools**; each host auto-discovers the plugin's components. All
runtime state is plain files under a project's `.rc/` directory.

## Repository Layout

| Path | Contents |
| --- | --- |
| `skills/<name>/SKILL.md` | Workflow skills (`rc-create-prd`, `rc-execute-task`, `rc-review-round`, …), each with frontmatter + optional `references/` and `scripts/`. |
| `agents/<name>.md` | Bundled specialist subagents (`rc-explorer`, `rc-librarian`, `rc-oracle`, `rc-fixer`) with `model:` tiers. |
| `commands/<name>.md` | Slash-command wrappers. |
| `hooks/hooks.json` + `hooks/scripts/*.sh` | Lifecycle hooks (bash), gated via `_lib.sh`. |
| `opencode/` | OpenCode-specific agent/command variants + `plugin/rc-hooks.ts`. |
| `extensions/rc-idea-factory/` | Bundled optional extension (skills + council agents). |
| `scripts/*.mjs` | Node helpers: `plugin-smoke.mjs` (component validation), `validate-tasks.mjs`. |
| `docs/` | Human-facing docs. |
| `.claude-plugin/` | `plugin.json` + `marketplace.json`. |

## Developing this repo

There is no build. Editing components is editing markdown, JSON, and small scripts.

```bash
node scripts/plugin-smoke.mjs           # validate all skills/agents/commands + hook wiring
node scripts/validate-tasks.mjs --selftest   # self-check the task validator
```

`node scripts/plugin-smoke.mjs` must pass before considering a change complete: it checks that
every `SKILL.md`/agent has valid frontmatter, commands are non-empty, and every hook command
points at a script that exists and is executable.

## Git — hands-off by default

- **NEVER** run `git restore`, `git checkout`, `git reset`, `git clean`, `git rm`, commit, push,
  or branch **without explicit user permission** — these can permanently lose work.
- If the worktree holds unexpected edits, read and work around them; do not revert them.

## Conventions

- **Skills** — `SKILL.md` frontmatter: `name`, `description` (required), plus `model`, `effort`,
  `user-invocable`, `argument-hint` as needed. The description drives when the skill fires; keep
  it precise, with a "Do not use for …" clause.
- **Agents** — `agents/*.md` frontmatter: `name`, `description` (required), `tools`,
  `model` (`haiku`/`sonnet`/`opus`/`inherit`), `color`. The bundled specialists are **leaf
  workers**: they carry no `Task`/`Agent` tool, so they cannot spawn subagents (the recursion cap).
- **Hooks** — bash, `set -u`, source `_lib.sh`, gate with `rc_hook_active <name> <profile>`
  (`minimal`|`standard`|`strict`), fail open on any environment problem, and feed the agent via
  `rc_block <name> <msg>` (exit 2 + stderr). Reference scripts with `${CLAUDE_PLUGIN_ROOT}`.
- **Commands** — thin `.md` wrappers; delegate to a skill rather than embedding logic.
- **Docs in sync** — when a skill's behavior, trigger, or output changes, update its description
  and any `references/` doc plus `skills/rc/SKILL.md`. If no doc change is needed, say so.
- **Match the surrounding style.** Read neighbouring components before adding a new one; reuse the
  established frontmatter fields and prose shape.

## Skill dispatch when using RC on a project

When these skills are installed in the host, prefer them for their domain: `rc-analyze`
(understand/trace existing code), `systematic-debugging` + `no-workarounds` (bug fixes — root
cause, never a workaround), `rc-final-verify` (before claiming any task done), and the workflow
skills for the pipeline phases. Activate every skill whose domain the change touches; a change
that spans domains needs all of their skills.

## Anti-patterns

1. Marking a change done without `node scripts/plugin-smoke.mjs` passing.
2. Editing a `SKILL.md`/agent and leaving `skills/rc/SKILL.md` (the front-door reference) stale.
3. Referencing a removed `rc <command>` (the CLI is gone — route to the plugin-native skill/script).
4. Running destructive git commands without explicit permission.
5. Giving a bundled specialist agent the `Task`/`Agent` tool (breaks the recursion cap).
6. Fixing a bug by patching the symptom instead of the root cause.
