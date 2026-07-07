---
name: rc-loop
description: Drives a task through a generate→verify→retry loop against an explicit, machine-checkable success gate (a test, build, lint, arbitrary command, file existence, or an independent reviewer), retrying on concrete failure feedback up to a cap, then reporting the outcome or escalating. Use when "done" can be defined by a concrete pass/fail check and the work may need several attempts. Do not use for subjective quality with no pass/fail gate (use rc-gan) or for open-ended exploration (use rc-analyze).
model: sonnet
effort: high
user-invocable: true
argument-hint: "[goal] [--success test|build|lint|command|fileExists|review] [--max 3]"
---

# Loop

Run a task in a disciplined loop until an explicit success gate passes or an attempt cap is hit. Each round acts on the *concrete failure output* of the last, so the loop converges instead of thrashing.

## Setup (define the loop before running it)

Establish these up front — ask the user for anything missing:

1. **Goal** — what the loop is trying to accomplish, in one sentence.
2. **Success gate** — how you'll know it passed. Pick one and make it machine-checkable:
   - `test` / `build` / `lint` — the project's real command for that gate.
   - `command` — an arbitrary shell command; success = exit 0. Provide the exact command.
   - `fileExists` — success = a given path exists (and, optionally, matches a pattern).
   - `review` — an independent reviewer (`rc-review` / `rc-code-review`) reports no blocking findings.
3. **Max attempts** — default 3.
4. **Context** — files or directories to read before the first attempt.

## Loop

For attempt `n` from 1 to `max`:

1. **Generate** — do the work (or delegate to `rc-exec`), addressing the previous attempt's specific failure.
2. **Verify** — run the success gate for real and read its full output. Never assume; require evidence (`rc-final-verify` discipline).
3. **Decide**:
   - Gate passes → **stop, success.** Report what passed and the evidence.
   - Gate fails and `n < max` → capture the concrete failure (exact error, failing test, missing file), feed it into the next attempt.
   - Gate fails and `n == max` → **stop, escalate.** Report every attempt, what was tried, and why it did not converge. Do not claim success.

## Rules

- The gate must be objective. If you cannot state a pass/fail check, this is the wrong skill — use `rc-gan` (subjective) or `rc-analyze` (exploration).
- Never fake or loosen the gate to make the loop pass. A gate that always passes is a bug.
- Bound the loop: escalate at the cap rather than looping indefinitely.
- Each retry must change something informed by the last failure — never repeat an identical attempt.
