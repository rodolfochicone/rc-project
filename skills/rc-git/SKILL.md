---
name: rc-git
description: Moves the current changes onto a proper feature branch, pushes it, and opens a PR — confirming each outward-facing step (branch, push, PR) and explicitly verifying the PR target branch. Also covers intelligent git rebase and merge-conflict resolution. Use when the user wants to ship local work as a branch and pull request, or needs to rebase a feature branch and resolve conflicts. Do not use for committing in place without a PR, or for force-pushing without going through the rebase guidance below.
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

## Rebase e resolução de conflitos

Handle git rebase operations and resolve merge conflicts intelligently while preserving features and maintaining code quality. Use this section when rebasing feature branches, resolving conflicts across commits, and ensuring clean linear history without losing changes.

### Quick start

For most rebases with multiple commits, use the squash-first strategy to resolve conflicts only once:

```bash
# Step 1: Backup current state
TIMESTAMP=$(date +%Y%m%d_%H%M%S)
git branch backup-rebase-$TIMESTAMP

# Step 2: Squash commits (interactive rebase on current branch)
git rebase -i $(git merge-base HEAD origin/main)

# Step 3: Rebase onto target
git rebase origin/main

# Step 4: If conflicts, resolve them once (see workflow below)
# Then continue: git rebase --continue

# Step 5: Force push safely
git push origin $(git rev-parse --abbrev-ref HEAD) --force-with-lease
```

This approach resolves conflicts once instead of per-commit, saving time and mental overhead.

### Core workflow: conflict analysis & resolution

Copy this checklist and mark progress:

```
Rebase Workflow:
- [ ] Step 1: Create safety backup
- [ ] Step 2: Fetch latest from target branch
- [ ] Step 3: Analyze conflict scope
- [ ] Step 4: Choose resolution strategy
- [ ] Step 5: Apply conflict resolutions
- [ ] Step 6: Validate merged code
- [ ] Step 7: Run tests
- [ ] Step 8: Force push safely
```

#### Step 1: Create safety backup

ALWAYS do this first. If rebase goes wrong, you can recover:

```bash
# Timestamped backup branch
TIMESTAMP=$(date +%Y%m%d_%H%M%S)
git branch backup-rebase-$TIMESTAMP

# Alternative: note your current HEAD SHA for manual recovery
git reflog
```

This costs nothing and saves hours of work if something goes wrong.

#### Step 2: Fetch latest changes

Ensure you have the most recent remote state:

```bash
# Fetch without modifying local branches
git fetch origin

# View the divergence
git log --oneline origin/main..HEAD  # Your commits
git log --oneline HEAD..origin/main  # New commits on main
```

Understand how many commits you're rebasing and how much main has changed.

#### Step 3: Analyze conflict scope

Before starting the rebase, predict conflicts:

```bash
# See which files you changed
git diff --name-only origin/main...HEAD

# See which files main changed
git diff --name-only origin/main HEAD

# Likely conflict areas: files changed in both
```

**Key insight**: If you changed `auth.ts` and so did main, you WILL get conflicts in `auth.ts`.

Anticipating conflicts helps you understand how to resolve them.

#### Step 4: Choose resolution strategy

**Strategy A: Squash first (recommended for 3+ commits)**

*When to use*: Multiple feature commits with many conflicts expected.

*Why*: Reduces conflicts to one resolution phase instead of per-commit.

```bash
# Interactive rebase on current branch first
git rebase -i $(git merge-base HEAD origin/main)

# In editor, change all "pick" to "squash" (or 's') except first commit
# Save and exit - commits are squashed into one
# Edit commit message to describe the entire feature

# Now rebase the squashed commit
git rebase origin/main

# Resolve conflicts once, then git rebase --continue
```

*Tradeoffs*: Lose individual commit history, but simpler conflict resolution.

**Strategy B: Interactive rebase with conflict awareness**

*When to use*: 1-2 commits, clean history, or complex per-commit logic.

```bash
git rebase -i origin/main

# In editor, you can:
# - Reorder commits to isolate conflict-prone ones
# - Drop commits that are already in main (git detects this)
# - Combine related commits before rebasing

# Save and exit - rebase proceeds, stopping at conflicts
```

*Tradeoffs*: More control, but more conflict-resolution iterations.

**Strategy C: Simple linear rebase (fastest, auto-resolution)**

*When to use*: Simple cases, no critical decisions, or in automated pipelines.

```bash
# Rebase all commits at once
git rebase origin/main

# If no conflicts, done
# If conflicts, you resolve each one
```

**Warning**: Not recommended for complex scenarios. Use Strategies A or B instead.

#### Step 5: Apply conflict resolutions

When `git rebase` pauses with conflicts:

```bash
# See which files conflict
git status

# For each conflicted file:
# - RECOMMENDED: Use merge tool for visual clarity
git mergetool --no-prompt

# - ALTERNATIVE: Manual edit in your editor
# Search for conflict markers: <<<<<<, ======, >>>>>>
```

**Conflict marker anatomy**

```javascript
<<<<<<< HEAD
// Your current feature code
function authenticate(token) {
  validateToken(token);
  return true;
}
=======
// Main branch code (incoming)
function authenticate(token) {
  if (!token) throw new Error("No token");
  validateToken(token);
  setSession(token);
  return true;
}
>>>>>>> origin/main
```

**Decision framework** (before deleting markers):

