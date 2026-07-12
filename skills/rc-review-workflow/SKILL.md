---
name: rc-review-workflow
description: Runs a review→fix→re-review loop over a PRD slug implementation by driving the Claude Code Workflow tool — fanning out multiple review lenses in parallel, synthesizing their findings into a reviews-NNN round, fixing the valid issues, and repeating until clean or a round cap. Use when the user wants an automated multi-lens review-and-remediate loop on a slug, orchestrated by Claude. Do not use for a single-pass review (use rc-code-review or rc-review-round), to fix an existing round (use rc-fix-reviews), or on non–Claude Code hosts (run the review skills manually).
argument-hint: "[slug]"
model: sonnet
effort: medium
---

# Review → fix → re-review as a Claude Workflow

Drive an automated **review → fix → re-review** loop over a feature slug's implementation using the Claude Code **`Workflow`** tool. Each round fans out several review lenses in parallel, synthesizes their findings into a `reviews-NNN/` round (the `rc-review-round` format), fixes the valid issues (the `rc-fix-reviews` contract), and repeats until a round surfaces **no new issues** or a round cap is hit. **Invoking this skill is the explicit opt-in to call the `Workflow` tool.**

This is **Claude Code only** — the `Workflow` tool does not exist in other agents. On other hosts, run the loop by hand: `/rc-review-round` then `/rc-fix-reviews`, repeated.

## When to stop and defer

- **No `Workflow` tool available** (non–Claude Code host) → stop and tell the user to run `/rc-review-round` + `/rc-fix-reviews` manually. Do not emulate the loop by hand here.
- **Slug not implemented** (`.rc/tasks/<slug>/` missing, or no task implementation to review) → stop; nothing to review yet.

## Phase 1 — Resolve the slug

Take `<slug>` from the argument and confirm `.rc/tasks/<slug>/` exists with its PRD/TechSpec and the implemented change set to review. The loop reviews the implementation of that slug, not an arbitrary diff.

## Phase 2 — Confirm parameters

Confirm with the user before launching (the loop edits the working tree during fix rounds):

