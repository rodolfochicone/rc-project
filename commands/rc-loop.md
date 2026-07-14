---
description: Autonomous creator loop — readiness gate, then walk .rc/ROADMAP.md phase by phase (plan → execute → verify → learn) until it's exhausted or a phase can't reach green.
argument-hint: [roadmap-or-milestone]
disable-model-invocation: true
---

You are starting the **RC autonomous loop** for: $ARGUMENTS

The loop is **earned, not default**. Most feature work belongs in `/rc-pipe` (human-gated). Drive
the three gates below in order; stop at the first one that fails — never skip ahead into the loop.

1. **Readiness gate.** Read `skills/rc-loop/references/loop-readiness.md` and answer its four
   questions with the user (harness strong? feedback fast? reliable stop condition? backlog big
   enough?). **Any "no" → stop** and tell the user to stay in `/rc-pipe` and spend the effort on the
   harness instead. A loop does not fix a weak harness — it compounds its errors.
2. **Intent gate.** If `.rc/ROADMAP.md` is missing, invoke `rc-roadmap` (`create`) to author the
   phases and **confirm the phase list and order with the user**. The loop executes intent; it
   never invents it. If the roadmap exists, invoke `rc-roadmap` (`status`) and show where it stands.
3. **Run the loop.** Invoke the `rc-loop` skill. It walks each actionable phase — load lessons →
   plan → execute → verify → record lessons → close — and stops on its own when the roadmap is
   exhausted, a phase can't reach a green gate, or it hits a decision it cannot assume.

Outward-facing actions stay human-gated: the loop leaves a green, committed working tree per phase.
Opening a PR (`/rc-git`), pushing, and Linear/Jira writes are **not** part of the loop.

At the end, report the phases closed, the phases left, and any lesson the loop recorded.
