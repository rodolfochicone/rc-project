# Roadmap format

The exact shape of `.rc/ROADMAP.md`. A loop reads this file every iteration, so the structure
is a contract, not a style preference.

## Template

```markdown
# Roadmap — <project or milestone name>

> Autonomous loop source of truth. `rc-loop` reads this file each iteration (phase goal +
> sub-items). The per-phase tasks are derived just-in-time by rc-create-tasks against the
> codebase, workflow memory, and confirmed lessons. A checkbox flips to `[x]` only when a
> verification PASS is recorded for that phase.

**Legend:** `[ ]` not started · `[~]` in progress · `[x]` done (verification PASS)

**Hard dependency order:** <state the chain, e.g. "Phase 3 (auth) precedes 4–6.">
Never run phases in parallel.

---

## Phase 1 — <epic title> `[x]`

> Done when: <one-line, observable, runnable acceptance gate>.

- [x] <high-level sub-item>
- [x] <high-level sub-item>

## Phase 2 — <epic title> `[ ]`

> Done when: <one-line acceptance gate>.

- [ ] <high-level sub-item>
- [ ] <high-level sub-item>
```

## Invariants the loop depends on

1. **Phase granularity is an epic.** Each phase ends in something runnable and is small enough to
   plan into tasks once the previous phase is done — but not so small it is a single task. If a
   phase reads like one task, merge it up; if it reads like a quarter's work, split it.
2. **Every phase has a `> Done when:` gate** — one observable, verifiable line. This is what the
   Verifier checks and what justifies flipping the checkbox.
3. **Checkbox states are exactly `[ ]`, `[~]`, `[x]`.** `[~]` marks the phase a loop is currently
   in (so a resumed loop knows where it stopped); `[x]` requires a recorded PASS.
4. **Dependency order is explicit and honored.** The next actionable phase is the first `[ ]`/`[~]`
   whose dependencies are all `[x]`. Phases never run in parallel — a creator loop that forks the
   working tree perpetuates conflicting side effects.
5. **Phase titles are stable.** The per-phase task directory (`.rc/tasks/<phase-slug>/`) is keyed
   off the title; renaming a phase mid-flight orphans its tasks and handoff.
6. **The roadmap holds outline only.** No task breakdowns, no code, no design detail inline — that
   lives in the phase's `task_NN.md`, its TechSpec, and workflow memory. Keeping the roadmap thin
   is what lets a human reorder it in one glance.

## What a phase is NOT

- Not a task list (that is `rc-create-tasks` output).
- Not a decision log (that is `rc-workflow-memory` shared memory / STATE decisions).
- Not a place to record why something failed (that is `.rc/LESSONS.md` via `rc-lessons`).
