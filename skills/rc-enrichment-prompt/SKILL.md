---
name: rc-enrichment-prompt
description: Transforms a rough, vague, or one-line user request into a structured, high-signal prompt in markdown before any work begins — surfacing the real objective, implicit context, ambiguities, scope, and acceptance criteria so the executing agent (or another LLM) wastes no turns guessing. Use when a request is underspecified, when the user asks to "enhance/enrich/improve this prompt", or as a first pass before a large or fuzzy task. Do not use when the request is already precise and scoped, for trivial one-step edits, or to actually execute the task (this only rewrites the prompt, it does not implement).
argument-hint: "[raw prompt to enrich]"
model: sonnet
effort: medium
---

# Prompt Enrichment

Turn a raw request into an optimized, execution-ready prompt. The output is a **rewritten prompt**, not an implementation. Never start coding or answering the underlying task here — produce the enriched prompt and stop.

## Required input

- **raw prompt** (`$ARGUMENTS` when invoked as an argument, otherwise the user's latest request): the original instruction to enrich. If none is discernible, ask the user to paste it instead of inventing one.

## Procedure

**Step 1 — Analyze the raw prompt.** Identify, without guessing beyond the evidence:
- **Core objective** — the single outcome the user actually wants.
- **Implicit context** — files, systems, conventions, or prior decisions the user assumes but did not state. Inspect the repo/conversation to ground these; do not fabricate.
- **Ambiguities** — vague terms, missing inputs, undefined success. List each as an open question.
- **Scope & size** — one-liner vs multi-step; what is explicitly out of scope.

**Step 2 — Resolve or flag ambiguities.** For each ambiguity: if the codebase or conversation answers it, fill it in and note the evidence. If it genuinely needs the user, add it to an **Open questions** block rather than assuming. Never paper over a real fork with a default silently.

**Step 3 — Rewrite into structured markdown.** Produce the enriched prompt using this skeleton, dropping any section that does not apply (do not pad):

```markdown
## Objective
[Clear, direct statement of the outcome.]

## Context
[Relevant current state: files, modules, dependencies, conventions, prior art. Cite concrete paths/symbols when known.]

## Requirements
- [Concrete, testable requirement]
- [...]

## Constraints
- [Standards, patterns, "do not touch", performance/security limits — pull from project rules when present.]

## Acceptance criteria
- [How success is verified — the command to run, the behavior to observe, the test that must pass.]

## Out of scope
- [What NOT to do, to prevent scope creep.]

## Open questions
- [Only genuine blockers that need the user; omit the section if none.]
```

**Step 4 — Right-size.** Match the enriched prompt's weight to the task. A one-line change gets a short Objective + Acceptance criteria, not a full template. A large feature gets the full structure. Enrichment adds signal, never ceremony.

**Step 5 — Output.** Print only the enriched prompt (inside a fenced block so it is copy-pasteable) plus, if any, the Open questions. Do not begin the task. If invoked inside a RC flow, the enriched prompt is the input the next skill/command consumes.

## Rules

- **Preserve intent.** Sharpen and structure the user's request; never substitute a different task.
- **Evidence over invention.** Every "Context" claim must be grounded in the repo/conversation. Unknowns go to Open questions, not into assumptions.
- **No implementation.** This skill ends at the rewritten prompt. Executing it is a separate step.
- **Language.** Match the language of the original prompt.
