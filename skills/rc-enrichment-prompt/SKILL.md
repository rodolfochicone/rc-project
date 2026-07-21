---
name: rc-enrichment-prompt
description: Transforms a rough, vague, or one-line user request into a structured, high-signal prompt in markdown before any work begins — surfacing the real objective, implicit context, ambiguities, scope, and acceptance criteria so the executing agent (or another LLM) wastes no turns guessing. Use when a request is underspecified, when the user asks to "enhance/enrich/improve this prompt", or as a first pass before a large or fuzzy task. Do not use when the request is already precise and scoped, for trivial one-step edits, or to actually execute the task (this only rewrites the prompt, it does not implement).
argument-hint: "[raw prompt to enrich]"
model: opus
effort: high
---

# Prompt Enrichment

Turn a raw request into an optimized, execution-ready prompt. The output is a **rewritten prompt**, not an implementation. Never start coding or answering the underlying task here. This skill produces, refines, and optionally saves the prompt — nothing else.

## Required input

- **raw prompt** (`$ARGUMENTS` when invoked as an argument, otherwise the user's latest request): the original instruction to enrich. If none is discernible, ask the user to paste it instead of inventing one.

## Resolving the `.rc` base directory

RC supports monorepos, where more than one `.rc` directory can exist. Before reading or writing any `.rc/...` path, resolve which `.rc` directory this run uses; its parent is the base directory. Treat every `.rc/...` path in this skill as relative to that base.

1. Search the project recursively for `.rc` directories, skipping `node_modules`, `.git`, `vendor`, and any `_archived/` directory.
2. Resolve the base from what you find:
   - **None found** — use `.rc/` at the project root, creating it on first write. Ordinary single-folder projects behave exactly as before.
   - **Exactly one found** — use it without asking.
   - **Two or more found** — a prompt carries no feature slug to disambiguate with, so ask the user which `.rc` to use via the interactive question tool that pauses execution, listing the discovered directories by their path relative to the project root.

## Procedure

When this skill instructs you to ask the user a question, you MUST use your runtime's dedicated interactive question tool — the tool or function that presents a question to the user and **pauses execution until the user responds**. Do not output questions as plain assistant text and continue generating; always use the mechanism that blocks until the user has answered.

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

**Step 5 — Output.** Print only the enriched prompt (inside a fenced block so it is copy-pasteable) plus, if any, the Open questions. If invoked inside a RC flow, the enriched prompt is the input the next skill/command consumes.

**Step 6 — Resolve the open questions.** Skip this step entirely if step 3 produced no `## Open questions` section. Otherwise ask all of them in one round via the interactive question tool that pauses execution. Fold each answer into the section it belongs to (Requirements, Constraints, Out of scope), drop the `## Open questions` block, and **reprint the full updated prompt**. One round only: whatever is still undecided becomes a line under `## Assumptions` in the prompt, not a second round of questions.

**Step 7 — Offer to save.** Always, whether or not step 6 ran. Ask the user — via the interactive question tool that pauses execution — whether to save the enriched prompt. If they decline, say the prompt was not saved and stop. If they accept, write it to `.rc/prompts/NN-<slug>.md` under the resolved base:

- `NN` is zero-padded to 2 digits and **increments past the highest `NN` already present** in `.rc/prompts/` — never the file count, since a deleted prompt must not recycle its number. With no files, start at `01`.
- `<slug>` is kebab-case derived from `## Objective`, 3–5 words, **in the language of the original prompt** (see Rules).
- Write the final enriched prompt as the file body, without the surrounding fenced block.
- Create `.rc/prompts/` if it does not exist.
- Print the path of the file written.

## Rules

- **Preserve intent.** Sharpen and structure the user's request; never substitute a different task.
- **Language.** Match the language of the original prompt.
