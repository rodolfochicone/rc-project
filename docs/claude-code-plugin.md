# Claude Code plugin — install and integration runbook

rc distributes its workflow skills and slash commands as a **Claude Code plugin** in
addition to the `rc setup` CLI installer. The plugin is an **additive, Claude-only**
channel: it ships the same skills and commands and auto-updates through the plugin
marketplace, pinned to each tagged release.

- **Marketplace name:** `rc-project`
- **Plugin name:** `rc`
- **Source of truth:** `.claude-plugin/marketplace.json` and `.claude-plugin/plugin.json`
  (kept in sync with the release version by `cmd/rc-plugin-sync`).

## Who this is for

- **End users** installing the plugin: follow [Install](#install) and [Update](#update).
- **Maintainers** validating a release: run the full [manual integration runbook](#manual-integration-runbook)
  before tagging, since plugin integration is validated manually by design (the automated
  gate is the manifest-consistency test under `make verify`).

## Prerequisite — GitHub access (repo is private)

`/plugin marketplace add rodolfochicone/rc-project` clones this repository, which is currently
private. The Claude Code environment must have GitHub read access first:

```bash
gh auth login                 # or:
export GH_TOKEN="$(gh auth token)"   # token with repo read scope (GITHUB_TOKEN also works)
```

Without it the marketplace add fails to clone. This is the same access `rc upgrade` and the
CLI update notifier require.

## Install

```text
/plugin marketplace add rodolfochicone/rc-project
/plugin install rc@rc-project
```

Plugin commands are namespaced under the plugin name and surface as:

- `/rc:rc-create-prd`
- `/rc:rc-create-techspec`
- `/rc:rc-create-tasks`
- `/rc:rc-review-round`
- `/rc:rc-final-verify`

**Pick one channel.** Use _either_ the plugin _or_ `rc setup` for Claude Code, not both —
installing the same skills through two channels produces duplicate commands. The plugin
covers Claude Code only; keep `rc setup` for other agents (codex, cursor-agent, droid,
gemini, …).

## Update

```text
/plugin marketplace update
```

The plugin `version` is pinned to each tagged release, so `/plugin marketplace update`
moves you to the version published with that tag. Because Claude Code installs the
plugin from the **git ref** (not from GoReleaser build artifacts), the pin must be
**committed before the tag is created** — see [Pinning the version before a release](#pinning-the-version-before-a-release).

## Pinning the version before a release

Claude Code serves the plugin from the repository **at the tagged ref** — the
marketplace `source` is `"."`, and `/plugin marketplace add rodolfochicone/rc-project`
clones the repo. The committed `.claude-plugin/*.json` at the tag is exactly what
users install; GoReleaser build artifacts are never consulted by the plugin
channel. The version bump must therefore be a **committed change that lands before
the tag**, not a release-time hook (a GoReleaser `before.hook` would run after the
tag and never commit its edits).

Maintainer pre-tag step, when cutting release `vX.Y.Z`:

```bash
go run ./cmd/rc-plugin-sync X.Y.Z   # stamps both manifests (v-prefix is normalized)
git add .claude-plugin/plugin.json .claude-plugin/marketplace.json
git commit -m "chore(release): pin plugin manifests to vX.Y.Z"
git tag vX.Y.Z                      # tag now carries the pinned manifests
```

The manifest-consistency test (`internal/pluginmanifest`, run under `make verify`)
keeps `plugin.json` and `marketplace.json` in agreement, but it does not enforce
that the committed version matches the tag — that is this manual step's job.

## Manual integration runbook

Run this end to end in a scratch Claude Code session before tagging a release. Each step
lists its expected result.

1. **Add the marketplace from a local path.** From a checkout of this repo:

   ```text
   /plugin marketplace add ./
   ```

   _Expected:_ Claude Code registers the `rc-project` marketplace from the local
   `.claude-plugin/marketplace.json`. (The local-path form avoids the private-repo clone and
   is the fastest way to validate uncommitted manifest changes.)

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

   _Expected:_ both resolve as plugin skills under the `rc:` namespace (alongside the rest of
   the `/rc:rc-*` set).

4. **Validate the update flow.** Bump the manifest `version` (locally, or via
   `go run ./cmd/rc-plugin-sync <new-version>`), then:

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
docs must match the manifests. The automated check lives in the manifest-consistency test
(`internal/pluginmanifest`), which runs under `make verify`. If you rename the plugin or
marketplace, update this runbook, the README "Install as a Claude Code plugin" section, and
`skills/rc/SKILL.md` to match.
