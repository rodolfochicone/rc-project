---
name: rc-new-project
description: Scaffolds a brand-new project by creating a private repository in the rodolfochicone organization from the rodolfochicone/typescript-template and cloning it into a subfolder of the current directory, guiding the user through any GitHub configuration problems. Use when the user wants to start a new TypeScript project from scratch. Do not use for adding files to an existing repo, cloning an arbitrary repository, or scaffolding from a non-rodolfochicone template.
model: sonnet
effort: medium
---

# New Project Scaffold

Take a developer from "empty folder" to "cloned, ready-to-code project" by creating a new repository from the rodolfochicone TypeScript template and cloning it locally. This skill is the agent-driven equivalent of the `rc init` command and is standalone — it does not read or require any `.rc/` workflow artifacts.

The repository is **always** created in the `rodolfochicone` organization, **private**, from the `rodolfochicone/typescript-template` template, and cloned into a subdirectory named after the repo inside the current working directory.

Ask every confirmation question in the user's language. Creating a remote repository is outward-facing: **never** create it without an explicit yes for that step. Use a TodoWrite list to track the phases and run them **in order**.

## Required Inputs

- The new repository name (a plain slug, e.g. `my-service`). If not provided, ask for it.
- Reject any name containing slashes or spaces — the owner is always the `rodolfochicone` org, so only a bare name is valid.

## Phase 0 — Check GitHub readiness

Inspect the real environment before doing anything outward-facing:

```bash
command -v gh                 # is the GitHub CLI installed?
gh auth status                # is gh authenticated?
pwd                           # the folder the project will be created in
ls -A                         # is the current folder empty / what is already here?
```

Base every later decision on this real state, not on assumptions.

## Phase 1 — Guard & guide configuration

Resolve any blocker before touching GitHub. Guide the user with concrete, copy-pasteable commands:

- **`gh` not installed** → stop and explain:
  - Install: https://cli.github.com (macOS: `brew install gh`).
  - Then `gh auth login`.
- **`gh` not authenticated** → stop and tell the user to run `gh auth login`, then confirm org access with `gh api user/memberships/orgs/rodolfochicone`.
- **No access to the `rodolfochicone` org** (e.g. the membership check fails) → stop and explain they need to:
  - Confirm membership: `gh api user/memberships/orgs/rodolfochicone`.
  - Ask an org admin for repository-creation permission.
  - Re-authenticate with org scope: `gh auth login --scopes "repo,read:org"`.

Only continue once `gh` is installed, authenticated, and the org is reachable.

## Phase 2 — Confirm the plan

Show the user exactly what will happen and get an explicit yes:

- Repository: `rodolfochicone/<name>` (private)
- Template: `rodolfochicone/typescript-template`
- Clone target: `./<name>` under the current directory

If `./<name>` already exists locally, stop and ask before proceeding.

## Phase 3 — Create & clone (confirm)

After the explicit yes, run a single command that creates the repo and clones it:

```bash
gh repo create rodolfochicone/<name> \
  --template rodolfochicone/typescript-template \
  --private \
  --clone
```

This creates `./<name>/` inside the current directory.

If the command fails:

- **Permission / `403` / "must be a member"** → this is an org-access problem; return to the configuration guidance in Phase 1.
- **Name already exists** → tell the user the repo name is taken and ask for a different one.
- **Any other error** → surface gh's own output verbatim; do not invent a cause.

## Phase 4 — Verify & report

Confirm the result is real before claiming success:

```bash
ls -A <name>                  # the clone exists and has the template files
git -C <name> remote -v       # origin points at rodolfochicone/<name>
```

Then report concisely: the repository URL, the local path (`./<name>`), and a suggested next step (`cd <name>` and install dependencies per the template's README).

## Guardrails

- Never create the repository in any owner other than `rodolfochicone`.
- Never create the repository without an explicit confirmation.
- Never fabricate a GitHub error cause — relay gh's real output and map it to the guidance above.
- Do not overwrite or clone into a non-empty existing `./<name>` directory without asking.
