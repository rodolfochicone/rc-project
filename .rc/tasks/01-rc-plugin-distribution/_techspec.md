# TechSpec: Distribute rc skills and commands as a Claude Code plugin

> No `_prd.md` exists for this feature. This TechSpec is informed directly by user-provided
> context and codebase exploration; business framing is intentionally light.

## Executive Summary

Expose the bundled `rc-*` skills and slash commands through Claude Code's native plugin +
marketplace mechanism so Claude Code users can install and **auto-update** them
independently of the `rc` binary. The repository's root layout (`skills/` with
`SKILL.md` directories, `commands/` with flat `.md` files) already matches the plugin
layout exactly, so the change is additive: two manifest files at the repo root
(`.claude-plugin/plugin.json` and `.claude-plugin/marketplace.json`), a Go consistency
test, a release-time version bump, and documentation. No skill content is duplicated and
no existing install path (`rc setup` / `rc sync`) changes.

**Primary trade-off:** hosting the plugin in `rc-project` itself gives a single physical
source of truth with zero sync automation, but installing the plugin clones the entire
private repository (~140MB of git history) to deliver ~300KB of assets and requires git
auth (`GH_TOKEN`). We accept the heavier clone and auth requirement in exchange for
eliminating a second repo and a mirror pipeline.

## System Architecture

### Component Overview

- **`.claude-plugin/marketplace.json`** (new, repo root) — the catalog. Declares one
  marketplace (`rc-project`) listing a single plugin `rc` with `source: "."` and a pinned
  `version`. This is what `/plugin marketplace add rodolfochicone/rc-project` reads.
- **`.claude-plugin/plugin.json`** (new, repo root) — the plugin manifest. Plugin root is
  the repo root, so `skills/` and `commands/` are auto-discovered. Carries `name`,
  `version`, `description`, `author`.
- **`skills/` + `commands/`** (existing, unchanged) — serve a dual role: `//go:embed`
  source for the binary AND plugin component directories. The embed globs
  (`*/SKILL.md */references/*` and the commands embed) do not match `.claude-plugin/`, so
  adding the manifests is invisible to the binary build.
- **Manifest consistency test** (new Go test) — asserts the manifests stay valid and in
  sync with the embedded assets, so a renamed/added skill or malformed manifest fails
  `make test` rather than reaching users.
- **Release version bump** (modified release flow) — sets the manifest `version` to the
  release tag at release time.

### Data flow

`/plugin marketplace add rodolfochicone/rc-project` → Claude Code clones the repo, reads
`.claude-plugin/marketplace.json` → `/plugin install rc@rc-project` copies the plugin into
`~/.claude/plugins/cache` → `skills/` and `commands/` are discovered and exposed as
`rc:rc-*`. `/plugin marketplace update` re-reads the catalog; a changed `version` triggers
an update.

## Implementation Design

### Core Interfaces

The plugin manifest is the core data contract. The consistency test parses it with this
Go shape (added in the test package; no production type is required):

```go
// pluginManifest mirrors the subset of .claude-plugin/plugin.json the test validates.
type pluginManifest struct {
    Name        string `json:"name"`        // must be "rc"
    Version     string `json:"version"`     // non-empty; pinned to release tag
    Description string `json:"description"`
}

// marketplaceManifest mirrors .claude-plugin/marketplace.json.
type marketplaceManifest struct {
    Name    string `json:"name"`
    Plugins []struct {
        Name    string `json:"name"`    // must include "rc"
        Source  string `json:"source"`  // "."
        Version string `json:"version"` // non-empty
    } `json:"plugins"`
}
```

### Data Models

`.claude-plugin/plugin.json`:

```json
{
  "name": "rc",
  "version": "0.14.0",
  "description": "RC AI-assisted development workflow skills and commands",
  "author": { "name": "Escale" }
}
```

`.claude-plugin/marketplace.json`:

```json
{
  "name": "rc-project",
  "owner": { "name": "Escale" },
  "plugins": [
    {
      "name": "rc",
      "source": ".",
      "version": "0.14.0",
      "description": "RC workflow skills and slash commands"
    }
  ]
}
```

### API Endpoints

Not applicable — no HTTP surface. User-facing entry points are Claude Code slash commands:

- `/plugin marketplace add rodolfochicone/rc-project`
- `/plugin install rc@rc-project`
- `/plugin marketplace update` (pull new tagged versions)
- Invocation post-install: `/rc:rc-create-prd`, `/rc:rc-execute-task`, etc.

## Integration Points

- **Claude Code plugin runtime** — consumes the manifests and component directories.
  Auth: `/plugin marketplace add` against the private `rodolfochicone/rc-project` repo requires
  git credentials (`GH_TOKEN` or configured SSH); the same applies to background
  auto-update. Documented as a prerequisite.
- **OSS GoReleaser release flow** — releases always run through `.goreleaser.oss.yml`
  (keyless OSS build, tags `vX.Y.Z`); this is the only release path. It gains a step that
  writes the release version into the manifests so the pinned `version` tracks the tag.

## Impact Analysis

