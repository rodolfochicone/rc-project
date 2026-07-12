---
description: Ship the current work as a branch + PR (rc-git), then distill session learnings (rc-memory).
argument-hint: "[linear-issue]"
disable-model-invocation: true
---

You are running the **ship + learn** step for: $ARGUMENTS

1. **Ship** — invoke the `rc-git` skill to move the current changes onto a feature branch, push, and open a PR. It confirms each outward-facing step (branch, push, PR) and verifies the PR target; pass the optional Linear issue above through to it.
2. **Learn** — once the PR is open, invoke the `rc-memory` skill to distill recurring corrections and patterns from this session into `.rc/memory/LEARNINGS.md`. It is a no-op when nothing durable is worth capturing, so it is safe to always run (capture is on by default; set `RC_INSTINCTS=0` to opt out).

Report the PR URL and any instincts recorded. If the ship step stops (nothing to ship, or the user declines an outward-facing action), stop there and do not run the learn step.