- **Round cap** — max rounds (default **3**); the loop also stops early when a round finds zero new **high/critical** issues (any medium/low it found are still fixed that round, but they don't trigger another round).
- **Auto-commit** — whether `rc-fix-reviews` commits after a clean verify each round.
- **Lenses** — default all four below; let the user drop any.

## Phase 3 — Build and launch the Workflow

Call the `Workflow` tool with a script that loops rounds. Within each round:

- **Lenses run in parallel and are read-only** — they only analyze, so `parallel()` is safe.
- **Synthesis and fix are sequential and write** — one round at a time, so nothing fights over the working tree.

Each lens `agent()` follows the contract of its skill (invoke the skill if available as a tool in the subagent; otherwise follow the focus described). The synthesis `agent()` follows `rc-review-round` (dedup vs prior rounds, write only NEW issues, no directory when zero). The fix `agent()` follows `rc-fix-reviews` (triage → fix valid → `make verify` via `rc-final-verify` → set issue status).

Pass `args` as real JSON: `{ slug, maxRounds, autoCommit, lenses }`. Reference script shape:

```js
export const meta = {
  name: 'rc-review-workflow',
  description: 'Loop review→fix→re-review over a PRD slug across multiple lenses',
  phases: [{ title: 'Review' }, { title: 'Synthesize' }, { title: 'Fix' }],
}

const LENSES = [
  { key: 'defects',      skill: 'rc-code-review',         focus: 'correctness, security, performance, project conventions' },
  { key: 'simplify',     skill: 'rc-simplify-review',     focus: 'over-engineering, needless complexity, dead abstractions' },
  { key: 'architecture', skill: 'architectural-analysis', focus: 'dead code, duplication, architectural anti-patterns, type confusion' },
  { key: 'adversarial',  skill: 'rc-adversarial-review',     focus: 'refute the other lenses — what did they miss or get wrong; default skeptical' },
]

const FINDINGS = { type: 'object', required: ['lens', 'findings'], properties: {
  lens: { type: 'string' },
  findings: { type: 'array', items: { type: 'object', required: ['file', 'severity', 'title', 'detail'], properties: {
    file: { type: 'string' }, line: { type: 'integer' },
    severity: { type: 'string', enum: ['critical', 'high', 'medium', 'low'] },
    title: { type: 'string' }, detail: { type: 'string' },
  } } },
} }
const ROUND = { type: 'object', required: ['round', 'newIssues', 'newBlocking'], properties: {
  round: { type: 'integer' }, newIssues: { type: 'integer' },
  newBlocking: { type: 'integer' }, dir: { type: 'string' } } }
const FIX = { type: 'object', required: ['round', 'resolved', 'gate'], properties: {
  round: { type: 'integer' }, resolved: { type: 'integer' }, invalid: { type: 'integer' },
  gate: { type: 'string', enum: ['pass', 'fail'] } } }

const lenses = LENSES.filter(l => (args.lenses ?? LENSES.map(x => x.key)).includes(l.key))
const maxRounds = args.maxRounds ?? 3
let lastRound = 0
for (let i = 0; i < maxRounds; i++) {
  const lensResults = await parallel(lenses.map(l => () =>
    agent(
      `Review the implementation under .rc/tasks/${args.slug}/ through the ${l.key} lens (${l.focus}). ` +
      `Follow the ${l.skill} skill's contract; if that skill is available as a tool, invoke it. ` +
      `Read PRD/TechSpec for intent; skip issues the linter already catches. ` +
      `Return findings (file, line, severity, title, detail).`,
      { label: `review:${l.key}`, phase: 'Review', schema: FINDINGS })))
  const all = lensResults.filter(Boolean).flatMap(r => r.findings.map(f => ({ ...f, lens: r.lens })))

  const rr = await agent(
    `Acting as rc-review-round for feature ${args.slug}: dedup these multi-lens findings, drop any already tracked ` +
    `in prior reviews-NNN/ rounds, and write ONLY new issues as issue files under .rc/tasks/${args.slug}/reviews-NNN/ ` +
    `(next 3-digit round number). Create no directory if there are zero new issues. ` +
    `Return {round, newIssues, newBlocking, dir}, where newBlocking counts only new high- or critical-severity issues. ` +
    `Findings JSON:\n${JSON.stringify(all)}`,
    { label: 'synthesize', phase: 'Synthesize', schema: ROUND })
  log(`round ${rr.round}: ${rr.newIssues} new issues (${rr.newBlocking} high/critical)`)
  if (!rr.newIssues) { lastRound = rr.round - 1; break }

  const fx = await agent(
    `Acting as rc-fix-reviews for feature ${args.slug}, round ${rr.round}: triage every issue file in ${rr.dir}, ` +
    `fix the valid ones (root cause, severity order), run make verify via rc-final-verify, set each issue status. ` +
    `Auto-commit: ${args.autoCommit}. Return {round, resolved, invalid, gate}.`,
    { label: `fix:round-${rr.round}`, phase: 'Fix', schema: FIX })
  log(`round ${rr.round}: resolved ${fx.resolved}, gate ${fx.gate}`)
  lastRound = rr.round
  if (fx.gate !== 'pass') break
  if (!rr.newBlocking) break
}
return { lastRound, maxRounds }
```

## Phase 4 — Report

The `Workflow` tool runs in the background and notifies on completion; read its result and relay what matters. Summarize: rounds run, new issues per round (and how many were high/critical), resolved vs. invalid, the final gate result, and whether the loop ended **converged** (no new high/critical issues) or hit the **round cap** / a **red gate**. Point at the `reviews-NNN/` directories for the issue trail.

## Guardrails

- **Claude-only, opt-in via this skill.** Calling `Workflow` is justified only because the user invoked this skill. Never call it for unrelated work.
- **Lenses read-only, fix sequential.** Run the review lenses in parallel (they only analyze); run synthesis and fix one round at a time so nothing fights over the working tree.
- **No new high/critical is the exit signal.** A round whose new issues are all medium/low (or none at all) means converged — fix them, then stop. Otherwise keep looping until the round cap or a red `make verify`, and report what remains.
- **Verify gates every fix round.** A round's fixes count only after a clean `make verify` via `rc-final-verify`. Never mark issues resolved on a red gate.
- **Adversarial lens is in-Workflow, not cross-LLM.** Here the `adversarial` lens is a skeptical Claude subagent, not the real cross-model `rc-adversarial-review` (which spawns the opposite LLM). For a true cross-LLM pass, run `/rc-adversarial-review` separately.
