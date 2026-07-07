---
name: rc-instincts
description: Distills atomic, confidence-scored "instincts" (trigger→action patterns) from captured tool observations and the current session, maintaining .rc/instincts/INSTINCTS.md so recurring corrections and workflows become durable, reusable guidance. Use to review or update learned instincts after a work session, or when asked to "learn" a repeated pattern. Do not use for one-off facts (use rc-project-memory), cross-task run memory (use rc-workflow-memory), or config security (use rc-audit).
model: sonnet
effort: medium
argument-hint: "[scope-note]"
---

# Instincts

Turn repetition into durable guidance. When the same correction, error resolution, or workflow shows up again and again, that pattern should stop being re-learned each session and become an **instinct** — a single `trigger → action` rule with a confidence score and evidence. This skill reads what was captured during work (the observation log and the current session) and maintains a small, curated `INSTINCTS.md` for the project. It is the lightweight, project-scoped core of continuous learning — no cross-project promotion, no automatic code generation; just a tended list of "how we do things here" that grows more confident with repetition and decays when contradicted.

## Inputs & location

- Resolve the project's `.rc` base directory (nearest `.rc` walking up; default `./.rc`). All paths below are relative to it.
- Raw capture (optional): `.rc/instincts/observations.jsonl`, appended by the `observe.sh` hook when `RC_INSTINCTS=1`. Each line is `{ts, tool, target}`. May be absent — then distill from the current session alone.
- Curated output: `.rc/instincts/INSTINCTS.md` (create if missing).

## The instinct format

`INSTINCTS.md` groups atomic instincts by domain. Each is one line, observable and actionable:

```
## code-style
- [0.7] when adding a new error path → wrap with `fmt.Errorf("…: %w", err)`  (evidence: corrected 3× in internal/; updated 2026-06-27; seen 4)

## workflow
- [0.5] when a task touches setup catalogs → update the exact-list test in bundle_test.go  (evidence: build broke once; updated 2026-06-27; seen 2)
```

- **Confidence** `0.3–0.9`. Start a fresh instinct at `0.4–0.5`. One trigger → one action; if you need "and", split it.
- **Domain** is a short tag (`code-style`, `workflow`, `testing`, `tooling`, `review`).
- **Evidence** is why it exists — the correction/error/repetition that justifies it. No evidence → not an instinct.

## Workflow

1. **Gather signal.** Read `observations.jsonl` if present, plus the current session: what the user corrected, what errors were resolved and how, which multi-step sequences repeated (3+ times), and stated tool/style preferences.
2. **Cluster into candidate instincts.** Group related signals into atomic `trigger → action` rules. Discard one-off, project-irrelevant, or already-obvious-from-CLAUDE.md patterns — instincts capture the non-obvious, repeated ones.
3. **Merge with existing `INSTINCTS.md` and evolve confidence:**
   - New pattern → add at `0.4–0.5` with evidence and `seen 1`.
   - Pattern that matches an existing instinct and recurred without being corrected → raise confidence (+0.1–0.2, cap `0.9`), bump `seen`, refresh the date.
   - Pattern the user **contradicted** this session → lower the matching instinct's confidence (−0.2–0.3); drop it if it falls below `0.3`.
   - Stale (no occurrence for a long stretch and low confidence) → drop it. Keep the file small and high-signal.
4. **Write `INSTINCTS.md`** sorted by domain then confidence (highest first). Then **prune** `observations.jsonl` (truncate it) so the next pass starts clean — the distilled instincts are the durable record, the raw log is transient.
5. **Promote the durable ones.** For any instinct that has reached high confidence (`≥0.8`) and is a genuine cross-task convention, record it via the `rc-workflow-memory` skill (and, if it is a lasting project convention/gotcha, as a `.rc/memory/` file) so it informs work even when this skill is not run. Instincts are the staging area; memory is the long-term store.
6. **Report** what changed: instincts added, reinforced, weakened, dropped, and promoted.

## Critical Rules

- Project-scoped only. Do not copy instincts between projects or to a global location — patterns from one stack mislead another.
- Every instinct needs evidence and an atomic trigger→action. No speculative or unsupported instincts.
- Confidence reflects repetition and lack of contradiction, not how clever the rule sounds. Lower it when the user pushes back.
- Keep `INSTINCTS.md` short — it is curated guidance, not a log. Prune aggressively; the raw log holds history.
- Never capture or store secrets or file contents — instincts are about patterns, not data. (`observe.sh` already records only tool + truncated target.)
- This skill writes only under `.rc/instincts/` (and promotes via the memory skills); it does not edit source code.

## Boundaries

The capture side is the `observe.sh` hook (opt-in via `RC_INSTINCTS=1`, shipped for both the Claude channel and opencode). This skill is the distill/curate side, run deliberately. It deliberately omits the heavier ECC ideas (cross-project promotion, clustering into generated skills/commands) — those are out of scope for this minimal loop.
