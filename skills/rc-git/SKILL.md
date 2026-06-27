---
name: rc-git
description: Moves the current changes onto a proper feature branch, pushes it, and opens a PR — confirming each outward-facing step (branch, push, PR) and explicitly verifying the PR target branch. Use when the user wants to ship local work as a branch and pull request. Do not use for committing in place without a PR, rewriting shared history, or force-pushing.
model: haiku
effort: medium
---

# Git Branch, Push & PR

Take the work currently in this repo from "changes on disk" to "PR opened", stopping for the user's confirmation at every outward-facing step. This skill is standalone — it does not read or require any `.rc/` workflow artifacts.

Ask every confirmation question in the user's language. An optional Jira ticket (e.g. `PROJ-123`) may be supplied when the skill is invoked; if present, use it and skip the ticket question.

**Always write the PR title and description in Brazilian Portuguese (pt-BR)**, regardless of the conversation language. Keep the conventional-commit type prefix (`feat:`, `fix:`, etc.) in the title; everything else is in pt-BR.

Use a TodoWrite list to track the phases. Run them **in order** and **never** push or open a PR without an explicit yes for that specific step.

## Phase 0 — Gather git state

Before anything else, inspect the real repository state by running:

```bash
git rev-parse --is-inside-work-tree           # is this a git repo?
git branch --show-current                      # current branch (empty = detached HEAD)
git symbolic-ref --quiet refs/remotes/origin/HEAD | sed 's@^refs/remotes/origin/@@'  # default branch
git status --short                             # working tree changes
git log --oneline @{u}..HEAD                   # local commits ahead of upstream
git remote -v                                  # remotes
command -v gh                                  # is the gh CLI available?
```

Base every later decision on this real state, not on assumptions.

## Phase 1 — Inspect & guard

- If this is **not** a git repo → stop and say so. Don't run anything else.
- If `HEAD` is **detached** → stop and ask the user to check out a branch first.
- Determine the **default branch** (the remote HEAD from Phase 0; if unknown, ask the user — commonly `main` or `master`).
- If the working tree is clean **and** there are no local commits ahead of the upstream/default → there's nothing to ship: tell the user and stop.

## Phase 2 — Summarize the changes

Build a concise, human-readable summary of what was done, from the actual diff — not guesswork:

- Working changes: `git status --short`, `git diff --stat`, `git diff --cached --stat`; read the diffs as needed.
- Any local commits ahead: `git log --oneline <base>..HEAD`.

Group by area/file and describe the intent (what changed and why), in a few bullets. Reuse this summary for the branch name, commit message, and PR body.

## Phase 3 — Create the feature branch (only if on the default branch)

Skip this whole phase if the current branch is **not** the default branch — the work is already on a feature branch; go to Phase 4.

If on the default branch:

1. **Jira ticket** — if a ticket was supplied at invocation, use it. Otherwise ask via **AskUserQuestion** whether a Jira ticket should go in the branch name. **Default is no ticket** — only include one if the user provides it (they can type it via "Other").
2. **Propose branch names** — suggest 2–3 candidates via **AskUserQuestion** (recommended first, plus "Other" for a custom name). Build each as kebab-case from the summary:
   - Pick a type prefix that fits the change: `feat/`, `fix/`, `chore/`, `refactor/`, `docs/`, or `test/`.
   - No ticket (default): `<type>/<short-slug>` (e.g. `feat/git-branch-flow`).
   - With ticket: `<type>/<TICKET>-<short-slug>` (e.g. `feat/PROJ-123-git-branch-flow`).
   - Keep the slug short (3–5 words), lowercase, ASCII, hyphen-separated.
3. **Confirm creation** — show the chosen branch name **and** the commit message you'll use (subject from the summary; body with the bullets). Ask to proceed. On yes:
   - `git switch -c <branch>` (this carries the uncommitted working changes onto the new branch; the default branch stays untouched).
   - `git add -A && git commit` with the message. Do **not** add any co-author trailer or tool attribution to the commit message.
4. **Optional cleanup** — if the default branch had local commits ahead of its upstream (those commits are now also on the feature branch), offer to reset the local default branch back to its upstream so it stays clean (`git branch -f <default> <default>@{u}` — only while not checked out on it). Only offer this if it actually applies; let the user decline.

## Phase 4 — Push (confirm)

Confirm via **AskUserQuestion** before pushing. There must be a remote (`origin`); if none, say so and stop. On yes:

- `git push -u origin <current-branch>`.

Report the result (and the remote branch URL if the push output gives one).

## Phase 5 — Open the PR (confirm + confirm the target)

1. **Confirm the target base branch first** — detect the default branch as the suggested base, but **always ask** via **AskUserQuestion** which branch the PR targets, because some projects open PRs against a branch other than the default. Offer the detected default (recommended) and let the user pick another (list candidates from `git branch -r` or accept a custom one via "Other").
2. **Build the PR body in pt-BR** from the "PR description template" below, filling each section from the Phase 2 summary and the real diff. Omit sections that don't apply rather than writing filler. Use the supplied Jira ticket (if any) in "Issue relacionada".
3. **Confirm opening the PR** — show the title (commit subject, in pt-BR) and the rendered body and the resolved `head → base`. On yes, open it with `gh`:
   - Write the rendered body to a temp file (e.g. `"$TMPDIR/rc-pr-body.md"`) and pass it with `--body-file` — markdown templates are multi-line and break with an inline `--body "..."`:
     `gh pr create --base <target> --head <current-branch> --title "<title>" --body-file <path>`.
   - Do **not** add any tool attribution or "generated with" footer to the PR body.
   - If `gh` is **not installed**, don't fail silently: print the GitHub "compare" URL (`<repo-url>/compare/<target>...<current-branch>?expand=1`) so the user can open it manually.

Report the PR URL.

## PR description template

Render the PR body in **Brazilian Portuguese (pt-BR)** using the structure below. Fill each section from the Phase 2 summary and the real diff; omit any section that doesn't apply rather than leaving placeholders. In "Tipo de mudança", check the single box matching the branch/commit type prefix.

```markdown
## Resumo
<o que mudou e por quê, em 1–3 frases>

## Tipo de mudança
- [ ] Feature
- [ ] Correção de bug
- [ ] Refatoração
- [ ] Documentação
- [ ] Chore / build / infra

## Mudanças
- <um bullet por área/arquivo, a partir do resumo da Phase 2>

## Como testar
<passos para o revisor validar localmente: comandos, testes ou fluxo a exercitar>

## Checklist
- [ ] Testado localmente
- [ ] Testes adicionados/atualizados quando aplicável
- [ ] Sem alterações não relacionadas ao objetivo da PR

## Issue relacionada
<TICKET do Jira ou link; "N/A" se não houver>
```

## Guardrails

- Outward-facing actions (commit, push, PR) happen **only** after an explicit yes for that specific step — approval for one step does not authorize the next.
- Never force-push, never rewrite shared history, never touch unrelated branches.
- Don't invent a summary — base it on the real diff. If the diff is large, summarize honestly rather than omitting.
