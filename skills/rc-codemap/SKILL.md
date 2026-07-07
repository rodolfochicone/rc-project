---
name: rc-codemap
description: Builds and maintains a hierarchical repository map — one `codemap.md` per directory summarizing the WHY and HOW of that folder — so agents navigate a codebase from architectural summaries instead of re-reading source on every task. Use to onboard into an unfamiliar repo, to give the pipeline a token-efficient navigation layer, or to refresh the map after structural changes. Do not use to diagnose a specific bug (use rc-analyze), to review a change set (use rc-code-review), or to record durable prose facts (use rc-project-memory).
model: sonnet
effort: medium
user-invocable: true
argument-hint: "[path] [--refresh]"
---

# Codemap

Maintain a hierarchical `codemap.md` in each significant directory of the repository. A codemap describes a folder's *purpose, key files, and how it connects to its neighbors* — the durable architecture, not line-level detail. Agents read the maps to understand a codebase in seconds and to read only the folders a task actually touches.

## What a `codemap.md` contains

Keep each map short (aim for under ~40 lines). For the directory it lives in:

- **Purpose** — one or two sentences: why this directory exists.
- **Key files** — the handful that matter, each with a one-line role. Skip trivial files.
- **Entry points / public surface** — what the rest of the repo calls into here.
- **Depends on / used by** — the main inbound and outbound directories.
- **Subdirectories** — one line each, linking to their own `codemap.md`.

Describe design intent, not implementation. A codemap should stay valid across refactors that don't change the architecture.

## Building or refreshing the map

1. Determine scope: the given `path`, or the repo root if none. Respect `.gitignore`; skip `node_modules`, `.git`, `vendor`, `dist`, build output, and generated dirs.
2. Detect init vs. update:
   - **Init** — no `codemap.md` exists under scope: walk the tree bottom-up and write one map per significant directory.
   - **Update** — maps exist: only re-analyze directories whose contents changed since the map was last written (compare against git status / mtimes). Leave unchanged folders alone.
3. For each directory, read its files enough to summarize intent (prefer the local `codemap.md` of child dirs over re-reading their files), then write/overwrite its `codemap.md`.
4. Keep a root `codemap.md` as the index that links down into the tree.

## Using the map

- Before touching an unfamiliar area, read the nearest `codemap.md` (and the root index), then descend only into the folders the task needs.
- When you change a directory's architecture, update its `codemap.md` in the same change so the map never drifts.

## Rules

- Maps are committed documentation — keep them in git alongside the code.
- Never paste code blocks, full APIs, or exhaustive file lists into a map — that is how it rots into noise.
- If a map contradicts the code, trust the code and correct the map.
