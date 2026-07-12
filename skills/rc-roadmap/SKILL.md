---
name: rc-roadmap
description: Author and read the loop roadmap — the human-owned list of high-level phases (epics) that an autonomous creator loop walks one at a time. Use to create or reorder .rc/ROADMAP.md before starting a loop, to add the next batch of phases when a loop exhausts, or to read the next unchecked phase. Creating and reordering phases is the human intent step a loop cannot do for itself. Do not use to plan a single phase's tasks (rc-create-tasks), to run the loop (rc-loop), or to track task-level status (the task files under .rc/tasks/).
argument-hint: "[create|next|status]"
user-invocable: true
model: sonnet
effort: medium
---

# Loop Roadmap

`.rc/ROADMAP.md` is the **source of truth an autonomous loop reads each iteration**. It holds
high-level phases (epics), not task detail — a phase is planned in full only when the one before
it is done, because a creator loop builds on what already exists and cannot plan the far future
blind. Authoring and reordering phases is a **human decision**: the loop executes intent, it
does not invent it (a loop resolves *how*, never *whether*). Confirm the roadmap with the user
before a loop consumes it.

Read `references/roadmap-format.md` before creating or editing the file — it holds the exact
template, the checkbox legend, and the invariants a loop depends on.

## create — author or extend the roadmap

1. Establish the project's `.rc/` base (nearest `.rc` walking up; default `./.rc`).
2. If `.rc/ROADMAP.md` is missing, do a brief exploration (or read the PRD/TechSpec under
   `.rc/tasks/` if present) and draft phases at **epic** granularity — each ends in something
   runnable, with a one-line **`> Done when:`** gate.
3. State the **hard dependency order** and mark that phases never run in parallel.
4. **Confirm the phase list and order with the user** before any loop runs. This is the intent
   gate; do not skip it.

When a loop exhausts (see `rc-loop`), the human re-enters here to research and append the next
batch of phases — the roadmap grows in stages, it is not written all at once.

## next — resolve the next actionable phase

Return the first phase whose checkbox is `[ ]` (or `[~]`) **and** whose dependencies are all
`[x]`. If the next phase's dependencies are unmet, report the blocker instead of guessing an
order. If every phase is `[x]`, report the roadmap is complete — the loop stops and the human
decides what comes next.

## status — summarize progress

Print the phase list with each checkbox state and the current actionable phase (or "complete").

## Rules

- **Phases are epics, not tasks.** Detail belongs in the per-phase `task_NN.md` (authored by
  `rc-create-tasks` when the loop reaches that phase), never inline in the roadmap.
- **A checkbox flips to `[x]` only on a recorded verification PASS** — never by hand to "move
  the loop along". The gate owns "done".
- **Human owns intent.** Creating, reordering, and re-scoping phases is a confirmed human step.
- **One roadmap per project** under `.rc/ROADMAP.md`. Keep it the durable outline; keep decisions
  and handoff in workflow memory (`rc-workflow-memory`), lessons in `.rc/LESSONS.md` (`rc-lessons`).
