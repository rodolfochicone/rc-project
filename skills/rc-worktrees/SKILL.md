---
name: rc-worktrees
description: An opinionated protocol for using Git worktrees as isolated "lanes" for parallel, risky, or experimental work, tracked in a local manifest, so several efforts (or several agents) proceed without colliding in one working tree. Use to run tasks in parallel, to sandbox a risky refactor, or to review an integration before merging. Do not use for ordinary single-branch work, for rewriting shared history, or for force-pushing.
model: sonnet
effort: medium
user-invocable: true
argument-hint: "[create|list|integrate|cleanup] [slug]"
---

# Worktrees

Manage Git worktrees as safe, isolated coding lanes. One caller owns lane planning, path/branch selection, delegation, diff validation, integration, and cleanup — so parallel work never corrupts the main tree.

All worktrees live under a single path; never create them as siblings of the repo:

```text
.rc/worktrees/<slug>/
```

Branches are prefixed `rc/<slug>` and based off an explicit base (default `main`).

## State manifest — `.rc/worktrees.json`

Track lanes in a local manifest so the layout is inspectable:

```json
{
  "version": "1.0.0",
  "lanes": [
    {
      "slug": "auth-oauth2",
      "branch": "rc/auth-oauth2",
      "path": ".rc/worktrees/auth-oauth2",
      "base": "main",
      "purpose": "move auth to OAuth2",
      "areas": ["src/auth", "src/config"],
      "status": "active"
    }
  ]
}
```

`.rc/worktrees/` is transient and should be gitignored; the manifest may be committed to make lanes visible to the team.

## Protocol

1. **Create** — `git worktree add -b rc/<slug> .rc/worktrees/<slug> <base>`; record the lane in the manifest with its `purpose` and the file `areas` it owns.
2. **Assign ownership** — give each active lane a disjoint set of `areas`. Two lanes must not edit the same files concurrently; if they must, serialize them.
3. **Work** — run the task inside `.rc/worktrees/<slug>/`. Keep changes within the lane's declared areas.
4. **Validate** — before integrating, run the project's verification gate inside the lane and review its diff (`git -C .rc/worktrees/<slug> diff <base>...`).
5. **Integrate** — merge or rebase the lane branch back onto `<base>` once green; resolve conflicts in the owner lane.
6. **Cleanup** — `git worktree remove .rc/worktrees/<slug>` and delete the branch if merged; drop the lane from the manifest.

## Rules

- Confirm before any outward step (branch creation is local and fine; pushing a lane branch or opening a PR needs the same confirmation as `rc-git`).
- Never run destructive git commands (`reset --hard`, `clean`, history rewrite) inside a lane without explicit permission.
- Keep `.rc/worktrees.json` in sync with the actual worktrees (`git worktree list`); reconcile stale entries during cleanup.
