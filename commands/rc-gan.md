---
description: Adversarial generator↔evaluator loop to drive subjective quality (UI/UX, CLI ergonomics, copy) up to a target score by exercising the running artifact.
argument-hint: "[target] [--threshold 7.0] [--max-iterations 8]"
disable-model-invocation: true
---

You are running the **rc GAN harness** for: $ARGUMENTS

Invoke the `rc-gan` skill with the target and any `--threshold` / `--max-iterations` flags above. Follow it exactly:

- Phase 0: write the spec and a weighted rubric, and state how the artifact will be exercised (run it for real).
- Loop: **generate** (improve + keep the build's verification gate green), then spawn a **fresh, independent evaluator** that exercises the running artifact and scores it against the rubric, then iterate on its concrete defects.
- Stop on threshold, plateau (no real gain over 2 iterations), or max iterations.

Report the rubric, the score progression, the outcome (passed / plateaued / capped), the artifact location, and any remaining gaps. This loop owns subjective quality only — correctness/security still go through `/rc-review`.