| Component | Impact Type | Description and Risk | Required Action |
|-----------|-------------|---------------------|-----------------|
| `.claude-plugin/plugin.json` | new | Plugin manifest at repo root. Low risk. | Create file |
| `.claude-plugin/marketplace.json` | new | Marketplace catalog at repo root. Low risk. | Create file |
| `skills/embed.go`, `commands/embed.go` | unchanged | Embed globs exclude `.claude-plugin/`; dual-use confirmed. Risk: a future glob change could accidentally embed manifests. | Verify build; no edit |
| Manifest consistency test | new | Guards manifest validity + asset sync in `make test`. Low risk. | Add `*_test.go` |
| Release flow (`.goreleaser.oss.yml` / `Makefile`) | modified | Bump manifest `version` to tag at release. Risk: forgotten bump silently stops updates. | Add release step |
| Docs (`README.md`, `CLAUDE.md`, `skills/rc/SKILL.md`) | modified | Document the plugin install path as additive to `rc setup`. Low risk. | Update docs |

## Testing Approach

### Unit Tests

- Parse `.claude-plugin/plugin.json`: valid JSON, `name == "rc"`, non-empty `version`.
- Parse `.claude-plugin/marketplace.json`: valid JSON, lists a plugin named `rc` with
  `source == "."` and non-empty `version`.
- Assert manifest/asset coherence against the embedded FS: every directory matched by
  `skills.FS` has a `SKILL.md`, and `commands.FS` exposes the expected `.md` files, so an
  added/renamed asset that breaks discovery fails the test. Reuse `skills.FS` / the
  `commands` package rather than re-globbing the filesystem.
- Use table-driven subtests with `t.Run`; no network, no `rc setup` invocation.

### Integration Tests

- Manual (documented runbook), not automated: add the marketplace from a **local path**
  (`/plugin marketplace add ./`) in a scratch Claude Code session, run
  `/plugin install rc@rc-project`, confirm `/rc:rc-*` skills and commands resolve and that
  `/plugin marketplace update` picks up a bumped `version`. Local-path add avoids needing
  the private remote during development.

## Development Sequencing

### Build Order

1. **Add `.claude-plugin/plugin.json`** at repo root — no dependencies.
2. **Add `.claude-plugin/marketplace.json`** at repo root — depends on step 1 (references
   the plugin `name`).
3. **Add the manifest consistency test** — depends on steps 1–2 (parses both manifests);
   must pass `make verify`.
4. **Wire the release version bump** into the OSS GoReleaser flow (`.goreleaser.oss.yml`,
   the only release path) — depends on step 2 (the field it rewrites).
5. **Update documentation** (`README.md`, `CLAUDE.md`, `skills/rc/SKILL.md`) describing the
   additive Claude-only install path — depends on steps 1–2.
6. **Validate** with the manual integration runbook — depends on steps 1–3.

### Technical Dependencies

- Claude Code version recent enough for plugin marketplaces (current releases support it).
- Git auth (`GH_TOKEN`) available to users for the private-repo marketplace add.

## Monitoring and Observability

Not applicable — distribution is client-side via Claude Code. The only operational signal
is the consistency test in CI (`make verify`); a failing build is the alert that manifests
drifted from assets. No runtime logging or metrics are introduced.

## Technical Considerations

### Key Decisions

- **Decision:** Host the single-plugin marketplace inside `rc-project`.
  **Rationale:** the root layout already conforms; single source of truth, no mirror.
  **Trade-off:** heavy clone (~140MB) + private-repo auth.
  **Alternatives rejected:** dedicated public repo + CI mirror (second source of truth);
  status quo (updates coupled to binary). See ADR-001.
- **Decision:** Pin `version` to release tags. **Rationale:** stable, predictable updates
  aligned to the existing `vX.Y.Z` cadence. **Trade-off:** version must be bumped per
  release. **Alternatives rejected:** commit-tracking, SHA pinning. See ADR-002.
- **Decision:** Keep the `rc-` name prefix; plugin is additive and Claude-only.
  **Rationale:** surgical — `rc setup`, cross-references, and memory keep working.
  **Trade-off:** verbose `rc:rc-*` invocation; double listing if both channels used.
  See ADR-003.

### Known Risks

- **Forgotten version bump** (medium likelihood) — users silently stop receiving updates.
  Mitigation: fold the bump into the release pipeline (step 4) and document it.
- **Heavy/private clone friction** (high likelihood for new users) — large download and
  auth requirement. Mitigation: document the `GH_TOKEN` prerequisite; revisit a dedicated
  public repo (ADR-001 alternative) if friction proves material.
- **Embed glob regression** (low likelihood) — a future change to the embed directives
  could pull `.claude-plugin/` into the binary. Mitigation: the consistency test and a
  passing `make build` guard against it.

## Architecture Decision Records

- [ADR-001: Distribute rc skills and commands as a Claude Code plugin hosted in rc-project](adrs/adr-001.md) — Host a single-plugin marketplace in the repo itself; single source of truth over a lighter mirror repo.
- [ADR-002: Pin plugin updates to release tags via the `version` field](adrs/adr-002.md) — Pinned `version` aligned to `vX.Y.Z` tags for stable updates over commit-tracking.
- [ADR-003: Keep the `rc-` name prefix; plugin is an additive Claude-only channel](adrs/adr-003.md) — Surgical naming; `rc setup` stays the path for non-Claude agents.
