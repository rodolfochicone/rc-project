# Loop readiness — should this even be a loop?

Autonomous looping is not a default; it is a mode you earn. Most enterprise feature work should
stay in human-gated spec-driven flow (`rc-pipe` / `rc-card`). Enter an autonomous loop only for
the cases it actually pays off — **migrations, large mechanical build-outs, and automations** —
and only after the four questions below are all "yes". A loop does not fix a weak harness, and it
does not supply intent.

## The four questions (all must be "yes")

1. **Is the harness strong enough that you barely review PRs?**
   If every PR still needs hand-correction, the harness is not ready — each loop iteration will
   perpetuate the same errors and compound them. Strengthen tests/lint/types first. The harness
   (tests, types, lint, compiler, spec-derived acceptance checks) is the sensor the loop steers by;
   a weak sensor steers it off a cliff.

2. **Is feedback fast?**
   The gate must run quickly (seconds, not minutes). A slow gate burns tokens and wall-clock on
   every iteration and stalls the loop. Slow suites are a defect to fix, not a cost to accept.

3. **Is there a reliable stop condition?**
   Something the loop can hit that says "stop and call the human" — roadmap exhausted, a phase that
   cannot reach green, or a decision the loop cannot assume. Without it, the loop wanders. `rc-loop`
   encodes these; confirm they actually fire for your project.

4. **Is the backlog big enough to be worth it?**
   Enough queued phases that unattended looping beats you planning and doing each one yourself. For
   one or two phases, human-gated flow is faster and safer.

If any answer is "no": stay in `rc-pipe` / `rc-card`, and spend the effort on the harness instead.

## Which loop is this?

| Type | Side effects? | Safe to repeat? | Use in RC | Autonomy |
| --- | --- | --- | --- | --- |
| **Fixed loop** | No | Yes — a re-run does not worsen the result | Benchmarks, evaluations, repeated review/QA passes, automations | Safest place for autonomy |
| **Creator loop** | Yes | No — each iteration builds on (and can corrupt) the last | Roadmap-driven build-out / migration via `rc-loop` | Highest value, highest risk; gate it hard |
| **Human-gated** | Yes | N/A | Everything else: intent (author roadmap), outward writes (PR, Linear, state moves), product/architecture forks | Not autonomous by design |

A **fixed** loop has no side effects — the second run cannot be poisoned by the first, so it is
the safest thing to automate. A **creator** loop generates artifacts that later iterations depend
on; a bug shipped in an early phase propagates, so it demands the strong harness the four questions
test for. **Human-gated** work — deciding *what* to build, and any action that reaches outside the
repo — is never folded into autonomous looping; the loop resolves *how*, never *whether*.

## The harness is everything

The single biggest predictor of a loop that works is the harness under it. Prefer sensors the
agent cannot rationalize away: a compiler error, a failing typed test, a lint violation are
unambiguous; "does this look right?" is not. For work with a visual or runtime surface, verify by
**driving the real thing and observing it** (the `verify` / `run` skills, browser automation),
not only by a green unit suite — a green logic test can hide a broken-looking or broken-at-runtime
result. Fold each miss the gate lets through back in as a lesson (`rc-lessons`) so the sensor gets
sharper every phase.
