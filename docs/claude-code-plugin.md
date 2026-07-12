# Installing RC — Claude Code plugin & other hosts

RC ships as a **Claude Code plugin** (marketplace manifests under `.claude-plugin/`);
**other hosts** install by cloning this repository and symlinking its asset directories
into the host's config paths. There is no CLI installer — the retired `rc setup` flow
is fully replaced by the two channels below.

- **Marketplace name:** `rc-project`
- **Plugin name:** `rc`
- **Source of truth:** `.claude-plugin/marketplace.json` and `.claude-plugin/plugin.json`

## Prerequisite — GitHub access (repo is private)

`/plugin marketplace add rodolfochicone/rc-project` (and the git clone below) requires GitHub
read access to this private repository:

```bash
gh auth login                        # or:
export GH_TOKEN="$(gh auth token)"   # token with repo read scope (GITHUB_TOKEN also works)
```

## Claude Code (plugin marketplace)

```text
/plugin marketplace add rodolfochicone/rc-project
/plugin install rc@rc-project
```

Plugin commands are namespaced under the plugin name and surface as `/rc:rc-*`
(e.g. `/rc:rc-create-prd`, `/rc:rc-execute-task`). Skills, agents, and hooks are
auto-discovered from the plugin — nothing else to configure.

Update:

```text
/plugin marketplace update
```

The plugin is the **only** channel for Claude Code — do not also symlink this repo's
`skills/` into `~/.claude/skills/`, or every skill and command shows up twice.

## OpenCode and other hosts (clone + symlink)

Clone once, symlink, and `git pull` to update — the symlinks always serve the current
checkout:

```bash
git clone git@github.com:rodolfochicone/rc-project.git ~/dev/rc/rc-project

# 1. Skills — the cross-tool skills path, read natively by OpenCode (and other
#    hosts that support the .agents/skills convention):
mkdir -p ~/.agents
ln -sfn ~/dev/rc/rc-project/skills ~/.agents/skills

# 2. OpenCode agents, commands, and the hooks plugin:
mkdir -p ~/.config/opencode/agents ~/.config/opencode/commands ~/.config/opencode/plugins
for f in ~/dev/rc/rc-project/opencode/agent/*.md;    do ln -sfn "$f" ~/.config/opencode/agents/;   done
for f in ~/dev/rc/rc-project/opencode/commands/*.md; do ln -sfn "$f" ~/.config/opencode/commands/; done
ln -sfn ~/dev/rc/rc-project/opencode/plugin/rc-hooks.ts ~/.config/opencode/plugins/rc-hooks.ts
```

Notes:

- If `~/.agents/skills` already exists with other content, link per skill instead:
  `for d in ~/dev/rc/rc-project/skills/*/; do ln -sfn "$d" ~/.agents/skills/"$(basename "$d")"; done`
- Rules need no install step: OpenCode reads the consumer project's `AGENTS.md`
  (falling back to `CLAUDE.md`) natively.
- The agents under this repo's top-level `agents/` (leaf workers and council archetypes)
  are in Claude Code plugin format and ship through the plugin; the OpenCode-format
  variants live under `opencode/agent/`.
- Update everything with `git pull` in the checkout — no re-install needed.

## Version pinning before a release (maintainers)

Claude Code serves the plugin from the **git ref** — the committed `.claude-plugin/*.json`
at the tag is exactly what users install. The version bump must land **before** the tag.
The old `rc-plugin-sync` tool was retired with the CLI; edit both manifests by hand:

```bash
# bump "version" in .claude-plugin/plugin.json AND .claude-plugin/marketplace.json (keep them equal)
git add .claude-plugin/plugin.json .claude-plugin/marketplace.json
git commit -m "chore(release): pin plugin manifests to vX.Y.Z"
git tag vX.Y.Z
```

## Manual integration runbook

Run this end to end in a scratch Claude Code session before tagging a release.

1. **Add the marketplace from a local path.** From a checkout of this repo:

   ```text
   /plugin marketplace add ./
   ```

   _Expected:_ Claude Code registers the `rc-project` marketplace from the local
   `.claude-plugin/marketplace.json` (local-path form avoids the private-repo clone and
   validates uncommitted manifest changes).

2. **Install the plugin.**

   ```text
   /plugin install rc@rc-project
   ```

   _Expected:_ the `rc` plugin installs without error.

3. **Verify the skills resolve as plugin commands** (e.g. `/rc:rc-create-prd`,
   `/rc:rc-execute-task`).

   _Expected:_ they resolve under the `rc:` namespace alongside the rest of the `/rc:rc-*` set.

4. **Validate the update flow.** Bump the manifest `version` locally, then:

   ```text
   /plugin marketplace update
   ```

   _Expected:_ Claude Code reports the new version and applies the update.

### Cleanup

```text
/plugin uninstall rc@rc-project
/plugin marketplace remove rc-project
```

## Consistency

The marketplace name (`rc-project`), plugin name (`rc`), and slash commands referenced in
the docs must match the manifests. The automated manifest-consistency test was retired with
the Go codebase — when renaming the plugin or marketplace, check by hand and update this
runbook, the README "Installation" section, and `skills/rc/SKILL.md` together.
