# AGENTS.md

This repository is a bundle of **AI-assisted development workflows** — skills, slash commands, hooks, and agents for **Claude Code** and **OpenCode**. It is **plain markdown and shell**: no Go, no build step, no binary.

The full working guidance lives in [`CLAUDE.md`](CLAUDE.md) and applies to every agent. In short:

- Edit markdown skills under `skills/<name>/SKILL.md`, commands under `commands/` (Claude) and `opencode/commands/` (OpenCode), hooks under `hooks/scripts/` (wired in `hooks/hooks.json`).
- Keep Claude Code and OpenCode at parity when changing commands or hook enforcement.
- Do **not** reintroduce Go, a build pipeline, a web app, or a daemon.
- Do **not** run destructive git commands or commit/branch/push without explicit permission.
- There is no `make verify`; verify by inspecting frontmatter/links and shell-checking hook scripts.
