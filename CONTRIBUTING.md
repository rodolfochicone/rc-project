# Contributing to RC

Thanks for your interest in contributing. RC is a **pure agent plugin** — skills, commands, agents,
and hooks (plus a couple of small Node/Bash helper scripts). There is no compiled binary and no
build step.

## Prerequisites

- `git`
- `node` (runs the validation scripts) and `jq` (used by the bash hooks)
- Optionally, an agent host (Claude Code, OpenCode, …) to try a change end-to-end

## Getting Started

```bash
git clone git@github.com:<your-user>/rc-project.git
cd rc-project
git remote add upstream git@github.com:rodolfochicone/rc-project.git
node scripts/plugin-smoke.mjs   # must pass — validates all components
```

## Development Workflow

1. Sync your `main` with upstream before creating a branch.
2. Create a feature branch from `main`.
3. Make your changes — components are markdown, JSON, and small scripts.
4. Validate:
   ```bash
   node scripts/plugin-smoke.mjs               # frontmatter of skills/agents/commands + hook wiring
   node scripts/validate-tasks.mjs --selftest  # if you touched the task validator
   ```
   `plugin-smoke` is the blocking gate; it must report `OK` before you open a PR.
5. Push to your fork and open a PR against `upstream/main`.

## Contributing components

- **Skills** — `skills/<name>/SKILL.md` with frontmatter (`name`, `description` required; plus
  `model`, `effort`, `user-invocable`, `argument-hint` as needed). Keep supporting files
  self-contained in the skill's `references/` and `scripts/`. Write skill content in **English**.
- **Agents** — `agents/<name>.md` with frontmatter (`name`, `description` required; `tools`,
  `model`, `color`). Bundled specialists are leaf workers — do not give them the `Task`/`Agent`
  tool.
- **Commands** — thin `commands/<name>.md` wrappers that delegate to a skill.
- **Hooks** — bash under `hooks/scripts/`, wired in `hooks/hooks.json`; source `_lib.sh`, gate with
  `rc_hook_active`, fail open, and reference scripts with `${CLAUDE_PLUGIN_ROOT}`.
- **Keep docs in sync** — update the affected skill's description/references and
  `skills/rc/SKILL.md` when behavior changes.

## Commit Messages

- Use [Conventional Commits](https://www.conventionalcommits.org/): `feat:`, `fix:`, `refactor:`,
  `docs:`, `test:`, `chore:`.
- Keep the subject line under 72 characters.
- Focus on **why**, not what.

## Pull Requests

- Keep PRs focused. One logical change per PR.
- Write a clear description with a summary and how you validated it.
- Link related issues when applicable.
- `node scripts/plugin-smoke.mjs` must pass.

## Releasing

RC ships through the **Claude Code plugin marketplace**, which reads the version from the manifests
on `main` — there is no build or package step. To cut a release:

1. Bump `version` in **both** `.claude-plugin/plugin.json` and `.claude-plugin/marketplace.json`
   (semver: minor for new skills/agents/hooks, patch for fixes). Do **not** bump `package.json`
   or `extensions/*/extension.toml` — they version independently of the plugin.
2. In `CHANGELOG.md`, move `## [Unreleased]` to `## [X.Y.Z] - YYYY-MM-DD`, leave a fresh empty
   `[Unreleased]`, and update the link refs at the bottom: point `[Unreleased]` at
   `compare/vX.Y.Z...main` and add `[X.Y.Z]` → `releases/tag/vX.Y.Z`.
3. `node scripts/plugin-smoke.mjs` must pass.
4. Commit `chore(release): vX.Y.Z` — the two manifests and `CHANGELOG.md`, nothing else. Push to `main`.
5. Annotated tag: `git tag -a vX.Y.Z -m 'rc vX.Y.Z — <summary>'` and push it.
6. Publish the GitHub release: `gh release create vX.Y.Z --notes-file -`, using the notes from the
   version's CHANGELOG section.

Hosts pick up the new version via `/plugin marketplace update`. There is no npm channel today.

> **Gotcha:** when publishing several releases in one go, `gh` marks as _Latest_ the one **created
> last by date**, not the highest semver. Fix it with `gh release edit vHIGHEST --latest`.

## Maintainers

- Rodolfo Chicone (<rchicone103@gmail.com>)

## License

By contributing, you agree that your contributions will be licensed under the [MIT License](LICENSE).