1. **Can you keep both?** YES → Merge them intelligently

   ```javascript
   function authenticate(token) {
     if (!token) throw new Error("No token"); // Keep main's validation
     validateToken(token);
     setSession(token); // Keep main's session setup
     return true; // Keep feature's return
   }
   ```

2. **Conflicting logic?** Understand WHY they differ, then decide
   - Did main add critical security checks? → Keep main's version
   - Did your feature add essential functionality? → Keep feature's version
   - Are they trying to do different things? → Combine intentionally

3. **Lost features?** NEVER let a feature silently disappear
   - If you added authentication logic, ensure it's in final version
   - If main improved database access, ensure that's preserved

**Key resolution principles**

DO:
- Keep both versions' important functionality when possible
- Use the merge tool for visual representation
- Add comments explaining merged conflicts: `// Merged from both versions: main's validation + feature's session setup`
- Test each file after resolution

DON'T:
- Mindlessly pick one version without understanding both
- Delete conflict markers without understanding the conflict
- Keep duplicate code - merge intelligently
- Skip testing before continuing

#### Step 6: Validate merged code

After resolving conflicts, validate the merged code:

```bash
# 1. Check syntax
npm run lint  # or eslint, pylint, etc.

# 2. Check types (if TypeScript)
npm run type-check  # or tsc --noEmit

# 3. Spot-check key files
git diff HEAD origin/main -- <conflicted-file>

# If validation fails:
# 1. Fix the issue in the file
# 2. git add <file>
# 3. git rebase --continue
```

**Important**: Validation catches mistakes BEFORE you commit them.

#### Step 7: Run tests

This is your safety net:

```bash
# Run full test suite
npm test

# Or specific tests for changed areas
npm test -- --testPathPattern=auth  # If auth.ts was changed

# If tests fail:
# 1. Understand what broke
# 2. Fix in the files
# 3. git add <files>
# 4. git rebase --continue
```

**Rule**: Never force-push code that fails tests.

#### Step 8: Force push safely

Use `--force-with-lease` instead of `--force`. It protects against accidentally overwriting others' work:

```bash
# SAFE: Protects others' commits
git push origin $(git rev-parse --abbrev-ref HEAD) --force-with-lease

# UNSAFE: Can overwrite others' work
git push origin your-branch -f  # Don't do this

# If force-with-lease fails:
# Someone else pushed to your branch
# Coordinate with them before forcing
```

### Common scenarios & strategies

**Scenario 1: Many small conflicts across 5+ commits** — use the squash-first strategy:

```bash
git rebase -i $(git merge-base HEAD origin/main)
# Mark all but first commit as 's' (squash)
# Save - commits squash into one

git rebase origin/main
# Resolve conflicts once
git rebase --continue

# Only one conflict-resolution phase!
```

*Why*: Each commit might have conflicts. Squashing before rebasing means one pass.

**Scenario 2: One specific commit has conflicts** — target that commit with interactive rebase:

```bash
git rebase -i origin/main

# In editor, move the problematic commit to the end
# Save - rebase proceeds, stopping at that commit

# When it stops, you know exactly which commit conflicts
git status  # See what changed in this commit

# Resolve, then continue
git rebase --continue
```

*Why*: Isolating the commit helps you understand what it's trying to do.

**Scenario 3: Conflicts keep repeating (same file, different commits)** — use git rerere (reuse recorded resolution):

```bash
# Enable rerere globally (one-time setup)
git config --global rerere.enabled true

# Now when you hit the same conflict in a second commit,
# Git automatically applies the first commit's resolution

git rebase origin/main

# Git remembers your first conflict resolution
# and replays it automatically for similar conflicts
```

*Why*: When rebasing and hitting the same file repeatedly, rerere saves manual work.

**Scenario 4: Rebase conflicts are too complex** — emergency escape plan:

```bash
# Abort the rebase - return to original state
git rebase --abort

# Fall back to merge (safer for complex scenarios)
git merge origin/main

# Or try a different approach:
# - Squash your entire feature branch first
# - Cherry-pick main's critical changes selectively
```

*Important*: It's okay to abort and rethink. Better than a broken rebase.

### Checklist: before you commit to rebasing

- [ ] Understand why you're rebasing (cleaner history, syncing with main, etc.)
- [ ] Backup your current branch
- [ ] Run tests on current branch - they should pass
- [ ] Fetch latest: `git fetch origin`
- [ ] Understand what you'll be rebasing (5 commits? 50?)
- [ ] Understand main's recent changes (1 commit? Major refactor?)
- [ ] Choose your strategy (squash-first, interactive, or simple linear — see above)
- [ ] Have your merge tool ready (if using GUI)
- [ ] Block 30 mins - don't rush conflict resolution
- [ ] Have tests ready to run after rebase

### When NOT to rebase

- Shared branches (use merge instead: `git merge origin/main`)
- Critical production code without comprehensive tests
- When multiple people are pushing to the same branch
- If you don't understand the conflicts you're seeing

**Default to merge if uncertain.** Merge is safer for collaborative work.

**Use rebase when**: Solo feature branch, clean history matters, no shared dependencies.

### Summary: the rebase philosophy

Rebasing is a tool for creating clean, linear commit history. **Used well**, it makes debugging and code review easier. **Used poorly**, it loses work.

**The key principle**: Understand every conflict before resolving it. Don't automate away the thinking.
