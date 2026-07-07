---
name: rc-doctor
description: Health-checks the rc installation — hook scripts present, syntactically valid and wired, manifests valid JSON with consistent versions, skill/command/agent frontmatter well-formed, Claude↔OpenCode guard-hook parity, and required tools (jq, gh) available — then prints a pass/fail report with the exact fix for each failure. Use after installing or updating the rc plugin, when a hook or command misbehaves, or before cutting a release of rc itself. Do not use to audit a project's own config for secrets and risks (use rc-audit) or to analyze context-window cost (use rc-context-budget).
model: haiku
effort: low
---

# Doctor

Diagnose the rc installation itself and report what is broken and how to fix it. Read-only: this skill never edits files, it prescribes.

## Locate the installation

Check in order and use the first that exists:

1. The current repo, if it is the rc source itself (`.claude-plugin/plugin.json` with `"name": "rc"` at the root).
2. The installed Claude Code plugin cache (`~/.claude/plugins/cache/*/rc/*/`).
3. An OpenCode install (`.opencode/` or `~/.config/opencode/` containing `plugin/rc-hooks.ts`).

Report which installation is being checked. If more than one exists, check the one for the runtime currently in use and mention the other.

## Checks

Run every check even after a failure; collect results.

1. **Manifests** — `jq empty` on `.claude-plugin/plugin.json`, `.claude-plugin/marketplace.json`, and `hooks/hooks.json`. The `version` in both manifests must match each other and the latest `CHANGELOG.md` release heading.
2. **Hook scripts** — every `command` in `hooks.json` resolves to an existing file under `hooks/scripts/`; every script passes `bash -n`; no script in `hooks/scripts/` (except `_lib.sh`) is left unwired.
3. **Guard-hook parity** — each guard hook wired in `hooks.json` (`git-guard`, `commit-guard`, `go-mod-guard`, `gateguard`, `go-fmt`, `observe`) is also referenced in `opencode/plugin/rc-hooks.ts`. Session hooks (`session-recall`, `phase-reminder`, `precompact-capture`) and `notify` are Claude-only by design — do not flag them.
4. **Skill frontmatter** — every `skills/*/SKILL.md` has `name` (matching its directory) and `description`; no description contains `<` or `>` (they break marketplace validation); every `references/` path a skill mentions exists.
5. **Commands and agents** — every file in `commands/`, `agents/`, `opencode/commands/`, and `opencode/agent/` has valid frontmatter with a `description`; every skill or agent a command routes to exists.
6. **Environment** — `jq` and `git` on PATH; `gh` on PATH (warn, not fail — only `rc-git` and `rc-new-project` need it); `.rc/` writable in the current project if one exists.

## Output

Print a table (`check | status | detail`), with status one of `pass` / `warn` / `fail`. For every non-pass row, give the exact command or edit that fixes it. End with a one-line verdict: healthy, or the count of failures to address.

## Critical Rules

- Read-only. Never fix anything yourself — prescribe the fix.
- Run all checks; never stop at the first failure.
- Distinguish design decisions from defects: Claude-only hooks and the intentionally different command sets (see `CLAUDE.md`) are not findings.
