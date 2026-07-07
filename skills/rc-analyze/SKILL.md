---
name: rc-analyze
description: Performs a deep, evidence-based analysis of an existing codebase to answer a specific analytical prompt — diagnosing the root cause of a bug, hunting inconsistencies and contradictions, understanding how something works, or tracing a behavior or data flow — then writes and prints a thorough report, ending with an actionable implementation plan when a change is implied. Use to understand, trace, or diagnose existing code, or to assess impact and feasibility. Do not use to review a change set or diff for defects (use rc-code-review), to generate remediation issues (use rc-review-round), to apply a plan's code changes (use rc-fix-analysis), or to edit source code directly.
model: opus
effort: xhigh
---

# Deep Codebase Analysis

Investigate the codebase to answer the user's analytical prompt — whether that is how something works, why a bug happens, or whether two parts of the system are consistent — then deliver a thorough, evidence-based report. This skill is read-only — it explains, traces, diagnoses, and assesses; it does not change code. It is standalone and stack-agnostic; it detects the stack and reasons in that stack's idioms.

## Code navigation (Serena)

If the Serena MCP is available, prefer its symbolic tools over whole-file reads — they are LSP-accurate and token-efficient:

- `get_symbols_overview` to grasp a file's structure before reading it; `find_symbol` (by name path, e.g. `Type/method`) to jump straight to a definition.
- `find_referencing_symbols` to map every caller of a symbol before reasoning about impact.

Fall back to Grep/Glob + Read when Serena is unavailable or for plain-text (non-symbol) searches.

## Required Inputs

- The analytical prompt: the question or topic to investigate (e.g. "how does session auth work", "why does the cache return stale data under load", "is the locking around the shared registry consistent or is there a race", "what is the impact of changing the config file format", "where does input validation happen and is it consistent").
- Optional: a slug or scope (specific files/directories/packages) to focus the analysis.

## Resolving the `.rc` base directory

rc supports monorepos, where more than one `.rc` directory can exist. Before reading or writing any `.rc/...` path, resolve which `.rc` directory this run uses; its parent is the base directory. Treat every `.rc/...` path in this skill as relative to that base.

1. Search the project recursively for `.rc` directories, skipping `node_modules`, `.git`, `vendor`, and any `_archived/` directory.
2. Resolve the base from what you find:
   - **None found** — use `.rc/` at the project root, creating it on first write. Ordinary single-folder projects behave exactly as before.
   - **Exactly one found** — use it without asking.
   - **Two or more found** — if the prompt names a feature that maps to one `.rc/tasks/<slug>/`, use that `.rc`. Otherwise ask the user which `.rc` to use via the interactive question tool that pauses execution, listing the discovered directories by their path relative to the project root.

## Workflow

1. Frame the question precisely. Restate the prompt as one or more concrete questions the report must answer. If the prompt is ambiguous or could be interpreted several materially different ways, ask the user one focused clarifying question before investigating; otherwise state your interpretation and proceed.

2. Establish project context and stack. Read `CLAUDE.md`, `AGENTS.md`, `CONTRIBUTING.md`, and any architecture notes or ADRs (including `.rc/tasks/<slug>/_prd.md`, `_techspec.md`, and `adrs/` when present and relevant). Identify the language(s) and frameworks from manifest and config files so the analysis reasons in idiomatic terms.

3. Investigate breadth-first, then depth-first.
   - Locate the relevant entry points and symbols with Grep/Glob and the symbol tools. Do not use web search for local code.
   - Map the relevant slice: callers, callees, interfaces, data structures, and the flow of control and data through them. Trace behavior end to end rather than reading a single file in isolation.
   - For broad or unfamiliar areas, spawn an `Explore` agent (or several in parallel) to map structure and naming conventions, then read the pivotal files yourself in full before drawing conclusions.
   - Ground every claim in a specific `file:line` reference. Read the code that proves a claim — never infer behavior from a name alone.

