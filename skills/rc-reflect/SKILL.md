---
name: rc-reflect
description: Reviews recent work to find repeated friction and recommends the smallest durable improvement — a new skill, agent, slash command, hook, project-memory fact, or instinct — grounded in evidence of what actually kept happening. Use when asked to learn from recent sessions, to turn a manually-repeated workflow into a reusable asset, or to improve your rc configuration based on real usage. Do not use for ordinary implementation, one-off debugging, or speculative asset creation without evidence (for atomic trigger→action corrections use rc-instincts; for durable facts use rc-project-memory).
model: sonnet
effort: high
user-invocable: true
argument-hint: "[focus]"
---

# Reflect

Look back over recent work and recommend the smallest useful improvement that would remove a *repeated* friction. The output is a ranked list of concrete suggestions with evidence — not new code, and not change for its own sake. "No change needed" is a valid, common answer.

## When to use

- The user asks to learn from recent sessions or to improve a recurring workflow.
- The user keeps doing something manually that could be a reusable asset.
- A recurring process might deserve to become a playbook or default.

Do not use for normal implementation, isolated bugs, or inventing assets nobody has needed twice.

## Evidence sources

Prefer evidence over intuition. Draw on:

- The current and recent session history (what was corrected, redone, or repeated).
- `.rc/instincts/INSTINCTS.md` and `.rc/instincts/observations.jsonl` (captured tool observations).
- `.rc/memory/` project facts and `.rc/tasks/*/memory/` workflow memory.
- Existing skills, commands, agents, hooks, and rules — so a suggestion reuses or extends what exists instead of duplicating it.

A pattern qualifies only if it appears **at least twice** (or once with clear cost). One-offs are not friction.

## What to recommend

Map each repeated pattern to the lightest asset that fixes it:

| Pattern | Recommend |
| --- | --- |
| A repeated trigger→action correction | an **instinct** (`rc-instincts`) |
| A durable project fact, convention, or gotcha | a **project-memory** file (`rc-project-memory`) |
| A multi-step workflow done by hand repeatedly | a **skill** or **slash command** |
| A rule that should be enforced, not just remembered | a **hook** |
| A recurring role or perspective | a **custom agent** |
| A constraint that belongs in every session | a **rule** or a `CLAUDE.md` line |

## Process

1. Gather evidence from the sources above, scoped to `focus` when given.
2. Cluster it into candidate patterns; drop anything seen only once without clear cost.
3. For each surviving pattern, propose the single lightest asset (table above), cite the evidence, and estimate the effort.
4. Rank by (frequency × cost saved) ÷ effort. Present the top few; explicitly say when the right answer is "no change."
5. Do not create the asset in this skill — hand off to the owning skill (`rc-instincts`, `rc-project-memory`) or propose the file for approval.
