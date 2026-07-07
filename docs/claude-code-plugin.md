# Claude Code plugin — install and integration runbook

rc distributes its workflow skills, slash commands, hooks, and agents as a **Claude Code
plugin**. Claude Code serves the plugin directly from this repository's layout (`skills/`,
`commands/`, `agents/`, `hooks/hooks.json`) and auto-updates through the plugin marketplace,
pinned to each tagged release.

- **Marketplace name:** `rc-project`
- **Plugin name:** `rc`
- **Source of truth:** `.claude-plugin/marketplace.json` and `.claude-plugin/plugin.json`
  (the two files must carry the same version — keep them in sync by hand).

For OpenCode, there is no marketplace; copy the `opencode/` bundle into your OpenCode config
(see the README).

## Prerequisite — GitHub access (repo is private)

`/plugin marketplace add rodolfochicone/rc-project` clones this repository, which is currently
private. The Claude Code environment must have GitHub read access first:

```bash
gh auth login                 # or:
export GH_TOKEN="$(gh auth token)"   # token with repo read scope (GITHUB_TOKEN also works)
```

Without it the marketplace add fails to clone.

## Install

```text
/plugin marketplace add rodolfochicone/rc-project
/plugin install rc@rc-project
```

Plugin commands are namespaced under the plugin name and surface as `/rc:rc-create-prd`,
`/rc:rc-create-techspec`, `/rc:rc-create-tasks`, `/rc:rc-review-round`, `/rc:rc-final-verify`,
and the rest of the `/rc:rc-*` set.

## Update

```text
/plugin marketplace update
```

The plugin `version` is pinned to each tagged release, so `/plugin marketplace update` moves
you to the version published with that tag. Because Claude Code installs the plugin from the
**git ref**, the pin must be **committed before the tag is created** — see
[Pinning the version before a release](#pinning-the-version-before-a-release).

## Pinning the version before a release

Claude Code serves the plugin from the repository **at the tagged ref** — the marketplace
`source` is `"."`, and `/plugin marketplace add rodolfochicone/rc-project` clones the repo. The
committed `.claude-plugin/*.json` at the tag is exactly what users install. The version bump
must therefore be a **committed change that lands before the tag**.

Maintainer pre-tag step, when cutting release `vX.Y.Z`:

```bash
# Set the same "version" in both manifests to X.Y.Z:
#   .claude-plugin/plugin.json
#   .claude-plugin/marketplace.json
git add .claude-plugin/plugin.json .claude-plugin/marketplace.json
git commit -m "chore(release): pin plugin manifests to vX.Y.Z"
git tag vX.Y.Z                      # tag now carries the pinned manifests
```

## Manual integration runbook

Run this end to end in a scratch Claude Code session before tagging a release. Each step lists
its expected result.

1. **Add the marketplace from a local path.** From a checkout of this repo:

   ```text
   /plugin marketplace add ./
   ```

   _Expected:_ Claude Code registers the `rc-project` marketplace from the local
   `.claude-plugin/marketplace.json`. (The local-path form avoids the private-repo clone and is
   the fastest way to validate uncommitted manifest changes.)

2. **Install the plugin.**

   ```text
   /plugin install rc@rc-project
   ```

   _Expected:_ the `rc` plugin installs without error.

3. **Verify the skills resolve as plugin commands.** In the session, confirm the namespaced
   commands are available, e.g.:

   ```text
   /rc:rc-create-prd
   /rc:rc-execute-task
   ```

   _Expected:_ both resolve as plugin skills under the `rc:` namespace.

4. **Validate the update flow.** Bump the manifest `version` in both files, commit, then:

   ```text
   /plugin marketplace update
   ```

   _Expected:_ Claude Code reports the new version and applies the update.

### Cleanup

Remove the scratch install when finished:

```text
/plugin uninstall rc@rc-project
/plugin marketplace remove rc-project
```

## Verifying manifest/doc consistency

The marketplace name (`rc-project`), plugin name (`rc`), and slash commands referenced in the
docs must match the manifests. If you rename the plugin or marketplace, update this runbook, the
README install section, and `skills/rc/SKILL.md` to match, and confirm `plugin.json` and
`marketplace.json` still carry the same `version`.
