# Contributing to rc

Thanks for your interest in contributing. This repository is **plain markdown and shell** —
skills, slash commands, hooks, and agents for Claude Code and OpenCode. There is no build step.

## Getting Started

```bash
git clone git@github.com:<your-user>/rc-project.git
cd rc-project
git remote add upstream git@github.com:rodolfochicone/rc-project.git
```

## Development Workflow

1. Sync your `main` with upstream before creating a branch.
2. Create a feature branch from `main`.
3. Make your changes.
4. Verify (see below).
5. Push to your fork and open a PR against `upstream/main`.

## Verification

There is no `make verify`. Check your change by inspection and by exercising the artifact:

- **Skills / commands** — confirm valid YAML frontmatter (`name`, `description`) and that any
  `references/` links resolve.
- **Hooks** — shell-check every script you touch (`bash -n hooks/scripts/<name>.sh`) and, when
  practical, run it against a sample tool payload.
- **Plugin manifests** — confirm `.claude-plugin/plugin.json` and `.claude-plugin/marketplace.json`
  stay valid JSON and share the same `version`.
- **Parity** — when changing a command or a hook, update both the Claude Code and OpenCode sides
  (`commands/` + `opencode/commands/`, `hooks/` + `opencode/plugin/rc-hooks.ts`).

## Skills

Each skill lives in `skills/<skill-name>/` with:

```
skills/<skill-name>/
  SKILL.md          # Skill definition with frontmatter
  references/       # Supporting persona and template files
```

When contributing a new skill or modifying an existing one:

- Follow the frontmatter format (`name`, `description`, optional `model`, `effort`).
- Keep references self-contained within the skill's `references/` directory.
- Write all skill content in English.

## Commit Messages

- Use [Conventional Commits](https://www.conventionalcommits.org/): `feat:`, `fix:`, `refactor:`, `docs:`, `test:`, `chore:`.
- Keep the subject line under 72 characters.
- Focus on **why**, not what.

## Pull Requests

- Keep PRs focused. One logical change per PR.
- Write a clear description with a summary.
- Link related issues when applicable.

## Maintainers

- Rodolfo Chicone (<rodolfo.chicone@escale.com.br>)

## License

By contributing, you agree that your contributions will be licensed under the [MIT License](LICENSE).