4. Reason about findings. Synthesize what the code actually does (not what it is named to do), surface invariants and assumptions, identify edge cases and failure modes, and call out contradictions, dead code, or divergence from documented conventions. When diagnosing a bug, trace to the root cause — the specific code and condition that produce the wrong behavior — rather than stopping at the symptom; state the precise trigger condition (the input or state that makes it manifest) and distinguish a root cause you have **confirmed** in the code from one you only **hypothesize**, naming the evidence that would settle it. When hunting inconsistencies, name the two (or more) places that disagree and why it matters. When more than one explanation fits the evidence, hold the competing hypotheses side by side and actively look for the code that would **disconfirm** your leading one — settling for the first plausible story is how confident-but-wrong conclusions get made. Separate established fact (backed by a reference) from inference (your reasoning), and label each clearly. State your confidence and name the open questions you could not resolve from the code.

5. Write the report and print it.
   - If a feature slug applies, write to `.rc/tasks/<slug>/analysis-NNN.md`; otherwise write to `.rc/analysis/<topic-slug>-NNN.md`, deriving `<topic-slug>` from the prompt. `NNN` is zero-padded and increments past any existing matching file so prior analyses are preserved.
   - Print the same content to the user.
   - Structure the report as:

   ```
   ANALYSIS — <topic>
   ==================
   Question:   <the framed question(s)>
   Scope:      <files/packages examined>
   Coverage:   <what was read in full vs. surveyed vs. deliberately not examined>
   Verdict:    <the one-paragraph answer up front>
   ```

   State coverage honestly: an investigation almost never touches every relevant file, and naming what you did *not* read lets the reader judge how complete the conclusion is and where a blind spot might hide.

   Then the body:
   - **Summary** — the direct answer to the prompt, in a few sentences.
   - **How it works** — the mechanism, traced step by step, each step anchored to `file:line`.
   - **Key findings** — the load-bearing observations, fact vs. inference labeled.
   - **Risks, edge cases & gaps** — failure modes, assumptions, dead code, convention divergence (omit if genuinely none).
   - **Open questions** — what remains unresolved and what would resolve it (omit if none).
   - **Implementation plan** — when the analysis points to a concrete change (a fix, refactor, or feature the prompt asks to scope), close the report with an ordered, actionable plan that the `rc-fix-analysis` skill can execute. Each step states the change, the target `file:line`, and a per-step done criterion. This section describes the plan only — it does not edit code. Omit it for purely explanatory analyses where no change is implied.
   - **References** — the `file:line` anchors the conclusions rest on.

6. Close with the bottom-line answer restated in one or two sentences, and state the report path. If an Implementation plan section was written, note that `rc-fix-analysis` can execute it.

## Project memory

Before investigating, search `.rc/memory/` (with Grep) for the question's key terms to recover
prior decisions, conventions, and gotchas that may already answer or inform it (see the
`rc-project-memory` skill).

## Critical Rules

- Do not modify source code — this skill is the read-only half of the pair, so its only writes are the report under `.rc/...`; code changes belong to `rc-fix-analysis`, which executes the plan.
- Lead with the answer, because the reader needs the verdict to act and should not have to reconstruct it from the trace.
- Every factual claim cites a `file:line`, and distinguish what the code does from what you infer it does — a named function may not do what its name says, so grounding is what separates analysis from a confident guess.
- Depth over breadth padding: answer the question fully but do not inflate the report with tangential surveys, since padding buries the load-bearing findings and wastes the reader's attention.
- Do not settle for the first explanation that fits — confirmation bias produces confident, wrong diagnoses; weigh competing hypotheses and seek the evidence that would disconfirm yours.
- Reason in the target stack's idioms, and surface a harmful convention rather than silently endorsing it, so the report does not normalize a latent problem.
- When the code is genuinely ambiguous or you ran out of evidence, say so — guessing to appear complete is worse than a named gap the reader can follow up on.

## Error Handling

- If the prompt is too vague to investigate, ask one focused clarifying question before proceeding.
- If the scope is enormous, state the triage you applied (what you analyzed in full vs. surveyed) so the reader knows the report's coverage.
- If a file relevant to the analysis cannot be read, note which file and continue with the rest.
