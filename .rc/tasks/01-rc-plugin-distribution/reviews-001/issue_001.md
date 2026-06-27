---
provider: manual
pr:
round: 1
round_created_at: 2026-06-25T16:43:59Z
status: resolved
file: .goreleaser.oss.yml
line: 22
severity: high
author: claude-code
provider_ref:
---

# Issue 001: Version pinning never reaches plugin users

## Review Comment

The feature's core promise (ADR-002, task_02) is that the Claude Code plugin
`version` is pinned to each tagged release so `/plugin marketplace update` moves
users to the version published with that tag. The implemented mechanism does not
achieve this, for two compounding reasons:

1. **Wrong lifecycle point.** Claude Code installs the plugin from the **git
   repository at a ref** — `/plugin marketplace add rodolfochicone/rc-project` clones
   the repo and reads `.claude-plugin/marketplace.json` (`source: "."`). It does
   *not* consume GoReleaser build artifacts. The release flow tags `main` HEAD
   first and then runs GoReleaser; the `before.hooks` step
   `go run ./cmd/rc-plugin-sync {{ .Version }}` rewrites the manifests in the
   working tree **after** the tag already exists, and those edits are never
   committed or pushed. The tag users install therefore keeps the pre-sync
   version. The hook only mutates a throwaway working tree.

2. **Wrong config file.** The sync hook lives only in `.goreleaser.oss.yml`
   (the manual/local keyless OSS path). The CI release in
   `.github/workflows/release.yml` (Release job) invokes `goreleaser-pro` with
   the default `.goreleaser.yml`, which has no plugin-sync hook
   (`before.hooks` is just `make frontend-build`). So in the real CI release the
   sync never runs at all.

Net effect: after cutting `vX.Y.Z`, the committed `.claude-plugin/*.json` at the
tag still reads the previous version, and plugin users never receive the bump.
The `internal/pluginmanifest.Sync` logic and the consistency tests are correct —
the defect is purely *where* the sync is wired.

Suggested fix: make the version bump a **committed change that lands before the
tag**, not a GoReleaser before-hook. For example, run
`go run ./cmd/rc-plugin-sync <version>` as part of the release-PR / pre-tag step
(the same place the version is otherwise decided), commit the manifest change,
then tag — so the tag itself carries the pinned manifests. If a GoReleaser hook
is kept as a safety net, it must be in the config CI actually uses
(`.goreleaser.yml`) and paired with a commit-back step; otherwise it is dead
wiring. Also correct `docs/claude-code-plugin.md` (Update section), which
currently tells maintainers the OSS GoReleaser flow stamps the manifests — that
claim is inaccurate for what users actually install.

Affected files:
- `.goreleaser.oss.yml` (hook in the non-CI config)
- `.github/workflows/release.yml` (CI uses `.goreleaser.yml`, which lacks the hook)
- `docs/claude-code-plugin.md` (documents the pinning as effective)

## Triage

- Decision: `VALID`
- Root cause: the version-pin mechanism is wired as a GoReleaser `before.hook`
  (`.goreleaser.oss.yml`), which runs *after* the tag is placed and never commits
  its edits. Because Claude Code installs the plugin from the git ref at the tag
  (not from GoReleaser artifacts), the stamped version cannot reach users. The
  hook is therefore not just ineffective but also a dirty-tree hazard during
  `goreleaser release --clean`.
- Fix approach (surgical, avoids touching the private CI release machinery):
  1. Remove the dead/risky plugin-sync `before.hook` from `.goreleaser.oss.yml`.
     This also resolves the secondary "wrong config" point — the mechanism no
     longer lives in GoReleaser at all, so the absence of the hook in the
     CI-used `.goreleaser.yml` is moot.
  2. Re-document the pinning in `docs/claude-code-plugin.md` as a **committed
     pre-tag step**: maintainers run `go run ./cmd/rc-plugin-sync <version>`,
     commit the manifest change, then tag — so the tag itself carries the pinned
     manifests regardless of which release config runs.
- `.github/workflows/release.yml` is intentionally left untouched: moving the
  pin to a committed pre-tag step makes it config-independent, so editing the CI
  workflow is unnecessary and out of scope.
- Notes: `internal/pluginmanifest.Sync` and the consistency tests are correct and
  unchanged; only the wiring + docs are corrected.
