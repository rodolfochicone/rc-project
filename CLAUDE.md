# CLAUDE.md

Guidance for coding agents working in this repository.

## What this repo is

rc is a bundle of **AI-assisted development workflows** distributed as **skills, slash commands, hooks, and agents** for **Claude Code** and **OpenCode**. It is **plain markdown and shell** — there is no Go, no build step, and no binary. Do not add a compiled language, a package manager, or a build pipeline; if a change seems to need one, stop and discuss first.

## Layout

| Path              | Responsibility                                                       |
| ----------------- | ------------------------------------------------------------------- |
| `skills/`         | Skills — each is a directory with `SKILL.md` (+ optional `references/`) |
| `commands/`       | Claude Code slash commands (markdown with frontmatter)              |
| `agents/`         | Claude Code plugin agents (one phase agent per file)               |
| `hooks/`          | `hooks.json` + shell scripts run at agent lifecycle events          |
| `opencode/`       | OpenCode `agent/`, `commands/`, and the `plugin/rc-hooks.ts` plugin |
| `rules/`          | Coding rules injected into agent context                            |
| `.claude-plugin/` | Plugin (`plugin.json`) + marketplace (`marketplace.json`) manifests |
| `docs/`           | `claude-code-plugin.md` install/maintainer runbook                  |

## Editing rules

- **Skills** are markdown. A skill lives in `skills/<name>/SKILL.md` with YAML frontmatter (`name`, `description`). Keep the description a precise trigger — it is how the agent decides relevance.
- **Commands** are markdown with frontmatter under `commands/` (Claude) and `opencode/commands/` (OpenCode). Keep the two in sync when a command exists on both sides.
- **Hooks** are POSIX/bash shell scripts under `hooks/scripts/`, wired in `hooks/hooks.json`. Paths use `${CLAUDE_PLUGIN_ROOT}`. When adding or renaming a hook, update `hooks.json` **and** `opencode/plugin/rc-hooks.ts` so both runtimes stay at parity.
- **Keep Claude Code and OpenCode in parity.** A workflow change usually touches both `commands/` + `opencode/commands/` and, if it affects enforcement, `hooks/` + `opencode/plugin/rc-hooks.ts`.
- **The command sets differ by design, and that is intentional — do not "fix" it by forcing a 1:1 match.** OpenCode has per-phase commands (`rc-prd`, `rc-techspec`, `rc-tasks`, `rc-fix`) because its commands route to phase *agents*; on Claude those phases are reached by invoking the skills directly (`/rc:rc-create-prd`, …) plus the `rc-plan`/`rc-pipe` aggregate commands, so separate per-phase commands are redundant. `rc-docs` is Claude-only. Session hooks (`session-recall`, `phase-reminder`, `precompact-capture`) and `notify` are Claude-only; OpenCode gets equivalent behavior through the plugin's own event API, so they are not mirrored in `rc-hooks.ts` — only the guard hooks are.
- **Match existing style** in each file. Skills follow a consistent SKILL.md shape — read a neighbor before writing a new one.

## Do not

- Do not reintroduce Go, a `go.mod`, a `Makefile`, or any build tooling.
- Do not add a web app, a daemon, an HTTP API, or an SDK.
- Do not run destructive git commands (`git restore`, `git reset`, `git clean`, `git checkout -- …`, `git rm`) without explicit user permission.
- Do not commit, branch, or push without explicit authorization.

## Verifying a change

There is no `make verify`. Verify by inspection and by exercising the artifact:

- Skill/command markdown: confirm valid frontmatter and that the references it links to exist.
- Hooks: shell-check the script (`bash -n hooks/scripts/<name>.sh`) and, when practical, run it against a sample tool payload.
- Plugin manifests: confirm `plugin.json` / `marketplace.json` stay valid JSON after edits.
